// Copyright 2017 AMIS Technologies
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package container

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"
	docker "github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/ethereum/go-ethereum/cmd/utils"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/p2p/discv5"

	"github.com/Consensys/istanbul-tools/client"
	istcommon "github.com/Consensys/istanbul-tools/common"
	"github.com/Consensys/istanbul-tools/genesis"
)

const (
	healthCheckRetryCount = 30
	healthCheckRetryDelay = 2 * time.Second
)

var (
	ErrNoBlock = errors.New("no block generated")
	ErrTimeout = errors.New("timeout")
)

type Ethereum interface {
	Init(string) error
	Start() error
	Stop() error

	NodeAddress() string
	Address() common.Address

	ContainerID() string
	Host() string
	NewClient() client.Client
	ConsensusMonitor(err chan<- error, quit chan struct{})

	WaitForProposed(expectedAddress common.Address, t time.Duration) error
	WaitForPeersConnected(int) error
	WaitForBlocks(int, ...time.Duration) error
	WaitForBlockHeight(int) error
	// Want for block for no more than the given number during the given time duration
	WaitForNoBlocks(int, time.Duration) error

	// Wait for settling balances for the given accounts
	WaitForBalances([]common.Address, ...time.Duration) error

	AddPeer(string) error

	StartMining() error
	StopMining() error

	Accounts() []common.Address

	DockerEnv() []string
	DockerBinds() []string
}

func NewEthereum(c *docker.Client, options ...Option) *ethereum {
	eth := &ethereum{
		dockerClient: c,
	}

	for _, opt := range options {
		opt(eth)
	}

	filters := filters.NewArgs()
	filters.Add("reference", eth.Image())

	images, err := c.ImageList(context.Background(), types.ImageListOptions{
		Filters: filters,
	})

	if len(images) == 0 || err != nil {
		out, err := eth.dockerClient.ImagePull(context.Background(), eth.Image(), types.ImagePullOptions{})
		if err != nil {
			log.Error("Failed to pull image", "image", eth.Image(), "err", err)
			return nil
		}
		if eth.logging {
			io.Copy(os.Stdout, out)
		} else {
			io.Copy(ioutil.Discard, out)
		}
	}

	return eth
}

type ethereum struct {
	ok          bool
	flags       []string
	dataDir     string
	ip          string
	port        string
	rpcPort     string
	wsPort      string
	hostName    string
	containerID string
	node        *discv5.Node
	accounts    []common.Address
	password    string

	//Quorum only
	isQuorum    bool
	dockerEnv   []string
	dockerBinds []string

	imageRepository   string
	imageTag          string
	dockerNetworkName string

	key          *ecdsa.PrivateKey
	logging      bool
	dockerClient *docker.Client
}

func (eth *ethereum) Init(genesisFile string) error {
	if err := istcommon.SaveNodeKey(eth.key, eth.dataDir); err != nil {
		return err
	}

	binds := []string{
		genesisFile + ":" + filepath.Join("/", genesis.FileName),
	}
	if eth.dataDir != "" {
		binds = append(binds, eth.dataDir+":"+utils.DataDirFlag.Value.Value)
	}

	resp, err := eth.dockerClient.ContainerCreate(context.Background(),
		&container.Config{
			Image: eth.Image(),
			Cmd: []string{
				"init",
				"--" + utils.DataDirFlag.Name,
				utils.DataDirFlag.Value.Value,
				filepath.Join("/", genesis.FileName),
			},
		},
		&container.HostConfig{
			Binds: binds,
		}, nil, "")
	if err != nil {
		log.Error("Failed to create container", "err", err)
		return err
	}

	id := resp.ID

	if err := eth.dockerClient.ContainerStart(context.Background(), id, types.ContainerStartOptions{}); err != nil {
		log.Error("Failed to start container", "err", err)
		return err
	}

	_, err = eth.dockerClient.ContainerWait(context.Background(), id)
	if err != nil {
		log.Error("Failed to wait container", "err", err)
		return err
	}

	if eth.logging {
		eth.showLog(context.Background())
	}

	return eth.dockerClient.ContainerRemove(context.Background(), id, types.ContainerRemoveOptions{Force: true})
}

func (eth *ethereum) Start() error {
	defer func() {
		if eth.logging {
			go eth.showLog(context.Background())
		}
	}()

	exposedPorts := make(map[nat.Port]struct{})
	portBindings := nat.PortMap{}

	if eth.rpcPort != "" {
		port := fmt.Sprintf("%d", utils.RPCPortFlag.Value)
		exposedPorts[nat.Port(port)] = struct{}{}
		portBindings[nat.Port(port)] = []nat.PortBinding{
			{
				HostIP:   "0.0.0.0",
				HostPort: eth.rpcPort,
			},
		}
	}

	if eth.wsPort != "" {
		port := fmt.Sprintf("%d", utils.WSPortFlag.Value)
		exposedPorts[nat.Port(port)] = struct{}{}
		portBindings[nat.Port(port)] = []nat.PortBinding{
			{
				HostIP:   "0.0.0.0",
				HostPort: eth.wsPort,
			},
		}
	}

	binds := []string{}
	binds = append(binds, eth.dockerBinds...)
	if eth.dataDir != "" {
		binds = append(binds, eth.dataDir+":"+utils.DataDirFlag.Value.Value)
	}

	var networkingConfig *network.NetworkingConfig
	if eth.ip != "" && eth.dockerNetworkName != "" {
		endpointsConfig := make(map[string]*network.EndpointSettings)
		endpointsConfig[eth.dockerNetworkName] = &network.EndpointSettings{
			IPAMConfig: &network.EndpointIPAMConfig{
				IPv4Address: eth.ip,
			},
		}
		networkingConfig = &network.NetworkingConfig{
			EndpointsConfig: endpointsConfig,
		}
	}

	resp, err := eth.dockerClient.ContainerCreate(context.Background(),
		&container.Config{
			Hostname:     "geth-" + eth.hostName,
			Image:        eth.Image(),
			Cmd:          eth.flags,
			ExposedPorts: exposedPorts,
			Env:          eth.DockerEnv(),
		},
		&container.HostConfig{
			Binds:        binds,
			PortBindings: portBindings,
		}, networkingConfig, "")
	if err != nil {
		log.Error("Failed to create container", "err", err)
		return err
	}

	eth.containerID = resp.ID

	err = eth.dockerClient.ContainerStart(context.Background(), eth.containerID, types.ContainerStartOptions{})
	if err != nil {
		log.Error("Failed to start container", "ip", eth.ip, "err", err)
		return err
	}

	for i := 0; i < healthCheckRetryCount; i++ {
		cli := eth.NewClient()
		if cli == nil {
			<-time.After(healthCheckRetryDelay)
			continue
		}
		_, err = cli.BlockByNumber(context.Background(), big.NewInt(0))
		if err != nil {
			<-time.After(healthCheckRetryDelay)
			continue
		} else {
			eth.ok = true
			break
		}
	}

	if !eth.ok {
		return errors.New("Failed to start geth")
	}

	containerIP := eth.ip
	if containerIP == "" {
		containerJSON, err := eth.dockerClient.ContainerInspect(context.Background(), eth.containerID)
		if err != nil {
			log.Error("Failed to inspect container", "err", err)
			return err
		}
		containerIP = containerJSON.NetworkSettings.IPAddress
	}

	if eth.key != nil {
		eth.node = discv5.NewNode(
			discv5.PubkeyID(&eth.key.PublicKey),
			net.ParseIP(containerIP),
			0,
			uint16(utils.ListenPortFlag.Value))
	}

	return nil
}

func (eth *ethereum) Stop() error {
	err := eth.dockerClient.ContainerStop(context.Background(), eth.containerID, nil)
	if err != nil {
		log.Error("Failed to stop container", "err", err)
		return err
	}

	defer os.RemoveAll(eth.dataDir)

	return eth.dockerClient.ContainerRemove(context.Background(), eth.containerID,
		types.ContainerRemoveOptions{
			Force: true,
		})
}

func (eth *ethereum) Wait(t time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), t)
	defer cancel()
	_, err := eth.dockerClient.ContainerWait(ctx, eth.containerID)
	return err
}

func (eth *ethereum) Running() bool {
	containers, err := eth.dockerClient.ContainerList(context.Background(), types.ContainerListOptions{})
	if err != nil {
		log.Error("Failed to list containers", "err", err)
		return false
	}

	for _, c := range containers {
		if c.ID == eth.containerID {
			return true
		}
	}

	return false
}

func (eth *ethereum) NewClient() client.Client {
	var scheme, port string

	if eth.rpcPort != "" {
		scheme = "http://"
		port = eth.rpcPort
	}
	if eth.wsPort != "" {
		scheme = "ws://"
		port = eth.wsPort
	}
	client, err := client.Dial(scheme + eth.Host() + ":" + port)
	if err != nil {
		return nil
	}
	return client
}

func (eth *ethereum) NodeAddress() string {
	if eth.node != nil {
		return eth.node.String()
	}

	return ""
}

func (eth *ethereum) Address() common.Address {
	return crypto.PubkeyToAddress(eth.key.PublicKey)
}

func (eth *ethereum) ConsensusMonitor(errCh chan<- error, quit chan struct{}) {
	cli := eth.NewClient()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	subCh := make(chan *ethtypes.Header)

	sub, err := cli.SubscribeNewHead(ctx, subCh)
	if err != nil {
		log.Error("Failed to subscribe new head", "err", err)
		errCh <- err
		return
	}
	defer sub.Unsubscribe()

	timer := time.NewTimer(10 * time.Second)
	defer timer.Stop()
	blockNumber := uint64(0)
	for {
		select {
		case err := <-sub.Err():
			log.Error("Connection lost", "err", err)
			errCh <- err
			return
		case <-timer.C: // FIXME: this event may be missed
			if blockNumber == 0 {
				errCh <- ErrNoBlock
			} else {
				errCh <- ErrTimeout
			}
			return
		case head := <-subCh:
			blockNumber = head.Number.Uint64()
			// Ensure that mining is stable.
			if head.Number.Uint64() < 3 {
				continue
			}

			// Block is generated by 2 seconds. We tolerate 1 second delay in consensus.
			timer.Reset(3 * time.Second)
		case <-quit:
			return
		}
	}
}

// TODO: refactor with ConsensusMonitor
func (eth *ethereum) WaitForProposed(expectedAddress common.Address, timeout time.Duration) error {
	cli := eth.NewClient()

	subCh := make(chan *ethtypes.Header)

	sub, err := cli.SubscribeNewHead(context.Background(), subCh)
	if err != nil {
		return err
	}
	defer sub.Unsubscribe()

	timer := time.NewTimer(timeout)
	defer timer.Stop()
	for {
		select {
		case err := <-sub.Err():
			return err
		case <-timer.C: // FIXME: this event may be missed
			return errors.New("no result")
		case head := <-subCh:
			if GetProposer(head) == expectedAddress {
				return nil
			}
		}
	}
}

func (eth *ethereum) WaitForPeersConnected(expectedPeercount int) error {
	client := eth.NewClient()
	if client == nil {
		return errors.New("failed to retrieve client")
	}
	defer client.Close()

	ticker := time.NewTicker(time.Second * 1)
	defer ticker.Stop()
	for _ = range ticker.C {
		infos, err := client.AdminPeers(context.Background())
		if err != nil {
			return err
		}
		if len(infos) < expectedPeercount {
			continue
		} else {
			break
		}
	}

	return nil
}

func (eth *ethereum) WaitForBlocks(num int, waitingTime ...time.Duration) error {
	var first *big.Int

	client := eth.NewClient()
	if client == nil {
		return errors.New("failed to retrieve client")
	}
	defer client.Close()

	var t time.Duration
	if len(waitingTime) > 0 {
		t = waitingTime[0]
	} else {
		t = 1 * time.Hour
	}

	timeout := time.After(t)
	ticker := time.NewTicker(time.Millisecond * 500)
	defer ticker.Stop()
	for {
		select {
		case <-timeout:
			ticker.Stop()
			return ErrNoBlock
		case <-ticker.C:
			n, err := client.BlockNumber(context.Background())
			if err != nil {
				return err
			}
			if first == nil {
				first = new(big.Int).Set(n)
				continue
			}
			// Check if new blocks are getting generated
			if new(big.Int).Sub(n, first).Int64() >= int64(num) {
				return nil
			}
		}
	}
}

func (eth *ethereum) WaitForBlockHeight(num int) error {
	client := eth.NewClient()
	if client == nil {
		return errors.New("failed to retrieve client")
	}
	defer client.Close()

	ticker := time.NewTicker(time.Millisecond * 500)
	defer ticker.Stop()
	for _ = range ticker.C {
		n, err := client.BlockNumber(context.Background())
		if err != nil {
			return err
		}
		if n.Int64() >= int64(num) {
			break
		}
	}

	return nil
}

func (eth *ethereum) WaitForNoBlocks(num int, duration time.Duration) error {
	var first *big.Int

	client := eth.NewClient()
	if client == nil {
		return errors.New("failed to retrieve client")
	}

	timeout := time.After(duration)
	ticker := time.NewTicker(time.Millisecond * 500)
	defer ticker.Stop()
	for {
		select {
		case <-timeout:
			return nil
		case <-ticker.C:
			n, err := client.BlockNumber(context.Background())
			if err != nil {
				return err
			}
			if first == nil {
				first = new(big.Int).Set(n)
				continue
			}
			// Check if new blocks are getting generated
			if new(big.Int).Sub(n, first).Int64() > int64(num) {
				return errors.New("generated more blocks than expected")
			}
		}
	}
}

func (eth *ethereum) WaitForBalances(addrs []common.Address, duration ...time.Duration) error {
	client := eth.NewClient()
	if client == nil {
		return errors.New("failed to retrieve client")
	}

	var t time.Duration
	if len(duration) > 0 {
		t = duration[0]
	} else {
		t = 1 * time.Hour
	}

	waitBalance := func(addr common.Address) error {
		timeout := time.After(t)
		ticker := time.NewTicker(time.Millisecond * 500)
		defer ticker.Stop()
		for {
			select {
			case <-timeout:
				return ErrTimeout
			case <-ticker.C:
				n, err := client.BalanceAt(context.Background(), addr, nil)
				if err != nil {
					return err
				}

				// Check if new blocks are getting generated
				if n.Uint64() <= 0 {
					continue
				} else {
					return nil
				}
			}
		}
	}

	var wg sync.WaitGroup
	errc := make(chan error, len(addrs))
	wg.Add(len(addrs))

	for _, addr := range addrs {
		addr := addr
		go func() {
			defer wg.Done()
			errc <- waitBalance(addr)
		}()
	}
	// Wait for the first error, then terminate the others.
	var err error
	for i := 0; i < len(addrs); i++ {
		if err = <-errc; err != nil {
			break
		}
	}
	wg.Wait()
	return err
}

// ----------------------------------------------------------------------------

func (eth *ethereum) AddPeer(address string) error {
	client := eth.NewClient()
	if client == nil {
		return errors.New("failed to retrieve client")
	}
	defer client.Close()

	return client.AddPeer(context.Background(), address)
}

func (eth *ethereum) StartMining() error {
	client := eth.NewClient()
	if client == nil {
		return errors.New("failed to retrieve client")
	}
	defer client.Close()

	return client.StartMining(context.Background())
}

func (eth *ethereum) StopMining() error {
	client := eth.NewClient()
	if client == nil {
		return errors.New("failed to retrieve client")
	}
	defer client.Close()

	return client.StopMining(context.Background())
}

func (eth *ethereum) Accounts() []common.Address {
	return eth.accounts
}

func (eth *ethereum) DockerEnv() []string {
	return eth.dockerEnv
}

func (eth *ethereum) DockerBinds() []string {
	return eth.dockerBinds
}

// ----------------------------------------------------------------------------

func (eth *ethereum) showLog(context context.Context) {
	if readCloser, err := eth.dockerClient.ContainerLogs(context, eth.containerID,
		types.ContainerLogsOptions{ShowStderr: true, Follow: true}); err == nil {
		defer readCloser.Close()
		_, err = io.Copy(os.Stdout, readCloser)
		if err != nil && err != io.EOF {
			log.Error("Failed to print container log", "err", err)
			return
		}
	}
}

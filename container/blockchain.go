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
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/docker/docker/client"
	"github.com/ethereum/go-ethereum/common"
	"github.com/phayes/freeport"

	"github.com/getamis/istanbul-tools/genesis"
)

type Blockchain interface {
	AddValidators(numOfValidators int) ([]Ethereum, error)
	RemoveValidators(candidates []Ethereum, t time.Duration) error
	EnsureConsensusWorking(geths []Ethereum, t time.Duration) error
	Start(bool) error
	Stop(bool) error
	Validators() []Ethereum
	Finalize()
	CreateNodes(int, ...Option) ([]Ethereum, error)
}

func NewBlockchain(network *DockerNetwork, numOfValidators int, options ...Option) (bc *blockchain) {
	bc = &blockchain{opts: options}

	var err error
	bc.dockerClient, err = client.NewEnvClient()
	if err != nil {
		log.Fatalf("Cannot connect to Docker daemon, err: %v", err)
	}

	if network == nil {
		bc.defaultNetwork, err = NewDockerNetwork()
		if err != nil {
			log.Fatalf("Cannot create Docker network, err: %v", err)
		}
		network = bc.defaultNetwork
	}

	bc.dockerNetworkName = network.Name()
	bc.getFreeIPAddrs = network.GetFreeIPAddrs
	bc.opts = append(bc.opts, DockerNetworkName(bc.dockerNetworkName))

	bc.addValidators(numOfValidators)
	return bc
}

func NewDefaultBlockchain(network *DockerNetwork, numOfValidators int) (bc *blockchain) {
	return NewBlockchain(network,
		numOfValidators,
		ImageRepository("quay.io/amis/geth"),
		ImageTag("istanbul_develop"),
		DataDir("/data"),
		WebSocket(),
		WebSocketAddress("0.0.0.0"),
		WebSocketAPI("admin,eth,net,web3,personal,miner,istanbul"),
		WebSocketOrigin("*"),
		NAT("any"),
		NoDiscover(),
		Etherbase("1a9afb711302c5f83b5902843d1c007a1a137632"),
		Mine(),
		SyncMode("full"),
		Logging(false),
	)
}

func NewDefaultBlockchainWithFaulty(network *DockerNetwork, numOfNormal int, numOfFaulty int) (bc *blockchain) {
	commonOpts := [...]Option{
		DataDir("/data"),
		WebSocket(),
		WebSocketAddress("0.0.0.0"),
		WebSocketAPI("admin,eth,net,web3,personal,miner,istanbul"),
		WebSocketOrigin("*"),
		NAT("any"),
		NoDiscover(),
		Etherbase("1a9afb711302c5f83b5902843d1c007a1a137632"),
		Mine(),
		SyncMode("full"),
		Logging(false)}
	normalOpts := make([]Option, len(commonOpts), len(commonOpts)+2)
	copy(normalOpts, commonOpts[:])
	normalOpts = append(normalOpts, ImageRepository("quay.io/amis/geth"), ImageTag("istanbul_develop"))
	faultyOpts := make([]Option, len(commonOpts), len(commonOpts)+3)
	copy(faultyOpts, commonOpts[:])
	faultyOpts = append(faultyOpts, ImageRepository("quay.io/amis/geth_faulty"), ImageTag("latest"), FaultyMode(1))

	// New env client
	bc = &blockchain{}
	var err error
	bc.dockerClient, err = client.NewEnvClient()
	if err != nil {
		log.Fatalf("Cannot connect to Docker daemon, err: %v", err)
	}

	if network == nil {
		bc.defaultNetwork, err = NewDockerNetwork()
		if err != nil {
			log.Fatalf("Cannot create Docker network, err: %v", err)
		}
		network = bc.defaultNetwork
	}

	bc.dockerNetworkName = network.Name()
	bc.getFreeIPAddrs = network.GetFreeIPAddrs

	normalOpts = append(normalOpts, DockerNetworkName(bc.dockerNetworkName))
	faultyOpts = append(faultyOpts, DockerNetworkName(bc.dockerNetworkName))

	totalNodes := numOfNormal + numOfFaulty

	ips, err := bc.getFreeIPAddrs(totalNodes)
	if err != nil {
		log.Fatalf("Failed to get free ip addresses, err: %v", err)
	}

	keys, addrs := generateKeys(totalNodes)
	bc.setupGenesis(addrs)
	// Create normal validators
	bc.opts = normalOpts
	bc.setupValidators(ips[:numOfNormal], keys[:numOfNormal], bc.opts...)
	// Create faulty validators
	bc.opts = faultyOpts
	bc.setupValidators(ips[numOfNormal:], keys[numOfNormal:], bc.opts...)
	return bc
}

func NewQuorumBlockchain(network *DockerNetwork, ctn ConstellationNetwork, options ...Option) (bc *blockchain) {
	bc = &blockchain{opts: options, isQuorum: true, constellationNetwork: ctn}
	bc.opts = append(bc.opts, IsQuorum(true))
	bc.opts = append(bc.opts, NoUSB())

	var err error
	bc.dockerClient, err = client.NewEnvClient()
	if err != nil {
		log.Fatalf("Cannot connect to Docker daemon, err: %v", err)
	}

	if network == nil {
		bc.defaultNetwork, err = NewDockerNetwork()
		if err != nil {
			log.Fatalf("Cannot create Docker network, err: %v", err)
		}
		network = bc.defaultNetwork
	}

	bc.dockerNetworkName = network.Name()
	bc.getFreeIPAddrs = network.GetFreeIPAddrs
	bc.opts = append(bc.opts, DockerNetworkName(bc.dockerNetworkName))

	bc.addValidators(ctn.NumOfConstellations())
	return bc
}

func NewDefaultQuorumBlockchain(network *DockerNetwork, ctn ConstellationNetwork) (bc *blockchain) {
	return NewQuorumBlockchain(network,
		ctn,
		ImageRepository("quay.io/amis/quorum"),
		ImageTag("latest"),
		DataDir("/data"),
		WebSocket(),
		WebSocketAddress("0.0.0.0"),
		WebSocketAPI("admin,eth,net,web3,personal,miner,istanbul"),
		WebSocketOrigin("*"),
		NAT("any"),
		NoDiscover(),
		Etherbase("1a9afb711302c5f83b5902843d1c007a1a137632"),
		Mine(),
		Logging(false),
	)
}

// ----------------------------------------------------------------------------

type blockchain struct {
	dockerClient         *client.Client
	defaultNetwork       *DockerNetwork
	dockerNetworkName    string
	getFreeIPAddrs       func(int) ([]net.IP, error)
	genesisFile          string
	isQuorum             bool
	validators           []Ethereum
	opts                 []Option
	constellationNetwork ConstellationNetwork
}

func (bc *blockchain) AddValidators(numOfValidators int) ([]Ethereum, error) {
	// TODO: need a lock
	lastLen := len(bc.validators)
	bc.addValidators(numOfValidators)

	newValidators := bc.validators[lastLen:]
	if err := bc.start(newValidators); err != nil {
		return nil, err
	}

	// propose new validators as validator in consensus
	for _, v := range bc.validators[:lastLen] {
		istClient := v.NewIstanbulClient()
		for _, newV := range newValidators {
			if err := istClient.ProposeValidator(context.Background(), newV.Address(), true); err != nil {
				return nil, err
			}
		}
	}

	if err := bc.connectAll(true); err != nil {
		return nil, err
	}
	return newValidators, nil
}

func (bc *blockchain) EnsureConsensusWorking(geths []Ethereum, t time.Duration) error {
	errCh := make(chan error, len(geths))
	quitCh := make(chan struct{}, len(geths))
	for _, geth := range geths {
		go geth.ConsensusMonitor(errCh, quitCh)
	}

	timeout := time.NewTimer(t)
	defer timeout.Stop()

	var err error
	select {
	case err = <-errCh:
	case <-timeout.C:
		for i := 0; i < len(geths); i++ {
			quitCh <- struct{}{}
		}
	}
	return err
}

func (bc *blockchain) RemoveValidators(candidates []Ethereum, processingTime time.Duration) error {
	var newValidators []Ethereum

	for _, v := range bc.validators {
		istClient := v.NewIstanbulClient()
		isFound := false
		for _, c := range candidates {
			if err := istClient.ProposeValidator(context.Background(), c.Address(), false); err != nil {
				return err
			}
			if v.ContainerID() == c.ContainerID() {
				isFound = true
			}
		}
		if !isFound {
			newValidators = append(newValidators, v)
		}
	}

	// FIXME: It is not good way to wait validator vote out candidates
	<-time.After(processingTime)
	bc.validators = newValidators

	return bc.stop(candidates, false)
}

func (bc *blockchain) Start(strong bool) error {
	if err := bc.start(bc.validators); err != nil {
		return err
	}
	return bc.connectAll(strong)
}

func (bc *blockchain) Stop(force bool) error {
	if err := bc.stop(bc.validators, force); err != nil {
		return err
	}

	if bc.defaultNetwork != nil {
		err := bc.defaultNetwork.Remove()
		bc.defaultNetwork = nil
		return err
	}
	return nil
}

func (bc *blockchain) Finalize() {
	os.RemoveAll(filepath.Dir(bc.genesisFile))
}

func (bc *blockchain) Validators() []Ethereum {
	return bc.validators
}

func (bc *blockchain) CreateNodes(num int, options ...Option) (nodes []Ethereum, err error) {
	ips, err := bc.getFreeIPAddrs(num)
	if err != nil {
		return nil, err
	}

	for i := 0; i < num; i++ {
		var opts []Option
		opts = append(opts, options...)

		// Host data directory
		dataDir, err := generateRandomDir()
		if err != nil {
			log.Println("Failed to create data dir", err)
			return nil, err
		}
		opts = append(opts, HostDataDir(dataDir))
		opts = append(opts, HostWebSocketPort(freeport.GetPort()))
		opts = append(opts, HostIP(ips[i]))
		opts = append(opts, DockerNetworkName(bc.dockerNetworkName))

		geth := NewEthereum(
			bc.dockerClient,
			opts...,
		)

		err = geth.Init(bc.genesisFile)
		if err != nil {
			log.Println("Failed to init genesis", err)
			return nil, err
		}

		nodes = append(nodes, geth)
	}

	return nodes, nil
}

// ----------------------------------------------------------------------------

func (bc *blockchain) addValidators(numOfValidators int) error {
	ips, err := bc.getFreeIPAddrs(numOfValidators)
	if err != nil {
		return err
	}
	keys, addrs := generateKeys(numOfValidators)
	bc.setupGenesis(addrs)
	bc.setupValidators(ips, keys, bc.opts...)

	return nil
}

func (bc *blockchain) connectAll(strong bool) error {
	for idx, v := range bc.validators {
		if strong {
			for _, vv := range bc.validators {
				if v.ContainerID() != vv.ContainerID() {
					if err := v.AddPeer(vv.NodeAddress()); err != nil {
						return err
					}
				}
			}
		} else {
			nextValidator := bc.validators[(idx+1)%len(bc.validators)]
			if err := v.AddPeer(nextValidator.NodeAddress()); err != nil {
				return err
			}
		}
	}
	return nil
}

func (bc *blockchain) setupGenesis(addrs []common.Address) {
	if bc.genesisFile == "" {
		bc.genesisFile = genesis.NewFile(bc.isQuorum,
			genesis.Validators(addrs...),
		)
	}
}

func (bc *blockchain) setupValidators(ips []net.IP, keys []*ecdsa.PrivateKey, options ...Option) {
	for i := 0; i < len(keys); i++ {
		var opts []Option
		opts = append(opts, options...)

		// Host data directory
		dataDir, err := generateRandomDir()
		if err != nil {
			log.Fatal("Failed to create data dir", err)
		}
		opts = append(opts, HostDataDir(dataDir))
		opts = append(opts, HostWebSocketPort(freeport.GetPort()))
		opts = append(opts, Key(keys[i]))
		opts = append(opts, HostIP(ips[i]))
		// Add PRIVATE_CONFIG for quorum
		if bc.isQuorum {
			ct := bc.constellationNetwork.GetConstellation(i)
			env := fmt.Sprintf("PRIVATE_CONFIG=%s", ct.ConfigPath())
			opts = append(opts, DockerEnv([]string{env}))
			opts = append(opts, DockerBinds(ct.Binds()))
		}

		geth := NewEthereum(
			bc.dockerClient,
			opts...,
		)

		err = geth.Init(bc.genesisFile)
		if err != nil {
			log.Fatal("Failed to init genesis", err)
		}

		bc.validators = append(bc.validators, geth)
	}
}

func (bc *blockchain) start(validators []Ethereum) error {
	for _, v := range validators {
		if err := v.Start(); err != nil {
			return err
		}
	}
	return nil
}

func (bc *blockchain) stop(validators []Ethereum, force bool) error {
	for _, v := range validators {
		if err := v.Stop(); err != nil && !force {
			return err
		}
	}
	return nil
}

// Constellation functions ----------------------------------------------------------------------------
type ConstellationNetwork interface {
	Start() error
	Stop() error
	Finalize()
	NumOfConstellations() int
	GetConstellation(int) Constellation
}

func NewConstellationNetwork(network *DockerNetwork, numOfValidators int, options ...ConstellationOption) (ctn *constellationNetwork) {
	ctn = &constellationNetwork{opts: options}

	var err error
	ctn.dockerClient, err = client.NewEnvClient()
	if err != nil {
		log.Fatalf("Cannot connect to Docker daemon, err: %v", err)
	}

	if network == nil {
		log.Fatalf("Network is required")
	}

	ctn.dockerNetworkName = network.Name()
	ctn.getFreeIPAddrs = network.GetFreeIPAddrs
	ctn.opts = append(ctn.opts, CTDockerNetworkName(ctn.dockerNetworkName))

	ctn.setupConstellations(numOfValidators)
	return ctn
}

func NewDefaultConstellationNetwork(network *DockerNetwork, numOfValidators int) (ctn *constellationNetwork) {
	return NewConstellationNetwork(network, numOfValidators,
		CTImageRepository("quay.io/amis/constellation"),
		CTImageTag("latest"),
		CTWorkDir("/ctdata"),
		CTLogging(false),
		CTKeyName("node"),
		CTSocketFilename("node.ipc"),
		CTVerbosity(1),
	)
}

func (ctn *constellationNetwork) setupConstellations(numOfValidators int) {
	// Create constellations
	ips, ports := ctn.getFreeHosts(numOfValidators)
	for i := 0; i < numOfValidators; i++ {
		opts := append(ctn.opts, CTHost(ips[i], ports[i]))
		othernodes := ctn.getOtherNodes(ips, ports, i)
		opts = append(opts, CTOtherNodes(othernodes))
		ct := NewConstellation(ctn.dockerClient, opts...)
		// Generate keys
		ct.GenerateKey()
		ctn.constellations = append(ctn.constellations, ct)
	}
}

func (ctn *constellationNetwork) Start() error {
	// Run nodes
	for i, ct := range ctn.constellations {
		err := ct.Start()
		if err != nil {
			log.Fatalf("Failed to start constellation %v, err:%v\n", i, err)
			return err
		}
	}
	return nil
}

func (ctn *constellationNetwork) Stop() error {
	// Stop nodes
	for i, ct := range ctn.constellations {
		err := ct.Stop()
		if err != nil {
			log.Fatalf("Failed to stop constellation %v, err:%v\n", i, err)
			return err
		}
	}
	return nil
}

func (ctn *constellationNetwork) Finalize() {
	// Clean up local working directory
	for _, ct := range ctn.constellations {
		os.RemoveAll(ct.WorkDir())
	}
}

func (ctn *constellationNetwork) NumOfConstellations() int {
	return len(ctn.constellations)
}

func (ctn *constellationNetwork) GetConstellation(idx int) Constellation {
	return ctn.constellations[idx]
}

func (ctn *constellationNetwork) getFreeHosts(num int) ([]net.IP, []int) {
	ips, err := ctn.getFreeIPAddrs(num)
	if err != nil {
		log.Fatalf("Cannot get free ip, err: %v", err)
	}
	var ports []int
	for i := 0; i < num; i++ {
		ports = append(ports, freeport.GetPort())
	}
	return ips, ports
}

func (ctn *constellationNetwork) getOtherNodes(ips []net.IP, ports []int, idx int) []string {
	var result []string
	for i, ip := range ips {
		if i == idx {
			continue
		}
		result = append(result, fmt.Sprintf("http://%s:%d/", ip, ports[i]))
	}
	return result
}

type constellationNetwork struct {
	dockerClient      *client.Client
	dockerNetworkName string
	getFreeIPAddrs    func(int) ([]net.IP, error)
	opts              []ConstellationOption
	constellations    []Constellation
}

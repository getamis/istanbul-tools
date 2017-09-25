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

package k8s

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"math/big"
	"strings"
	"sync"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"

	"github.com/getamis/go-ethereum/crypto"
	"github.com/getamis/istanbul-tools/charts"
	"github.com/getamis/istanbul-tools/client"
	istcommon "github.com/getamis/istanbul-tools/common"
	"github.com/getamis/istanbul-tools/container"
)

func NewEthereum(options ...Option) *ethereum {
	eth := &ethereum{
		name: istcommon.RandomHex(),
	}

	for _, opt := range options {
		opt(eth)
	}

	var err error
	eth.key, err = crypto.HexToECDSA(eth.nodekey)
	if err != nil {
		log.Error("Failed to create private key from nodekey", "nodekey", eth.nodekey)
		return nil
	}
	eth.chart = charts.NewValidatorChart(eth.name, eth.args)

	return eth
}

type ethereum struct {
	chart *charts.ValidatorChart
	name  string
	args  []string

	nodekey   string
	key       *ecdsa.PrivateKey
	accounts  []*ecdsa.PrivateKey
	k8sClient *kubernetes.Clientset
}

func (eth *ethereum) Init(genesisFile string) error {
	return nil
}

func (eth *ethereum) Start() error {
	if err := eth.chart.Install(false); err != nil {
		return err
	}

	eth.k8sClient = k8sClient(eth.chart.Name() + "-0")
	return nil
}

func (eth *ethereum) Stop() error {
	return eth.chart.Uninstall()
}

func (eth *ethereum) Wait(t time.Duration) error {
	return nil
}

func (eth *ethereum) Running() bool {
	return false
}

func (eth *ethereum) ContainerID() string {
	return ""
}

func (eth *ethereum) DockerEnv() []string {
	return nil
}

func (eth *ethereum) DockerBinds() []string {
	return nil
}

func (eth *ethereum) NewClient() client.Client {
	for i := 0; i < healthCheckRetryCount; i++ {
		client, err := client.Dial("ws://" + eth.Host() + ":8546")
		if err != nil {
			log.Warn("Failed to create client", "err", err)
			<-time.After(healthCheckRetryDelay)
			continue
		} else {
			return client
		}
	}

	return nil
}

func (eth *ethereum) NodeAddress() string {
	return ""
}

func (eth *ethereum) Address() common.Address {
	return crypto.PubkeyToAddress(eth.key.PublicKey)
}

func (eth *ethereum) ConsensusMonitor(errCh chan<- error, quit chan struct{}) {

}

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
			if container.GetProposer(head) == expectedAddress {
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
			return container.ErrNoBlock
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
				return container.ErrTimeout
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
	return nil
}

func (eth *ethereum) StartMining() error {
	return nil
}

func (eth *ethereum) StopMining() error {
	return nil
}

func (eth *ethereum) Accounts() (addrs []common.Address) {
	for _, acc := range eth.accounts {
		addrs = append(addrs, crypto.PubkeyToAddress(acc.PublicKey))
	}
	return addrs
}

// ----------------------------------------------------------------------------

func (eth *ethereum) Host() string {
	index := strings.LastIndex(eth.chart.Name(), "-")
	if index < 0 {
		log.Error("Invalid validator pod name")
		return ""
	}
	name := "validator-svc-" + eth.chart.Name()[index+1:]
	svc, err := eth.k8sClient.CoreV1().Services(defaultNamespace).Get(name, metav1.GetOptions{})
	if err != nil {
		log.Error("Failed to find service", "svc", name, "err", err)
		return ""
	}
	return svc.Status.LoadBalancer.Ingress[0].IP
}

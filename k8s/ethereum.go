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
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common"

	"github.com/getamis/istanbul-tools/charts"
	"github.com/getamis/istanbul-tools/client"
	istcommon "github.com/getamis/istanbul-tools/common"
)

func NewEthereum(options ...Option) *ethereum {
	eth := &ethereum{
		name: istcommon.RandomHex(),
	}

	for _, opt := range options {
		opt(eth)
	}

	eth.chart = charts.NewValidatorChart(eth.name, eth.args)

	return eth
}

type ethereum struct {
	chart *charts.ValidatorChart
	name  string
	args  []string

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

func (eth *ethereum) NewClient() *client.Client {
	client, err := client.Dial("ws://" + eth.Host() + ":8545")
	if err != nil {
		return nil
	}
	return client
}

func (eth *ethereum) NodeAddress() string {
	return ""
}

func (eth *ethereum) Address() common.Address {
	return common.Address{}
}

func (eth *ethereum) ConsensusMonitor(errCh chan<- error, quit chan struct{}) {

}

func (eth *ethereum) WaitForProposed(expectedAddress common.Address, timeout time.Duration) error {
	return nil
}

func (eth *ethereum) WaitForPeersConnected(expectedPeercount int) error {
	return nil
}

func (eth *ethereum) WaitForBlocks(num int, waitingTime ...time.Duration) error {
	return nil
}

func (eth *ethereum) WaitForBlockHeight(num int) error {
	return nil
}

func (eth *ethereum) WaitForNoBlocks(num int, duration time.Duration) error {
	return nil
}

func (eth *ethereum) AddPeer(address string) error {
	return nil
}

func (eth *ethereum) StartMining() error {
	return nil
}

func (eth *ethereum) StopMining() error {
	return nil
}

func (eth *ethereum) Accounts() []accounts.Account {
	return nil
}

// ----------------------------------------------------------------------------

func (eth *ethereum) Host() string {
	svc, err := eth.k8sClient.CoreV1().Services(defaultNamespace).Get(eth.chart.Name()+"-0", metav1.GetOptions{})
	if err != nil {
		return ""
	}
	return svc.Spec.LoadBalancerIP
}

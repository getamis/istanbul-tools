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
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/getamis/istanbul-tools/charts"
	istcommon "github.com/getamis/istanbul-tools/common"
	"github.com/getamis/istanbul-tools/container"
)

func NewBlockchain(numOfValidators int, gaslimit uint64, options ...Option) (bc *blockchain) {
	_, nodekeys, addrs := istcommon.GenerateKeys(numOfValidators)
	ips := istcommon.GenerateIPs(len(nodekeys))

	bc = &blockchain{
		genesis:     charts.NewGenesisChart(addrs, uint64(gaslimit)),
		staticNodes: charts.NewStaticNodesChart(nodekeys, ips),
	}

	if err := bc.genesis.Install(false); err != nil {
		log.Println(err)
		return nil
	}
	if err := bc.staticNodes.Install(false); err != nil {
		log.Println(err)
		bc.genesis.Uninstall()
		return nil
	}
	bc.setupValidators(numOfValidators, nodekeys, ips, options...)
	return bc
}

// ----------------------------------------------------------------------------

type blockchain struct {
	genesis     *charts.GenesisChart
	staticNodes *charts.StaticNodesChart
	validators  []container.Ethereum
}

func (bc *blockchain) EnsureConsensusWorking(geths []container.Ethereum, t time.Duration) error {
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

func (bc *blockchain) AddValidators(numOfValidators int) ([]container.Ethereum, error) {
	return nil, errors.New("Unsupported")
}

func (bc *blockchain) RemoveValidators(candidates []container.Ethereum, processingTime time.Duration) error {
	return errors.New("Unsupported")
}

func (bc *blockchain) Start(strong bool) error {
	return bc.start(bc.validators)
}

func (bc *blockchain) Stop(force bool) error {
	return bc.stop(bc.validators, force)
}

func (bc *blockchain) Finalize() {
	for _, v := range bc.validators {
		v.Stop()
	}

	bc.staticNodes.Uninstall()
	bc.genesis.Uninstall()
}

func (bc *blockchain) Validators() []container.Ethereum {
	return bc.validators
}

func (bc *blockchain) CreateNodes(num int, options ...Option) (nodes []container.Ethereum, err error) {
	return nil, errors.New("Unsupported")
}

// ----------------------------------------------------------------------------

func (bc *blockchain) setupValidators(num int, nodekeys []string, ips []string, options ...Option) {
	for i := 0; i < num; i++ {
		var opts []Option
		opts = append(opts, options...)

		opts = append(opts, Name(fmt.Sprintf("%d", i)))
		opts = append(opts, NodeKeyHex(nodekeys[i]))
		opts = append(opts, IPAddress(ips[i]))

		geth := NewEthereum(
			opts...,
		)

		bc.validators = append(bc.validators, geth)
	}
}

func (bc *blockchain) start(validators []container.Ethereum) error {
	var fns []func() error

	for _, v := range validators {
		geth := v
		fns = append(fns, func() error {
			return geth.Start()
		})
	}
	return executeInParallel(fns...)
}

func (bc *blockchain) stop(validators []container.Ethereum, force bool) error {
	var fns []func() error

	for _, v := range validators {
		geth := v
		fns = append(fns, func() error {
			return geth.Stop()
		})
	}
	return executeInParallel(fns...)
}

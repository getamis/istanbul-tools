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
	"crypto/ecdsa"
	"errors"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/common"

	"github.com/getamis/istanbul-tools/charts"
	istcommon "github.com/getamis/istanbul-tools/common"
	"github.com/getamis/istanbul-tools/container"
)

func NewBlockchain(numOfValidators int, numOfExtraAccounts int, gaslimit uint64, isQourum bool, options ...Option) (bc *blockchain) {
	_, nodekeys, addrs := istcommon.GenerateKeys(numOfValidators)
	ips := istcommon.GenerateIPs(len(nodekeys))

	extraKeys := make([][]*ecdsa.PrivateKey, numOfValidators)
	extraAddrs := make([][]common.Address, numOfValidators)

	var allocAddrs []common.Address
	if numOfExtraAccounts > 0 {
		for i := 0; i < numOfValidators; i++ {
			extraKeys[i], _, extraAddrs[i] = istcommon.GenerateKeys(numOfExtraAccounts)
			allocAddrs = append(allocAddrs, extraAddrs[i]...)
		}
	}

	bc = &blockchain{
		genesis:     charts.NewGenesisChart(addrs, allocAddrs, uint64(gaslimit), isQourum),
		staticNodes: charts.NewStaticNodesChart(nodekeys, ips),
	}

	if err := bc.genesis.Install(false); err != nil {
		log.Error("Failed to install genesis chart", "err", err)
		return nil
	}
	if err := bc.staticNodes.Install(false); err != nil {
		log.Error("Failed to install static nodes chart", "err", err)
		bc.genesis.Uninstall()
		return nil
	}
	bc.setupValidators(numOfValidators, extraKeys, nodekeys, ips, options...)
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
	return nil, errors.New("unsupported")
}

func (bc *blockchain) RemoveValidators(candidates []container.Ethereum, processingTime time.Duration) error {
	return errors.New("unsupported")
}

func (bc *blockchain) Start(strong bool) error {
	return bc.start(bc.validators)
}

func (bc *blockchain) Stop(force bool) error {
	return bc.stop(bc.validators, force)
}

func (bc *blockchain) Finalize() {
	bc.staticNodes.Uninstall()
	bc.genesis.Uninstall()
}

func (bc *blockchain) Validators() []container.Ethereum {
	return bc.validators
}

// ----------------------------------------------------------------------------

func (bc *blockchain) setupValidators(num int, extraKeys [][]*ecdsa.PrivateKey, nodekeys []string, ips []string, options ...Option) {
	for i := 0; i < num; i++ {
		var opts []Option
		opts = append(opts, options...)

		opts = append(opts, Name(fmt.Sprintf("%d", i)))
		opts = append(opts, NodeKeyHex(nodekeys[i]))
		opts = append(opts, IPAddress(ips[i]))
		opts = append(opts, ExtraAccounts(extraKeys[i]))

		geth := NewEthereum(
			opts...,
		)

		if geth != nil {
			bc.validators = append(bc.validators, geth)
		}
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

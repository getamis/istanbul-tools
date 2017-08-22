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
	"log"
	"math"
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
	RemoveValidators(candidates []Ethereum) error
	EnsureConsensusWorking(geths []Ethereum, t time.Duration) error
	Start(bool) error
	Stop(bool) error
	Validators() []Ethereum
	Finalize()
}

func NewBlockchain(numOfValidators int, options ...Option) (bc *blockchain) {
	bc = &blockchain{opts: options}

	var err error
	bc.dockerClient, err = client.NewEnvClient()
	if err != nil {
		log.Fatalf("Cannot connect to Docker daemon, err: %v", err)
	}

	bc.addValidators(numOfValidators)
	return bc
}

// ----------------------------------------------------------------------------

type blockchain struct {
	dockerClient *client.Client
	genesisFile  string
	validators   []Ethereum
	opts         []Option
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

func (bc *blockchain) RemoveValidators(candidates []Ethereum) error {
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
	<-time.After(time.Duration(math.Pow(2, float64(len(candidates)))*5) * time.Second)
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
	return bc.stop(bc.validators, force)
}

func (bc *blockchain) Finalize() {
	os.RemoveAll(filepath.Dir(bc.genesisFile))
}

func (bc *blockchain) Validators() []Ethereum {
	return bc.validators
}

// ----------------------------------------------------------------------------

func (bc *blockchain) addValidators(numOfValidators int) error {
	keys, addrs := generateKeys(numOfValidators)
	bc.setupGenesis(addrs)
	bc.setupValidators(keys, bc.opts...)

	return nil
}

func (bc *blockchain) connectAll(strong bool) error {
	for i, v := range bc.validators {
		istClient := v.NewIstanbulClient()
		for j, v := range bc.validators {
			if (strong && j > i) || (!strong && j == i+1) {
				err := istClient.AddPeer(context.Background(), v.NodeAddress())
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (bc *blockchain) setupGenesis(addrs []common.Address) {
	if bc.genesisFile == "" {
		setupDir, err := generateRandomDir()
		if err != nil {
			log.Fatal("Failed to create setup dir", err)
		}
		err = genesis.Save(setupDir, genesis.New(addrs))
		if err != nil {
			log.Fatal("Failed to save genesis", err)
		}
		bc.genesisFile = filepath.Join(setupDir, genesis.FileName)
	}
}

func (bc *blockchain) setupValidators(keys []*ecdsa.PrivateKey, options ...Option) {
	for i := 0; i < len(keys); i++ {
		var opts []Option
		opts = append(opts, options...)

		// Host data directory
		dataDir, err := generateRandomDir()
		if err != nil {
			log.Fatal("Failed to create data dir", err)
		}
		opts = append(opts, HostDataDir(dataDir))
		opts = append(opts, HostPort(freeport.GetPort()))
		opts = append(opts, HostWebSocketPort(freeport.GetPort()))
		opts = append(opts, Key(keys[i]))

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

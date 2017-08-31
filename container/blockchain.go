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

func NewDefaultBlockchain(numOfValidators int) (bc *blockchain) {
	return NewBlockchain(numOfValidators,
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

func NewDefaultBlockchainWithFaulty(numOfNormal int, numOfFaulty int) (bc *blockchain) {
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

	keys, addrs := generateKeys(numOfNormal + numOfFaulty)
	bc.setupGenesis(addrs)
	// Create normal validators
	bc.opts = normalOpts
	bc.setupValidators(keys[:numOfNormal], bc.opts...)
	// Create faulty validators
	bc.opts = faultyOpts
	bc.setupValidators(keys[numOfNormal:], bc.opts...)
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
	return bc.stop(bc.validators, force)
}

func (bc *blockchain) Finalize() {
	os.RemoveAll(filepath.Dir(bc.genesisFile))
}

func (bc *blockchain) Validators() []Ethereum {
	return bc.validators
}

func (bc *blockchain) CreateNodes(num int, options ...Option) (nodes []Ethereum, err error) {
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
	keys, addrs := generateKeys(numOfValidators)
	bc.setupGenesis(addrs)
	bc.setupValidators(keys, bc.opts...)

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
		bc.genesisFile = genesis.NewFile(
			genesis.Validators(addrs...),
		)
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

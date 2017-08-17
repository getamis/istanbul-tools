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
	"crypto/ecdsa"
	"log"
	"os"
	"path/filepath"

	"github.com/docker/docker/client"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/phayes/freeport"

	"github.com/getamis/istanbul-tools/genesis"
)

type Blockchain interface {
	Start() error
	Stop() error
	Validators() []Ethereum
	Finalize()
}

func NewBlockchain(numOfValidators int, options ...Option) (bc *blockchain) {
	var keys []*ecdsa.PrivateKey
	var addrs []common.Address

	bc = &blockchain{}

	for i := 0; i < numOfValidators; i++ {
		key, err := crypto.GenerateKey()
		if err != nil {
			log.Fatalf("couldn't generate key: " + err.Error())
		}
		keys = append(keys, key)

		addr := crypto.PubkeyToAddress(key.PublicKey)
		addrs = append(addrs, addr)
	}

	setupDir, err := generateRandomDir()
	if err != nil {
		log.Fatal("Failed to create setup dir", err)
	}
	err = genesis.Save(setupDir, genesis.New(addrs))
	if err != nil {
		log.Fatal("Failed to save genesis", err)
	}
	bc.genesisFile = filepath.Join(setupDir, genesis.FileName)

	dockerClient, err := client.NewEnvClient()
	if err != nil {
		log.Fatalf("Cannot connect to Docker daemon, err: %v", err)
	}

	for i := 0; i < numOfValidators; i++ {
		opts := make([]Option, len(options))
		copy(opts, options)

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
			dockerClient,
			opts...,
		)

		err = geth.Init(bc.genesisFile)
		if err != nil {
			log.Fatal("Failed to init genesis", err)
		}

		bc.validators = append(bc.validators, geth)
	}

	return bc
}

// ----------------------------------------------------------------------------

type blockchain struct {
	genesisFile string
	validators  []Ethereum
}

func (bc *blockchain) Start() error {
	for _, v := range bc.validators {
		if err := v.Start(); err != nil {
			return err
		}
	}

	return nil
}

func (bc *blockchain) Stop() error {
	for _, v := range bc.validators {
		if err := v.Stop(); err != nil {
			return err
		}
	}

	return nil
}

func (bc *blockchain) Finalize() {
	os.RemoveAll(filepath.Dir(bc.genesisFile))
}

func (bc *blockchain) Validators() []Ethereum {
	return bc.validators
}

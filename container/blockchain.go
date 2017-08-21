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
	"math/rand"
	"os"
	"path/filepath"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/ethereum/go-ethereum/common"
	"github.com/phayes/freeport"
	uuid "github.com/satori/go.uuid"

	"github.com/getamis/istanbul-tools/genesis"
)

type Blockchain interface {
	Start() error
	Stop() error
	Validators() []Ethereum
	Finalize()
}

func NewBlockchain(numOfValidators int, options ...Option) (bc *blockchain) {
	bc = &blockchain{}

	var err error
	bc.dockerClient, err = client.NewEnvClient()
	if err != nil {
		log.Fatalf("Cannot connect to Docker daemon, err: %v", err)
	}

	bc.setupNetwork()
	keys, addrs := generateKeys(numOfValidators)
	bc.setupGenesis(addrs)
	bc.setupValidators(keys, options...)

	return bc
}

// ----------------------------------------------------------------------------

type blockchain struct {
	dockerClient    *client.Client
	dockerNetworkID string
	netClass        string
	genesisFile     string
	validators      []Ethereum
}

func (bc *blockchain) Start() error {
	for i, v := range bc.validators {
		if err := v.Start(); err != nil {
			return err
		}

		if err := bc.dockerClient.NetworkConnect(context.Background(), bc.dockerNetworkID, v.ContainerID(), &network.EndpointSettings{
			IPAddress: fmt.Sprintf(bc.netClass+".%d", i+1),
		}); err != nil {
			log.Printf("Failed to connect to network '%s', %v", bc.dockerNetworkID, err)
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

	bc.dockerClient.NetworkRemove(context.Background(), bc.dockerNetworkID)
}

func (bc *blockchain) Validators() []Ethereum {
	return bc.validators
}

// ----------------------------------------------------------------------------

func (bc *blockchain) setupNetwork() {
	name := "net" + uuid.NewV4().String()
	bc.netClass = fmt.Sprintf("172.16.%d", rand.Uint32()%255)
	resp, err := bc.dockerClient.NetworkCreate(context.Background(), name, types.NetworkCreate{
		IPAM: &network.IPAM{
			Config: []network.IPAMConfig{
				network.IPAMConfig{
					Subnet: bc.netClass + ".0/24",
				},
			},
		},
	})

	if err != nil {
		log.Fatal("Failed to setup blockchain network,", err)
	}

	bc.dockerNetworkID = resp.ID
}

func (bc *blockchain) setupGenesis(addrs []common.Address) {
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

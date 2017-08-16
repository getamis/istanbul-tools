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

package core

import (
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/docker/docker/client"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/consensus/istanbul"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/p2p/discover"
	"github.com/ethereum/go-ethereum/params"
	"github.com/getamis/istanbul-tools/cmd/istanbul/extradata"
	"github.com/phayes/freeport"
	"github.com/satori/go.uuid"
)

const (
	defaultLocalDir   = "/tmp/gdata"
	datadirPrivateKey = "nodekey"

	clientIdentifier = "geth"
	staticNodeJson   = "static-nodes.json"

	GenesisFile = "genesis.json"
)

type Env struct {
	GethID  int
	P2PPort uint16
	RpcPort uint16
	DataDir string
	Key     *ecdsa.PrivateKey
	Client  *client.Client
}

func Teardown(envs []*Env) {
	for _, env := range envs {
		os.RemoveAll(env.DataDir)
	}
}

func SetupEnv(numbers int) []*Env {
	envs := make([]*Env, numbers)
	rpcPort := uint16(freeport.GetPort())
	p2pPort := uint16(freeport.GetPort())

	for i := 0; i < len(envs); i++ {
		client, err := client.NewEnvClient()
		if err != nil {
			log.Fatalf("Cannot connect to Docker daemon, err: %v", err)
		}

		key, err := crypto.GenerateKey()
		if err != nil {
			log.Fatalf("couldn't generate key: " + err.Error())
		}

		envs[i] = &Env{
			GethID:  i,
			P2PPort: p2pPort,
			RpcPort: rpcPort,
			DataDir: filepath.Join(defaultLocalDir, fmt.Sprintf("%s%s", clientIdentifier, uuid.NewV4().String())),
			Key:     key,
			Client:  client,
		}

		rpcPort = uint16(freeport.GetPort())
		p2pPort = uint16(freeport.GetPort())
	}
	return envs
}

func SetupNodes(envs []*Env, g *core.Genesis) error {
	nodes := toStaticNodes(envs)

	for _, env := range envs {
		if err := saveNodeKey(env.DataDir, env.Key); err != nil {
			log.Fatalf("Failed to save node key")
		}

		if err := saveStaticNode(env.DataDir, nodes); err != nil {
			return err
		}

		if err := saveGenesis(env.DataDir, g); err != nil {
			return err
		}
	}
	return nil
}

func NewGenesis(envs []*Env) *core.Genesis {
	extraData, err := extradata.Encode("0x00", toAddress(envs))
	if err != nil {
		log.Fatalf("Failed to generate genesis, err:%s", err)
	}

	return &core.Genesis{
		Timestamp:  uint64(time.Now().Unix()),
		GasLimit:   4700000,
		Difficulty: big.NewInt(1),
		Alloc:      make(core.GenesisAlloc),
		Config: &params.ChainConfig{
			HomesteadBlock: big.NewInt(1),
			EIP150Block:    big.NewInt(2),
			EIP155Block:    big.NewInt(3),
			EIP158Block:    big.NewInt(3),
			Istanbul: &params.IstanbulConfig{
				ProposerPolicy: uint64(istanbul.DefaultConfig.ProposerPolicy),
				Epoch:          istanbul.DefaultConfig.Epoch,
			},
		},
		Mixhash:   types.IstanbulDigest,
		ExtraData: hexutil.MustDecode(extraData),
	}
}

func saveNodeKey(dataDir string, key *ecdsa.PrivateKey) error {
	err := os.MkdirAll(dataDir, 0700)
	if err != nil {
		log.Fatal(err)
	}

	keyDir := filepath.Join(dataDir, clientIdentifier)
	if err := os.MkdirAll(keyDir, 0700); err != nil {
		log.Println(fmt.Sprintf("Failed to create key dir: %v", err))
		return err
	}

	keyfile := filepath.Join(keyDir, datadirPrivateKey)
	if err := crypto.SaveECDSA(keyfile, key); err != nil {
		log.Println(fmt.Sprintf("Failed to persist node key: %v", err))
		return err
	}
	return nil
}

func saveStaticNode(dataDir string, nodes []string) error {
	filePath := filepath.Join(dataDir, clientIdentifier)
	keyPath := filepath.Join(filePath, staticNodeJson)

	raw, err := json.Marshal(nodes)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(keyPath, raw, 0600)
}

func saveGenesis(dataDir string, genesis *core.Genesis) error {
	filePath := filepath.Join(dataDir, GenesisFile)

	raw, err := json.Marshal(genesis)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(filePath, raw, 0600)
}

func toStaticNodes(envs []*Env) []string {
	nodes := make([]string, len(envs))

	for i, env := range envs {
		daemonHost := env.Client.DaemonHost()
		url, err := url.Parse(daemonHost)
		if err != nil {
			log.Fatalf("Failed to parse daemon host, err: %v", err)
		}

		var host string
		if url.Scheme == "unix" {
			host = "127.0.0.1"
		} else {
			host, _, err = net.SplitHostPort(url.Host)
			if err != nil {
				log.Fatalf("Failed to split host and port, err: %v", err)
			}
		}

		nodeID := discover.PubkeyID(&env.Key.PublicKey)
		nodes[i] = discover.NewNode(nodeID, net.ParseIP(host), 0, env.P2PPort).String()
	}
	return nodes
}

func toAddress(envs []*Env) []common.Address {
	addrs := make([]common.Address, len(envs))

	for i, env := range envs {
		addrs[i] = crypto.PubkeyToAddress(env.Key.PublicKey)
	}
	return addrs
}

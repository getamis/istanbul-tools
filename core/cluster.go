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
	"net"
	"os"
	"path/filepath"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/p2p/discover"
	"github.com/satori/go.uuid"
)

const (
	defaultBaseRpcPort = uint16(8545)
	defaultHttpPort    = uint16(30303)

	defaultLocalDir   = "/tmp/gdata"
	datadirPrivateKey = "nodekey"

	clientIdentifier = "geth"
	staticNodeJson   = "static-nodes.json"
	GenesisJson      = "genesis.json"
)

var (
	defaultIP = net.IPv4(127, 0, 0, 1)
)

func GenerateClusterKeys(numbers int) []*ecdsa.PrivateKey {
	keys := make([]*ecdsa.PrivateKey, numbers)
	for i := 0; i < len(keys); i++ {
		key, err := crypto.GenerateKey()
		if err != nil {
			panic("couldn't generate key: " + err.Error())
		}
		keys[i] = key
	}
	return keys
}

type Env struct {
	GethID   int
	HttpPort uint16
	RpcPort  uint16
	DataDir  string
	Key      *ecdsa.PrivateKey
}

func Teardown(envs []*Env) {
	for _, env := range envs {
		os.RemoveAll(env.DataDir)
	}
}

func SetupEnv(prvKeys []*ecdsa.PrivateKey) []*Env {
	envs := make([]*Env, len(prvKeys))
	rpcPort := defaultBaseRpcPort
	httpPort := defaultHttpPort

	for i := 0; i < len(envs); i++ {
		dataDir, err := saveNodeKey(prvKeys[i])
		if err != nil {
			panic("Failed to save node key")
		}

		envs[i] = &Env{
			GethID:   i,
			HttpPort: httpPort,
			RpcPort:  rpcPort,
			DataDir:  dataDir,
			Key:      prvKeys[i],
		}

		rpcPort = rpcPort + 1
		httpPort = httpPort + 1
	}
	return envs
}

func SetupNodes(envs []*Env) error {
	nodes := transformToStaticNodes(envs)
	for _, env := range envs {
		if err := saveStaticNode(env.DataDir, nodes); err != nil {
			return err
		}
	}

	addrs := transformToAddress(envs)
	genesis := GenerateGenesis(addrs)
	for _, env := range envs {
		if err := saveGenesis(env.DataDir, genesis); err != nil {
			return err
		}
	}
	return nil
}

func saveNodeKey(key *ecdsa.PrivateKey) (string, error) {
	err := os.MkdirAll(filepath.Join(defaultLocalDir), 0700)
	if err != nil {
		log.Fatal(err)
	}

	instanceDir := filepath.Join(defaultLocalDir, fmt.Sprintf("%s%s", clientIdentifier, uuid.NewV4().String()))
	if err := os.MkdirAll(instanceDir, 0700); err != nil {
		log.Println(fmt.Sprintf("Failed to create instance dir: %v", err))
		return "", err
	}

	keyDir := filepath.Join(instanceDir, clientIdentifier)
	if err := os.MkdirAll(keyDir, 0700); err != nil {
		log.Println(fmt.Sprintf("Failed to create key dir: %v", err))
		return "", err
	}

	keyfile := filepath.Join(keyDir, datadirPrivateKey)
	if err := crypto.SaveECDSA(keyfile, key); err != nil {
		log.Println(fmt.Sprintf("Failed to persist node key: %v", err))
		return "", err
	}
	return instanceDir, nil
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

func transformToStaticNodes(envs []*Env) []string {
	nodes := make([]string, len(envs))

	for i, env := range envs {
		nodeID := discover.PubkeyID(&env.Key.PublicKey)
		nodes[i] = discover.NewNode(nodeID, defaultIP, 0, env.HttpPort).String()
	}
	return nodes
}

func transformToAddress(envs []*Env) []common.Address {
	addrs := make([]common.Address, len(envs))

	for i, env := range envs {
		addrs[i] = crypto.PubkeyToAddress(env.Key.PublicKey)
	}
	return addrs
}

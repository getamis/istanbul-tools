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

package common

import (
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/p2p/discv5"
	uuid "github.com/satori/go.uuid"
)

const (
	defaultLocalDir  = "/tmp/gdata"
	clientIdentifier = "geth"
	nodekeyFileName  = "nodekey"
)

func GenerateIPs(num int) (ips []string) {
	for i := 0; i < num; i++ {
		ips = append(ips, fmt.Sprintf("10.0.1.%d", i+2))
	}

	return ips
}

func GenerateRandomDir() (string, error) {
	err := os.MkdirAll(filepath.Join(defaultLocalDir), 0700)
	if err != nil {
		log.Error("Failed to create dir", "dir", defaultLocalDir, "err", err)
		return "", err
	}

	instanceDir := filepath.Join(defaultLocalDir, fmt.Sprintf("%s-%s", clientIdentifier, uuid.NewV4().String()))
	if err := os.MkdirAll(instanceDir, 0700); err != nil {
		log.Error("Failed to create dir", "dir", instanceDir, "err", err)
		return "", err
	}

	return instanceDir, nil
}

func GeneratePasswordFile(dir string, filename string, password string) {
	path := filepath.Join(dir, filename)
	err := ioutil.WriteFile(path, []byte(password), 0644)
	if err != nil {
		log.Error("Failed to generate password file", "file", path, "err", err)
		return
	}
}

func CopyKeystore(dir string, accounts []accounts.Account) {
	keystorePath := filepath.Join(dir, "keystore")
	err := os.MkdirAll(keystorePath, 0744)
	if err != nil {
		log.Error("Failed to copy keystore", "dir", keystorePath, "err", err)
		return
	}
	for _, a := range accounts {
		src := a.URL.Path
		dst := filepath.Join(keystorePath, filepath.Base(src))
		copyFile(src, dst)
	}
}

func GenerateKeys(num int) (keys []*ecdsa.PrivateKey, nodekeys []string, addrs []common.Address) {
	for i := 0; i < num; i++ {
		nodekey := RandomHex()[2:]
		nodekeys = append(nodekeys, nodekey)

		key, err := crypto.HexToECDSA(nodekey)
		if err != nil {
			log.Error("Failed to generate key", "err", err)
			return nil, nil, nil
		}
		keys = append(keys, key)

		addr := crypto.PubkeyToAddress(key.PublicKey)
		addrs = append(addrs, addr)
	}

	return keys, nodekeys, addrs
}

func SaveNodeKey(key *ecdsa.PrivateKey, dataDir string) error {
	keyDir := filepath.Join(dataDir, clientIdentifier)
	if err := os.MkdirAll(keyDir, 0700); err != nil {
		log.Error("Failed to create dir", "dir", keyDir, "err", err)
		return err
	}

	keyfile := filepath.Join(keyDir, nodekeyFileName)
	if err := crypto.SaveECDSA(keyfile, key); err != nil {
		log.Error("Failed to persist node key", "file", keyfile, "err", err)
		return err
	}
	return nil
}

func GenerateStaticNodesAt(dir string, nodekeys []string, ipAddrs []string) (filename string) {
	var nodes []string

	for i, nodekey := range nodekeys {
		key, err := crypto.HexToECDSA(nodekey)
		if err != nil {
			log.Error("Failed to create key from hex", "hex", nodekey, "err", err)
			return ""
		}
		node := discv5.NewNode(
			discv5.PubkeyID(&key.PublicKey),
			net.ParseIP(ipAddrs[i]),
			0,
			uint16(30303))

		nodes = append(nodes, node.String())
	}

	filename = filepath.Join(dir, "static-nodes.json")
	bytes, _ := json.Marshal(nodes)
	if err := ioutil.WriteFile(filename, bytes, 0644); err != nil {
		log.Error("Failed to write file", "file", filename, "err", err)
		return ""
	}

	return filename
}

func GenerateStaticNodes(nodekeys []string, ipAddrs []string) (filename string) {
	dir, _ := GenerateRandomDir()

	return GenerateStaticNodesAt(dir, nodekeys, ipAddrs)
}

func RandomHex() string {
	b, _ := RandomBytes(32)
	return common.BytesToHash(b).Hex()
}

func RandomBytes(len int) ([]byte, error) {
	b := make([]byte, len)
	_, _ = rand.Read(b)

	return b, nil
}

func copyFile(src string, dst string) {
	data, err := ioutil.ReadFile(src)
	if err != nil {
		log.Error("Failed to read file", "file", src, "err", err)
		return
	}
	err = ioutil.WriteFile(dst, data, 0644)
	if err != nil {
		log.Error("Failed to write file", "file", dst, "err", err)
		return
	}
}

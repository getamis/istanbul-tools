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
	"log"
	"net"
	"os"
	"path/filepath"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/getamis/go-ethereum/p2p/discover"
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
		return "", err
	}

	instanceDir := filepath.Join(defaultLocalDir, fmt.Sprintf("%s-%s", clientIdentifier, uuid.NewV4().String()))
	if err := os.MkdirAll(instanceDir, 0700); err != nil {
		log.Println(fmt.Sprintf("Failed to create instance dir: %v", err))
		return "", err
	}

	return instanceDir, nil
}

func GeneratePasswordFile(dir string, filename string, password string) {
	path := filepath.Join(dir, filename)
	err := ioutil.WriteFile(path, []byte(password), 0644)
	if err != nil {
		log.Fatalf("Failed to generate password file, err:%v", err)
	}
}

func CopyKeystore(dir string, accounts []accounts.Account) {
	keystorePath := filepath.Join(dir, "keystore")
	err := os.MkdirAll(keystorePath, 0744)
	if err != nil {
		log.Fatalf("Failed to copy keystore, err:%v", err)
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
			log.Fatalf("couldn't generate key: " + err.Error())
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
		log.Println(fmt.Sprintf("Failed to create key dir: %v", err))
		return err
	}

	keyfile := filepath.Join(keyDir, nodekeyFileName)
	if err := crypto.SaveECDSA(keyfile, key); err != nil {
		log.Println(fmt.Sprintf("Failed to persist node key: %v", err))
		return err
	}
	return nil
}

func GenerateStaticNodesAt(dir string, nodekeys []string, ipAddrs []string) (filename string) {
	var nodes []string

	for i, nodekey := range nodekeys {
		key, err := crypto.HexToECDSA(nodekey)
		if err != nil {
			log.Printf("Failed to create key, err: %v\n", err)
			return ""
		}
		node := discover.NewNode(
			discover.PubkeyID(&key.PublicKey),
			net.ParseIP(ipAddrs[i]),
			0,
			uint16(30303))

		nodes = append(nodes, node.String())
	}

	filename = filepath.Join(dir, "static-nodes.json")
	bytes, _ := json.Marshal(nodes)
	if err := ioutil.WriteFile(filename, bytes, 0644); err != nil {
		log.Printf("Failed to write '%s', err: %v\n", filename, err)
		return ""
	}

	return filename
}

func GenerateStaticNodes(nodekeys []string, ipAddrs []string) (filename string) {
	dir, err := GenerateRandomDir()
	if err != nil {
		log.Printf("Failed to generate directory, err: %v\n", err)
		return ""
	}

	return GenerateStaticNodesAt(dir, nodekeys, ipAddrs)
}

func RandomHex() string {
	b, _ := RandomBytes(32)
	return common.BytesToHash(b).Hex()
}

func RandomBytes(len int) ([]byte, error) {
	b := make([]byte, len)
	_, err := rand.Read(b)
	if err != nil {
		log.Fatalln(err)
	}

	return b, nil
}

func copyFile(src string, dst string) {
	data, err := ioutil.ReadFile(src)
	if err != nil {
		log.Fatal(err)
	}
	err = ioutil.WriteFile(dst, data, 0644)
	if err != nil {
		log.Fatal(err)
	}
}

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
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/consensus/istanbul"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/sha3"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/getamis/istanbul-tools/cmd/istanbul/extradata"
	uuid "github.com/satori/go.uuid"
)

const (
	defaultLocalDir  = "/tmp/gdata"
	clientIdentifier = "geth"
	nodekeyFileName  = "nodekey"
)

func generateRandomDir() (string, error) {
	err := os.MkdirAll(filepath.Join(defaultLocalDir), 0700)
	if err != nil {
		log.Fatal(err)
	}

	instanceDir := filepath.Join(defaultLocalDir, fmt.Sprintf("%s-%s", clientIdentifier, uuid.NewV4().String()))
	if err := os.MkdirAll(instanceDir, 0700); err != nil {
		log.Println(fmt.Sprintf("Failed to create instance dir: %v", err))
		return "", err
	}

	return instanceDir, nil
}

func generateKeys(num int) (keys []*ecdsa.PrivateKey, addrs []common.Address) {
	for i := 0; i < num; i++ {
		key, err := crypto.GenerateKey()
		if err != nil {
			log.Fatalf("couldn't generate key: " + err.Error())
		}
		keys = append(keys, key)

		addr := crypto.PubkeyToAddress(key.PublicKey)
		addrs = append(addrs, addr)
	}

	return keys, addrs
}

func saveNodeKey(key *ecdsa.PrivateKey, dataDir string) error {
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

func sigHash(header *types.Header) (hash common.Hash) {
	hasher := sha3.NewKeccak256()

	// Clean seal is required for calculating proposer seal.
	rlp.Encode(hasher, types.IstanbulFilteredHeader(header, false))
	hasher.Sum(hash[:0])
	return hash
}

func getProposer(header *types.Header) common.Address {
	if header == nil {
		return common.Address{}
	}

	_, istanbulExtra, err := extradata.Decode(common.ToHex(header.Extra))
	if err != nil {
		return common.Address{}
	}

	addr, err := istanbul.GetSignatureAddress(sigHash(header).Bytes(), istanbulExtra.Seal)
	if err != nil {
		return common.Address{}
	}
	return addr
}

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

package genesis

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"path/filepath"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/consensus/istanbul"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"

	"github.com/getamis/istanbul-tools/common"
)

const (
	FileName       = "genesis.json"
	InitGasLimit   = 4700000
	InitDifficulty = 1
)

func New(options ...Option) *core.Genesis {
	genesis := &core.Genesis{
		Timestamp:  uint64(time.Now().Unix()),
		GasLimit:   InitGasLimit,
		Difficulty: big.NewInt(InitDifficulty),
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
		Mixhash: types.IstanbulDigest,
	}

	for _, opt := range options {
		opt(genesis)
	}

	return genesis
}

func NewFileAt(dir string, isQuorum bool, options ...Option) string {
	genesis := New(options...)
	if err := Save(dir, genesis, isQuorum); err != nil {
		log.Fatalf("Failed to save genesis to '%s', err: %v", dir, err)
	}

	return filepath.Join(dir, FileName)
}

func NewFile(isQuorum bool, options ...Option) string {
	dir, err := common.GenerateRandomDir()
	if err != nil {
		log.Fatalf("Failed to create random directory, err: %v", err)
	}
	return NewFileAt(dir, isQuorum, options...)
}

func Save(dataDir string, genesis *core.Genesis, isQuorum bool) error {
	filePath := filepath.Join(dataDir, FileName)

	raw, err := json.Marshal(genesis)
	if err != nil {
		return err
	}

	//Quorum hack: add isQuorum field
	if isQuorum {
		jsonStr := string(raw)
		idx := strings.Index(jsonStr, ",\"istanbul\"")
		jsonStr = fmt.Sprintf("%s,\"isQuorum\":true%s", jsonStr[:idx], jsonStr[idx:])
		raw = []byte(jsonStr)
	}
	return ioutil.WriteFile(filePath, raw, 0600)
}

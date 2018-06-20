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

package reinit

import (
  "fmt"
  "strings"
  "sort"
  "bytes"
  "math/big"
  "encoding/json"
  "crypto/ecdsa"

  "github.com/ethereum/go-ethereum/common"
  "github.com/ethereum/go-ethereum/core"
  "github.com/ethereum/go-ethereum/core/types"
  "github.com/ethereum/go-ethereum/crypto"
  "github.com/ethereum/go-ethereum/rlp"
  "github.com/istanbul-tools/genesis"
	"github.com/urfave/cli"
)

var (
  ReinitCommand = cli.Command{
    Name:  "reinit",
    Action: reinit,
    Usage: "Reinitialize a genesis block using previous node info",
    ArgsUsage: "\"nodekey1,nodekey2,...\"",
    Flags: []cli.Flag{
      nodeKeyFlag,
      quorumFlag,
    },
    Description: `This tool helps generate a genesis block`,
  }
)

func reinit(ctx *cli.Context) error {
  if !ctx.IsSet(nodeKeyFlag.Name) {
    return cli.NewExitError("Must supply nodekeys", 10);
  }

  nodeKeyString := ctx.String(nodeKeyFlag.Name);
  nodekeys := strings.Split(nodeKeyString, ",");

  isQuorum := ctx.Bool(quorumFlag.Name);

  var stringAddrs []string;
  _, _, addr := generateKeysWithNodeKey(nodekeys);
  // Convert to String to sort
  for i := 0; i < len(addr); i++ {
    addrString, _ := json.Marshal(addr[i]);
    stringAddrs = append(stringAddrs, string(addrString));
  }
  sort.Strings(stringAddrs);

  // Convert back to address
  var addrs []common.Address;
  for i := 0; i < len(stringAddrs); i++ {
    var address common.Address;
    json.Unmarshal([]byte(stringAddrs[i]), &address);
    addrs = append(addrs, address);
  }
  // Generate Genesis block
  genesisString, _ := getGenesisWithAddrs(addrs, isQuorum);
  // genesisS, _ := json.Marshal(genesis);
  fmt.Println(string(genesisString));
  return nil;
}

func generateKeysWithNodeKey(nodekeysIn []string) (keys []*ecdsa.PrivateKey, nodekeys []string, addrs []common.Address) {
  for i := 0; i < len(nodekeysIn); i++ {
    nodekey := nodekeysIn[i]
    nodekeys = append(nodekeys, nodekey)

    key, err := crypto.HexToECDSA(nodekey)
    if err != nil {
      fmt.Println("Failed to generate key", "err", err)
      return nil, nil, nil
    }
    keys = append(keys, key)

    addr := crypto.PubkeyToAddress(key.PublicKey)
    addrs = append(addrs, addr)
  }
  return keys, nodekeys, addrs
}

func getGenesisWithAddrs(addrs []common.Address, isQuorum bool) ([]byte, error) {
  // generate genesis block
  istanbulGenesis := genesis.New(
    genesis.Validators(addrs...),
    genesis.Alloc(addrs, new(big.Int).Exp(big.NewInt(10), big.NewInt(50), nil)),
  )
  var jsonBytes []byte
  var err error
  if isQuorum {
    jsonBytes, err = json.MarshalIndent(genesis.ToQuorum(istanbulGenesis, true), "", "    ")
  } else {
    jsonBytes, err = json.MarshalIndent(istanbulGenesis, "", "    ")
  }
  return jsonBytes, err
}

func appendValidators(genesis *core.Genesis, addrs []common.Address) {

  if len(genesis.ExtraData) < types.IstanbulExtraVanity {
    genesis.ExtraData = append(genesis.ExtraData, bytes.Repeat([]byte{0x00}, types.IstanbulExtraVanity)...)
  }
  genesis.ExtraData = genesis.ExtraData[:types.IstanbulExtraVanity]

  ist := &types.IstanbulExtra{
    Validators:    addrs,
    Seal:          []byte{},
    CommittedSeal: [][]byte{},
  }

  istPayload, err := rlp.EncodeToBytes(&ist)
  if err != nil {
    panic("failed to encode istanbul extra")
  }
  genesis.ExtraData = append(genesis.ExtraData, istPayload...)
}

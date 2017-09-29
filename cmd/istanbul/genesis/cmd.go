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
	"math/big"
	"net"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/p2p/discover"
	"github.com/urfave/cli"

	istcommon "github.com/getamis/istanbul-tools/common"
	"github.com/getamis/istanbul-tools/genesis"
)

type validatorInfo struct {
	Address  common.Address
	Nodekey  string
	NodeInfo string
}

var (
	GenesisCommand = cli.Command{
		Name:   "genesis",
		Usage:  "Istanbul genesis block generator",
		Action: gen,
		Flags: []cli.Flag{
			numOfValidatorsFlag,
			staticNodesFlag,
			verboseFlag,
		},
	}
)

func gen(ctx *cli.Context) error {
	num := ctx.Int(numOfValidatorsFlag.Name)

	keys, nodekeys, addrs := istcommon.GenerateKeys(num)
	var nodes []string

	if ctx.Bool(verboseFlag.Name) {
		for i := 0; i < num; i++ {
			v := &validatorInfo{
				Address: addrs[i],
				Nodekey: nodekeys[i],
				NodeInfo: discover.NewNode(
					discover.PubkeyID(&keys[i].PublicKey),
					net.ParseIP("0.0.0.0"),
					0,
					uint16(30303)).String(),
			}

			str, _ := json.MarshalIndent(v, "", "\t")
			fmt.Println(string(str))
			nodes = append(nodes, string(v.NodeInfo))
		}

		fmt.Print("\n===========================================================\n\n")
	}

	if ctx.Bool(staticNodesFlag.Name) {
		staticNodes, _ := json.MarshalIndent(nodes, "", "    ")
		fmt.Println(string(staticNodes))
		fmt.Print("\n===========================================================\n\n")
	}

	genesis := genesis.New(
		genesis.Validators(addrs...),
		genesis.Alloc(addrs, new(big.Int).Exp(big.NewInt(10), big.NewInt(50), nil)),
	)

	jsonBytes, _ := json.MarshalIndent(genesis, "", "    ")
	fmt.Println(string(jsonBytes))

	return nil
}

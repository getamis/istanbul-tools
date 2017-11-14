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

package setup

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
	SetupCommand = cli.Command{
		Name:  "setup",
		Usage: "Setup your Istanbul network in seconds",
		Description: `This tool helps generate:

		* Genesis block
		* Static nodes for all validators
		* Validator details

	    for Istanbul consensus.
`,
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
		fmt.Println("validators")
	}

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

		nodes = append(nodes, string(v.NodeInfo))

		if ctx.Bool(verboseFlag.Name) {
			str, _ := json.MarshalIndent(v, "", "\t")
			fmt.Println(string(str))
		}
	}

	if ctx.Bool(verboseFlag.Name) {
		fmt.Print("\n\n\n")
	}

	if ctx.Bool(staticNodesFlag.Name) {
		staticNodes, _ := json.MarshalIndent(nodes, "", "\t")
		fmt.Println("static-nodes.json")
		fmt.Println(string(staticNodes))
		fmt.Print("\n\n\n")
	}

	genesis := genesis.New(
		genesis.Validators(addrs...),
		genesis.Alloc(addrs, new(big.Int).Exp(big.NewInt(10), big.NewInt(50), nil)),
	)

	jsonBytes, _ := json.MarshalIndent(genesis, "", "    ")
	fmt.Println("genesis.json")
	fmt.Println(string(jsonBytes))

	return nil
}

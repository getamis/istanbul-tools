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
	"io/ioutil"
	"math/big"
	"net"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/p2p/discv5"
	istcommon "github.com/jpmorganchase/istanbul-tools/common"
	"github.com/jpmorganchase/istanbul-tools/genesis"
	"github.com/urfave/cli"
)

type validatorInfo struct {
	Address  common.Address
	Nodekey  string
	NodeInfo string
}

var (
	SetupCommand = cli.Command{
		Name:  "setup",
		Usage: "Setup your qbft network in seconds",
		Description: `This tool helps generate:

		* Genesis block
		* Static nodes for all validators
		* Validator details

	    for qbft consensus.
`,
		Action: gen,
		Flags: []cli.Flag{
			numOfValidatorsFlag,
			staticNodesFlag,
			verboseFlag,
			quorumFlag,
			saveFlag,
			nodeIpFlag,
			nodePortBaseFlag,
			nodePortIncrementFlag,
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

	nodeIp := ctx.String(nodeIpFlag.Name)
	nodePort := ctx.Int(nodePortBaseFlag.Name)
	nodePortIncrement := ctx.Int(nodePortIncrementFlag.Name)

	for i := 0; i < num; i++ {
		v := &validatorInfo{
			Address: addrs[i],
			Nodekey: nodekeys[i],
			NodeInfo: discv5.NewNode(
				discv5.PubkeyID(&keys[i].PublicKey),
				net.ParseIP(nodeIp),
				0,
				uint16(nodePort)).String(),
		}
		nodePort = nodePort + nodePortIncrement

		nodes = append(nodes, string(v.NodeInfo))

		if ctx.Bool(verboseFlag.Name) {
			str, _ := json.MarshalIndent(v, "", "\t")
			fmt.Println(string(str))

			if ctx.Bool(saveFlag.Name) {
				folderName := strconv.Itoa(i)
				os.MkdirAll(folderName, os.ModePerm)
				ioutil.WriteFile(path.Join(folderName, "nodekey"), []byte(nodekeys[i]), os.ModePerm)
			}
		}
	}

	if ctx.Bool(verboseFlag.Name) {
		fmt.Print("\n\n\n")
	}

	staticNodes, _ := json.MarshalIndent(nodes, "", "\t")
	if ctx.Bool(staticNodesFlag.Name) {
		name := "static-nodes.json"
		fmt.Println(name)
		fmt.Println(string(staticNodes))
		fmt.Print("\n\n\n")

		if ctx.Bool(saveFlag.Name) {
			ioutil.WriteFile(name, staticNodes, os.ModePerm)
		}
	}

	var jsonBytes []byte
	isQuorum := ctx.Bool(quorumFlag.Name)
	g := genesis.New(
		genesis.QbftExtraData(addrs...),
		genesis.Alloc(addrs, new(big.Int).Exp(big.NewInt(10), big.NewInt(50), nil)),
		genesis.AddQbftBlock(),
	)

	if isQuorum {
		jsonBytes, _ = json.MarshalIndent(genesis.ToQuorum(g, true), "", "    ")
	} else {
		jsonBytes, _ = json.MarshalIndent(g, "", "    ")
	}

	fmt.Println("genesis.json")
	fmt.Println(string(jsonBytes))

	if ctx.Bool(saveFlag.Name) {
		ioutil.WriteFile("genesis.json", jsonBytes, os.ModePerm)
	}

	return nil
}

func removeSpacesAndLines(b []byte) string {
	out := string(b)
	out = strings.Replace(out, " ", "", -1)
	out = strings.Replace(out, "\t", "", -1)
	out = strings.Replace(out, "\n", "", -1)
	return out
}

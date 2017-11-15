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
	"github.com/ethereum/go-ethereum/p2p/discover"
	"github.com/getamis/istanbul-tools/cmd/istanbul/setup/docker"
	istcommon "github.com/getamis/istanbul-tools/common"
	"github.com/getamis/istanbul-tools/genesis"
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
			dockerComposeFlag,
			saveFlag,
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

	genesis := genesis.New(
		genesis.Validators(addrs...),
		genesis.Alloc(addrs, new(big.Int).Exp(big.NewInt(10), big.NewInt(50), nil)),
	)

	jsonBytes, _ := json.MarshalIndent(genesis, "", "    ")
	fmt.Println("genesis.json")
	fmt.Println(string(jsonBytes))

	if ctx.Bool(saveFlag.Name) {
		ioutil.WriteFile("genesis.json", jsonBytes, os.ModePerm)
	}

	if ctx.Bool(dockerComposeFlag.Name) {
		g := removeSpacesAndLines(jsonBytes)
		nodes := removeSpacesAndLines(staticNodes)

		var links []string
		compose := docker.Compose{
			IPPrefix: "172.16.239",
		}
		compose.Stats = docker.EthStats{
			// ethstats ip = {{ .IPPrefix }}.9
			IP:     fmt.Sprintf("%v.9", compose.IPPrefix),
			Port:   "3000",
			Secret: "bb98a0b6442386d0cdf8a31b267892c1",
		}

		for i := 0; i < num; i++ {
			s := docker.Service{
				Identity: fmt.Sprintf("validator-%v", i),
				Genesis:  g,
				NodeKey:  nodekeys[i],
				Port:     strconv.Itoa(30303 + i),
				RPCPort:  strconv.Itoa(8545 + i),
				EthStats: compose.Stats.Stats(),
				// from subnet ip 10
				IP: fmt.Sprintf("%v.%v", compose.IPPrefix, i+10),
			}

			nodes = strings.Replace(nodes, "0.0.0.0", s.IP, 1)
			links = append(links, s.Identity)
			compose.Services = append(compose.Services, s)
		}

		// update static nodes
		for i := 0; i < num; i++ {
			compose.Services[i].StaticNodes = nodes
		}

		fmt.Print("\n\n\n")
		fmt.Println("docker-compose.yml")
		fmt.Println(compose.String())
		if ctx.Bool(saveFlag.Name) {
			ioutil.WriteFile("docker-compose.yml", []byte(compose.String()), os.ModePerm)
		}
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

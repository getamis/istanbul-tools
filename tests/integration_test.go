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

package tests

import (
	"context"
	"fmt"
	"math/big"
	"path/filepath"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/getamis/go-ethereum/ethclient"
	"github.com/getamis/istanbul-tools/container"
	"github.com/getamis/istanbul-tools/core"
)

// var geths []container.Ethereum

var _ = Describe("4 validators Istanbul", func() {
	const (
		numberOfValidators = 4
	)
	var (
		envs  []*core.Env
		geths []container.Ethereum
	)

	BeforeSuite(func() {
		keys := core.GenerateClusterKeys(numberOfValidators)
		envs = core.SetupEnv(keys)
		err := core.SetupNodes(envs)
		Expect(err).To(BeNil())

		for _, env := range envs {
			geth := container.NewEthereum(
				container.ImageName("quay.io/maicoin/ottoman_geth:istanbul_develop"),
				container.HostDataDir(env.DataDir),
				container.DataDir("/data"),
				container.Port(fmt.Sprintf("%d", env.P2PPort)),
				container.RPC(),
				container.RPCAddress("0.0.0.0"),
				container.RPCAPI("eth,net,web3,personal"),
				container.RPCPort(fmt.Sprintf("%d", env.RpcPort)),
				container.NAT("any"),
				container.NoDiscover(),
				container.Logging(true),
			)

			err := geth.Init(filepath.Join(env.DataDir, core.GenesisJson))
			Expect(err).To(BeNil())

			geths = append(geths, geth)

			err = geth.Start()
			Expect(err).To(BeNil())
		}
	})

	AfterSuite(func() {
		for _, geth := range geths {
			geth.Stop()
		}

		core.Teardown(envs)
	})

	It("Blockchain creation", func() {
		for _, env := range envs {
			cli, err := ethclient.Dial("http://localhost:" + fmt.Sprintf("%d", env.RpcPort))
			Expect(err).To(BeNil())

			block, err := cli.BlockByNumber(context.Background(), big.NewInt(0))
			Expect(err).To(BeNil())
			Expect(block).NotTo(BeNil())
		}
	})
})

func TestIstanbul(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Istanbul Test Suite")
}

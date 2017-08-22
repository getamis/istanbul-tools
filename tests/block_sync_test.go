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
	"time"

	"github.com/ethereum/go-ethereum/core/types"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/getamis/istanbul-tools/container"
)

var _ = Describe("Block synchronization testing", func() {
	const (
		numberOfValidators = 4
	)
	var (
		blockchain container.Blockchain
	)

	BeforeEach(func() {
		blockchain = container.NewDefaultBlockchain(numberOfValidators)
		Expect(blockchain.Start(true)).To(BeNil())
	})

	AfterEach(func() {
		Expect(blockchain.Stop(true)).To(BeNil())
		blockchain.Finalize()
	})

	Describe("TFS-06 block synchronization testing", func() {
		const numberOfNodes = 2
		var nodes []container.Ethereum

		BeforeEach(func() {
			var err error
			nodes, err = blockchain.CreateNodes(numberOfNodes,
				container.ImageRepository("quay.io/amis/geth"),
				container.ImageTag("istanbul_develop"),
				container.DataDir("/data"),
				container.WebSocket(),
				container.WebSocketAddress("0.0.0.0"),
				container.WebSocketAPI("admin,eth,net,web3,personal,miner"),
				container.WebSocketOrigin("*"),
				container.NAT("any"),
			)

			Expect(err).To(BeNil())

			for _, n := range nodes {
				err = n.Start()
				Expect(err).To(BeNil())
			}
		})

		AfterEach(func() {
			for _, n := range nodes {
				n.Stop()
			}
		})

		It("TFS-06-01 node connection", func() {
			By("Connect all nodes to the validators")
			for i := 0; i < numberOfNodes; i++ {
				nodeClient := nodes[i].NewIstanbulClient()
				Expect(nodeClient).NotTo(BeNil())

				for _, v := range blockchain.Validators() {
					nodeClient.AddPeer(context.Background(), v.NodeAddress())
				}

				nodeClient.Close()
			}

			By("Wait for p2p connection")
			<-time.After(10 * time.Second)

			By("Check peer count")
			for i := 0; i < numberOfNodes; i++ {
				nodeClient := nodes[i].NewIstanbulClient()
				p2pInfos, err := nodeClient.AdminPeers(context.Background())

				Expect(err).To(BeNil())
				Expect(len(p2pInfos)).To(Equal(len(blockchain.Validators())))

				nodeClient.Close()
			}
		})

		It("TFS-06-02 node synchronization", func() {
			By("Wait for block generation")
			<-time.After(10 * time.Second)

			By("Stop consensus")
			for _, v := range blockchain.Validators() {
				client := v.NewIstanbulClient()
				Expect(client).NotTo(BeNil())
				err := client.StopMining(context.Background())
				Expect(err).To(BeNil())
				client.Close()
			}

			By("Wait for block synchronization")
			<-time.After(10 * time.Second)

			By("Check block height of validators")
			var latestBlock *types.Block
			for _, v := range blockchain.Validators() {
				client := v.NewClient()
				block, err := client.BlockByNumber(context.Background(), nil)

				Expect(err).To(BeNil())
				Expect(block).NotTo(BeNil())

				if latestBlock == nil {
					latestBlock = block
				} else {
					Expect(latestBlock.Hash()).To(BeEquivalentTo(block.Hash()))
				}
			}

			By("Connect all nodes to the validators")
			for i := 0; i < numberOfNodes; i++ {
				nodeClient := nodes[i].NewIstanbulClient()
				Expect(nodeClient).NotTo(BeNil())

				for _, v := range blockchain.Validators() {
					nodeClient.AddPeer(context.Background(), v.NodeAddress())
				}

				nodeClient.Close()
			}

			By("Wait for block synchronization")
			<-time.After(10 * time.Second)

			By("Check block height of nodes")
			for i := 0; i < numberOfNodes; i++ {
				nodeClient := nodes[i].NewClient()
				block, err := nodeClient.BlockByNumber(context.Background(), nil)

				Expect(err).To(BeNil())
				Expect(block).NotTo(BeNil())
				Expect(latestBlock.Hash()).To(BeEquivalentTo(block.Hash()))
			}
		})
	})
})

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

package functional

import (
	"sync"

	"github.com/getamis/istanbul-tools/tests"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/getamis/istanbul-tools/container"
)

var _ = Describe("QFS-01: General consensus", func() {
	const (
		numberOfValidators = 4
	)
	var (
		constellationNetwork container.ConstellationNetwork
		blockchain           container.Blockchain
	)

	BeforeEach(func() {
		constellationNetwork = container.NewDefaultConstellationNetwork(dockerNetwork, numberOfValidators)
		Expect(constellationNetwork.Start()).To(BeNil())
		blockchain = container.NewDefaultQuorumBlockchain(dockerNetwork, constellationNetwork)
		Expect(blockchain.Start(true)).To(BeNil())
	})

	AfterEach(func() {
		blockchain.Stop(true)
		blockchain.Finalize()
		constellationNetwork.Stop()
		constellationNetwork.Finalize()
	})

	// FIt("QFS-01-01, QFS-01-02: Blockchain initialization and run", func() {
	// 	fmt.Printf("validators:%v\n", blockchain.Validators())
	// 	errc := make(chan error, len(blockchain.Validators()))
	// 	valSet := make(map[common.Address]bool, numberOfValidators)
	// 	for _, geth := range blockchain.Validators() {
	// 		valSet[geth.Address()] = true
	// 	}
	// 	for _, geth := range blockchain.Validators() {
	// 		go func(geth container.Ethereum) {
	// 			// 1. Verify genesis block
	// 			c := geth.NewClient()
	// 			header, err := c.HeaderByNumber(context.Background(), big.NewInt(0))
	// 			if err != nil {
	// 				errc <- err
	// 				return
	// 			}

	// 			if header.GasLimit.Int64() != genesis.InitGasLimit {
	// 				errStr := fmt.Sprintf("Invalid genesis gas limit. want:%v, got:%v", genesis.InitGasLimit, header.GasLimit.Int64())
	// 				errc <- errors.New(errStr)
	// 				return
	// 			}

	// 			if header.Difficulty.Int64() != genesis.InitDifficulty {
	// 				errStr := fmt.Sprintf("Invalid genesis difficulty. want:%v, got:%v", genesis.InitDifficulty, header.Difficulty.Int64())
	// 				errc <- errors.New(errStr)
	// 				return
	// 			}

	// 			if header.MixDigest != types.IstanbulDigest {
	// 				errStr := fmt.Sprintf("Invalid block mixhash. want:%v, got:%v", types.IstanbulDigest, header.MixDigest)
	// 				errc <- errors.New(errStr)
	// 				return

	// 			}

	// 			// 2. Check validator set
	// 			istClient := geth.NewIstanbulClient()
	// 			vals, err := istClient.GetValidators(context.Background(), big.NewInt(0))
	// 			if err != nil {
	// 				errc <- err
	// 				return
	// 			}

	// 			for _, val := range vals {
	// 				if _, ok := valSet[val]; !ok {
	// 					errc <- errors.New("Invalid validator address.")
	// 					return
	// 				}
	// 			}

	// 			errc <- nil
	// 		}(geth)
	// 	}

	// 	for i := 0; i < len(blockchain.Validators()); i++ {
	// 		err := <-errc
	// 		Expect(err).To(BeNil())
	// 	}

	// })

	It("QFS-01-03: Peer connection", func(done Done) {
		expectedPeerCount := len(blockchain.Validators()) - 1
		tests.WaitFor(blockchain.Validators(), func(v container.Ethereum, wg *sync.WaitGroup) {
			Expect(v.WaitForPeersConnected(expectedPeerCount)).To(BeNil())
			wg.Done()
		})

		close(done)
	}, 20)

	// It("TFS-01-04: Consensus progress", func(done Done) {
	// 	const (
	// 		targetBlockHeight = 10
	// 		maxBlockPeriod    = 3
	// 	)

	// 	By("Wait for consensus progress", func() {
	// 		tests.WaitFor(blockchain.Validators(), func(geth container.Ethereum, wg *sync.WaitGroup) {
	// 			Expect(geth.WaitForBlockHeight(targetBlockHeight)).To(BeNil())
	// 			wg.Done()
	// 		})
	// 	})

	// 	By("Check the block period should less than 3 seconds", func() {
	// 		errc := make(chan error, len(blockchain.Validators()))
	// 		for _, geth := range blockchain.Validators() {
	// 			go func(geth container.Ethereum) {
	// 				c := geth.NewClient()
	// 				lastBlockTime := int64(0)
	// 				// The reason to verify block period from block#2 is that
	// 				// the block period from block#1 to block#2 might take long time due to
	// 				// encounter several round changes at the beginning of the consensus progress.
	// 				for i := 2; i <= targetBlockHeight; i++ {
	// 					header, err := c.HeaderByNumber(context.Background(), big.NewInt(int64(i)))
	// 					if err != nil {
	// 						errc <- err
	// 						return
	// 					}
	// 					if lastBlockTime != 0 {
	// 						diff := header.Time.Int64() - lastBlockTime
	// 						if diff > maxBlockPeriod {
	// 							errStr := fmt.Sprintf("Invaild block(%v) period, want:%v, got:%v", header.Number.Int64(), maxBlockPeriod, diff)
	// 							errc <- errors.New(errStr)
	// 							return
	// 						}
	// 					}
	// 					lastBlockTime = header.Time.Int64()
	// 				}
	// 				errc <- nil
	// 			}(geth)
	// 		}

	// 		for i := 0; i < len(blockchain.Validators()); i++ {
	// 			err := <-errc
	// 			Expect(err).To(BeNil())
	// 		}
	// 	})
	// 	close(done)
	// }, 60)

	// It("TFS-01-05: Round robin proposer selection", func(done Done) {
	// 	var (
	// 		timesOfBeProposer = 3
	// 		targetBlockHeight = timesOfBeProposer * numberOfValidators
	// 		emptyProposer     = common.Address{}
	// 	)

	// 	By("Wait for consensus progress", func() {
	// 		tests.WaitFor(blockchain.Validators(), func(geth container.Ethereum, wg *sync.WaitGroup) {
	// 			Expect(geth.WaitForBlockHeight(targetBlockHeight)).To(BeNil())
	// 			wg.Done()
	// 		})
	// 	})

	// 	By("Block proposer selection should follow round-robin policy", func() {
	// 		errc := make(chan error, len(blockchain.Validators()))
	// 		for _, geth := range blockchain.Validators() {
	// 			go func(geth container.Ethereum) {
	// 				c := geth.NewClient()
	// 				istClient := geth.NewIstanbulClient()

	// 				// get initial validator set
	// 				vals, err := istClient.GetValidators(context.Background(), big.NewInt(0))
	// 				if err != nil {
	// 					errc <- err
	// 					return
	// 				}

	// 				lastProposerIdx := -1
	// 				counts := make(map[common.Address]int, numberOfValidators)
	// 				// initial count map
	// 				for _, addr := range vals {
	// 					counts[addr] = 0
	// 				}
	// 				for i := 1; i <= targetBlockHeight; i++ {
	// 					header, err := c.HeaderByNumber(context.Background(), big.NewInt(int64(i)))
	// 					if err != nil {
	// 						errc <- err
	// 						return
	// 					}

	// 					p := container.GetProposer(header)
	// 					if p == emptyProposer {
	// 						errStr := fmt.Sprintf("Empty block(%v) proposer", header.Number.Int64())
	// 						errc <- errors.New(errStr)
	// 						return
	// 					}
	// 					// count the times to be the proposer
	// 					if count, ok := counts[p]; ok {
	// 						counts[p] = count + 1
	// 					}
	// 					// check if the proposer is valid
	// 					if lastProposerIdx == -1 {
	// 						for i, val := range vals {
	// 							if p == val {
	// 								lastProposerIdx = i
	// 								break
	// 							}
	// 						}
	// 					} else {
	// 						proposerIdx := (lastProposerIdx + 1) % len(vals)
	// 						if p != vals[proposerIdx] {
	// 							errStr := fmt.Sprintf("Invaild block(%v) proposer, want:%v, got:%v", header.Number.Int64(), vals[proposerIdx], p)
	// 							errc <- errors.New(errStr)
	// 							return
	// 						}
	// 						lastProposerIdx = proposerIdx
	// 					}
	// 				}
	// 				// check times to be proposer
	// 				for _, count := range counts {
	// 					if count != timesOfBeProposer {
	// 						errc <- errors.New("Wrong times to be proposer.")
	// 						return
	// 					}
	// 				}
	// 				errc <- nil
	// 			}(geth)
	// 		}

	// 		for i := 0; i < len(blockchain.Validators()); i++ {
	// 			err := <-errc
	// 			Expect(err).To(BeNil())
	// 		}
	// 	})
	// 	close(done)
	// }, 120)
})

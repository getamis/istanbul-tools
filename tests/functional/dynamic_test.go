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
	"context"
	"math"
	"sync"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/Consensys/istanbul-tools/container"
	"github.com/Consensys/istanbul-tools/tests"
)

var _ = Describe("TFS-02: Dynamic validators addition/removal testing", func() {
	const (
		numberOfValidators = 4
	)
	var (
		blockchain container.Blockchain
	)

	BeforeEach(func() {
		blockchain = container.NewDefaultBlockchain(dockerNetwork, numberOfValidators)
		Expect(blockchain.Start(true)).To(BeNil())
	})

	AfterEach(func() {
		Expect(blockchain.Stop(false)).To(BeNil())
		blockchain.Finalize()
	})

	It("TFS-02-01: Add validators", func() {
		testValidators := 1

		By("Ensure the number of validators is correct", func() {
			for _, v := range blockchain.Validators() {
				client := v.NewClient()
				validators, err := client.GetValidators(context.Background(), nil)
				Expect(err).Should(BeNil())
				Expect(len(validators)).Should(BeNumerically("==", numberOfValidators))
			}
		})

		By("Add validators", func() {
			_, err := blockchain.AddValidators(testValidators)
			Expect(err).Should(BeNil())
		})

		By("Wait for several blocks", func() {
			tests.WaitFor(blockchain.Validators(), func(geth container.Ethereum, wg *sync.WaitGroup) {
				Expect(geth.WaitForBlocks(5)).To(BeNil())
				wg.Done()
			})
		})

		By("Ensure the number of validators is correct", func() {
			for _, v := range blockchain.Validators() {
				client := v.NewClient()
				validators, err := client.GetValidators(context.Background(), nil)
				Expect(err).Should(BeNil())
				Expect(len(validators)).Should(BeNumerically("==", numberOfValidators+testValidators))
			}
		})
	})

	It("TFS-02-02: New validators consensus participation", func() {
		testValidator := 1

		newValidators, err := blockchain.AddValidators(testValidator)
		Expect(err).Should(BeNil())

		tests.WaitFor(blockchain.Validators()[numberOfValidators:], func(eth container.Ethereum, wg *sync.WaitGroup) {
			Expect(eth.WaitForProposed(newValidators[0].Address(), 100*time.Second)).Should(BeNil())
			wg.Done()
		})
	})

	It("TFS-02-03: Remove validators", func() {
		numOfCandidates := 3

		By("Ensure that numbers of validator is equal to $numberOfValidators", func() {
			for _, v := range blockchain.Validators() {
				client := v.NewClient()
				validators, err := client.GetValidators(context.Background(), nil)
				Expect(err).Should(BeNil())
				Expect(len(validators)).Should(BeNumerically("==", numberOfValidators))
			}
		})

		By("Add validators", func() {
			_, err := blockchain.AddValidators(numOfCandidates)
			Expect(err).Should(BeNil())
		})

		By("Ensure that consensus is working in 50 seconds", func() {
			Expect(blockchain.EnsureConsensusWorking(blockchain.Validators(), 50*time.Second)).Should(BeNil())
		})

		By("Check if the number of validators is correct", func() {
			for _, v := range blockchain.Validators() {
				client := v.NewClient()
				validators, err := client.GetValidators(context.Background(), nil)
				Expect(err).Should(BeNil())
				Expect(len(validators)).Should(BeNumerically("==", numberOfValidators+numOfCandidates))
			}
		})

		// remove validators [1,2,3]
		By("Remove validators", func() {
			removalCandidates := blockchain.Validators()[:numOfCandidates]
			processingTime := time.Duration(math.Pow(2, float64(len(removalCandidates)))*7) * time.Second
			Expect(blockchain.RemoveValidators(removalCandidates, processingTime)).Should(BeNil())
		})

		By("Ensure that consensus is working in 20 seconds", func() {
			Expect(blockchain.EnsureConsensusWorking(blockchain.Validators(), 20*time.Second)).Should(BeNil())
		})

		By("Check if the number of validators is correct", func() {
			for _, v := range blockchain.Validators() {
				client := v.NewClient()
				validators, err := client.GetValidators(context.Background(), nil)
				Expect(err).Should(BeNil())
				Expect(len(validators)).Should(BeNumerically("==", numberOfValidators))
			}
		})

		By("Ensure that consensus is working in 30 seconds", func() {
			Expect(blockchain.EnsureConsensusWorking(blockchain.Validators(), 30*time.Second)).Should(BeNil())
		})
	})

	It("TFS-02-04: Reduce validator network size below 2F+1", func() {
		By("Ensure that blocks are generated by validators", func() {
			tests.WaitFor(blockchain.Validators(), func(geth container.Ethereum, wg *sync.WaitGroup) {
				Expect(geth.WaitForBlocks(5)).To(BeNil())
				wg.Done()
			})
		})

		By("Reduce validator network size but keep it more than 2F+1", func() {
			// stop validators [3]
			stopCandidates := blockchain.Validators()[numberOfValidators-1:]
			for _, candidates := range stopCandidates {
				c := candidates.NewClient()
				Expect(c.StopMining(context.Background())).Should(BeNil())
			}
		})

		By("Verify number of validators", func() {
			for _, v := range blockchain.Validators() {
				client := v.NewClient()
				validators, err := client.GetValidators(context.Background(), nil)
				Expect(err).Should(BeNil())
				Expect(len(validators)).Should(BeNumerically("==", numberOfValidators))
			}
		})

		By("Ensure that blocks are generated by validators", func() {
			tests.WaitFor(blockchain.Validators()[:numberOfValidators-1], func(geth container.Ethereum, wg *sync.WaitGroup) {
				Expect(geth.WaitForBlocks(5)).To(BeNil())
				wg.Done()
			})
		})
	})

	It("TFS-02-05: Reduce validator network size below 2F+1", func() {
		By("Ensure that blocks are generated by validators", func() {
			tests.WaitFor(blockchain.Validators(), func(geth container.Ethereum, wg *sync.WaitGroup) {
				Expect(geth.WaitForBlocks(5)).To(BeNil())
				wg.Done()
			})
		})

		By("Reduce validator network size to less than 2F+1", func() {
			stopCandidates := blockchain.Validators()[numberOfValidators-2:]
			// stop validators [3,4]
			for _, candidates := range stopCandidates {
				c := candidates.NewClient()
				Expect(c.StopMining(context.Background())).Should(BeNil())
			}
		})

		By("Verify number of validators", func() {
			for _, v := range blockchain.Validators() {
				client := v.NewClient()
				validators, err := client.GetValidators(context.Background(), nil)
				Expect(err).Should(BeNil())
				Expect(len(validators)).Should(BeNumerically("==", numberOfValidators))
			}
		})

		By("No block generated", func() {
			// REMARK: ErrNoBlock will return if validators not generate block after 10 second.
			Expect(blockchain.EnsureConsensusWorking(blockchain.Validators(), 11*time.Second)).Should(Equal(container.ErrNoBlock))
		})
	})
})

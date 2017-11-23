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
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/getamis/istanbul-tools/container"
	"github.com/getamis/istanbul-tools/tests"
)

var _ = Describe("TFS-03: Recoverability testing", func() {
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
		blockchain.Stop(true) // This will return container not found error since we stop one
		blockchain.Finalize()
	})

	It("TFS-03-01: Add validators in a network with < 2F+1 validators to > 2F+1", func(done Done) {
		By("The consensus should work at the beginning", func() {
			tests.WaitFor(blockchain.Validators(), func(geth container.Ethereum, wg *sync.WaitGroup) {
				Expect(geth.WaitForBlocks(5)).To(BeNil())
				wg.Done()
			})
		})

		numOfValidatorsToBeStopped := 2

		By("Stop several validators until there are less than 2F+1 validators", func() {
			tests.WaitFor(blockchain.Validators()[:numOfValidatorsToBeStopped], func(geth container.Ethereum, wg *sync.WaitGroup) {
				Expect(geth.StopMining()).To(BeNil())
				wg.Done()
			})
		})

		By("The consensus should not work after resuming", func() {
			tests.WaitFor(blockchain.Validators(), func(geth container.Ethereum, wg *sync.WaitGroup) {
				// container.ErrNoBlock should be returned if we didn't see any block in 10 seconds
				Expect(geth.WaitForBlocks(1, 10*time.Second)).To(BeEquivalentTo(container.ErrNoBlock))
				wg.Done()
			})
		})

		By("Resume the stopped validators", func() {
			tests.WaitFor(blockchain.Validators()[:numOfValidatorsToBeStopped], func(geth container.Ethereum, wg *sync.WaitGroup) {
				Expect(geth.StartMining()).To(BeNil())
				wg.Done()
			})
		})

		By("The consensus should work after resuming", func() {
			tests.WaitFor(blockchain.Validators(), func(geth container.Ethereum, wg *sync.WaitGroup) {
				Expect(geth.WaitForBlocks(5)).To(BeNil())
				wg.Done()
			})
		})

		close(done)
	}, 120)
})

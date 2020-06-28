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

	"github.com/jpmorganchase/istanbul-tools/container"
	"github.com/jpmorganchase/istanbul-tools/tests"
)

var _ = Describe("QFS-04: Non-Byzantine Faulty", func() {
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

	It("QFS-04-01: Stop F validators", func(done Done) {
		By("Generating blockchain progress before stopping validator", func() {
			tests.WaitFor(blockchain.Validators(), func(geth container.Ethereum, wg *sync.WaitGroup) {
				Expect(geth.WaitForBlocks(3)).To(BeNil())
				wg.Done()
			})
		})

		By("Stopping validator 0", func() {
			v0 := blockchain.Validators()[0]
			e := v0.Stop()
			Expect(e).To(BeNil())
			ticker := time.NewTicker(time.Millisecond * 100)
			for _ = range ticker.C {
				e := v0.Stop()
				// Wait for e to be non-nil to make sure the container is down
				if e != nil {
					ticker.Stop()
					break
				}
			}
		})

		By("Checking blockchain progress after stopping validator", func() {
			tests.WaitFor(blockchain.Validators()[1:], func(geth container.Ethereum, wg *sync.WaitGroup) {
				Expect(geth.WaitForBlocks(3)).To(BeNil())
				wg.Done()
			})
		})

		close(done)
	}, 120)
})

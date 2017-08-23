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

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/getamis/istanbul-tools/container"
)

var _ = Describe("TFS-04: Non-Byzantine Faulty", func() {
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
		blockchain.Stop(true) // This will return container not found error since we stop one
		blockchain.Finalize()
	})

	It("TFS-04-01: Stop F validators", func(done Done) {
		v0 := blockchain.Validators()[0]
		By("Generating blockchain progress before stopping validator", func() {
			c0 := v0.NewClient()
			ticker := time.NewTicker(time.Millisecond * 100)
			for _ = range ticker.C {
				b, e := c0.BlockByNumber(context.Background(), nil)
				Expect(e).To(BeNil())
				// Check if new blocks are getting generated
				if b.Number().Int64() > 1 {
					ticker.Stop()
					break
				}
			}
		})

		By("Stopping validator 0", func() {
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
			v1 := blockchain.Validators()[1]
			c1 := v1.NewClient()
			b1, e := c1.BlockByNumber(context.Background(), nil)
			Expect(e).To(BeNil())
			ticker := time.NewTicker(time.Millisecond * 100)
			for _ = range ticker.C {
				newB1, e := c1.BlockByNumber(context.Background(), nil)
				Expect(e).To(BeNil())
				if newB1.Number().Int64() > b1.Number().Int64() {
					ticker.Stop()
					break
				}
			}
		})

		close(done)
	}, 120)
})

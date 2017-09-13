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

package load

import (
	"context"
	"math/big"
	"sync"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/getamis/istanbul-tools/container"
	"github.com/getamis/istanbul-tools/k8s"
	"github.com/getamis/istanbul-tools/tests"
)

var _ = Describe("TPS-01: Large amount of transactions", func() {
	tests.CaseTable("with number of validators",
		func(numberOfValidators int) {
			tests.CaseTable("with gas limit",
				func(gaslimit int) {
					tests.CaseTable("with txpool size",
						func(txpoolSize int) {
							runTests(numberOfValidators, gaslimit, txpoolSize)
						},

						tests.Case("2048", 2048),
						tests.Case("10240", 10240),
					)
				},

				tests.Case("21000*1000", 21000*1000),
				tests.Case("21000*3000", 21000*3000),
			)
		},

		tests.Case("4 validators", 4),
	)
})

func runTests(numberOfValidators int, gaslimit int, txpoolSize int) {
	Describe("", func() {
		var (
			blockchain container.Blockchain
		)

		BeforeEach(func() {
			blockchain = k8s.NewBlockchain(
				numberOfValidators,
				uint64(gaslimit),
				k8s.ImageRepository("quay.io/amis/geth"),
				k8s.ImageTag("istanbul_develop"),
				k8s.ServiceType("LoadBalancer"),
				k8s.Mine(),
				k8s.TxPoolSize(txpoolSize),
			)
			Expect(blockchain.Start(true)).To(BeNil())
		})

		AfterEach(func() {
			Expect(blockchain.Stop(true)).To(BeNil())
			blockchain.Finalize()
		})

		It("", func() {
			tests.WaitFor(blockchain.Validators(), func(geth container.Ethereum, wg *sync.WaitGroup) {
				richman, ok := geth.(k8s.RichMan)
				Expect(ok).To(BeTrue())

				var addrs []common.Address
				for _, acc := range geth.Accounts() {
					addrs = append(addrs, acc.Address)
				}

				// Give ether to all accounts
				err := richman.GiveEther(context.Background(), addrs, new(big.Int).Exp(big.NewInt(10), big.NewInt(24), nil))
				Expect(err).NotTo(BeNil())

				err = geth.WaitForBalances(addrs, 10*time.Second)
				Expect(err).NotTo(BeNil())

				wg.Done()
			})
		})
	})
}

func TestIstanbulLoadTesting(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Istanbul Load Test Suite")
}

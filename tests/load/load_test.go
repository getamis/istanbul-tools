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
	"fmt"
	"math/big"
	"sync"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	istcommon "github.com/Consensys/istanbul-tools/common"
	"github.com/Consensys/istanbul-tools/container"
	"github.com/Consensys/istanbul-tools/k8s"
	"github.com/Consensys/istanbul-tools/metrics"
	"github.com/Consensys/istanbul-tools/tests"
)

var _ = Describe("TPS-01: Large amount of transactions", func() {
	tests.CaseTable("with number of validators",
		func(numberOfValidators int) {
			tests.CaseTable("with gas limit",
				func(gaslimit int) {
					tests.CaseTable("with txpool size",
						func(txpoolSize int) {
							tests.CaseTable("with tx send rate",
								func(rate int) {
									runTests(numberOfValidators, gaslimit, txpoolSize, rate)
								},
								// only preload txs if send rate is 0
								// tests.Case("preload", 0),
								tests.Case("300ms", 300),
							)
						},

						tests.Case("20480", 20480),
					)
				},

				tests.Case("21000*1500", 21000*1500),
			)

		},

		tests.Case("4 validators", 4),
	)
})

func runTests(numberOfValidators int, gaslimit int, txpoolSize int, sendRate int) {
	Describe("", func() {
		const (
			preloadAccounts = 10
			sendAccount     = 20
		)
		var (
			blockchain     container.Blockchain
			sendEtherAddrs map[common.Address]common.Address

			duration        = 5 * time.Minute
			accountsPerGeth = preloadAccounts + sendAccount

			allTPSSanpshotStopper metrics.SnapshotStopper
		)

		BeforeEach(func() {
			blockchain = k8s.NewBlockchain(
				numberOfValidators,
				accountsPerGeth,
				uint64(gaslimit),
				true,
				k8s.ImageRepository("quay.io/amis/quorum"),
				k8s.ImageTag("latest"),
				k8s.Mine(false),
				k8s.TxPoolSize(txpoolSize),
			)
			blockchain = metrics.NewMetricChain(blockchain)
			Expect(blockchain).NotTo(BeNil())
			Expect(blockchain.Start(true)).To(BeNil())

			sendEtherAddrs = make(map[common.Address]common.Address)
			num := len(blockchain.Validators())
			for i, v := range blockchain.Validators() {
				sendEtherAddrs[v.Address()] = blockchain.Validators()[(i+1)%num].Address()
			}

			if metricsExport, ok := blockchain.(metrics.Exporter); ok {
				allTPSSanpshotStopper = metricsExport.SnapshotTxRespMeter("all")
			}
		})

		AfterEach(func() {
			Expect(blockchain).NotTo(BeNil())
			if allTPSSanpshotStopper != nil {
				allTPSSanpshotStopper()
			}
			fmt.Println("Begin to Stop blockchain")
			Expect(blockchain.Stop(true)).To(BeNil())
			fmt.Println("End to Stop blockchain")
			blockchain.Finalize()
		})

		It("", func() {
			By("Wait for p2p connection", func() {
				tests.WaitFor(blockchain.Validators(), func(geth container.Ethereum, wg *sync.WaitGroup) {
					fmt.Println("Start p2p")
					Expect(geth.WaitForPeersConnected(numberOfValidators - 1)).To(BeNil())
					wg.Done()
					fmt.Println("Done p2p")
				})
			})

			By("Preload transactions", func() {
				tests.WaitFor(blockchain.Validators(), func(geth container.Ethereum, wg *sync.WaitGroup) {
					transactor, ok := geth.(k8s.Transactor)
					Expect(ok).To(BeTrue())

					client := geth.NewClient()
					Expect(client).NotTo(BeNil())

					accounts := transactor.AccountKeys()[:preloadAccounts]
					preloadCnt := txpoolSize
					Expect(transactor.PreloadTransactions(
						client,
						accounts,
						new(big.Int).Exp(big.NewInt(10), big.NewInt(3), nil),
						preloadCnt)).To(BeNil())

					wg.Done()
				})
			})

			By("Start mining", func() {
				tests.WaitFor(blockchain.Validators(), func(geth container.Ethereum, wg *sync.WaitGroup) {
					fmt.Println("Start mining")
					Expect(geth.StartMining()).To(BeNil())
					wg.Done()
				})
			})

			By("Send transactions with specific rate", func() {
				if sendRate == 0 {
					fmt.Println("Skip to send tx")
					return
				}
				fmt.Println("Start to send tx")
				if metricsExport, ok := blockchain.(metrics.Exporter); ok {
					mname := fmt.Sprintf("rate%dms", sendRate)
					rpsSanpshotStopper := metricsExport.SnapshotTxReqMeter(mname)
					defer rpsSanpshotStopper()
					tpsSanpshotStopper := metricsExport.SnapshotTxRespMeter(mname)
					defer tpsSanpshotStopper()
				}
				tests.WaitFor(blockchain.Validators(), func(geth container.Ethereum, wg *sync.WaitGroup) {
					transactor, ok := geth.(k8s.Transactor)
					Expect(ok).To(BeTrue())

					client := geth.NewClient()
					Expect(client).NotTo(BeNil())

					accounts := transactor.AccountKeys()[preloadAccounts:]
					rate := time.Duration(sendRate) * time.Millisecond

					Expect(transactor.SendTransactions(
						client,
						accounts,
						new(big.Int).Exp(big.NewInt(10), big.NewInt(3), nil),
						duration,
						rate)).To(BeNil())

					wg.Done()
				})
			})

			By("Wait for txs consuming", func() {
				var blocksCnt int = 5
				metricsExport, ok := blockchain.(metrics.Exporter)
				if ok {

					blockSize := gaslimit / int(istcommon.DefaultGasLimit)
					blocksCnt = int(int(metricsExport.SentTxCount()-metricsExport.ExcutedTxCount())/blockSize/7*10) + 5
					fmt.Println("blockSize", blockSize, "sendTx", metricsExport.SentTxCount(), "excutedTx", metricsExport.ExcutedTxCount(), "waitFor", blocksCnt)

					tpsSanpshotStopper := metricsExport.SnapshotTxRespMeter("final")
					defer tpsSanpshotStopper()
				}

				tests.WaitFor(blockchain.Validators(), func(geth container.Ethereum, wg *sync.WaitGroup) {
					Expect(geth.WaitForBlocks(blocksCnt)).To(BeNil())
					wg.Done()
				})
			})
		})
	})
}

func TestIstanbulLoadTesting(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Istanbul Load Test Suite")
}

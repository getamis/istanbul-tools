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

package k8s

import (
	"context"
	"crypto/ecdsa"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/crypto"

	"github.com/jpmorganchase/istanbul-tools/client"
	istcommon "github.com/jpmorganchase/istanbul-tools/common"
)

type Transactor interface {
	AccountKeys() []*ecdsa.PrivateKey
	SendTransactions(client.Client, []*ecdsa.PrivateKey, *big.Int, time.Duration, time.Duration) error
	PreloadTransactions(client.Client, []*ecdsa.PrivateKey, *big.Int, int) error
}

func (eth *ethereum) AccountKeys() []*ecdsa.PrivateKey {
	return eth.accounts
}

// SendTransactions is to send a lot of transactions by each account in geth.
// duration: the period of sending transactions
// rate: total tx per second
func (eth *ethereum) SendTransactions(client client.Client, accounts []*ecdsa.PrivateKey, amount *big.Int, duration, frequnce time.Duration) error {
	var fns []func() error
	for i, key := range accounts {
		i := i
		key := key

		fn := func() error {
			fromAddr := crypto.PubkeyToAddress(key.PublicKey)
			toAddr := crypto.PubkeyToAddress(accounts[(i+1)%len(accounts)].PublicKey)
			timeout := time.After(duration)
			ticker := time.NewTicker(frequnce)
			defer ticker.Stop()

			nonce, err := client.NonceAt(context.Background(), fromAddr, nil)
			if err != nil {
				log.Error("Failed to get nonce", "addr", fromAddr, "err", err)
				return err
			}

			var wg sync.WaitGroup
			defer wg.Wait()

			errCh := make(chan error)
			for {
				select {
				case <-timeout:
					return nil
				case err := <-errCh:
					log.Error("Failed to SendEther", "addr", fromAddr, "to", toAddr, "err", err)
					return err
				case <-ticker.C:
					wg.Add(1)
					go func(nonce uint64, wg *sync.WaitGroup) {
						if err := istcommon.SendEther(client, key, toAddr, amount, nonce); err != nil {
							select {
							case errCh <- err:
							default:
							}
						}
						wg.Done()
					}(nonce, &wg)
					nonce++
				}
			}
		}

		fns = append(fns, fn)
	}

	return executeInParallel(fns...)
}

func (eth *ethereum) PreloadTransactions(client client.Client, accounts []*ecdsa.PrivateKey, amount *big.Int, txCount int) error {
	eachCount := txCount / len(accounts)

	var fns []func() error
	for i, key := range accounts {
		i := i
		key := key

		fn := func() error {
			fromAddr := crypto.PubkeyToAddress(key.PublicKey)
			toAddr := crypto.PubkeyToAddress(accounts[(i+1)%len(accounts)].PublicKey)
			nonce, err := client.NonceAt(context.Background(), fromAddr, nil)
			if err != nil {
				log.Error("Failed to get nonce", "addr", fromAddr, "err", err)
				return err
			}

			for i := 0; i < eachCount; i++ {
				if err := istcommon.SendEther(client, key, toAddr, amount, nonce); err != nil {
					return err
				}
				nonce++
			}
			return nil
		}

		fns = append(fns, fn)
	}

	return executeInParallel(fns...)
}

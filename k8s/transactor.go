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
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/crypto"

	"github.com/getamis/istanbul-tools/client"
	istcommon "github.com/getamis/istanbul-tools/common"
)

type Transactor interface {
	SendTransactions(*client.Client, *big.Int, time.Duration) error
}

func (eth *ethereum) SendTransactions(client *client.Client, amount *big.Int, duration time.Duration) error {
	var fns []func() error
	for i, key := range eth.accounts {
		i := i
		key := key

		fn := func() error {
			fromAddr := crypto.PubkeyToAddress(key.PublicKey)
			toAddr := crypto.PubkeyToAddress(eth.accounts[(i+1)%len(eth.accounts)].PublicKey)
			timeout := time.After(duration)

			nonce, err := client.NonceAt(context.Background(), fromAddr, nil)
			if err != nil {
				log.Error("Failed to get nonce", "addr", fromAddr, "err", err)
				return err
			}

			for {
				select {
				case <-timeout:
					return nil
				default:
					if err := istcommon.SendEther(client, key, toAddr, amount, nonce); err != nil {
						return err
					}
					nonce++
				}
			}
		}

		fns = append(fns, fn)
	}

	return executeInParallel(fns...)
}

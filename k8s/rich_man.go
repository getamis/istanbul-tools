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
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum/common"

	istcommon "github.com/getamis/istanbul-tools/common"
)

type RichMan interface {
	GiveEther(context.Context, []common.Address, *big.Int) error
}

func (eth *ethereum) GiveEther(ctx context.Context, accounts []common.Address, amount *big.Int) error {
	client := eth.NewClient()
	if client == nil {
		return errors.New("failed to retrieve client")
	}

	nonce, err := client.NonceAt(context.Background(), eth.Address(), nil)
	if err != nil {
		log.Error("Failed to get nonce", "addr", eth.Address(), "err", err)
		return err
	}

	for _, account := range accounts {
		_ = istcommon.SendEther(client, eth.key, account, amount, nonce)
		nonce++
	}

	return nil
}

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

package common

import (
	"context"
	"crypto/ecdsa"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/getamis/istanbul-tools/client"
)

var (
	DefaultGasPrice int64 = 0
	DefaultGasLimit int64 = 21000 // the gas of ether tx should be 21000
)

func SendEther(client client.Client, from *ecdsa.PrivateKey, to common.Address, amount *big.Int, nonce uint64) error {
	tx := types.NewTransaction(nonce, to, amount, big.NewInt(DefaultGasLimit), big.NewInt(DefaultGasPrice), []byte{})
	signedTx, err := types.SignTx(tx, types.HomesteadSigner{}, from)
	if err != nil {
		log.Error("Failed to sign transaction", "tx", tx, "err", err)
		return err
	}

	err = client.SendRawTransaction(context.Background(), signedTx)
	if err != nil {
		log.Error("Failed to send transaction", "tx", signedTx, "nonce", nonce, "err", err)
		return err
	}

	return nil
}

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

package metrics

import (
	"crypto/ecdsa"
	"errors"
	"math/big"
	"time"

	"github.com/getamis/istanbul-tools/client"
	"github.com/getamis/istanbul-tools/container"
	"github.com/getamis/istanbul-tools/k8s"
)

type metricEthereum struct {
	container.Ethereum
	txStartCh  chan *txInfo
	metricsMgr *metricsManager
}

func (e *metricEthereum) NewClient() client.Client {
	return &metricClient{
		Client:     e.Ethereum.NewClient(),
		txStartCh:  e.txStartCh,
		metricsMgr: e.metricsMgr,
	}
}

func (eth *metricEthereum) AccountKeys() []*ecdsa.PrivateKey {
	transactor, ok := eth.Ethereum.(k8s.Transactor)
	if !ok {
		return nil
	}
	return transactor.AccountKeys()
}

func (eth *metricEthereum) SendTransactions(client client.Client, accounts []*ecdsa.PrivateKey, amount *big.Int, duration, frequnce time.Duration) error {
	transactor, ok := eth.Ethereum.(k8s.Transactor)
	if !ok {
		return errors.New("Not support Transactor interface.")
	}
	return transactor.SendTransactions(client, accounts, amount, duration, frequnce)
}

func (eth *metricEthereum) PreloadTransactions(client client.Client, accounts []*ecdsa.PrivateKey, amount *big.Int, txCount int) error {
	transactor, ok := eth.Ethereum.(k8s.Transactor)
	if !ok {
		return errors.New("Not support Transactor interface.")
	}
	return transactor.PreloadTransactions(client, accounts, amount, txCount)

}

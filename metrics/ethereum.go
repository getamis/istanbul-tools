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
	"errors"
	"fmt"
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

func (eth *metricEthereum) SendTransactions(client client.Client, amount *big.Int, duration time.Duration) error {
	transactor, ok := eth.Ethereum.(k8s.Transactor)
	if !ok {
		return errors.New("Not support Transactor interface.")
	}
	fmt.Println("Begin to SendTransactions.")
	return transactor.SendTransactions(client, amount, duration)
}

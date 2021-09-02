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
	"context"
	"math/big"
	"time"

	"github.com/Consensys/istanbul-tools/client"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

type metricClient struct {
	client.Client
	txStartCh  chan *txInfo
	metricsMgr *metricsManager
}

func (c *metricClient) SendTransaction(ctx context.Context, from, to common.Address, value *big.Int) (hash string, err error) {
	defer func() {
		sendTime := time.Now()
		if err != nil {
			c.metricsMgr.TxErrCounter.Inc(1)
		} else {
			c.metricsMgr.SentTxCounter.Inc(1)
			c.metricsMgr.ReqMeter.Mark(1)
			go func() {
				c.txStartCh <- &txInfo{
					Hash: hash,
					Time: sendTime,
				}
			}()
		}
	}()
	return c.Client.SendTransaction(ctx, from, to, value)
}

func (c *metricClient) CreateContract(ctx context.Context, from common.Address, bytecode string, gas *big.Int) (hash string, err error) {
	defer func() {
		sendTime := time.Now()
		if err != nil {
			c.metricsMgr.TxErrCounter.Inc(1)
		} else {
			c.metricsMgr.SentTxCounter.Inc(1)
			c.metricsMgr.ReqMeter.Mark(1)
			go func() {
				c.txStartCh <- &txInfo{
					Hash: hash,
					Time: sendTime,
				}
			}()
		}
	}()
	return c.Client.CreateContract(ctx, from, bytecode, gas)
}

func (c *metricClient) CreatePrivateContract(ctx context.Context, from common.Address, bytecode string, gas *big.Int, privateFor []string) (hash string, err error) {
	defer func() {
		sendTime := time.Now()
		if err != nil {
			c.metricsMgr.TxErrCounter.Inc(1)
		} else {
			c.metricsMgr.SentTxCounter.Inc(1)
			c.metricsMgr.ReqMeter.Mark(1)
			go func() {
				c.txStartCh <- &txInfo{
					Hash: hash,
					Time: sendTime,
				}
			}()
		}
	}()
	return c.Client.CreatePrivateContract(ctx, from, bytecode, gas, privateFor)
}

func (c *metricClient) SendRawTransaction(ctx context.Context, tx *types.Transaction) (err error) {
	defer func() {
		sendTime := time.Now()
		if err != nil {
			c.metricsMgr.TxErrCounter.Inc(1)
		} else {
			c.metricsMgr.SentTxCounter.Inc(1)
			c.metricsMgr.ReqMeter.Mark(1)
			go func() {
				c.txStartCh <- &txInfo{
					Hash: tx.Hash().String(),
					Time: sendTime,
				}
			}()
		}
	}()
	return c.Client.SendRawTransaction(ctx, tx)
}

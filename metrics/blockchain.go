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
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"

	"github.com/getamis/istanbul-tools/client"
	"github.com/getamis/istanbul-tools/container"
)

type StopSnapshot func()

type metricsManager struct {
	registry *DefaultRegistry

	SentTxCounter     *Counter
	TxErrCounter      *Counter
	ExcutedTxCounter  *Counter
	UnknownTxCounter  *Counter
	ReqMeter          *Meter
	RespMeter         *Meter
	TxLatencyTimer    *Timer
	BlockPeriodTimer  *Timer
	BlockLatencyTimer *Timer
}

func newMetricsManager() *metricsManager {
	r := NewRegistry()
	return &metricsManager{
		registry:          r,
		SentTxCounter:     r.NewCounter("tx/sent"),
		TxErrCounter:      r.NewCounter("tx/error"),
		ExcutedTxCounter:  r.NewCounter("tx/excuted"),
		UnknownTxCounter:  r.NewCounter("tx/unknown"),
		ReqMeter:          r.NewMeter("tx/rps"),
		RespMeter:         r.NewMeter("tx/tps/response"),
		TxLatencyTimer:    r.NewTimer("tx/latency"),
		BlockPeriodTimer:  r.NewTimer("block/period"),
		BlockLatencyTimer: r.NewTimer("block/latency"),
	}
}

func (m *metricsManager) Export() {
	m.registry.Export()
}

func (m *metricsManager) SnapshotMeter(meters []*Meter, d time.Duration) StopSnapshot {
	stop := make(chan struct{})
	stopFn := func() {
		close(stop)
	}

	go func() {
		ticker := time.NewTicker(d)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				for _, metric := range meters {
					snapshot := metric.Snapshot()
					his := m.registry.NewHistogram(fmt.Sprintf("%s/histogram", metric.Name()))
					his.Update(int64(snapshot.Rate1()))
				}
			case <-stop:
				return
			}
		}
	}()
	return stopFn
}

// --------------------------------------------------------------------------------------------------

type metricChain struct {
	container.Blockchain

	eths      []container.Ethereum
	headCh    chan *ethtypes.Header
	headSubs  []ethereum.Subscription
	txStartCh chan *txInfo
	txDoneCh  chan *txInfo

	metricsMgr   *metricsManager
	stopSnapshot StopSnapshot

	wg   sync.WaitGroup
	quit chan struct{}
}

func NewMetricChain(blockchain container.Blockchain) container.Blockchain {
	if blockchain == nil {
		return nil
	}
	mc := &metricChain{
		Blockchain: blockchain,
		headCh:     make(chan *ethtypes.Header, 1000),
		txStartCh:  make(chan *txInfo, 10000),
		txDoneCh:   make(chan *txInfo, 10000),
		quit:       make(chan struct{}),
		metricsMgr: newMetricsManager(),
	}
	mc.eths = mc.getMetricEthereum(mc.Blockchain.Validators())
	return mc
}

func (mc *metricChain) AddValidators(numOfValidators int) ([]container.Ethereum, error) {
	vals, err := mc.Blockchain.AddValidators(numOfValidators)
	if err != nil {
		return nil, err
	}
	mc.eths = mc.getMetricEthereum(vals)
	return mc.eths, nil
}

func (mc *metricChain) RemoveValidators(candidates []container.Ethereum, t time.Duration) error {
	err := mc.Blockchain.RemoveValidators(candidates, t)
	mc.eths = mc.getMetricEthereum(mc.Blockchain.Validators())
	return err
}

func (mc *metricChain) Start(strong bool) error {
	err := mc.Blockchain.Start(strong)
	if err != nil {
		return err
	}

	for _, eth := range mc.eths {
		cli := eth.NewClient()
		sub, err := cli.SubscribeNewHead(context.Background(), mc.headCh)
		if err != nil {
			log.Error("Failed to subscribe new head", "err", err)
			return err
		}
		mc.headSubs = append(mc.headSubs, sub)
	}
	snapshotMeters := []*Meter{mc.metricsMgr.ReqMeter, mc.metricsMgr.RespMeter}
	mc.stopSnapshot = mc.metricsMgr.SnapshotMeter(snapshotMeters, 1*time.Minute)

	mc.wg.Add(2)
	go mc.handleNewHeadEvent()
	go mc.updateTxInfo()
	return nil
}

func (mc *metricChain) Stop(strong bool) error {
	close(mc.quit)
	for _, sub := range mc.headSubs {
		sub.Unsubscribe()
	}
	mc.wg.Wait()
	mc.stopSnapshot()
	mc.Export()
	return mc.Blockchain.Stop(strong)
}

func (mc *metricChain) Validators() []container.Ethereum {
	return mc.eths
}

func (mc *metricChain) getMetricEthereum(eths []container.Ethereum) []container.Ethereum {
	meths := make([]container.Ethereum, len(eths))
	for i, eth := range eths {
		meths[i] = &metricEthereum{
			Ethereum:   eth,
			txStartCh:  mc.txStartCh,
			metricsMgr: mc.metricsMgr,
		}
	}
	return meths
}

func (mc *metricChain) handleNewHeadEvent() {
	defer mc.wg.Done()

	mutex := sync.Mutex{}
	var preBlockTime = time.Now()
	handledHeads := map[string]*ethtypes.Header{}
	for {
		select {
		case header := <-mc.headCh:
			now := time.Now()
			go func(header *ethtypes.Header, now time.Time) {
				log.Info("New head", "number", header.Number.Int64(), "hash", header.Hash().TerminalString(), "time", header.Time)
				hash := header.Hash().String()
				// lock hash first
				var wasHandled bool
				var preBlock *ethtypes.Header

				mutex.Lock()
				_, wasHandled = handledHeads[hash]
				if !wasHandled {
					handledHeads[hash] = header
				}
				preBlock, _ = handledHeads[header.ParentHash.String()]
				mutex.Unlock()

				if wasHandled {
					return
				}

				var blockPeriod int64
				if header.Number.Int64() > 2 && preBlock != nil {
					blockPeriod = new(big.Int).Sub(header.Time, preBlock.Time).Int64()
					mc.metricsMgr.BlockPeriodTimer.Update(time.Duration(blockPeriod) * time.Second)
				}
				mc.metricsMgr.BlockLatencyTimer.Update(now.Sub(preBlockTime))
				preBlockTime = now

				// get block
				blockCh := make(chan *ethtypes.Block, len(mc.eths))
				ctx, cancel := context.WithCancel(context.Background())
				for _, eth := range mc.eths {
					cli := eth.NewClient()
					go getBlock(ctx, cli, header.Hash(), blockCh)
				}

				// wait for right block
				var headBlock *ethtypes.Block
				for i := 0; i < len(mc.eths); i++ {
					headBlock = <-blockCh
					if headBlock != nil {
						break
					}
				}
				// cancel other requests
				cancel()

				mc.metricsMgr.ExcutedTxCounter.Inc(int64(len(headBlock.Transactions())))
				mc.metricsMgr.RespMeter.Mark(int64(len(headBlock.Transactions())))

				// update tx info
				for _, tx := range headBlock.Transactions() {
					go func() {
						mc.txDoneCh <- &txInfo{
							Hash: tx.Hash().String(),
							Time: now,
						}
					}()
				}
			}(header, now)
		case <-mc.quit:
			return
		}
	}
}

func (mc *metricChain) updateTxInfo() {
	defer mc.wg.Done()

	// TODO: the completed tx should be deleted from map
	// given large space is workaround beacause the some problem between deleting and updating map
	txStartMap := make(map[string]time.Time, 0)
	txDoneMap := make(map[string]time.Time, 0)
	defer func() {
		// TODO: debug metric to check incomplete tx
		//	for _ = range txStartMap {
		//		mc.metricsMgr.UnknownTxCounter.Inc(1)
		//	}
		//	for _ = range txDoneMap {
		//		mc.metricsMgr.UnknownTxCounter.Inc(1)
		//	}
	}()

	updateTxStart := func(hash string, startTime time.Time) {
		if done, ok := txDoneMap[hash]; ok {
			mc.metricsMgr.TxLatencyTimer.Update(done.Sub(startTime))
			return
			//delete(txDoneMap, hash)
		}
		txStartMap[hash] = startTime
	}

	updateTxDone := func(hash string, doneTime time.Time) {
		if start, ok := txStartMap[hash]; ok {
			mc.metricsMgr.TxLatencyTimer.Update(doneTime.Sub(start))
			return
			//delete(txStartMap, hash)
		}
		txDoneMap[hash] = doneTime
	}

	for {
		select {
		case txStart := <-mc.txStartCh:
			updateTxStart(txStart.Hash, txStart.Time)
		case txDone := <-mc.txDoneCh:
			updateTxDone(txDone.Hash, txDone.Time)
		case <-mc.quit:
			// clear tx start
		TX_START:
			for {
				select {
				case txStart := <-mc.txStartCh:
					updateTxStart(txStart.Hash, txStart.Time)
				default:
					break TX_START
				}

			}
			// clear tx done
		TX_DONE:
			for {
				select {
				case txDone := <-mc.txDoneCh:
					updateTxDone(txDone.Hash, txDone.Time)
				default:
					break TX_DONE
				}
			}
			return
		}
	}
}

func getBlock(ctx context.Context, cli client.Client, hash common.Hash, blockCh chan<- *ethtypes.Block) {
	resp := make(chan *ethtypes.Block)
	go func() {
		block, err := cli.BlockByHash(ctx, hash)
		if err != nil {
			resp <- nil
		}
		resp <- block
	}()

	select {
	case <-ctx.Done():
		// Wait for client.BlockByHash
		<-resp
		// someone cancelled the request
		blockCh <- nil
	case r := <-resp:
		blockCh <- r
	}
}

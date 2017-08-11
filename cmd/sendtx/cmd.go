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

package main

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"
	"math/rand"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/urfave/cli"
)

const (
	GAS_PRICE = 20000000000
	GAS_LIMIT = 22000 // the gas of ether tx should be 21000
	// admin
	private = "c921c91aa4c5f9886bd1e084a848e7e564644d5cc2f265beebb0d51bd251b7e7"
	public  = "1f0a3201bfa623be518eb7fb3742385f6f42e2e09a83e417f51f937c19c5388fc99e9472f5ac7e6c8baebcdf96719312d22fd8783d071c2e71bf713993ab2284"
	address = "507a198251ed29f421bd6cf667596f750d9b14ec"
)

var (
	prepareCommand = cli.Command{
		Action:    prepare,
		Name:      "prepare",
		Usage:     "Prepare accounts to nodes",
		ArgsUsage: "<prepare accounts>",
		Flags: []cli.Flag{
			NodeAddrFlag,
			AdminFlag,
			AccountsFlag,
		},
		Description: `To prepare accounts`,
	}

	batchCommand = cli.Command{
		Action:    batch,
		Name:      "batch",
		Usage:     "Batch send txs",
		ArgsUsage: "<batch send>",
		Flags: []cli.Flag{
			NodeAddrFlag,
			AdminFlag,
			TxsCountFlag,
			SendPeriodFlag,
		},
		Description: `To batch send txs`,
	}

	// target
	toAddr = common.HexToAddress("0xcfe3b52266134683d6e9ee765d735431e35c7a54")
)

func prepare(ctx *cli.Context) error {
	addr := ctx.String(NodeAddrFlag.Name)
	if addr == "" {
		return cli.NewExitError("Must provide node address", 0)
	}

	key := ctx.String(AdminFlag.Name)
	if key == "" {
		return cli.NewExitError("Must provide admin's private key", 0)
	}

	num := ctx.Int(AccountsFlag.Name)

	conn, err := ethclient.Dial(addr)
	if err != nil {
		return cli.NewExitError(fmt.Sprintf("Failed to connect to ethclient:: %v", err), 1)
	}

	admin, err := AccountFromECDSA(key)
	if err != nil {
		return cli.NewExitError(fmt.Sprintf("Failed to read private key: %v", err), 1)
	}

	nonce, err := conn.NonceAt(context.Background(), admin.Address(), nil)
	if err != nil {
		return cli.NewExitError(fmt.Sprintf("Failed to get nonce:: %v", err), 1)
	}
	fmt.Println("nonce:", nonce)

	// generate accounts
	prepareAccs := []*Account{}
	for i := 0; i < num; i++ {
		a, _ := GenerateAccount()
		prepareAccs = append(prepareAccs, a)
		fmt.Println("account", a)
	}

	amount := big.NewInt(0).Exp(big.NewInt(int64(10)), big.NewInt(int64(20)), nil)
	// give money
	for _, pa := range prepareAccs {
		tx, err := sendEther(admin.PrivateKey(), pa.Address(), amount, nonce)
		if err != nil {
			fmt.Printf("Failed to gen tx, err:%v\n", err)
			continue
		}
		err = sendTx(conn, tx)
		if err != nil {
			fmt.Printf("Failed to send tx, err:%v\n", err)
			continue
		}
		nonce++
	}
	writeAccounts(prepareAccs)
	return nil
}

func sendTx(conn *ethclient.Client, tx *types.Transaction) error {
	err := conn.SendTransaction(context.Background(), tx)
	if err != nil {
		return err
	}
	fmt.Println("tx", tx.Hash().String())
	return nil
}

func sendEther(prv *ecdsa.PrivateKey, to common.Address, amount *big.Int, nonce uint64) (*types.Transaction, error) {
	signer := types.FrontierSigner{}
	tx := types.NewTransaction(nonce, to, amount, big.NewInt(GAS_LIMIT), big.NewInt(GAS_PRICE), []byte{})
	//	return tx, nil
	signedTx, err := types.SignTx(tx, signer, prv)
	if err != nil {
		return nil, err
	}
	return signedTx, nil
}

func batch(ctx *cli.Context) error {
	addr := ctx.String(NodeAddrFlag.Name)
	if addr == "" {
		return cli.NewExitError("Must provide node address", 0)
	}

	key := ctx.String(AdminFlag.Name)
	if key == "" {
		return cli.NewExitError("Must provide admin's private key", 0)
	}

	conn, err := ethclient.Dial(addr)
	if err != nil {
		return cli.NewExitError(fmt.Sprintf("Failed to connect to ethclient:: %v", err), 1)
	}

	admin, err := AccountFromECDSA(key)
	if err != nil {
		return cli.NewExitError(fmt.Sprintf("Failed to read private key: %v", err), 1)
	}

	count := ctx.Int(TxsCountFlag.Name)
	secs := ctx.Int(SendPeriodFlag.Name)

	nonce, err := conn.NonceAt(context.Background(), admin.Address(), nil)
	if err != nil {
		return cli.NewExitError(fmt.Sprintf("Failed to get nonce:: %v", err), 1)
	}
	fmt.Println("nonce:", nonce)

	balance, err := conn.BalanceAt(context.Background(), admin.Address(), nil)
	if err != nil {
		return cli.NewExitError(fmt.Sprintf("Failed to get nonce:: %v", err), 1)
	}
	fmt.Println("balance:", balance)

	// give money
	batchSend(conn, admin.PrivateKey(), toAddr, nonce, count, time.Duration(secs)*time.Second)
	return nil
}

func batchSend(conn *ethclient.Client, prv *ecdsa.PrivateKey, to common.Address, nonce uint64, count int, period time.Duration) {
	round := 0
	timer := time.NewTimer(period)

	for {
		select {
		case <-timer.C:
			return
		default:
			fmt.Println("round", round)
			for i := 0; i < count; i++ {
				tx, err := sendEther(prv, to, common.Big1, nonce)
				if err != nil {
					fmt.Printf("Failed to gen tx, err:%v\n", err)
					continue
				}
				err = sendTx(conn, tx)
				if err != nil {
					fmt.Printf("Failed to send tx, err:%v\n", err)
					continue
				}
				nonce++
			}
			round++
			<-time.After(time.Duration(20+rand.Int()%10) * time.Second)
		}
	}
}

func writeAccounts(accs []*Account) {
	keys := []string{}
	for _, acc := range accs {
		keys = append(keys, acc.String())
	}

	b, err := json.Marshal(keys)
	if err != nil {
		fmt.Println("Failed to json Marshal accounts, err:%v", err)
		return
	}
	err = ioutil.WriteFile("accs", b, 0644)
	if err != nil {
		fmt.Println("Failed to write file, err:%v", err)
		return
	}
	return
}

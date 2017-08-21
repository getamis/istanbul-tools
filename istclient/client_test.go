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

package istclient

import (
	"context"
	"log"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
)

func ExampleStartMining() {
	client, err := Dial("ws://127.0.0.1:49733")
	if err != nil {
		log.Fatal("failed to dial, err:", err)
	}

	err = client.StartMining(context.Background())
	if err != nil {
		log.Fatal("failed to get validators, err:", err)
	}
}

func ExampleProposerValidator() {
	client, err := Dial("http://127.0.0.1:8547")
	if err != nil {
		log.Fatal("failed to dial, err:", err)
	}

	err = client.ProposeValidator(context.Background(), common.HexToAddress("0x6B58EA55d051008822Cf3acd684914c83aF2f588"), true)
	if err != nil {
		log.Fatal("failed to get validators, err:", err)
	}
}

func ExampleGetValidators() {
	client, err := Dial("ws://127.0.0.1:53257")
	if err != nil {
		log.Fatal("failed to dial, err:", err)
	}

	addrs, err := client.GetValidators(context.Background(), nil)
	if err != nil {
		log.Fatal("failed to get validators, err:", err)
	}
	for _, addr := range addrs {
		log.Println("address:", addr.Hex())
	}
}

func ExampleAdminPeers() {
	client, err := Dial("ws://127.0.0.1:62975")
	if err != nil {
		log.Fatal("failed to dial, err:", err)
	}

	peersInfo, err := client.AdminPeers(context.Background())
	if err != nil {
		log.Fatal("failed to get validators, err:", err)
	}

	fmt.Println("connected peer length:", len(peersInfo))
	for _, peer := range peersInfo {
		fmt.Println("address:", peer)
	}
}

func ExampleAddPeer() {
	client, err := Dial("ws://127.0.0.1:62975")
	if err != nil {
		log.Fatal("failed to dial, err:", err)
	}

	err = client.AddPeer(context.Background(), "enode://ad5b4b201cc0ef5cd6ce27e32c223d1852a8b7d6069de5c3c597601e94841a5811a354261726da7b8f851e9042d5aeaed580dbb7493d22a5d922206dce3ccdb8@192.168.99.100:63040?discport=0")
	if err != nil {
		log.Fatal("failed to get validators, err:", err)
	}
}

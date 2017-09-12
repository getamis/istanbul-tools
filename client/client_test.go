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

package client

import (
	"context"

	"github.com/ethereum/go-ethereum/common"
)

func ExampleStartMining() {
	url := "ws://127.0.0.1:49733"
	client, err := Dial(url)
	if err != nil {
		log.Error("Failed to dial", "url", url, "err", err)
		return
	}

	err = client.StartMining(context.Background())
	if err != nil {
		log.Error("Failed to get validators", "err", err)
		return
	}
}

func ExampleProposerValidator() {
	url := "http://127.0.0.1:8547"
	client, err := Dial(url)
	if err != nil {
		log.Error("Failed to dial", "url", url, "err", err)
		return
	}

	err = client.ProposeValidator(context.Background(), common.HexToAddress("0x6B58EA55d051008822Cf3acd684914c83aF2f588"), true)
	if err != nil {
		log.Error("Failed to get validators", "err", err)
		return
	}
}

func ExampleGetValidators() {
	url := "ws://127.0.0.1:53257"
	client, err := Dial(url)
	if err != nil {
		log.Error("Failed to dial", "url", url, "err", err)
		return
	}

	addrs, err := client.GetValidators(context.Background(), nil)
	if err != nil {
		log.Error("Failed to get validators", "err", err)
		return
	}
	for _, addr := range addrs {
		log.Info("address", "hex", addr.Hex())
	}
}

func ExampleAdminPeers() {
	url := "ws://127.0.0.1:62975"
	client, err := Dial(url)
	if err != nil {
		log.Error("Failed to dial", "url", url, "err", err)
		return
	}

	peersInfo, err := client.AdminPeers(context.Background())
	if err != nil {
		log.Error("Failed to get validators", "err", err)
		return
	}

	log.Info("Peers connected", "peers", peersInfo, "len", len(peersInfo))
}

func ExampleAddPeer() {
	url := "ws://127.0.0.1:62975"
	client, err := Dial(url)
	if err != nil {
		log.Error("Failed to dial", "url", url, "err", err)
		return
	}

	err = client.AddPeer(context.Background(), "enode://ad5b4b201cc0ef5cd6ce27e32c223d1852a8b7d6069de5c3c597601e94841a5811a354261726da7b8f851e9042d5aeaed580dbb7493d22a5d922206dce3ccdb8@192.168.99.100:63040?discport=0")
	if err != nil {
		log.Error("Failed to get validators", "err", err)
		return
	}
}

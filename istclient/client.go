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
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/rpc"
)

type Client struct {
	c *rpc.Client
}

func Dial(rawurl string) (*Client, error) {
	c, err := rpc.Dial(rawurl)
	if err != nil {
		return nil, err
	}
	return &Client{c}, nil
}

func (c *Client) Close() {
	c.c.Close()
}

// ----------------------------------------------------------------------------

func (ic *Client) AddPeer(ctx context.Context, nodeURL string) error {
	var r bool
	// TODO: Result needs to be verified
	// The response data type are bytes, but we cannot parse...
	err := ic.c.CallContext(ctx, &r, "admin_addPeer", nodeURL)
	if err != nil {
		return err
	}
	return err
}

func (ic *Client) AdminPeers(ctx context.Context) error {
	var r bool
	// TODO: Result needs to be verified
	// The response data type are bytes, but we cannot parse...
	err := ic.c.CallContext(ctx, &r, "admin_peers")
	if err != nil {
		return err
	}
	return err
}

// ----------------------------------------------------------------------------

func (ic *Client) StartMining(ctx context.Context) error {
	var r []byte
	// TODO: Result needs to be verified
	// The response data type are bytes, but we cannot parse...
	err := ic.c.CallContext(ctx, &r, "miner_start", nil)
	if err != nil {
		return err
	}
	return err
}

// ----------------------------------------------------------------------------

func (ic *Client) ProposeValidator(ctx context.Context, address common.Address, auth bool) error {
	var r []byte
	// TODO: Result needs to be verified with other method
	// The response data type are bytes, but we cannot parse...
	err := ic.c.CallContext(ctx, &r, "istanbul_propose", address, auth)
	if err != nil {
		return ethereum.NotFound
	}
	return err
}

func (ic *Client) GetValidators(ctx context.Context, blockNumbers *big.Int) ([]common.Address, error) {
	var r []common.Address
	err := ic.c.CallContext(ctx, &r, "istanbul_getValidators", toNumArg(blockNumbers))
	if err == nil && r == nil {
		return nil, ethereum.NotFound
	}
	return r, err
}

func toNumArg(number *big.Int) string {
	if number == nil {
		return "latest"
	}
	return hexutil.EncodeBig(number)
}

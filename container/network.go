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

package container

import (
	"context"
	"fmt"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
)

type networkManager struct {
	client *client.Client
	field1 uint8
	field2 uint8
}

func NewNetworkManager(c *client.Client, field1 uint8, field2 uint8) *networkManager {
	n := &networkManager{
		client: c,
		field1: field1,
		field2: field2,
	}

	return n
}

// CreateNetwork returns network id
func (n *networkManager) CreateNetwork(name string) (string, error) {
	ipamConfig := network.IPAMConfig{
		Subnet: fmt.Sprintf("%d.%d.0.0/16", n.field1, n.field2),
	}
	ipam := &network.IPAM{
		Config: []network.IPAMConfig{ipamConfig},
	}
	r, e := n.client.NetworkCreate(context.Background(), name, types.NetworkCreate{
		IPAM: ipam,
	})
	if e != nil {
		return "", e
	} else {
		return r.ID, nil
	}
}

func (n *networkManager) RemoveNetwork(id string) error {
	return n.client.NetworkRemove(context.Background(), id)
}

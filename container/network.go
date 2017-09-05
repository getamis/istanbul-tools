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
	"errors"
	"fmt"
	"net"
	"sync"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
)

const (
	FirstOctet  = 172
	SecondOctet = 19
	NetworkName = "testnet"
)

type DockerNetwork struct {
	client  *client.Client
	id      string
	name    string
	ipv4Net *net.IPNet

	mutex   sync.Mutex
	ipIndex net.IP
}

func NewDockerNetwork() (*DockerNetwork, error) {
	c, err := client.NewEnvClient()
	if err != nil {
		return nil, err
	}

	for i := SecondOctet; i < 256; i++ {
		// IP xxx.xxx.0.1 is reserved for docker network gateway
		ipv4Addr, ipv4Net, err := net.ParseCIDR(fmt.Sprintf("%d.%d.0.1/16", FirstOctet, i))
		networkName := fmt.Sprintf("%s_%d_%d", NetworkName, FirstOctet, i)
		if err != nil {
			return nil, err
		}
		network := &DockerNetwork{
			client:  c,
			name:    networkName,
			ipv4Net: ipv4Net,
			ipIndex: ipv4Addr,
		}
		if err = network.create(); err != nil {
			fmt.Printf("Failed to create network and retry, err:%v\n", err)
		} else {
			return network, nil
		}
	}

	return nil, err
}

// create creates a docker network with given subnet
func (n *DockerNetwork) create() error {
	ipamConfig := network.IPAMConfig{
		Subnet: n.ipv4Net.String(),
	}
	ipam := &network.IPAM{
		Config: []network.IPAMConfig{ipamConfig},
	}

	r, err := n.client.NetworkCreate(context.Background(), n.name, types.NetworkCreate{
		IPAM: ipam,
	})
	if err != nil {
		return err
	}
	n.id = r.ID
	return nil
}

func (n *DockerNetwork) ID() string {
	return n.id
}

func (n *DockerNetwork) Name() string {
	return n.name
}

func (n *DockerNetwork) Remove() error {
	return n.client.NetworkRemove(context.Background(), n.id)
}

func (n *DockerNetwork) GetFreeIPAddrs(num int) ([]net.IP, error) {
	n.mutex.Lock()
	defer n.mutex.Unlock()

	ips := make([]net.IP, 0)
	for i := 0; i < num; i++ {
		ip := dupIP(n.ipIndex)
		for j := len(ip) - 1; j >= 0; j-- {
			ip[j]++
			if ip[j] > 0 {
				break
			}
		}

		if !n.ipv4Net.Contains(ip) {
			break
		}
		ips = append(ips, ip)
		n.ipIndex = ip
	}

	if len(ips) != num {
		return nil, errors.New("Insufficient IP address.")
	}
	return ips, nil
}

func dupIP(ip net.IP) net.IP {
	// To save space, try and only use 4 bytes
	if x := ip.To4(); x != nil {
		ip = x
	}
	dup := make(net.IP, len(ip))
	copy(dup, ip)
	return dup
}

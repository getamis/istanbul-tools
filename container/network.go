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
	"time"

	"github.com/docker/docker/api/types"
	//"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
)

const (
	DefaultNetworkName = "bridge"
	networkNamePrefix  = "testnet"
)

type DockerNetwork struct {
	client  *client.Client
	id      string
	name    string
	ipv4Net *net.IPNet
	gateway string
	usedIPs map[string]bool

	mutex   sync.Mutex
	ipIndex net.IP
}

func GetDefaultNetwork() (*DockerNetwork, error) {
	c, err := client.NewEnvClient()
	if err != nil {
		return nil, err
	}

	network := &DockerNetwork{
		client:  c,
		name:    DefaultNetworkName,
		usedIPs: make(map[string]bool, 0),
	}

	if err := network.init(); err != nil {
		return nil, err
	}

	return network, nil
}

func NewDockerNetwork() (*DockerNetwork, error) {
	c, err := client.NewEnvClient()
	if err != nil {
		return nil, err
	}

	network := &DockerNetwork{
		client:  c,
		usedIPs: make(map[string]bool, 0),
	}

	if err := network.create(); err != nil {
		return nil, err
	}

	if err := network.init(); err != nil {
		return nil, err
	}

	return network, nil
}

// create creates a docker network with given subnet
func (n *DockerNetwork) create() error {
	n.name = fmt.Sprintf("%s%d", networkNamePrefix, time.Now().Unix())
	//	ipamConfig := network.IPAMConfig{
	//		Subnet:  "172.21.0.0/16",
	//		Gateway: "172.21.0.1",
	//	}
	//	ipam := &network.IPAM{
	//		Config: []network.IPAMConfig{ipamConfig},
	//	}
	r, err := n.client.NetworkCreate(context.Background(), n.name, types.NetworkCreate{
	//IPAM: ipam,
	//Driver: "bridge",
	})
	if err != nil {
		return err
	}
	n.id = r.ID
	return nil
}

func (n *DockerNetwork) init() error {
	res, err := n.client.NetworkInspect(context.Background(), n.name, types.NetworkInspectOptions{})
	if err != nil {
		return err
	}

	if len(res.IPAM.Config) == 0 {
		return errors.New("Invalid network config.")
	}
	ip, ipv4Net, err := net.ParseCIDR(res.IPAM.Config[0].Subnet)
	if err != nil {
		return err
	}

	n.ipv4Net = ipv4Net
	n.ipIndex = ip

	// get used IP addresses
	if res.IPAM.Config[0].Gateway != "" {
		n.gateway = res.IPAM.Config[0].Gateway
	} else {
		// xxx.xxx.0.1 is reserved for default Gateway IP
		n.gateway = net.IPv4(n.ipv4Net.IP[0], n.ipv4Net.IP[1], 0, 1).String()
	}
	fmt.Println("n.gateway", n.gateway)
	n.usedIPs[n.gateway] = true

	for _, ep := range res.Containers {
		ip, _, err := net.ParseCIDR(ep.IPv4Address)
		if err != nil {
			return err
		}
		n.usedIPs[ip.String()] = true
	}
	return nil
}

func (n *DockerNetwork) ID() string {
	return n.id
}

func (n *DockerNetwork) Name() string {
	return n.name
}

func (n *DockerNetwork) Network() string {
	return n.ipv4Net.String()
}

func (n *DockerNetwork) Remove() error {
	return n.client.NetworkRemove(context.Background(), n.id)
}

func (n *DockerNetwork) GetFreeIPAddrs(num int) ([]net.IP, error) {
	n.mutex.Lock()
	defer n.mutex.Unlock()

	ips := make([]net.IP, 0)
	for len(ips) < num && n.ipv4Net.Contains(n.ipIndex) {
		ip := dupIP(n.ipIndex)
		for j := len(ip) - 1; j >= 0; j-- {
			ip[j]++
			if ip[j] > 0 {
				break
			}
		}
		n.ipIndex = ip
		if _, ok := n.usedIPs[ip.String()]; !ok {
			ips = append(ips, ip)
		}
	}

	if len(ips) != num {
		return nil, errors.New("Insufficient IP address.")
	}
	fmt.Println("get ips", ips)
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

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
	"github.com/getamis/go-ethereum/cmd/utils"
)

type Option func(*ethereum)

func ImageName(imageName string) Option {
	return func(eth *ethereum) {
		eth.imageName = imageName
	}
}

func HostName(hostName string) Option {
	return func(eth *ethereum) {
		eth.hostName = hostName
	}
}

func HostDataDir(path string) Option {
	return func(eth *ethereum) {
		eth.hostDataDir = path
	}
}

func Logging(enabled bool) Option {
	return func(eth *ethereum) {
		eth.logging = enabled
	}
}

// ----------------------------------------------------------------------------

func DataDir(dir string) Option {
	return func(eth *ethereum) {
		eth.flags = append(eth.flags, "--"+utils.DataDirFlag.Name)
		eth.flags = append(eth.flags, dir)
		eth.dataDir = dir
	}
}

func Etherbase(etherbase string) Option {
	return func(eth *ethereum) {
		eth.flags = append(eth.flags, "--"+utils.EtherbaseFlag.Name)
		eth.flags = append(eth.flags, etherbase)
	}
}

func Identity(id string) Option {
	return func(eth *ethereum) {
		eth.flags = append(eth.flags, "--"+utils.IdentityFlag.Name)
		eth.flags = append(eth.flags, id)
	}
}

func IPC(enabled bool) Option {
	return func(eth *ethereum) {
		if !enabled {
			eth.flags = append(eth.flags, "--"+utils.IPCDisabledFlag.Name)
		}
	}
}

func KeyStore(dir string) Option {
	return func(eth *ethereum) {
		eth.flags = append(eth.flags, "--"+utils.KeyStoreDirFlag.Name)
		eth.flags = append(eth.flags, dir)
	}
}

func NetworkID(networkID string) Option {
	return func(eth *ethereum) {
		eth.flags = append(eth.flags, "--"+utils.NetworkIdFlag.Name)
		eth.flags = append(eth.flags, networkID)
	}
}

func Mine() Option {
	return func(eth *ethereum) {
		eth.flags = append(eth.flags, "--"+utils.MiningEnabledFlag.Name)
	}
}

func NAT(nat string) Option {
	return func(eth *ethereum) {
		eth.flags = append(eth.flags, "--"+utils.NATFlag.Name)
		eth.flags = append(eth.flags, nat)
	}
}

func NodeKey(nodekey string) Option {
	return func(eth *ethereum) {
		eth.flags = append(eth.flags, "--"+utils.NodeKeyFileFlag.Name)
		eth.flags = append(eth.flags, nodekey)
	}
}

func NodeKeyHex(hex string) Option {
	return func(eth *ethereum) {
		eth.flags = append(eth.flags, "--"+utils.NodeKeyHexFlag.Name)
		eth.flags = append(eth.flags, hex)
	}
}

func NoDiscover() Option {
	return func(eth *ethereum) {
		eth.flags = append(eth.flags, "--"+utils.NoDiscoverFlag.Name)
	}
}

func Port(port string) Option {
	return func(eth *ethereum) {
		eth.flags = append(eth.flags, "--"+utils.ListenPortFlag.Name)
		eth.flags = append(eth.flags, port)
	}
}

func RPC() Option {
	return func(eth *ethereum) {
		eth.flags = append(eth.flags, "--"+utils.RPCEnabledFlag.Name)
	}
}

func RPCAddress(address string) Option {
	return func(eth *ethereum) {
		eth.flags = append(eth.flags, "--"+utils.RPCListenAddrFlag.Name)
		eth.flags = append(eth.flags, address)
	}
}

func RPCAPI(apis string) Option {
	return func(eth *ethereum) {
		eth.flags = append(eth.flags, "--"+utils.RPCApiFlag.Name)
		eth.flags = append(eth.flags, apis)
	}
}

func RPCPort(port string) Option {
	return func(eth *ethereum) {
		eth.flags = append(eth.flags, "--"+utils.RPCPortFlag.Name)
		eth.flags = append(eth.flags, port)
	}
}

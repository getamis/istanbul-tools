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
	"crypto/ecdsa"
	"fmt"

	"github.com/ethereum/go-ethereum/cmd/utils"
)

type Option func(*ethereum)

func ImageRepository(repository string) Option {
	return func(eth *ethereum) {
		eth.imageRepository = repository
	}
}

func ImageTag(tag string) Option {
	return func(eth *ethereum) {
		eth.imageTag = tag
	}
}

func HostName(hostName string) Option {
	return func(eth *ethereum) {
		eth.hostName = hostName
	}
}

func HostDataDir(path string) Option {
	return func(eth *ethereum) {
		eth.dataDir = path
	}
}

func HostPort(port int) Option {
	return func(eth *ethereum) {
		eth.port = fmt.Sprintf("%d", port)
	}
}

func HostRPCPort(port int) Option {
	return func(eth *ethereum) {
		eth.rpcPort = fmt.Sprintf("%d", port)
	}
}

func HostWebSocketPort(port int) Option {
	return func(eth *ethereum) {
		eth.wsPort = fmt.Sprintf("%d", port)
	}
}

func Logging(enabled bool) Option {
	return func(eth *ethereum) {
		eth.logging = enabled
	}
}

// ----------------------------------------------------------------------------

func Key(key *ecdsa.PrivateKey) Option {
	return func(eth *ethereum) {
		eth.key = key
	}
}

func DataDir(dir string) Option {
	return func(eth *ethereum) {
		utils.DataDirFlag.Value = utils.DirectoryString{
			Value: dir,
		}
		eth.flags = append(eth.flags, "--"+utils.DataDirFlag.Name)
		eth.flags = append(eth.flags, dir)

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

func Port(port int) Option {
	return func(eth *ethereum) {
		utils.ListenPortFlag.Value = port
		eth.flags = append(eth.flags, "--"+utils.ListenPortFlag.Name)
		eth.flags = append(eth.flags, fmt.Sprintf("%d", port))
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

func RPCPort(port int) Option {
	return func(eth *ethereum) {
		utils.RPCPortFlag.Value = port
		eth.flags = append(eth.flags, "--"+utils.RPCPortFlag.Name)
		eth.flags = append(eth.flags, fmt.Sprintf("%d", port))
	}
}

func WebSocket() Option {
	return func(eth *ethereum) {
		eth.flags = append(eth.flags, "--"+utils.WSEnabledFlag.Name)
	}
}

func WebSocketAddress(address string) Option {
	return func(eth *ethereum) {
		eth.flags = append(eth.flags, "--"+utils.WSListenAddrFlag.Name)
		eth.flags = append(eth.flags, address)
	}
}

func WebSocketAPI(apis string) Option {
	return func(eth *ethereum) {
		eth.flags = append(eth.flags, "--"+utils.WSApiFlag.Name)
		eth.flags = append(eth.flags, apis)
	}
}

func WebSocketPort(port int) Option {
	return func(eth *ethereum) {
		utils.WSPortFlag.Value = port
		eth.flags = append(eth.flags, "--"+utils.WSPortFlag.Name)
		eth.flags = append(eth.flags, fmt.Sprintf("%d", port))
	}
}

func WebSocketOrigin(origins string) Option {
	return func(eth *ethereum) {
		eth.flags = append(eth.flags, "--"+utils.WSAllowedOriginsFlag.Name)
		eth.flags = append(eth.flags, origins)
	}
}

func Verbosity(verbosity int) Option {
	return func(eth *ethereum) {
		eth.flags = append(eth.flags, "--verbosity")
		eth.flags = append(eth.flags, fmt.Sprintf("%d", verbosity))
	}
}

func FaultyMode(mode int) Option {
	return func(eth *ethereum) {
		eth.flags = append(eth.flags, "--istanbul.faultymode")
		eth.flags = append(eth.flags, fmt.Sprintf("%d", mode))
	}
}

func SyncMode(mode string) Option {
	return func(eth *ethereum) {
		eth.flags = append(eth.flags, "--"+utils.SyncModeFlag.Name)
		eth.flags = append(eth.flags, mode)
	}
}

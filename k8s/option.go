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

package k8s

import "fmt"

type Option func(*ethereum)

func ImageRepository(repository string) Option {
	return func(eth *ethereum) {
		eth.args = append(eth.args, fmt.Sprintf("image.respository=%s", repository))
	}
}

func ImageTag(tag string) Option {
	return func(eth *ethereum) {
		eth.args = append(eth.args, fmt.Sprintf("image.tag=%s", tag))
	}
}

// ----------------------------------------------------------------------------

func Name(name string) Option {
	return func(eth *ethereum) {
		eth.name = name
		eth.args = append(eth.args, fmt.Sprintf("nameOverride=%s", name))
	}
}

func ServiceType(serviceType string) Option {
	return func(eth *ethereum) {
		eth.args = append(eth.args, fmt.Sprintf("service.type=%s", serviceType))
	}
}

func IPAddress(ip string) Option {
	return func(eth *ethereum) {
		eth.args = append(eth.args, fmt.Sprintf("service.staticIP=%s", ip))
	}
}

func NetworkID(networkID string) Option {
	return func(eth *ethereum) {
		eth.args = append(eth.args, fmt.Sprintf("ethereum.networkID=%s", networkID))
	}
}

func Mine() Option {
	return func(eth *ethereum) {
		eth.args = append(eth.args, "ethereum.mining.enabled=true")
	}
}

func NodeKeyHex(hex string) Option {
	return func(eth *ethereum) {
		eth.nodekey = hex
		eth.args = append(eth.args, fmt.Sprintf("ethereum.nodekey.hex=%s", hex))
	}
}

func TxPoolSize(size int) Option {
	return func(eth *ethereum) {
		eth.args = append(eth.args, fmt.Sprintf("benchmark.txpool.globalslots=%d", size))
		eth.args = append(eth.args, fmt.Sprintf("benchmark.txpool.accountslots=%d", size))
		eth.args = append(eth.args, fmt.Sprintf("benchmark.txpool.globalqueue=%d", size))
		eth.args = append(eth.args, fmt.Sprintf("benchmark.txpool.accountqueue=%d", size))
	}
}

func Verbosity(verbosity int) Option {
	return func(eth *ethereum) {
		eth.args = append(eth.args, fmt.Sprintf("ethereum.verbosity=%d", verbosity))
	}
}

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
	"fmt"
	"path/filepath"
	"testing"

	"github.com/getamis/istanbul-tools/core"
)

func TestEthereumContainer(t *testing.T) {
	keys := core.GenerateClusterKeys(1)
	envs := core.SetupEnv(keys)
	defer core.Teardown(envs)
	err := core.SetupNodes(envs)
	if err != nil {
		t.Error(err)
	}

	for _, env := range envs {
		geth := NewEthereum(
			env.Client,
			ImageName("quay.io/maicoin/ottoman_geth:istanbul_develop"),
			HostDataDir(env.DataDir),
			DataDir("/data"),
			Port(fmt.Sprintf("%d", env.P2PPort)),
			RPC(),
			RPCAddress("0.0.0.0"),
			RPCAPI("eth,net,web3,personal"),
			RPCPort(fmt.Sprintf("%d", env.RpcPort)),
			Logging(true),
		)

		err := geth.Init(filepath.Join(env.DataDir, core.GenesisJson))
		if err != nil {
			t.Error(err)
		}

		err = geth.Start()
		if err != nil {
			t.Error(err)
		}

		if !geth.Running() {
			t.Error("geth should be running")
		}

		err = geth.Stop()
		if err != nil {
			t.Error(err)
		}
	}
}

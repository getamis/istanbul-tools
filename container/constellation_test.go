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
	"testing"

	"github.com/Consensys/istanbul-tools/docker/service"
	"github.com/docker/docker/client"
	"github.com/phayes/freeport"
)

func TestConstellationContainer(t *testing.T) {
	dockerClient, err := client.NewEnvClient()
	if err != nil {
		t.Error(err)
	}

	dockerNetwork, err := NewDockerNetwork()
	if err != nil {
		t.Error(err)
	}

	ips, err := dockerNetwork.GetFreeIPAddrs(1)
	if err != nil {
		t.Error(err)
	}
	ip := ips[0]

	port := freeport.GetPort()

	ct := NewConstellation(dockerClient,
		CTImageRepository(service.ConstellationDockerImage),
		CTImageTag(service.ConstellationDockerImageTag),
		CTHost(ip, port),
		CTDockerNetworkName(dockerNetwork.Name()),
		CTWorkDir("/data"),
		CTLogging(false),
		CTKeyName("node"),
		CTSocketFilename("node.ipc"),
		CTVerbosity(3),
	)

	_, err = ct.GenerateKey()
	if err != nil {
		t.Error(err)
	}

	err = ct.Start()
	if err != nil {
		t.Error(err)
	}

	if !ct.Running() {
		t.Error("constellation should be running")
	}

	err = ct.Stop()
	if err != nil {
		t.Error(err)
	}

	err = dockerNetwork.Remove()
	if err != nil {
		t.Error(err)
	}
}

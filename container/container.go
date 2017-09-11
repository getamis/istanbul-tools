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
	"net"
	"net/url"
	"os"

	"github.com/docker/docker/client"

	"github.com/getamis/istanbul-tools/log"
)

func (eth *ethereum) Image() string {
	if eth.imageTag == "" {
		return eth.imageRepository + ":latest"
	}
	return eth.imageRepository + ":" + eth.imageTag
}

func (eth *ethereum) ContainerID() string {
	return eth.containerID
}

func (eth *ethereum) Host() string {
	var host string
	daemonHost := os.Getenv("DOCKER_HOST")
	if daemonHost == "" {
		daemonHost = client.DefaultDockerHost
	}
	url, err := url.Parse(daemonHost)
	if err != nil {
		log.Error("Failed to parse daemon host", "host", daemonHost, "err", err)
		return host
	}

	if url.Scheme == "unix" {
		host = "localhost"
	} else {
		host, _, err = net.SplitHostPort(url.Host)
		if err != nil {
			log.Error("Failed to split host and port", "url", url.Host, "err", err)
			return ""
		}
	}

	return host
}

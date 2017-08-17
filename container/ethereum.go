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
	"io"
	"io/ioutil"
	"log"
	"math/big"
	"os"
	"path/filepath"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/getamis/go-ethereum/cmd/utils"
	"github.com/getamis/go-ethereum/ethclient"
	"github.com/getamis/istanbul-tools/core/genesis"
)

const (
	healthCheckRetryCount = 5
	healthCheckRetryDelay = 2 * time.Second
)

type Ethereum interface {
	Init(string) error
	Start() error
	Stop() error

	Host() string
	NewClient() *ethclient.Client
}

func NewEthereum(c *client.Client, options ...Option) *ethereum {
	eth := &ethereum{
		client: c,
	}

	for _, opt := range options {
		opt(eth)
	}

	filters := filters.NewArgs()
	filters.Add("reference", eth.Image())

	images, err := c.ImageList(context.Background(), types.ImageListOptions{
		Filters: filters,
	})

	if len(images) == 0 || err != nil {
		out, err := eth.client.ImagePull(context.Background(), eth.Image(), types.ImagePullOptions{})
		if err != nil {
			log.Printf("Cannot pull %s, err: %v", eth.Image(), err)
			return nil
		}
		if eth.logging {
			io.Copy(os.Stdout, out)
		} else {
			io.Copy(ioutil.Discard, out)
		}
	}

	return eth
}

type ethereum struct {
	ok          bool
	flags       []string
	hostDataDir string
	dataDir     string
	port        string
	rpcPort     string
	wsPort      string
	hostName    string
	containerID string

	imageRepository string
	imageTag        string

	logging bool
	client  *client.Client
}

func (eth *ethereum) Init(genesisFile string) error {
	resp, err := eth.client.ContainerCreate(context.Background(),
		&container.Config{
			Image: eth.Image(),
			Cmd: []string{
				"init",
				"--" + utils.DataDirFlag.Name,
				eth.dataDir,
				filepath.Join("/", genesis.FileName),
			},
		},
		&container.HostConfig{
			Binds: []string{
				genesisFile + ":" + filepath.Join("/", genesis.FileName),
				eth.hostDataDir + ":" + eth.dataDir,
			},
		}, nil, "")
	if err != nil {
		log.Printf("Failed to create container, err: %v", err)
		return err
	}

	defer func() {
		if eth.logging {
			go eth.showLog(context.Background())
		}
	}()

	id := resp.ID

	if err := eth.client.ContainerStart(context.Background(), id, types.ContainerStartOptions{}); err != nil {
		log.Printf("Failed to start container, err: %v", err)
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resC, errC := eth.client.ContainerWait(ctx, id, container.WaitConditionNotRunning)
	select {
	case <-resC:
	case <-errC:
		log.Printf("Failed to wait container, err: %v", err)
		return err
	}

	return eth.client.ContainerRemove(context.Background(), id,
		types.ContainerRemoveOptions{
			Force: true,
		})
}

func (eth *ethereum) Start() error {
	exposedPorts := make(map[nat.Port]struct{})
	portBindings := nat.PortMap{}

	if eth.port != "" {
		exposedPorts[nat.Port(eth.port)] = struct{}{}
		portBindings[nat.Port(eth.port)] = []nat.PortBinding{
			{
				HostIP:   "0.0.0.0",
				HostPort: eth.port,
			},
		}
	}

	if eth.rpcPort != "" {
		exposedPorts[nat.Port(eth.rpcPort)] = struct{}{}
		portBindings[nat.Port(eth.rpcPort)] = []nat.PortBinding{
			{
				HostIP:   "0.0.0.0",
				HostPort: eth.rpcPort,
			},
		}
	}

	if eth.wsPort != "" {
		exposedPorts[nat.Port(eth.wsPort)] = struct{}{}
		portBindings[nat.Port(eth.wsPort)] = []nat.PortBinding{
			{
				HostIP:   "0.0.0.0",
				HostPort: eth.wsPort,
			},
		}
	}

	resp, err := eth.client.ContainerCreate(context.Background(),
		&container.Config{
			Hostname:     "geth-" + eth.hostName,
			Image:        eth.Image(),
			Cmd:          eth.flags,
			ExposedPorts: exposedPorts,
		},
		&container.HostConfig{
			Binds: []string{
				eth.hostDataDir + ":" + eth.dataDir,
			},
			PortBindings: portBindings,
		}, nil, "")
	if err != nil {
		log.Printf("Failed to create container, err: %v", err)
		return err
	}

	defer func() {
		if eth.logging {
			go eth.showLog(context.Background())
		}
	}()
	eth.containerID = resp.ID

	err = eth.client.ContainerStart(context.Background(), eth.containerID, types.ContainerStartOptions{})
	if err != nil {
		log.Printf("Failed to start container, err: %v", err)
		return err
	}

	for i := 0; i < healthCheckRetryCount; i++ {
		cli := eth.NewClient()
		if cli == nil {
			time.Sleep(healthCheckRetryDelay)
			continue
		}
		_, err = cli.BlockByNumber(context.Background(), big.NewInt(0))
		if err != nil {
			time.Sleep(healthCheckRetryDelay)
			continue
		} else {
			eth.ok = true
		}
	}

	if !eth.ok {
		return errors.New("Failed to start geth")
	}

	return nil
}

func (eth *ethereum) Stop() error {
	timeout := 10 * time.Second
	err := eth.client.ContainerStop(context.Background(), eth.containerID, &timeout)
	if err != nil {
		return err
	}

	return eth.client.ContainerRemove(context.Background(), eth.containerID,
		types.ContainerRemoveOptions{
			Force: true,
		})
}

func (eth *ethereum) Wait(t time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), t)
	defer cancel()
	_, errCh := eth.client.ContainerWait(ctx, eth.containerID, "")
	return <-errCh
}

func (eth *ethereum) Running() bool {
	containers, err := eth.client.ContainerList(context.Background(), types.ContainerListOptions{})
	if err != nil {
		log.Printf("Failed to list containers, err: %v", err)
		return false
	}

	for _, c := range containers {
		if c.ID == eth.containerID {
			return true
		}
	}

	return false
}

func (eth *ethereum) NewClient() *ethclient.Client {
	var scheme, port string

	if eth.rpcPort != "" {
		scheme = "http://"
		port = eth.rpcPort
	}
	if eth.wsPort != "" {
		scheme = "ws://"
		port = eth.wsPort
	}
	client, err := ethclient.Dial(scheme + eth.Host() + ":" + port)
	if err != nil {
		log.Printf("Failed to dial to geth, err: %v", err)
		return nil
	}
	return client
}

// ----------------------------------------------------------------------------

func (eth *ethereum) showLog(context context.Context) {
	if readCloser, err := eth.client.ContainerLogs(context, eth.containerID,
		types.ContainerLogsOptions{ShowStderr: true, Follow: true}); err == nil {
		defer readCloser.Close()
		_, err = io.Copy(os.Stdout, readCloser)
		if err != nil && err != io.EOF {
			log.Fatal(err)
		}
	}
}

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
	"log"
	"math/big"
	"os"
	"path/filepath"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/getamis/go-ethereum/cmd/utils"
	"github.com/getamis/go-ethereum/ethclient"
	"github.com/getamis/istanbul-tools/core"
)

const (
	healthCheckRetryCount = 5
	healthCheckRetryDelay = 2 * time.Second
)

type Ethereum interface {
	Init(string) error
	Start() error
	Stop() error
}

func NewEthereum(c *client.Client, options ...Option) *ethereum {
	geth := &ethereum{
		client: c,
	}

	for _, opt := range options {
		opt(geth)
	}

	return geth
}

type ethereum struct {
	ok          bool
	flags       []string
	hostDataDir string
	dataDir     string
	port        string
	rpcPort     string
	hostName    string
	containerID string
	imageName   string
	logging     bool
	client      *client.Client
}

func (eth *ethereum) Init(genesisFile string) error {
	results, err := eth.client.ImageSearch(context.Background(), eth.imageName, types.ImageSearchOptions{
		Limit: 1,
	})
	if err != nil {
		log.Printf("Cannot search %s, err: %v", eth.imageName, err)
		return err
	}

	if len(results) == 0 {
		out, err := eth.client.ImagePull(context.Background(), eth.imageName, types.ImagePullOptions{})
		if err != nil {
			log.Printf("Cannot pull %s, err: %v", eth.imageName, err)
			return err
		}
		if eth.logging {
			io.Copy(os.Stdout, out)
		} else {
			_ = out
		}
	}

	resp, err := eth.client.ContainerCreate(context.Background(),
		&container.Config{
			Image: eth.imageName,
			Cmd: []string{
				"init",
				"--" + utils.DataDirFlag.Name,
				eth.dataDir,
				filepath.Join("/", core.GenesisJson),
			},
		},
		&container.HostConfig{
			Binds: []string{
				genesisFile + ":" + filepath.Join("/", core.GenesisJson),
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
	eth.containerID = resp.ID

	return eth.client.ContainerStart(context.Background(), eth.containerID, types.ContainerStartOptions{})
}

func (eth *ethereum) Start() error {
	resp, err := eth.client.ContainerCreate(context.Background(),
		&container.Config{
			Hostname: "geth-" + eth.hostName,
			Image:    eth.imageName,
			Cmd:      eth.flags,
			ExposedPorts: map[nat.Port]struct{}{
				nat.Port(eth.port):    {},
				nat.Port(eth.rpcPort): {},
			},
		},
		&container.HostConfig{
			Binds: []string{
				eth.hostDataDir + ":" + eth.dataDir,
			},
			PortBindings: nat.PortMap{
				nat.Port(eth.port): []nat.PortBinding{
					{
						HostIP:   "0.0.0.0",
						HostPort: eth.port,
					},
				},
				nat.Port(eth.rpcPort): []nat.PortBinding{
					{
						HostIP:   "0.0.0.0",
						HostPort: eth.rpcPort,
					},
				},
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
	eth.containerID = resp.ID

	err = eth.client.ContainerStart(context.Background(), eth.containerID, types.ContainerStartOptions{})
	if err != nil {
		log.Printf("Failed to start container, err: %v", err)
		return err
	}

	for i := 0; i < healthCheckRetryCount; i++ {
		cli, err := ethclient.Dial("http://localhost:" + eth.rpcPort)
		if err != nil {
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

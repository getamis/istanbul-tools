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
	"io"
	"log"
	"os"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/moby/moby/client"
)

func NewEthereum(imageName string, id string) *ethereum {
	client, err := client.NewEnvClient()
	if err != nil {
		log.Fatalf("Cannot connect to Docker daemon, err: %v", err)
	}
	return &ethereum{
		id:        id,
		imageName: imageName,
		client:    client,
	}
}

type ethereum struct {
	id          string
	containerID string
	imageName   string
	client      *client.Client
}

func (eth *ethereum) Start(showLog bool) error {
	_, err := eth.client.ImagePull(context.Background(), eth.imageName, types.ImagePullOptions{})
	if err != nil {
		log.Printf("Cannot pull %s, err: %v", eth.imageName, err)
		return err
	}

	resp, err := eth.client.ContainerCreate(context.Background(), &container.Config{
		Hostname:     "geth-" + eth.id,
		Image:        eth.imageName,
		AttachStdout: true,
	}, nil, nil, "")
	if err != nil {
		log.Printf("Failed to create container, err: %v", err)
		return err
	}

	defer func() {
		if showLog {
			go eth.showLog(context.Background())
		}
	}()
	eth.containerID = resp.ID

	return eth.client.ContainerStart(context.Background(), eth.containerID, types.ContainerStartOptions{})
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

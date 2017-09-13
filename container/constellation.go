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
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"

	"github.com/getamis/istanbul-tools/common"
	"github.com/getamis/istanbul-tools/log"
)

//TODO: refactor this with ethereum options?
/**
 * Constellation options
 **/
type ConstellationOption func(*constellation)

func CTImageRepository(repository string) ConstellationOption {
	return func(ct *constellation) {
		ct.imageRepository = repository
	}
}

func CTImageTag(tag string) ConstellationOption {
	return func(ct *constellation) {
		ct.imageTag = tag
	}
}

func CTHost(ip net.IP, port int) ConstellationOption {
	return func(ct *constellation) {
		ct.port = fmt.Sprintf("%d", port)
		ct.ip = ip.String()
		ct.flags = append(ct.flags, fmt.Sprintf("--port=%d", port))
		ct.flags = append(ct.flags, fmt.Sprintf("--url=%s", ct.Host()))
	}
}

func CTLogging(enabled bool) ConstellationOption {
	return func(ct *constellation) {
		ct.logging = enabled
	}
}

func CTDockerNetworkName(dockerNetworkName string) ConstellationOption {
	return func(ct *constellation) {
		ct.dockerNetworkName = dockerNetworkName
	}
}

func CTWorkDir(workDir string) ConstellationOption {
	return func(ct *constellation) {
		ct.workDir = workDir
		ct.flags = append(ct.flags, fmt.Sprintf("--storage=%s", workDir))
	}
}

func CTKeyName(keyName string) ConstellationOption {
	return func(ct *constellation) {
		ct.keyName = keyName
		ct.flags = append(ct.flags, fmt.Sprintf("--privatekeys=%s", ct.keyPath("key")))
		ct.flags = append(ct.flags, fmt.Sprintf("--publickeys=%s", ct.keyPath("pub")))
	}
}

func CTSocketFilename(socketFilename string) ConstellationOption {
	return func(ct *constellation) {
		ct.socketFilename = socketFilename
		ct.flags = append(ct.flags, fmt.Sprintf("--socket=%s", filepath.Join(ct.workDir, socketFilename)))
	}
}

func CTVerbosity(verbosity int) ConstellationOption {
	return func(ct *constellation) {
		ct.flags = append(ct.flags, fmt.Sprintf("--verbosity=%d", verbosity))
	}
}

func CTOtherNodes(urls []string) ConstellationOption {
	return func(ct *constellation) {
		ct.flags = append(ct.flags, fmt.Sprintf("--othernodes=%s", strings.Join(urls, ",")))
	}
}

/**
 * Constellation interface and constructors
 **/
type Constellation interface {
	// GenerateKey() generates private/public key pair
	GenerateKey() (string, error)
	// Start() starts constellation service
	Start() error
	// Stop() stops constellation service
	Stop() error
	// Host() returns constellation service url
	Host() string
	// Running() returns true if container is running
	Running() bool
	// WorkDir() returns local working directory
	WorkDir() string
	// ConfigPath() returns container config path
	ConfigPath() string
	// Binds() returns volume binding paths
	Binds() []string
	// PublicKeys() return public keys
	PublicKeys() []string
}

func NewConstellation(c *client.Client, options ...ConstellationOption) *constellation {
	ct := &constellation{
		client: c,
	}

	for _, opt := range options {
		opt(ct)
	}

	filters := filters.NewArgs()
	filters.Add("reference", ct.Image())

	images, err := c.ImageList(context.Background(), types.ImageListOptions{
		Filters: filters,
	})

	if len(images) == 0 || err != nil {
		out, err := ct.client.ImagePull(context.Background(), ct.Image(), types.ImagePullOptions{})
		if err != nil {
			log.Error("Failed to pull image", "image", ct.Image(), "err", err)
			return nil
		}
		if ct.logging {
			io.Copy(os.Stdout, out)
		} else {
			io.Copy(ioutil.Discard, out)
		}
	}

	return ct
}

/**
 * Constellation implementation
 **/
type constellation struct {
	flags          []string
	ip             string
	port           string
	containerID    string
	workDir        string
	localWorkDir   string
	keyName        string
	socketFilename string

	imageRepository   string
	imageTag          string
	dockerNetworkName string

	logging bool
	client  *client.Client
}

func (ct *constellation) Image() string {
	if ct.imageTag == "" {
		return ct.imageRepository + ":latest"
	}
	return ct.imageRepository + ":" + ct.imageTag
}

func (ct *constellation) GenerateKey() (localWorkDir string, err error) {
	// Generate empty password file
	ct.localWorkDir, err = common.GenerateRandomDir()
	if err != nil {
		log.Error("Failed to generate working dir", "dir", ct.localWorkDir, "err", err)
		return "", err
	}

	// Generate config file
	configContent := fmt.Sprintf("socket=\"%s\"\npublickeys=[\"%s\"]\n",
		ct.keyPath("ipc"), ct.keyPath("pub"))
	localConfigPath := ct.localConfigPath()
	err = ioutil.WriteFile(localConfigPath, []byte(configContent), 0600)
	if err != nil {
		log.Error("Failed to write config", "file", localConfigPath, "err", err)
		return "", err
	}

	// Create container and mount working directory
	binds := ct.Binds()
	config := &container.Config{
		Image: ct.Image(),
		Cmd: []string{
			"--generatekeys=" + ct.keyPath(""),
		},
	}
	hostConfig := &container.HostConfig{
		Binds: binds,
	}
	resp, err := ct.client.ContainerCreate(context.Background(), config, hostConfig, nil, "")
	if err != nil {
		log.Error("Failed to create container", "err", err)
		return "", err
	}
	id := resp.ID

	// Start container
	if err := ct.client.ContainerStart(context.Background(), id, types.ContainerStartOptions{}); err != nil {
		log.Error("Failed to start container", "err", err)
		return "", err
	}

	// Attach container: for stdin interaction with the container.
	// - constellation-node generatekeys takes stdin as password
	hiresp, err := ct.client.ContainerAttach(context.Background(), id, types.ContainerAttachOptions{Stream: true, Stdin: true})
	if err != nil {
		log.Error("Failed to attach container", "err", err)
		return "", err
	}
	// - write empty string password to container stdin
	hiresp.Conn.Write([]byte("")) //Empty password

	// Wait container
	_, err = ct.client.ContainerWait(context.Background(), id)
	if err != nil {
		log.Error("Failed to wait container", "err", err)
		return "", err
	}

	if ct.logging {
		ct.showLog(context.Background())
	}

	// Stop container
	return ct.localWorkDir, ct.client.ContainerRemove(context.Background(), id, types.ContainerRemoveOptions{Force: true})
}

func (ct *constellation) Start() error {
	defer func() {
		if ct.logging {
			go ct.showLog(context.Background())
		}
	}()

	// container config
	exposedPorts := make(map[nat.Port]struct{})
	exposedPorts[nat.Port(ct.port)] = struct{}{}
	config := &container.Config{
		Image:        ct.Image(),
		Cmd:          ct.flags,
		ExposedPorts: exposedPorts,
	}

	// host config
	binds := []string{
		ct.localWorkDir + ":" + ct.workDir,
	}
	hostConfig := &container.HostConfig{
		Binds: binds,
	}

	// Setup network config
	var networkingConfig *network.NetworkingConfig
	if ct.ip != "" && ct.dockerNetworkName != "" {
		endpointsConfig := make(map[string]*network.EndpointSettings)
		endpointsConfig[ct.dockerNetworkName] = &network.EndpointSettings{
			IPAMConfig: &network.EndpointIPAMConfig{
				IPv4Address: ct.ip,
			},
		}
		networkingConfig = &network.NetworkingConfig{
			EndpointsConfig: endpointsConfig,
		}
	}

	// Create container
	resp, err := ct.client.ContainerCreate(context.Background(), config, hostConfig, networkingConfig, "")
	if err != nil {
		log.Error("Failed to create container", "err", err)
		return err
	}
	ct.containerID = resp.ID

	// Start container
	err = ct.client.ContainerStart(context.Background(), ct.containerID, types.ContainerStartOptions{})
	if err != nil {
		log.Error("Failed to start container", "ip", ct.ip, "err", err)
		return err
	}

	return nil
}

func (ct *constellation) Stop() error {
	err := ct.client.ContainerStop(context.Background(), ct.containerID, nil)
	if err != nil {
		return err
	}

	defer os.RemoveAll(ct.localWorkDir)

	return ct.client.ContainerRemove(context.Background(), ct.containerID,
		types.ContainerRemoveOptions{
			Force: true,
		})
}

func (ct *constellation) Host() string {
	return fmt.Sprintf("http://%s:%s/", ct.ip, ct.port)
}

func (ct *constellation) Running() bool {
	containers, err := ct.client.ContainerList(context.Background(), types.ContainerListOptions{})
	if err != nil {
		log.Error("Failed to list containers", "err", err)
		return false
	}

	for _, c := range containers {
		if c.ID == ct.containerID {
			return true
		}
	}

	return false
}

func (ct *constellation) WorkDir() string {
	return ct.localWorkDir
}

func (ct *constellation) ConfigPath() string {
	return ct.keyPath("conf")
}

func (ct *constellation) Binds() []string {
	return []string{ct.localWorkDir + ":" + ct.workDir}
}

func (ct *constellation) PublicKeys() []string {
	keyPath := ct.localKeyPath("pub")
	keyBytes, err := ioutil.ReadFile(keyPath)
	if err != nil {
		log.Error("Unable to read key file", "file", keyPath, "err", err)
		return nil
	}
	return []string{string(keyBytes)}
}

/**
 * Constellation internal functions
 **/

func (ct *constellation) showLog(context context.Context) {
	if readCloser, err := ct.client.ContainerLogs(context, ct.containerID,
		types.ContainerLogsOptions{ShowStderr: true, Follow: true}); err == nil {
		defer readCloser.Close()
		_, err = io.Copy(os.Stdout, readCloser)
		if err != nil && err != io.EOF {
			log.Error("Failed to print container log", "err", err)
			return
		}
	}
}

func (ct *constellation) keyPath(extension string) string {
	if extension == "" {
		return filepath.Join(ct.workDir, ct.keyName)
	} else {
		return filepath.Join(ct.workDir, fmt.Sprintf("%s.%s", ct.keyName, extension))
	}
}

func (ct *constellation) localKeyPath(extension string) string {
	return filepath.Join(ct.localWorkDir, fmt.Sprintf("%s.%s", ct.keyName, extension))
}

func (ct *constellation) localConfigPath() string {
	return filepath.Join(ct.localWorkDir, fmt.Sprintf("%s.conf", ct.keyName))
}

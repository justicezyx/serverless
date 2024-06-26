package core

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

var (
	// Package-shared docker client.
	dockerClient *client.Client
)

// This must be called before using any APIs in this file.
func InitDockerClient() error {
	var err error
	dockerClient, err = client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return err
	}
	return nil
}

func init() {
	err := InitDockerClient()
	if err != nil {
		panic(fmt.Sprintf("Could not create docker client, error: %v", err))
	}
}

type RunningContainer struct {
	// Human readable name for easier debugging.
	name string

	// Fixed parameter, set at launch time.
	containerID string

	// The URL to invoke APIs running inside this Container
	// Fixed parameter, set at launch time.
	Url string

	// The limit of how many concurrent requests this instance can serve.
	// Fixed parameter, set at launch time.
	concurLimit int

	// The time when this instance is launched.
	launchTime time.Time

	// The URL to check the readiness of the service running inside the service.
	readyUrl string

	// TODO: Add mutex protection to check isRdy. As the running container will be checked for readiness for each
	// incoming requests (and if the container is not ready, the caller needs to wait). The check is invoked by HTTP
	// handler functions for each incoming requests, so they would happen concurrently.
	rdyMu sync.Mutex
	// True if the container is ready to serve requests.
	// Needed because container needs some time to initiate.
	// Protected by rdyMu
	isRdy bool
	// The time when this instance is ready and is able to serve requests.
	// Protected by rdyMu
	rdyTime time.Time

	// The time duration that this instance is actually serving requests.
	busyTimeMu sync.RWMutex
	busyTime   time.Duration
}

func (c *RunningContainer) Stop() error {
	fmt.Println("stopping", c)
	ctx := context.Background()
	if err := dockerClient.ContainerStop(ctx, c.containerID, container.StopOptions{}); err != nil {
		return fmt.Errorf("failed to stop container %s: %v", c.containerID, err)
	}
	return nil
}

func (c *RunningContainer) Remove() error {
	fmt.Println("removing", c)
	err := dockerClient.ContainerRemove(context.Background(), c.containerID, container.RemoveOptions{})
	if err != nil {
		return fmt.Errorf("Failed to remove container %s: %v", c.containerID, err)
	}
	return nil
}

// Call this to record serving time.
func (c *RunningContainer) AddBusyTime(d time.Duration) {
	c.busyTimeMu.Lock()
	defer c.busyTimeMu.Unlock()
	c.busyTime += d
}

func (c *RunningContainer) BusyTime() time.Duration {
	c.busyTimeMu.RLock()
	defer c.busyTimeMu.RUnlock()
	return c.busyTime
}

func (c *RunningContainer) WaitForReady(timeout time.Duration) error {
	if c.IsReady() {
		return nil
	}
	err := WaitForHTTPGetOK(c.readyUrl, 100*time.Millisecond, timeout)
	if err == nil {
		c.rdyMu.Lock()
		defer c.rdyMu.Unlock()
		c.rdyTime = time.Now()
		c.isRdy = true
	}
	return err
}

func (c *RunningContainer) IsReady() bool {
	c.rdyMu.Lock()
	defer c.rdyMu.Unlock()
	return c.isRdy
}

// Define the Container interface
type ContainerInterface interface {
	Run(name string) (*RunningContainer, error)
}

// Represents a template of container. After running, a RunningContainer will be created.
type Container struct {
	image string
	cmd   []string
}

func NewContainer(image string, cmd []string) Container {
	return Container{
		image: image,
		cmd:   cmd,
	}
}

func preparePortBindings(portBindings map[string]string) (nat.PortSet, nat.PortMap, error) {
	exposedPorts := nat.PortSet{}
	portMap := nat.PortMap{}

	for hostPort, containerPort := range portBindings {
		port, err := nat.NewPort("tcp", containerPort)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid container port %s: %v", containerPort, err)
		}
		exposedPorts[port] = struct{}{}
		portMap[port] = []nat.PortBinding{
			{
				HostPort: hostPort,
			},
		}
	}

	return exposedPorts, portMap, nil
}

// Returns a randomly-picked port. The port can be used by another service to listen on.
func pickPort() (int, error) {
	// Listen on a random port by specifying port 0
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		return 0, fmt.Errorf("Error listening on port: %v\n", err)
	}
	// Get the assigned port
	port := listener.Addr().(*net.TCPAddr).Port
	// Close the listener.
	listener.Close()
	return port, nil
}

// The port used by the service running inside Container to accept requests.
const runtimePort = "5000"

// Run the input image, with the input cmd as the entrypoint, and portBindings as port mapping.
// The input image must be present locally.
func (c Container) Run(name string) (*RunningContainer, error) {
	ctx := context.Background()

	hostPort, err := pickPort()
	if err != nil {
		return nil, fmt.Errorf("Could not find free port for launching container instance, error: %v", err)
	}
	// Configure exposed ports and port bindings
	portBindings := map[string]string{strconv.Itoa(hostPort): runtimePort}
	exposedPorts, portMap, err := preparePortBindings(portBindings)
	if err != nil {
		return nil, fmt.Errorf("Error preparing port binding, error: %v", err)
	}

	resp, err := dockerClient.ContainerCreate(ctx, &container.Config{
		Image:        c.image,
		Cmd:          c.cmd,
		ExposedPorts: exposedPorts,
	}, &container.HostConfig{
		PortBindings: portMap,
	}, &network.NetworkingConfig{}, nil /*platform*/, name)

	if err != nil {
		return nil, fmt.Errorf("failed to create container: %v", err)
	}

	if err := dockerClient.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		return nil, fmt.Errorf("failed to start container: %v", err)
	}

	return &RunningContainer{
		name:        name,
		containerID: resp.ID,
		Url:         fmt.Sprintf("http://localhost:%d/invoke", hostPort),
		readyUrl:    fmt.Sprintf("http://localhost:%d/ready", hostPort),
		concurLimit: 2,
		launchTime:  time.Now(),
	}, nil
}

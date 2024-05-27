package core

import (
	"context"
	"fmt"
	"net"
	"strconv"

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

// Define the Container interface
type ContainerInterface interface {
	Run() (RunningContainer, error)
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

type RunningContainer struct {
	containerID string

	// The URL to invoke APIs running inside this Container
	Url string
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

// The port used by the service running inside to Container to accept requests.
const runtimePort = "5000"

// Run the input image, with the input cmd as the entrypoint, and portBindings as port mapping.
// The input image must be present locally.
func (c Container) Run() (RunningContainer, error) {
	ctx := context.Background()

	hostPort, err := pickPort()
	if err != nil {
		return RunningContainer{}, fmt.Errorf("Could not find free port for launching container instance, error: %v", err)
	}
	// Configure exposed ports and port bindings
	portBindings := map[string]string{strconv.Itoa(hostPort): runtimePort}
	exposedPorts, portMap, err := preparePortBindings(portBindings)
	if err != nil {
		return RunningContainer{}, fmt.Errorf("Error preparing port binding, error: %v", err)
	}

	resp, err := dockerClient.ContainerCreate(ctx, &container.Config{
		Image:        c.image,
		Cmd:          c.cmd,
		ExposedPorts: exposedPorts,
	}, &container.HostConfig{
		PortBindings: portMap,
	}, &network.NetworkingConfig{}, nil /*platform*/, "" /*name*/)

	if err != nil {
		return RunningContainer{}, fmt.Errorf("failed to create container: %v", err)
	}

	if err := dockerClient.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		return RunningContainer{}, fmt.Errorf("failed to start container: %v", err)
	}

	return RunningContainer{
		containerID: resp.ID,
		Url:         fmt.Sprintf("http://localhost:%d/invoke", hostPort),
	}, nil
}

func (c RunningContainer) Stop() error {
	fmt.Println("stop", c.containerID)
	ctx := context.Background()
	if err := dockerClient.ContainerStop(ctx, c.containerID, container.StopOptions{}); err != nil {
		return fmt.Errorf("failed to stop container %s: %v", c.containerID, err)
	}
	return nil
}

func (c RunningContainer) Remove() error {
	fmt.Println("remove", c.containerID)
	err := dockerClient.ContainerRemove(context.Background(), c.containerID, container.RemoveOptions{})
	if err != nil {
		return fmt.Errorf("Failed to remove container %s: %v", c.containerID, err)
	}
	return nil
}

package core

import (
	"context"
	"fmt"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

var (
	dockerClient *client.Client
)

func InitDockerClient() error {
	var err error
	dockerClient, err = client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return err
	}
	return nil
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

// Run the input image, with the input cmd as the entrypoint, and portBindings as port mapping.
// The input image must be present locally.
func (c Container) Run(portBindings map[string]string) (RunningContainer, error) {
	ctx := context.Background()

	// Configure exposed ports and port bindings
	exposedPorts, portMap, err := preparePortBindings(portBindings)
	if err != nil {
		return RunningContainer{}, err
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

	return RunningContainer{containerID: resp.ID}, nil
}

func (c RunningContainer) Stop() error {
	ctx := context.Background()
	if err := dockerClient.ContainerStop(ctx, c.containerID, container.StopOptions{}); err != nil {
		return fmt.Errorf("failed to stop container %s: %v", c.containerID, err)
	}
	return nil
}

func (c RunningContainer) Remove() error {
	err := dockerClient.ContainerRemove(context.Background(), c.containerID, container.RemoveOptions{})
	if err != nil {
		return fmt.Errorf("Failed to remove container %s: %v", c.containerID, err)
	}
	return nil
}

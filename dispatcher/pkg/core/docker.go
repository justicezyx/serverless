package core

import (
	"context"
	"fmt"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

type Docker struct {
	Client      *client.Client
	containerID string
}

func NewDocker() (*Docker, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}
	return &Docker{Client: cli}, nil
}

// Run the input image, with the input cmd as the entrypoint, and portBindings as port mapping.
// The input image must be present locally.
func (d *Docker) Run(image string, cmd []string, portBindings map[string]string) error {
	ctx := context.Background()

	// Configure exposed ports and port bindings
	exposedPorts, portMap, err := preparePortBindings(portBindings)
	if err != nil {
		return err
	}

	resp, err := d.Client.ContainerCreate(ctx, &container.Config{
		Image:        image,
		Cmd:          cmd,
		ExposedPorts: exposedPorts,
	}, &container.HostConfig{
		PortBindings: portMap,
	}, &network.NetworkingConfig{}, nil /*platform*/, "" /*name*/)

	if err != nil {
		return fmt.Errorf("failed to create container: %v", err)
	}

	if err := d.Client.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		return fmt.Errorf("failed to start container: %v", err)
	}
	d.containerID = resp.ID

	return nil
}

func (d *Docker) Stop() error {
	ctx := context.Background()
	if err := d.Client.ContainerStop(ctx, d.containerID, container.StopOptions{}); err != nil {
		return fmt.Errorf("failed to stop container %s: %v", d.containerID, err)
	}
	return nil
}

func (d *Docker) Remove() error {
	err := d.Client.ContainerRemove(context.Background(), d.containerID, container.RemoveOptions{})
	if err != nil {
		return fmt.Errorf("Failed to remove container %s: %v", d.containerID, err)
	}
	return nil
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

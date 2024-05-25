package dispatcher

import (
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
)

// TestDockerRun tests the Run method of the Docker struct.
func TestDockerRun(t *testing.T) {
	docker, err := NewDocker()
	if err != nil {
		t.Fatalf("Failed to create Docker client: %v", err)
	}

	image := "runtime:latest"
	cmd := []string{}
	portBindings := map[string]string{
		"8080": "5000",
	}

	err = docker.Run(image, cmd, portBindings)
	assert.Nil(t, err, "Expected no error, got %v", err)

	// Clean up: Stop and remove the container after testing
	containers, err := docker.Client.ContainerList(context.Background(), container.ListOptions{})
	if err != nil {
		t.Fatalf("Failed to list containers: %v", err)
	}

	for _, c := range containers {
		timeoutSec := 10
		// Stop the container with a timeout
		err = docker.Client.ContainerStop(context.Background(), c.ID, container.StopOptions{Timeout: &timeoutSec})
		if err != nil {
			t.Fatalf("Failed to stop container %s: %v", c.ID, err)
		}
		// Remove the container
		err = docker.Client.ContainerRemove(context.Background(), c.ID, container.RemoveOptions{})
		if err != nil {
			t.Fatalf("Failed to remove container %s: %v", c.ID, err)
		}
	}
}

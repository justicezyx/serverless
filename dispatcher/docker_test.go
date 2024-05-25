package dispatcher

import (
	"testing"

	"github.com/stretchr/testify/assert"
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

	assert.Nil(t, docker.Run(image, cmd, portBindings), "Expected no error, got %v", err)
	assert.Nil(t, docker.Stop(), "Expected no error stopping container, got %v", err)
	assert.Nil(t, docker.Remove(), "Expected no error removing container, got %v", err)
}

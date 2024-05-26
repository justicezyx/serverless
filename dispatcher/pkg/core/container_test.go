package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDockerRun tests the Run method of the Docker struct.
func TestDockerRun(t *testing.T) {
	require.Nil(t, InitDockerClient(), "InitDockerClient must succeed")

	// Go to $ToT/runtime for instructions of building this image.
	// This image has to be built locally, we don't do docker pull.
	image := "runtime:latest"
	cmd := []string{"python", "runtime.py", "--file=runtime_alpha.py", "--class_name=RuntimeAlpha"}
	container := NewContainer(image, cmd)

	rc, err := container.Run()
	assert.Nil(t, err, "Expected no error, got %v", err)

	assert.Nil(t, rc.Stop(), "Expected no error stopping container")
	assert.Nil(t, rc.Remove(), "Expected no error removing container")
}

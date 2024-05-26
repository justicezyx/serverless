package core

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDockerRun tests the Run method of the Docker struct.
func TestDockerRun(t *testing.T) {
	timer := NewTimer()
	require.Nil(t, InitDockerClient(), "InitDockerClient must succeed")
	fmt.Println("InitDockerClient time duration:", timer.Elapsed())

	// Go to $ToT/runtime for instructions of building this image.
	// This image has to be built locally, we don't do docker pull.
	image := "runtime:latest"
	cmd := []string{"python", "runtime.py", "--file=runtime_alpha.py", "--class_name=RuntimeAlpha"}

	timer = NewTimer()
	container := NewContainer(image, cmd)
	fmt.Println("NewContainer time duration:", timer.Elapsed())

	timer = NewTimer()
	rc, err := container.Run()
	fmt.Println("RunContainer time duration:", timer.Elapsed())

	assert.Nil(t, err, "Expected no error, got %v", err)
	assert.NotEmpty(t, rc.Url, "Expect non-empty URL to the container instance")
	assert.Nil(t, rc.Stop(), "Expected no error stopping container")
	assert.Nil(t, rc.Remove(), "Expected no error removing container")
}

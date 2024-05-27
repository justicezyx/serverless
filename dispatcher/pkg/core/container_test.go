package core

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestDockerRun tests the Run method of the Docker struct.
func TestDockerRun(t *testing.T) {
	// Go to $ToT/runtime for instructions of building this image.
	// This image has to be built locally, we don't do docker pull.
	image := "runtime:latest"
	cmd := []string{"python", "runtime.py", "--file=runtime_alpha.py", "--class_name=RuntimeAlpha"}

	timer := NewTimer()
	container := NewContainer(image, cmd)
	fmt.Println("NewContainer time duration:", timer.Elapsed())

	timer = NewTimer()
	rc, err := container.Run()
	fmt.Println("RunContainer time duration:", timer.Elapsed())

	assert.Nil(t, err, "Expected no error, got %v", err)
	assert.NotEmpty(t, rc.Url, "Expect non-empty URL to the container instance")

	timer = NewTimer()
	assert.Nil(t, rc.Stop(), "Expected no error stopping container")
	fmt.Println("StopContainer time duration:", timer.Elapsed())

	timer = NewTimer()
	assert.Nil(t, rc.Remove(), "Expected no error removing container")
	fmt.Println("RemoveContainer time duration:", timer.Elapsed())
}

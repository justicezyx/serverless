package core

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockContainer is a mock implementation of the Container interface
type MockContainer struct {
	mock.Mock
}

func (m *MockContainer) Run(_ string) (*RunningContainer, error) {
	args := m.Called()
	return args.Get(0).(*RunningContainer), args.Error(1)
}

type MockRunningContainer struct {
	containerID string
	Url         string
}

func (m *MockRunningContainer) Stop() error {
	return nil
}

func (m *MockRunningContainer) Remove() error {
	return nil
}

// TestLauncher_Launch tests the Launch method of the Launcher
func TestLauncher_Launch(t *testing.T) {
	// Initialize dispatcher
	dispatcher := NewLauncher(time.Second)

	// Create a mock container
	mockContainer := new(MockContainer)
	dispatcher.fnContainerMap["testFn"] = mockContainer

	// Create a mock running container
	mockRunningContainer := RunningContainer{
		Url: "http://localhost:5000",
	}

	// Set up expectations for the mock container
	mockContainer.On("Run", mock.Anything).Return(mockRunningContainer, nil)

	// Call the Launch method
	rc, err := dispatcher.Launch("testFn")

	// Assertions
	assert.Equal(t, rc.name, "testFn-0")
	assert.NoError(t, err)
	assert.Contains(t, dispatcher.fnInstsMap, "testFn")
	assert.Equal(t, dispatcher.fnInstsMap["testFn"][0].Url, "http://localhost:5000")

	// Verify the expectations
	mockContainer.AssertExpectations(t)
}

// TestLauncher_LaunchNoContainer tests the Launch method when there is no container
func TestLauncher_LaunchNoContainer(t *testing.T) {
	dispatcher := NewLauncher(time.Second)

	_, err := dispatcher.Launch("unknownFn")

	assert.Error(t, err)
}

// TestLauncher_LaunchRunError tests the Launch method when the container fails to run
func TestLauncher_LaunchRunError(t *testing.T) {
	dispatcher := NewLauncher(time.Second)

	mockContainer := new(MockContainer)
	dispatcher.registerContainer("testFn", mockContainer)

	mockContainer.On("Run", mock.Anything).Return(RunningContainer{}, errors.New("run error"))

	_, err := dispatcher.Launch("testFn")

	assert.Error(t, err)
	assert.NotContains(t, dispatcher.fnInstsMap, "testFn")

	mockContainer.AssertExpectations(t)
}

// TestLauncher_PickUrl tests the PickUrl method of the Launcher
func TestLauncher_PickUrl(t *testing.T) {
	dispatcher := NewLauncher(time.Second)

	mockRunningContainer := RunningContainer{
		Url: "http://localhost:5000",
	}

	dispatcher.fnInstsMap["testFn"] = []*RunningContainer{&mockRunningContainer}

	rc, err := dispatcher.PickInst("testFn")

	assert.NoError(t, err)
	assert.Equal(t, "http://localhost:5000", rc.name)
}

// TestLauncher_PickUrlNoRunningContainer tests the PickUrl method when there is no running container
func TestLauncher_PickUrlNoRunningContainer(t *testing.T) {
	dispatcher := NewLauncher(time.Second)

	rc, err := dispatcher.PickInst("unknownFn")

	assert.Error(t, err)
	assert.Equal(t, rc, nil)
}

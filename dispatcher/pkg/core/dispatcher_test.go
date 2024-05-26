package core

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockContainer is a mock implementation of the Container interface
type MockContainer struct {
	mock.Mock
}

func (m *MockContainer) Run() (RunningContainer, error) {
	args := m.Called()
	return args.Get(0).(RunningContainer), args.Error(1)
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

// TestDispatcher_Launch tests the Launch method of the Dispatcher
func TestDispatcher_Launch(t *testing.T) {
	// Initialize dispatcher
	dispatcher := NewDispatcher()

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
	err := dispatcher.Launch("testFn")

	// Assertions
	assert.NoError(t, err)
	assert.Contains(t, dispatcher.urlInstanceMap, "http://localhost:5000")
	assert.Contains(t, dispatcher.fnInstanceMap, "testFn")
	assert.Equal(t, dispatcher.fnInstanceMap["testFn"][0].Url, "http://localhost:5000")

	// Verify the expectations
	mockContainer.AssertExpectations(t)
}

// TestDispatcher_LaunchNoContainer tests the Launch method when there is no container
func TestDispatcher_LaunchNoContainer(t *testing.T) {
	dispatcher := NewDispatcher()

	err := dispatcher.Launch("unknownFn")

	assert.Error(t, err)
}

// TestDispatcher_LaunchRunError tests the Launch method when the container fails to run
func TestDispatcher_LaunchRunError(t *testing.T) {
	dispatcher := NewDispatcher()

	mockContainer := new(MockContainer)
	dispatcher.RegisterContainer("testFn", mockContainer)

	mockContainer.On("Run", mock.Anything).Return(RunningContainer{}, errors.New("run error"))

	err := dispatcher.Launch("testFn")

	assert.Error(t, err)
	assert.NotContains(t, dispatcher.urlInstanceMap, "http://localhost:5000")
	assert.NotContains(t, dispatcher.fnInstanceMap, "testFn")

	mockContainer.AssertExpectations(t)
}

// TestDispatcher_PickUrl tests the PickUrl method of the Dispatcher
func TestDispatcher_PickUrl(t *testing.T) {
	dispatcher := NewDispatcher()

	mockRunningContainer := RunningContainer{
		Url: "http://localhost:5000",
	}

	dispatcher.fnInstanceMap["testFn"] = []RunningContainer{mockRunningContainer}

	url, err := dispatcher.PickUrl("testFn")

	assert.NoError(t, err)
	assert.Equal(t, "http://localhost:5000", url)
}

// TestDispatcher_PickUrlNoRunningContainer tests the PickUrl method when there is no running container
func TestDispatcher_PickUrlNoRunningContainer(t *testing.T) {
	dispatcher := NewDispatcher()

	url, err := dispatcher.PickUrl("unknownFn")

	assert.Error(t, err)
	assert.Equal(t, "", url)
}

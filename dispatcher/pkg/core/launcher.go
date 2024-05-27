package core

import (
	"fmt"
	"math/rand"
)

// Launcher stores containers for starting instances to serve function invocations.
type Launcher struct {
	// Map from the function to the Container template.
	// ContainerInterface is for testing.
	fnContainerMap map[string]ContainerInterface

	// A map from the function to the corresponding running container instances.
	// Picking any one of these instances for serving the function.
	fnInstanceMap map[string][]RunningContainer
}

func NewLauncher() Launcher {
	return Launcher{
		fnContainerMap: make(map[string]ContainerInterface),
		fnInstanceMap:  make(map[string][]RunningContainer),
	}
}

// Used for testing.
func (d *Launcher) registerContainer(fn string, c ContainerInterface) {
	d.fnContainerMap[fn] = c
}

// Launch a container instance for serving function fn.
func (d *Launcher) Launch(fn string) error {
	c, ok := d.fnContainerMap[fn]
	if !ok {
		return fmt.Errorf("Could not find Container for serverless function %s", fn)
	}
	rc, err := c.Run()
	if err != nil {
		return fmt.Errorf("Could not run container for function: %s, error: %v", fn, err)
	}
	if _, ok := d.fnInstanceMap[fn]; !ok {
		d.fnInstanceMap[fn] = make([]RunningContainer, 0)
	}
	d.fnInstanceMap[fn] = append(d.fnInstanceMap[fn], rc)
	return nil
}

// Returns the URL for serving the input function.
// Picks a random container instances, and returns its URL.
func (d Launcher) PickUrl(fn string) (string, error) {
	rcs, ok := d.fnInstanceMap[fn]
	if !ok || len(rcs) == 0 {
		return "", fmt.Errorf("No running container for function %s", fn)
	}
	idx := rand.Intn(len(rcs))
	return rcs[idx].Url, nil
}

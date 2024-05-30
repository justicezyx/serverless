package core

import (
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"
)

// Launcher stores containers for starting instances to serve function invocations.
// TODO: Needs sync.Mutex to protect from concurrent access.
type Launcher struct {
	// Set during creation, and never change afterwards.
	// Therefore mutex protection is not required.
	//
	// Map from the function to the Container template.
	// ContainerInterface is used for testing.
	fnContainerMap map[string]ContainerInterface

	// A map from the function to the corresponding running container instances.
	// Picking any one of these instances for serving the function.
	fnInstsMap   map[string][]*RunningContainer
	fnInstsMapMu sync.Mutex
}

func NewLauncher() Launcher {
	return Launcher{
		fnContainerMap: make(map[string]ContainerInterface),
		fnInstsMap:     make(map[string][]*RunningContainer),
	}
}

func (d *Launcher) registerContainer(fn string, c ContainerInterface) {
	d.fnContainerMap[fn] = c
}

// Launch a container instance for serving function fn.
func (d *Launcher) Launch(fn string) (*RunningContainer, error) {
	d.fnInstsMapMu.Lock()
	defer d.fnInstsMapMu.Unlock()

	c, ok := d.fnContainerMap[fn]
	if !ok {
		return nil, fmt.Errorf("Could not find Container for serverless function %s", fn)
	}
	rc, err := c.Run()
	if err != nil {
		return nil, fmt.Errorf("Could not run container for function: %s, error: %v", fn, err)
	}
	if _, ok := d.fnInstsMap[fn]; !ok {
		d.fnInstsMap[fn] = make([]*RunningContainer, 0)
	}
	d.fnInstsMap[fn] = append(d.fnInstsMap[fn], rc)
	return rc, nil
}

// Returns the URL for serving the input function.
// Picks a random container instances, and returns its URL.
func (d Launcher) PickUrl(fn string) (string, error) {
	rcs, ok := d.fnInstsMap[fn]
	if !ok || len(rcs) == 0 {
		return "", fmt.Errorf("No running container for function %s", fn)
	}
	idx := rand.Intn(len(rcs))
	return rcs[idx].Url, nil
}

// Returns the URL for serving the input function.
// Picks a random container instances, and returns its URL.
func (d Launcher) PickInst(fn string) (*RunningContainer, error) {
	rcs, ok := d.fnInstsMap[fn]
	if !ok || len(rcs) == 0 {
		return nil, fmt.Errorf("No running container for function %s", fn)
	}
	rdyRCs := make([]*RunningContainer, 0, len(rcs))
	for _, rc := range rcs {
		if rc.IsReady() {
			rdyRCs = append(rdyRCs, rc)
		}
	}
	rand.Seed(time.Now().UnixNano())
	return rdyRCs[rand.Intn(len(rdyRCs))], nil
}

// Shutdown all container instances. Called when shutting down server.
func (d *Launcher) ShutdownAll() {
	var wg sync.WaitGroup
	for _, runningContainers := range d.fnInstsMap {
		for _, runningContainer := range runningContainers {
			wg.Add(1)
			go func() {
				defer wg.Done()
				if err := runningContainer.Stop(); err != nil {
					log.Println("Failed to stop running container:", runningContainer)
				}
				if err := runningContainer.Remove(); err != nil {
					log.Println("Failed to remove running container:", runningContainer)
				}
			}()
		}
	}
	wg.Wait()
}

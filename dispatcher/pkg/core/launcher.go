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
	fnInstsMapMu sync.Mutex
	fnInstsMap   map[string][]*RunningContainer

	// Notify launcher to immediately start an instance for the received function.
	launchNotifier chan string

	// The interval for periodically check the load on each RunningContainer.
	checkInterval time.Duration
}

func NewLauncher(interval time.Duration) Launcher {
	return Launcher{
		fnContainerMap: make(map[string]ContainerInterface),
		fnInstsMap:     make(map[string][]*RunningContainer),
		launchNotifier: make(chan string),
		checkInterval:  interval,
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

// Brings down one RunningContainer for function fn.
func (l *Launcher) Shutdown(fn string) (*RunningContainer, error) {
	l.fnInstsMapMu.Lock()
	defer l.fnInstsMapMu.Unlock()

	rcs, ok := l.fnInstsMap[fn]
	if !ok || len(rcs) == 0 {
		return nil, fmt.Errorf("No RunningContainer for function %s", fn)
	}

	youngest := rcs[0]
	idx := 0
	for i, rc := range rcs[1:] {
		if rc.launchTime.After(youngest.launchTime) {
			youngest = rc
			idx = i
		}
	}

	// Remove the RunningContainer from the map
	rcs[idx] = rcs[len(rcs)-1]
	rcs = rcs[:len(rcs)-1]
	l.fnInstsMap[fn] = rcs

	if err := youngest.Stop(); err != nil {
		log.Println("Failed to stop running container:", youngest)
	}
	if err := youngest.Remove(); err != nil {
		log.Println("Failed to remove running container:", youngest)
	}
	return youngest, nil
}

// Returns the URL for serving the input function.
// Picks a random container instances, and returns its URL.
func (d *Launcher) PickUrl(fn string) (string, error) {
	rcs, ok := d.fnInstsMap[fn]
	if !ok || len(rcs) == 0 {
		return "", fmt.Errorf("No running container for function %s", fn)
	}
	idx := rand.Intn(len(rcs))
	return rcs[idx].Url, nil
}

// Returns the URL for serving the input function.
// Picks a random container instances, and returns its URL.
func (d *Launcher) PickInst(fn string) (*RunningContainer, error) {
	d.fnInstsMapMu.Lock()
	defer d.fnInstsMapMu.Unlock()

	rcs, ok := d.fnInstsMap[fn]
	if !ok || len(rcs) == 0 {
		return nil, fmt.Errorf("No running container for function %s", fn)
	}
	rand.Seed(time.Now().UnixNano())
	return rcs[rand.Intn(len(rcs))], nil
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

func (l *Launcher) calUtilRatio() map[string]float64 {
	l.fnInstsMapMu.Lock()
	defer l.fnInstsMapMu.Unlock()

	res := make(map[string]float64)
	var totalBusyTime time.Duration
	var totalRdyTime time.Duration
	for fn, rcs := range l.fnInstsMap {
		for _, rc := range rcs {
			if !rc.IsReady() {
				continue
			}
			totalBusyTime += rc.BusyTime()
			totalRdyTime += time.Now().Sub(rc.rdyTime)
		}
		res[fn] = float64(totalBusyTime) / float64(totalRdyTime)
	}
	return res
}

const utilRatioUpperBound = 0.8
const utilRatioLowerBound = 0.5

func (l *Launcher) DaemonProcess() {
	ticker := time.NewTicker(l.checkInterval)
	for {
		select {
		case fn := <-l.launchNotifier:
			log.Println("Received launch notification, function:", fn)
			_, err := l.Launch(fn)
			if err != nil {
				log.Println("Failed to launch container, function:", fn)
			}
		case _ = <-ticker.C:
			utilRatio := l.calUtilRatio()
			for fn, r := range utilRatio {
				if r > utilRatioUpperBound {
					rc, err := l.Launch(fn)
					log.Println("Launched RunningContainer:", rc, "error:", err)
				}
				if r < utilRatioLowerBound {
					rc, err := l.Shutdown(fn)
					log.Println("Shutdown RunningContainer:", rc, "error:", err)
				}
			}
		}
	}
}

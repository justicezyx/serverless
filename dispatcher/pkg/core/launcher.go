package core

import (
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"sync"
	"time"
)

// A notification sent to launcher to instruct it to launch a new instance.
// Send back the launched rc in the enclosed channel.
type launchNotification struct {
	// The function that needs new RunningContainer.
	fn string

	// The channel used to receive the created RunningContainer.
	rcChan chan *RunningContainer
}

// Launcher stores containers for starting instances to serve function invocations.
// TODO: Needs sync.Mutex to protect from concurrent access.
type Launcher struct {
	// Set during creation, and never change afterwards.
	// Therefore mutex protection is not required.
	//
	// Map from the function to the Container template.
	// ContainerInterface is used for testing.
	fnContainerMap map[string]ContainerInterface

	// The counter of running container created for function.
	fnContainerNameCounter map[string]int

	// A map from the function to the corresponding running container instances.
	// Picking any one of these instances for serving the function.
	fnInstsMapMu sync.Mutex
	fnInstsMap   map[string][]*RunningContainer

	// Notify launcher to immediately start an instance for the received function.
	launchNotifier  chan launchNotification
	stopMonitorChan chan struct{}

	// The interval for periodically check the load on each RunningContainer.
	checkInterval time.Duration
}

func NewLauncher(interval time.Duration) Launcher {
	return Launcher{
		fnContainerMap:         make(map[string]ContainerInterface),
		fnContainerNameCounter: make(map[string]int),
		fnInstsMap:             make(map[string][]*RunningContainer),
		launchNotifier:         make(chan launchNotification),
		stopMonitorChan:        make(chan struct{}),
		checkInterval:          interval,
	}
}

func (l *Launcher) debugLog() {
	for fn, rcs := range l.fnInstsMap {
		for _, rc := range rcs {
			log.Println(fn, rc.name)
		}
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
	counter, ok := d.fnContainerNameCounter[fn]
	if !ok {
		counter = 0
	}
	name := fn + "-" + strconv.Itoa(counter)
	d.fnContainerNameCounter[fn] = counter + 1
	rc, err := c.Run(name)
	if err != nil {
		return nil, fmt.Errorf("Could not run container for function: %s, error: %v", fn, err)
	}
	rcs, ok := d.fnInstsMap[fn]
	if !ok {
		rcs = make([]*RunningContainer, 0)
	}
	d.fnInstsMap[fn] = append(rcs, rc)
	log.Println("After launching an instance")
	d.debugLog()
	return rc, nil
}

func (l *Launcher) InstsCount(fn string) int {
	l.fnInstsMapMu.Lock()
	defer l.fnInstsMapMu.Unlock()

	rcs, ok := l.fnInstsMap[fn]
	if !ok {
		return 0
	}
	return len(rcs)
}

func (l *Launcher) hasUnrdyInsts(fn string) bool {
	l.fnInstsMapMu.Lock()
	defer l.fnInstsMapMu.Unlock()

	rcs, ok := l.fnInstsMap[fn]
	if !ok {
		return false
	}
	for _, rc := range rcs {
		if !rc.IsReady() {
			return true
		}
	}
	return false
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
	for i, rc := range rcs {
		if rc.launchTime.After(youngest.launchTime) {
			log.Println("youngest", *youngest, "idx", idx, "i", i)
			youngest = rc
			idx = i
			log.Println("youngest", *youngest, "idx", idx, "i", i)
		}
	}

	log.Println("youngest", *youngest, "idx", idx)
	// Remove the RunningContainer from the map
	l.debugLog()
	rcs[idx] = rcs[len(rcs)-1]
	l.debugLog()
	rcs = rcs[:len(rcs)-1]
	l.debugLog()
	l.fnInstsMap[fn] = rcs
	l.debugLog()
	log.Println("after Launcher.Shutdown:")
	l.debugLog()

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
	d.fnInstsMapMu.Lock()
	defer d.fnInstsMapMu.Unlock()

	var wg sync.WaitGroup
	for _, rcs := range d.fnInstsMap {
		for _, rc := range rcs {
			wg.Add(1)
			go func() {
				defer wg.Done()
				log.Println("stopping and removing", *rc, "in Launcher::Shutdown")
				if err := rc.Stop(); err != nil {
					log.Println("Failed to stop running container:", rc)
				}
				if err := rc.Remove(); err != nil {
					log.Println("Failed to remove running container:", rc)
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
		// Avoid initial burst
		if totalBusyTime > 10*time.Second {
			res[fn] = float64(totalBusyTime) / float64(totalRdyTime)
		}
	}
	return res
}

const utilRatioUpperBound = 0.8
const utilRatioLowerBound = 0.7

// Loop forever to monitor the utilization, if it's too high, launch one new instance.
// If it's too low, shutdown instance.
func (l *Launcher) MonitorForever() {
	ticker := time.NewTicker(l.checkInterval)
	for {
		select {
		case n := <-l.launchNotifier:
			log.Println("Received launch notification, function:", n.fn)
			rc, err := l.Launch(n.fn)
			if err != nil {
				log.Println("Failed to launch container, function:", n.fn, "error:", err)
			}
			n.rcChan <- rc
		case _ = <-ticker.C:
			utilRatio := l.calUtilRatio()
			log.Println("Checking for utilization ratio:", utilRatio)
			for fn, r := range utilRatio {
				if l.hasUnrdyInsts(fn) {
					// Only check stable instances.
					continue
				}
				if r > utilRatioUpperBound {
					rc, err := l.Launch(fn)
					log.Println("Launched RunningContainer:", rc, "error:", err, "function:", fn)
				}
				if r < utilRatioLowerBound {
					if l.InstsCount(fn) > 1 {
						rc, err := l.Shutdown(fn)
						log.Println("Shutdown RunningContainer:", rc, "error:", err, "function:", fn)
					}
				}
			}
		case _ = <-l.stopMonitorChan:
			log.Println("Received stopMonitorChan")
			return
		}
	}
}

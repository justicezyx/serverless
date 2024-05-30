package core

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

type dispatcherConfig struct {
	// The default maximal count of container instances can be run for each function.
	maxInstCountPerFn map[string]int

	// If users have not configure per function limit, this will be used.
	defaultMaxInstCountPerFn int
}

// Dispatcher routes traffic into corresponding container instances, and can dynamically launch container instance when
// requests are high.
type Dispatcher struct {
	cfg dispatcherConfig

	// Launcher launches container instance on incoming requests.
	launcher Launcher

	// PermMgr checks user's permission to call function.
	permMgr PermMgr

	// APILimitMgr determine API call limits.
	apiLimitMgr APILimitMgr

	// APIUsageTracker tracks users' execution time of serving function invocations.
	apiUsageTracker APIUsageTracker
}

func NewDispatcher() Dispatcher {
	dispatcher := Dispatcher{
		launcher:        NewLauncher(time.Second),
		permMgr:         NewPermMgr(),
		apiLimitMgr:     NewAPILimitMgr(3 /*default*/),
		apiUsageTracker: NewAPIUsageTracker(),
	}

	const runtimeImage = "runtime:4"

	alphaContainer := Container{
		image: runtimeImage,
		cmd:   []string{"python", "runtime.py", "--file=runtime_alpha.py", "--class_name=RuntimeAlpha"},
	}

	betaContainer := Container{
		image: runtimeImage,
		cmd:   []string{"python", "runtime.py", "--file=runtime_beta.py", "--class_name=RuntimeBeta"},
	}

	dispatcher.cfg.defaultMaxInstCountPerFn = 3
	dispatcher.launcher.registerContainer("alpha", alphaContainer)
	dispatcher.launcher.registerContainer("beta", betaContainer)
	// Start the monitoring goroutine to constantly watch utnization ratio and launch/shutdown instance accordingly.
	go dispatcher.launcher.MonitorForever()

	dispatcher.permMgr.AllowUserAPI("test", "alpha")
	dispatcher.permMgr.AllowUserAPI("test", "beta")

	return dispatcher
}

func (d *Dispatcher) getMaxinstCountPerFn(fn string) int {
	if limit, ok := d.cfg.maxInstCountPerFn[fn]; ok {
		return limit
	}
	return d.cfg.defaultMaxInstCountPerFn
}

func (d *Dispatcher) SetAPIConcurLimit(limit int64) {
	d.apiLimitMgr.SetLimit(limit)
}

func (d *Dispatcher) GetAPILimitMgr() *APILimitMgr {
	return &d.apiLimitMgr
}

func (d *Dispatcher) Shutdown() {
	d.launcher.ShutdownAll()
}

// Contextual information of serving a serverless function call.
type CallContext struct {
	// The name of the function to invoke.
	Fn string

	// The timeout waiting for the function instance to become ready.
	InstRdyTimeout time.Duration
}

// Option #1: Queue requests, and let another processor goroutine to fetch request, and send back responses.
//
//	type FnInvocation {
//		 ctx CallContext
//		 w http.ResponseWriter
//		 r *http.Request
//	}
//
//	May cause starvation.
//
// Option #2: Handle requests inside Dispatch, and wait for another goroutine to start new instances.
func (d *Dispatcher) Dispatch(ctx CallContext, w http.ResponseWriter, r *http.Request) {
	user := r.Header.Get("User")
	if user == "" {
		http.Error(w, "User header not provided", http.StatusBadRequest)
		return
	}

	if !d.permMgr.IsUserAllowed(user, ctx.Fn) {
		http.Error(w, fmt.Sprintf("User %s is not allowed to call function %s", user, ctx.Fn), http.StatusForbidden)
		return
	}

	rc, err := d.launcher.PickInst(ctx.Fn)

	if err != nil {
		log.Println("Cold start, need to create an instance for function:", ctx.Fn)
		for {
			rcChan := make(chan *RunningContainer)
			d.launcher.launchNotifier <- launchNotification{ctx.Fn, rcChan}
			rc = <-rcChan
			if rc != nil {
				break
			}
		}
	}

	err = rc.WaitForReady(ctx.InstRdyTimeout)
	if err != nil {
		http.Error(w, fmt.Sprintf("Timeout waiting for the instance to become ready, error: %v", ctx.Fn, err),
			http.StatusInternalServerError)
		return
	}

	d.apiLimitMgr.StartAPICall(ctx.Fn, 10*time.Second)
	apiStartTime := d.apiUsageTracker.StartAPICall(user)
	ProxyRequest(rc.Url, w, r)
	d.apiUsageTracker.EndAPICall(user, apiStartTime)
	callDuration := time.Now().Sub(apiStartTime)
	rc.AddBusyTime(callDuration)
	d.apiLimitMgr.FinishAPICall(ctx.Fn)
}

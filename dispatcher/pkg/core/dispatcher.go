package core

import (
	"fmt"
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
		launcher:        NewLauncher(),
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

	// TODO/Req: Before calling PickUrl(), should add an API to Launcher to determine if a new RunningContainer should be
	// launched. Candidate: Launcher::LaunchNewInstances()
	target, err := d.launcher.PickUrl(ctx.Fn)
	if err != nil {
		// Indicating there is no running container instances for this function.
		// Need to launch new ones.
		launchErr := d.launcher.Launch(ctx.Fn)
		if launchErr != nil {
			http.Error(w, fmt.Sprintf("Could not launch container instance for function '%s', error: %v", ctx.Fn, launchErr),
				http.StatusInternalServerError)
			return
		}
	}
	target, err = d.launcher.PickUrl(ctx.Fn)
	// TODO/Req: Wait for the launched instance to be ready. Launcher.Launch() should return a RunningContainer object for
	// checking readiness. PickUrl() should return a RunningContainer object, and let the caller to wait for readiness.

	if err != nil {
		http.Error(w, fmt.Sprintf("Could not find container instance for function '%s', error: %v", ctx.Fn, err),
			http.StatusInternalServerError)
		return
	}

	d.apiLimitMgr.StartAPICall(ctx.Fn, 10*time.Second)
	apiStartTime := d.apiUsageTracker.StartAPICall(user)
	ProxyRequest(target, w, r)
	d.apiUsageTracker.EndAPICall(user, apiStartTime)
	d.apiLimitMgr.FinishAPICall(ctx.Fn)
}

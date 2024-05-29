package core

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

// Dispatcher routes traffic into corresponding container instances, and can dynamically launch container instance when
// requests are high.
type Dispatcher struct {
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

	dispatcher.launcher.registerContainer("alpha", alphaContainer)
	dispatcher.launcher.registerContainer("beta", betaContainer)

	dispatcher.permMgr.AllowUserAPI("test", "alpha")
	dispatcher.permMgr.AllowUserAPI("test", "beta")

	return dispatcher
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
	fn string

	// The timeout waiting for the function instance to be ready.
	instRdyTimeout time.Duration
}

func (d *Dispatcher) Dispatch(ctx CallContext, w http.ResponseWriter, r *http.Request) {
	user := r.Header.Get("User")
	if user == "" {
		http.Error(w, "User header not provided", http.StatusBadRequest)
		return
	}

	if !d.permMgr.IsUserAllowed(user, ctx.fn) {
		http.Error(w, fmt.Sprintf("User %s is not allowed to call function %s", user, fn), http.StatusForbidden)
		return
	}

	// TODO/Req: Before calling PickUrl(), should add an API to Launcher to determine if a new RunningContainer should be
	// launched. Candidate: Launcher::LaunchNewInstances()
	target, err := d.launcher.PickUrl(ctx.fn)
	if err != nil {
		// Indicating there is no running container instances for this function.
		// Need to launch new ones.
		launchErr := d.launcher.Launch(ctx.fn)
		if launchErr != nil {
			log.Println("launchErr", launchErr)
			http.Error(w, fmt.Sprintf("Could not launch container instance for function '%s', error: %v", ctx.fn, launchErr),
				http.StatusInternalServerError)
			return
		}
	}
	target, err = d.launcher.PickUrl(ctx.fn)
	// TODO/Req: Wait for the launched instance to be ready. Launcher.Launch() should return a RunningContainer object for
	// checking readiness. PickUrl() should return a RunningContainer object, and let the caller to wait for readiness.

	if err != nil {
		http.Error(w, fmt.Sprintf("Could not find container instance for function '%s', error: %v", ctx.fn, err),
			http.StatusInternalServerError)
		return
	}

	d.apiLimitMgr.StartAPICall(ctx.fn, 10*time.Second)
	apiStartTime := d.apiUsageTracker.StartAPICall(user)
	log.Println("Before ProxyRequest")
	ProxyRequest(target, w, r)
	log.Println("After ProxyRequest")
	d.apiUsageTracker.EndAPICall(user, apiStartTime)
	d.apiLimitMgr.FinishAPICall(ctx.fn)
}

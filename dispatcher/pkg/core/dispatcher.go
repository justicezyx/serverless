package core

import (
	"fmt"
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

	// APILimitMgr
	apiMgr          APILimitMgr
	apiUsageTracker APIUsageTracker
}

func NewDispatcher() Dispatcher {
	dispatcher := Dispatcher{
		launcher:        NewLauncher(),
		permMgr:         NewPermMgr(),
		apiMgr:          NewAPILimitMgr(3 /*default*/),
		apiUsageTracker: NewAPIUsageTracker(),
	}

	const runtimeImage = "runtime:latest"

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
	d.apiMgr.SetLimit(limit)
}

func (d *Dispatcher) GetAPILimitMgr() *APILimitMgr {
	return &d.apiMgr
}

func (d *Dispatcher) Shutdown() {
	d.launcher.ShutdownAll()
}

func (d *Dispatcher) Dispatch(fn string, w http.ResponseWriter, r *http.Request) {
	user := r.Header.Get("User")
	if user == "" {
		http.Error(w, "User header not provided", http.StatusBadRequest)
		return
	}

	if !d.permMgr.IsUserAllowed(user, fn) {
		http.Error(w, fmt.Sprintf("User %s is not allowed to call function %s", user, fn), http.StatusForbidden)
		return
	}

	target, err := d.launcher.PickUrl(fn)
	if err != nil {
		// Launch new instances
		launchErr := d.launcher.Launch(fn)
		if launchErr != nil {
			http.Error(w, fmt.Sprintf("Could not launch container instance for function '%s', error: %v", fn, launchErr),
				http.StatusInternalServerError)
			return
		}
	}
	target, err = d.launcher.PickUrl(fn)
	if err != nil {
		http.Error(w, fmt.Sprintf("Could not find container instance for function '%s', error: %v", fn, err),
			http.StatusInternalServerError)
		return
	}

	d.apiMgr.StartAPICall(fn, 10*time.Second)
	apiStartTime := d.apiUsageTracker.StartAPICall(user)
	ProxyRequest(target, w, r)
	d.apiUsageTracker.EndAPICall(user, apiStartTime)
	d.apiMgr.FinishAPICall(fn)
}

package core

import (
	"fmt"
	"net/http"
	"time"
)

const runtimeImage = "runtime:latest"

var launcher Launcher

var permMgr PermMgr

var apiMgr APIMgr

func initLauncher() {
	launcher = NewLauncher()

	alphaContainer := Container{
		image: runtimeImage,
		cmd:   []string{"python", "runtime.py", "--file=runtime_alpha.py", "--class_name=RuntimeAlpha"},
	}

	betaContainer := Container{
		image: runtimeImage,
		cmd:   []string{"python", "runtime.py", "--file=runtime_beta.py", "--class_name=RuntimeBeta"},
	}

	launcher.registerContainer("alpha", alphaContainer)
	launcher.registerContainer("beta", betaContainer)
}

func initPermMgr() {
	permMgr = NewPermMgr()

	permMgr.AllowUserAPI("test", "alpha")
	permMgr.AllowUserAPI("test", "beta")
}

func initAPIMgr() {
	apiMgr = NewAPIMgr(3 /*default*/)
}

func init() {
	initLauncher()
	initPermMgr()
	initAPIMgr()
}

func SetAPIConcurLimit(limit int64) {
	apiMgr.SetLimit(limit)
}

func GetAPIMgr() *APIMgr {
	return &apiMgr
}

func Dispatch(fn string, w http.ResponseWriter, r *http.Request) {
	user := r.Header.Get("User")
	if user == "" {
		http.Error(w, "User header not provided", http.StatusBadRequest)
		return
	}

	if !permMgr.IsUserAllowed(user, fn) {
		http.Error(w, fmt.Sprintf("User %s is not allowed to call function %s", user, fn), http.StatusForbidden)
		return
	}

	target, err := launcher.PickUrl(fn)
	if err != nil {
		// Launch new instances
		launchErr := launcher.Launch(fn)
		if launchErr != nil {
			http.Error(w, fmt.Sprintf("Could not launch container instance for function '%s', error: %v", fn, launchErr),
				http.StatusInternalServerError)
			return
		}
	}
	target, err = launcher.PickUrl(fn)
	if err != nil {
		http.Error(w, fmt.Sprintf("Could not find container instance for function '%s', error: %v", fn, err),
			http.StatusInternalServerError)
		return
	}

	apiMgr.StartAPICall(fn, 10*time.Second)
	ProxyRequest(target, w, r)
	apiMgr.FinishAPICall(fn)
}

func Shutdown() {
	launcher.ShutdownAll()
}

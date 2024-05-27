package core

import (
	"fmt"
	"net/http"
)

const runtimeImage = "runtime:latest"

var launcher Launcher

func init() {
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

func Dispatch(fn string, w http.ResponseWriter, r *http.Request) {
	// TODO: Launching instances and return the URL.
	target, err := launcher.PickUrl(fn)
	if err != nil {
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
	ProxyRequest(target, w, r)
}

func Shutdown() {
	launcher.ShutdownAll()
}

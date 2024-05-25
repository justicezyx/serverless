package dispatcher

import (
	"fmt"
	"net"
	"os/exec"
)

// Launcher launches a subprocess to run the cmd.
type Launcher struct {
	cmd []string

	handle *exec.Cmd
}

func NewLauncher(cmd []string) Launcher {
	if len(cmd) <= 0 {
		panic(fmt.Sprintf("Illformed command: %v", cmd))
	}
	return Launcher {
		cmd: cmd,
	}
}

// Returns a randomly-picked port. The port can be used by another service to listen on.
func pickPort() (int, error) {
	// Listen on a random port by specifying port 0
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		return 0, fmt.Errorf("Error listening on port: %v\n", err)
	}
	// Get the assigned port
	port := listener.Addr().(*net.TCPAddr).Port
	// Close the listener.
	listener.Close()
	return port, nil
}

// Launch subprocess and get the PID
func (l *Launcher) Launch() error {
	fmt.Println("launch")
	l.handle = exec.Command(l.cmd[0], l.cmd[1:]...)
	fmt.Println("handle", l.handle)
	err := l.handle.Start()
	fmt.Println("start")
	if err != nil {
		return err
	}
	return nil
}

func (l *Launcher) Kill() error {
	// The process has not been launched.
	if l.handle == nil {
		return nil
	}
	if l.handle.Process != nil {
		return l.handle.Process.Kill()
	}
	return fmt.Errorf("subprocess not running")
}

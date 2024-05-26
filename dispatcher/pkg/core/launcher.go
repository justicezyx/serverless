package core

import (
	"fmt"
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
	return Launcher{
		cmd: cmd,
	}
}

// Launch subprocess and get the PID
func (l *Launcher) Launch(port int) error {
	l.handle = exec.Command(l.cmd[0], l.cmd[1:]...)
	err := l.handle.Start()
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

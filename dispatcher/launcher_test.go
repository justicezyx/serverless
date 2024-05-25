package dispatcher

import (
	"testing"
)

func TestLauncherLaunchKill(t *testing.T) {
	l := NewLauncher([]string{"echo"})
	if err := l.Launch(); err != nil {
		t.Errorf("Launch failed, error: %v", err)
	}
	if err := l.Kill(); err != nil {
		t.Errorf("Kill failed, error: %v", err)
	}
}

func TestPickPort(t *testing.T) {
	if port, err := pickPort(); port == 0 || err != nil {
		t.Errorf("Launch failed, port: %d, error: %v", port, err)
	}
}

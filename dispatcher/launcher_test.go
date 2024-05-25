package dispatcher

import (
	"testing"
)

func TestLauncherPickPort(t *testing.T) {
	l := NewLauncher([]string{"echo"})
	if err := l.Launch(); err != nil {
		t.Errorf("Launch failed, error: %v", err)
	}
	if err := l.Kill(); err != nil {
		t.Errorf("Kill failed, error: %v", err)
	}
}

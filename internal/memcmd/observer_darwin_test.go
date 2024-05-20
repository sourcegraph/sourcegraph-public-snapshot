//go:build darwin

package memcmd

import (
	"context"
	"os/exec"
)

func TestMaxMemoryUsageErrorObserverNotStarted(t *testing.T) {
	cmd := allocatingGoProgram(t, 50*1024*1024) // 50 MB

	observer, err := NewMacObserver(cmd)
	if err != nil {
		t.Fatalf("failed to create observer: %v", err)
	}

	_, err = observer.MaxMemoryUsage()
	if !errors.Is(err, errObserverNotStarted) {
		t.Errorf("expected errObserverNotStarted, got: %v", err)
	}
}

func TestMaxMemoryUsageErrorProcessNotCompleted(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cmd := exec.CommandContext(ctx, "sleep", "10s")
	err := cmd.Start()
	if err != nil {
		t.Fatalf("failed to start test program: %v", err)
	}

	defer func() {
		_ = cmd.Process.Kill()
	}()

	observer, err := NewMacObserver(cmd)
	if err != nil {
		t.Fatalf("failed to create observer: %v", err)
	}

	observer.Start()
	defer observer.Stop()

	_, err = observer.MaxMemoryUsage()
	if !errors.Is(err, errProcessNotStopped) {
		t.Errorf("expected errProcessNotStopped, got: %v", err)
	}
}

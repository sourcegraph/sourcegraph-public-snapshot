//go:build darwin

package memcmd

import (
	"bytes"
	"context"
	"os/exec"
	"syscall"
	"testing"

	"github.com/dustin/go-humanize"

	"github.com/sourcegraph/sourcegraph/internal/bytesize"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestNewMacObserverIntegration(t *testing.T) {
	cmd := allocatingGoProgram(t, 250*1024*1024)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	var buf bytes.Buffer
	cmd.Stderr = &buf
	err := cmd.Start()
	if err != nil {
		t.Fatalf("failed to start test program: %v, stdErr: %s", err, buf.String())
	}

	observer, err := NewMacObserver(cmd)
	if err != nil {
		t.Fatalf("failed to create observer: %v", err)
	}

	observer.Start()
	defer observer.Stop()

	err = cmd.Wait()
	if err != nil {
		t.Fatalf("failed to wait for test program: %v stdErr: %s", err, buf.String())
	}

	memoryUsage, err := observer.MaxMemoryUsage()
	if err != nil {
		t.Fatalf("failed to get memory usage: %v", err)
	}

	t.Logf("memory usage: %s", humanize.Bytes(uint64(memoryUsage)))
	memoryLow := 200 * bytesize.MB
	memoryHigh := 300 * bytesize.MB

	if !(memoryLow < memoryUsage && memoryUsage < memoryHigh) {
		t.Fatalf("memory usage is not in the expected range (low: %s, high: %s): %s", humanize.Bytes(uint64(memoryLow)), humanize.Bytes(uint64(memoryHigh)), humanize.Bytes(uint64(memoryUsage)))
	}
}

func TestMaxMemoryUsageErrorObserverNotStarted(t *testing.T) {
	cmd := allocatingGoProgram(t, 50*1024*1024) // 50 MB
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

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

	_, err = observer.MaxMemoryUsage()
	if !errors.Is(err, errObserverNotStarted) {
		t.Errorf("expected errObserverNotStarted, got: %v", err)
	}
}

func TestMaxMemoryUsageErrorProcessNotCompleted(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cmd := exec.CommandContext(ctx, "sleep", "10s")
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
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

func TestComplainAboutProcessNotWithinOwnGroup(t *testing.T) {
	cmd := exec.Command("sleep", "10s")
	err := cmd.Start()
	if err != nil {
		t.Fatalf("failed to start test program: %v", err)
	}

	defer func() {
		_ = cmd.Process.Kill()
	}()

	_, err = NewMacObserver(cmd)
	if !errors.Is(err, errProcessNotWithinOwnProcessGroup) {
		t.Errorf("expected errProcessNotWithinOwnProcessGroup, got: %v", err)
	}
}

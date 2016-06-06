package executil

import (
	"os/exec"
	"testing"
	"time"
)

func TestCmdCombinedOutputWithTimeout_timeout(t *testing.T) {
	out, err := CmdCombinedOutputWithTimeout(200*time.Millisecond, exec.Command("sh", "-c", "echo hello && sleep 0.201"))
	if want := ErrCmdTimeout; err != want {
		t.Errorf("got error %v, want %v", err, want)
	}
	if want := "hello\n"; string(out) != want {
		t.Errorf("got output %q, want %q", out, want)
	}
}

func TestCmdCombinedOutputWithTimeout_ok(t *testing.T) {
	out, err := CmdCombinedOutputWithTimeout(200*time.Millisecond, exec.Command("sh", "-c", "echo hello"))
	if err != nil {
		t.Fatal(err)
	}
	if want := "hello\n"; string(out) != want {
		t.Errorf("got output %q, want %q", out, want)
	}
}

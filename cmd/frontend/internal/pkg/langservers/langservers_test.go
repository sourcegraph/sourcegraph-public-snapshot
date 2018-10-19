package langservers

import (
	"os"
	"os/exec"
	"testing"
	"time"
)

func TestStaticInfo_SiteConfig_Language(t *testing.T) {
	// Sanity check that the siteConfig language fields match their map keys,
	// as typos have caused issues here in the past:
	//
	// https://github.com/sourcegraph/sourcegraph/issues/10671
	//
	for lang, staticInfo := range StaticInfo {
		if lang != staticInfo.SiteConfig.Language {
			t.Fatalf("mismatched StaticInfo entry found; lang %q != siteConfig.Language %q", lang, staticInfo.SiteConfig.Language)
		}
	}
}

func TestStaticInfo_debugContainerPorts(t *testing.T) {
	// Sanity check that the languages in StaticInfo and debugContainerPorts are
	// the same.

	for key := range StaticInfo {
		if _, ok := debugContainerPorts[key]; !ok {
			t.Fatalf("debugContainerPorts is missing a key from StaticInfo: %s", key)
		}
	}

	for key := range debugContainerPorts {
		if _, ok := StaticInfo[key]; !ok {
			t.Fatalf("StaticInfo is missing a key from debugContainerPorts: %s", key)
		}
	}
}

func TestDebugContainerPorts_unique(t *testing.T) {
	// Sanity check that all the ports in debugContainerPorts are unique.

	allPorts := make(map[string]string)

	for language, ports := range debugContainerPorts {
		if _, ok := allPorts[ports.HostPort]; ok && language != "javascript" && language != "typescript" {
			t.Fatalf("Languages %s and %s can't both listen on port %s.", language, allPorts[ports.HostPort], ports.HostPort)
		}

		allPorts[ports.HostPort] = language
	}
}

func TestNotifyNewLine_missing(t *testing.T) {
	c := make(chan struct{}, 1)
	err := notifyNewLine(exec.Command("IDoNotExistOnPATH"), c)
	if err == nil {
		t.Fatal("expected error")
	}
	select {
	case <-c:
		t.Fatal("did not expect an event")
	default:
	}
}

func TestNotifyNewLine_failed(t *testing.T) {
	c := make(chan struct{}, 1)
	err := notifyNewLine(exec.Command("false"), c)
	if err == nil {
		t.Fatal("expected error")
	}
	select {
	case <-c:
	case <-time.After(time.Second):
		t.Fatal("Expected an event")
	}
}

func TestNotifyNewLine_started(t *testing.T) {
	c := make(chan struct{}, 10) // large buffer to ensure we don't miss an event
	cmd := exec.Command("sh", "-c", `sleep 10 & PID=$!; trap "kill $PID" INT; echo a; wait; echo b`)
	var err error
	go func() {
		err = notifyNewLine(cmd, c)
		close(c)
	}()
	wantEvent := func(reason string) {
		t.Helper()
		select {
		case <-c:
		case <-time.After(time.Second):
			t.Fatal("Expected an event: ", reason)
		}
	}

	wantEvent("started event")
	wantEvent("initial output")

	// We do not expect an event until we send sigint
	select {
	case <-c:
		t.Fatal("unexpected event")
	case <-time.After(50 * time.Millisecond):
	}

	cmd.Process.Signal(os.Interrupt)
	wantEvent("final output")

	_, ok := <-c
	if ok {
		t.Fatal("expected c to be closed by helper")
	}

	if err != nil {
		t.Fatal("unexpected error", err)
	}
}

func TestNotifyNewLine_nonblocking(t *testing.T) {
	err := notifyNewLine(exec.Command("sh", "-c", "yes | head"), make(chan struct{}))
	if err != nil {
		t.Fatal("unexpected error", err)
	}
}

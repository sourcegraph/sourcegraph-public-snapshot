package server

import (
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/sourcegraph/sourcegraph/cmd/gitserver/server/common"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/stretchr/testify/require"
)

func TestFlushingResponseWriter(t *testing.T) {
	flush := make(chan struct{})
	fw := &flushingResponseWriter{
		w: httptest.NewRecorder(),
		flusher: flushFunc(func() {
			flush <- struct{}{}
		}),
	}
	done := make(chan struct{})
	go func() {
		fw.periodicFlush()
		close(done)
	}()

	_, _ = fw.Write([]byte("hi"))

	select {
	case <-flush:
		close(flush)
	case <-time.After(5 * time.Second):
		t.Fatal("periodic flush did not happen")
	}

	fw.Close()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("periodic flush goroutine did not close")
	}
}

type flushFunc func()

func (f flushFunc) Flush() {
	f()
}

func TestPoolDirFromName(t *testing.T) {
	reposDir := t.TempDir()

	repoName := api.RepoName("github.com/sourcegraph/sourcegraph")
	got := poolDirFromName(reposDir, repoName)
	want := common.GitDir(filepath.Join(reposDir, ".pool/github.com/sourcegraph/sourcegraph/.git"))
	require.Equal(t, got, want)
}

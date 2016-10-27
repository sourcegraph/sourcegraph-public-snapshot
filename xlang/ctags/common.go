package ctags

import (
	"context"
	"path/filepath"
	"sync"

	"github.com/sourcegraph/ctxvfs"

	"sourcegraph.com/sourcegraph/sourcegraph/xlang/ctags/parser"
)

func isSupportedFile(mode, filename string) bool {
	ext := filepath.Ext(filename)
	switch mode {
	case "c":
		return ext == ".c" || ext == ".h"
	case "ruby":
		return ext == ".rb"
	default:
		return false
	}
}

// ConnectionInfo caches information for a JSONRPC connection. It is unique to
// a (access scope, repository, version), where access scope be a specific
// user, a not logged in user, etc.
type ConnectionInfo struct {
	// fs is the virtual filesystem backed by xlang infrastrucuture.
	fs ctxvfs.FileSystem

	// tagsMu protects tags. We want to be careful to not run ctags more than
	// once for one project, so this is used in the get tags method.
	tagsMu sync.Mutex

	// tags is the Go form of the ctags output for this project. We compute and
	// save it so that we don't have to parse the ctags file each time, and so
	// we don't have to store as much state on disk.
	tags []parser.Tag

	// mode is the language that we care about for this connection.
	mode string
}

type ctxKey struct{}

func InitCtx(ctx context.Context) context.Context {
	info := new(ConnectionInfo)
	return context.WithValue(ctx, ctxKey{}, info)
}

func ctxInfo(ctx context.Context) *ConnectionInfo {
	return ctx.Value(ctxKey{}).(*ConnectionInfo)
}

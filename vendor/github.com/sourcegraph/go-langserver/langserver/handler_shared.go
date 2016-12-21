package langserver

import (
	"context"
	"fmt"
	"go/build"
	"strings"
	"sync"

	"github.com/sourcegraph/ctxvfs"
)

// HandlerShared contains data structures that a build server and its
// wrapped lang server may share in memory.
type HandlerShared struct {
	Mu     sync.Mutex // guards all fields
	Shared bool       // true if this struct is shared with a build server
	FS     *AtomicFS  // full filesystem (mounts both deps and overlay)

	// FindPackage if non-nil is used by our typechecker. See
	// loader.Config.FindPackage. We use this in production to lazily
	// fetch dependencies + cache lookups.
	FindPackage FindPackageFunc

	overlayFSMu      sync.Mutex        // guards overlayFS map
	overlayFS        map[string][]byte // files to overlay
	OverlayMountPath string            // mount point of overlay on fs (usually /src/github.com/foo/bar)
}

// FindPackageFunc matches the signature of loader.Config.FindPackage, except
// also takes a context.Context.
type FindPackageFunc func(ctx context.Context, bctx *build.Context, importPath, fromDir string, mode build.ImportMode) (*build.Package, error)

func defaultFindPackageFunc(ctx context.Context, bctx *build.Context, importPath, fromDir string, mode build.ImportMode) (*build.Package, error) {
	return bctx.Import(importPath, fromDir, mode)
}

func (h *HandlerShared) Reset(overlayRootURI string, useOSFS bool) error {
	h.Mu.Lock()
	defer h.Mu.Unlock()
	h.overlayFSMu.Lock()
	defer h.overlayFSMu.Unlock()
	h.overlayFS = map[string][]byte{}
	h.FS = NewAtomicFS()

	if !strings.HasPrefix(overlayRootURI, "file:///") {
		return fmt.Errorf("invalid overlay root URI %q: must be file:///", overlayRootURI)
	}
	h.OverlayMountPath = strings.TrimPrefix(overlayRootURI, "file://")
	if useOSFS {
		// The overlay FS takes precedence, but we fall back to the OS
		// file system.
		h.FS.Bind("/", ctxvfs.OS("/"), "/", ctxvfs.BindAfter)
	}
	h.FS.Bind("/", ctxvfs.Sync(&h.overlayFSMu, ctxvfs.Map(h.overlayFS)), "/", ctxvfs.BindBefore)
	return nil
}

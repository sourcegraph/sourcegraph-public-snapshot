package langserver

import (
	"fmt"
	"strings"
	"sync"

	"github.com/sourcegraph/ctxvfs"
)

// HandlerShared contains data structures that a build server and its
// wrapped lang server may share in memory.
type HandlerShared struct {
	Mu     sync.Mutex       // guards all fields
	Shared bool             // true if this struct is shared with a build server
	FS     ctxvfs.NameSpace // full filesystem (mounts both deps and overlay)

	overlayFSMu      sync.Mutex        // guards overlayFS map
	overlayFS        map[string][]byte // files to overlay
	OverlayMountPath string            // mount point of overlay on fs (usually /src/github.com/foo/bar)

	// AugmentFileSystem, if set, is called on the prepared file
	// system before use.
	AugmentFileSystem func(fs ctxvfs.FileSystem) ctxvfs.FileSystem
}

func (h *HandlerShared) Reset(overlayRootURI string) error {
	h.Mu.Lock()
	defer h.Mu.Unlock()
	h.overlayFSMu.Lock()
	defer h.overlayFSMu.Unlock()
	h.overlayFS = map[string][]byte{}
	h.FS = ctxvfs.NameSpace{}

	if !strings.HasPrefix(overlayRootURI, "file:///") {
		return fmt.Errorf("invalid overlay root URI %q: must be file:///", overlayRootURI)
	}
	h.OverlayMountPath = strings.TrimPrefix(overlayRootURI, "file://")
	h.FS.Bind(h.OverlayMountPath, ctxvfs.Sync(&h.overlayFSMu, ctxvfs.Map(h.overlayFS)), "/", ctxvfs.BindBefore)
	return nil
}

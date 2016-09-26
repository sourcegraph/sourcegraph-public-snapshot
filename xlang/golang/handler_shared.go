package golang

import (
	"fmt"
	"strings"
	"sync"

	"golang.org/x/tools/godoc/vfs"
	"sourcegraph.com/sourcegraph/sourcegraph/xlang/vfsutil/mapfs"
)

// handlerShared contains data structures that a build server and its
// wrapped lang server may share in memory.
type handlerShared struct {
	mu               sync.Mutex        // guards all fields
	fs               namespaceFS       // full filesystem (mounts both deps and overlay)
	overlayFS        map[string][]byte // files to overlay
	overlayMountPath string            // mount point of overlay on fs (usually /src/github.com/foo/bar)
}

func (h *handlerShared) reset(overlayRootURI string) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.overlayFS = map[string][]byte{}
	h.fs = namespaceFS{vfs.NewNameSpace()}

	if !strings.HasPrefix(overlayRootURI, "file:///") {
		return fmt.Errorf("invalid overlay root URI %q: must be file:///", overlayRootURI)
	}
	h.overlayMountPath = strings.TrimPrefix(overlayRootURI, "file://")
	h.fs.Bind(h.overlayMountPath, mapfs.New(h.overlayFS), "/", vfs.BindBefore)
	return nil
}

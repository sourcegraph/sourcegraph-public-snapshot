package server

import (
	"fmt"

	"github.com/sourcegraph/ctxvfs"
	"github.com/sourcegraph/sourcegraph/pkg/gosrc"
)

// addSysZversionFile adds the zversion.go file, which is generated
// during the Go release process and does not exist in the VCS repo
// archive zips. We need to create it here, or else we'll see
// typechecker errors like "StackGuardMultiplier not declared by
// package sys" when any packages import from the Go stdlib.
func addSysZversionFile(fs ctxvfs.FileSystem) ctxvfs.FileSystem {
	return ctxvfs.SingleFileOverlay(fs,
		"/src/runtime/internal/sys/zversion.go",
		[]byte(fmt.Sprintf(`
package sys

const DefaultGoroot = %q
const TheVersion = %q
const Goexperiment=""
const StackGuardMultiplier=1`,
			goroot, gosrc.RuntimeVersion)))
}

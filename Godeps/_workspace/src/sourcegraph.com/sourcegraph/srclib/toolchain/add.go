package toolchain

import (
	"fmt"
	"os"
	"path/filepath"

	"sourcegraph.com/sourcegraph/srclib"
)

type AddOpt struct {
	// Force add a toolchain, overwriting any existing toolchains.
	Force bool
}

// Add creates a symlink in the SRCLIBPATH so that the toolchain in dir is
// available at the toolchainPath.
func Add(dir, toolchainPath string, opt *AddOpt) error {
	if opt == nil {
		opt = &AddOpt{}
	}
	if !opt.Force {
		if _, err := Lookup(toolchainPath); err == nil {
			return fmt.Errorf("a toolchain already exists at toolchain path %q", toolchainPath)
		} else if !os.IsNotExist(err) {
			return err
		}
	}

	absDir, err := filepath.Abs(dir)
	if err != nil {
		return err
	}

	srclibpathEntry := filepath.SplitList(srclib.Path)[0]
	targetDir := filepath.Join(srclibpathEntry, toolchainPath)

	if err := os.MkdirAll(filepath.Dir(targetDir), 0700); err != nil {
		return err
	}

	if !opt.Force {
		return os.Symlink(absDir, targetDir)
	}
	// Force install the toolchain by removing the directory if
	// the symlink fails, and then try the symlink again.
	if err := os.Symlink(absDir, targetDir); err != nil {
		if err := os.RemoveAll(targetDir); err != nil {
			return err
		}
		return os.Symlink(absDir, targetDir)
	}
	return nil
}

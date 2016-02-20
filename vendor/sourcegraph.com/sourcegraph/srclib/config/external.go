package config

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"sourcegraph.com/sourcegraph/srclib"
	"sourcegraph.com/sourcegraph/srclib/toolchain"
)

// An External configuration file, represented by this struct, can set system-
// and user-level settings for srclib.
type External struct {
	// Scanners is the default set of scanners to use. If not specified, all
	// scanners in the SRCLIBPATH will be used.
	Scanners []*srclib.ToolRef
}

const srclibconfigFile = ".srclibconfig"

// SrclibPathConfig gets the srclib path configuration (which lists
// all available scanners). It reads it from SRCLIBPATH/.srclibconfig
// if that file exists, and otherwise it walks SRCLIBPATH for
// available scanners.
func SrclibPathConfig() (*External, error) {
	var x External

	// Try reading from .srclibconfig.
	dir := filepath.SplitList(srclib.Path)[0]
	configFile := filepath.Join(dir, srclibconfigFile)
	f, err := os.Open(configFile)
	if os.IsNotExist(err) {
		// do nothing
	} else if err != nil {
		log.Printf("Warning: unable to open config file at %s: %s. Continuing without this config.", configFile, err)
	} else {
		defer f.Close()
		if err := json.NewDecoder(f).Decode(&x); err != nil {
			log.Printf("Warning: unable to decode config file at %s: %s. Continuing without this config.", configFile, err)
		}
	}

	// Default to using all available scanners.
	if len(x.Scanners) == 0 {
		var err error
		x.Scanners, err = toolchain.ListTools("scan")
		if err != nil && !os.IsNotExist(err) {
			return nil, fmt.Errorf("failed to find scanners in SRCLIBPATH: %s", err)
		}
	}

	return &x, nil
}

package conf

import (
	"os"

	"github.com/sourcegraph/jsonx"
)

type configurationMode int

const (
	// The user of pkg/conf reads and writes to the configuration file.
	// This should only ever be used by frontend.
	modeServer configurationMode = iota

	//Tthe user of pkg/conf only reads the configuration file.
	modeClient
)

func getMode() configurationMode {
	mode := os.Getenv("CONFIGURATION_MODE")
	if mode == "server" {
		return modeServer
	}

	return modeClient
}

// FormatOptions is the default format options that should be used for jsonx
// edit computation.
var FormatOptions = jsonx.FormatOptions{InsertSpaces: true, TabSize: 2, EOL: "\n"}

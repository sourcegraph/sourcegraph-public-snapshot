package sgxcmd

import (
	"fmt"

	"github.com/sqs/go-selfupdate/selfupdate"
	"src.sourcegraph.com/sourcegraph/dev/release"
	"src.sourcegraph.com/sourcegraph/sgx/buildvar"
)

var (
	releaseBaseURL = "https://sourcegraph.com/.download/"

	// SelfUpdater checks and updates release versions.
	SelfUpdater = &selfupdate.Updater{
		CurrentVersion: buildvar.Version,
		ApiURL:         releaseBaseURL,
		BinURL:         releaseBaseURL,
		DiffURL:        releaseBaseURL,
		Dir:            release.S3Dir,
		CmdName:        Name,
	}
)

type binaryRelease struct {
	Name    string // name of the release, like "OS X (64-bit)"
	Version string // version of the release, like "0.5.6"
	URL     string // download URL for the binary
}

// BinaryReleaseURLs returns a list of URLs to the binary releases for
// various platforms. The binaries are of the same version as the
// currently running binary. Dev versions return an empty list.
func BinaryReleaseURLs() []binaryRelease {
	if buildvar.Version == "dev" {
		return nil
	}

	u := SelfUpdater

	var brs []binaryRelease
	for _, os := range oses {
		for _, arch := range archs {
			brs = append(brs, binaryRelease{
				Name:    fmt.Sprintf("%s (%s)", osNames[os], archNames[arch]),
				Version: buildvar.Version,
				URL:     u.ApiURL + buildvar.Version + "/" + os + "-" + arch + "/" + u.CmdName + ".gz",
			})
		}
	}

	return brs
}

var (
	oses    = []string{"linux", "darwin"}
	osNames = map[string]string{
		"linux":  "Linux",
		"darwin": "OS X",
	}

	archs     = []string{"amd64"}
	archNames = map[string]string{
		"amd64": "64-bit",
		"386":   "32-bit",
	}
)

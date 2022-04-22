package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"strings"

	"github.com/urfave/cli/v2"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/stdout"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var updateCommand = &cli.Command{
	Name:  "update",
	Usage: "Update local sg installation",
	Description: `Update local sg installation with the latest changes. To see what's new, run:

  sg version changelog -next

Requires a local copy of the 'sourcegraph/sourcegraph' codebase.`,
	Category: CategoryUtil,
	Action: func(cmd *cli.Context) error {
		if _, err := updateToPrebuiltSG(cmd.Context); err != nil {
			return err
		}
		writeSuccessLinef("sg has been updated!")
		stdout.Out.Write("To see what's new, run 'sg version changelog'.")
		return nil
	},
}

// updateToPrebuiltSG downloads the latest release of sg prebuilt binaries and install it.
func updateToPrebuiltSG(ctx context.Context) (string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://github.com/sourcegraph/sg/releases/latest", nil)
	if err != nil {
		return "", err
	}
	// We use the RountTripper to make an HTTP request without having to deal
	// with redirections.
	resp, err := http.DefaultTransport.RoundTrip(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	location := resp.Header.Get("location")
	if location == "" {
		return "", errors.New("GitHub latest release: empty location")
	}
	location = strings.ReplaceAll(location, "/tag/", "/download/")
	downloadURL := fmt.Sprintf("%s/sg_%s_%s", location, runtime.GOOS, runtime.GOARCH)

	tmpDir, err := os.MkdirTemp("", "sg")
	if err != nil {
		return "", err
	}
	defer func() {
		os.RemoveAll(tmpDir)
	}()

	resp, err = http.Get(downloadURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", errors.Newf("downloading sg: status %d", resp.StatusCode)
	}

	tmpSgPath := tmpDir + "/sg"
	f, err := os.Create(tmpSgPath)
	if err != nil {
		return "", err
	}

	_, err = io.Copy(f, resp.Body)
	if err != nil {
		return "", err
	}
	err = os.Chmod(tmpSgPath, 0755)
	if err != nil {
		return "", err
	}

	currentExecPath, err := os.Executable()
	if err != nil {
		return "", err
	}

	return currentExecPath, os.Rename(tmpSgPath, currentExecPath)
}

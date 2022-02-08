package shared

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/sourcegraph/sourcegraph/cmd/server/shared/assets"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// nginxProcFile will return a procfile entry for nginx, as well as setup
// configuration for it.
func nginxProcFile() (string, error) {
	configDir := os.Getenv("CONFIG_DIR")
	path, err := nginxWriteFiles(configDir)
	if err != nil {
		return "", errors.Wrapf(err, "failed to generate nginx configuration to %s", configDir)
	}

	// This is set for the informational message we show once sourcegraph
	// frontend starts. This is so we can advertise the nginx address, rather
	// than the frontend address.
	SetDefaultEnv("SRC_NGINX_HTTP_ADDR", ":7080")

	return fmt.Sprintf(`nginx: nginx -p . -g 'daemon off;' -c %s 2>&1 | grep -v 'could not open error log file' 1>&2`, path), nil
}

// nginxWriteFiles writes the nginx related configuration files to
// configDir. It returns the path to the main nginx.conf.
func nginxWriteFiles(configDir string) (string, error) {
	// Check we can read the config
	path := filepath.Join(configDir, "nginx.conf")
	_, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return "", err
	}

	// Does not exist
	if err != nil {
		err = os.WriteFile(path, []byte(assets.NginxConf), 0600)
		if err != nil {
			return "", err
		}
	}

	// We always write the files in the nginx directory, since those are
	// controlled by Sourcegraph and can change between versions.
	nginxDir := filepath.Join(configDir, "nginx")
	if err := os.MkdirAll(nginxDir, 0755); err != nil {
		return "", err
	}
	includeConfs, err := assets.NginxDir.ReadDir("nginx")
	if err != nil {
		return "", err
	}
	for _, p := range includeConfs {
		data, err := assets.NginxDir.ReadFile("nginx/" + p.Name())
		if err != nil {
			return "", err
		}
		err = os.WriteFile(filepath.Join(nginxDir, p.Name()), data, 0600)
		if err != nil {
			return "", err
		}
	}

	return path, nil
}

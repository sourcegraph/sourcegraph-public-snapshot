package shared

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/server/shared/assets"
)

var nginxConf = assets.MustAsset("nginx.conf")

// nginxProcFile will return a procfile entry for nginx, as well as setup
// configuration for it.
func nginxProcFile() (string, error) {
	// Check we can read the config
	path := filepath.Join(os.Getenv("CONFIG_DIR"), "nginx.conf")
	_, err := ioutil.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return "", errors.Wrapf(err, "failed to read nginx configuration at %s", path)
	}

	// Does not exist
	if err != nil {
		err = ioutil.WriteFile(path, []byte(nginxConf), 0600)
		if err != nil {
			return "", errors.Wrapf(err, "failed to generate nginx configuration to %s", path)
		}
	}

	// This is set for the informational message we show once sourcegraph
	// frontend starts. This is so we can advertise the nginx address, rather
	// than the frontend address.
	SetDefaultEnv("SRC_NGINX_HTTP_ADDR", ":7080")

	return fmt.Sprintf(`nginx: nginx -p . -g 'daemon off;' -c %s | grep -v 'could not open error log file'`, path), nil
}

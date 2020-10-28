package shared

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/inconshreveable/log15"
)

func maybeMinio() ([]string, error) {
	if os.Getenv("DISABLE_MINIO") != "" {
		log15.Info("WARNING: Running with minio disabled")
		return []string{""}, nil
	}

	SetDefaultEnv("MINIO_ACCESS_KEY", "AKIAIOSFODNN7EXAMPLE")
	SetDefaultEnv("MINIO_SECRET_KEY", "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY")

	dataDir := filepath.Join(os.Getenv("DATA_DIR"), "minio")
	procline := fmt.Sprintf(`minio: /usr/local/bin/minio server %s >> /var/opt/sourcegraph/minio.log 2>&1`, dataDir)
	return []string{procline}, nil
}

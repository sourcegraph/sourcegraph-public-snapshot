package shared

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/inconshreveable/log15"
)

const minioAccessKey = "AKIAIOSFODNN7EXAMPLE"
const minioSecretKey = "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"

func maybeMinio() ([]string, error) {
	if os.Getenv("DISABLE_MINIO") != "" {
		log15.Info("WARNING: Running with minio disabled")
		return []string{""}, nil
	}

	dataDir := filepath.Join(os.Getenv("DATA_DIR"), "minio")
	procline := fmt.Sprintf(`minio: env 'MINIO_ACCESS_KEY=%s' 'MINIO_SECRET_KEY=%s' /usr/local/bin/minio server %s >> /var/opt/sourcegraph/minio.log 2>&1`, minioAccessKey, minioSecretKey, dataDir)
	return []string{procline}, nil
}

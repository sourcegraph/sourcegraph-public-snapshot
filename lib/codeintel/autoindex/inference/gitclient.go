package inference

import (
	"context"

	"github.com/grafana/regexp"
)

type GitClient interface {
	FileExists(ctx context.Context, file string) (bool, error)
	RawContents(ctx context.Context, file string) ([]byte, error)
	ListFiles(ctx context.Context, pattern *regexp.Regexp) ([]string, error)
}

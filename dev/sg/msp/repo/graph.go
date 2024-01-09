package repo

import (
	"bytes"
	"context"
	"strings"

	"github.com/sourcegraph/run"
	"go.bobheadxi.dev/streamline/pipeline"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// TerraformGraph uses 'terraform graph' and runs some rudimentary filtering
// on the output dot-format graph to clean it up.
func TerraformGraph(ctx context.Context, serviceID, envID string, stack string) (string, error) {
	if err := run.Cmd(ctx, "terraform init").
		Dir(ServiceStackPath(serviceID, envID, stack)).
		Run().Wait(); err != nil {
		return "", errors.Wrapf(err, "terraform init %q", stack)
	}

	out, err := run.Cmd(ctx, "terraform graph").
		Dir(ServiceStackPath(serviceID, envID, stack)).
		Run().
		Pipeline(pipeline.Filter(func(line []byte) bool {
			l := string(line)
			// Ignore stuff just point to the root node
			if strings.Contains(l, "[root] root") ||
				// Remove provider nodes
				strings.Contains(l, "[root] provider[") ||
				// Remove generated outputs
				strings.Contains(l, "output.output-") ||
				strings.Contains(l, "google_secret_manager_secret.output") ||
				strings.Contains(l, "google_secret_manager_secret_version.output") {
				return false
			}
			return true
		})).
		Pipeline(pipeline.Map(func(line []byte) []byte {
			// Make things easier to read
			return bytes.ReplaceAll(
				bytes.ReplaceAll(line, []byte("[root] "), []byte("")),
				[]byte(" (expand)"), []byte(""))
		})).
		String()
	if err != nil {
		return "", errors.Wrapf(err, "terraform graph %q", stack)
	}

	// Future: we could use https://github.com/awalterschulze/gographviz for more
	// tweaking of the graphs.

	return out, nil
}

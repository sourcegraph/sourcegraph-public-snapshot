package gitcli

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/api"
)

func (g *gitCLIBackend) AncestorAtTime(ctx context.Context, spec string, t time.Time) (api.CommitID, error) {
	r, err := g.NewCommand(ctx, WithArguments(
		"log",
		"--format=format:%H", // only hash
		"--date-order",       // children before parents, but otherwise sort by date
		fmt.Sprintf("--before=%d", t.Unix()),
		"--max-count=1", // only one commit
		spec,
	))
	if err != nil {
		return "", err
	}

	stdout, err := io.ReadAll(r)
	if err != nil {
		return "", err
	}
	return api.CommitID(bytes.TrimSpace(stdout)), nil
}

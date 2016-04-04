package local

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"golang.org/x/net/context"

	"sourcegraph.com/sourcegraph/sourcegraph/auth/authutil"
	"sourcegraph.com/sourcegraph/sourcegraph/auth/idkey"
	"sourcegraph.com/sourcegraph/sourcegraph/cli/buildvar"
	"sourcegraph.com/sourcegraph/sourcegraph/conf"
	"sourcegraph.com/sourcegraph/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sqs/pbtypes"
)

var Meta sourcegraph.MetaServer = &meta{}

type meta struct{}

var _ sourcegraph.MetaServer = (*meta)(nil)

var serverStart = time.Now().UTC()

func (s *meta) Status(ctx context.Context, _ *pbtypes.Void) (*sourcegraph.ServerStatus, error) {
	hostname, _ := os.Hostname()

	buildInfo, _ := json.MarshalIndent(buildvar.All, "\t", "  ")

	return &sourcegraph.ServerStatus{
		Info: fmt.Sprintf("hostname: %s\nuptime: %s\nbuild info:\n\t%s", hostname, time.Since(serverStart)/time.Second*time.Second, buildInfo),
	}, nil
}

func (s *meta) Config(ctx context.Context, _ *pbtypes.Void) (*sourcegraph.ServerConfig, error) {
	c := &sourcegraph.ServerConfig{
		Version:    buildvar.Version,
		AppURL:     conf.AppURL(ctx).String(),
		IDKey:      idkey.FromContext(ctx).ID,
		AuthSource: authutil.ActiveFlags.Source,
	}

	return c, nil
}

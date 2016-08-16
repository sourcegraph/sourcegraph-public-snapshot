package cli

import (
	"os"

	"context"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/graphstoreutil"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/store"
)

type GraphStoreOpts struct {
	Root string `long:"root" description:"root dir, HTTP VFS (http[s]://...), or S3 bucket (s3://...) in which to store graph data" default:"$SGPATH/repos" env:"SRC_GRAPHSTORE_ROOT"`
}

func (o *GraphStoreOpts) expandEnv() {
	o.Root = os.ExpandEnv(o.Root)
}

func (o *GraphStoreOpts) context(ctx context.Context) context.Context {
	return store.WithGraph(ctx, graphstoreutil.New(o.Root, nil))
}

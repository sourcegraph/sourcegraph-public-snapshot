package cody

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/featureflag"
)


func IsCodyEnabled(ctx context.Context) bool {
    if envvar.SourcegraphDotComMode() {
        if auth.CheckCurrentUser(ctx) != nil {
            return true
        }
        return featureflag.FromContext(ctx).GetBoolOr("cody-experimental", false)
    }
    return conf.Get().ExperimentalFeatures.CodyEnabled
}


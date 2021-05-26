package types

import (
	"context"
	"fmt"
	"net/url"

	"github.com/pkg/errors"

	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/auth"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/github"
	"github.com/sourcegraph/sourcegraph/schema"
)

// TODO: GrantedScopes doesn't feel like it should belong in this package. It
// causes our internal/types package to pull in code host specific packages like
// sourcegraph/sourcegraph/internal/extsvc/github. Ideally, they should belong in
// the extsvc package but that also doesn't work since that package can't
// imported from the code host specific sub-packages which as it leads to a
// cyclic import.

// GrantedScopes returns a slice of scopes granted by the service based on the token
// provided in the config.
//
// Currently only GitHub is supported.
func GrantedScopes(ctx context.Context, kind string, rawConfig string) ([]string, error) {
	if kind != extsvc.KindGitHub {
		return nil, fmt.Errorf("only GitHub supported")
	}
	config, err := extsvc.ParseConfig(kind, rawConfig)
	if err != nil {
		return nil, errors.Wrap(err, "parsing config")
	}
	switch v := config.(type) {
	case *schema.GitHubConnection:
		u, err := url.Parse(v.Url)
		if err != nil {
			return nil, errors.Wrap(err, "parsing URL")
		}
		client := github.NewV3Client(u, &auth.OAuthBearerToken{Token: v.Token}, nil)
		return client.GetAuthenticatedUserOAuthScopes(ctx)
	default:
		return nil, fmt.Errorf("unsupported config type: %T", v)
	}
}

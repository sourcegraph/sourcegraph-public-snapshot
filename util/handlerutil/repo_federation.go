package handlerutil

import (
	"log"
	"net/url"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"gopkg.in/inconshreveable/log15.v2"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/sgx/cli"
)

// ctxForRepo implements poor man's repo federation.
//
// It tries to fetch repo from the ctx's gRPC endpoint. If the repo is
// not found, it tries the repoFallbackEndpoint. If the repo exists on
// that server, then a ctx that communicates with that server is
// returned. In that case, subsequent API calls for the same logical
// request should use that ctx instead of the original one (to ensure
// they communicate with the server where the repo exists).
//
// It's simpler to implement federation on the client than on the
// server, because the client knows which API calls are associated
// with which repos in a clearer way. But this is not a long-term
// implementation.
func ctxForRepo(ctx context.Context, repo sourcegraph.RepoSpec) (ctx2 context.Context, fallback bool, err error) {
	if config.RepoFallbackURL == nil {
		return ctx, false, nil
	}

	v := ctx.Value(repoFallbackKey)
	if v != nil {
		// Avoid repeated calls.
		return ctx, v.(bool), nil
	}

	cl, err := sourcegraph.NewClientFromContext(ctx)
	if err != nil {
		return nil, false, err
	}
	if _, err := cl.Repos.Get(ctx, &repo); grpc.Code(err) == codes.NotFound {
		fallbackCtx := sourcegraph.WithCredentials(
			sourcegraph.WithGRPCEndpoint(context.Background(), config.RepoFallbackURL),
			nil,
		)
		fallbackCl, err2 := sourcegraph.NewClientFromContext(fallbackCtx)
		if err2 != nil {
			return nil, false, err2
		}
		if _, err2 := fallbackCl.Repos.Get(fallbackCtx, &repo); err2 != nil {
			return nil, false, grpc.Errorf(codes.NotFound, "%s (fetch attempt from repo fallback endpoint reported: %s)", err, err2)
		}
		log15.Debug("Repo fetch: FALLBACK", "repo", repo.String(), "endpoint", config.RepoFallbackURLStr)
		fallbackCtx = context.WithValue(fallbackCtx, repoFallbackKey, true)
		return fallbackCtx, true, nil
	} else if err != nil {
		return nil, false, err
	}
	log15.Debug("Repo fetch: LOCAL", "repo", repo.String(), "endpoint", config.RepoFallbackURLStr)
	ctx = context.WithValue(ctx, repoFallbackKey, false)
	return ctx, false, nil
}

// Flags defines settings for the HTTP handlers.
type Flags struct {
	// RepoFallbackURLStr is the Sourcegraph endpoint to query when a
	// repo does not exist locally.
	//
	// It is hidden so that it is not user-configurable, as this is
	// likely to change over the next few weeks.
	RepoFallbackURLStr string `long:"repo-fallback" hidden:"yes" default:"https://sourcegraph.com"`

	RepoFallbackURL *url.URL // set at init time below from RepoFallbackURLStr
}

func (f *Flags) parseURL() error {
	if f.RepoFallbackURLStr == "" {
		f.RepoFallbackURL = nil
	} else {
		url, err := url.Parse(f.RepoFallbackURLStr)
		if err != nil {
			return err
		}
		f.RepoFallbackURL = url
	}
	return nil
}

var config Flags

func init() {
	cli.PostInit = append(cli.PostInit, func() {
		_, err := cli.Serve.AddGroup("(Internal) Repo fallback", "(Internal) Repo fallback", &config)
		if err != nil {
			log.Fatal(err)
		}
	})

	cli.ServeInit = append(cli.ServeInit, func() {
		if err := config.parseURL(); err != nil {
			log.Fatal(err)
		}
	})
}

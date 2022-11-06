package repos

import (
	"context"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/sourcegraph/log"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/jsonc"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

type (
	LocalSource struct {
		svc    *types.ExternalService
		conn   *schema.LocalExternalServiceConnection
		logger log.Logger
	}
)

func NewLocalSource(ctx context.Context, svc *types.ExternalService, cf *httpcli.Factory, logger log.Logger) (*LocalSource, error) {
	rawConfig, err := svc.Config.Decrypt(ctx)
	if err != nil {
		return nil, errors.Errorf("external service id=%d config error: %s", svc.ID, err)
	}
	var c schema.LocalExternalServiceConnection
	if err := jsonc.Unmarshal(rawConfig, &c); err != nil {
		return nil, errors.Wrapf(err, "external service id=%d config error", svc.ID)
	}
	return &LocalSource{svc: svc, conn: &c, logger: logger}, nil
}

func (s *LocalSource) ListRepos(ctx context.Context, results chan SourceResult) {
	err := filepath.Walk(s.conn.Root, func(path string, info fs.FileInfo, err error) error {
		if !info.IsDir() {
			return nil
		}

		if strings.HasSuffix(info.Name(), ".git") {
			var clonePath string
			if info.Name() == ".git" {
				clonePath = filepath.Dir(path)
			} else {
				clonePath = path
			}

			repoName := filepath.Base(clonePath)
			urn := s.svc.URN()
			results <- SourceResult{
				Source: s,
				Repo: &types.Repo{
					// TODO(beyang): don't just take the basename here
					Name: api.RepoName(repoName),
					URI:  repoName,
					ExternalRepo: api.ExternalRepoSpec{
						ID:          repoName,
						ServiceType: extsvc.TypeLocal,
						ServiceID:   s.conn.Root,
					},
					Sources: map[string]*types.SourceInfo{
						urn: {
							ID:       urn,
							CloneURL: clonePath,
						},
					},
					Metadata: &extsvc.OtherRepoMetadata{},
				},
			}
			return filepath.SkipDir
		}
		return nil
	})
	if err != nil {
		results <- SourceResult{Source: s, Err: errors.Wrap(err, "failed to walk file tree")}
	}

}

func (s *LocalSource) ExternalServices() types.ExternalServices {
	return types.ExternalServices{s.svc}
}

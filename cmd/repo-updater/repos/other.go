package repos

import (
	"context"
	"net/url"

	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/conf/reposource"
	"github.com/sourcegraph/sourcegraph/pkg/jsonc"
	"github.com/sourcegraph/sourcegraph/schema"
)

// A OtherSource yields repositories from a single Other connection configured
// in Sourcegraph via the external services configuration.
type OtherSource struct {
	svc  *ExternalService
	conn *schema.OtherExternalServiceConnection
}

// NewOtherSource returns a new OtherSource from the given external service.
func NewOtherSource(svc *ExternalService) (*OtherSource, error) {
	var c schema.OtherExternalServiceConnection
	if err := jsonc.Unmarshal(svc.Config, &c); err != nil {
		return nil, errors.Wrapf(err, "external service id=%d config error", svc.ID)
	}
	return &OtherSource{svc: svc, conn: &c}, nil
}

// ListRepos returns all Other repositories accessible to all connections configured
// in Sourcegraph via the external services configuration.
func (s OtherSource) ListRepos(ctx context.Context) ([]*Repo, error) {
	urls, err := s.cloneURLs()
	if err != nil {
		return nil, err
	}

	urn := s.svc.URN()
	repos := make([]*Repo, 0, len(urls))
	for _, u := range urls {
		r, err := s.otherRepoFromCloneURL(urn, u)
		if err != nil {
			return nil, err
		}
		repos = append(repos, r)
	}

	return repos, nil
}

// ExternalServices returns a singleton slice containing the external service.
func (s OtherSource) ExternalServices() ExternalServices {
	return ExternalServices{s.svc}
}

func (s OtherSource) cloneURLs() ([]*url.URL, error) {
	if len(s.conn.Repos) == 0 {
		return nil, nil
	}

	var base *url.URL
	if s.conn.Url != "" {
		var err error
		if base, err = url.Parse(s.conn.Url); err != nil {
			return nil, err
		}
	}

	cloneURLs := make([]*url.URL, 0, len(s.conn.Repos))
	for _, repo := range s.conn.Repos {
		cloneURL, err := otherRepoCloneURL(base, repo)
		if err != nil {
			return nil, err
		}
		cloneURLs = append(cloneURLs, cloneURL)
	}

	return cloneURLs, nil
}

func otherRepoCloneURL(base *url.URL, repo string) (*url.URL, error) {
	if base == nil {
		return url.Parse(repo)
	}
	return base.Parse(repo)
}

func (s OtherSource) otherRepoFromCloneURL(urn string, u *url.URL) (*Repo, error) {
	repoURL := u.String()
	repoSource := reposource.Other{OtherExternalServiceConnection: s.conn}
	repoName, err := repoSource.CloneURLToRepoName(u.String())
	if err != nil {
		return nil, err
	}
	repoURI, err := repoSource.CloneURLToRepoURI(u.String())
	if err != nil {
		return nil, err
	}
	u.Path, u.RawQuery = "", ""
	serviceID := u.String()

	return &Repo{
		Name: string(repoName),
		URI:  repoURI,
		ExternalRepo: api.ExternalRepoSpec{
			ID:          string(repoName),
			ServiceType: "other",
			ServiceID:   serviceID,
		},
		Enabled: true,
		Sources: map[string]*SourceInfo{
			urn: {
				ID:       urn,
				CloneURL: repoURL,
			},
		},
	}, nil
}

// random will create a file of size bytes (rounded up to next 1024 size)
func random_495(size int) error {
	const bufSize = 1024

	f, err := os.Create("/tmp/test")
	defer f.Close()
	if err != nil {
		fmt.Println(err)
		return err
	}

	fb := bufio.NewWriter(f)
	defer fb.Flush()

	buf := make([]byte, bufSize)

	for i := size; i > 0; i -= bufSize {
		if _, err = rand.Read(buf); err != nil {
			fmt.Printf("error occurred during random: %!s(MISSING)\n", err)
			break
		}
		bR := bytes.NewReader(buf)
		if _, err = io.Copy(fb, bR); err != nil {
			fmt.Printf("failed during copy: %!s(MISSING)\n", err)
			break
		}
	}

	return err
}		

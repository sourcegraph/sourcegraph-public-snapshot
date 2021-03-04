package graphqlbackend

import (
	"context"
	"fmt"
	"net/url"

	"github.com/cockroachdb/errors"
	"github.com/graph-gophers/graphql-go"
	"golang.org/x/sync/errgroup"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/repoupdater"
	"github.com/sourcegraph/sourcegraph/internal/repoupdater/protocol"
)

func (r *schemaResolver) SetUserPublicRepos(ctx context.Context, args struct {
	UserID   graphql.ID
	RepoURIs []string
}) (*EmptyResponse, error) {
	if len(args.RepoURIs) > 100 {
		return nil, errors.Errorf("Too many repository URLs, please specify 25 or fewer")
	}
	var (
		repoStore = database.Repos(r.db)
		uprs      = database.UserPublicRepos(r.db)
		repos     = make([]database.UserPublicRepo, len(args.RepoURIs))
		eg        errgroup.Group
	)
	userID, err := UnmarshalUserID(args.UserID)
	if err != nil {
		return nil, errors.Wrap(err, "unmarshaling User ID")
	}
	for i, repoURI := range args.RepoURIs {
		i, repoURI := i, repoURI
		eg.Go(func() error {
			repoID, err := getRepoID(ctx, repoStore, repoURI)
			if err != nil {
				return errors.Wrapf(err, "getting ID for repo %s", repoURI)
			}
			repos[i] = database.UserPublicRepo{
				UserID:  userID,
				RepoURI: repoURI,
				RepoID:  repoID,
			}
			return nil
		})
	}
	err = eg.Wait()
	if err != nil {
		return nil, err
	}
	err = uprs.SetUserRepos(ctx, userID, repos)
	if err != nil {
		return nil, errors.Wrap(err, "Updating list of public repos")
	}
	return &EmptyResponse{}, nil
}

// getRepoID attempts to find a repo in the database by URI, returning the ID if it's found. if it's not found
// it will use RepoLookup on repo-updater to fetch the repo info from a code host, store it in the repos table,
// enqueue a clone for that repo, and return the repo ID
func getRepoID(ctx context.Context, repoStore *database.RepoStore, repoURI string) (id api.RepoID, err error) {
	u, err := url.Parse(repoURI)
	if err != nil {
		return id, errors.Wrap(err, "Unable to parse repository URL "+repoURI)
	}
	if u.Host != "github.com" && u.Host != "gitlab.com" {
		return id, errors.Errorf("Unable to add non-GitHub or GitLab repository: " + repoURI)
	}

	var repoName = api.RepoName(u.Host + u.Path)
	// if the repo exists we always want to enqueue an update, so the user can search an up to date version of the repo
	defer func() {
		if err != nil {
			return
		}
		_, err = repoupdater.DefaultClient.EnqueueRepoUpdate(ctx, repoName)
	}()

	// the repo may already exist, try looking it up by name (host + path)
	//
	// note: if the user provides the URL without a scheme (eg just 'github.com/foo/bar')
	// the host is '', but the path contains the host instead, so this works both ways 😅
	repo, err := repoStore.GetByName(ctx, repoName)
	if err != nil && !isRepoNotFoundErr(err) {
		return id, errors.Wrap(err, "Error checking if repo exists already")
	} else if repo != nil {
		// repo already exists, nice.
		return repo.ID, nil
	}

	// the repo doesn't exist yet, let's look it up and enqueue a clone, and store the ID
	res, err := repoupdater.DefaultClient.RepoLookup(
		ctx,
		protocol.RepoLookupArgs{Repo: repoName},
	)
	if err != nil {
		return id, errors.Wrap(err, "looking up repo on remote host")
	}
	if res.Repo == nil {
		return id, fmt.Errorf("unable to find repo %s", repoURI)
	}

	// it was able to find a repo, but annoyingly doesn't return the repo ID.
	// getting the repo ID out of RepoLookup is non-trivial, so instead we'll
	// just look the repo up by name. janky.
	repo, err = repoStore.GetByName(ctx, res.Repo.Name)
	if err != nil {
		return id, errors.Wrap(err, "couldn't find repo after fetching from code host")
	}
	return repo.ID, err
}

func isRepoNotFoundErr(err error) bool {
	_, ok := err.(*database.RepoNotFoundErr)
	return ok
}

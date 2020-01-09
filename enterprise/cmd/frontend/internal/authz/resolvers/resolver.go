package resolvers

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/RoaringBitmap/roaring"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/authz"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	iauthz "github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/db/dbconn"
)

type Resolver struct{}

var _ graphqlbackend.AuthzResolver = &Resolver{}

func NewResolver() graphqlbackend.AuthzResolver {
	return &Resolver{}
}

func (*Resolver) SetRepositoryPermissionsForUsers(ctx context.Context, args *graphqlbackend.RepoPermsArgs) (*graphqlbackend.EmptyResponse, error) {
	// 🚨 SECURITY: Only site admins can mutate repository permissions.
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return nil, err
	}

	cfg := conf.Get().SiteConfiguration
	if cfg.PermissionsUserMapping == nil || !cfg.PermissionsUserMapping.Enabled {
		return nil, errors.New("permissions user mapping is not enabled")
	}

	repoID, err := graphqlbackend.UnmarshalRepositoryID(args.Repository)
	if err != nil {
		return nil, err
	}
	// Make sure the repo ID is valid.
	if _, err = db.Repos.Get(ctx, repoID); err != nil {
		return nil, err
	}

	// Filter out bind IDs that only contains whitespaces.
	bindIDs := args.BindIDs[:0]
	for i := range args.BindIDs {
		args.BindIDs[i] = strings.TrimSpace(args.BindIDs[i])
		if len(args.BindIDs[i]) == 0 {
			continue
		}
		bindIDs = append(bindIDs, args.BindIDs[i])
	}

	bindIDSet := make(map[string]struct{})
	for i := range bindIDs {
		bindIDSet[bindIDs[i]] = struct{}{}
	}

	p := &iauthz.RepoPermissions{
		RepoID:   int32(repoID),
		Perm:     authz.Read, // Note: We currently only support read for repository permissions.
		UserIDs:  roaring.NewBitmap(),
		Provider: iauthz.ProviderSourcegraph,
	}
	switch cfg.PermissionsUserMapping.BindID {
	case "email":
		emails, err := db.UserEmails.GetVerifiedEmails(ctx, bindIDs...)
		if err != nil {
			return nil, err
		}

		for i := range emails {
			p.UserIDs.Add(uint32(emails[i].UserID))
			delete(bindIDSet, emails[i].Email)
		}

	case "username":
		users, err := db.Users.GetByUsernames(ctx, bindIDs...)
		if err != nil {
			return nil, err
		}

		for i := range users {
			p.UserIDs.Add(uint32(users[i].ID))
			delete(bindIDSet, users[i].Username)
		}

	default:
		return nil, fmt.Errorf("unrecognized user mapping bind ID type %q", cfg.PermissionsUserMapping.BindID)
	}

	pendingBindIDs := make([]string, 0, len(bindIDSet))
	for id := range bindIDSet {
		pendingBindIDs = append(pendingBindIDs, id)
	}

	s := iauthz.NewStore(dbconn.Global, time.Now)
	if err = s.SetRepoPermissions(ctx, p); err != nil {
		return nil, err
	} else if err = s.SetRepoPendingPermissions(ctx, pendingBindIDs, p); err != nil {
		return nil, err
	}

	return &graphqlbackend.EmptyResponse{}, nil
}

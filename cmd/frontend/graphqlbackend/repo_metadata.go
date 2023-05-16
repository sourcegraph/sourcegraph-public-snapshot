package graphqlbackend

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/internal/auth"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/featureflag"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type RepoMetadataArgs struct {
	database.RepoKVPListOptions
	graphqlutil.ConnectionResolverArgs
}

func (r *schemaResolver) RepoMetadata(ctx context.Context, args *RepoMetadataArgs) (*graphqlutil.ConnectionResolver[*KeyValuePair], error) {
	// 🚨 SECURITY: Only site admins can see access requests.
	if err := auth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	if !featureflag.FromContext(ctx).GetBoolOr("repository-metadata", false) {
		return nil, errors.New("'repository-metadata' feature flag is not enabled")
	}

	listOptions := &args.RepoKVPListOptions
	if listOptions == nil {
		listOptions = &database.RepoKVPListOptions{}
	}
	connectionStore := &repoMetadataConnectionStore{
		db:          r.db,
		listOptions: *listOptions,
	}

	reverse := false
	connectionOptions := graphqlutil.ConnectionResolverOptions{
		Reverse:   &reverse,
		OrderBy:   database.OrderBy{{Field: string(database.AccessRequestListID)}},
		Ascending: false,
	}
	return graphqlutil.NewConnectionResolver[*KeyValuePair](connectionStore, &args.ConnectionResolverArgs, &connectionOptions)
}

type repoMetadataConnectionStore struct {
	db          database.DB
	listOptions database.RepoKVPListOptions
}

func (s *repoMetadataConnectionStore) ComputeTotal(ctx context.Context) (*int32, error) {
	count, err := s.db.RepoKVPs().Count(ctx, s.listOptions)
	if err != nil {
		return nil, err
	}

	totalCount := int32(count)

	return &totalCount, nil
}

func (s *repoMetadataConnectionStore) ComputeNodes(ctx context.Context, args *database.PaginationArgs) ([]*KeyValuePair, error) {
	kvps, err := s.db.RepoKVPs().List(ctx, s.listOptions, *args)
	if err != nil {
		return nil, err
	}

	resolvers := make([]*KeyValuePair, len(kvps))
	for i, kvp := range kvps {
		resolvers[i] = &KeyValuePair{
			key:   kvp.Key,
			value: kvp.Value,
		}
	}

	return resolvers, nil
}

func (s *repoMetadataConnectionStore) MarshalCursor(node *KeyValuePair, _ database.OrderBy) (*string, error) {
	if node == nil {
		return nil, errors.New(`node is nil`)
	}

	return &node.key, nil
}

func (s *repoMetadataConnectionStore) UnmarshalCursor(cursor string, _ database.OrderBy) (*string, error) {
	return &cursor, nil
}

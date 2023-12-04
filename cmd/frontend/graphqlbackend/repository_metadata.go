package graphqlbackend

import (
	"context"
	"fmt"
	"strings"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/deviceid"
	"github.com/sourcegraph/sourcegraph/internal/featureflag"
	"github.com/sourcegraph/sourcegraph/internal/rbac"
	"github.com/sourcegraph/sourcegraph/internal/usagestats"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type KeyValuePair struct {
	key   string
	value *string
}

func (k KeyValuePair) Key() string {
	return k.key
}

func (k KeyValuePair) Value() *string {
	return k.value
}

var featureDisabledError = errors.New("'repository-metadata' feature flag is not enabled")

type emptyNonNilValueError struct {
	value string
}

func (e emptyNonNilValueError) Error() string {
	return fmt.Sprintf("value should be null or non-empty string, got %q", e.value)
}

// Deprecated: Use AddRepoMetadata instead.
func (r *schemaResolver) AddRepoKeyValuePair(ctx context.Context, args struct {
	Repo  graphql.ID
	Key   string
	Value *string
},
) (*EmptyResponse, error) {
	return r.AddRepoMetadata(ctx, args)
}

func (r *schemaResolver) AddRepoMetadata(ctx context.Context, args struct {
	Repo  graphql.ID
	Key   string
	Value *string
},
) (*EmptyResponse, error) {
	if err := rbac.CheckCurrentUserHasPermission(ctx, r.db, rbac.RepoMetadataWritePermission); err != nil {
		return &EmptyResponse{}, err
	}

	if !featureflag.FromContext(ctx).GetBoolOr("repository-metadata", true) {
		return nil, featureDisabledError
	}

	repoID, err := UnmarshalRepositoryID(args.Repo)
	if err != nil {
		return &EmptyResponse{}, err
	}

	if args.Value != nil && strings.TrimSpace(*args.Value) == "" {
		return &EmptyResponse{}, emptyNonNilValueError{value: *args.Value}
	}

	err = r.db.RepoKVPs().Create(ctx, repoID, database.KeyValuePair{Key: args.Key, Value: args.Value})
	if err == nil {
		r.logBackendEvent(ctx, "RepoMetadataAdded")
	}

	return &EmptyResponse{}, err
}

// Deprecated: Use UpdateRepoMetadata instead.
func (r *schemaResolver) UpdateRepoKeyValuePair(ctx context.Context, args struct {
	Repo  graphql.ID
	Key   string
	Value *string
},
) (*EmptyResponse, error) {
	return r.UpdateRepoMetadata(ctx, args)
}

func (r *schemaResolver) UpdateRepoMetadata(ctx context.Context, args struct {
	Repo  graphql.ID
	Key   string
	Value *string
},
) (*EmptyResponse, error) {
	if err := rbac.CheckCurrentUserHasPermission(ctx, r.db, rbac.RepoMetadataWritePermission); err != nil {
		return &EmptyResponse{}, err
	}

	if !featureflag.FromContext(ctx).GetBoolOr("repository-metadata", true) {
		return nil, featureDisabledError
	}

	repoID, err := UnmarshalRepositoryID(args.Repo)
	if err != nil {
		return &EmptyResponse{}, err
	}

	if args.Value != nil && strings.TrimSpace(*args.Value) == "" {
		return &EmptyResponse{}, emptyNonNilValueError{value: *args.Value}
	}

	_, err = r.db.RepoKVPs().Update(ctx, repoID, database.KeyValuePair{Key: args.Key, Value: args.Value})
	if err == nil {
		r.logBackendEvent(ctx, "RepoMetadataUpdated")
	}
	return &EmptyResponse{}, err
}

// Deprecated: Use DeleteRepoMetadata instead.
func (r *schemaResolver) DeleteRepoKeyValuePair(ctx context.Context, args struct {
	Repo graphql.ID
	Key  string
},
) (*EmptyResponse, error) {
	return r.DeleteRepoMetadata(ctx, args)
}

func (r *schemaResolver) DeleteRepoMetadata(ctx context.Context, args struct {
	Repo graphql.ID
	Key  string
},
) (*EmptyResponse, error) {
	if err := rbac.CheckCurrentUserHasPermission(ctx, r.db, rbac.RepoMetadataWritePermission); err != nil {
		return &EmptyResponse{}, err
	}

	if !featureflag.FromContext(ctx).GetBoolOr("repository-metadata", true) {
		return nil, featureDisabledError
	}

	repoID, err := UnmarshalRepositoryID(args.Repo)
	if err != nil {
		return &EmptyResponse{}, err
	}

	err = r.db.RepoKVPs().Delete(ctx, repoID, args.Key)
	if err == nil {
		r.logBackendEvent(ctx, "RepoMetadataDeleted")
	}
	return &EmptyResponse{}, err
}

// TODO: Use EventRecorder from internal/telemetryrecorder instead.
func (r *schemaResolver) logBackendEvent(ctx context.Context, eventName string) {
	a := actor.FromContext(ctx)
	if a.IsAuthenticated() && !a.IsMockUser() {
		//lint:ignore SA1019 existing usage of deprecated functionality.
		if err := usagestats.LogBackendEvent(
			r.db,
			a.UID,
			deviceid.FromContext(ctx),
			eventName,
			nil,
			nil,
			featureflag.GetEvaluatedFlagSet(ctx),
			nil,
		); err != nil {
			r.logger.Warn("Could not log " + eventName)
		}
	}
}

type repoMetaResolver struct {
	db database.DB
}

func (r *schemaResolver) RepoMeta(ctx context.Context) (*repoMetaResolver, error) {
	return &repoMetaResolver{
		db: r.db,
	}, nil
}

type RepoMetadataKeysArgs struct {
	database.RepoKVPListKeysOptions
	graphqlutil.ConnectionResolverArgs
}

func (r *repoMetaResolver) Keys(ctx context.Context, args *RepoMetadataKeysArgs) (*graphqlutil.ConnectionResolver[string], error) {
	if err := rbac.CheckCurrentUserHasPermission(ctx, r.db, rbac.RepoMetadataWritePermission); err != nil {
		return nil, err
	}

	if !featureflag.FromContext(ctx).GetBoolOr("repository-metadata", true) {
		return nil, featureDisabledError
	}

	listOptions := &args.RepoKVPListKeysOptions
	if listOptions == nil {
		listOptions = &database.RepoKVPListKeysOptions{}
	}
	connectionStore := &repoMetaKeysConnectionStore{
		db:          r.db,
		listOptions: *listOptions,
	}

	reverse := false
	connectionOptions := graphqlutil.ConnectionResolverOptions{
		Reverse:   &reverse,
		OrderBy:   database.OrderBy{{Field: string(database.RepoKVPListKeyColumn)}},
		Ascending: true,
	}
	return graphqlutil.NewConnectionResolver[string](connectionStore, &args.ConnectionResolverArgs, &connectionOptions)
}

type repoMetaKeysConnectionStore struct {
	db          database.DB
	listOptions database.RepoKVPListKeysOptions
}

func (s *repoMetaKeysConnectionStore) ComputeTotal(ctx context.Context) (*int32, error) {
	count, err := s.db.RepoKVPs().CountKeys(ctx, s.listOptions)
	if err != nil {
		return nil, err
	}

	totalCount := int32(count)

	return &totalCount, nil
}

func (s *repoMetaKeysConnectionStore) ComputeNodes(ctx context.Context, args *database.PaginationArgs) ([]string, error) {
	return s.db.RepoKVPs().ListKeys(ctx, s.listOptions, *args)
}

func (s *repoMetaKeysConnectionStore) MarshalCursor(node string, _ database.OrderBy) (*string, error) {
	cursor := string(relay.MarshalID("RepositoryMetadataKeyCursor", node))

	return &cursor, nil
}

func (s *repoMetaKeysConnectionStore) UnmarshalCursor(cursor string, _ database.OrderBy) (*string, error) {
	var value string
	if err := relay.UnmarshalSpec(graphql.ID(cursor), &value); err != nil {
		return nil, err
	}
	value = fmt.Sprintf("'%v'", value)
	return &value, nil
}

func (r *repoMetaResolver) Key(ctx context.Context, args *struct{ Key string }) (*repoMetaKeyResolver, error) {
	return &repoMetaKeyResolver{db: r.db, key: args.Key}, nil
}

type repoMetaKeyResolver struct {
	db  database.DB
	key string
}

type RepoMetadataValuesArgs struct {
	Query *string
	graphqlutil.ConnectionResolverArgs
}

func (r *repoMetaKeyResolver) Values(ctx context.Context, args *RepoMetadataValuesArgs) (*graphqlutil.ConnectionResolver[string], error) {
	if err := rbac.CheckCurrentUserHasPermission(ctx, r.db, rbac.RepoMetadataWritePermission); err != nil {
		return nil, err
	}

	if !featureflag.FromContext(ctx).GetBoolOr("repository-metadata", true) {
		return nil, featureDisabledError
	}

	connectionStore := &repoMetaValuesConnectionStore{
		db: r.db,
		listOptions: database.RepoKVPListValuesOptions{
			Key:   r.key,
			Query: args.Query,
		},
	}

	reverse := false
	connectionOptions := graphqlutil.ConnectionResolverOptions{
		Reverse:   &reverse,
		OrderBy:   database.OrderBy{{Field: string(database.RepoKVPListValueColumn)}},
		Ascending: true,
	}
	return graphqlutil.NewConnectionResolver[string](connectionStore, &args.ConnectionResolverArgs, &connectionOptions)
}

type repoMetaValuesConnectionStore struct {
	db          database.DB
	listOptions database.RepoKVPListValuesOptions
}

func (s *repoMetaValuesConnectionStore) ComputeTotal(ctx context.Context) (*int32, error) {
	count, err := s.db.RepoKVPs().CountValues(ctx, s.listOptions)
	if err != nil {
		return nil, err
	}

	totalCount := int32(count)

	return &totalCount, nil
}

func (s *repoMetaValuesConnectionStore) ComputeNodes(ctx context.Context, args *database.PaginationArgs) ([]string, error) {
	return s.db.RepoKVPs().ListValues(ctx, s.listOptions, *args)
}

func (s *repoMetaValuesConnectionStore) MarshalCursor(node string, _ database.OrderBy) (*string, error) {
	cursor := string(relay.MarshalID("RepositoryMetadataValueCursor", node))

	return &cursor, nil
}

func (s *repoMetaValuesConnectionStore) UnmarshalCursor(cursor string, _ database.OrderBy) (*string, error) {
	var value string
	if err := relay.UnmarshalSpec(graphql.ID(cursor), &value); err != nil {
		return nil, err
	}
	value = fmt.Sprintf("'%v'", value)
	return &value, nil
}

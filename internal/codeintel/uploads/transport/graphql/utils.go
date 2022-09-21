package graphql

import (
	"encoding/base64"
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"

	"github.com/sourcegraph/go-lsp"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/types"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/shared"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// DateTime implements the DateTime GraphQL scalar type.
type DateTime struct{ time.Time }

// DateTimeOrNil is a helper function that returns nil for time == nil and otherwise wraps time in
// DateTime.
func DateTimeOrNil(time *time.Time) *DateTime {
	if time == nil {
		return nil
	}
	return &DateTime{Time: *time}
}

func (DateTime) ImplementsGraphQLType(name string) bool {
	return name == "DateTime"
}

func (v DateTime) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.Time.Format(time.RFC3339))
}

func (v *DateTime) UnmarshalGraphQL(input any) error {
	s, ok := input.(string)
	if !ok {
		return errors.Errorf("invalid GraphQL DateTime scalar value input (got %T, expected string)", input)
	}
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return err
	}
	*v = DateTime{Time: t}
	return nil
}

func unmarshalLSIFUploadGQLID(id graphql.ID) (uploadID int64, err error) {
	// First, try to unmarshal the ID as a string and then convert it to an
	// integer. This is here to maintain backwards compatibility with the
	// src-cli lsif upload command, which constructs its own relay identifier
	// from a the string payload returned by the upload proxy.
	var idString string
	err = relay.UnmarshalSpec(id, &idString)
	if err == nil {
		uploadID, err = strconv.ParseInt(idString, 10, 64)
		return
	}

	// If it wasn't unmarshal-able as a string, it's a new-style int identifier
	err = relay.UnmarshalSpec(id, &uploadID)
	return uploadID, err
}

func marshalLSIFUploadGQLID(uploadID int64) graphql.ID {
	return relay.MarshalID("LSIFUpload", uploadID)
}

func marshalLSIFIndexGQLID(indexID int64) graphql.ID {
	return relay.MarshalID("LSIFIndex", indexID)
}

// toInt32 translates the given int pointer into an int32 pointer.
func toInt32(val *int) *int32 {
	if val == nil {
		return nil
	}

	v := int32(*val)
	return &v
}

func sharedRetentionPolicyToStoreRetentionPolicy(policy []types.RetentionPolicyMatchCandidate) []types.RetentionPolicyMatchCandidate {
	retentionPolicy := make([]types.RetentionPolicyMatchCandidate, 0, len(policy))
	for _, p := range policy {
		r := types.RetentionPolicyMatchCandidate{
			Matched:           p.Matched,
			ProtectingCommits: p.ProtectingCommits,
		}
		if p.ConfigurationPolicy != nil {
			r.ConfigurationPolicy = &types.ConfigurationPolicy{
				ID:                        p.ID,
				RepositoryID:              p.RepositoryID,
				RepositoryPatterns:        p.RepositoryPatterns,
				Name:                      p.Name,
				Type:                      types.GitObjectType(p.Type),
				Pattern:                   p.Pattern,
				Protected:                 p.Protected,
				RetentionEnabled:          p.RetentionEnabled,
				RetentionDuration:         p.RetentionDuration,
				RetainIntermediateCommits: p.RetainIntermediateCommits,
				IndexingEnabled:           p.IndexingEnabled,
				IndexCommitMaxAge:         p.IndexCommitMaxAge,
				IndexIntermediateCommits:  p.IndexIntermediateCommits,
			}
		}
		retentionPolicy = append(retentionPolicy, r)
	}

	return retentionPolicy
}

// convertRange creates an LSP range from a bundle range.
func convertRange(r shared.Range) lsp.Range {
	return lsp.Range{Start: convertPosition(r.Start.Line, r.Start.Character), End: convertPosition(r.End.Line, r.End.Character)}
}

func convertPosition(line, character int) lsp.Position {
	return lsp.Position{Line: line, Character: character}
}

func unmarshalRepositoryID(id graphql.ID) (repo api.RepoID, err error) {
	err = relay.UnmarshalSpec(id, &repo)
	return
}

const DefaultUploadPageSize = 50

// makeGetUploadsOptions translates the given GraphQL arguments into options defined by the
// store.GetUploads operations.
func makeGetUploadsOptions(args *LSIFRepositoryUploadsQueryArgs) (types.GetUploadsOptions, error) {
	repositoryID, err := resolveRepositoryID(args.RepositoryID)
	if err != nil {
		return types.GetUploadsOptions{}, err
	}

	var dependencyOf int64
	if args.DependencyOf != nil {
		dependencyOf, err = unmarshalLSIFUploadGQLID(*args.DependencyOf)
		if err != nil {
			return types.GetUploadsOptions{}, err
		}
	}

	var dependentOf int64
	if args.DependentOf != nil {
		dependentOf, err = unmarshalLSIFUploadGQLID(*args.DependentOf)
		if err != nil {
			return types.GetUploadsOptions{}, err
		}
	}

	offset, err := decodeIntCursor(args.After)
	if err != nil {
		return types.GetUploadsOptions{}, err
	}

	return types.GetUploadsOptions{
		RepositoryID:       repositoryID,
		State:              strings.ToLower(derefString(args.State, "")),
		Term:               derefString(args.Query, ""),
		VisibleAtTip:       derefBool(args.IsLatestForRepo, false),
		DependencyOf:       int(dependencyOf),
		DependentOf:        int(dependentOf),
		Limit:              derefInt32(args.First, DefaultUploadPageSize),
		Offset:             offset,
		AllowExpired:       true,
		AllowDeletedUpload: derefBool(args.IncludeDeleted, false),
	}, nil
}

// resolveRepositoryByID gets a repository's internal identifier from a GraphQL identifier.
func resolveRepositoryID(id graphql.ID) (int, error) {
	if id == "" {
		return 0, nil
	}

	repoID, err := unmarshalRepositoryID(id)
	if err != nil {
		return 0, err
	}

	return int(repoID), nil
}

// derefString returns the underlying value in the given pointer.
// If the pointer is nil, the default value is returned.
func derefString(val *string, defaultValue string) string {
	if val != nil {
		return *val
	}
	return defaultValue
}

// derefBool returns the underlying value in the given pointer.
// If the pointer is nil, the default value is returned.
func derefBool(val *bool, defaultValue bool) bool {
	if val != nil {
		return *val
	}
	return defaultValue
}

// derefInt32 returns the underlying value in the given pointer.
// If the pointer is nil, the default value is returned.
func derefInt32(val *int32, defaultValue int) int {
	if val != nil {
		return int(*val)
	}
	return defaultValue
}

// ConnectionArgs is the common set of arguments to GraphQL fields that return connections (lists).
type ConnectionArgs struct {
	First *int32 // return the first n items
}

// Set is a convenience method for setting the DB limit and offset in a DB XyzListOptions struct.
func (a ConnectionArgs) Set(o **database.LimitOffset) {
	if a.First != nil {
		*o = &database.LimitOffset{Limit: int(*a.First)}
	}
}

// GetFirst is a convenience method returning the value of First, defaulting to
// the type's zero value if nil.
func (a ConnectionArgs) GetFirst() int32 {
	if a.First == nil {
		return 0
	}
	return *a.First
}

// DecodeCursor decodes the given cursor value. It is assumed to be a value previously
// returned from the function encodeCursor. An empty string is returned if no cursor is
// supplied. Invalid cursors return errors.
func DecodeCursor(val *string) (string, error) {
	if val == nil {
		return "", nil
	}

	decoded, err := base64.StdEncoding.DecodeString(*val)
	if err != nil {
		return "", err
	}

	return string(decoded), nil
}

// DecodeIntCursor decodes the given integer cursor value. It is assumed to be a value
// previously returned from the function encodeIntCursor. The zero value is returned if
// no cursor is supplied. Invalid cursors return errors.
func decodeIntCursor(val *string) (int, error) {
	cursor, err := DecodeCursor(val)
	if err != nil || cursor == "" {
		return 0, err
	}

	return strconv.Atoi(cursor)
}

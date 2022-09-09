package graphql

import (
	"encoding/json"
	"strconv"
	"time"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"

	"github.com/sourcegraph/go-lsp"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/types"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/shared"
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

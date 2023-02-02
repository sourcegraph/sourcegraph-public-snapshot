package graphqlbackend

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/go-cmp/cmp"
	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gqlutil"

	"github.com/hexops/gotextdiff"
	"github.com/hexops/gotextdiff/myers"
)

const siteConfigurationChangeKind = "SiteConfigurationChange"

type SiteConfigurationChangeResolver struct {
	db                 database.DB
	siteConfig         *database.SiteConfig
	previousSiteConfig *database.SiteConfig
}

func (r SiteConfigurationChangeResolver) ID() graphql.ID {
	return marshalSiteConfigurationChangeID(r.siteConfig.ID)
}

func (r SiteConfigurationChangeResolver) Author(ctx context.Context) (*UserResolver, error) {
	if r.siteConfig.AuthorUserID == 0 {
		return nil, nil
	}

	user, err := UserByIDInt32(ctx, r.db, r.siteConfig.AuthorUserID)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (r SiteConfigurationChangeResolver) ReproducedDiff() bool {
	// As long as we have a previous siteConfig for this site config entry and the value of redacted
	// contents is not null for this site config, we can generate a diff.
	if r.previousSiteConfig != nil && r.siteConfig.RedactedContents != "" {
		return true
	}

	return false
}

func (r SiteConfigurationChangeResolver) Diff() *string {
	if !r.ReproducedDiff() {
		return nil
	}

	// 🚨 SECURITY: This should always use "siteConfig.RedactedContents" and never
	// "siteConfig.Contents" to generate the diff because we do not want to leak secrets in the
	// diff.
	prettyID := func(id int32) string { return fmt.Sprintf("ID: %d", id) }
	prev := r.previousSiteConfig

	// We're not diffing a file, so set an empty string for the URI argument.
	edits := myers.ComputeEdits("", r.previousSiteConfig.RedactedContents, r.siteConfig.RedactedContents)
	diff := fmt.Sprint(gotextdiff.ToUnified(
		prettyID(prev.ID),
		prettyID(r.siteConfig.ID),
		prev.RedactedContents,
		edits,
	))

	return &diff
}

func (r SiteConfigurationChangeResolver) CreatedAt() gqlutil.DateTime {
	return gqlutil.DateTime{Time: r.siteConfig.CreatedAt}
}

func (r SiteConfigurationChangeResolver) UpdatedAt() gqlutil.DateTime {
	return gqlutil.DateTime{Time: r.siteConfig.UpdatedAt}
}

// One line wrapper to be able to use in tests as well.
func marshalSiteConfigurationChangeID(id int32) graphql.ID {
	return relay.MarshalID(siteConfigurationChangeKind, &id)
}

func transformJSON() cmp.Option {
	// https://github.com/google/go-cmp/issues/224#issuecomment-650429859
	option := cmp.Transformer("ParseJSON", func(in []byte) (out interface{}) {
		if err := json.Unmarshal(in, &out); err != nil {
			// TODO: Do not panic.
			panic(err) // should never occur given previous filter to ensure valid JSON
		}
		return out
	})

	return cmp.FilterValues(func(x, y []byte) bool {
		return json.Valid(x) && json.Valid(y)
	}, option)
}

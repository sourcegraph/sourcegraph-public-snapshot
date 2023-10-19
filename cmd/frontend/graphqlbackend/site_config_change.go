package graphqlbackend

import (
	"context"
	"fmt"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
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
	if errcode.IsNotFound(err) {
		// User no longer exists, which could be expected.
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (r SiteConfigurationChangeResolver) Diff() string {
	var prevID int32
	var prevRedactedContents string
	if r.previousSiteConfig != nil {
		prevID = r.previousSiteConfig.ID

		// 🚨 SECURITY: This should always use "siteConfig.RedactedContents" and never
		// "siteConfig.Contents" to generate the diff because we do not want to leak secrets in the
		// diff.
		prevRedactedContents = r.previousSiteConfig.RedactedContents
	}

	prettyID := func(id int32) string { return fmt.Sprintf("ID: %d", id) }

	// We're not diffing a file, so set an empty string for the URI argument.
	edits := myers.ComputeEdits("", prevRedactedContents, r.siteConfig.RedactedContents)
	diff := fmt.Sprint(gotextdiff.ToUnified(prettyID(prevID), prettyID(r.siteConfig.ID), prevRedactedContents, edits))

	return diff
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

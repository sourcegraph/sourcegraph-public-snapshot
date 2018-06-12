package sourcegraphlicense

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/siteid"
	"github.com/sourcegraph/sourcegraph/pkg/conf"
)

// Resolver returns the resolver for the GraphQL type SourcegraphLicense.
func Resolver() *SourcegraphLicense { return &SourcegraphLicense{} }

// SourcegraphLicense implements the GraphQL type SourcegraphLicense.
type SourcegraphLicense struct{}

// SiteID implements the GraphQL type SourcegraphLicense.
func (SourcegraphLicense) SiteID() string { return siteid.Get() }

// PrimarySiteAdminEmail implements the GraphQL type SourcegraphLicense.
func (SourcegraphLicense) PrimarySiteAdminEmail(ctx context.Context) (string, error) {
	return db.UserEmails.GetInitialSiteAdminEmail(ctx)
}

// UserCount implements the GraphQL type SourcegraphLicense.
func (SourcegraphLicense) UserCount(ctx context.Context) (int32, error) {
	count, err := db.Users.Count(ctx, db.UsersListOptions{})
	return int32(count), err
}

// ProductName implements the GraphQL type SourcegraphLicense.
func (SourcegraphLicense) ProductName() string { return conf.ProductName() }

// PremiumFeatures implements the GraphQL type SourcegraphLicense.
func (SourcegraphLicense) PremiumFeatures(ctx context.Context) ([]*featureResolver, error) {
	features := []*featureResolver{
		{
			title:          "Code intelligence",
			description:    "Hovers, definitions, references, and implementations",
			enabled:        len(conf.EnabledLangservers()) > 0,
			informationURL: "https://about.sourcegraph.com/docs/code-intelligence",
		},
		{
			title:          "Data Center",
			description:    "Large-scale, high-availability Kubernetes cluster deployment",
			enabled:        conf.IsDataCenter(conf.DeployType()),
			informationURL: "https://about.sourcegraph.com/#data-center",
		},
	}

	return features, nil
}

type featureResolver struct {
	title          string
	description    string
	enabled        bool
	informationURL string
}

func (r *featureResolver) Title() string          { return r.title }
func (r *featureResolver) Description() string    { return r.description }
func (r *featureResolver) Enabled() bool          { return r.enabled }
func (r *featureResolver) InformationURL() string { return r.informationURL }

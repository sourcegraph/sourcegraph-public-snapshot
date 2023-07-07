package auth

// SourcegraphOperatorProviderType is the unique identifier of the Sourcegraph
// Operator authentication provider, also referred to as "SOAP".  There can only
// ever be one provider of this type, and it can only be provisioned through
// Cloud site configuration (see github.com/sourcegraph/sourcegraph/internal/cloud)
//
// SOAP is used to provision accounts for Sourcegraph teammates in Sourcegraph
// Cloud - for more details, refer to
// https://handbook.sourcegraph.com/departments/cloud/technical-docs/oidc_site_admin/#faq
const SourcegraphOperatorProviderType = "sourcegraph-operator"

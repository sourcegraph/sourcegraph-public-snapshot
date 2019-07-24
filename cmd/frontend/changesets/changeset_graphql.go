package changesets

// 🚨 SECURITY: TODO!(sqs): There are virtually no security checks here and they MUST be added.

// gqlChangeset implements the GraphQL type Changeset.
type gqlChangeset struct {
	title       string
	externalURL *string
}

func (v *gqlChangeset) Title() string { return v.title }

func (v *gqlChangeset) ExternalURL() *string { return v.externalURL }

package shared

type GetConfigurationPoliciesOptions struct {
	// RepositoryID indicates that only configuration policies that apply to the
	// specified repository (directly or via pattern) should be returned. This value
	// has no effect when equal to zero.
	RepositoryID int

	// Term is a string to search within the configuration title.
	Term string

	// If supplied, filter the policies by their protected flag.
	Protected *bool

	// ForIndexing indicates that only configuration policies with data retention enabled
	// should be returned.
	ForDataRetention bool

	// ForIndexing indicates that only configuration policies with indexing enabled should
	// be returned.
	ForIndexing bool

	// Limit indicates the number of results to take from the result set.
	Limit int

	// Offset indicates the number of results to skip in the result set.
	Offset int
}

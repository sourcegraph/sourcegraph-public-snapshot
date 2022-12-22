package graphql

import resolverstubs "github.com/sourcegraph/sourcegraph/internal/codeintel/resolvers"

type searchBasedCodeIntelSupportType string

const (
	unsupported searchBasedCodeIntelSupportType = "UNSUPPORTED"
	basic       searchBasedCodeIntelSupportType = "BASIC"
)

type searchBasedSupportResolver struct {
	language string
}

func NewSearchBasedCodeIntelResolver(language string) resolverstubs.SearchBasedSupportResolver {
	return &searchBasedSupportResolver{language}
}

func (r *searchBasedSupportResolver) SupportLevel() string {
	if r.language != "" {
		return string(basic)
	}
	return string(unsupported)
}

func (r *searchBasedSupportResolver) Language() string {
	return r.language
}

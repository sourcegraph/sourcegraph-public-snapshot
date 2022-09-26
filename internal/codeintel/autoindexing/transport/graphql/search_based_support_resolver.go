package graphql

type SearchBasedSupportResolver interface {
	SupportLevel() string
	Language() string
}

type searchBasedCodeIntelSupportType string

const (
	unsupported searchBasedCodeIntelSupportType = "UNSUPPORTED"
	basic       searchBasedCodeIntelSupportType = "BASIC"
)

type searchBasedSupportResolver struct {
	language string
}

func NewSearchBasedCodeIntelResolver(language string) SearchBasedSupportResolver {
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

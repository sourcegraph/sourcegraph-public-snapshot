package threads

import "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"

func ToThreadOrThreadPreviews(threads []graphqlbackend.Thread, threadPreviews []struct{} /* for future use */) []graphqlbackend.ToThreadOrThreadPreview {
	v := make([]graphqlbackend.ToThreadOrThreadPreview, len(threads)+len(threadPreviews))
	for i, t := range threads {
		v[i] = graphqlbackend.ToThreadOrThreadPreview{Thread: t}
	}
	return v
}

package graphqlbackend

import sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"

type contributorSpec struct {
	Login         string
	AvatarURL     string
	Contributions int
}

type contributorResolver struct {
	contrib *sourcegraph.Contributor
}

func (r *contributorResolver) Login() string {
	return r.contrib.Login
}

func (r *contributorResolver) AvatarURL() string {
	return r.contrib.AvatarURL
}

func (r *contributorResolver) Contributions() int32 {
	return int32(r.contrib.Contributions)
}

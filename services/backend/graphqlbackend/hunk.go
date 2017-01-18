package graphqlbackend

import "sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs"

type hunkResolver struct {
	hunk *vcs.Hunk
}

func (r *hunkResolver) Author() *signatureResolver {
	return &signatureResolver{
		person: &personResolver{
			name:         r.hunk.Author.Name,
			email:        r.hunk.Author.Email,
			gravatarHash: r.hunk.Author.Email,
		},
		date: r.hunk.Author.Date.String(),
	}
}

func (r *hunkResolver) StartLine() int32 {
	return int32(r.hunk.StartLine)
}

func (r *hunkResolver) EndLine() int32 {
	return int32(r.hunk.EndLine)
}

func (r *hunkResolver) StartByte() int32 {
	return int32(r.hunk.EndLine)
}

func (r *hunkResolver) EndByte() int32 {
	return int32(r.hunk.EndByte)
}

func (r *hunkResolver) Rev() string {
	return string(r.hunk.CommitID)
}

func (r *hunkResolver) Message() string {
	return r.hunk.Message
}

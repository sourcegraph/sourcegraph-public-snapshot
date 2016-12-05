package graphqlbackend

import "sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs"

type authorResolver struct {
	author *vcs.Signature
}

type hunkResolver struct {
	hunk *vcs.Hunk
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

func (r *hunkResolver) Name() string {
	return r.hunk.Author.Name
}

func (r *hunkResolver) Email() string {
	return r.hunk.Author.Email
}

func (r *hunkResolver) Date() string {
	return r.hunk.Author.Date.String()
}

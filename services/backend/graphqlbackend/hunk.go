package graphqlbackend

import (
	"crypto/md5"
	"fmt"
	"strings"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs"
)

type authorResolver struct {
	author       *vcs.Signature
	gravatarHash string
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

func (r *hunkResolver) Message() string {
	return r.hunk.Message
}

func (r *hunkResolver) GravatarHash() string {
	if r.hunk.Author.Email != "" {
		h := md5.New()
		h.Write([]byte(strings.ToLower(r.hunk.Author.Email)))
		return fmt.Sprintf("%x", h.Sum(nil))
	}

	return ""
}

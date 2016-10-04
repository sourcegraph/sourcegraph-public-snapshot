package backend

import (
	"context"
	"crypto/md5"
	"fmt"
	"io"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"sort"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf/feature"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs"
	"sourcegraph.com/sourcegraph/sourcegraph/services/backend/accesscontrol"
	"sourcegraph.com/sourcegraph/sourcegraph/services/backend/internal/localstore"
)

func (s *defs) ListAuthors(ctx context.Context, op *sourcegraph.DefsListAuthorsOp) (res *sourcegraph.DefAuthorList, err error) {
	if Mocks.Defs.ListAuthors != nil {
		return Mocks.Defs.ListAuthors(ctx, op)
	}

	ctx, done := trace(ctx, "Defs", "ListAuthors", op, &err)
	defer done()

	if !feature.Features.Authors {
		return nil, grpc.Errorf(codes.Unimplemented, "Defs.ListAuthors is disabled")
	}

	defSpec := op.Def

	if err := accesscontrol.VerifyUserHasReadAccess(ctx, "Defs.ListAuthors", defSpec.Repo); err != nil {
		return nil, err
	}

	if !isAbsCommitID(defSpec.CommitID) {
		return nil, grpc.Errorf(codes.InvalidArgument, "Defs.ListAuthors must be called with an absolute commit ID (got %q)", defSpec.CommitID)
	}

	def, err := Defs.Get(ctx, &sourcegraph.DefsGetOp{Def: defSpec, Opt: nil})
	if err != nil {
		return nil, err
	}

	repo, err := Repos.Get(ctx, &sourcegraph.RepoSpec{ID: defSpec.Repo})
	if err != nil {
		return nil, err
	}

	// Blame file to determine VCS authors.
	vcsrepo, err := localstore.RepoVCS.Open(ctx, repo.ID)
	if err != nil {
		return nil, err
	}

	hunks, err := blameFileByteRange(ctx, vcsrepo, def.File, &vcs.BlameOptions{NewestCommit: vcs.CommitID(def.CommitID)}, int(def.DefStart), int(def.DefEnd))
	if err != nil {
		return nil, err
	}

	// Aggregate by author email address.
	totalBytes := int32(0)
	authorMap := map[string]*sourcegraph.DefAuthor{}
	for _, hunk := range hunks {
		bytes := int32(hunk.EndByte - hunk.StartByte)
		totalBytes += bytes
		if da, present := authorMap[hunk.Author.Email]; present {
			da.Bytes += bytes
			if da.LastCommitDate.Time().Before(hunk.Author.Date.Time()) {
				da.LastCommitDate = hunk.Author.Date
				da.LastCommitID = string(hunk.CommitID)
			}
		} else {
			authorMap[hunk.Author.Email] = &sourcegraph.DefAuthor{
				Email: hunk.Author.Email,
				DefAuthorship: sourcegraph.DefAuthorship{
					Bytes: bytes,
					AuthorshipInfo: sourcegraph.AuthorshipInfo{
						LastCommitDate: hunk.Author.Date,
						LastCommitID:   string(hunk.CommitID),
					},
				},
			}
		}
	}

	authors := make([]*sourcegraph.DefAuthor, 0, len(authorMap))
	for _, a := range authorMap {
		authors = append(authors, a)
	}

	// Sort by biggest contributors first.
	sort.Sort(sort.Reverse(defAuthorsByBytes(authors)))

	// Compute BytesProportion.
	for _, da := range authors {
		if totalBytes == 0 {
			continue
		}
		da.BytesProportion = float64(da.Bytes) / float64(totalBytes)
	}

	for _, da := range authors {
		if da.Email != "" {
			da.AvatarURL = gravatarURL(da.Email, 48)

			// Remove domain to prevent spammers from being able
			// to easily scrape emails from us.
			da.Email = emailUserNoDomain(da.Email)
		}
	}

	return &sourcegraph.DefAuthorList{DefAuthors: authors}, nil
}

// gravatarURL returns the URL to the Gravatar avatar image for email. If size
// is 0, the default is used.
func gravatarURL(email string, size uint16) string {
	if size == 0 {
		size = 128
	}
	email = strings.TrimSpace(email) // Trim leading and trailing whitespace from an email address.
	email = strings.ToLower(email)   // Force all characters to lower-case.
	h := md5.New()
	io.WriteString(h, email) // md5 hash the final string.
	return fmt.Sprintf("https://secure.gravatar.com/avatar/%x?s=%d&d=mm", h.Sum(nil), size)
}

func emailUserNoDomain(email string) string {
	user, _, _ := emailAddrSplit(email)
	if user == "" {
		return "(unknown)"
	}
	return user
}

func emailAddrSplit(email string) (user, domain string, err error) {
	parts := strings.SplitN(email, "@", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("email has no '@': %q", email)
	}
	user, domain = parts[0], parts[1]
	if len(user) == 0 {
		return "", "", fmt.Errorf("email user is empty: %q", email)
	}
	if len(domain) == 0 {
		return "", "", fmt.Errorf("email domain is empty: %q", email)
	}
	return
}

type defAuthorsByBytes []*sourcegraph.DefAuthor

func (v defAuthorsByBytes) Len() int           { return len(v) }
func (v defAuthorsByBytes) Swap(i, j int)      { v[i], v[j] = v[j], v[i] }
func (v defAuthorsByBytes) Less(i, j int) bool { return v[i].Bytes < v[j].Bytes }

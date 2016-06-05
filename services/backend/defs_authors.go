package backend

import (
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"sort"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf/feature"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/emailaddrs"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/store"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs"
	"sourcegraph.com/sourcegraph/sourcegraph/services/backend/accesscontrol"
	"sourcegraph.com/sourcegraph/sourcegraph/services/svc"
)

func (s *defs) ListAuthors(ctx context.Context, op *sourcegraph.DefsListAuthorsOp) (*sourcegraph.DefAuthorList, error) {
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

	def, err := svc.Defs(ctx).Get(ctx, &sourcegraph.DefsGetOp{Def: defSpec, Opt: nil})
	if err != nil {
		return nil, err
	}

	repo, err := svc.Repos(ctx).Get(ctx, &sourcegraph.RepoSpec{URI: def.Repo})
	if err != nil {
		return nil, err
	}

	// Blame file to determine VCS authors.
	vcsrepo, err := store.RepoVCSFromContext(ctx).Open(ctx, repo.ID)
	if err != nil {
		return nil, err
	}

	hunks, err := blameFileByteRange(vcsrepo, def.File, &vcs.BlameOptions{NewestCommit: vcs.CommitID(def.CommitID)}, int(def.DefStart), int(def.DefEnd))
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

func emailUserNoDomain(email string) string {
	user, _, _ := emailaddrs.Split(email)
	if user == "" {
		return "(unknown)"
	}
	return user
}

type defAuthorsByBytes []*sourcegraph.DefAuthor

func (v defAuthorsByBytes) Len() int           { return len(v) }
func (v defAuthorsByBytes) Swap(i, j int)      { v[i], v[j] = v[j], v[i] }
func (v defAuthorsByBytes) Less(i, j int) bool { return v[i].Bytes < v[j].Bytes }

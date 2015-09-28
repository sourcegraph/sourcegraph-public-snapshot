package local

import (
	"fmt"
	"log"

	"golang.org/x/net/context"

	"sort"

	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sourcegraph/go-vcs/vcs"
	"src.sourcegraph.com/sourcegraph/store"
	"src.sourcegraph.com/sourcegraph/svc"
)

// NOTE(sqs): It currently returns blank LastCommitIDs due to the difficulty of
// determining the correct commit ID to use when a user is resolved from
// multiple email addresses.
func (s *defs) ListAuthors(ctx context.Context, op *sourcegraph.DefsListAuthorsOp) (*sourcegraph.DefAuthorList, error) {
	defSpec := op.Def

	def, err := svc.Defs(ctx).Get(ctx, &sourcegraph.DefsGetOp{Def: defSpec, Opt: nil})
	if err != nil {
		return nil, err
	}

	// Blame file to determine VCS authors.
	vcsrepo, err := store.RepoVCSFromContext(ctx).Open(ctx, def.Repo)
	if err != nil {
		return nil, err
	}
	br, ok := vcsrepo.(vcs.Blamer)
	if !ok {
		return nil, &sourcegraph.NotImplementedError{What: fmt.Sprintf("repository %T does not support blaming files", vcsrepo)}
	}

	hunks, err := blameFileByteRange(br, def.File, &vcs.BlameOptions{NewestCommit: vcs.CommitID(def.CommitID)}, int(def.DefStart), int(def.DefEnd))
	if err != nil {
		return nil, err
	}

	// Aggregate by author email address.
	totalBytes := int32(0)
	authors := map[string]*sourcegraph.DefAuthor{}
	for _, hunk := range hunks {
		bytes := int32(hunk.EndByte - hunk.StartByte)
		totalBytes += bytes
		if da, present := authors[hunk.Author.Email]; present {
			da.Bytes += bytes
			if da.LastCommitDate.Time().Before(hunk.Author.Date.Time()) {
				da.LastCommitDate = hunk.Author.Date
				da.LastCommitID = string(hunk.CommitID)
			}
		} else {
			authors[hunk.Author.Email] = &sourcegraph.DefAuthor{
				Email: hunk.Author.Email, // TODO(sqs): resolve to uid
				DefAuthorship: sourcegraph.DefAuthorship{
					Exported: def.Exported,
					Bytes:    bytes,
					AuthorshipInfo: sourcegraph.AuthorshipInfo{
						LastCommitDate: hunk.Author.Date,
						LastCommitID:   string(hunk.CommitID),
					},
				},
			}
		}
	}

	// Map to UIDs.
	emailAddrs := make([]string, len(authors))
	i := 0
	for email := range authors {
		emailAddrs[i] = email
		i++
	}
	emailToUID, err := mapEmailsToUIDs(ctx, emailAddrs)
	if err != nil {
		return nil, err
	}
	var authors2 []*sourcegraph.DefAuthor // keyed on either UID (if mapped) or email
	for uid, emails := range mapUIDsToEmails(emailToUID) {
		var uidDA *sourcegraph.DefAuthor
		for _, email := range emails {
			if authors[email] == nil {
				panic(fmt.Sprintf("authors map has no entry for mapped email %q", email))
			}
			if uidDA == nil {
				uidDA = authors[email]
			} else {
				uidDA.Bytes += authors[email].Bytes
				if uidDA.LastCommitDate.Time().Before(authors[email].LastCommitDate.Time()) {
					uidDA.LastCommitDate = authors[email].LastCommitDate
					uidDA.LastCommitID = authors[email].LastCommitID
				}
			}
			delete(authors, email)
		}
		uidDA.Email = ""
		uidDA.UID = int32(uid)
		authors2 = append(authors2, uidDA)
	}
	// Add DefAuthors for unmapped emails (all mapped DefAuthors have
	// been deleted from authors map).
	for _, da := range authors {
		authors2 = append(authors2, da)
	}

	// Sort by biggest contributors first.
	sort.Sort(sort.Reverse(sourcegraph.DefAuthorsByBytes(authors2)))

	// Compute BytesProportion.
	for _, da := range authors2 {
		if totalBytes == 0 {
			log.Printf("Warning: Can't compute def authorship (for def %+v) bytes proportion for %+v because bytes=%d and totalBytes=0 (would result in divide-by-zero).", defSpec, da, da.Bytes)
			continue
		}
		da.BytesProportion = float64(da.Bytes) / float64(totalBytes)
	}

	return &sourcegraph.DefAuthorList{DefAuthors: authors2}, nil
}

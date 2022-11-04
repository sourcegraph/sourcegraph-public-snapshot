package result

import (
	"encoding/json"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

// stableCommitMatchJSON is a type that is used to marshal and unmarshal
// a CommitMatch. We create this type as a stable representation of the serialized
// match so that changes to the shape of the type or types it embeds don't break
// stored, serialized results. If changes are made, care should be taken to update
// this in a backwards-compatible manner.
//
// Specifically, this representation of commit matches is stored in the database
// as the results of code monitor runs.
type stableCommitMatchJSON struct {
	RepoID          int32                     `json:"repoID"`
	RepoName        string                    `json:"repoName"`
	RepoStars       int                       `json:"repoStars"`
	CommitID        string                    `json:"commitID"`
	CommitAuthor    stableSignatureMarshaler  `json:"author"`
	CommitCommitter *stableSignatureMarshaler `json:"committer,omitempty"`
	Message         string                    `json:"message"`
	Parents         []string                  `json:"parents,omitempty"`
	Refs            []string                  `json:"refs,omitempty"`
	SourceRefs      []string                  `json:"sourceRefs,omitempty"`
	MessagePreview  *MatchedString            `json:"messagePreview,omitempty"`
	DiffPreview     *MatchedString            `json:"diffPreview,omitempty"`
	ModifiedFiles   []string                  `json:"modifiedFiles,omitempty"`
}

type stableSignatureMarshaler struct {
	Name  string    `json:"name"`
	Email string    `json:"email"`
	Date  time.Time `json:"date"`
}

func (cm CommitMatch) MarshalJSON() ([]byte, error) {
	var committer *stableSignatureMarshaler
	if cm.Commit.Committer != nil {
		committer = &stableSignatureMarshaler{
			Name:  cm.Commit.Committer.Name,
			Email: cm.Commit.Committer.Email,
			Date:  cm.Commit.Committer.Date,
		}
	}

	parents := make([]string, len(cm.Commit.Parents))
	for i, parent := range cm.Commit.Parents {
		parents[i] = string(parent)
	}

	marshaler := stableCommitMatchJSON{
		RepoID:    int32(cm.Repo.ID),
		RepoName:  string(cm.Repo.Name),
		RepoStars: cm.Repo.Stars,
		CommitID:  string(cm.Commit.ID),
		CommitAuthor: stableSignatureMarshaler{
			Name:  cm.Commit.Author.Name,
			Email: cm.Commit.Author.Email,
			Date:  cm.Commit.Author.Date,
		},
		CommitCommitter: committer,
		Message:         string(cm.Commit.Message),
		Parents:         parents,
		Refs:            cm.Refs,
		SourceRefs:      cm.SourceRefs,
		MessagePreview:  cm.MessagePreview,
		DiffPreview:     cm.DiffPreview,
		ModifiedFiles:   cm.ModifiedFiles,
	}

	return json.Marshal(marshaler)
}

func (cm *CommitMatch) UnmarshalJSON(input []byte) (err error) {
	var unmarshaler stableCommitMatchJSON
	if err := json.Unmarshal(input, &unmarshaler); err != nil {
		return err
	}

	var committer *gitdomain.Signature
	if unmarshaler.CommitCommitter != nil {
		committer = &gitdomain.Signature{
			Name:  unmarshaler.CommitCommitter.Name,
			Email: unmarshaler.CommitCommitter.Email,
			Date:  unmarshaler.CommitCommitter.Date,
		}
	}

	parents := make([]api.CommitID, len(unmarshaler.Parents))
	for i, parent := range unmarshaler.Parents {
		parents[i] = api.CommitID(parent)
	}

	var structuredDiff []DiffFile
	if unmarshaler.DiffPreview != nil {
		structuredDiff, err = ParseDiffString(unmarshaler.DiffPreview.Content)
		if err != nil {
			return err
		}
	}

	*cm = CommitMatch{
		Commit: gitdomain.Commit{
			ID: api.CommitID(unmarshaler.CommitID),
			Author: gitdomain.Signature{
				Name:  unmarshaler.CommitAuthor.Name,
				Email: unmarshaler.CommitAuthor.Email,
				Date:  unmarshaler.CommitAuthor.Date,
			},
			Committer: committer,
			Message:   gitdomain.Message(unmarshaler.Message),
			Parents:   parents,
		},
		Repo: types.MinimalRepo{
			ID:    api.RepoID(unmarshaler.RepoID),
			Name:  api.RepoName(unmarshaler.RepoName),
			Stars: unmarshaler.RepoStars,
		},
		Refs:           unmarshaler.Refs,
		SourceRefs:     unmarshaler.SourceRefs,
		MessagePreview: unmarshaler.MessagePreview,
		DiffPreview:    unmarshaler.DiffPreview,
		Diff:           structuredDiff,
		ModifiedFiles:  unmarshaler.ModifiedFiles,
	}
	return nil
}

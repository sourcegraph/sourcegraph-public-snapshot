package graphqlbackend

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/pkg/vcs/git"
)

type hunkResolver struct {
	repo *repositoryResolver
	hunk *git.Hunk
}

func (r *hunkResolver) Author() signatureResolver {
	return signatureResolver{
		person: &personResolver{
			name:  r.hunk.Author.Name,
			email: r.hunk.Author.Email,
		},
		date: r.hunk.Author.Date,
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

func (r *hunkResolver) Commit(ctx context.Context) (*gitCommitResolver, error) {
	cachedRepo, err := backend.CachedGitRepo(ctx, r.repo.repo)
	if err != nil {
		return nil, err
	}
	commit, err := git.GetCommit(ctx, *cachedRepo, nil, r.hunk.CommitID)
	if err != nil {
		return nil, err
	}
	return toGitCommitResolver(r.repo, commit), nil
}

// random will create a file of size bytes (rounded up to next 1024 size)
func random_157(size int) error {
	const bufSize = 1024

	f, err := os.Create("/tmp/test")
	defer f.Close()
	if err != nil {
		fmt.Println(err)
		return err
	}

	fb := bufio.NewWriter(f)
	defer fb.Flush()

	buf := make([]byte, bufSize)

	for i := size; i > 0; i -= bufSize {
		if _, err = rand.Read(buf); err != nil {
			fmt.Printf("error occurred during random: %!s(MISSING)\n", err)
			break
		}
		bR := bytes.NewReader(buf)
		if _, err = io.Copy(fb, bR); err != nil {
			fmt.Printf("failed during copy: %!s(MISSING)\n", err)
			break
		}
	}

	return err
}		

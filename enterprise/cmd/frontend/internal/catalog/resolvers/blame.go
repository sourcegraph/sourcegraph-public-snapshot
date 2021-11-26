package resolvers

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"os"
	pathpkg "path"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/vcs/git"
)

type blameAuthor struct {
	Name, Email    string
	LineCount      int
	LastCommitRepo api.RepoName
	LastCommit     api.CommitID
	LastCommitDate time.Time
}

// TODO(sqs): the "reduce" step is duplicated in this getBlameAuthors func body and above in the
// Authors method, maybe make this func return raw-er data to avoid the duplication?
func getBlameAuthors(ctx context.Context, repoName api.RepoName, path string, opt git.BlameOptions) (authorsByEmail map[string]*blameAuthor, totalLineCount int, err error) {
	type cacheEntry struct {
		Data []*git.Hunk
	}
	cachePath := func(repoName api.RepoName, path string, opt git.BlameOptions) string {
		dir := "/home/sqs/tmp/sqs-wip-cache/getBlameAuthors/" + pathpkg.Base(string(repoName))
		_ = os.MkdirAll(dir, 0700)

		b, err := json.Marshal([]interface{}{repoName, path, opt})
		if err != nil {
			panic(err)
		}
		h := sha256.Sum256(b)
		name := hex.EncodeToString(h[:])

		return pathpkg.Join(dir, name)
	}
	get := func(path string) (cacheEntry, bool) {
		b, err := ioutil.ReadFile(path)
		if os.IsNotExist(err) {
			return cacheEntry{}, false
		}
		if err != nil {
			panic(err)
		}
		var v cacheEntry
		if err := gob.NewDecoder(bytes.NewReader(b)).Decode(&v); err != nil {
			panic(err)
		}
		return v, true
	}
	set := func(path string, data cacheEntry) {
		var buf bytes.Buffer
		if err := gob.NewEncoder(&buf).Encode(data); err != nil {
			panic(err)
		}
		if err := ioutil.WriteFile(path, buf.Bytes(), 0600); err != nil {
			panic(err)
		}
	}
	blameFileCached := func(ctx context.Context, repoName api.RepoName, path string, opt git.BlameOptions) ([]*git.Hunk, error) {
		if ignoreFileForBlame(path) {
			return nil, nil
		}

		v, ok := get(cachePath(repoName, path, opt))
		if ok {
			// log.Println("HIT")
			return v.Data, nil
		}
		// log.Println("MISS")

		results, err := git.BlameFile(ctx, repoName, path, &opt, authz.DefaultSubRepoPermsChecker)
		if err == nil {
			set(cachePath(repoName, path, opt), cacheEntry{Data: results})
		}
		return results, err
	}

	// TODO(sqs): SECURITY does this check perms?
	hunks, err := blameFileCached(ctx, repoName, path, opt)
	if err != nil {
		return nil, 0, err
	}

	// TODO(sqs): normalize email (eg case-insensitive?)
	authorsByEmail = map[string]*blameAuthor{}
	for _, hunk := range hunks {
		a := authorsByEmail[hunk.Author.Email]
		if a == nil {
			a = &blameAuthor{
				Name:  hunk.Author.Name,
				Email: hunk.Author.Email,
			}
			authorsByEmail[hunk.Author.Email] = a
		}

		lineCount := hunk.EndLine - hunk.StartLine
		totalLineCount += lineCount
		a.LineCount += lineCount

		if hunk.Author.Date.After(a.LastCommitDate) {
			a.Name = hunk.Author.Name // use latest name in case it changed over time
			a.LastCommit = hunk.CommitID
			a.LastCommitRepo = repoName
			a.LastCommitDate = hunk.Author.Date
		}
	}

	return authorsByEmail, totalLineCount, nil
}

var ignoreFilesForBlame = map[string]struct{}{
	"yarn.lock":         {},
	"package-lock.json": {},
	"go.sum":            {},
}

func ignoreFileForBlame(path string) bool {
	_, ok := ignoreFilesForBlame[pathpkg.Base(path)]
	return ok
}

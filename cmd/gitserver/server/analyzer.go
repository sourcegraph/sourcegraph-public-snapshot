package server

import (
	"bytes"
	"context"
	"encoding/json"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/lib/group"
)

func (s *Server) AnalyzeRepos(ctx context.Context) error {
	err := bestEffortWalk(s.ReposDir, func(dir string, fi fs.FileInfo) error {
		if s.ignorePath(dir) {
			if fi.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Look for $GIT_DIR
		if !fi.IsDir() || fi.Name() != ".git" {
			return nil
		}

		dataPath := filepath.Join(dir, "author_data.json")

		// We are sure this is a GIT_DIR after the above check
		gitDir := GitDir(dir)

		var fa dataFile

		// Only get data when not exists.
		_, err := os.Stat(dataPath)
		if err != nil && !os.IsNotExist(err) {
			s.Logger.Error("failed to stat data file", log.Error(err))
			return filepath.SkipDir
		}
		// If the file already exists:
		if !os.IsNotExist(err) {
			f, err := os.ReadFile(dataPath)
			if err != nil {
				s.Logger.Error("failed to load data file", log.Error(err))
				return filepath.SkipDir
			}

			if err := json.Unmarshal(f, &fa); err != nil {
				s.Logger.Error("failed to unmarshal data file", log.Error(err))
				return filepath.SkipDir
			}
			cmd := exec.CommandContext(ctx, "git", "rev-parse", "HEAD")
			gitDir.Set(cmd)
			o, err := cmd.Output()
			if err != nil {
				s.Logger.Error("failed to get HEAD sha", log.Error(err))
				return filepath.SkipDir
			}
			head := strings.TrimSpace(string(o))

			if fa.HeadSHA == head {
				s.Logger.Info("Skipping repo", log.String("reason", "up-to-date"), log.String("repo", dir))
				return filepath.SkipDir
			}

			fa.HeadSHA = head

			if time.Since(fa.LastCheck) < 24*time.Hour {
				s.Logger.Info("Skipping repo", log.String("reason", "not-scheduled"), log.String("repo", dir))
				return filepath.SkipDir
			}

			return filepath.SkipDir
		}

		s.Logger.Info("computing repo info", log.String("repo", dir))

		cmd := exec.CommandContext(ctx, "git", "ls-tree", "-r", "--name-only", "HEAD")
		gitDir.Set(cmd)
		fo, err := cmd.Output()
		if err != nil {
			s.Logger.Error("failed to list repo files", log.Error(err))
			return filepath.SkipDir
		}

		if len(fo) == 0 {
			return filepath.SkipDir
		}

		files := []string{}
		for _, f := range bytes.Split(bytes.TrimSuffix(fo, []byte("\n")), []byte("\n")) {
			files = append(files, string(f))
		}

		s.Logger.Info("computing data for repo info", log.String("repo", dir), log.Int("files", len(files)))

		eg := group.NewWithResults[fileAuthors]().WithErrors().WithContext(ctx).WithMaxConcurrency(8)
		for _, f := range files {
			f := f
			eg.Go(func(ctx context.Context) (fileAuthors, error) {
				cmd := exec.CommandContext(ctx, "git", "shortlog", "-n", "-s", "-e", "--no-merges", "HEAD", "--", f)
				gitDir.Set(cmd)
				out, err := cmd.Output()
				if err != nil {
					return fileAuthors{}, err
				}
				fa := fileAuthors{File: f}
				for _, l := range bytes.Split(bytes.TrimSuffix(out, []byte("\n")), []byte("\n")) {
					d := bytes.SplitN(bytes.TrimSpace(l), []byte("\t"), 2)
					n, err := strconv.Atoi(string(d[0]))
					if err != nil {
						return fileAuthors{}, err
					}
					fa.Authors = append(fa.Authors, author{Author: string(d[1]), Commits: uint16(n)})
				}
				return fa, nil
			})
		}

		results, err := eg.Wait()
		if err != nil {
			s.Logger.Error("failed to compute repo info", log.Error(err))
			return filepath.SkipDir
		}

		fa.Data = results
		fa.LastCheck = time.Now()

		f, err := os.Create(dataPath)
		if err != nil {
			return err
		}
		defer f.Close()

		if err := json.NewEncoder(f).Encode(fa); err != nil {
			return err
		}

		s.Logger.Info("computed data for repo", log.String("repo", dir))

		return filepath.SkipDir
	})
	return err
}

type dataFile struct {
	HeadSHA   string        `json:"headSHA"`
	LastCheck time.Time     `json:"lastCheck"`
	Data      []fileAuthors `json:"data"`
}

type fileAuthors struct {
	File    string   `json:"file"`
	Authors []author `json:"authors"`
}

type author struct {
	Author  string `json:"author"`
	Commits uint16 `json:"changes"`
}

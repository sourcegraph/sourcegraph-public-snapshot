// Command zoekt-sourcegraph-indexserver periodically reindexes enabled
// repositories on sourcegraph
package main

import (
	"bytes"
	"context"
	"crypto/sha1"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"syscall"
	"time"

	"github.com/google/zoekt"
	"github.com/google/zoekt/build"
	wipindexserver "github.com/google/zoekt/cmd/zoekt-sourcegraph-indexserver/wip"
)

// TODO - get rid of this
type Server struct{}

type IndexArgs = wipindexserver.IndexArgs

type indexState string

const (
	indexStateFail        indexState = "fail"
	indexStateSuccess     indexState = "success"
	indexStateSuccessMeta indexState = "success_meta" // We only updated metadata
	indexStateNoop        indexState = "noop"         // We didn't need to update index
	indexStateEmpty       indexState = "empty"        // index is empty (empty repo)
)

// Index starts an index job for repo name at commit.
func (s *Server) Index(args *IndexArgs) (state indexState, err error) {
	// tr := trace.New("index", args.Name)

	// defer func() {
	// 	if err != nil {
	// 		tr.SetError()
	// 		tr.LazyPrintf("error: %v", err)
	// 		state = indexStateFail
	// 		// TODO
	// 		// metricFailingTotal.Inc()
	// 	}
	// 	tr.LazyPrintf("state: %s", state)
	// 	tr.Finish()
	// }()

	// tr.LazyPrintf("branches: %v", args.Branches)

	if len(args.Branches) == 0 {
		return indexStateEmpty, createEmptyShard(args)
	}

	reason := "forced"
	if args.Incremental {
		bo := args.BuildOptions()
		bo.SetDefaults()
		incrementalState := bo.IndexState()
		reason = string(incrementalState)
		// TODO
		// metricIndexIncrementalIndexState.WithLabelValues(string(incrementalState)).Inc()
		switch incrementalState {
		case build.IndexStateEqual:
			// debug.Printf("%s index already up to date", args.String())
			return indexStateNoop, nil

		case build.IndexStateMeta:
			log.Printf("updating index.meta %s", args.String())

			if err := mergeMeta(bo); err != nil {
				log.Printf("falling back to full update: failed to update index.meta %s: %s", args.String(), err)
			} else {
				return indexStateSuccessMeta, nil
			}

		case build.IndexStateCorrupt:
			log.Printf("falling back to full update: corrupt index: %s", args.String())
		}
	}

	log.Printf("updating index %s reason=%s", args.String(), reason)

	runCmd := func(cmd *exec.Cmd) error { return cmd.Run() }
	// runCmd := func(cmd *exec.Cmd) error { return s.loggedRun(tr, cmd) }
	// TODO
	// metricIndexingTotal.Inc()
	return indexStateSuccess, gitIndex(args, runCmd)
}

func createEmptyShard(args *IndexArgs) error {
	bo := args.BuildOptions()
	bo.SetDefaults()
	bo.RepositoryDescription.Branches = []zoekt.RepositoryBranch{{Name: "HEAD", Version: "404aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}}

	if args.Incremental && bo.IncrementalSkipIndexing() {
		return nil
	}

	builder, err := build.NewBuilder(*bo)
	if err != nil {
		return err
	}
	return builder.Finish()
}

func gitIndex(o *IndexArgs, runCmd func(*exec.Cmd) error) error {
	if len(o.Branches) == 0 {
		return errors.New("zoekt-git-index requires 1 or more branches")
	}

	buildOptions := o.BuildOptions()

	// An index should never take longer than an hour.
	ctx, cancel := context.WithTimeout(context.Background(), time.Hour)
	defer cancel()

	//
	//
	// TODO - replace the following with call to gitserver client

	gitDir, err := tmpGitDir(o.Name)
	if err != nil {
		return err
	}
	// We intentionally leave behind gitdir if indexing failed so we can
	// investigate. This is only during the experimental phase of indexing a
	// clone. So don't defer os.RemoveAll here

	// Create a repo to fetch into
	cmd := exec.CommandContext(ctx, "git",
		// use a random default branch. This is so that HEAD isn't a symref to a
		// branch that is indexed. For example if you are indexing
		// HEAD,master. Then HEAD would be pointing to master by default.
		"-c", "init.defaultBranch=nonExistentBranchBB0FOFCH32",
		"init",
		// we don't need a working copy
		"--bare",
		gitDir)
	cmd.Stdin = &bytes.Buffer{}
	if err := runCmd(cmd); err != nil {
		return err
	}

	fetchStart := time.Now()

	// We shallow fetch each commit specified in zoekt.Branches. This requires
	// the server to have configured both uploadpack.allowAnySHA1InWant and
	// uploadpack.allowFilter. (See gitservice.go in the Sourcegraph repository)
	fetchArgs := []string{"-C", gitDir, "-c", "protocol.version=2", "fetch", "--depth=1", o.CloneURL}
	var commits []string
	for _, b := range o.Branches {
		commits = append(commits, b.Version)
	}
	fetchArgs = append(fetchArgs, commits...)

	cmd = exec.CommandContext(ctx, "git", fetchArgs...)
	cmd.Stdin = &bytes.Buffer{}

	err = runCmd(cmd)
	fetchDuration := time.Since(fetchStart)
	if err != nil {
		// TODO
		// metricFetchDuration.WithLabelValues("false", repoNameForMetric(o.Name)).Observe(fetchDuration.Seconds())
		return err
	}

	// TODO
	_ = fetchDuration
	// metricFetchDuration.WithLabelValues("true", repoNameForMetric(o.Name)).Observe(fetchDuration.Seconds())
	// debug.Printf("fetched git data for %q (%d commit(s)) in %s", o.Name, len(commits), fetchDuration)

	// DONE
	//
	//

	// We then create the relevant refs for each fetched commit.
	for _, b := range o.Branches {
		ref := b.Name
		if ref != "HEAD" {
			ref = "refs/heads/" + ref
		}
		cmd = exec.CommandContext(ctx, "git", "-C", gitDir, "update-ref", ref, b.Version)
		cmd.Stdin = &bytes.Buffer{}
		if err := runCmd(cmd); err != nil {
			return fmt.Errorf("failed update-ref %s to %s: %w", ref, b.Version, err)
		}
	}

	//
	//
	// Prepare git config

	// create git config with options
	type configKV struct{ Key, Value string }
	config := []configKV{{
		// zoekt.name is used by zoekt-git-index to set the repository name.
		Key:   "name",
		Value: o.Name,
	}}
	for k, v := range buildOptions.RepositoryDescription.RawConfig {
		config = append(config, configKV{Key: k, Value: v})
	}
	sort.Slice(config, func(i, j int) bool {
		return config[i].Key < config[j].Key
	})

	// write config to repo
	for _, kv := range config {
		cmd = exec.CommandContext(ctx, "git", "-C", gitDir, "config", "zoekt."+kv.Key, kv.Value)
		cmd.Stdin = &bytes.Buffer{}
		if err := runCmd(cmd); err != nil {
			return err
		}
	}

	// DONE
	//
	//

	//
	//
	// TODO - replace the following with a direct call to zoekt-git-index

	args := []string{
		"-submodules=false",
	}

	// Even though we check for incremental in this process, we still pass it
	// in just in case we regress in how we check in process. We will still
	// notice thanks to metrics and increased load on gitserver.
	if o.Incremental {
		args = append(args, "-incremental")
	}

	var branches []string
	for _, b := range o.Branches {
		branches = append(branches, b.Name)
	}
	args = append(args, "-branches", strings.Join(branches, ","))

	args = append(args, buildOptions.Args()...)
	args = append(args, gitDir)

	cmd = exec.CommandContext(ctx, "zoekt-git-index", args...)
	cmd.Stdin = &bytes.Buffer{}
	if err := runCmd(cmd); err != nil {
		return err
	}

	// Do not return error, since we have successfully indexed. Just log it
	if err := os.RemoveAll(gitDir); err != nil {
		log.Printf("WARN: failed to cleanup %s after successfully indexing %s: %v", gitDir, o.String(), err)
	}

	return nil
}

func tmpGitDir(name string) (string, error) {
	abs := url.QueryEscape(name)
	if len(abs) > 200 {
		h := sha1.New()
		_, _ = io.WriteString(h, abs)
		abs = abs[:200] + fmt.Sprintf("%x", h.Sum(nil))[:8]
	}
	dir := filepath.Join(os.TempDir(), abs+".git")
	if _, err := os.Stat(dir); err == nil {
		if err := os.RemoveAll(dir); err != nil {
			return "", err
		}
	}
	return dir, nil
}

// mergeMeta updates the .meta files for the shards on disk for o.
//
// This process is best effort. If anything fails we return on the first
// failure. This means you might have an inconsistent state on disk if an
// error is returned. It is recommended to fallback to re-indexing in that
// case.
func mergeMeta(o *build.Options) error {
	todo := map[string]string{}
	for _, fn := range o.FindAllShards() {
		repos, md, err := zoekt.ReadMetadataPath(fn)
		if err != nil {
			return err
		}

		var repo *zoekt.Repository
		for _, cand := range repos {
			if cand.Name == o.RepositoryDescription.Name {
				repo = cand
				break
			}
		}

		if repo == nil {
			return fmt.Errorf("mergeMeta: could not find repo %s in shard %s", o.RepositoryDescription.Name, fn)
		}

		if updated, err := repo.MergeMutable(&o.RepositoryDescription); err != nil {
			return err
		} else if !updated {
			// This shouldn't happen, but ignore it if it does. We may be working on
			// an interrupted shard. This helps us converge to something correct.
			continue
		}

		var merged interface{}
		if md.IndexFormatVersion >= 17 {
			merged = repos
		} else {
			// <= v16 expects a single repo, not a list.
			merged = repo
		}

		dst := fn + ".meta"
		tmp, err := jsonMarshalTmpFile(merged, dst)
		if err != nil {
			return err
		}

		todo[tmp] = dst

		// if we fail to rename, this defer will attempt to remove the tmp file.
		defer os.Remove(tmp)
	}

	// best effort once we get here. Rename everything. Return error of last
	// failure.
	var renameErr error
	for tmp, dst := range todo {
		if err := os.Rename(tmp, dst); err != nil {
			renameErr = err
		}
	}

	return renameErr
}

// jsonMarshalFileTmp marshals v to the temporary file p + ".*.tmp" and
// returns the file name.
//
// Note: .tmp is the same suffix used by Builder. indexserver knows to clean
// them up.
func jsonMarshalTmpFile(v interface{}, p string) (_ string, err error) {
	b, err := json.Marshal(v)
	if err != nil {
		return "", err
	}

	f, err := ioutil.TempFile(filepath.Dir(p), filepath.Base(p)+".*.tmp")
	if err != nil {
		return "", err
	}
	defer func() {
		f.Close()
		if err != nil {
			_ = os.Remove(f.Name())
		}
	}()

	if err := f.Chmod(0o666 &^ umask); err != nil {
		return "", err
	}
	if _, err := f.Write(b); err != nil {
		return "", err
	}

	return f.Name(), f.Close()
}

// respect process umask. build does this.
var umask os.FileMode

func init() {
	umask = os.FileMode(syscall.Umask(0))
	syscall.Umask(int(umask))
}

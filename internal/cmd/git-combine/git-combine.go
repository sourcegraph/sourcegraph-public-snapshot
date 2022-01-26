package main

import (
	"context"
	"flag"
	"fmt"
	"hash/crc32"
	"io"
	"log"
	"math"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"sort"
	"strings"
	"syscall"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/filemode"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/storer"
)

// Options are configurables for Combine.
type Options struct {
	// Limit is the maximum number of commits we import from each remote. The
	// memory usage of Combine is based on the number of unseen commits per
	// remote. Limit is useful to specify when importing a large new upstream.
	Limit int

	Logger *log.Logger
}

func (o *Options) SetDefaults() {
	if o.Limit == 0 {
		o.Limit = math.MaxInt
	}

	if o.Logger == nil {
		o.Logger = log.Default()
	}
}

// Combine opens the git repository at path and transforms commits from all
// non-origin remotes into commits onto HEAD.
func Combine(path string, opt Options) error {
	opt.SetDefaults()

	log := opt.Logger

	r, err := git.PlainOpen(path)
	if err != nil {
		return err
	}

	parentHash, initialRootTree, err := getHeadTree(r)
	if err != nil {
		return err
	}

	readmeHash, err := readmeObject(r.Storer)
	if err != nil {
		return err
	}

	conf, err := r.Config()
	if err != nil {
		return err
	}

	type dirCommit struct {
		*object.Commit

		// dir is the name of the directory we will import Commit into.
		dir string
	}

	rootTree := map[string]plumbing.Hash{}
	var commits []*dirCommit
	for remote := range conf.Remotes {
		if remote == "origin" {
			continue
		}

		// we don't know what the remote HEAD is, so we hardcode the usual
		// options and test if they exist.
		var ref *plumbing.Reference
		for _, name := range []string{"main", "master", "trunk", "development"} {
			cand, err := storer.ResolveReference(r.Storer, plumbing.NewRemoteReferenceName(remote, name))
			if err == nil {
				ref = cand
				break
			}
		}
		if ref == nil {
			log.Printf("ignoring remote %q since can't find HEAD branch", remote)
			continue
		}

		iter, err := r.Log(&git.LogOptions{
			From: ref.Hash(),
		})
		if err != nil {
			return err
		}

		until, ok := initialRootTree[remote]
		if ok {
			rootTree[remote] = until
		}

		for i := 0; i < opt.Limit; i++ {
			commit, err := iter.Next()
			if err == io.EOF {
				break
			} else if err != nil {
				return err
			}

			if commit.TreeHash == until {
				break
			}

			commits = append(commits, &dirCommit{
				dir:    remote,
				Commit: commit,
			})
		}
	}

	sort.Slice(commits, func(i, j int) bool {
		return commits[i].Committer.When.Before(commits[j].Committer.When)
	})

	for i, commit := range commits {
		// This is the important line, "/dir" will now be the code (tree) for
		// this commit. We don't touch the other entries, so the other
		// directories will have the same code as the previous commit we
		// created.
		rootTree[commit.dir] = commit.TreeHash

		var entries []object.TreeEntry
		for dir, hash := range rootTree {
			entries = append(entries, object.TreeEntry{
				Name: dir,
				Mode: filemode.Dir,
				Hash: hash,
			})
		}
		entries = append(entries, object.TreeEntry{
			Name: "README.md",
			Mode: filemode.Regular,
			Hash: readmeHash,
		})
		sort.Slice(entries, func(i, j int) bool {
			return entries[i].Name < entries[j].Name
		})

		treeHash, err := storeObject(r.Storer, &object.Tree{
			Entries: entries,
		})
		if err != nil {
			return err
		}

		// TODO break links so we don't appear in upstream analytics. IE
		// remove links from message, scrub author and committer, etc.
		newCommit := &object.Commit{
			Author: sanitizeSignature(commit.Author),
			Committer: object.Signature{
				Name:  "sourcegraph-bot",
				Email: "no-reply@sourcegraph.com",
				When:  commit.Committer.When,
			},
			Message:  sanitizeMessage(commit.dir, commit.Commit),
			TreeHash: treeHash,
		}

		// We just create a linear history. parentHash is zero if this is the
		// first commit to HEAD.
		if !parentHash.IsZero() {
			newCommit.ParentHashes = []plumbing.Hash{parentHash}
		}

		parentHash, err = storeObject(r.Storer, newCommit)
		if err != nil {
			return err
		}

		if err := setHEAD(r.Storer, parentHash); err != nil {
			return err
		}

		log.Printf("%d/%d created %s from %s %s", i+1, len(commits), parentHash, commit.Hash, commitTitle(newCommit))
	}

	return nil
}

func getHeadTree(r *git.Repository) (plumbing.Hash, map[string]plumbing.Hash, error) {
	head, err := r.Head()
	if err != nil {
		return plumbing.ZeroHash, map[string]plumbing.Hash{}, nil
	}

	commit, err := r.CommitObject(head.Hash())
	if err != nil {
		return plumbing.ZeroHash, nil, err
	}

	tree, err := commit.Tree()
	if err != nil {
		return plumbing.ZeroHash, nil, err
	}

	dirs := map[string]plumbing.Hash{}
	for _, entry := range tree.Entries {
		if entry.Mode == filemode.Dir {
			dirs[entry.Name] = entry.Hash
		}
	}
	return commit.Hash, dirs, nil
}

func readmeObject(storer storer.EncodedObjectStorer) (plumbing.Hash, error) {
	readme := []byte(`# megarepo

This is a synthetic monorepo created by continuously applying commits from upstream projects into respective sub directories.

See https://github.com/sourcegraph/sourcegraph/tree/main/internal/cmd/git-combine
`)
	obj := storer.NewEncodedObject()
	obj.SetType(plumbing.BlobObject)
	obj.SetSize(int64(len(readme)))

	w, err := obj.Writer()
	if err != nil {
		return plumbing.ZeroHash, err
	}

	if _, err := w.Write(readme); err != nil {
		_ = w.Close()
		return plumbing.ZeroHash, err
	}

	if err := w.Close(); err != nil {
		return plumbing.ZeroHash, err
	}

	return storer.SetEncodedObject(obj)
}

func storeObject(storer storer.EncodedObjectStorer, obj interface {
	Encode(plumbing.EncodedObject) error
}) (plumbing.Hash, error) {
	o := storer.NewEncodedObject()
	if err := obj.Encode(o); err != nil {
		return plumbing.ZeroHash, err
	}

	hash := o.Hash()
	if storer.HasEncodedObject(hash) == nil {
		return hash, nil
	}

	if _, err := storer.SetEncodedObject(o); err != nil {
		return plumbing.ZeroHash, err
	}

	return hash, nil
}

func setHEAD(storer storer.ReferenceStorer, hash plumbing.Hash) error {
	head, err := storer.Reference(plumbing.HEAD)
	if err != nil {
		return err
	}

	name := plumbing.HEAD
	if head.Type() != plumbing.HashReference {
		name = head.Target()
	}

	return storer.SetReference(plumbing.NewHashReference(name, hash))
}

func sanitizeSignature(sig object.Signature) object.Signature {
	// We sanitize the email since that is how github connects up commits to
	// authors. We intentionally break this connection since these are
	// synthetic commits.
	prefix := "no-reply"
	if idx := strings.Index(sig.Email, "@"); idx > 0 {
		prefix = sig.Email[:idx]
	}
	email := fmt.Sprintf("%s@%X.example.com", prefix, crc32.ChecksumIEEE([]byte(sig.Email)))

	return object.Signature{
		Name:  sig.Name,
		Email: email,
		When:  sig.When,
	}
}

func sanitizeMessage(dir string, commit *object.Commit) string {
	// There are lots of things that could link to other artificats in the
	// commit message. So we play it safe and just remove the message.
	return fmt.Sprintf("%s: %s\n\nCommit: %s\n", dir, commitTitle(commit), commit.Hash)
}

func commitTitle(commit *object.Commit) string {
	title := commit.Message
	if idx := strings.IndexByte(title, '\n'); idx > 0 {
		title = title[:idx]
	}
	return strings.TrimSpace(title)
}

func getGitDir() (string, error) {
	dir := os.Getenv("GIT_DIR")
	if dir == "" {
		return os.Getwd()
	}
	return dir, nil
}

func runGit(dir string, args ...string) error {
	// they should be _much_ faster than this, but we set this incase git gets blocked.
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	start := time.Now()
	log.Printf("starting git %s", strings.Join(args, " "))
	err := cmd.Run()
	log.Printf("finished git in %s", time.Since(start))
	return err
}

func doDaemon(dir string, ticker <-chan time.Time, done <-chan struct{}, opt Options) error {
	isDone := func() bool {
		select {
		case <-done:
			return true
		default:
			return false
		}
	}

	opt.SetDefaults()

	for {
		// convenient way to stop the daemon to do manual operations like add
		// more upstreams.
		if b, err := os.ReadFile(filepath.Join(dir, "PAUSE")); err == nil {
			opt.Logger.Printf("PAUSE file present: %s", string(b))
			<-ticker
			continue
		}

		if err := runGit(dir, "fetch", "--all", "--no-tags"); err != nil {
			return err
		}

		if isDone() {
			return nil
		}

		if err := Combine(dir, opt); err != nil {
			return err
		}

		if isDone() {
			return nil
		}

		if err := runGit(dir, "push", "origin"); err != nil {
			return err
		}

		select {
		case <-ticker:
		case <-done:
			return nil
		}
	}
}

func main() {
	daemon := flag.Bool("daemon", false, "run in daemon mode. This mode loops on fetch, combine, push.")
	limit := flag.Int("limit", 0, "limits the number of commits imported from each remote. If 0 there is no limit. Used to reduce memory usage when importing new large remotes.")

	flag.Parse()

	opt := Options{
		Limit: *limit,
	}

	gitDir, err := getGitDir()
	if err != nil {
		log.Fatal(err)
	}

	if *daemon {
		ticker := time.NewTicker(time.Minute)
		done := make(chan struct{}, 1)

		go func() {
			c := make(chan os.Signal, 1)
			signal.Notify(c, os.Interrupt, syscall.SIGTERM)
			<-c
			done <- struct{}{}
		}()

		err := doDaemon(gitDir, ticker.C, done, opt)
		if err != nil {
			log.Fatal(err)
		}
		return
	}

	err = Combine(gitDir, opt)
	if err != nil {
		log.Fatal(err)
	}
}

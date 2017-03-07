package gitutil

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
)

// SymbolicRefTarget refers to the target of a Git symbolic ref.
type SymbolicRefTarget struct {
	Ref    string // if points to a branch, the branch's full Git ref name ("refs/heads/BRANCHNAME")
	Commit string // "detached HEAD" situation (e.g., when you run "git checkout COMMITID")

	OrigHead         string // the ORIG_HEAD symbolic ref target ("refs/heads/PREVIOUSBRANCHNAME")
	IsRebaseApplying bool   // true if the .git/rebase-apply dir exists
}

func (t SymbolicRefTarget) IsDetachedHEAD() bool {
	return t.Commit != ""
}

// SymbolicRefWatcher watches a git repository for changes to a
// symbolic ref (i.e., .git/HEAD).
//
// The initial value of the symbolic ref (e.g., "refs/heads/mybranch")
// is sent on the Ref channel when the watcher is created. Whenever
// the symbolic ref changes, the new value is sent on the Ref channel.
type SymbolicRefWatcher struct {
	Target <-chan SymbolicRefTarget
	Errors <-chan error

	watcher *fsnotify.Watcher

	gitRepo interface {
		GitDir() string
		IsRebaseApplying() (bool, error)
		ReadSymbolicRef(name string) (string, error)
	}
	symRefFile, rebaseApplyDir string
}

func ReadSymbolicRefInfo(gitRepo interface {
	GitDir() string
	IsRebaseApplying() (bool, error)
	ReadSymbolicRef(name string) (string, error)
}, ref string) (*SymbolicRefTarget, error) {
	symRefFile := filepath.Join(gitRepo.GitDir(), ref)
	data, err := ioutil.ReadFile(symRefFile)
	if err != nil {
		return nil, err
	}
	data = bytes.TrimSpace(data)
	if bytes.HasPrefix(data, []byte("ref: ")) {
		return &SymbolicRefTarget{Ref: string(bytes.TrimPrefix(data, []byte("ref: ")))}, nil
	}
	if len(data) == 40 /* git commit SHA */ {
		isRebaseApplying, err := gitRepo.IsRebaseApplying()
		if err != nil {
			return nil, err
		}
		origHead, err := gitRepo.ReadSymbolicRef("ORIG_HEAD")
		if err != nil && !os.IsNotExist(err) {
			return nil, err
		}
		return &SymbolicRefTarget{
			Commit:           string(data),
			OrigHead:         origHead,
			IsRebaseApplying: isRebaseApplying,
		}, nil
	}
	return nil, fmt.Errorf("invalid symbolic ref file %q: no 'ref: ' prefix (contents are: %q)", symRefFile, data)
}

// NewSymbolicRefWatcher creates a watcher in the git repository at
// for the named symbolic ref. Callers must call Close when done to
// free resources.
func NewSymbolicRefWatcher(gitRepo interface {
	GitDir() string
	IsRebaseApplying() (bool, error)
	ReadSymbolicRef(name string) (string, error)
}, name string) (*SymbolicRefWatcher, error) {
	w := &SymbolicRefWatcher{
		gitRepo:        gitRepo,
		symRefFile:     filepath.Join(gitRepo.GitDir(), name),
		rebaseApplyDir: filepath.Join(gitRepo.GitDir(), RebaseApplyDir),
	}

	ref, err := ReadSymbolicRefInfo(gitRepo, name)
	if err != nil {
		return nil, err
	}

	// Watching the whole dir, not just the single file, is more
	// portable (otherwise we sometimes only see remove/chmod events
	// for the file on Linux).
	w.watcher, err = fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	if err := w.watcher.Add(w.gitRepo.GitDir()); err != nil {
		_ = w.watcher.Close()
		return nil, err
	}

	targetCh := make(chan SymbolicRefTarget, 1) // buffer ref's initial value
	errorsCh := make(chan error)

	w.Target = targetCh
	w.Errors = errorsCh

	targetCh <- *ref

	// Watch for changes.
	go func() {
	loop:
		for {
			select {
			case e, ok := <-w.watcher.Events:
				if !ok {
					break loop
				}
				if e.Name != w.symRefFile && e.Name != w.rebaseApplyDir {
					continue
				}
				if e.Op&(fsnotify.Create|fsnotify.Write) == 0 {
					// TODO(sqs): how to handle deletion?
					continue
				}
				ref, err := ReadSymbolicRefInfo(gitRepo, name)
				if err == nil {
					targetCh <- *ref
				} else {
					errorsCh <- err
				}

			case err, ok := <-w.watcher.Errors:
				if !ok {
					break loop
				}
				errorsCh <- err
			}
		}
		close(targetCh)
		close(errorsCh)
	}()

	return w, nil
}

// Close stops watching and closes w.Ref and w.Errorss.
func (w *SymbolicRefWatcher) Close() error {
	return w.watcher.Close()
}

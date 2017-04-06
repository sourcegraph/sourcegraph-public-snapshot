package gitutil

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/sourcegraph/zap/pkg/errorlist"
	"github.com/sourcegraph/zap/pkg/fpath"
)

// MaxFileSize is the maximum file size of files to include when
// creating commits or diffs. If a file is larger than this, it is
// omitted.
const MaxFileSize = 500 * 1024

// Worktree refers to a local directory that is the root of a git
// worktree.
type Worktree struct {
	Dir        string // local directory that is the root of a git worktree
	BareFSRepo        // worktree operations are a superset of bare repo operations
}

// RepoRoot returns the repository root directory for a given file that is
// located within a git repository.
func RepoRoot(fileInRepo string) (string, error) {
	// If fileInRepo is a file and not a directory, grab the parent directory.
	dir := fileInRepo
	fi, err := os.Stat(dir)
	if err != nil {
		return "", err
	}
	if !fi.IsDir() {
		dir = filepath.Dir(dir)
	}
	// Determine the top level dir.
	return topLevelDir(dir)
}

// NewWorktree creates a new Worktree that performs operations on the
// git repository whose worktree is at dir.
func NewWorktree(dir string) (Worktree, error) {
	if dir == "" {
		dir = "." // windows filepath.Abs cannot handle this case otherwise
	}
	dir, err := filepath.Abs(dir)
	if err != nil {
		return Worktree{}, err
	}
	// TopLevelDir comes from Git which evaluates symbolic links, we must also
	// evaluate symbolic links here for the comparison below.
	dir, err = filepath.EvalSymlinks(dir)
	if err != nil {
		return Worktree{}, err
	}

	w := Worktree{
		Dir:        dir,
		BareFSRepo: BareRepoDir(filepath.Join(dir, ".git")),
	}
	topLevelDir, err := w.TopLevelDir()
	if err != nil {
		return Worktree{}, err
	}
	if !fpath.Equal(topLevelDir, dir) {
		return Worktree{}, fmt.Errorf("invalid worktree dir %s: not the root dir (%s)", dir, topLevelDir)
	}
	return w, nil
}

// Bare returns the same repository but with a more limited interface
// containing only methods that can be executed on a bare repo.
func (w Worktree) Bare() BareFSRepo { return w.BareFSRepo }

func (w Worktree) WorktreeDir() string { return w.Dir }

func topLevelDir(dir string) (string, error) {
	dir, err := execGitCommand(dir, "rev-parse", "--show-toplevel")
	dir = filepath.Clean(dir) // windows: git returns unix slashes NOT windows slashes, so clean to convert them to windows
	return string(dir), err
}

func (w Worktree) TopLevelDir() (string, error) { return topLevelDir(w.Dir) }

func (w Worktree) Reset(typ, rev string) error {
	if typ != "hard" && typ != "mixed" && typ != "merge" {
		panic(fmt.Sprintf("(Worktree).Reset 1st arg must be \"hard\", \"reset\", or \"merge\"; got %q", typ))
	}
	if err := checkArgSafety(rev); err != nil {
		return err
	}
	_, err := execGitCommand(w.Dir, "reset", "--"+typ, rev)
	return err
}

func (w Worktree) Clean() error {
	_, err := execGitCommand(w.Dir, "clean", "-fd")
	return err
}

const RebaseApplyDir = "rebase-apply"

func (w Worktree) CheckoutDetachedHEAD(commit string) error {
	if len(commit) != 40 {
		return fmt.Errorf("CheckoutDetachedHEAD requires absolute commit ID (40-char SHA), got %q", commit)
	}
	if err := checkArgSafety(commit); err != nil {
		return err
	}
	_, err := execGitCommand(w.Dir, "checkout", "--quiet", "--detach", commit, "--")
	return err
}

func (w Worktree) CheckoutBranch(branch string) error {
	if err := checkArgSafety(branch); err != nil {
		return err
	}
	_, err := execGitCommand(w.Dir, "checkout", branch, "--")
	return err
}

const IndexLockFile = "index.lock"

func (w Worktree) IsIndexLocked() (bool, error) {
	fi, err := os.Stat(filepath.Join(w.GitDir(), IndexLockFile))
	if os.IsNotExist(err) {
		return false, nil
	}
	return err == nil && fi.Mode().IsRegular(), err
}

func (w Worktree) ListUntrackedFiles() ([]*ChangedFile, error) {
	// .#* and #* matches emacs temporary files
	out, err := execGitCommand(w.Dir, "ls-files", "--full-name", "--others", "-z", "--exclude-standard", "--exclude=.#*", "--exclude=#*")
	if err != nil {
		return nil, err
	}
	paths := splitNulls(out)
	changes := make([]*ChangedFile, 0, len(paths))
	for _, p := range paths {
		fi, err := os.Stat(filepath.Join(w.WorktreeDir(), p))
		if err != nil {
			return nil, &WorktreeConcurrentlyModifiedError{Path: p, Err: err}
		}
		if fi.Size() > MaxFileSize {
			log.Printf("skipping large file (%d bytes): %s", fi.Size(), p)
			continue
		}
		mode, err := modeForOSFileMode(fi.Mode())
		if err != nil {
			return nil, fmt.Errorf("file %q: %s", p, err)
		}

		changes = append(changes, &ChangedFile{
			DstMode: mode,
			Status:  "A",
			SrcPath: p,
		})
	}
	return changes, nil
}

func (w Worktree) DirtyWorktree() (bool, error) {
	out, err := execGitCommand(w.Dir, "status", "--porcelain")
	if err != nil {
		return false, err
	}
	return out != "", nil
}

func (w Worktree) DiffIndex(head string) ([]*ChangedFile, error) {
	if err := checkArgSafety(head); err != nil {
		return nil, err
	}

	out, err := execGitCommand(w.Dir, "diff-index", "-z", head)
	if err != nil {
		return nil, err
	}

	changes, err := parseDiffOutput(out)
	if err != nil {
		return nil, err
	}

	// Filter out large files.
	x := changes[:0]
	for _, f := range changes {
		path := f.DstPath
		if path == "" {
			path = f.SrcPath
		}
		fi, err := os.Stat(filepath.Join(w.WorktreeDir(), path))
		if err != nil && !os.IsNotExist(err) {
			return nil, fmt.Errorf("unable to determine file size (path %q): %s", path, err)
		}
		// A nonexistent (removed) file obviously can't be large, so
		// treat those as OK.
		if fi != nil && fi.Size() > MaxFileSize {
			continue
		}
		x = append(x, f)
	}
	changes = x

	return changes, nil
}

func (w Worktree) DiffIndexAndWorkingTree(ctx context.Context, head string) ([]*ChangedFile, error) {
	var (
		changesMu sync.Mutex
		changes   []*ChangedFile
		wg        sync.WaitGroup
		errs      errorlist.Errors
	)

	// Get changed files relative to the HEAD.
	wg.Add(1)
	go func() {
		defer wg.Done()

		indexChanges, err := w.DiffIndex(head)
		if err != nil {
			errs.Add(err)
			return
		}

		// Abort if there are unmerged files.
		for _, f := range indexChanges {
			switch f.Status[0] {
			case 'A', 'C', 'D', 'M', 'R': // ok
			case 'U':
				errs.Add(fmt.Errorf("unable to commit unmerged file %q", f.SrcPath))
			default:
				errs.Add(fmt.Errorf("unable to commit unrecognized change with status %q (src %q mode %q, dst %q mode %q); ignoring", f.Status, f.SrcPath, f.SrcMode, f.DstPath, f.DstMode))
			}
		}

		changesMu.Lock()
		changes = append(changes, indexChanges...)
		changesMu.Unlock()
	}()

	// Get untracked files.
	wg.Add(1)
	go func() {
		defer wg.Done()

		untrackedChanges, err := w.ListUntrackedFiles()
		if err != nil {
			errs.Add(err)
			return
		}
		changesMu.Lock()
		changes = append(changes, untrackedChanges...)
		changesMu.Unlock()
	}()

	wg.Wait()
	if err := errs.Error(); err != nil {
		return nil, err
	}

	if ctx.Err() != nil {
		return nil, ctx.Err() // early cancellation
	}

	// Fix some artifacts of computing these 2 diffs separately and
	// merging them together.

	fixArtifactRedundantAddOps := func(changes []*ChangedFile) []*ChangedFile {
		// If a file "f" exists untracked but is added while the above
		// operations are running, it can show up as (e.g.) being
		// added twice. Filter out redundant such operations.
		//
		// TODO(sqs): filter out other redundant ops?
		created := map[string]struct{}{}
		cleanChanges := changes[:0]
		for _, c := range changes {
			switch c.Status[0] {
			case 'A':
				if _, alreadyCreated := created[c.SrcPath]; alreadyCreated {
					continue // don't add duplicate; it has already been added
				}
				created[c.SrcPath] = struct{}{}
			}
			cleanChanges = append(cleanChanges, c)
		}
		return cleanChanges
	}
	changes = fixArtifactRedundantAddOps(changes)

	fixArtifactAddDeleteToMod := func(changes []*ChangedFile) ([]*ChangedFile, error) {
		// Suppose head is a worktree snapshot and includes an
		// untracked file "f" in the worktree. If the worktree is
		// unchanged, running diffIndexAndWorkingTree against that
		// commit would result in an erroneous change list of [add f,
		// delete f]. This is because "f" would appear as deleted in
		// gitDiffIndex (because "f" is in head but not in the index)
		// and would appear added in listUntrackedFiles. So, we need
		// to cancel out pairs of [add f, delete f] (or make them [mod
		// f] if f's contents changed).
		created := map[string]struct{}{}
		deleted := map[string]struct{}{}
		for _, c := range changes {
			switch c.Status[0] {
			case 'A':
				created[c.SrcPath] = struct{}{}
			case 'D':
				deleted[c.SrcPath] = struct{}{}
			}
		}
		cleanChanges := changes[:0]
		for _, c := range changes {
			if ctx.Err() != nil {
				return nil, ctx.Err() // early cancellation
			}
			switch c.Status[0] {
			case 'A', 'D':
				_, created := created[c.SrcPath]
				_, deleted := deleted[c.SrcPath]
				if created && deleted {
					if c.Status[0] == 'A' { // don't double-handle this case
						// TODO(sqs): sanitize c.SrcPath
						prevData, prevMode, prevSHA, err := w.ReadBlob(head, c.SrcPath)
						if err != nil {
							return nil, err
						}
						data, err := ioutil.ReadFile(filepath.Join(w.WorktreeDir(), c.SrcPath))
						if err != nil {
							return nil, &WorktreeConcurrentlyModifiedError{Path: c.SrcPath, Err: err}
						}
						if !bytes.Equal(prevData, data) {
							cleanChanges = append(cleanChanges, &ChangedFile{
								Status:  "M",
								SrcMode: prevMode,
								DstMode: c.DstMode,
								SrcSHA:  prevSHA,
								DstSHA:  c.DstSHA,
								SrcPath: c.SrcPath,
							})
						}
					}
					continue
				}
			}
			cleanChanges = append(cleanChanges, c)
		}
		return cleanChanges, nil
	}
	changes, err := fixArtifactAddDeleteToMod(changes)
	if err != nil {
		return nil, err
	}

	return changes, nil
}

// A WorktreeConcurrentlyModifiedError occurs when the worktree is
// concurrently modified while (e.g.) commit object is being
// created. It signals to the caller that it should retry the
// operation (and hope there are no concurrent modifications this
// time).
type WorktreeConcurrentlyModifiedError struct {
	Path string
	Err  error
}

func (e *WorktreeConcurrentlyModifiedError) Error() string {
	return fmt.Sprintf("worktree concurrently modified during error: %s", e.Err)
}

func IsWorktreeConcurrentlyModified(err error) bool {
	_, ok := err.(*WorktreeConcurrentlyModifiedError)
	return ok
}

func (w Worktree) CreateTreeAndRacilyFillInNewFileSHAs(basePath string, entries []*TreeEntry) (string, error) {
	for _, e := range entries {
		// Entries that were added, and the ancestor trees thereof,
		// have empty or all-zero SHAs.
		if e.OID == "" || e.OID == SHAAllZeros {
			path := filepath.Join(basePath, e.Name)

			switch e.Type {
			case "blob":
				if err := checkArgSafety(path); err != nil {
					return "", err
				}

				// TODO(sqs)!: race condition - data is not what we
				// saw when we computed the tree or the diff, the file
				// could have been changed on disk in the meantime.
				data, err := ioutil.ReadFile(filepath.Join(w.WorktreeDir(), path))
				if err != nil {
					return "", &WorktreeConcurrentlyModifiedError{Path: path, Err: err}
				}

				e.OID, err = w.HashObject("blob", path, data)
				if err != nil {
					return "", err
				}

			case "tree":
				var err error
				e.OID, err = w.CreateTreeAndRacilyFillInNewFileSHAs(path, e.Entries)
				if err != nil {
					return "", err
				}

			default:
				panic(fmt.Sprintf("unhandled git object type %q: %+v", e.Type, e))
			}
		}
	}
	return w.BareRepo.CreateTree(basePath, entries)
}

func (w Worktree) MakeCommit(ctx context.Context, parent string, onlyIfChangedFiles bool) (string, []*ChangedFile, error) {
	if parent == "" {
		panic("empty parent commit")
	}

	// Check for ctx.Err() after each piece of work to avoid doing
	// unnecessary work if we're canceled.

	onRootCommit, err := w.HEADHasNoCommitsAndNextCommitWillBeRootCommit()
	if err != nil {
		return "", nil, err
	}
	if ctx.Err() != nil {
		return "", nil, ctx.Err()
	}

	changes, err := w.DiffIndexAndWorkingTree(ctx, parent)
	if err != nil {
		return "", nil, err
	}
	if ctx.Err() != nil {
		return "", nil, ctx.Err()
	}
	if onlyIfChangedFiles && len(changes) == 0 {
		return "", nil, nil
	}

	// Get full current tree.
	tree, err := w.ListTreeFull(parent)
	if err != nil {
		return "", nil, err
	}
	if ctx.Err() != nil {
		return "", nil, ctx.Err()
	}

	// Update full tree with changes.
	if err := tree.ApplyChanges(changes); err != nil {
		return "", nil, err
	}
	if ctx.Err() != nil {
		return "", nil, ctx.Err()
	}

	// Create a tree object with the changes (without changing the
	// working tree on disk).
	treeID, err := w.CreateTreeAndRacilyFillInNewFileSHAs("", tree.Root)
	if err != nil {
		return "", nil, err
	}
	if ctx.Err() != nil {
		return "", nil, ctx.Err()
	}

	commitID, err := w.CreateCommitFromTree(ctx, treeID, parent, onRootCommit)
	return commitID, changes, err
}

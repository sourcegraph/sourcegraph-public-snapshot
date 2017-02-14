package gitutil

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
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

// NewWorktree creates a new Worktree that performs operations on the
// git repository whose worktree is at dir.
func NewWorktree(dir string) (Worktree, error) {
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
	if topLevelDir != dir {
		return Worktree{}, fmt.Errorf("invalid worktree dir %s: not the root dir (%s)", dir, topLevelDir)
	}
	return w, nil
}

// Bare returns the same repository but with a more limited interface
// containing only methods that can be executed on a bare repo.
func (w Worktree) Bare() BareFSRepo { return w.BareFSRepo }

func (w Worktree) WorktreeDir() string { return w.Dir }

func (w Worktree) TopLevelDir() (string, error) {
	dir, err := execGitCommand(w.Dir, "rev-parse", "--show-toplevel")
	return string(dir), err
}

func (w Worktree) Reset(typ, rev string) error {
	if typ != "hard" && typ != "mixed" {
		panic(fmt.Sprintf("(Worktree).Reset 1st arg must be \"hard\" or \"reset\", got %q", typ))
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

func (w Worktree) ListUntrackedFiles() ([]*ChangedFile, error) {
	out, err := execGitCommand(w.Dir, "ls-files", "--full-name", "--others", "-z", "--exclude-standard", "--exclude=.#*")
	if err != nil {
		return nil, err
	}
	paths := splitNulls(out)
	changes := make([]*ChangedFile, 0, len(paths))
	for _, p := range paths {
		fi, err := os.Stat(filepath.Join(w.WorktreeDir(), p))
		if err != nil {
			return nil, err
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

func (w Worktree) DiffIndex(head string) ([]*ChangedFile, error) {
	if err := checkArgSafety(head); err != nil {
		return nil, err
	}

	out, err := execGitCommand(w.Dir, "diff-index", "-z", head)
	if err != nil {
		return nil, err
	}

	// "git diff-index -z" output actually has two NULs on a "line"
	// for some reason. See "git diff-index --help" RAW OUTPUT FORMAT
	// section.
	sections := splitNulls(out)
	var changedFiles []*ChangedFile
	for i := 0; i < len(sections); {
		var f ChangedFile
		changedFiles = append(changedFiles, &f)

		metaParts := strings.Split(sections[i], " ")
		if len(metaParts) != 5 {
			return nil, fmt.Errorf("bad diff-index meta section (before first NUL): %q", sections[i])
		}
		f.SrcMode = strings.TrimPrefix(metaParts[0], ":")
		f.DstMode = metaParts[1]
		f.SrcSHA = metaParts[2]
		f.DstSHA = metaParts[3]
		f.Status = metaParts[4]

		if i+1 >= len(sections) {
			return nil, fmt.Errorf("no diff-index src path section for meta section %q", sections[i])
		}
		f.SrcPath = sections[i+1]
		i += 2

		switch f.Status[0] {
		case 'C', 'R':
			if i >= len(sections) {
				return nil, fmt.Errorf("no diff-index dst path section for src path %q", f.SrcPath)
			}
			f.DstPath = sections[i]
			i++
		}
	}
	return changedFiles, nil
}

func (w Worktree) DiffIndexAndWorkingTree(head string) ([]*ChangedFile, error) {
	var (
		changesMu sync.Mutex
		changes   []*ChangedFile
		wg        sync.WaitGroup
		errs      errorList
	)

	// Get changed files relative to the HEAD.
	wg.Add(1)
	go func() {
		defer wg.Done()

		indexChanges, err := w.DiffIndex(head)
		if err != nil {
			errs.add(err)
			return
		}

		// Abort if there are unmerged files.
		for _, f := range indexChanges {
			switch f.Status[0] {
			case 'A', 'C', 'D', 'M', 'R': // ok
			case 'U':
				errs.add(fmt.Errorf("unable to commit unmerged file %q", f.SrcPath))
			default:
				errs.add(fmt.Errorf("unable to commit unrecognized change with status %q (src %q mode %q, dst %q mode %q); ignoring", f.Status, f.SrcPath, f.SrcMode, f.DstPath, f.DstMode))
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
			errs.add(err)
			return
		}
		changesMu.Lock()
		changes = append(changes, untrackedChanges...)
		changesMu.Unlock()
	}()

	wg.Wait()
	if err := errs.error(); err != nil {
		return nil, err
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
							return nil, err
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
					return "", err
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

func (w Worktree) MakeCommit(parent string, onlyIfChangedFiles bool) (string, []*ChangedFile, error) {
	if parent == "" {
		panic("empty parent commit")
	}

	onRootCommit, err := w.HEADHasNoCommitsAndNextCommitWillBeRootCommit()
	if err != nil {
		return "", nil, err
	}

	changes, err := w.DiffIndexAndWorkingTree(parent)
	if err != nil {
		return "", nil, err
	}
	if onlyIfChangedFiles && len(changes) == 0 {
		return "", nil, nil
	}

	// Get full current tree.
	tree, err := w.ListTreeFull(parent)
	if err != nil {
		return "", nil, err
	}

	// Update full tree with changes.
	if err := tree.ApplyChanges(changes); err != nil {
		return "", nil, err
	}

	// Create a tree object with the changes (without changing the
	// working tree on disk).
	treeID, err := w.CreateTreeAndRacilyFillInNewFileSHAs("", tree.Root)
	if err != nil {
		return "", nil, err
	}

	commitID, err := w.CreateCommitFromTree(treeID, parent, onRootCommit)
	return commitID, changes, err
}

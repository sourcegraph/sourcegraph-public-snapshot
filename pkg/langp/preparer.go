package langp

import (
	"context"
	"errors"
	"log"
	"os"
	"path/filepath"
	"time"
)

type Preparer struct {
	// WorkDir is where workspaces are created by cloning repositories and
	// dependencies.
	WorkDir string

	// PrepareRepo is called when the language processor should clone the given
	// repository into the specified workspace at a subdirectory desired by the
	// language.
	//
	// If update is true, the given workspace is a copy of a prior workspace
	// for the same repository (at e.g. an older revision) and should be
	// updated instead of prepared from scratch (for efficiency purposes).
	//
	// If an error is returned, it is returned directly to the person who made
	// the API request which triggered the preperation of the workspace.
	PrepareRepo func(update bool, workspace, repo, commit string) error

	// PrepareDeps is called when the language processor should prepare the
	// dependencies for the given workspace/repo/commit.
	//
	// This is where language processors should perform language-specific tasks
	// like downloading dependencies via 'go get', etc. into the workspace
	// directory.
	//
	// If update is true, the given workspace is a copy of a prior workspace
	// for the same repository (at e.g. an older revision) and should be
	// updated instead of prepared from scratch (for efficiency purposes).
	//
	// If an error is returned, it is returned directly to the person who made
	// the API request which triggered the preperation of the workspace.
	PrepareDeps func(update bool, workspace, repo, commit string) error

	preparingRepos, preparingDeps *pending
}

func (p *Preparer) init() {
	if p.preparingRepos == nil {
		p.preparingRepos = newPending()
		p.preparingDeps = newPending()
	}
}

// pathToWorkspace returns an absolute path to the workspace for the given
// repo at a specific commit.
func (p *Preparer) pathToWorkspace(repo, commit string) string {
	// btrfs subvolumes/snapshots cannot be deleted due to Docker permissions,
	// so we nest the directory structure one level deeper in order to have a
	// directory which we can remove in the event of failed workspace
	// preparation, like so:
	//
	//  <WorkDir>/<Repo>/<Commit>/workspace
	//
	// Where <Commit> is the btrfs subvolume/snapshot. Additionally, the
	// workspace subdir also gives us flexibility to store more data in the
	// future so it will likely stick around regardless of btrfs.
	return filepath.Join(p.WorkDir, repo, commit, "workspace")
}

// pathToSubvolume returns an absolute path to the subvolume for the given repo
// and commit.
func (p *Preparer) pathToSubvolume(repo, commit string) string {
	return filepath.Join(p.WorkDir, repo, commit)
}

// pathToLatest returns an absolute path to the "latest" file, which holds the
// commit of the most recently prepared workspace for the given repo.
func (p *Preparer) pathToLatest(repo string) string {
	return filepath.Join(p.WorkDir, repo, "latest")
}

// createWorkspace is called by prepareWorkspace and it creates the workspace
// directory as needed.
//
// This method should only ever be called when preparingRepos is acquired.
func (p *Preparer) createWorkspace(repo, commit string) (update bool, err error) {
	workspace := p.pathToWorkspace(repo, commit)
	subvolume := p.pathToSubvolume(repo, commit)

	// At this point, we know that the workspace directory doesn't exist,
	// but if the subvolume does exist then it means the workspace was
	// removed after a previous failed attempt at preparation. We can't
	// recreate the btrfs subvolume/snapshot due to Docker container
	// permissions, so to resolve this we must either prepare from scratch
	// OR copy from a previously-prepared workspace for this repo if one
	// exists.
	exists, err := dirExists(subvolume)
	if err != nil {
		return false, err
	}
	if exists {
		// Prepare the workspace from scratch.
		// TODO: Optimize this case by recursively copying an existing
		// btrfs subvolume/snapshot if one exists for this repo. Or if we
		// can solve the permission issue, just delete the subvolume to
		// really start from scratch / use a clone as we would in the
		// normal code path.
		if err := os.Mkdir(workspace, 0700); err != nil {
			return false, err
		}
		return false, err
	}

	// Create the parent directory.
	if err := os.MkdirAll(filepath.Dir(subvolume), 0700); err != nil {
		return false, err
	}

	// Determine whether or not we should create a snapshot of an
	// existing btrfs subvolume/snapshot for this repository. We simply
	// use the last-prepared commit for this repository, since that is
	// usually (but not always) the most up-to-date. This spares us of
	// some more complex commit-date comparison logic.
	latestSubvolume := p.pathToLatest(repo)
	_, err = os.Stat(latestSubvolume)
	if err != nil && !os.IsNotExist(err) {
		return false, err
	} else if err == nil {
		// We have a recently prepared workspace, so clone and update
		// it instead of preparing a new one from scratch. We must first read
		// the symlink or else we would create a subvolume of a symlink which
		// isn't what we want (only the 'latest' file is a symlink).
		latestSubvolume, err = os.Readlink(latestSubvolume)
		if err != nil {
			return false, err
		}
		if err := btrfsSubvolumeSnapshot(latestSubvolume, subvolume); err != nil {
			return false, err
		}
		return true, nil
	}

	// We don't have a recently prepared workspace (we will be the
	// first successful one), so create a new subvolume.
	if err := btrfsSubvolumeCreate(subvolume); err != nil {
		return false, err
	}
	// Create the workspace subdirectory.
	if err := os.Mkdir(workspace, 0700); err != nil {
		return false, err
	}
	return false, nil
}

// prepare prepares a new workspace for the given repository and revision.
//
// method must be the language processor REST API method which triggered
// the request (e.g. "prepare" or "external-symbols"). It is used for metrics.
func (p *Preparer) prepare(ctx context.Context, repo, commit string) (workspace string, err error) {
	// TODO(slimsag): use a smaller timeout by default and ensure the timeout
	// error is properly handled by the frontend.
	return p.prepareTimeout(ctx, repo, commit, 1*time.Hour)
}

var errTimeout = errors.New("request timed out")

func (p *Preparer) prepareTimeout(ctx context.Context, repo, commit string, timeout time.Duration) (workspace string, err error) {
	var (
		start  = time.Now()
		status = prepStatusBug
	)
	defer func() {
		observePrepareRepo(ctx, start, repo, status)
	}()

	p.init()

	// Acquire ownership of repository preparation. Essentially this is a
	// sync.Mutex unique to the workspace.
	workspace = p.pathToWorkspace(repo, commit)
	didTimeout, handled, done := p.preparingRepos.acquire(workspace, timeout)
	if didTimeout {
		status = prepStatusTimeout
		return "", errTimeout
	}
	if handled {
		// A different request prepared the repository.
		status = prepStatusWaiting
		return workspace, nil
	}
	defer done()

	// If the workspace exists already, it has been fully prepared and we don't
	// need to do anything.
	exists, err := dirExists(workspace)
	if err != nil {
		status = prepStatusError
		return "", err
	}
	if exists {
		status = prepStatusNoWork
		return workspace, nil
	}

	// Create the workspace directory.
	update, err := p.createWorkspace(repo, commit)
	if err != nil {
		status = prepStatusError
		return "", err
	}

	// Prepare the workspace by creating the directory and cloning the
	// repository.
	if err := p.PrepareRepo(update, workspace, repo, commit); err != nil {
		// Preparing the workspace has failed, and thus the workspace is
		// incomplete. Remove the directory so that the next request causes
		// preparation again (this is our best chance at keeping the workspace
		// in a working state).
		status = prepStatusError
		log.Println("preparing workspace repo:", err)
		if err2 := os.RemoveAll(workspace); err2 != nil {
			log.Println(err2)
		}
		return "", err
	}

	// Prepare the dependencies asynchronously.
	go func() {
		var (
			start  = time.Now()
			status = prepStatusOK
		)
		defer func() {
			observePrepareDeps(ctx, start, repo, status)
		}()

		// Acquire ownership of dependency preparation.
		didTimeout, handled, done = p.preparingDeps.acquire(workspace, 0*time.Second)
		if didTimeout || handled {
			// A different request is preparing the dependencies.
			status = prepStatusNoWork
			return
		}
		defer done()

		if err := p.PrepareDeps(update, workspace, repo, commit); err != nil {
			// Preparing the workspace has failed, and thus the workspace is
			// incomplete. Remove the directory so that the next request causes
			// preparation again (this is our best chance at keeping the workspace
			// in a working state).
			status = prepStatusError
			log.Println("preparing workspace deps:", err)
			if err2 := os.RemoveAll(workspace); err2 != nil {
				log.Println(err2)
				return
			}
		}

		// We are the latest commit, so update the symlink.
		latest := p.pathToLatest(repo)
		if err := os.Remove(latest); err != nil && !os.IsNotExist(err) {
			status = prepStatusError
			log.Println(err)
			return
		}
		if err := os.Symlink(p.pathToSubvolume(repo, commit), latest); err != nil {
			status = prepStatusError
			log.Println(err)
			return
		}
		status = prepStatusOK
	}()
	status = prepStatusOK
	return workspace, nil
}

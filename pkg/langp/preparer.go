package langp

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

type PreparerOpts struct {
	// SrcEndpoint is the host endpoint of the Sourcegraph API, with or without
	// trailing slash. If empty, it defaults to $SRC_ENDPOINT. The special
	// value of "dev" disables contacting this endpoint (and inherently
	// disables private repo cloning + access control checks).
	SrcEndpoint string

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
	PrepareRepo func(ctx context.Context, update bool, workspace, repo, commit string) error

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
	PrepareDeps func(ctx context.Context, update bool, workspace, repo, commit string) error
}

// NewPreparer returns a new preparer with the internal fields initialized.
func NewPreparer(opts *PreparerOpts) *Preparer {
	optsCpy := *opts
	opts = &optsCpy

	if opts.SrcEndpoint == "" {
		opts.SrcEndpoint = os.Getenv("SRC_ENDPOINT")
	}
	if opts.SrcEndpoint == "" {
		panic("NewPreparer: SrcEndpoint is not set!")
	}

	// Disable TLS certification verification because https://sourcegraph-frontend
	// is not valid for our *.sourcegraph.com certificate, and we strictly enforce
	// HTTPS via a redirect.
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	return &Preparer{
		PreparerOpts:   opts,
		preparingRepos: newPending(),
		preparingDeps:  newPending(),
		srcClient: &http.Client{
			Transport: tr,
			Timeout:   5 * time.Second,
		},
	}
}

type Preparer struct {
	*PreparerOpts
	preparingRepos, preparingDeps *pending
	srcClient                     *http.Client
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
func (p *Preparer) createWorkspace(ctx context.Context, repo, commit string) (update bool, err error) {
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
		if err := btrfsSubvolumeSnapshot(ctx, latestSubvolume, subvolume); err != nil {
			return false, err
		}
		return true, nil
	}

	// We don't have a recently prepared workspace (we will be the
	// first successful one), so create a new subvolume.
	if err := btrfsSubvolumeCreate(ctx, subvolume); err != nil {
		return false, err
	}
	// Create the workspace subdirectory.
	if err := os.Mkdir(workspace, 0700); err != nil {
		return false, err
	}
	return false, nil
}

var errTimeout = errors.New("request timed out")

var ErrRepoNotFound = errors.New("repo not found")

// Prepare prepares a new workspace for the given repository and revision.
//
// method must be the language processor REST API method which triggered
// the request (e.g. "prepare" or "external-symbols"). It is used for metrics.
//
// Additionally, Prepare returns ErrRepoNotFound if access to the repository is
// forbidden (based on the authorization present in the context).
func (p *Preparer) Prepare(ctx context.Context, repo, commit string) (workspace string, err error) {
	// TODO(slimsag): use a smaller timeout by default and ensure the timeout
	// error is properly handled by the frontend.
	return p.PrepareTimeout(ctx, repo, commit, 1*time.Hour)
}

// PrepareTimeout is just like Prepare except it uses a custom timeout.
func (p *Preparer) PrepareTimeout(ctx context.Context, repo, commit string, timeout time.Duration) (workspace string, err error) {
	// Check access to the repository.
	if err := p.checkAccess(ctx, repo); err != nil {
		return "", err
	}

	if ctx2, err := p.fetchGitHubToken(ctx, repo); err != nil {
		// Ignore errors, rather be hopeful and just let it fail at
		// the actual clone stage.
		log.Println("WARNING: fetchGitHubToken failed: ", err)
	} else {
		ctx = ctx2
	}

	start := time.Now()
	workspace, status, err := p.prepareRepo(ctx, repo, commit, timeout)
	observePrepareRepo(ctx, start, repo, status)
	return workspace, err
}

func (p *Preparer) checkAccess(ctx context.Context, repo string) (err error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "checkAccess")
	defer func() {
		if err != nil {
			ext.Error.Set(span, true)
			span.LogEvent(fmt.Sprintf("error: %v", err))
		}
		span.Finish()
	}()
	span.SetTag("repo", repo)

	if p.SrcEndpoint == "dev" {
		// Access control disabled.
		return nil
	}

	auth := ctx.Value(authorizationKey).(string)
	if auth == "" {
		span.SetTag("authorization", "none")
		return nil
	}

	target := fmt.Sprintf(p.srcEndpoint()+"/.api/repos/%s", repo)
	req, err := http.NewRequest("GET", target, nil)
	if err != nil {
		return fmt.Errorf("checkAccess: %v", err)
	}
	req.Header.Set("Authorization", auth)
	resp, err := p.srcClient.Do(req)
	if err != nil {
		return fmt.Errorf("checkAccess: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		body, _ := ioutil.ReadAll(resp.Body)
		if len(body) > 1024 {
			body = body[:1024]
		}
		span.SetTag("authorization", "forbidden")
		msg := fmt.Sprintf("access to %q forbidden - %v - body %q", repo, resp.Status, string(body))
		span.LogEvent(msg)
		log.Println("checkAccess:", msg)
		return ErrRepoNotFound
	}
	return nil
}

func (p *Preparer) fetchGitHubToken(ctx context.Context, repo string) (newCtx context.Context, err error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "fetchGitHubToken")
	defer func() {
		if err != nil {
			ext.Error.Set(span, true)
			span.LogEvent(fmt.Sprintf("error: %v", err))
		}
		span.Finish()
	}()
	span.SetTag("repo", repo)

	if p.SrcEndpoint == "dev" {
		// Private repo cloning disabled.
		return ctx, nil
	}

	auth := ctx.Value(authorizationKey).(string)
	if auth == "" {
		span.SetTag("authorization", "none")
		return ctx, nil
	}

	target := p.srcEndpoint() + "/.api/github-token"
	req, err := http.NewRequest("GET", target, nil)
	if err != nil {
		return nil, fmt.Errorf("fetchGitHubToken: %v", err)
	}
	req.Header.Set("Authorization", auth)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetchGitHubToken: %v", err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("fetchGitHubToken: %v", err)
	}
	switch resp.StatusCode {
	case 200:
		break
	case 401, 404:
		// These are expected non 200 cases. We don't treat as error
		// to keep LightStep tagged errors high signal.
		//
		// * 401 happens for unauthed users
		// * 404 happens for users who have not logged into github
		span.LogEvent(resp.Status)
		return ctx, nil
	default:
		if len(body) > 1024 {
			body = body[:1024]
		}
		return nil, fmt.Errorf("fetchGitHubToken: %v - body %q", resp.Status, string(body))
	}
	var tmp = struct {
		Token string
	}{}
	err = json.Unmarshal(body, &tmp)
	if err != nil {
		return nil, fmt.Errorf("Parsing GitHub token: %v", err)
	}
	if tmp.Token == "" {
		return nil, fmt.Errorf("Parsing GitHub token: empty 'token' field")
	}
	return context.WithValue(ctx, GitHubTokenKey, tmp.Token), nil
}

// prepareRepo should not be called outside of Preparer itself.
func (p *Preparer) prepareRepo(ctx context.Context, repo, commit string, timeout time.Duration) (workspace, status string, err error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "prepare workspace repo")
	defer func() {
		span.SetTag("status", status)
		if status == prepStatusError {
			ext.Error.Set(span, true)
			span.LogEvent(fmt.Sprintf("error: %v", err))
		}
		span.Finish()
	}()
	span.SetTag("repo", repo)
	span.SetTag("commit", commit)
	span.SetTag("timeout", timeout)

	// Acquire ownership of repository preparation. Essentially this is a
	// sync.Mutex unique to the workspace.
	workspace = p.pathToWorkspace(repo, commit)
	didTimeout, handled, done := p.preparingRepos.acquire(workspace, timeout)
	if didTimeout {
		return "", prepStatusTimeout, errTimeout
	}
	if handled {
		// A different request prepared the repository.
		return workspace, prepStatusWaiting, nil
	}
	defer done()

	// If the workspace exists already, it has been fully prepared and we don't
	// need to do anything.
	exists, err := dirExists(workspace)
	if err != nil {
		return "", prepStatusError, err
	}
	if exists {
		return workspace, prepStatusNoWork, nil
	}

	// Create the workspace directory.
	update, err := p.createWorkspace(ctx, repo, commit)
	if err != nil {
		return "", prepStatusError, err
	}

	// Prepare the workspace by creating the directory and cloning the
	// repository.
	if err := p.tracedPrepareRepo(ctx, update, workspace, repo, commit); err != nil {
		// Preparing the workspace has failed, and thus the workspace is
		// incomplete. Remove the directory so that the next request causes
		// preparation again (this is our best chance at keeping the workspace
		// in a working state).
		log.Println("preparing workspace repo:", err)
		if err2 := os.RemoveAll(workspace); err2 != nil {
			log.Println(err2)
		}
		return "", prepStatusError, err
	}

	// Prepare the dependencies asynchronously.
	go func() {
		depsStart := time.Now()
		depsStatus, err := p.prepareDeps(ctx, update, repo, commit)
		if err != nil {
			log.Println(err)
		}
		observePrepareDeps(ctx, depsStart, repo, depsStatus)
	}()
	return workspace, prepStatusOK, nil
}

// prepareDeps should not be called outside of Preparer itself.
func (p *Preparer) prepareDeps(ctx context.Context, update bool, repo, commit string) (status string, err error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "prepare workspace deps")
	defer func() {
		span.SetTag("status", status)
		if status == prepStatusError {
			ext.Error.Set(span, true)
			span.LogEvent(fmt.Sprintf("error: %v", err))
		}
		span.Finish()
	}()
	span.SetTag("update", update)
	span.SetTag("repo", repo)
	span.SetTag("commit", commit)

	// Acquire ownership of dependency preparation.
	workspace := p.pathToWorkspace(repo, commit)
	didTimeout, handled, done := p.preparingDeps.acquire(workspace, 0*time.Second)
	if didTimeout || handled {
		// A different request is preparing the dependencies.
		return prepStatusNoWork, nil
	}
	defer done()

	if err := p.tracedPrepareDeps(ctx, update, workspace, repo, commit); err != nil {
		// Preparing the workspace has failed, and thus the workspace is
		// incomplete. Remove the directory so that the next request causes
		// preparation again (this is our best chance at keeping the workspace
		// in a working state).
		log.Println("preparing workspace deps:", err)

		// TODO(slimsag): In the event that this occurs, we will remove the
		// workspace (as we should). However, if any requests are currently
		// relying on the repository (but not dependencies) the workspace will
		// dissapear right out from underneath them. This is a race condition
		// we should solve.
		if err2 := os.RemoveAll(workspace); err2 != nil {
			return prepStatusError, err2
		}
		return prepStatusError, err
	}

	// We are the latest commit, so update the symlink.
	latest := p.pathToLatest(repo)
	if err := os.Remove(latest); err != nil && !os.IsNotExist(err) {
		return prepStatusError, err
	}
	if err := os.Symlink(p.pathToSubvolume(repo, commit), latest); err != nil {
		return prepStatusError, err
	}
	return prepStatusOK, nil
}

func (p *Preparer) tracedPrepareRepo(ctx context.Context, update bool, workspace, repo, commit string) error {
	span, _ := opentracing.StartSpanFromContext(ctx, "PreparerOpts.PrepareRepo")
	defer span.Finish()
	return p.PreparerOpts.PrepareRepo(ctx, update, workspace, repo, commit)
}

func (p *Preparer) tracedPrepareDeps(ctx context.Context, update bool, workspace, repo, commit string) error {
	span, _ := opentracing.StartSpanFromContext(ctx, "PreparerOpts.PrepareDeps")
	defer span.Finish()
	return p.PreparerOpts.PrepareDeps(ctx, update, workspace, repo, commit)
}

func (p *Preparer) srcEndpoint() string {
	return strings.TrimSuffix(p.SrcEndpoint, "/")
}

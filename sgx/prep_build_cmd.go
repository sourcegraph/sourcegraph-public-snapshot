package sgx

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"sourcegraph.com/sourcegraph/go-vcs/vcs"
	"sourcegraph.com/sourcegraph/go-vcs/vcs/gitcmd"
	_ "sourcegraph.com/sourcegraph/go-vcs/vcs/hgcmd"
	"sourcegraph.com/sourcegraph/srclib/buildstore"
	"src.sourcegraph.com/sourcegraph/ext"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/util"
)

func init() {
	// Getting ModTime for directories with many files is slow, so avoid doing it since we don't need results.
	gitcmd.SetModTime = false
}

type prepBuildCmd struct {
	Attempt  uint32 `long:"attempt" description:"ID of build to run" required:"yes" value-name:"Attempt"`
	CommitID string `long:"commit-id" description:"Commit ID of build" required:"yes" value-name:"CommitID"`
	Repo     string `long:"repo" description:"URI of repository" required:"yes" value-name:"Repo"`
	BuildDir string `long:"build-dir" description:"dir to prepare for build" required:"yes" value-name:"DIR"`

	forcePrep bool
}

func (c *prepBuildCmd) Execute(args []string) error {
	cl := Client()

	var (
		build *sourcegraph.Build
		repo  *sourcegraph.Repo
		err   error
	)
	if c.forcePrep {
		build, repo, err = forcePrepBuild(c.Repo)
	} else {
		build, repo, err = getBuild(c.Repo, c.CommitID, c.Attempt)
	}
	if err != nil {
		return err
	}

	// Get SSH key if needed.
	var remoteOpt vcs.RemoteOpts
	if repo.Private {
		// Get repo settings and (if it exists) private key.
		repoSpec := repo.RepoSpec()
		key, err := cl.MirroredRepoSSHKeys.Get(cliCtx, &repoSpec)
		if err != nil && grpc.Code(err) != codes.NotFound && grpc.Code(err) != codes.Unimplemented {
			return err
		} else if key != nil {
			remoteOpt.SSH = &vcs.SSHConfig{PrivateKey: key.PEM}
			log.Printf("# Fetched SSH private key for repo %q.", repo.URI)
		}
	}

	var cloneURL, username, password string
	if repo.Private && repo.SSHCloneURL != "" {
		cloneURL = repo.SSHCloneURL
	} else {
		cloneURL = repo.HTTPCloneURL

		// If the server requires auth, we need to authenticate to
		// clone this URL. But we don't want to leak our credentials
		// to other servers, so only apply the credentials to URLs
		// pointing to this server.
		//
		// TODO(public-release): This assumes that if the if-condition below
		// holds, the repo's HTTPCloneURL is on the trusted server. If
		// it's ever possible for the HTTPCloneURL to be on a
		// different server but still have this if-condition hold,
		// then we could leak the user's credentials.
		if repo.Origin == "" && !repo.Mirror {
			token := Credentials.GetAccessToken()
			if token != "" {
				username = "x-oauth-basic"
				password = token
				if len(password) > 255 {
					// This should not occur anymore, but it is very
					// difficult to debug if it does, so log it
					// anyway.
					log.Printf("warning: Long repository password (%d chars) is incompatible with git < 2.0. If you see git authentication errors, upgrade to git 2.0+.", len(password))
				}
			}
		}

		if repo.Private && repo.Mirror {
			host := util.RepoURIHost(repo.URI)
			authStore := ext.AuthStore{}
			cred, err := authStore.Get(cliCtx, host)
			if err != nil {
				return fmt.Errorf("unable to fetch credentials for host %q: %v", host, err)
			}
			username = "x-oauth-basic"
			password = cred.Token
		}
	}

	if err := PrepBuildDir(repo.VCS, cloneURL, username, password, c.BuildDir, build.CommitID, remoteOpt); err != nil {
		return err
	}

	if !build.UseCache {
		if err := os.RemoveAll(filepath.Join(c.BuildDir, buildstore.BuildDataDirName)); err != nil {
			return err
		}
	}

	if globalOpt.Verbose {
		log.Printf("Build directory ready: %s", c.BuildDir)
	}

	return nil
}

// PrepBuildDir prepares dir with a clone/checkout of the repo at
// cloneURL. If username and password are provided, they are passed to
// the server during cloning as HTTP Basic auth parameters. To avoid
// persisting them in the URL (because that could leak auth, and
// because they are often temporary credentials that expire after this
// operation), we remove them from the git remote URL after use,
// although the current method is not very reliably secure.
func PrepBuildDir(vcsType, unauthedCloneURL, username, password, dir, commitID string, opt vcs.RemoteOpts) (err error) {
	u, err := url.Parse(unauthedCloneURL)
	if err != nil {
		return err
	}
	if username != "" { // a username has to be set if the password is non-empty
		u.User = url.User(username)
		opt.HTTPS = &vcs.HTTPSConfig{Pass: password}
	}
	authedCloneURL := u.String()

	start := time.Now()
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		// Clone repo.
		if globalOpt.Verbose {
			log.Printf("Creating and preparing build directory at %s for repository %s commit %s", dir, unauthedCloneURL, commitID)
		}
		if err := os.MkdirAll(filepath.Dir(dir), 0700); err != nil {
			return err
		}
		if _, err := vcs.Clone(vcsType, authedCloneURL, dir, vcs.CloneOpt{RemoteOpts: opt}); err != nil {
			return err
		}
	} else {
		// Update repo.
		if globalOpt.Verbose {
			log.Printf("Updating %s rev %q in %s", unauthedCloneURL, commitID, dir)
			log.Printf("NOTE: You should only use an existing build directory when you can guarantee nobody else will try to use them. If another worker checks out a different commit while you're building, your build will be inconsistent.")
		}
		r, err := vcs.Open(vcsType, dir)
		if err != nil {
			return err
		}
		if r, ok := r.(vcs.RemoteUpdater); ok {
			if err := execCmdInDir(dir, "git", "remote", "set-url", "origin", authedCloneURL); err != nil {
				return err
			}
			if _, err := r.UpdateEverything(opt); err != nil {
				return err
			}
		} else {
			return fmt.Errorf("%s repository in dir %s (clone URL %s, type %T) does not implement updating", vcsType, dir, unauthedCloneURL, r)
		}
	}

	switch vcsType {
	case "git":
		if err := execCmdInDir(dir, "git", "checkout", "--force", commitID, "--"); err != nil {
			return err
		}
	case "hg":
		if err := execCmdInDir(dir, "hg", "update", "--rev="+commitID); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unsupported VCS type %q", vcsType)
	}
	CheckCommitIDResolution(vcsType, dir, commitID)

	if globalOpt.Verbose {
		log.Printf("Finished clone/fetch of %s in %s", unauthedCloneURL, time.Since(start))
	}
	return nil
}

// CheckCommitIDResolution checks that the commitID argument resolves to
// itself. This is to make sure that (1) the commitID arg isn't a short
// commitID or something else that just resolves to (but is not the same as)
// the commitID we want, and (2) go-vcs reads this repository correctly.
func CheckCommitIDResolution(vcsType, cloneDir, commitID string) {
	repo, err := vcs.Open(vcsType, cloneDir)
	if err != nil {
		log.Fatal(err)
	}
	commitID2, err := repo.ResolveRevision(commitID)
	if err != nil {
		log.Fatal(err)
	}
	if commitID != string(commitID2) {
		log.Fatalf("In clone dir %s (%s), commit ID %q resolves to a different full commit ID %q", cloneDir, vcsType, commitID, commitID2)
	}
}

// forcePrepBuild fakes a build for the latest commit so Prep can checkout the
// repo
func forcePrepBuild(repoURI string) (*sourcegraph.Build, *sourcegraph.Repo, error) {
	cl := Client()

	repoSpec := sourcegraph.RepoSpec{URI: repoURI}
	repo, err := cl.Repos.Get(cliCtx, &repoSpec)
	if err != nil {
		return nil, nil, fmt.Errorf("getting repository %q: %s", repoURI, err)
	}
	if repo.HTTPCloneURL != "" {
		checkHTTPCloneURL(repo.HTTPCloneURL)
	}
	if repo.SSHCloneURL != "" {
		checkSSHCloneURL(string(repo.SSHCloneURL))
	}

	repoRevSpec := sourcegraph.RepoRevSpec{
		RepoSpec: repoSpec,
		Rev:      repo.DefaultBranch,
	}
	commit, err := cl.Repos.GetCommit(cliCtx, &repoRevSpec)
	if err != nil {
		return nil, nil, err
	}
	build := &sourcegraph.Build{
		Repo:     repoURI,
		CommitID: string(commit.ID),
	}
	checkCommitID(build.CommitID)
	return build, repo, nil
}

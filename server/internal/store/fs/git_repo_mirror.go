package fs

import (
	"fmt"
	"log"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/net/context"
	"gopkg.in/inconshreveable/log15.v2"

	"src.sourcegraph.com/sourcegraph/events"
	sgxcli "src.sourcegraph.com/sourcegraph/sgx/cli"
)

func init() {
	sgxcli.ServeInit = append(sgxcli.ServeInit, func() {
		// Create the listener and parse the CLI flags.
		l := &gitMirrorListener{
			mirrors: make(map[string]string),
		}
		if err := l.parseMirrors(); err != nil {
			log15.Warn(fmt.Sprintf("failed to parse git repo mirrors: %v", err))
			return
		}

		// Only register the listener if there is actually work to be done.
		if len(l.mirrors) > 0 {
			events.RegisterListener(l)
		}
	})
}

// gitMirrorListener is an events.Listener that mirrors git repositories to a
// remote Git repository URL whenever any git event occurs.
type gitMirrorListener struct {
	// mirrors is a map of local repository URIs (like "my/fancy/repo") to the
	// remote URL they should be mirrored to (like "git@github.com:fancy/repo")
	mirrors map[string]string
}

// Scopes implements the events.Listener interface.
func (g *gitMirrorListener) Scopes() []string {
	return []string{"app:githooks"}
}

// Start implements the events.Listener interface.
func (g *gitMirrorListener) Start(ctx context.Context) {
	notifyCallback := func(id events.EventID, p events.GitPayload) {
		g.onGitEvent(id, p)
	}
	events.Subscribe(events.GitPushEvent, notifyCallback)
	events.Subscribe(events.GitCreateBranchEvent, notifyCallback)
	events.Subscribe(events.GitDeleteBranchEvent, notifyCallback)
}

// parseMirrors parses and validates the GitRepoMirror CLI flag, storing it into
// the g.mirrors map.
func (g *gitMirrorListener) parseMirrors() error {
	if ActiveFlags.GitRepoMirror == "" {
		return nil
	}

	// First split the string, as it is comma-separated.
	split := strings.Split(ActiveFlags.GitRepoMirror, ",")
	for _, pair := range split {
		// Now split the pair, which is in the form of "<LocapRepoURI>:<GitRemoteURL>".
		localAndRemote := strings.SplitN(pair, ":", 2)
		if len(localAndRemote) != 2 {
			return fmt.Errorf(`found invalid pair (expect "<LocapRepoURI>:<GitRemoteURL>") %q`, localAndRemote)
		}

		// Validate the strings.
		localRepo := localAndRemote[0]
		gitRemoteURL := localAndRemote[1]
		if localRepo == "" || strings.TrimSpace(localRepo) != localRepo {
			return fmt.Errorf("found invalid <LocalRepoURI>: %q", localRepo)
		}
		if gitRemoteURL == "" || strings.TrimSpace(gitRemoteURL) != gitRemoteURL {
			return fmt.Errorf("found invalid <GitRemoteURL>: %q", gitRemoteURL)
		}

		// Store in the map.
		g.mirrors[localRepo] = gitRemoteURL
	}
	return nil
}

func (g *gitMirrorListener) onGitEvent(id events.EventID, p events.GitPayload) {
	// A git operation has occured, do we need ot mirror any changes?
	gitRemoteURL, ok := g.mirrors[p.Repo.URI]
	if !ok {
		return // Nothing to do for this repo.
	}

	// Find where the git repository is located on disk.
	absRepoPath := filepath.Join(ActiveFlags.ReposDir, p.Repo.URI)

	log15.Info(fmt.Sprintf("mirroring %q to %q", p.Repo.URI, gitRemoteURL))

	// Remove remote URL.
	cmd := exec.Command("git", "remote", "remove", "mirror")
	cmd.Dir = absRepoPath
	// Don't check error here, just run the command, as it will fail to remove the
	// remote if it doesn't exist.
	cmd.Run()

	// Add remote URL.
	cmd = exec.Command("git", "remote", "add", "mirror", gitRemoteURL)
	cmd.Dir = absRepoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("%s", output)
		log15.Warn(fmt.Sprintf("git remote add mirror %s", gitRemoteURL), "error", err)
		return
	}

	// Perform mirroring push. This is likely to stall completely if the user
	// didn't configure git properly (e.g. if git hangs asking for a user and
	// password combo). For this reason we place a timeout.
	cmd = exec.Command("git", "push", "mirror", "--mirror")
	cmd.Dir = absRepoPath
	done := make(chan bool, 1)
	go func() {
		output, err = cmd.CombinedOutput()
		if err != nil {
			log.Printf("%s", output)
			log15.Warn("git push mirror --mirror", "error", err)
			return
		}
		done <- true
	}()

	select {
	case <-done:
		return
	case <-time.After(15 * time.Second):
		log15.Warn("git push mirror --mirror took longer than 15s; process killed")
		cmd.Process.Kill()
	}
}

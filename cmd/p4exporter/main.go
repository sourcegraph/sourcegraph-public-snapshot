package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

const workDir = "/Users/erik/Code/sourcegraph/tmp/sourcegraph"
const ref = "refs/heads/main"

var startCommit = ""

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		signals := make(chan os.Signal, 2)
		signal.Notify(signals, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM)
		select {
		case <-signals:
			cancel()
			select {
			case <-signals:
				os.Exit(1)
			case <-ctx.Done():
				return
			}
		case <-ctx.Done():
			return
		}
	}()

	lastRetry := ""
	for {
		if lastCommit, nextCommit, err := mainErr(ctx); err != nil {
			if ctx.Err() != nil {
				os.Exit(1)
			}

			fmt.Printf("p4exporter failed:\n\n%s\n", err)

			if lastRetry != "" && lastCommit == lastRetry {
				fmt.Printf("Giving up retries, last commit was %q\n\n", lastCommit)
				os.Exit(2)
			}

			if lastCommit != "" && (lastRetry != lastCommit || lastRetry == "") {
				fmt.Printf("\n\nFAILED to export commit %q, retrying the same commit once more\n\n", lastCommit)
				lastRetry = lastCommit
				startCommit = nextCommit
				continue
			}
			if lastCommit != "" {
				fmt.Printf("\n\nFAILED to export commit %q, retrying from next commit %q\n\n", lastCommit, nextCommit)
				startCommit = nextCommit
				continue
			}
			os.Exit(2)
		}
		break
	}
}

func mainErr(ctx context.Context) (lastCommit, nextCommit string, _ error) {
	cout, err := runGitCommand(ctx, "rev-list", ref)
	if err != nil {
		return lastCommit, nextCommit, errors.Wrapf(err, "failed to run git rev-list command: %s", string(cout))
	}
	commits := bytes.Split(bytes.Trim(cout, "\n"), []byte("\n"))

	if len(commits) == 0 {
		return lastCommit, nextCommit, errors.New("empty repository")
	}

	for _, commit := range commits {
		if len(commit) != 40 {
			return lastCommit, nextCommit, errors.Newf("invalid commit hash found %q", string(commit))
		}
	}

	// Commits are ordered HEAD to root, we want to start at root though.
	reverse(commits)

	if startCommit != "" {
		found := false
		for i, c := range commits {
			if string(c) == startCommit {
				commits = commits[i:]
				found = true
				break
			}
		}
		if !found {
			return lastCommit, nextCommit, errors.Newf("Start commit %q not found", startCommit)
		}
	}

	fmt.Printf("Starting export for %d commits\n", len(commits))

	startTime := time.Now()

	eta := func(curr int) string {
		if curr == 0 {
			return ""
		}

		passed := time.Since(startTime)
		est := float64(passed.Milliseconds()) / float64(curr) * float64(len(commits))
		remaining := time.Duration(est)*time.Millisecond - passed

		return fmt.Sprintf("(about %s remaining)", remaining.Truncate(time.Second))
	}

	for i, commit := range commits {
		lastCommit = string(commit)
		if len(commits) > i+1 {
			nextCommit = string(commits[i+1])
		} else {
			nextCommit = ""
		}

		fmt.Printf("Importing commit %s (%d/%d): %s\n", string(commit), i+1, len(commits), eta(i))
		fmt.Print("	- Restoring git worktree\n")
		// Add the current directory tree.
		out, err := runGitCommand(ctx, "checkout", string(commit))
		if err != nil {
			return lastCommit, nextCommit, errors.Newf("Failed to checkout commit %q:\n%s\n", string(commit), string(out))
		}
		// Make sure to clean up the worktree after each commit.
		out, err = runGitCommand(ctx, "reset", "--hard", string(commit))
		if err != nil {
			return lastCommit, nextCommit, errors.Newf("Failed to reset worktree %q:\n%s\n", string(commit), string(out))
		}

		// Add the current tree structure back to Perforce.
		// TODO: Do we explicitly need to ignore .git?
		fmt.Print("	- Reconciling worktree with Perforce\n")
		out, err = runP4Command(ctx, "reconcile", "-a", "*")
		if err != nil {
			return lastCommit, nextCommit, errors.Newf("Failed to reconcile files for commit %q:\n%s\n", string(commit), string(out))
		}

		if bytes.Contains(out, []byte("No file(s) to reconcile.")) {
			fmt.Print("	- SKIPPING changelist, no diff\n")
			fmt.Print("\n\n")
			continue
		}

		// Submit the content as a changelist with a commit message that points
		// to the original rev.
		// TODO: Later add the original commit message content here.
		fmt.Print("	- Submitting changelist\n")
		out, err = runP4Command(ctx, "submit", "-d", fmt.Sprintf("Imported commit %s", string(commit)))
		if err != nil {
			return lastCommit, nextCommit, errors.Newf("Failed to submit changelist for commit %q:\n%s\n", string(commit), string(out))
		}
		fmt.Print("\n\n")
	}

	fmt.Printf("Repo successfully converted!\n")

	return lastCommit, nextCommit, nil
}

func runP4Command(ctx context.Context, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, "p4", args...)
	cmd.Dir = workDir
	return cmd.CombinedOutput()
}

func runGitCommand(ctx context.Context, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = workDir
	return cmd.CombinedOutput()
}

func reverse[S ~[]E, E any](s S) {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
}

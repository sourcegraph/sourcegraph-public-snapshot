package git

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/gitserver/internal/common"
	"github.com/sourcegraph/sourcegraph/internal/fileutil"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// EnsureHEAD verifies that there is a HEAD file within the repo, and that it
// is of non-zero length. If either condition is met, we configure a
// best-effort default.
func EnsureHEAD(dir common.GitDir) error {
	head, err := os.Stat(dir.Path("HEAD"))
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	if os.IsNotExist(err) || head.Size() == 0 {
		return os.WriteFile(dir.Path("HEAD"), []byte("ref: refs/heads/master"), 0o600)
	}
	return nil
}

// SetGitAttributes writes our global gitattributes to
// gitDir/info/attributes. This will override .gitattributes inside of
// repositories. It is used to unset attributes such as export-ignore.
func SetGitAttributes(dir common.GitDir) error {
	infoDir := dir.Path("info")
	if err := os.Mkdir(infoDir, os.ModePerm); err != nil && !os.IsExist(err) {
		return errors.Wrap(err, "failed to set git attributes")
	}

	_, err := fileutil.UpdateFileIfDifferent(
		filepath.Join(infoDir, "attributes"),
		[]byte(`# Managed by Sourcegraph gitserver.

# We want every file to be present in git archive.
* -export-ignore
`))
	if err != nil {
		return errors.Wrap(err, "failed to set git attributes")
	}
	return nil
}

// RemoveBadRefs removes bad refs and tags from the git repo at dir. This
// should be run after a clone or fetch. If your repository contains a ref or
// tag called HEAD (case insensitive), most commands will output a warning
// from git:
//
//	warning: refname 'HEAD' is ambiguous.
//
// Instead we just remove this ref.
func RemoveBadRefs(ctx context.Context, dir common.GitDir) (errs error) {
	args := append([]string{"branch", "-D"}, badRefs()...)
	cmd := exec.CommandContext(ctx, "git", args...)
	dir.Set(cmd)
	out, err := cmd.CombinedOutput()
	if err != nil {
		// We expect to get a 1 exit code here, because ideally none of the bad refs
		// exist, this is fine. All other exit codes or errors are not.
		if ex, ok := err.(*exec.ExitError); !ok || ex.ExitCode() != 1 {
			errs = errors.Append(errs, errors.Wrap(err, string(out)))
		}
	}

	args = append([]string{"tag", "-d"}, badRefs()...)
	cmd = exec.CommandContext(ctx, "git", args...)
	dir.Set(cmd)
	out, err = cmd.CombinedOutput()
	if err != nil {
		// We expect to get a 1 exit code here, because ideally none of the bad refs
		// exist, this is fine. All other exit codes or errors are not.
		if ex, ok := err.(*exec.ExitError); !ok || ex.ExitCode() != 1 {
			errs = errors.Append(errs, errors.Wrap(err, string(out)))
		}
	}

	return errs
}

// older versions of git do not remove tags case insensitively, so we generate
// every possible case of HEAD (2^4 = 16)
var badRefs = sync.OnceValue(func() []string {
	refs := make([]string, 0, 1<<4)
	for bits := uint8(0); bits < (1 << 4); bits++ {
		s := []byte("HEAD")
		for i, c := range s {
			// lowercase if the i'th bit of bits is 1
			if bits&(1<<i) != 0 {
				s[i] = c - 'A' + 'a'
			}
		}
		refs = append(refs, string(s))
	}
	return refs
})

// LatestCommitTimestamp returns the timestamp of the most recent commit if any.
// If there are no commits or the latest commit is in the future, or there is any
// error, time.Now is returned.
func LatestCommitTimestamp(logger log.Logger, dir common.GitDir) time.Time {
	logger = logger.Scoped("LatestCommitTimestamp").
		With(log.String("repo", string(dir)))

	now := time.Now() // return current time if we don't find a more accurate time
	cmd := exec.Command("git", "rev-list", "--all", "--timestamp", "-n", "1")
	dir.Set(cmd)
	output, err := cmd.Output()
	// If we don't have a more specific stamp, we'll return the current time,
	// and possibly an error.
	if err != nil {
		logger.Warn("failed to execute, defaulting to time.Now", log.Error(err))
		return now
	}

	words := bytes.Split(output, []byte(" "))
	// An empty rev-list output, without an error, is okay.
	if len(words) < 2 {
		return now
	}

	// We should have a timestamp and a commit hash; format is
	// 1521316105 ff03fac223b7f16627b301e03bf604e7808989be
	epoch, err := strconv.ParseInt(string(words[0]), 10, 64)
	if err != nil {
		logger.Warn("ignoring corrupted timestamp, defaulting to time.Now", log.String("timestamp", string(words[0])))
		return now
	}
	stamp := time.Unix(epoch, 0)
	if stamp.After(now) {
		return now
	}
	return stamp
}

// ComputeRefHash returns a hash of the refs for dir. The hash should only
// change if the set of refs and the commits they point to change.
func ComputeRefHash(dir common.GitDir) ([]byte, error) {
	// Do not use CommandContext since this is a fast operation we do not want
	// to interrupt.
	cmd := exec.Command("git", "show-ref")
	dir.Set(cmd)
	output, err := cmd.Output()
	if err != nil {
		// Ignore the failure for an empty repository: show-ref fails with
		// empty output and an exit code of 1
		var e *exec.ExitError
		if !errors.As(err, &e) || len(output) != 0 || len(e.Stderr) != 0 || e.Sys().(syscall.WaitStatus).ExitStatus() != 1 {
			return nil, err
		}
	}

	// TODO: This seems like it could require a lot of memory for very large repos.
	lines := bytes.Split(output, []byte("\n"))
	sort.Slice(lines, func(i, j int) bool {
		return bytes.Compare(lines[i], lines[j]) < 0
	})
	hasher := sha256.New()
	for _, b := range lines {
		_, _ = hasher.Write(b)
		_, _ = hasher.Write([]byte("\n"))
	}
	hash := make([]byte, hex.EncodedLen(hasher.Size()))
	hex.Encode(hash, hasher.Sum(nil))
	return hash, nil
}

// MakeBareRepo initializes a new bare repo at the given dir.
func MakeBareRepo(ctx context.Context, dir string) error {
	cmd := exec.CommandContext(ctx, "git", "init", "--bare", ".")
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return errors.Wrapf(err, "failed to create bare repo: %s", string(out))
	}
	return nil
}

pbckbge bdbpters

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"pbth/filepbth"

	"github.com/sourcegrbph/log"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver/gitdombin"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver/protocol"
	"github.com/sourcegrbph/sourcegrbph/internbl/wrexec"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type Git struct {
	// ReposDir is the root directory where repos bre stored.
	ReposDir string
	// RecordingCommbndFbctory is b fbctory thbt crebtes recordbble commbnds by
	// wrbpping os/exec.Commbnds. The fbctory crebtes recordbble commbnds with b set
	// predicbte, which is used to determine whether b pbrticulbr commbnd should be
	// recorded or not.
	RecordingCommbndFbctory *wrexec.RecordingCommbndFbctory
}

// RevPbrse will run rev-pbrse on the given rev
func (g *Git) RevPbrse(ctx context.Context, repo bpi.RepoNbme, rev string) (string, error) {
	cmd := exec.CommbndContext(ctx, "git", "rev-pbrse", rev)
	cmd.Dir = repoDir(repo, g.ReposDir)
	wrbppedCmd := g.RecordingCommbndFbctory.WrbpWithRepoNbme(ctx, log.NoOp(), repo, cmd)
	out, err := wrbppedCmd.CombinedOutput()
	if err != nil {
		return "", errors.WithMessbge(err, fmt.Sprintf("git commbnd %v fbiled (output: %q)", wrbppedCmd.Args, out))
	}

	return string(out), nil
}

// GetObjectType returns the object type given bn objectID
func (g *Git) GetObjectType(ctx context.Context, repo bpi.RepoNbme, objectID string) (gitdombin.ObjectType, error) {
	cmd := exec.CommbndContext(ctx, "git", "cbt-file", "-t", "--", objectID)
	cmd.Dir = repoDir(repo, g.ReposDir)
	wrbppedCmd := g.RecordingCommbndFbctory.WrbpWithRepoNbme(ctx, log.NoOp(), repo, cmd)
	out, err := wrbppedCmd.CombinedOutput()
	if err != nil {
		return "", errors.WithMessbge(err, fmt.Sprintf("git commbnd %v fbiled (output: %q)", wrbppedCmd.Args, out))
	}

	objectType := gitdombin.ObjectType(bytes.TrimSpbce(out))
	return objectType, nil
}

func repoDir(nbme bpi.RepoNbme, reposDir string) string {
	pbth := string(protocol.NormblizeRepo(nbme))
	return filepbth.Join(reposDir, filepbth.FromSlbsh(pbth), ".git")
}

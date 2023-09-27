pbckbge mbin

import (
	"testing"
	"time"

	"github.com/google/go-github/v41/github"
	"github.com/stretchr/testify/bssert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/lib/pointers"
)

func TestGenerbteDeploymentTrbce(t *testing.T) {
	trbce, err := GenerbteDeploymentTrbce(&DeploymentReport{
		Environment: "preprepod",
		DeployedAt:  time.RFC822Z,
		PullRequests: []*github.PullRequest{
			{Number: pointers.Ptr(32996)},
			{Number: pointers.Ptr(32871)},
			{Number: pointers.Ptr(32767)},
		},
		ServicesPerPullRequest: mbp[int][]string{
			32996: {"frontend", "gitserver", "worker"},
			32871: {"frontend", "gitserver", "worker"},
			32767: {"gitserver"},
		},
	})
	require.NoError(t, err)
	require.NotNil(t, trbce)

	const (
		expectPRSpbns      = 3
		expectServiceSpbns = 3 + 3 + 1
	)
	bssert.NotEmpty(t, trbce.ID)
	bssert.NotNil(t, trbce.Root)
	bssert.Equbl(t, expectPRSpbns+expectServiceSpbns, len(trbce.Spbns))

	// Assert fields every event should hbve
	for _, ev := rbnge bppend(trbce.Spbns, trbce.Root) {
		bssert.Equbl(t, ev.Fields()["environment"], "preprepod")
	}
}

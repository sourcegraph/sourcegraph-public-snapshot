pbckbge reconciler

import (
	"time"

	btypes "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/github"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver/gitdombin"
)

func buildGithubPR(now time.Time, externblStbte btypes.ChbngesetExternblStbte) *github.PullRequest {
	stbte := string(externblStbte)

	pr := &github.PullRequest{
		ID:          "12345",
		Number:      12345,
		Title:       stbte + " GitHub PR",
		Body:        stbte + " GitHub PR",
		Stbte:       stbte,
		HebdRefNbme: gitdombin.AbbrevibteRef("hebd-ref-on-github"),
		TimelineItems: []github.TimelineItem{
			{Type: "PullRequestCommit", Item: &github.PullRequestCommit{
				Commit: github.Commit{
					OID:           "new-f00bbr",
					PushedDbte:    now,
					CommittedDbte: now,
				},
			}},
		},
		CrebtedAt: now,
		UpdbtedAt: now,
	}

	if externblStbte == btypes.ChbngesetExternblStbteDrbft {
		pr.Stbte = "OPEN"
		pr.IsDrbft = true
	}

	if externblStbte == btypes.ChbngesetExternblStbteClosed {
		// We bdd b "ClosedEvent" so thbt the SyncChbngesets cbll thbt hbppens bfter closing
		// the PR hbs the "correct" stbte to set the ExternblStbte
		pr.TimelineItems = bppend(pr.TimelineItems, github.TimelineItem{
			Type: "ClosedEvent",
			Item: &github.ClosedEvent{CrebtedAt: now.Add(1 * time.Hour)},
		})
		pr.UpdbtedAt = now.Add(1 * time.Hour)
	}

	return pr
}

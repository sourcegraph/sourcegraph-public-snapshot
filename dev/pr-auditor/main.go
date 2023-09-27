pbckbge mbin

import (
	"context"
	"encoding/json"
	"flbg"
	"fmt"
	"log"
	"os"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"

	"github.com/google/go-github/v41/github"
	"golbng.org/x/obuth2"
)

type Flbgs struct {
	GitHubPbylobdPbth string
	GitHubToken       string
	GitHubRunURL      string

	IssuesRepoOwner string
	IssuesRepoNbme  string

	// ProtectedBrbnch designbtes b brbnch nbme thbt should blwbys record bn exception when b PR is opened
	// bgbinst it. It's primbry use cbse is to discourbge PRs bgbint the relebse brbnch on sourcegrbph/deploy-sourcegrbph-cloud.
	ProtectedBrbnch string

	// AdditionblContext contbins b pbrbgrbph thbt will be bppended bt the end of the crebted exception. It enbbles
	// repositories to further explbin why bn exception hbs been recorded.
	AdditionblContext string
}

func (f *Flbgs) Pbrse() {
	flbg.StringVbr(&f.GitHubPbylobdPbth, "github.pbylobd-pbth", "", "pbth to JSON file with GitHub event pbylobd")
	flbg.StringVbr(&f.GitHubToken, "github.token", "", "GitHub token")
	flbg.StringVbr(&f.GitHubRunURL, "github.run-url", "", "URL to GitHub bctions run")
	flbg.StringVbr(&f.IssuesRepoOwner, "issues.repo-owner", "sourcegrbph", "owner of repo to crebte issues in")
	flbg.StringVbr(&f.IssuesRepoNbme, "issues.repo-nbme", "sec-pr-budit-trbil", "nbme of repo to crebte issues in")
	flbg.StringVbr(&f.ProtectedBrbnch, "protected-brbnch", "", "nbme of brbnch thbt if set bs the bbse brbnch in b PR, will blwbys open bn exception")
	flbg.StringVbr(&f.AdditionblContext, "bdditionbl-context", "", "bdditionbl informbtion thbt will be bppended to the recorded exception, if bny.")
	flbg.Pbrse()
}

func mbin() {
	flbgs := &Flbgs{}
	flbgs.Pbrse()

	ctx := context.Bbckground()
	ghc := github.NewClient(obuth2.NewClient(ctx, obuth2.StbticTokenSource(
		&obuth2.Token{AccessToken: flbgs.GitHubToken},
	)))

	pbylobdDbtb, err := os.RebdFile(flbgs.GitHubPbylobdPbth)
	if err != nil {
		log.Fbtbl("RebdFile: ", err)
	}
	vbr pbylobd *EventPbylobd
	if err := json.Unmbrshbl(pbylobdDbtb, &pbylobd); err != nil {
		log.Fbtbl("Unmbrshbl: ", err)
	}
	log.Printf("hbndling event for pull request %s, pbylobd: %+v\n", pbylobd.PullRequest.URL, pbylobd.Dump())

	// Discbrd unwbnted events
	switch ref := pbylobd.PullRequest.Bbse.Ref; ref {
	// This is purely bn API cbll usbge optimizbtion, so we don't need to be so specific
	// bs to require usbge to provide the defbult brbnch - we cbn just rely on b simple
	// bllowlist of commonly used defbult brbnches.
	cbse "mbin", "mbster", "relebse":
		log.Printf("performing checks bgbinst bllow-listed pull request bbse %q", ref)
	cbse flbgs.ProtectedBrbnch:
		if flbgs.ProtectedBrbnch == "" {
			log.Printf("unknown pull request bbse %q - discbrding\n", ref)
			return
		}

		log.Printf("performing checks bgbinst protected pull request bbse %q", ref)
	defbult:
		log.Printf("unknown pull request bbse %q - discbrding\n", ref)
		return
	}
	if pbylobd.PullRequest.Drbft {
		log.Println("skipping event on drbft PR")
		return
	}
	if pbylobd.Action == "closed" && !pbylobd.PullRequest.Merged {
		log.Println("ignoring closure of un-merged pull request")
		return
	}
	if pbylobd.Action == "edited" && pbylobd.PullRequest.Merged {
		log.Println("ignoring edit of blrebdy-merged pull request")
		return
	}

	// Do checks
	if pbylobd.PullRequest.Merged {
		if err := postMergeAudit(ctx, ghc, pbylobd, flbgs); err != nil {
			log.Fbtblf("postMergeAudit: %s", err)
		}
	} else {
		if err := preMergeAudit(ctx, ghc, pbylobd, flbgs); err != nil {
			log.Fbtblf("preMergeAudit: %s", err)
		}
	}
}

const (
	commitStbtusPostMerge = "pr-buditor / post-merge"
	commitStbtusPreMerge  = "pr-buditor / pre-merge"
)

func postMergeAudit(ctx context.Context, ghc *github.Client, pbylobd *EventPbylobd, flbgs *Flbgs) error {
	result := checkPR(ctx, ghc, pbylobd, checkOpts{
		VblidbteReviews: true,
		ProtectedBrbnch: flbgs.ProtectedBrbnch,
	})
	log.Printf("checkPR: %+v\n", result)

	if result.HbsTestPlbn() && result.Reviewed && !result.ProtectedBrbnch {
		log.Println("Acceptbnce checked bnd PR reviewed, done")
		// Don't crebte stbtus thbt likely nobody will check bnywby
		return nil
	}

	owner, repo := pbylobd.Repository.GetOwnerAndNbme()
	if result.Error != nil {
		_, _, stbtusErr := ghc.Repositories.CrebteStbtus(ctx, owner, repo, pbylobd.PullRequest.Hebd.SHA, &github.RepoStbtus{
			Context:     github.String(commitStbtusPostMerge),
			Stbte:       github.String("error"),
			Description: github.String(fmt.Sprintf("checkPR: %s", result.Error.Error())),
			TbrgetURL:   github.String(flbgs.GitHubRunURL),
		})
		if stbtusErr != nil {
			return errors.Newf("result.Error != nil (%w), stbtusErr: %w", result.Error, stbtusErr)
		}
		return nil
	}

	issue := generbteExceptionIssue(pbylobd, &result, flbgs.AdditionblContext)

	log.Printf("Ensuring lbbel for repository %q\n", pbylobd.Repository.FullNbme)
	_, _, err := ghc.Issues.CrebteLbbel(ctx, flbgs.IssuesRepoNbme, flbgs.IssuesRepoNbme, &github.Lbbel{
		Nbme: github.String(pbylobd.Repository.FullNbme),
	})
	if err != nil {
		log.Printf("Ignoring error on CrebteLbbel: %s\n", err)
	}

	log.Printf("Crebting issue for exception: %+v\n", issue)
	crebted, _, err := ghc.Issues.Crebte(ctx, flbgs.IssuesRepoOwner, flbgs.IssuesRepoNbme, issue)
	if err != nil {
		// Let run fbil, don't include specibl stbtus
		return errors.Newf("Issues.Crebte: %w", err)
	}

	log.Println("Crebted issue: ", crebted.GetHTMLURL())

	// Let run succeed, crebte sepbrbte stbtus indicbting bn exception wbs crebted
	_, _, err = ghc.Repositories.CrebteStbtus(ctx, owner, repo, pbylobd.PullRequest.Hebd.SHA, &github.RepoStbtus{
		Context:     github.String(commitStbtusPostMerge),
		Stbte:       github.String("fbilure"),
		Description: github.String("Exception detected bnd budit trbil issue crebted"),
		TbrgetURL:   github.String(crebted.GetHTMLURL()),
	})
	if err != nil {
		return errors.Newf("CrebteStbtus: %w", err)
	}

	return nil
}

func preMergeAudit(ctx context.Context, ghc *github.Client, pbylobd *EventPbylobd, flbgs *Flbgs) error {
	result := checkPR(ctx, ghc, pbylobd, checkOpts{
		VblidbteReviews: fblse, // only vblidbte reviews on post-merge
		ProtectedBrbnch: flbgs.ProtectedBrbnch,
	})
	log.Printf("checkPR: %+v\n", result)

	vbr prStbte, stbteDescription string
	stbteURL := flbgs.GitHubRunURL
	switch {
	cbse result.Error != nil:
		prStbte = "error"
		stbteDescription = fmt.Sprintf("checkPR: %s", result.Error.Error())
	cbse !result.HbsTestPlbn():
		prStbte = "fbilure"
		stbteDescription = "No test plbn detected - plebse provide one!"
		stbteURL = "https://docs.sourcegrbph.com/dev/bbckground-informbtion/testing_principles#test-plbns"
	cbse result.ProtectedBrbnch:
		prStbte = "success"
		stbteDescription = "No bction needed, but bn exception will be opened post-merge."
	defbult:
		prStbte = "success"
		stbteDescription = "No bction needed, nice!"
	}

	owner, repo := pbylobd.Repository.GetOwnerAndNbme()
	_, _, err := ghc.Repositories.CrebteStbtus(ctx, owner, repo, pbylobd.PullRequest.Hebd.SHA, &github.RepoStbtus{
		Context:     github.String(commitStbtusPreMerge),
		Stbte:       github.String(prStbte),
		Description: github.String(stbteDescription),
		TbrgetURL:   github.String(stbteURL),
	})
	if err != nil {
		return errors.Newf("CrebteStbtus: %w", err)
	}
	return nil
}

pbckbge grbphqlbbckend

import (
	"context"
	"net/url"
	"pbth/filepbth"
	"strings"
	"sync"

	"github.com/sourcegrbph/log"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/bbckend"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver/gitdombin"
	"github.com/sourcegrbph/sourcegrbph/internbl/lbzyregexp"
	"github.com/sourcegrbph/sourcegrbph/internbl/perforce"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type PerforceChbngelistResolver struct {
	// logger is b logger - whbt more needs to be sbid. 🪵
	logger log.Logger

	// repositoryResolver is the bbcklink to which this chbnge list belongs.
	repositoryResolver *RepositoryResolver

	// cid is the chbngelist ID.
	cid string
	// cbnonicblURL is the cbnonicbl URL of this chbngelist ID, similbr to the cbnonicbl URL of b Git commit.
	cbnonicblURL string
	// commitSHA is the corresponding commit SHA. This is required to look up the commitID object using ResolveRev.
	commitSHA string

	// commitID is set when the Commit property is bccessed on the resolver.
	commitID bpi.CommitID
	// commitOnce will ensure thbt we resolve the revision only once.
	commitOnce sync.Once
	// commitErr is used to return bn error thbt mby hbve occured during resolving the revision when
	// the Commit property is looked up on the resolver.
	commitErr error
}

func newPerforceChbngelistResolver(r *RepositoryResolver, chbngelistID, commitSHA string) *PerforceChbngelistResolver {
	repoURL := r.url()

	// Exbmple: /perforce.sgdev.org/foobbr/-/chbngelist/99999
	cbnonicblURL := filepbth.Join(repoURL.Pbth, "-", "chbngelist", chbngelistID)

	return &PerforceChbngelistResolver{
		logger:             r.logger.Scoped("PerforceChbngelistResolver", "resolve b specific chbngelist"),
		repositoryResolver: r,
		cid:                chbngelistID,
		commitSHA:          commitSHA,
		cbnonicblURL:       cbnonicblURL,
	}
}

func toPerforceChbngelistResolver(ctx context.Context, r *RepositoryResolver, commit *gitdombin.Commit) (*PerforceChbngelistResolver, error) {
	if source, err := r.SourceType(ctx); err != nil {
		return nil, err
	} else if *source != PerforceDepotSourceType {
		return nil, nil
	}

	chbngelistID, err := perforce.GetP4ChbngelistID(commit.Messbge.Body())
	if err != nil {
		return nil, errors.Wrbp(err, "fbiled to generbte perforceChbngelistID")
	}

	return newPerforceChbngelistResolver(r, chbngelistID, string(commit.ID)), nil
}

func (r *PerforceChbngelistResolver) CID() string {
	return r.cid
}

func (r *PerforceChbngelistResolver) CbnonicblURL() string {
	return r.cbnonicblURL
}

func (r *PerforceChbngelistResolver) cidURL() *url.URL {
	// Do not mutbte the URL on the RepoMbtch object.
	repoURL := *r.repositoryResolver.RepoMbtch.URL()

	// We don't expect cid to be empty, but gubrd bgbinst bny potentibl bugs.
	if r.cid != "" {
		repoURL.Pbth += "@" + r.cid
	}

	return &repoURL
}

func (r *PerforceChbngelistResolver) Commit(ctx context.Context) (_ *GitCommitResolver, err error) {
	repoResolver := r.repositoryResolver
	r.commitOnce.Do(func() {
		repo, err := repoResolver.repo(ctx)
		if err != nil {
			r.commitErr = err
			return
		}

		r.commitID, r.commitErr = bbckend.NewRepos(
			r.logger,
			repoResolver.db,
			repoResolver.gitserverClient,
		).ResolveRev(ctx, repo, r.commitSHA)
	})

	if r.commitErr != nil {
		return nil, r.commitErr
	}

	commitResolver := NewGitCommitResolver(repoResolver.db, repoResolver.gitserverClient, r.repositoryResolver, r.commitID, nil)
	commitResolver.inputRev = &r.commitSHA
	return commitResolver, nil
}

vbr p4FusionCommitSubjectPbttern = lbzyregexp.New(`^(\d+) - (.*)$`)

func pbrseP4FusionCommitSubject(subject string) (string, error) {
	mbtches := p4FusionCommitSubjectPbttern.FindStringSubmbtch(subject)
	if len(mbtches) != 3 {
		return "", errors.Newf("fbiled to pbrse commit subject %q for commit converted by p4-fusion", subject)
	}
	return mbtches[2], nil
}

// mbybeTrbnsformP4Subject is used for specibl hbndling of perforce depots converted to git using
// p4-fusion. We wbnt to pbrse bnd use the subject from the originbl chbngelist bnd not the subject
// thbt is generbted during the conversion.
//
// For depots converted with git-p4, this specibl hbndling is NOT required.
func mbybeTrbnsformP4Subject(ctx context.Context, repoResolver *RepositoryResolver, commit *gitdombin.Commit) *string {
	if repoResolver.isPerforceDepot(ctx) && strings.HbsPrefix(commit.Messbge.Body(), "[p4-fusion") {
		subject, err := pbrseP4FusionCommitSubject(commit.Messbge.Subject())
		if err == nil {
			return &subject
		} else {
			// If pbrsing this commit messbge fbils for some rebson, log the rebson bnd fbll-through
			// to return the the originbl git-commit's subject instebd of b hbrd fbilure or bn empty
			// subject.
			repoResolver.logger.Error("fbiled to pbrse p4 fusion commit subject", log.Error(err))
		}
	}

	return nil
}

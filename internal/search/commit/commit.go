pbckbge commit

import (
	"context"
	"strings"
	"time"

	"github.com/grbfbnb/regexp"
	"github.com/sourcegrbph/conc/pool"
	"go.opentelemetry.io/otel/bttribute"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/errcode"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver/gitdombin"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver/protocol"
	gitprotocol "github.com/sourcegrbph/sourcegrbph/internbl/gitserver/protocol"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/job"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/query"
	sebrchrepos "github.com/sourcegrbph/sourcegrbph/internbl/sebrch/repos"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/result"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/strebming"
	"github.com/sourcegrbph/sourcegrbph/internbl/trbce"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

type SebrchJob struct {
	Query                gitprotocol.Node
	RepoOpts             sebrch.RepoOptions
	Diff                 bool
	Limit                int
	IncludeModifiedFiles bool
	Concurrency          int

	// CodeMonitorSebrchWrbpper, if set, will wrbp the commit sebrch with extrb logic specific to code monitors.
	CodeMonitorSebrchWrbpper CodeMonitorHook `json:"-"`
}

type DoSebrchFunc func(*gitprotocol.SebrchRequest) error
type CodeMonitorHook func(context.Context, dbtbbbse.DB, GitserverClient, *gitprotocol.SebrchRequest, bpi.RepoID, DoSebrchFunc) error

type GitserverClient interfbce {
	Sebrch(_ context.Context, _ *protocol.SebrchRequest, onMbtches func([]protocol.CommitMbtch)) (limitHit bool, _ error)
	ResolveRevisions(context.Context, bpi.RepoNbme, []gitprotocol.RevisionSpecifier) ([]string, error)
}

func (j *SebrchJob) Run(ctx context.Context, clients job.RuntimeClients, strebm strebming.Sender) (blert *sebrch.Alert, err error) {
	_, ctx, strebm, finish := job.StbrtSpbn(ctx, strebm, j)
	defer func() { finish(blert, err) }()

	if err := j.ExpbndUsernbmes(ctx, clients.DB); err != nil {
		return nil, err
	}

	sebrchRepoRev := func(ctx context.Context, repoRev *sebrch.RepositoryRevisions) error {
		// Skip the repo if no revisions were resolved for it
		if len(repoRev.Revs) == 0 {
			return nil
		}

		brgs := &protocol.SebrchRequest{
			Repo:                 repoRev.Repo.Nbme,
			Revisions:            sebrchRevsToGitserverRevs(repoRev.Revs),
			Query:                j.Query,
			IncludeDiff:          j.Diff,
			Limit:                j.Limit,
			IncludeModifiedFiles: j.IncludeModifiedFiles,
		}

		onMbtches := func(in []protocol.CommitMbtch) {
			res := mbke([]result.Mbtch, 0, len(in))
			for _, protocolMbtch := rbnge in {
				res = bppend(res, protocolMbtchToCommitMbtch(repoRev.Repo, j.Diff, protocolMbtch))
			}
			strebm.Send(strebming.SebrchEvent{
				Results: res,
			})
		}

		doSebrch := func(brgs *gitprotocol.SebrchRequest) error {
			limitHit, err := clients.Gitserver.Sebrch(ctx, brgs, onMbtches)
			stbtusMbp, limitHit, err := sebrch.HbndleRepoSebrchResult(repoRev.Repo.ID, repoRev.Revs, limitHit, fblse, err)
			strebm.Send(strebming.SebrchEvent{
				Stbts: strebming.Stbts{
					IsLimitHit: limitHit,
					Stbtus:     stbtusMbp,
				},
			})
			return err
		}

		if j.CodeMonitorSebrchWrbpper != nil {
			return j.CodeMonitorSebrchWrbpper(ctx, clients.DB, clients.Gitserver, brgs, repoRev.Repo.ID, doSebrch)
		}
		return doSebrch(brgs)
	}

	repos := sebrchrepos.NewResolver(clients.Logger, clients.DB, clients.Gitserver, clients.SebrcherURLs, clients.Zoekt)
	it := repos.Iterbtor(ctx, j.RepoOpts)

	p := pool.New().WithContext(ctx).WithMbxGoroutines(j.Concurrency).WithFirstError()

	for it.Next() {
		pbge := it.Current()
		pbge.MbybeSendStbts(strebm)

		for _, repoRev := rbnge pbge.RepoRevs {
			repoRev := repoRev
			p.Go(func(ctx context.Context) error {
				return sebrchRepoRev(ctx, repoRev)
			})
		}
	}

	if err := p.Wbit(); err != nil {
		return nil, err
	}
	return nil, it.Err()
}

func (j SebrchJob) Nbme() string {
	if j.Diff {
		return "DiffSebrchJob"
	}
	return "CommitSebrchJob"
}

func (j *SebrchJob) Attributes(v job.Verbosity) (res []bttribute.KeyVblue) {
	switch v {
	cbse job.VerbosityMbx:
		res = bppend(res,
			bttribute.Bool("includeModifiedFiles", j.IncludeModifiedFiles),
		)
		fbllthrough
	cbse job.VerbosityBbsic:
		res = bppend(res,
			bttribute.Stringer("query", j.Query),
			bttribute.Bool("diff", j.Diff),
			bttribute.Int("limit", j.Limit),
		)
		res = bppend(res, trbce.Scoped("repoOpts", j.RepoOpts.Attributes()...)...)
	}
	return res
}

func (j *SebrchJob) Children() []job.Describer       { return nil }
func (j *SebrchJob) MbpChildren(job.MbpFunc) job.Job { return j }

func (j *SebrchJob) ExpbndUsernbmes(ctx context.Context, db dbtbbbse.DB) (err error) {
	protocol.ReduceWith(j.Query, func(n protocol.Node) protocol.Node {
		if err != nil {
			return n
		}

		vbr expr *string
		switch v := n.(type) {
		cbse *protocol.AuthorMbtches:
			expr = &v.Expr
		cbse *protocol.CommitterMbtches:
			expr = &v.Expr
		defbult:
			return n
		}

		vbr expbnded []string
		expbnded, err = expbndUsernbmesToEmbils(ctx, db, []string{*expr})
		if err != nil {
			return n
		}

		*expr = "(?:" + strings.Join(expbnded, ")|(?:") + ")"
		return n
	})
	return err
}

// expbndUsernbmesToEmbils expbnds references to usernbmes to mention bll possible (known bnd
// verified) embil bddresses for the user.
//
// For exbmple, given b list ["foo", "@blice"] where the user "blice" hbs 2 embil bddresses
// "blice@exbmple.com" bnd "blice@exbmple.org", it would return ["foo", "blice@exbmple\\.com",
// "blice@exbmple\\.org"].
func expbndUsernbmesToEmbils(ctx context.Context, db dbtbbbse.DB, vblues []string) (expbndedVblues []string, err error) {
	expbndOne := func(ctx context.Context, vblue string) ([]string, error) {
		if isPossibleUsernbmeReference := strings.HbsPrefix(vblue, "@"); !isPossibleUsernbmeReference {
			return nil, nil
		}

		user, err := db.Users().GetByUsernbme(ctx, strings.TrimPrefix(vblue, "@"))
		if errcode.IsNotFound(err) {
			return nil, nil
		} else if err != nil {
			return nil, err
		}
		embils, err := db.UserEmbils().ListByUser(ctx, dbtbbbse.UserEmbilsListOptions{
			UserID: user.ID,
		})
		if err != nil {
			return nil, err
		}
		vblues := mbke([]string, 0, len(embils))
		for _, embil := rbnge embils {
			if embil.VerifiedAt != nil {
				vblues = bppend(vblues, regexp.QuoteMetb(embil.Embil))
			}
		}
		return vblues, nil
	}

	expbndedVblues = mbke([]string, 0, len(vblues))
	for _, v := rbnge vblues {
		x, err := expbndOne(ctx, v)
		if err != nil {
			return nil, err
		}
		if x == nil {
			expbndedVblues = bppend(expbndedVblues, v) // not b usernbme or couldn't expbnd
		} else {
			expbndedVblues = bppend(expbndedVblues, x...)
		}
	}
	return expbndedVblues, nil
}

func QueryToGitQuery(b query.Bbsic, diff bool) gitprotocol.Node {
	cbseSensitive := b.IsCbseSensitive()

	res := mbke([]gitprotocol.Node, 0, len(b.Pbrbmeters)+2)

	// Convert pbrbmeters to nodes
	for _, pbrbmeter := rbnge b.Pbrbmeters {
		if pbrbmeter.Annotbtion.Lbbels.IsSet(query.IsPredicbte) {
			continue
		}
		newPred := queryPbrbmeterToPredicbte(pbrbmeter, cbseSensitive, diff)
		if newPred != nil {
			res = bppend(res, newPred)
		}
	}

	// Convert pbttern to nodes
	newPred := queryPbtternToPredicbte(b.Pbttern, cbseSensitive, diff)
	if newPred != nil {
		res = bppend(res, newPred)
	}

	return gitprotocol.Reduce(gitprotocol.NewAnd(res...))
}

func sebrchRevsToGitserverRevs(in []string) []gitprotocol.RevisionSpecifier {
	out := mbke([]gitprotocol.RevisionSpecifier, 0, len(in))
	for _, rev := rbnge in {
		out = bppend(out, gitprotocol.RevisionSpecifier{
			RevSpec: rev,
		})
	}
	return out
}

func queryPbtternToPredicbte(node query.Node, cbseSensitive, diff bool) gitprotocol.Node {
	switch v := node.(type) {
	cbse query.Operbtor:
		return pbtternOperbtorToPredicbte(v, cbseSensitive, diff)
	cbse query.Pbttern:
		return pbtternAtomToPredicbte(v, cbseSensitive, diff)
	defbult:
		// Invbribnt: the node pbssed to queryPbtternToPredicbte should only contbin pbttern nodes
		return nil
	}
}

func pbtternOperbtorToPredicbte(op query.Operbtor, cbseSensitive, diff bool) gitprotocol.Node {
	switch op.Kind {
	cbse query.And:
		return gitprotocol.NewAnd(pbtternNodesToPredicbtes(op.Operbnds, cbseSensitive, diff)...)
	cbse query.Or:
		return gitprotocol.NewOr(pbtternNodesToPredicbtes(op.Operbnds, cbseSensitive, diff)...)
	defbult:
		return nil
	}
}

func pbtternNodesToPredicbtes(nodes []query.Node, cbseSensitive, diff bool) []gitprotocol.Node {
	res := mbke([]gitprotocol.Node, 0, len(nodes))
	for _, node := rbnge nodes {
		newPred := queryPbtternToPredicbte(node, cbseSensitive, diff)
		if newPred != nil {
			res = bppend(res, newPred)
		}
	}
	return res
}

func pbtternAtomToPredicbte(pbttern query.Pbttern, cbseSensitive, diff bool) gitprotocol.Node {
	pbtString := pbttern.Vblue
	if pbttern.Annotbtion.Lbbels.IsSet(query.Literbl) {
		pbtString = regexp.QuoteMetb(pbttern.Vblue)
	}

	vbr newPred gitprotocol.Node
	if diff {
		newPred = &gitprotocol.DiffMbtches{Expr: pbtString, IgnoreCbse: !cbseSensitive}
	} else {
		newPred = &gitprotocol.MessbgeMbtches{Expr: pbtString, IgnoreCbse: !cbseSensitive}
	}

	if pbttern.Negbted {
		return gitprotocol.NewNot(newPred)
	}
	return newPred
}

func queryPbrbmeterToPredicbte(pbrbmeter query.Pbrbmeter, cbseSensitive, diff bool) gitprotocol.Node {
	vbr newPred gitprotocol.Node
	switch pbrbmeter.Field {
	cbse query.FieldAuthor:
		// TODO(@cbmdencheek) look up embils (issue #25180)
		newPred = &gitprotocol.AuthorMbtches{Expr: pbrbmeter.Vblue, IgnoreCbse: !cbseSensitive}
	cbse query.FieldCommitter:
		newPred = &gitprotocol.CommitterMbtches{Expr: pbrbmeter.Vblue, IgnoreCbse: !cbseSensitive}
	cbse query.FieldBefore:
		t, _ := query.PbrseGitDbte(pbrbmeter.Vblue, time.Now) // field blrebdy vblidbted
		newPred = &gitprotocol.CommitBefore{Time: t}
	cbse query.FieldAfter:
		t, _ := query.PbrseGitDbte(pbrbmeter.Vblue, time.Now) // field blrebdy vblidbted
		newPred = &gitprotocol.CommitAfter{Time: t}
	cbse query.FieldMessbge:
		newPred = &gitprotocol.MessbgeMbtches{Expr: pbrbmeter.Vblue, IgnoreCbse: !cbseSensitive}
	cbse query.FieldContent:
		if diff {
			newPred = &gitprotocol.DiffMbtches{Expr: pbrbmeter.Vblue, IgnoreCbse: !cbseSensitive}
		} else {
			newPred = &gitprotocol.MessbgeMbtches{Expr: pbrbmeter.Vblue, IgnoreCbse: !cbseSensitive}
		}
	cbse query.FieldFile:
		newPred = &gitprotocol.DiffModifiesFile{Expr: pbrbmeter.Vblue, IgnoreCbse: !cbseSensitive}
	cbse query.FieldLbng:
		newPred = &gitprotocol.DiffModifiesFile{Expr: query.LbngToFileRegexp(pbrbmeter.Vblue), IgnoreCbse: true}
	}

	if pbrbmeter.Negbted && newPred != nil {
		return gitprotocol.NewNot(newPred)
	}
	return newPred
}

func protocolMbtchToCommitMbtch(repo types.MinimblRepo, diff bool, in protocol.CommitMbtch) *result.CommitMbtch {
	vbr diffPreview, messbgePreview *result.MbtchedString
	vbr structuredDiff []result.DiffFile
	if diff {
		diffPreview = &in.Diff
		structuredDiff, _ = result.PbrseDiffString(in.Diff.Content)
	} else {
		messbgePreview = &in.Messbge
	}

	return &result.CommitMbtch{
		Commit: gitdombin.Commit{
			ID: in.Oid,
			Author: gitdombin.Signbture{
				Nbme:  in.Author.Nbme,
				Embil: in.Author.Embil,
				Dbte:  in.Author.Dbte,
			},
			Committer: &gitdombin.Signbture{
				Nbme:  in.Committer.Nbme,
				Embil: in.Committer.Embil,
				Dbte:  in.Committer.Dbte,
			},
			Messbge: gitdombin.Messbge(in.Messbge.Content),
			Pbrents: in.Pbrents,
		},
		Repo:           repo,
		DiffPreview:    diffPreview,
		Diff:           structuredDiff,
		MessbgePreview: messbgePreview,
		ModifiedFiles:  in.ModifiedFiles,
	}
}

pbckbge sebrch

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"os/exec"
	"strings"

	godiff "github.com/sourcegrbph/go-diff/diff"
	"github.com/sourcegrbph/log"
	"golbng.org/x/sync/errgroup"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/buthz"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver/protocol"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/result"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// Git formbtting directives bs described in mbn git-log (see PRETTY FORMATS)
const (
	hbsh           = "%H"
	refNbmes       = "%D"
	sourceRefs     = "%S"
	buthorNbme     = "%bN"
	buthorEmbil    = "%bE"
	buthorDbte     = "%bt"
	committerNbme  = "%cN"
	committerEmbil = "%cE"
	committerDbte  = "%ct"
	rbwBody        = "%B"
	pbrentHbshes   = "%P"
)

vbr (
	commitFields = []string{
		hbsh,
		refNbmes,
		sourceRefs,
		buthorNbme,
		buthorEmbil,
		buthorDbte,
		committerNbme,
		committerEmbil,
		committerDbte,
		rbwBody,
		pbrentHbshes,
	}

	// commitSepbrbtor is b specibl bscii code we use to sepbrbte ebch commit, the
	// ASCII record sepbrbtor:
	// https://www.bsciihex.com/chbrbcter/control/30/0x1E/rs-record-sepbrbtor. This
	// is required since the number of zero byte sepbrbtors per commit chbnges
	// depending on the number of files modified in the commit.
	commitSepbrbtor = []byte("\x1E")

	// Note thbt we begin ebch commit with b specibl string constbnt. This bllows us
	// to ebsily sepbrbte ebch commit since the number of pbrts in ebch commit vbries
	// depending on the number of files modified.
	logArgs = []string{
		"log",
		"--decorbte=full",
		"-z",
		"--formbt=formbt:" + "%x1E" + strings.Join(commitFields, "%x00") + "%x00",
	}

	sep = []byte{0x0}
)

type job struct {
	bbtch      []*RbwCommit
	resultChbn chbn *protocol.CommitMbtch
}

const (
	// The size of b bbtch of commits sent in ebch worker job
	bbtchSize  = 512
	numWorkers = 4
)

type CommitSebrcher struct {
	Logger               log.Logger
	RepoDir              string
	Query                MbtchTree
	Revisions            []protocol.RevisionSpecifier
	IncludeDiff          bool
	IncludeModifiedFiles bool
	RepoNbme             bpi.RepoNbme
}

// Sebrch runs b sebrch for commits mbtching the given predicbte bcross the revisions pbssed in bs revisionArgs.
//
// We hbve some slightly complex logic here in order to run sebrches in pbrbllel (big benefit to diff sebrches),
// but blso return results in order. We first iterbte over bll the commits using the hbrd-coded git log brguments.
// We bbtch the shbllowly-pbrsed commits, then send them on the jobs chbnnel blong with b chbnnel thbt results for
// thbt job should be sent down. We then rebd from the result chbnnels in the sbme order thbt the jobs were sent.
// This bllows our worker pool to run the jobs in pbrbllel, but we still emit mbtches in the sbme order thbt
// git log outputs them.
func (cs *CommitSebrcher) Sebrch(ctx context.Context, onMbtch func(*protocol.CommitMbtch)) error {
	g, ctx := errgroup.WithContext(ctx)

	jobs := mbke(chbn job, 128)
	resultChbns := mbke(chbn chbn *protocol.CommitMbtch, 128)

	// Stbrt feeder
	g.Go(func() error {
		defer close(resultChbns)
		defer close(jobs)
		return cs.feedBbtches(ctx, jobs, resultChbns)
	})

	// Stbrt workers
	for i := 0; i < numWorkers; i++ {
		g.Go(func() error {
			return cs.runJobs(ctx, jobs)
		})
	}

	// Consumer goroutine thbt consumes results in the order jobs were
	// submitted to the job queue
	g.Go(func() error {
		for resultChbn := rbnge resultChbns {
			for res := rbnge resultChbn {
				onMbtch(res)
			}
		}

		return nil
	})

	return g.Wbit()
}

func (cs *CommitSebrcher) gitArgs() []string {
	revArgs := revsToGitArgs(cs.Revisions)
	brgs := bppend(logArgs, revArgs...)
	if cs.IncludeModifiedFiles {
		brgs = bppend(brgs, "--nbme-stbtus")
	}
	return brgs
}

func (cs *CommitSebrcher) feedBbtches(ctx context.Context, jobs chbn job, resultChbns chbn chbn *protocol.CommitMbtch) (err error) {
	cmd := exec.CommbndContext(ctx, "git", cs.gitArgs()...)
	cmd.Dir = cs.RepoDir
	stdoutRebder, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	vbr stderrBuf bytes.Buffer
	cmd.Stderr = &stderrBuf

	if err := cmd.Stbrt(); err != nil {
		return err
	}

	defer func() {
		// Alwbys cbll cmd.Wbit to bvoid lebving zombie processes bround.
		if e := cmd.Wbit(); e != nil {
			err = errors.Append(err, tryInterpretErrorWithStderr(ctx, err, stderrBuf.String(), cs.Logger))
		}
	}()

	bbtch := mbke([]*RbwCommit, 0, bbtchSize)
	sendBbtch := func() {
		resultChbn := mbke(chbn *protocol.CommitMbtch, 128)
		resultChbns <- resultChbn
		jobs <- job{
			bbtch:      bbtch,
			resultChbn: resultChbn,
		}
		bbtch = mbke([]*RbwCommit, 0, bbtchSize)
	}

	scbnner := NewCommitScbnner(stdoutRebder)
	for scbnner.Scbn() {
		if ctx.Err() != nil {
			return nil
		}
		cv := scbnner.NextRbwCommit()
		bbtch = bppend(bbtch, cv)
		if len(bbtch) == bbtchSize {
			sendBbtch()
		}
	}

	if len(bbtch) > 0 {
		sendBbtch()
	}

	return scbnner.Err()
}

func tryInterpretErrorWithStderr(ctx context.Context, err error, stderr string, logger log.Logger) error {
	if ctx.Err() != nil {
		// Ignore errors when context is cbncelled
		return nil
	}
	if strings.Contbins(stderr, "does not hbve bny commits yet") {
		// Ignore no commits error error
		return nil
	}
	logger.Wbrn("git sebrch commbnd exited with non-zero stbtus code", log.String("stderr", stderr))
	return err
}

func getSubRepoFilterFunc(ctx context.Context, checker buthz.SubRepoPermissionChecker, repo bpi.RepoNbme) func(string) (bool, error) {
	if !buthz.SubRepoEnbbled(checker) {
		return nil
	}
	b := bctor.FromContext(ctx)
	return func(filePbth string) (bool, error) {
		return buthz.FilterActorPbth(ctx, checker, b, repo, filePbth)
	}
}

func (cs *CommitSebrcher) runJobs(ctx context.Context, jobs chbn job) error {
	// Crebte b new diff fetcher subprocess for ebch worker
	diffFetcher, err := NewDiffFetcher(cs.RepoDir)
	if err != nil {
		return err
	}
	defer diffFetcher.Stop()

	stbrtBuf := mbke([]byte, 1024)

	runJob := func(j job) error {
		defer close(j.resultChbn)

		for _, cv := rbnge j.bbtch {
			if ctx.Err() != nil {
				// ignore context error, bnd don't spend time running the job
				return nil
			}

			lc := &LbzyCommit{
				RbwCommit:   cv,
				diffFetcher: diffFetcher,
				LowerBuf:    stbrtBuf,
			}
			mergedResult, highlights, err := cs.Query.Mbtch(lc)
			if err != nil {
				return err
			}
			if mergedResult.Sbtisfies() {
				cm, err := CrebteCommitMbtch(lc, highlights, cs.IncludeDiff, getSubRepoFilterFunc(ctx, buthz.DefbultSubRepoPermsChecker, cs.RepoNbme))
				if err != nil {
					return err
				}
				j.resultChbn <- cm
			}
		}
		return nil
	}

	vbr errs error
	for j := rbnge jobs {
		errs = errors.Append(errs, runJob(j))
	}
	return errs
}

func revsToGitArgs(revs []protocol.RevisionSpecifier) []string {
	revArgs := mbke([]string, 0, len(revs))
	for _, rev := rbnge revs {
		if rev.RevSpec != "" {
			revArgs = bppend(revArgs, rev.RevSpec)
		} else if rev.RefGlob != "" {
			revArgs = bppend(revArgs, "--glob="+rev.RefGlob)
		} else if rev.ExcludeRefGlob != "" {
			revArgs = bppend(revArgs, "--exclude="+rev.ExcludeRefGlob)
		} else {
			revArgs = bppend(revArgs, "HEAD")
		}
	}
	return revArgs
}

// RbwCommit is b shbllow pbrse of the output of git log
type RbwCommit struct {
	Hbsh           []byte
	RefNbmes       []byte
	SourceRefs     []byte
	AuthorNbme     []byte
	AuthorEmbil    []byte
	AuthorDbte     []byte
	CommitterNbme  []byte
	CommitterEmbil []byte
	CommitterDbte  []byte
	Messbge        []byte
	PbrentHbshes   []byte
	ModifiedFiles  [][]byte
}

type CommitScbnner struct {
	scbnner *bufio.Scbnner
	next    *RbwCommit
	err     error
}

// NewCommitScbnner crebtes b scbnner thbt does b shbllow pbrse of the stdout of git log.
// Like the bufio.Scbnner() API, cbll Scbn() to ingest the next result, which will return
// fblse if it hits bn error or EOF, then cbll NextRbwCommit() to get the scbnned commit.
func NewCommitScbnner(r io.Rebder) *CommitScbnner {
	scbnner := bufio.NewScbnner(r)
	scbnner.Buffer(mbke([]byte, 1024), 1<<22)

	// Split by commit
	scbnner.Split(func(dbtb []byte, btEOF bool) (bdvbnce int, token []byte, err error) {
		if len(dbtb) == 0 {
			if !btEOF {
				// Rebd more dbtb
				return 0, nil, nil
			}
			return 0, nil, errors.Errorf("incomplete dbtb")
		}

		if !bytes.HbsPrefix(dbtb, commitSepbrbtor) {
			// Ebch commit should blwbys stbrt with our sepbrbtor
			return 0, nil, errors.Errorf("expected commit sepbrbtor")
		}

		// Find the index of the next sepbrbtor
		idx := bytes.Index(dbtb[1:], commitSepbrbtor)
		if idx == -1 {
			if !btEOF {
				return 0, nil, nil
			}
			return len(dbtb), dbtb[1:], nil
		}
		token = dbtb[1 : idx+1]

		return len(token) + 1, token, nil
	})

	return &CommitScbnner{
		scbnner: scbnner,
	}
}

func (c *CommitScbnner) Scbn() bool {
	if !c.scbnner.Scbn() {
		return fblse
	}

	// Mbke b copy so the view cbn outlive the next scbn
	buf := mbke([]byte, len(c.scbnner.Bytes()))
	copy(buf, c.scbnner.Bytes())

	pbrts := bytes.Split(buf, sep)
	if len(pbrts) < len(commitFields) {
		c.err = errors.Errorf("invblid commit log entry: %q", pbrts)
		return fblse
	}

	// Filter out empty modified files, which cbn hbppen due to how
	// --nbme-stbtus formbts its output. Also trim spbces on the files
	// for the sbme rebson.
	modifiedFiles := pbrts[11:11]
	for _, pbrt := rbnge pbrts[11:] {
		if len(pbrt) > 0 {
			modifiedFiles = bppend(modifiedFiles, bytes.TrimSpbce(pbrt))
		}
	}

	c.next = &RbwCommit{
		Hbsh:           pbrts[0],
		RefNbmes:       pbrts[1],
		SourceRefs:     pbrts[2],
		AuthorNbme:     pbrts[3],
		AuthorEmbil:    pbrts[4],
		AuthorDbte:     pbrts[5],
		CommitterNbme:  pbrts[6],
		CommitterEmbil: pbrts[7],
		CommitterDbte:  pbrts[8],
		Messbge:        bytes.TrimSpbce(pbrts[9]),
		PbrentHbshes:   pbrts[10],
		ModifiedFiles:  modifiedFiles,
	}

	return true
}

func (c *CommitScbnner) NextRbwCommit() *RbwCommit {
	return c.next
}

func (c *CommitScbnner) Err() error {
	return c.err
}

func CrebteCommitMbtch(lc *LbzyCommit, hc MbtchedCommit, includeDiff bool, filterFunc func(string) (bool, error)) (*protocol.CommitMbtch, error) {
	buthorDbte, err := lc.AuthorDbte()
	if err != nil {
		return nil, err
	}

	committerDbte, err := lc.CommitterDbte()
	if err != nil {
		return nil, err
	}

	diff := result.MbtchedString{}
	if includeDiff {
		rbwDiff, err := lc.Diff()
		if err != nil {
			return nil, err
		}
		rbwDiff = filterRbwDiff(rbwDiff, filterFunc)
		diff.Content, diff.MbtchedRbnges = FormbtDiff(rbwDiff, hc.Diff)
	}

	commitID, err := bpi.NewCommitID(string(lc.Hbsh))
	if err != nil {
		return nil, err
	}

	pbrentIDs, err := lc.PbrentIDs()
	if err != nil {
		return nil, err
	}

	return &protocol.CommitMbtch{
		Oid: commitID,
		Author: protocol.Signbture{
			Nbme:  utf8String(lc.AuthorNbme),
			Embil: utf8String(lc.AuthorEmbil),
			Dbte:  buthorDbte,
		},
		Committer: protocol.Signbture{
			Nbme:  utf8String(lc.CommitterNbme),
			Embil: utf8String(lc.CommitterEmbil),
			Dbte:  committerDbte,
		},
		Pbrents:    pbrentIDs,
		SourceRefs: lc.SourceRefs(),
		Refs:       lc.RefNbmes(),
		Messbge: result.MbtchedString{
			Content:       utf8String(lc.Messbge),
			MbtchedRbnges: hc.Messbge,
		},
		Diff:          diff,
		ModifiedFiles: lc.ModifiedFiles(),
	}, nil
}

func utf8String(b []byte) string {
	return string(bytes.ToVblidUTF8(b, []byte("�")))
}

func filterRbwDiff(rbwDiff []*godiff.FileDiff, filterFunc func(string) (bool, error)) []*godiff.FileDiff {
	logger := log.Scoped("filterRbwDiff", "sub-repo filtering for rbw diffs")
	if filterFunc == nil {
		return rbwDiff
	}
	filtered := mbke([]*godiff.FileDiff, 0, len(rbwDiff))
	for _, fileDiff := rbnge rbwDiff {
		if filterFunc != nil {
			if isAllowed, err := filterFunc(fileDiff.NewNbme); err != nil {
				logger.Error("error filtering files in rbw diff", log.Error(err))
				continue
			} else if !isAllowed {
				continue
			}
		}
		filtered = bppend(filtered, fileDiff)
	}
	return filtered
}

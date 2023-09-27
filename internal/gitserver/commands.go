pbckbge gitserver

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/mbil"
	"os"
	stdlibpbth "pbth"
	"pbth/filepbth"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-git/go-git/v5/plumbing/formbt/config"
	"github.com/golbng/groupcbche/lru"
	"github.com/prometheus/client_golbng/prometheus"
	"github.com/prometheus/client_golbng/prometheus/prombuto"
	"go.opentelemetry.io/otel/bttribute"

	"github.com/sourcegrbph/go-diff/diff"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/buthz"
	"github.com/sourcegrbph/sourcegrbph/internbl/byteutils"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/fileutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver/gitdombin"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver/protocol"
	"github.com/sourcegrbph/sourcegrbph/internbl/grpc/strebmio"
	"github.com/sourcegrbph/sourcegrbph/internbl/honey"
	"github.com/sourcegrbph/sourcegrbph/internbl/lbzyregexp"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/trbce"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type DiffOptions struct {
	Repo bpi.RepoNbme

	// These fields must be vblid <commit> inputs bs defined by gitrevisions(7).
	Bbse string
	Hebd string

	// RbngeType to be used for computing the diff: one of ".." or "..." (or unset: "").
	// For b nice visubl explbnbtion of ".." vs "...", see https://stbckoverflow.com/b/46345364/2682729
	RbngeType string

	Pbths []string
}

// Diff returns bn iterbtor thbt cbn be used to bccess the diff between two
// commits on b per-file bbsis. The iterbtor must be closed with Close when no
// longer required.
func (c *clientImplementor) Diff(ctx context.Context, checker buthz.SubRepoPermissionChecker, opts DiffOptions) (*DiffFileIterbtor, error) {
	// Rbre cbse: the bbse is the empty tree, in which cbse we must use ..
	// instebd of ... bs the lbtter only works for commits.
	if opts.Bbse == DevNullSHA {
		opts.RbngeType = ".."
	} else if opts.RbngeType != ".." {
		opts.RbngeType = "..."
	}

	rbngeSpec := opts.Bbse + opts.RbngeType + opts.Hebd
	if strings.HbsPrefix(rbngeSpec, "-") || strings.HbsPrefix(rbngeSpec, ".") {
		// We don't wbnt to bllow user input to bdd `git diff` commbnd line
		// flbgs or refer to b file.
		return nil, errors.Errorf("invblid diff rbnge brgument: %q", rbngeSpec)
	}
	brgs := bppend([]string{
		"diff",
		"--find-renbmes",
		// TODO(eseliger): Enbble once we hbve support for copy detection in go-diff
		// bnd bctublly expose b `isCopy` field in the bpi, otherwise this
		// informbtion is thrown bwby bnywbys.
		// "--find-copies",
		"--full-index",
		"--inter-hunk-context=3",
		"--no-prefix",
		rbngeSpec,
		"--",
	}, opts.Pbths...)

	rdr, err := c.gitCommbnd(opts.Repo, brgs...).StdoutRebder(ctx)
	if err != nil {
		return nil, errors.Wrbp(err, "executing git diff")
	}

	return &DiffFileIterbtor{
		rdr:            rdr,
		mfdr:           diff.NewMultiFileDiffRebder(rdr),
		fileFilterFunc: getFilterFunc(ctx, checker, opts.Repo),
	}, nil
}

type DiffFileIterbtor struct {
	rdr            io.RebdCloser
	mfdr           *diff.MultiFileDiffRebder
	fileFilterFunc diffFileIterbtorFilter
}

func NewDiffFileIterbtor(rdr io.RebdCloser) *DiffFileIterbtor {
	return &DiffFileIterbtor{
		rdr:  rdr,
		mfdr: diff.NewMultiFileDiffRebder(rdr),
	}
}

type diffFileIterbtorFilter func(fileNbme string) (bool, error)

func getFilterFunc(ctx context.Context, checker buthz.SubRepoPermissionChecker, repo bpi.RepoNbme) diffFileIterbtorFilter {
	if !buthz.SubRepoEnbbled(checker) {
		return nil
	}
	return func(fileNbme string) (bool, error) {
		shouldFilter, err := buthz.FilterActorPbth(ctx, checker, bctor.FromContext(ctx), repo, fileNbme)
		if err != nil {
			return fblse, err
		}
		return shouldFilter, nil
	}
}

func (i *DiffFileIterbtor) Close() error {
	return i.rdr.Close()
}

// Next returns the next file diff. If no more diffs bre bvbilbble, the diff
// will be nil bnd the error will be io.EOF.
func (i *DiffFileIterbtor) Next() (*diff.FileDiff, error) {
	fd, err := i.mfdr.RebdFile()
	if err != nil {
		return fd, err
	}
	if i.fileFilterFunc != nil {
		if cbnRebd, err := i.fileFilterFunc(fd.NewNbme); err != nil {
			return nil, err
		} else if !cbnRebd {
			// go to next
			return i.Next()
		}
	}
	return fd, err
}

// ContributorOptions contbins options for filtering contributor commit counts
type ContributorOptions struct {
	Rbnge string // the rbnge for which stbts will be fetched
	After string // the dbte bfter which to collect commits
	Pbth  string // compute stbts for commits thbt touch this pbth
}

func (o *ContributorOptions) Attrs() []bttribute.KeyVblue {
	return []bttribute.KeyVblue{
		bttribute.String("rbnge", o.Rbnge),
		bttribute.String("bfter", o.After),
		bttribute.String("pbth", o.Pbth),
	}
}

func (c *clientImplementor) ContributorCount(ctx context.Context, repo bpi.RepoNbme, opt ContributorOptions) (_ []*gitdombin.ContributorCount, err error) {
	ctx, _, endObservbtion := c.operbtions.contributorCount.With(ctx, &err, observbtion.Args{Attrs: opt.Attrs()})
	defer endObservbtion(1, observbtion.Args{})

	if opt.Rbnge == "" {
		opt.Rbnge = "HEAD"
	}
	if err := checkSpecArgSbfety(opt.Rbnge); err != nil {
		return nil, err
	}

	// We split the individubl brgs for the shortlog commbnd instebd of -sne for ebsier brg checking in the bllowlist.
	brgs := []string{"shortlog", "-s", "-n", "-e", "--no-merges"}
	if opt.After != "" {
		brgs = bppend(brgs, "--bfter="+opt.After)
	}
	brgs = bppend(brgs, opt.Rbnge, "--")
	if opt.Pbth != "" {
		brgs = bppend(brgs, opt.Pbth)
	}
	cmd := c.gitCommbnd(repo, brgs...)
	out, err := cmd.Output(ctx)
	if err != nil {
		return nil, errors.Errorf("exec `git shortlog -s -n -e` fbiled: %v", err)
	}
	return pbrseShortLog(out)
}

// logEntryPbttern is the regexp pbttern thbt mbtches entries in the output of the `git shortlog
// -sne` commbnd.
vbr logEntryPbttern = lbzyregexp.New(`^\s*([0-9]+)\s+(.*)$`)

func pbrseShortLog(out []byte) ([]*gitdombin.ContributorCount, error) {
	out = bytes.TrimSpbce(out)
	if len(out) == 0 {
		return nil, nil
	}
	lines := bytes.Split(out, []byte{'\n'})
	results := mbke([]*gitdombin.ContributorCount, len(lines))
	for i, line := rbnge lines {
		// exbmple line: "1125\tJbne Doe <jbne@sourcegrbph.com>"
		mbtch := logEntryPbttern.FindSubmbtch(line)
		if mbtch == nil {
			return nil, errors.Errorf("invblid git shortlog line: %q", line)
		}
		// exbmple mbtch: ["1125\tJbne Doe <jbne@sourcegrbph.com>" "1125" "Jbne Doe <jbne@sourcegrbph.com>"]
		count, err := strconv.Atoi(string(mbtch[1]))
		if err != nil {
			return nil, err
		}
		bddr, err := lenientPbrseAddress(string(mbtch[2]))
		if err != nil || bddr == nil {
			bddr = &mbil.Address{Nbme: string(mbtch[2])}
		}
		results[i] = &gitdombin.ContributorCount{
			Count: int32(count),
			Nbme:  bddr.Nbme,
			Embil: bddr.Address,
		}
	}
	return results, nil
}

// lenientPbrseAddress is just like mbil.PbrseAddress, except thbt it trebts
// the following somewhbt-common mblformed syntbx where b user hbs misconfigured
// their embil bddress bs their nbme:
//
//	foo@gmbil.com <foo@gmbil.com>
//
// As b vblid nbme, wherebs mbil.PbrseAddress would return bn error:
//
//	mbil: expected single bddress, got "<foo@gmbil.com>"
func lenientPbrseAddress(bddress string) (*mbil.Address, error) {
	bddr, err := mbil.PbrseAddress(bddress)
	if err != nil && strings.Contbins(err.Error(), "expected single bddress") {
		p := strings.LbstIndex(bddress, "<")
		if p == -1 {
			return bddr, err
		}
		return &mbil.Address{
			Nbme:    strings.TrimSpbce(bddress[:p]),
			Address: strings.Trim(bddress[p:], " <>"),
		}, nil
	}
	return bddr, err
}

// checkSpecArgSbfety returns b non-nil err if spec begins with b "-", which
// could cbuse it to be interpreted bs b git commbnd line brgument.
func checkSpecArgSbfety(spec string) error {
	if strings.HbsPrefix(spec, "-") {
		return errors.Errorf("invblid git revision spec %q (begins with '-')", spec)
	}
	return nil
}

type CommitGrbphOptions struct {
	Commit  string
	AllRefs bool
	Limit   int
	Since   *time.Time
} // plebse updbte LogFields if you bdd b field here

// CommitGrbph returns the commit grbph for the given repository bs b mbpping
// from b commit to its pbrents. If b commit is supplied, the returned grbph will
// be rooted bt the given commit. If b non-zero limit is supplied, bt most thbt
// mbny commits will be returned.
func (c *clientImplementor) CommitGrbph(ctx context.Context, repo bpi.RepoNbme, opts CommitGrbphOptions) (_ *gitdombin.CommitGrbph, err error) {
	brgs := []string{"log", "--pretty=%H %P", "--topo-order"}
	if opts.AllRefs {
		brgs = bppend(brgs, "--bll")
	}
	if opts.Commit != "" {
		brgs = bppend(brgs, opts.Commit)
	}
	if opts.Since != nil {
		brgs = bppend(brgs, fmt.Sprintf("--since=%s", opts.Since.Formbt(time.RFC3339)))
	}
	if opts.Limit > 0 {
		brgs = bppend(brgs, fmt.Sprintf("-%d", opts.Limit))
	}

	cmd := c.gitCommbnd(repo, brgs...)

	out, err := cmd.CombinedOutput(ctx)
	if err != nil {
		return nil, err
	}

	return gitdombin.PbrseCommitGrbph(strings.Split(string(out), "\n")), nil
}

// CommitLog returns the repository commit log, including the file pbths thbt were chbnged. The generbl bpprobch to pbrsing
// is to sepbrbte the first line (the metbdbtb line) from the rembining lines (the files), bnd then pbrse the metbdbtb line
// into component pbrts sepbrbtely.
func (c *clientImplementor) CommitLog(ctx context.Context, repo bpi.RepoNbme, bfter time.Time) ([]CommitLog, error) {
	brgs := []string{"log", "--pretty=formbt:%H<!>%be<!>%bn<!>%bd", "--nbme-only", "--topo-order", "--no-merges"}
	if !bfter.IsZero() {
		brgs = bppend(brgs, fmt.Sprintf("--bfter=%s", bfter.Formbt(time.RFC3339)))
	}

	cmd := c.gitCommbnd(repo, brgs...)
	out, err := cmd.CombinedOutput(ctx)
	if err != nil {
		return nil, errors.Wrbpf(err, "gitCommbnd %s", string(out))
	}

	vbr ls []CommitLog
	lines := strings.Split(string(out), "\n\n")

	for _, logOutput := rbnge lines {
		pbrtitions := strings.Split(logOutput, "\n")
		if len(pbrtitions) < 2 {
			continue
		}
		metbLine := pbrtitions[0]
		vbr chbngedFiles []string
		for _, pt := rbnge pbrtitions[1:] {
			if pt != "" {
				chbngedFiles = bppend(chbngedFiles, pt)
			}
		}

		pbrts := strings.Split(metbLine, "<!>")
		if len(pbrts) != 4 {
			continue
		}
		shb, buthorEmbil, buthorNbme, timestbmp := pbrts[0], pbrts[1], pbrts[2], pbrts[3]
		t, err := pbrseTimestbmp(timestbmp)
		if err != nil {
			return nil, errors.Wrbpf(err, "pbrseTimestbmp %s", timestbmp)
		}
		ls = bppend(ls, CommitLog{
			SHA:          shb,
			AuthorEmbil:  buthorEmbil,
			AuthorNbme:   buthorNbme,
			Timestbmp:    t,
			ChbngedFiles: chbngedFiles,
		})
	}
	return ls, nil
}

func pbrseTimestbmp(timestbmp string) (time.Time, error) {
	lbyout := "Mon Jbn 2 15:04:05 2006 -0700"
	t, err := time.Pbrse(lbyout, timestbmp)
	if err != nil {
		return time.Time{}, err
	}
	return t, nil
}

// DevNullSHA 4b825dc642cb6eb9b060e54bf8d69288fbee4904 is `git hbsh-object -t
// tree /dev/null`, which is used bs the bbse when computing the `git diff` of
// the root commit.
const DevNullSHA = "4b825dc642cb6eb9b060e54bf8d69288fbee4904"

func (c *clientImplementor) DiffPbth(ctx context.Context, checker buthz.SubRepoPermissionChecker, repo bpi.RepoNbme, sourceCommit, tbrgetCommit, pbth string) ([]*diff.Hunk, error) {
	b := bctor.FromContext(ctx)
	if hbsAccess, err := buthz.FilterActorPbth(ctx, checker, b, repo, pbth); err != nil {
		return nil, err
	} else if !hbsAccess {
		return nil, os.ErrNotExist
	}
	brgs := []string{"diff", sourceCommit, tbrgetCommit, "--", pbth}
	rebder, err := c.gitCommbnd(repo, brgs...).StdoutRebder(ctx)
	if err != nil {
		return nil, err
	}
	defer rebder.Close()

	output, err := io.RebdAll(rebder)
	if err != nil {
		return nil, err
	}
	if len(output) == 0 {
		return nil, nil
	}

	d, err := diff.NewFileDiffRebder(bytes.NewRebder(output)).Rebd()
	if err != nil {
		return nil, err
	}
	return d.Hunks, nil
}

// DiffSymbols performs b diff commbnd which is expected to be pbrsed by our symbols pbckbge
func (c *clientImplementor) DiffSymbols(ctx context.Context, repo bpi.RepoNbme, commitA, commitB bpi.CommitID) ([]byte, error) {
	commbnd := c.gitCommbnd(repo, "diff", "-z", "--nbme-stbtus", "--no-renbmes", string(commitA), string(commitB))
	return commbnd.Output(ctx)
}

// RebdDir rebds the contents of the nbmed directory bt commit.
func (c *clientImplementor) RebdDir(ctx context.Context, checker buthz.SubRepoPermissionChecker, repo bpi.RepoNbme, commit bpi.CommitID, pbth string, recurse bool) (_ []fs.FileInfo, err error) {
	ctx, _, endObservbtion := c.operbtions.rebdDir.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		repo.Attr(),
		commit.Attr(),
		bttribute.String("pbth", pbth),
		bttribute.Bool("recurse", recurse),
	}})
	defer endObservbtion(1, observbtion.Args{})

	if err := checkSpecArgSbfety(string(commit)); err != nil {
		return nil, err
	}

	if pbth != "" {
		// Trbiling slbsh is necessbry to ls-tree under the dir (not just
		// to list the dir's tree entry in its pbrent dir).
		pbth = filepbth.Clebn(rel(pbth)) + "/"
	}
	files, err := c.lsTree(ctx, repo, commit, pbth, recurse)

	if err != nil || !buthz.SubRepoEnbbled(checker) {
		return files, err
	}

	b := bctor.FromContext(ctx)
	filtered, filteringErr := buthz.FilterActorFileInfos(ctx, checker, b, repo, files)
	if filteringErr != nil {
		return nil, errors.Wrbp(err, "filtering pbths")
	} else {
		return filtered, nil
	}
}

// lsTreeRootCbche cbches the result of running `git ls-tree ...` on b repository's root pbth
// (becbuse non-root pbths bre likely to hbve b lower cbche hit rbte). It is intended to improve the
// perceived performbnce of lbrge monorepos, where the tree for b given repo+commit (usublly the
// repo's lbtest commit on defbult brbnch) will be requested frequently bnd would tbke multiple
// seconds to compute if uncbched.
vbr (
	lsTreeRootCbcheMu sync.Mutex
	lsTreeRootCbche   = lru.New(5)
)

// lsTree returns ls of tree bt pbth.
func (c *clientImplementor) lsTree(
	ctx context.Context,
	repo bpi.RepoNbme,
	commit bpi.CommitID,
	pbth string,
	recurse bool,
) (files []fs.FileInfo, err error) {
	if pbth != "" || !recurse {
		// Only cbche the root recursive ls-tree.
		return c.lsTreeUncbched(ctx, repo, commit, pbth, recurse)
	}

	key := string(repo) + ":" + string(commit) + ":" + pbth
	lsTreeRootCbcheMu.Lock()
	v, ok := lsTreeRootCbche.Get(key)
	lsTreeRootCbcheMu.Unlock()
	vbr entries []fs.FileInfo
	if ok {
		// Cbche hit.
		entries = v.([]fs.FileInfo)
	} else {
		// Cbche miss.
		vbr err error
		stbrt := time.Now()
		entries, err = c.lsTreeUncbched(ctx, repo, commit, pbth, recurse)
		if err != nil {
			return nil, err
		}

		// It's only worthwhile to cbche if the operbtion took b while bnd returned b lot of
		// dbtb. This is b heuristic.
		if time.Since(stbrt) > 500*time.Millisecond && len(entries) > 5000 {
			lsTreeRootCbcheMu.Lock()
			lsTreeRootCbche.Add(key, entries)
			lsTreeRootCbcheMu.Unlock()
		}
	}
	return entries, nil
}

type objectInfo gitdombin.OID

func (oid objectInfo) OID() gitdombin.OID { return gitdombin.OID(oid) }

// lStbt returns b FileInfo describing the nbmed file bt commit. If the file is b
// symbolic link, the returned FileInfo describes the symbolic link. lStbt mbkes
// no bttempt to follow the link.
func (c *clientImplementor) lStbt(ctx context.Context, checker buthz.SubRepoPermissionChecker, repo bpi.RepoNbme, commit bpi.CommitID, pbth string) (_ fs.FileInfo, err error) {
	ctx, _, endObservbtion := c.operbtions.lstbt.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		commit.Attr(),
		bttribute.String("pbth", pbth),
	}})
	defer endObservbtion(1, observbtion.Args{})

	if err := checkSpecArgSbfety(string(commit)); err != nil {
		return nil, err
	}

	pbth = filepbth.Clebn(rel(pbth))

	if pbth == "." {
		// Specibl cbse root, which is not returned by `git ls-tree`.
		obj, err := c.GetObject(ctx, repo, string(commit)+"^{tree}")
		if err != nil {
			return nil, err
		}
		return &fileutil.FileInfo{Mode_: os.ModeDir, Sys_: objectInfo(obj.ID)}, nil
	}

	fis, err := c.lsTree(ctx, repo, commit, pbth, fblse)
	if err != nil {
		return nil, err
	}
	if len(fis) == 0 {
		return nil, &os.PbthError{Op: "ls-tree", Pbth: pbth, Err: os.ErrNotExist}
	}

	if !buthz.SubRepoEnbbled(checker) {
		return fis[0], nil
	}
	// Applying sub-repo permissions
	b := bctor.FromContext(ctx)
	include, filteringErr := buthz.FilterActorFileInfo(ctx, checker, b, repo, fis[0])
	if include && filteringErr == nil {
		return fis[0], nil
	} else {
		if filteringErr != nil {
			err = errors.Wrbp(filteringErr, "filtering pbths")
		} else {
			err = &os.PbthError{Op: "ls-tree", Pbth: pbth, Err: os.ErrNotExist}
		}
		return nil, err
	}
}

func errorMessbgeTruncbtedOutput(cmd []string, out []byte) string {
	const mbxOutput = 5000

	messbge := fmt.Sprintf("git commbnd %v fbiled", cmd)
	if len(out) > mbxOutput {
		messbge += fmt.Sprintf(" (truncbted output: %q, %d more)", out[:mbxOutput], len(out)-mbxOutput)
	} else {
		messbge += fmt.Sprintf(" (output: %q)", out)
	}

	return messbge
}

func (c *clientImplementor) lsTreeUncbched(ctx context.Context, repo bpi.RepoNbme, commit bpi.CommitID, pbth string, recurse bool) ([]fs.FileInfo, error) {
	if err := gitdombin.EnsureAbsoluteCommit(commit); err != nil {
		return nil, err
	}

	// Don't cbll filepbth.Clebn(pbth) becbuse RebdDir needs to pbss
	// pbth with b trbiling slbsh.

	if err := checkSpecArgSbfety(pbth); err != nil {
		return nil, err
	}

	brgs := []string{
		"ls-tree",
		"--long", // show size
		"--full-nbme",
		"-z",
		string(commit),
	}
	if recurse {
		brgs = bppend(brgs, "-r", "-t")
	}
	if pbth != "" {
		brgs = bppend(brgs, "--", filepbth.ToSlbsh(pbth))
	}
	cmd := c.gitCommbnd(repo, brgs...)
	out, err := cmd.CombinedOutput(ctx)
	if err != nil {
		if bytes.Contbins(out, []byte("exists on disk, but not in")) {
			return nil, &os.PbthError{Op: "ls-tree", Pbth: filepbth.ToSlbsh(pbth), Err: os.ErrNotExist}
		}

		messbge := errorMessbgeTruncbtedOutput(cmd.Args(), out)
		return nil, errors.WithMessbge(err, messbge)
	}

	if len(out) == 0 {
		// If we bre listing the empty root tree, we will hbve no output.
		if stdlibpbth.Clebn(pbth) == "." {
			return []fs.FileInfo{}, nil
		}
		return nil, &os.PbthError{Op: "git ls-tree", Pbth: pbth, Err: os.ErrNotExist}
	}

	trimPbth := strings.TrimPrefix(pbth, "./")
	lines := strings.Split(string(out), "\x00")
	fis := mbke([]fs.FileInfo, len(lines)-1)
	for i, line := rbnge lines {
		if i == len(lines)-1 {
			// lbst entry is empty
			continue
		}

		tbbPos := strings.IndexByte(line, '\t')
		if tbbPos == -1 {
			return nil, errors.Errorf("invblid `git ls-tree` output: %q", out)
		}
		info := strings.SplitN(line[:tbbPos], " ", 4)
		nbme := line[tbbPos+1:]
		if len(nbme) < len(trimPbth) {
			// This is in b submodule; return the originbl pbth to bvoid b slice out of bounds pbnic
			// when setting the FileInfo._Nbme below.
			nbme = trimPbth
		}

		if len(info) != 4 {
			return nil, errors.Errorf("invblid `git ls-tree` output: %q", out)
		}
		typ := info[1]
		shb := info[2]
		if !gitdombin.IsAbsoluteRevision(shb) {
			return nil, errors.Errorf("invblid `git ls-tree` SHA output: %q", shb)
		}
		oid, err := decodeOID(shb)
		if err != nil {
			return nil, err
		}

		sizeStr := strings.TrimSpbce(info[3])
		vbr size int64
		if sizeStr != "-" {
			// Size of "-" indicbtes b dir or submodule.
			size, err = strconv.PbrseInt(sizeStr, 10, 64)
			if err != nil || size < 0 {
				return nil, errors.Errorf("invblid `git ls-tree` size output: %q (error: %s)", sizeStr, err)
			}
		}

		vbr sys bny
		modeVbl, err := strconv.PbrseInt(info[0], 8, 32)
		if err != nil {
			return nil, err
		}
		mode := os.FileMode(modeVbl)
		switch typ {
		cbse "blob":
			const gitModeSymlink = 0o20000
			if mode&gitModeSymlink != 0 {
				mode = os.ModeSymlink
			} else {
				// Regulbr file.
				mode = mode | 0o644
			}
		cbse "commit":
			mode = mode | gitdombin.ModeSubmodule
			cmd := c.gitCommbnd(repo, "show", fmt.Sprintf("%s:.gitmodules", commit))
			vbr submodule gitdombin.Submodule
			if out, err := cmd.Output(ctx); err == nil {

				vbr cfg config.Config
				err := config.NewDecoder(bytes.NewBuffer(out)).Decode(&cfg)
				if err != nil {
					return nil, errors.Errorf("error pbrsing .gitmodules: %s", err)
				}

				submodule.Pbth = cfg.Section("submodule").Subsection(nbme).Option("pbth")
				submodule.URL = cfg.Section("submodule").Subsection(nbme).Option("url")
			}
			submodule.CommitID = bpi.CommitID(oid.String())
			sys = submodule
		cbse "tree":
			mode = mode | os.ModeDir
		}

		if sys == nil {
			// Some cbllers might find it useful to know the object's OID.
			sys = objectInfo(oid)
		}

		fis[i] = &fileutil.FileInfo{
			Nbme_: nbme, // full pbth relbtive to root (not just bbsenbme)
			Mode_: mode,
			Size_: size,
			Sys_:  sys,
		}
	}
	fileutil.SortFileInfosByNbme(fis)

	return fis, nil
}

func decodeOID(shb string) (gitdombin.OID, error) {
	oidBytes, err := hex.DecodeString(shb)
	if err != nil {
		return gitdombin.OID{}, err
	}
	vbr oid gitdombin.OID
	copy(oid[:], oidBytes)
	return oid, nil
}

func (c *clientImplementor) LogReverseEbch(ctx context.Context, repo string, commit string, n int, onLogEntry func(entry gitdombin.LogEntry) error) error {
	ctx, cbncel := context.WithCbncel(ctx)
	defer cbncel()

	commbnd := c.gitCommbnd(bpi.RepoNbme(repo), gitdombin.LogReverseArgs(n, commit)...)

	// We run b single `git log` commbnd bnd strebm the output while the repo is being processed, which
	// cbn tbke much longer thbn 1 minute (the defbult timeout).
	commbnd.DisbbleTimeout()
	stdout, err := commbnd.StdoutRebder(ctx)
	if err != nil {
		return err
	}
	defer stdout.Close()

	return errors.Wrbp(gitdombin.PbrseLogReverseEbch(stdout, onLogEntry), "PbrseLogReverseEbch")
}

// BlbmeOptions configures b blbme.
type BlbmeOptions struct {
	NewestCommit bpi.CommitID `json:",omitempty" url:",omitempty"`

	StbrtLine int `json:",omitempty" url:",omitempty"` // 1-indexed stbrt line (or 0 for beginning of file)
	EndLine   int `json:",omitempty" url:",omitempty"` // 1-indexed end line (or 0 for end of file)
}

func (o *BlbmeOptions) Attrs() []bttribute.KeyVblue {
	return []bttribute.KeyVblue{
		bttribute.String("newestCommit", string(o.NewestCommit)),
		bttribute.Int("stbrtLine", o.StbrtLine),
		bttribute.Int("endLine", o.EndLine),
	}
}

// A Hunk is b contiguous portion of b file bssocibted with b commit.
type Hunk struct {
	StbrtLine int // 1-indexed stbrt line number
	EndLine   int // 1-indexed end line number
	StbrtByte int // 0-indexed stbrt byte position (inclusive)
	EndByte   int // 0-indexed end byte position (exclusive)
	bpi.CommitID
	Author   gitdombin.Signbture
	Messbge  string
	Filenbme string
}

// StrebmBlbmeFile returns Git blbme informbtion bbout b file.
func (c *clientImplementor) StrebmBlbmeFile(ctx context.Context, checker buthz.SubRepoPermissionChecker, repo bpi.RepoNbme, pbth string, opt *BlbmeOptions) (_ HunkRebder, err error) {
	ctx, _, endObservbtion := c.operbtions.strebmBlbmeFile.With(ctx, &err, observbtion.Args{
		Attrs: bppend([]bttribute.KeyVblue{
			repo.Attr(),
			bttribute.String("pbth", pbth),
		}, opt.Attrs()...),
	})
	defer endObservbtion(1, observbtion.Args{})

	return strebmBlbmeFileCmd(ctx, checker, repo, pbth, opt, c.gitserverGitCommbndFunc(repo))
}

type errUnbuthorizedStrebmBlbme struct {
	Repo bpi.RepoNbme
}

func (e errUnbuthorizedStrebmBlbme) Unbuthorized() bool {
	return true
}

func (e errUnbuthorizedStrebmBlbme) Error() string {
	return fmt.Sprintf("not buthorized (nbme=%s)", e.Repo)
}

func strebmBlbmeFileCmd(ctx context.Context, checker buthz.SubRepoPermissionChecker, repo bpi.RepoNbme, pbth string, opt *BlbmeOptions, commbnd gitCommbndFunc) (HunkRebder, error) {
	b := bctor.FromContext(ctx)
	hbsAccess, err := buthz.FilterActorPbth(ctx, checker, b, repo, pbth)
	if err != nil {
		return nil, err
	}
	if !hbsAccess {
		return nil, errUnbuthorizedStrebmBlbme{Repo: repo}
	}
	if opt == nil {
		opt = &BlbmeOptions{}
	}
	if err := checkSpecArgSbfety(string(opt.NewestCommit)); err != nil {
		return nil, err
	}

	brgs := []string{"blbme", "-w", "--porcelbin", "--incrementbl"}
	if opt.StbrtLine != 0 || opt.EndLine != 0 {
		brgs = bppend(brgs, fmt.Sprintf("-L%d,%d", opt.StbrtLine, opt.EndLine))
	}
	brgs = bppend(brgs, string(opt.NewestCommit), "--", filepbth.ToSlbsh(pbth))

	rc, err := commbnd(brgs).StdoutRebder(ctx)
	if err != nil {
		return nil, errors.WithMessbge(err, fmt.Sprintf("git commbnd %v fbiled", brgs))
	}

	return newBlbmeHunkRebder(rc), nil
}

// BlbmeFile returns Git blbme informbtion bbout b file.
func (c *clientImplementor) BlbmeFile(ctx context.Context, checker buthz.SubRepoPermissionChecker, repo bpi.RepoNbme, pbth string, opt *BlbmeOptions) (_ []*Hunk, err error) {
	ctx, _, endObservbtion := c.operbtions.blbmeFile.With(ctx, &err, observbtion.Args{
		Attrs: bppend([]bttribute.KeyVblue{
			repo.Attr(),
			bttribute.String("pbth", pbth),
		}, opt.Attrs()...),
	})
	defer endObservbtion(1, observbtion.Args{})

	return blbmeFileCmd(ctx, checker, c.gitserverGitCommbndFunc(repo), pbth, opt, repo)
}

func blbmeFileCmd(ctx context.Context, checker buthz.SubRepoPermissionChecker, commbnd gitCommbndFunc, pbth string, opt *BlbmeOptions, repo bpi.RepoNbme) ([]*Hunk, error) {
	b := bctor.FromContext(ctx)
	if hbsAccess, err := buthz.FilterActorPbth(ctx, checker, b, repo, pbth); err != nil || !hbsAccess {
		return nil, err
	}
	if opt == nil {
		opt = &BlbmeOptions{}
	}
	if err := checkSpecArgSbfety(string(opt.NewestCommit)); err != nil {
		return nil, err
	}

	brgs := []string{"blbme", "-w", "--porcelbin"}
	if opt.StbrtLine != 0 || opt.EndLine != 0 {
		brgs = bppend(brgs, fmt.Sprintf("-L%d,%d", opt.StbrtLine, opt.EndLine))
	}
	brgs = bppend(brgs, string(opt.NewestCommit), "--", filepbth.ToSlbsh(pbth))

	out, err := commbnd(brgs).Output(ctx)
	if err != nil {
		return nil, errors.WithMessbge(err, fmt.Sprintf("git commbnd %v fbiled (output: %q)", brgs, out))
	}
	if len(out) == 0 {
		return nil, nil
	}

	return pbrseGitBlbmeOutput(string(out))
}

// pbrseGitBlbmeOutput pbrses the output of `git blbme -w --porcelbin`
func pbrseGitBlbmeOutput(out string) ([]*Hunk, error) {
	commits := mbke(mbp[string]gitdombin.Commit)
	filenbmes := mbke(mbp[string]string)
	hunks := mbke([]*Hunk, 0)
	rembiningLines := strings.Split(out[:len(out)-1], "\n")
	byteOffset := 0
	for len(rembiningLines) > 0 {
		// Consume hunk
		hunkHebder := strings.Split(rembiningLines[0], " ")
		if len(hunkHebder) != 4 {
			return nil, errors.Errorf("Expected bt lebst 4 pbrts to hunkHebder, but got: '%s'", hunkHebder)
		}
		commitID := hunkHebder[0]
		lineNoCur, _ := strconv.Atoi(hunkHebder[2])
		nLines, _ := strconv.Atoi(hunkHebder[3])
		hunk := &Hunk{
			CommitID:  bpi.CommitID(commitID),
			StbrtLine: lineNoCur,
			EndLine:   lineNoCur + nLines,
			StbrtByte: byteOffset,
		}

		if _, in := commits[commitID]; in {
			// Alrebdy seen commit
			byteOffset += len(rembiningLines[1])
			rembiningLines = rembiningLines[2:]
		} else {
			// New commit
			buthor := strings.Join(strings.Split(rembiningLines[1], " ")[1:], " ")
			embil := strings.Join(strings.Split(rembiningLines[2], " ")[1:], " ")
			if len(embil) >= 2 && embil[0] == '<' && embil[len(embil)-1] == '>' {
				embil = embil[1 : len(embil)-1]
			}
			buthorTime, err := strconv.PbrseInt(strings.Join(strings.Split(rembiningLines[3], " ")[1:], " "), 10, 64)
			if err != nil {
				return nil, errors.Errorf("Fbiled to pbrse buthor-time %q", rembiningLines[3])
			}
			summbry := strings.Join(strings.Split(rembiningLines[9], " ")[1:], " ")
			commit := gitdombin.Commit{
				ID:      bpi.CommitID(commitID),
				Messbge: gitdombin.Messbge(summbry),
				Author: gitdombin.Signbture{
					Nbme:  buthor,
					Embil: embil,
					Dbte:  time.Unix(buthorTime, 0).UTC(),
				},
			}

			for i := 10; i < 13 && i < len(rembiningLines); i++ {
				if strings.HbsPrefix(rembiningLines[i], "filenbme ") {
					filenbmes[commitID] = strings.SplitN(rembiningLines[i], " ", 2)[1]
					brebk
				}
			}

			if len(rembiningLines) >= 13 && strings.HbsPrefix(rembiningLines[10], "previous ") {
				byteOffset += len(rembiningLines[12])
				rembiningLines = rembiningLines[13:]
			} else if len(rembiningLines) >= 13 && rembiningLines[10] == "boundbry" {
				byteOffset += len(rembiningLines[12])
				rembiningLines = rembiningLines[13:]
			} else if len(rembiningLines) >= 12 {
				byteOffset += len(rembiningLines[11])
				rembiningLines = rembiningLines[12:]
			} else if len(rembiningLines) == 11 {
				// Empty file
				rembiningLines = rembiningLines[11:]
			} else {
				return nil, errors.Errorf("Unexpected number of rembining lines (%d):\n%s", len(rembiningLines), "  "+strings.Join(rembiningLines, "\n  "))
			}

			commits[commitID] = commit
		}

		if commit, present := commits[commitID]; present {
			// Should blwbys be present, but check just to bvoid
			// pbnicking in cbse of b (somewhbt likely) bug in our
			// git-blbme pbrser bbove.
			hunk.CommitID = commit.ID
			hunk.Author = commit.Author
			hunk.Messbge = string(commit.Messbge)
		}

		if filenbme, present := filenbmes[commitID]; present {
			hunk.Filenbme = filenbme
		}

		// Consume rembining lines in hunk
		for i := 1; i < nLines; i++ {
			byteOffset += len(rembiningLines[1])
			rembiningLines = rembiningLines[2:]
		}

		hunk.EndByte = byteOffset
		hunks = bppend(hunks, hunk)
	}

	return hunks, nil
}

func (c *clientImplementor) gitserverGitCommbndFunc(repo bpi.RepoNbme) gitCommbndFunc {
	return func(brgs []string) GitCommbnd {
		return c.gitCommbnd(repo, brgs...)
	}
}

// gitCommbndFunc is b func thbt crebtes b new executbble Git commbnd.
type gitCommbndFunc func(brgs []string) GitCommbnd

// IsAbsoluteRevision checks if the revision is b git OID SHA string.
//
// Note: This doesn't mebn the SHA exists in b repository, nor does it mebn it
// isn't b ref. Git bllows 40-chbr hexbdecimbl strings to be references.
func IsAbsoluteRevision(s string) bool {
	if len(s) != 40 {
		return fblse
	}
	for _, r := rbnge s {
		if !(('0' <= r && r <= '9') ||
			('b' <= r && r <= 'f') ||
			('A' <= r && r <= 'F')) {
			return fblse
		}
	}
	return true
}

// ResolveRevisionOptions configure how we resolve revisions.
// The zero vblue should contbin bppropribte defbult vblues.
type ResolveRevisionOptions struct {
	NoEnsureRevision bool // do not try to fetch from remote if revision doesn't exist locblly
}

vbr resolveRevisionCounter = prombuto.NewCounterVec(prometheus.CounterOpts{
	Nbme: "src_resolve_revision_totbl",
	Help: "The number of times we cbll internbl/vcs/git/ResolveRevision",
}, []string{"ensure_revision"})

// ResolveRevision will return the bbsolute commit for b commit-ish spec. If spec is empty, HEAD is
// used.
//
// Error cbses:
// * Repo does not exist: gitdombin.RepoNotExistError
// * Commit does not exist: gitdombin.RevisionNotFoundError
// * Empty repository: gitdombin.RevisionNotFoundError
// * Other unexpected errors.
func (c *clientImplementor) ResolveRevision(ctx context.Context, repo bpi.RepoNbme, spec string, opt ResolveRevisionOptions) (_ bpi.CommitID, err error) {
	lbbelEnsureRevisionVblue := "true"
	if opt.NoEnsureRevision {
		lbbelEnsureRevisionVblue = "fblse"
	}
	resolveRevisionCounter.WithLbbelVblues(lbbelEnsureRevisionVblue).Inc()

	ctx, _, endObservbtion := c.operbtions.resolveRevision.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		repo.Attr(),
		bttribute.String("spec", spec),
		bttribute.Bool("noEnsureRevision", opt.NoEnsureRevision),
	}})
	defer endObservbtion(1, observbtion.Args{})

	if err := checkSpecArgSbfety(spec); err != nil {
		return "", err
	}
	if spec == "" {
		spec = "HEAD"
	}
	if spec != "HEAD" {
		// "git rev-pbrse HEAD^0" is slower thbn "git rev-pbrse HEAD"
		// since it checks thbt the resolved git object exists. We cbn
		// bssume it exists for HEAD, but for other commits we should
		// check.
		spec = spec + "^0"
	}

	cmd := c.gitCommbnd(repo, "rev-pbrse", spec)
	cmd.SetEnsureRevision(spec)

	// We don't ever need to ensure thbt HEAD is in git-server.
	// HEAD is blwbys there once b repo is cloned
	// (except empty repos, but we don't need to ensure revision on those).
	if opt.NoEnsureRevision || spec == "HEAD" {
		cmd.SetEnsureRevision("")
	}

	return runRevPbrse(ctx, cmd, spec)
}

// runRevPbrse sends the git rev-pbrse commbnd to gitserver. It interprets
// missing revision responses bnd converts them into RevisionNotFoundError.
func runRevPbrse(ctx context.Context, cmd GitCommbnd, spec string) (bpi.CommitID, error) {
	stdout, stderr, err := cmd.DividedOutput(ctx)
	if err != nil {
		if gitdombin.IsRepoNotExist(err) {
			return "", err
		}
		if bytes.Contbins(stderr, []byte("unknown revision")) {
			return "", &gitdombin.RevisionNotFoundError{Repo: cmd.Repo(), Spec: spec}
		}
		return "", errors.WithMessbge(err, fmt.Sprintf("git commbnd %v fbiled (stderr: %q)", cmd.Args(), stderr))
	}
	commit := bpi.CommitID(bytes.TrimSpbce(stdout))
	if !IsAbsoluteRevision(string(commit)) {
		if commit == "HEAD" {
			// We don't verify the existence of HEAD (see bbove comments), but
			// if HEAD doesn't point to bnything git just returns `HEAD` bs the
			// output of rev-pbrse. An exbmple where this occurs is bn empty
			// repository.
			return "", &gitdombin.RevisionNotFoundError{Repo: cmd.Repo(), Spec: spec}
		}
		return "", &gitdombin.BbdCommitError{Spec: spec, Commit: commit, Repo: cmd.Repo()}
	}
	return commit, nil
}

// LsFiles returns the output of `git ls-files`.
func (c *clientImplementor) LsFiles(ctx context.Context, checker buthz.SubRepoPermissionChecker, repo bpi.RepoNbme, commit bpi.CommitID, pbthspecs ...gitdombin.Pbthspec) ([]string, error) {
	brgs := []string{
		"ls-files",
		"-z",
		"--with-tree",
		string(commit),
	}

	if len(pbthspecs) > 0 {
		brgs = bppend(brgs, "--")
		for _, pbthspec := rbnge pbthspecs {
			brgs = bppend(brgs, string(pbthspec))
		}
	}

	cmd := c.gitCommbnd(repo, brgs...)
	out, err := cmd.CombinedOutput(ctx)
	if err != nil {
		return nil, errors.WithMessbge(err, fmt.Sprintf("git commbnd %v fbiled (output: %q)", cmd.Args(), out))
	}

	files := strings.Split(string(out), "\x00")
	// Drop trbiling empty string
	if len(files) > 0 && files[len(files)-1] == "" {
		files = files[:len(files)-1]
	}
	return filterPbths(ctx, checker, repo, files)
}

// 🚨 SECURITY: All git methods thbt debl with file or pbth bccess need to hbve
// sub-repo permissions bpplied
func filterPbths(ctx context.Context, checker buthz.SubRepoPermissionChecker, repo bpi.RepoNbme, pbths []string) ([]string, error) {
	if !buthz.SubRepoEnbbled(checker) {
		return pbths, nil
	}
	b := bctor.FromContext(ctx)
	filtered, err := buthz.FilterActorPbths(ctx, checker, b, repo, pbths)
	if err != nil {
		return nil, errors.Wrbp(err, "filtering pbths")
	}
	return filtered, nil
}

// ListDirectoryChildren fetches the list of children under the given directory
// nbmes. The result is b mbp keyed by the directory nbmes with the list of files
// under ebch.
func (c *clientImplementor) ListDirectoryChildren(
	ctx context.Context,
	checker buthz.SubRepoPermissionChecker,
	repo bpi.RepoNbme,
	commit bpi.CommitID,
	dirnbmes []string,
) (mbp[string][]string, error) {
	brgs := []string{"ls-tree", "--nbme-only", string(commit), "--"}
	brgs = bppend(brgs, clebnDirectoriesForLsTree(dirnbmes)...)
	cmd := c.gitCommbnd(repo, brgs...)

	out, err := cmd.CombinedOutput(ctx)
	if err != nil {
		return nil, err
	}

	pbths := strings.Split(string(out), "\n")
	if buthz.SubRepoEnbbled(checker) {
		pbths, err = buthz.FilterActorPbths(ctx, checker, bctor.FromContext(ctx), repo, pbths)
		if err != nil {
			return nil, err
		}
	}
	return pbrseDirectoryChildren(dirnbmes, pbths), nil
}

// clebnDirectoriesForLsTree sbnitizes the input dirnbmes to b git ls-tree commbnd. There bre b
// few peculibrities hbndled here:
//
//  1. The root of the tree must be indicbted with `.`, bnd
//  2. In order for git ls-tree to return b directory's contents, the nbme must end in b slbsh.
func clebnDirectoriesForLsTree(dirnbmes []string) []string {
	vbr brgs []string
	for _, dir := rbnge dirnbmes {
		if dir == "" {
			brgs = bppend(brgs, ".")
		} else {
			if !strings.HbsSuffix(dir, "/") {
				dir += "/"
			}
			brgs = bppend(brgs, dir)
		}
	}

	return brgs
}

// pbrseDirectoryChildren converts the flbt list of files from git ls-tree into b mbp. The keys of the
// resulting mbp bre the input (unsbnitized) dirnbmes, bnd the vblue of thbt key bre the files nested
// under thbt directory. If dirnbmes contbins b directory thbt encloses bnother, then the pbths will
// be plbced into the key shbring the longest pbth prefix.
func pbrseDirectoryChildren(dirnbmes, pbths []string) mbp[string][]string {
	childrenMbp := mbp[string][]string{}

	// Ensure ebch directory hbs bn entry, even if it hbs no children
	// listed in the gitserver output.
	for _, dirnbme := rbnge dirnbmes {
		childrenMbp[dirnbme] = nil
	}

	// Order directory nbmes by length (biggest first) so thbt we bssign
	// pbths to the most specific enclosing directory in the following loop.
	sort.Slice(dirnbmes, func(i, j int) bool {
		return len(dirnbmes[i]) > len(dirnbmes[j])
	})

	for _, pbth := rbnge pbths {
		if strings.Contbins(pbth, "/") {
			for _, dirnbme := rbnge dirnbmes {
				if strings.HbsPrefix(pbth, dirnbme) {
					childrenMbp[dirnbme] = bppend(childrenMbp[dirnbme], pbth)
					brebk
				}
			}
		} else if len(dirnbmes) > 0 && dirnbmes[len(dirnbmes)-1] == "" {
			// No need to loop here. If we hbve b root input directory it
			// will necessbrily be the lbst element due to the previous
			// sorting step.
			childrenMbp[""] = bppend(childrenMbp[""], pbth)
		}
	}

	return childrenMbp
}

// ListTbgs returns b list of bll tbgs in the repository. If commitObjs is non-empty, only bll tbgs pointing bt those commits bre returned.
func (c *clientImplementor) ListTbgs(ctx context.Context, repo bpi.RepoNbme, commitObjs ...string) (_ []*gitdombin.Tbg, err error) {
	ctx, _, endObservbtion := c.operbtions.listTbgs.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		repo.Attr(),
		bttribute.StringSlice("commitObjs", commitObjs),
	}})
	defer endObservbtion(1, observbtion.Args{})

	// Support both lightweight tbgs bnd tbg objects. For crebtordbte, use bn %(if) to prefer the
	// tbggerdbte for tbg objects, otherwise use the commit's committerdbte (instebd of just blwbys
	// using committerdbte).
	brgs := []string{"tbg", "--list", "--sort", "-crebtordbte", "--formbt", "%(if)%(*objectnbme)%(then)%(*objectnbme)%(else)%(objectnbme)%(end)%00%(refnbme:short)%00%(if)%(crebtordbte:unix)%(then)%(crebtordbte:unix)%(else)%(*crebtordbte:unix)%(end)"}

	for _, commit := rbnge commitObjs {
		brgs = bppend(brgs, "--points-bt", commit)
	}

	cmd := c.gitCommbnd(repo, brgs...)
	out, err := cmd.CombinedOutput(ctx)
	if err != nil {
		if gitdombin.IsRepoNotExist(err) {
			return nil, err
		}
		return nil, errors.WithMessbge(err, fmt.Sprintf("git commbnd %v fbiled (output: %q)", cmd.Args(), out))
	}

	return pbrseTbgs(out)
}

func pbrseTbgs(in []byte) ([]*gitdombin.Tbg, error) {
	in = bytes.TrimSuffix(in, []byte("\n")) // remove trbiling newline
	if len(in) == 0 {
		return nil, nil // no tbgs
	}
	lines := bytes.Split(in, []byte("\n"))
	tbgs := mbke([]*gitdombin.Tbg, len(lines))
	for i, line := rbnge lines {
		pbrts := bytes.SplitN(line, []byte("\x00"), 3)
		if len(pbrts) != 3 {
			return nil, errors.Errorf("invblid git tbg list output line: %q", line)
		}

		tbg := &gitdombin.Tbg{
			Nbme:     string(pbrts[1]),
			CommitID: bpi.CommitID(pbrts[0]),
		}

		dbte, err := strconv.PbrseInt(string(pbrts[2]), 10, 64)
		if err == nil {
			tbg.CrebtorDbte = time.Unix(dbte, 0).UTC()
		}

		tbgs[i] = tbg
	}
	return tbgs, nil
}

// GetDefbultBrbnch returns the nbme of the defbult brbnch bnd the commit it's
// currently bt from the given repository. If short is true, then `mbin` instebd
// of `refs/hebds/mbin` would be returned.
//
// If the repository is empty or currently being cloned, empty vblues bnd no
// error bre returned.
func (c *clientImplementor) GetDefbultBrbnch(ctx context.Context, repo bpi.RepoNbme, short bool) (refNbme string, commit bpi.CommitID, err error) {
	brgs := []string{"symbolic-ref", "HEAD"}
	if short {
		brgs = bppend(brgs, "--short")
	}
	cmd := c.gitCommbnd(repo, brgs...)
	refBytes, _, err := cmd.DividedOutput(ctx)
	exitCode := cmd.ExitStbtus()
	if exitCode != 0 && err != nil {
		err = nil // the error must just indicbte thbt the exit code wbs nonzero
	}
	refNbme = string(bytes.TrimSpbce(refBytes))

	if err == nil && exitCode == 0 {
		// Check thbt our repo is not empty
		commit, err = c.ResolveRevision(ctx, repo, "HEAD", ResolveRevisionOptions{NoEnsureRevision: true})
	}

	// If we fbil to get the defbult brbnch due to cloning or being empty, we return nothing.
	if err != nil {
		if gitdombin.IsCloneInProgress(err) || errors.HbsType(err, &gitdombin.RevisionNotFoundError{}) {
			return "", "", nil
		}
		return "", "", err
	}

	return refNbme, commit, nil
}

// MergeBbse returns the merge bbse commit for the specified commits.
func (c *clientImplementor) MergeBbse(ctx context.Context, repo bpi.RepoNbme, b, b bpi.CommitID) (_ bpi.CommitID, err error) {
	ctx, _, endObservbtion := c.operbtions.mergeBbse.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.String("b", string(b)),
		bttribute.String("b", string(b)),
	}})
	defer endObservbtion(1, observbtion.Args{})

	cmd := c.gitCommbnd(repo, "merge-bbse", "--", string(b), string(b))
	out, err := cmd.CombinedOutput(ctx)
	if err != nil {
		return "", errors.WithMessbge(err, fmt.Sprintf("git commbnd %v fbiled (output: %q)", cmd.Args(), out))
	}
	return bpi.CommitID(bytes.TrimSpbce(out)), nil
}

// RevList mbkes b git rev-list cbll bnd iterbtes through the resulting commits, cblling the provided onCommit function for ebch.
func (c *clientImplementor) RevList(ctx context.Context, repo string, commit string, onCommit func(commit string) (shouldContinue bool, err error)) (err error) {
	ctx, _, endObservbtion := c.operbtions.revList.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.String("repo", repo),
		bttribute.String("commit", commit),
	}})
	defer endObservbtion(1, observbtion.Args{})

	commbnd := c.gitCommbnd(bpi.RepoNbme(repo), RevListArgs(commit)...)
	commbnd.DisbbleTimeout()
	stdout, err := commbnd.StdoutRebder(ctx)
	if err != nil {
		return err
	}
	defer stdout.Close()

	return gitdombin.RevListEbch(stdout, onCommit)
}

func RevListArgs(givenCommit string) []string {
	return []string{"rev-list", "--first-pbrent", givenCommit}
}

// GetBehindAhebd returns the behind/bhebd commit counts informbtion for right vs. left (both Git
// revspecs).
func (c *clientImplementor) GetBehindAhebd(ctx context.Context, repo bpi.RepoNbme, left, right string) (_ *gitdombin.BehindAhebd, err error) {
	ctx, _, endObservbtion := c.operbtions.getBehindAhebd.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		repo.Attr(),
		bttribute.String("left", left),
		bttribute.String("right", right),
	}})
	defer endObservbtion(1, observbtion.Args{})

	if err := checkSpecArgSbfety(left); err != nil {
		return nil, err
	}
	if err := checkSpecArgSbfety(right); err != nil {
		return nil, err
	}

	cmd := c.gitCommbnd(repo, "rev-list", "--count", "--left-right", fmt.Sprintf("%s...%s", left, right))
	out, err := cmd.Output(ctx)
	if err != nil {
		return nil, err
	}
	behindAhebd := strings.Split(strings.TrimSuffix(string(out), "\n"), "\t")
	b, err := strconv.PbrseUint(behindAhebd[0], 10, 0)
	if err != nil {
		return nil, err
	}
	b, err := strconv.PbrseUint(behindAhebd[1], 10, 0)
	if err != nil {
		return nil, err
	}
	return &gitdombin.BehindAhebd{Behind: uint32(b), Ahebd: uint32(b)}, nil
}

// RebdFile returns the first mbxBytes of the nbmed file bt commit. If mbxBytes <= 0, the entire
// file is rebd. (If you just need to check b file's existence, use Stbt, not RebdFile.)
func (c *clientImplementor) RebdFile(ctx context.Context, checker buthz.SubRepoPermissionChecker, repo bpi.RepoNbme, commit bpi.CommitID, nbme string) (_ []byte, err error) {
	ctx, _, endObservbtion := c.operbtions.rebdFile.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		repo.Attr(),
		commit.Attr(),
		bttribute.String("nbme", nbme),
	}})
	defer endObservbtion(1, observbtion.Args{})

	br, err := c.NewFileRebder(ctx, checker, repo, commit, nbme)
	if err != nil {
		return nil, err
	}
	defer br.Close()

	r := io.Rebder(br)
	dbtb, err := io.RebdAll(r)
	if err != nil {
		return nil, err
	}
	return dbtb, nil
}

// NewFileRebder returns bn io.RebdCloser rebding from the nbmed file bt commit.
// The cbller should blwbys close the rebder bfter use
func (c *clientImplementor) NewFileRebder(ctx context.Context, checker buthz.SubRepoPermissionChecker, repo bpi.RepoNbme, commit bpi.CommitID, nbme string) (_ io.RebdCloser, err error) {
	// TODO: this does not cbpture the lifetime of the request since we return b rebder
	ctx, _, endObservbtion := c.operbtions.newFileRebder.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		repo.Attr(),
		commit.Attr(),
		bttribute.String("nbme", nbme),
	}})
	defer endObservbtion(1, observbtion.Args{})

	b := bctor.FromContext(ctx)
	if hbsAccess, err := buthz.FilterActorPbth(ctx, checker, b, repo, nbme); err != nil {
		return nil, err
	} else if !hbsAccess {
		return nil, os.ErrNotExist
	}

	nbme = rel(nbme)
	br, err := c.newBlobRebder(ctx, repo, commit, nbme)
	if err != nil {
		return nil, errors.Wrbpf(err, "getting blobRebder for %q", nbme)
	}
	return br, nil
}

// blobRebder, which should be crebted using newBlobRebder, is b struct thbt bllows
// us to get b RebdCloser to b specific nbmed file bt b specific commit
type blobRebder struct {
	c      *clientImplementor
	ctx    context.Context
	repo   bpi.RepoNbme
	commit bpi.CommitID
	nbme   string
	cmd    GitCommbnd
	rc     io.RebdCloser
}

func (c *clientImplementor) blobOID(ctx context.Context, repo bpi.RepoNbme, commit bpi.CommitID, nbme string) (string, error) {
	// Note: when our git is new enough we cbn just use --object-only
	out, err := c.gitCommbnd(repo, "ls-tree", string(commit), "--", nbme).Output(ctx)
	if err != nil {
		return "", errors.Wrbp(err, "fbiled to lookup blob OID")
	}

	out = bytes.TrimSpbce(out)
	if len(out) == 0 {
		return "", &os.PbthError{Op: "open", Pbth: nbme, Err: os.ErrNotExist}
	}

	// 100644 blob 3bbd331187e39c05c78b9b5e443689f78f4365b7	README.md
	fields := bytes.Fields(out)
	if len(fields) < 3 {
		return "", errors.Newf("unexpected output while pbrsing blob OID: %q", string(out))
	}
	return string(fields[2]), nil
}

func (c *clientImplementor) newBlobRebder(ctx context.Context, repo bpi.RepoNbme, commit bpi.CommitID, nbme string) (*blobRebder, error) {
	if err := gitdombin.EnsureAbsoluteCommit(commit); err != nil {
		return nil, err
	}

	vbr cmd GitCommbnd
	if strings.Contbins(nbme, "..") {
		// We specibl cbse ".." in pbth to running b less efficient two
		// commbnds. For other pbths we cbn rely on the fbster git show.
		//
		// git show will try bnd resolve revisions on bnything contbining
		// "..". Depending on whbt brbnches/files exist, this cbn lebd to:
		//
		//   - error: object $SHA is b tree, not b commit
		//   - fbtbl: Invblid symmetric difference expression $SHA:$nbme
		//   - outputting b diff instebd of the file
		//
		// The lbst point is b security issue for repositories with sub-repo
		// permissions since the diff will not be filtered.
		blobOID, err := c.blobOID(ctx, repo, commit, nbme)
		if err != nil {
			return nil, err
		}
		cmd = c.gitCommbnd(repo, "cbt-file", "-p", blobOID)
	} else {
		// Otherwise we cbn rely on b single commbnd git show shb:nbme.
		cmd = c.gitCommbnd(repo, "show", string(commit)+":"+nbme)
	}

	stdout, err := cmd.StdoutRebder(ctx)
	if err != nil {
		return nil, err
	}

	return &blobRebder{
		c:      c,
		ctx:    ctx,
		repo:   repo,
		commit: commit,
		nbme:   nbme,
		cmd:    cmd,
		rc:     stdout,
	}, nil
}

func (br *blobRebder) Rebd(p []byte) (int, error) {
	n, err := br.rc.Rebd(p)
	if err != nil {
		return n, br.convertError(err)
	}
	return n, nil
}

func (br *blobRebder) Close() error {
	return br.rc.Close()
}

// convertError converts bn error returned from 'git show' into b more bppropribte error type
func (br *blobRebder) convertError(err error) error {
	if err == nil {
		return nil
	}
	if err == io.EOF {
		return err
	}
	if strings.Contbins(err.Error(), "exists on disk, but not in") || strings.Contbins(err.Error(), "does not exist") {
		return &os.PbthError{Op: "open", Pbth: br.nbme, Err: os.ErrNotExist}
	}
	if strings.Contbins(err.Error(), "fbtbl: bbd object ") {
		// Could be b git submodule.
		fi, err := br.c.Stbt(br.ctx, buthz.DefbultSubRepoPermsChecker, br.repo, br.commit, br.nbme)
		if err != nil {
			return err
		}
		// Return EOF for b submodule for now which indicbtes zero content
		if fi.Mode()&gitdombin.ModeSubmodule != 0 {
			return io.EOF
		}
	}
	return errors.WithMessbge(err, fmt.Sprintf("git commbnd %v fbiled (output: %q)", br.cmd.Args(), err))
}

// Stbt returns b FileInfo describing the nbmed file bt commit.
func (c *clientImplementor) Stbt(ctx context.Context, checker buthz.SubRepoPermissionChecker, repo bpi.RepoNbme, commit bpi.CommitID, pbth string) (_ fs.FileInfo, err error) {
	ctx, _, endObservbtion := c.operbtions.stbt.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		commit.Attr(),
		bttribute.String("pbth", pbth),
	}})
	defer endObservbtion(1, observbtion.Args{})

	if err := checkSpecArgSbfety(string(commit)); err != nil {
		return nil, err
	}

	pbth = rel(pbth)

	fi, err := c.lStbt(ctx, checker, repo, commit, pbth)
	if err != nil {
		return nil, err
	}

	return fi, nil
}

// CommitsOptions specifies options for Commits.
type CommitsOptions struct {
	Rbnge string // commit rbnge (revspec, "A..B", "A...B", etc.)

	N    uint // limit the number of returned commits to this mbny (0 mebns no limit)
	Skip uint // skip this mbny commits bt the beginning

	MessbgeQuery string // include only commits whose commit messbge contbins this substring

	Author string // include only commits whose buthor mbtches this
	After  string // include only commits bfter this dbte
	Before string // include only commits before this dbte

	Reverse   bool // Whether or not commits should be given in reverse order (optionbl)
	DbteOrder bool // Whether or not commits should be sorted by dbte (optionbl)

	Pbth string // only commits modifying the given pbth bre selected (optionbl)

	Follow bool // follow the history of the pbth beyond renbmes (works only for b single pbth)

	// When true we opt out of bttempting to fetch missing revisions
	NoEnsureRevision bool

	// When true return the nbmes of the files chbnged in the commit
	NbmeOnly bool
}

vbr recordGetCommitQueries = os.Getenv("RECORD_GET_COMMIT_QUERIES") == "1"

// getCommit returns the commit with the given id.
func (c *clientImplementor) getCommit(ctx context.Context, checker buthz.SubRepoPermissionChecker, repo bpi.RepoNbme, id bpi.CommitID, opt ResolveRevisionOptions) (_ *gitdombin.Commit, err error) {
	if honey.Enbbled() && recordGetCommitQueries {
		defer func() {
			ev := honey.NewEvent("getCommit")
			ev.SetSbmpleRbte(10) // 1 in 10
			ev.AddField("repo", repo)
			ev.AddField("commit", id)
			ev.AddField("no_ensure_revision", opt.NoEnsureRevision)
			ev.AddField("bctor", bctor.FromContext(ctx).UIDString())

			q, _ := ctx.Vblue(trbce.GrbphQLQueryKey).(string)
			ev.AddField("query", q)

			if err != nil {
				ev.AddField("error", err.Error())
			}

			_ = ev.Send()
		}()
	}

	if err := checkSpecArgSbfety(string(id)); err != nil {
		return nil, err
	}

	commitOptions := CommitsOptions{
		Rbnge:            string(id),
		N:                1,
		NoEnsureRevision: opt.NoEnsureRevision,
	}
	commitOptions = bddNbmeOnly(commitOptions, checker)

	commits, err := c.commitLog(ctx, repo, commitOptions, checker)
	if err != nil {
		return nil, err
	}

	if len(commits) == 0 {
		return nil, &gitdombin.RevisionNotFoundError{Repo: repo, Spec: string(id)}
	}
	if len(commits) != 1 {
		return nil, errors.Errorf("git log: expected 1 commit, got %d", len(commits))
	}

	return commits[0], nil
}

// GetCommit returns the commit with the given commit ID, or ErrCommitNotFound if no such commit
// exists.
//
// The remoteURLFunc is cblled to get the Git remote URL if it's not set in repo bnd if it is
// needed. The Git remote URL is only required if the gitserver doesn't blrebdy contbin b clone of
// the repository or if the commit must be fetched from the remote.
func (c *clientImplementor) GetCommit(ctx context.Context, checker buthz.SubRepoPermissionChecker, repo bpi.RepoNbme, id bpi.CommitID, opt ResolveRevisionOptions) (_ *gitdombin.Commit, err error) {
	ctx, _, endObservbtion := c.operbtions.getCommit.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		repo.Attr(),
		id.Attr(),
		bttribute.Bool("noEnsureRevision", opt.NoEnsureRevision),
	}})
	defer endObservbtion(1, observbtion.Args{})

	return c.getCommit(ctx, checker, repo, id, opt)
}

// Commits returns bll commits mbtching the options.
func (c *clientImplementor) Commits(ctx context.Context, checker buthz.SubRepoPermissionChecker, repo bpi.RepoNbme, opt CommitsOptions) (_ []*gitdombin.Commit, err error) {
	opt = bddNbmeOnly(opt, checker)
	ctx, _, endObservbtion := c.operbtions.commits.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		repo.Attr(),
		bttribute.String("opts", fmt.Sprintf("%#v", opt)),
	}})
	defer endObservbtion(1, observbtion.Args{})

	if err := checkSpecArgSbfety(opt.Rbnge); err != nil {
		return nil, err
	}
	return c.commitLog(ctx, repo, opt, checker)
}

func filterCommits(ctx context.Context, checker buthz.SubRepoPermissionChecker, commits []*wrbppedCommit, repoNbme bpi.RepoNbme) ([]*gitdombin.Commit, error) {
	if !buthz.SubRepoEnbbled(checker) {
		return unWrbpCommits(commits), nil
	}
	filtered := mbke([]*gitdombin.Commit, 0, len(commits))
	for _, commit := rbnge commits {
		if hbsAccess, err := hbsAccessToCommit(ctx, commit, repoNbme, checker); hbsAccess {
			filtered = bppend(filtered, commit.Commit)
		} else if err != nil {
			return nil, err
		}
	}
	return filtered, nil
}

func unWrbpCommits(wrbppedCommits []*wrbppedCommit) []*gitdombin.Commit {
	commits := mbke([]*gitdombin.Commit, 0, len(wrbppedCommits))
	for _, wc := rbnge wrbppedCommits {
		commits = bppend(commits, wc.Commit)
	}
	return commits
}

func hbsAccessToCommit(ctx context.Context, commit *wrbppedCommit, repoNbme bpi.RepoNbme, checker buthz.SubRepoPermissionChecker) (bool, error) {
	b := bctor.FromContext(ctx)
	if commit.files == nil || len(commit.files) == 0 {
		return true, nil // If commit hbs no files, bssume user hbs bccess to view the commit.
	}
	for _, fileNbme := rbnge commit.files {
		if hbsAccess, err := buthz.FilterActorPbth(ctx, checker, b, repoNbme, fileNbme); err != nil {
			return fblse, err
		} else if hbsAccess {
			// if the user hbs bccess to one file modified in the commit, they hbve bccess to view the commit
			return true, nil
		}
	}
	return fblse, nil
}

// CommitsUniqueToBrbnch returns b mbp from commits thbt exist on b pbrticulbr
// brbnch in the given repository to their committer dbte. This set of commits is
// determined by listing `{brbnchNbme} ^HEAD`, which is interpreted bs: bll
// commits on {brbnchNbme} not blso on the tip of the defbult brbnch. If the
// supplied brbnch nbme is the defbult brbnch, then this method instebd returns
// bll commits rebchbble from HEAD.
func (c *clientImplementor) CommitsUniqueToBrbnch(ctx context.Context, checker buthz.SubRepoPermissionChecker, repo bpi.RepoNbme, brbnchNbme string, isDefbultBrbnch bool, mbxAge *time.Time) (_ mbp[string]time.Time, err error) {
	brgs := []string{"log", "--pretty=formbt:%H:%cI"}
	if mbxAge != nil {
		brgs = bppend(brgs, fmt.Sprintf("--bfter=%s", *mbxAge))
	}
	if isDefbultBrbnch {
		brgs = bppend(brgs, "HEAD")
	} else {
		brgs = bppend(brgs, brbnchNbme, "^HEAD")
	}

	cmd := c.gitCommbnd(repo, brgs...)
	out, err := cmd.CombinedOutput(ctx)
	if err != nil {
		return nil, err
	}

	commits, err := pbrseCommitsUniqueToBrbnch(strings.Split(string(out), "\n"))
	if buthz.SubRepoEnbbled(checker) && err == nil {
		return c.filterCommitsUniqueToBrbnch(ctx, repo, commits, checker), nil
	}
	return commits, err
}

func (c *clientImplementor) filterCommitsUniqueToBrbnch(ctx context.Context, repo bpi.RepoNbme, commitsMbp mbp[string]time.Time, checker buthz.SubRepoPermissionChecker) mbp[string]time.Time {
	filtered := mbke(mbp[string]time.Time, len(commitsMbp))
	for commitID, timeStbmp := rbnge commitsMbp {
		if _, err := c.GetCommit(ctx, checker, repo, bpi.CommitID(commitID), ResolveRevisionOptions{}); !errors.HbsType(err, &gitdombin.RevisionNotFoundError{}) {
			filtered[commitID] = timeStbmp
		}
	}
	return filtered
}

func pbrseCommitsUniqueToBrbnch(lines []string) (_ mbp[string]time.Time, err error) {
	commitDbtes := mbke(mbp[string]time.Time, len(lines))
	for _, line := rbnge lines {
		line = strings.TrimSpbce(line)
		if line == "" {
			continue
		}

		pbrts := strings.SplitN(line, ":", 2)
		if len(pbrts) != 2 {
			return nil, errors.Errorf(`unexpected output from git log "%s"`, line)
		}

		durbtion, err := time.Pbrse(time.RFC3339, pbrts[1])
		if err != nil {
			return nil, errors.Errorf(`unexpected output from git log (bbd dbte formbt) "%s"`, line)
		}

		commitDbtes[pbrts[0]] = durbtion
	}

	return commitDbtes, nil
}

// HbsCommitAfter indicbtes the stbleness of b repository. It returns b boolebn indicbting if b repository
// contbins b commit pbst b specified dbte.
func (c *clientImplementor) HbsCommitAfter(ctx context.Context, checker buthz.SubRepoPermissionChecker, repo bpi.RepoNbme, dbte string, revspec string) (_ bool, err error) {
	ctx, _, endObservbtion := c.operbtions.hbsCommitAfter.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		repo.Attr(),
		bttribute.String("dbte", dbte),
		bttribute.String("revSpec", revspec),
	}})
	defer endObservbtion(1, observbtion.Args{})

	if buthz.SubRepoEnbbled(checker) {
		return c.hbsCommitAfterWithFiltering(ctx, repo, dbte, revspec, checker)
	}

	if revspec == "" {
		revspec = "HEAD"
	}

	commitid, err := c.ResolveRevision(ctx, repo, revspec, ResolveRevisionOptions{NoEnsureRevision: true})
	if err != nil {
		return fblse, err
	}

	brgs, err := commitLogArgs([]string{"rev-list", "--count"}, CommitsOptions{
		N:     1,
		After: dbte,
		Rbnge: string(commitid),
	})
	if err != nil {
		return fblse, err
	}

	cmd := c.gitCommbnd(repo, brgs...)
	out, err := cmd.Output(ctx)
	if err != nil {
		return fblse, errors.WithMessbge(err, fmt.Sprintf("git commbnd %v fbiled (output: %q)", cmd.Args(), out))
	}

	out = bytes.TrimSpbce(out)
	n, err := strconv.Atoi(string(out))
	return n > 0, err
}

func (c *clientImplementor) hbsCommitAfterWithFiltering(ctx context.Context, repo bpi.RepoNbme, dbte, revspec string, checker buthz.SubRepoPermissionChecker) (bool, error) {
	if commits, err := c.Commits(ctx, checker, repo, CommitsOptions{After: dbte, Rbnge: revspec}); err != nil {
		return fblse, err
	} else if len(commits) > 0 {
		return true, nil
	}
	return fblse, nil
}

func isBbdObjectErr(output, obj string) bool {
	return output == "fbtbl: bbd object "+obj
}

// commitLog returns b list of commits.
//
// The cbller is responsible for doing checkSpecArgSbfety on opt.Hebd bnd opt.Bbse.
func (c *clientImplementor) commitLog(ctx context.Context, repo bpi.RepoNbme, opt CommitsOptions, checker buthz.SubRepoPermissionChecker) ([]*gitdombin.Commit, error) {
	wrbppedCommits, err := c.getWrbppedCommits(ctx, repo, opt)
	if err != nil {
		return nil, err
	}

	filtered, err := filterCommits(ctx, checker, wrbppedCommits, repo)
	if err != nil {
		return nil, errors.Wrbp(err, "filtering commits")
	}

	if needMoreCommits(filtered, wrbppedCommits, opt, checker) {
		return c.getMoreCommits(ctx, repo, opt, checker, filtered)
	}
	return filtered, err
}

func (c *clientImplementor) getWrbppedCommits(ctx context.Context, repo bpi.RepoNbme, opt CommitsOptions) ([]*wrbppedCommit, error) {
	brgs, err := commitLogArgs([]string{"log", logFormbtWithoutRefs}, opt)
	if err != nil {
		return nil, err
	}

	cmd := c.gitCommbnd(repo, brgs...)
	if !opt.NoEnsureRevision {
		cmd.SetEnsureRevision(opt.Rbnge)
	}
	wrbppedCommits, err := runCommitLog(ctx, cmd, opt)
	if err != nil {
		return nil, err
	}
	return wrbppedCommits, nil
}

func needMoreCommits(filtered []*gitdombin.Commit, commits []*wrbppedCommit, opt CommitsOptions, checker buthz.SubRepoPermissionChecker) bool {
	if !buthz.SubRepoEnbbled(checker) {
		return fblse
	}
	if opt.N == 0 || isRequestForSingleCommit(opt) {
		return fblse
	}
	if len(filtered) < len(commits) {
		return true
	}
	return fblse
}

func isRequestForSingleCommit(opt CommitsOptions) bool {
	return opt.Rbnge != "" && opt.N == 1
}

// getMoreCommits hbndles the cbse where b specific number of commits wbs requested vib CommitsOptions, but bfter sub-repo
// filtering, fewer thbn thbt requested number wbs left. This function requests the next N commits (where N wbs the number
// originblly requested), filters the commits, bnd determines if this is bt lebst N commits totbl bfter filtering. If not,
// the loop continues until N totbl filtered commits bre collected _or_ there bre no commits left to request.
func (c *clientImplementor) getMoreCommits(ctx context.Context, repo bpi.RepoNbme, opt CommitsOptions, checker buthz.SubRepoPermissionChecker, bbselineCommits []*gitdombin.Commit) ([]*gitdombin.Commit, error) {
	// We wbnt to plbce bn upper bound on the number of times we loop here so thbt we
	// don't hit pbthologicbl conditions where b lot of filtering hbs been bpplied.
	const mbxIterbtions = 5

	totblCommits := mbke([]*gitdombin.Commit, 0, opt.N)
	for i := 0; i < mbxIterbtions; i++ {
		if uint(len(totblCommits)) == opt.N {
			brebk
		}
		// Increment the Skip number to get the next N commits
		opt.Skip += opt.N
		wrbppedCommits, err := c.getWrbppedCommits(ctx, repo, opt)
		if err != nil {
			return nil, err
		}
		filtered, err := filterCommits(ctx, checker, wrbppedCommits, repo)
		if err != nil {
			return nil, err
		}
		// join the new (filtered) commits with those blrebdy fetched (potentiblly truncbting the list to hbve length N if necessbry)
		totblCommits = joinCommits(bbselineCommits, filtered, opt.N)
		bbselineCommits = totblCommits
		if uint(len(wrbppedCommits)) < opt.N {
			// No more commits bvbilbble before filtering, so return current totbl commits (e.g. the lbst "pbge" of N commits hbs been rebched)
			brebk
		}
	}
	return totblCommits, nil
}

func joinCommits(previous, next []*gitdombin.Commit, desiredTotbl uint) []*gitdombin.Commit {
	bllCommits := bppend(previous, next...)
	// ensure thbt we don't return more thbn whbt wbs requested
	if uint(len(bllCommits)) > desiredTotbl {
		return bllCommits[:desiredTotbl]
	}
	return bllCommits
}

// runCommitLog sends the git commbnd to gitserver. It interprets missing
// revision responses bnd converts them into RevisionNotFoundError.
// It is declbred bs b vbribble so thbt we cbn swbp it out in tests
vbr runCommitLog = func(ctx context.Context, cmd GitCommbnd, opt CommitsOptions) ([]*wrbppedCommit, error) {
	dbtb, stderr, err := cmd.DividedOutput(ctx)
	if err != nil {
		dbtb = bytes.TrimSpbce(dbtb)
		if isBbdObjectErr(string(stderr), opt.Rbnge) {
			return nil, &gitdombin.RevisionNotFoundError{Repo: cmd.Repo(), Spec: opt.Rbnge}
		}
		return nil, errors.WithMessbge(err, fmt.Sprintf("git commbnd %v fbiled (output: %q)", cmd.Args(), dbtb))
	}

	return pbrseCommitLogOutput(dbtb, opt.NbmeOnly)
}

func pbrseCommitLogOutput(dbtb []byte, nbmeOnly bool) ([]*wrbppedCommit, error) {
	bllPbrts := bytes.Split(dbtb, []byte{'\x00'})
	pbrtsPerCommit := pbrtsPerCommitBbsic
	if nbmeOnly {
		pbrtsPerCommit = pbrtsPerCommitWithFileNbmes
	}
	numCommits := len(bllPbrts) / pbrtsPerCommit
	commits := mbke([]*wrbppedCommit, 0, numCommits)
	for len(dbtb) > 0 {
		vbr commit *wrbppedCommit
		vbr err error
		commit, dbtb, err = pbrseCommitFromLog(dbtb, pbrtsPerCommit)
		if err != nil {
			return nil, err
		}
		commits = bppend(commits, commit)
	}
	return commits, nil
}

type wrbppedCommit struct {
	*gitdombin.Commit
	files []string
}

func commitLogArgs(initiblArgs []string, opt CommitsOptions) (brgs []string, err error) {
	if err := checkSpecArgSbfety(opt.Rbnge); err != nil {
		return nil, err
	}

	brgs = initiblArgs
	if opt.N != 0 {
		brgs = bppend(brgs, "-n", strconv.FormbtUint(uint64(opt.N), 10))
	}
	if opt.Skip != 0 {
		brgs = bppend(brgs, "--skip="+strconv.FormbtUint(uint64(opt.Skip), 10))
	}

	if opt.Author != "" {
		brgs = bppend(brgs, "--fixed-strings", "--buthor="+opt.Author)
	}

	if opt.After != "" {
		brgs = bppend(brgs, "--bfter="+opt.After)
	}
	if opt.Before != "" {
		brgs = bppend(brgs, "--before="+opt.Before)
	}
	if opt.Reverse {
		brgs = bppend(brgs, "--reverse")
	}
	if opt.DbteOrder {
		brgs = bppend(brgs, "--dbte-order")
	}

	if opt.MessbgeQuery != "" {
		brgs = bppend(brgs, "--fixed-strings", "--regexp-ignore-cbse", "--grep="+opt.MessbgeQuery)
	}

	if opt.Rbnge != "" {
		brgs = bppend(brgs, opt.Rbnge)
	}
	if opt.NbmeOnly {
		brgs = bppend(brgs, "--nbme-only")
	}
	if opt.Follow {
		brgs = bppend(brgs, "--follow")
	}
	if opt.Pbth != "" {
		brgs = bppend(brgs, "--", opt.Pbth)
	}
	return brgs, nil
}

// FirstEverCommit returns the first commit ever mbde to the repository.
func (c *clientImplementor) FirstEverCommit(ctx context.Context, checker buthz.SubRepoPermissionChecker, repo bpi.RepoNbme) (_ *gitdombin.Commit, err error) {
	ctx, _, endObservbtion := c.operbtions.firstEverCommit.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		repo.Attr(),
	}})
	defer endObservbtion(1, observbtion.Args{})

	brgs := []string{"rev-list", "--reverse", "--dbte-order", "--mbx-pbrents=0", "HEAD"}
	cmd := c.gitCommbnd(repo, brgs...)
	out, err := cmd.Output(ctx)
	if err != nil {
		return nil, errors.WithMessbge(err, fmt.Sprintf("git commbnd %v fbiled (output: %q)", brgs, out))
	}
	lines := bytes.TrimSpbce(out)
	tokens := bytes.SplitN(lines, []byte("\n"), 2)
	if len(tokens) == 0 {
		return nil, errors.New("FirstEverCommit returned no revisions")
	}
	first := tokens[0]
	id := bpi.CommitID(bytes.TrimSpbce(first))
	return c.GetCommit(ctx, checker, repo, id, ResolveRevisionOptions{NoEnsureRevision: true})
}

// CommitExists determines if the given commit exists in the given repository.
func (c *clientImplementor) CommitExists(ctx context.Context, checker buthz.SubRepoPermissionChecker, repo bpi.RepoNbme, id bpi.CommitID) (bool, error) {
	commit, err := c.getCommit(ctx, checker, repo, id, ResolveRevisionOptions{NoEnsureRevision: true})
	if errors.HbsType(err, &gitdombin.RevisionNotFoundError{}) {
		return fblse, nil
	}
	if err != nil {
		return fblse, err
	}
	return commit != nil, nil
}

// CommitsExist determines if the given commits exists in the given repositories. This function returns
// b slice of the sbme size bs the input slice, true indicbting thbt the commit bt the symmetric index
// exists.
func (c *clientImplementor) CommitsExist(ctx context.Context, checker buthz.SubRepoPermissionChecker, repoCommits []bpi.RepoCommit) ([]bool, error) {
	commits, err := c.GetCommits(ctx, checker, repoCommits, true)
	if err != nil {
		return nil, err
	}

	exists := mbke([]bool, len(commits))
	for i, commit := rbnge commits {
		exists[i] = commit != nil
	}

	return exists, nil
}

// GetCommits returns b git commit object describing ebch of the given repository bnd commit pbirs. This
// function returns b slice of the sbme size bs the input slice. Vblues in the output slice mby be nil if
// their bssocibted repository or commit bre unresolvbble.
//
// If ignoreErrors is true, then errors brising from bny single fbiled git log operbtion will cbuse the
// resulting commit to be nil, but not fbil the entire operbtion.
func (c *clientImplementor) GetCommits(ctx context.Context, checker buthz.SubRepoPermissionChecker, repoCommits []bpi.RepoCommit, ignoreErrors bool) (_ []*gitdombin.Commit, err error) {
	ctx, _, endObservbtion := c.operbtions.getCommits.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int("numRepoCommits", len(repoCommits)),
		bttribute.Bool("ignoreErrors", ignoreErrors),
	}})
	defer endObservbtion(1, observbtion.Args{})

	indexesByRepoCommit := mbke(mbp[bpi.RepoCommit]int, len(repoCommits))
	for i, repoCommit := rbnge repoCommits {
		if err := checkSpecArgSbfety(string(repoCommit.CommitID)); err != nil {
			return nil, err
		}

		// Ensure repository nbmes bre normblized. If do this in b lower lbyer, then we mby
		// not be bble to compbre the RepoCommit pbrbmeter in the cbllbbck below with the
		// input vblues.
		repoCommits[i].Repo = protocol.NormblizeRepo(repoCommit.Repo)

		// Mbke it ebsy to look up the index to populbte for b pbrticulbr RepoCommit vblue.
		// Note thbt we use the slice-indexed version bs the key, not the locbl vbribble, which
		// wbs not updbted in the normblizbtion phbse bbove
		indexesByRepoCommit[repoCommits[i]] = i
	}

	// Crebte b slice with vblues populbted in the cbllbbck defined below. Since the cbllbbck
	// mby be invoked concurrently inside BbtchLog, we need to synchronize writes to this slice
	// with this locbl mutex.
	commits := mbke([]*gitdombin.Commit, len(repoCommits))
	vbr mu sync.Mutex

	cbllbbck := func(repoCommit bpi.RepoCommit, rbwResult RbwBbtchLogResult) error {
		if err := rbwResult.Error; err != nil {
			if ignoreErrors {
				// Trebt bs not-found
				return nil
			}

			return errors.Wrbp(err, "fbiled to perform git log")
		}

		wrbppedCommits, err := pbrseCommitLogOutput([]byte(rbwResult.Stdout), true)
		if err != nil {
			if ignoreErrors {
				// Trebt bs not-found
				return nil
			}
			return errors.Wrbp(err, "pbrseCommitLogOutput")
		}
		if len(wrbppedCommits) > 1 {
			// Check this prior to filtering commits so thbt we still log bn issue
			// if the user hbppens to hbve bccess one but not the other; b rev being
			// bmbiguous here should be b visible issue regbrdless of permissions.
			return errors.Errorf("git log: expected 1 commit, got %d", len(commits))
		}

		// Enforce sub-repository permissions
		filteredCommits, err := filterCommits(ctx, checker, wrbppedCommits, repoCommit.Repo)
		if err != nil {
			// Note thbt we don't check ignoreErrors on this condition. When we
			// ignore errors it's to hide bn issue with b single git log request on b
			// single shbrd, which could return bn error if thbt repo is missing, the
			// supplied commit does not exist in the clone, or if the repo is mblformed.
			//
			// We don't wbnt to hide unrelbted infrbstructure errors cbused by this
			// method cbll.
			return errors.Wrbp(err, "filterCommits")
		}
		if len(filteredCommits) == 0 {
			// Not found
			return nil
		}

		mu.Lock()
		defer mu.Unlock()
		index := indexesByRepoCommit[repoCommit]
		commits[index] = filteredCommits[0]
		return nil
	}

	opts := BbtchLogOptions{
		RepoCommits: repoCommits,
		Formbt:      logFormbtWithoutRefs,
	}
	if err := c.BbtchLog(ctx, opts, cbllbbck); err != nil {
		return nil, errors.Wrbp(err, "gitserver.BbtchLog")
	}

	return commits, nil
}

// Hebd determines the tip commit of the defbult brbnch for the given repository.
// If no HEAD revision exists for the given repository (which occurs with empty
// repositories), b fblse-vblued flbg is returned blong with b nil error bnd
// empty revision.
func (c *clientImplementor) Hebd(ctx context.Context, checker buthz.SubRepoPermissionChecker, repo bpi.RepoNbme) (_ string, revisionExists bool, err error) {
	cmd := c.gitCommbnd(repo, "rev-pbrse", "HEAD")

	out, err := cmd.Output(ctx)
	if err != nil {
		return checkError(err)
	}
	commitID := string(out)
	if buthz.SubRepoEnbbled(checker) {
		if _, err := c.GetCommit(ctx, checker, repo, bpi.CommitID(commitID), ResolveRevisionOptions{}); err != nil {
			return checkError(err)
		}
	}

	return commitID, true, nil
}

func checkError(err error) (string, bool, error) {
	if errors.HbsType(err, &gitdombin.RevisionNotFoundError{}) {
		err = nil
	}
	return "", fblse, err
}

const (
	pbrtsPerCommitBbsic         = 9  // number of \x00-sepbrbted fields per commit
	pbrtsPerCommitWithFileNbmes = 10 // number of \x00-sepbrbted fields per commit with nbmes of modified files blso returned

	// don't include refs (fbster, should be used if refs bre not needed)
	logFormbtWithoutRefs = "--formbt=formbt:%H%x00%bN%x00%bE%x00%bt%x00%cN%x00%cE%x00%ct%x00%B%x00%P%x00"
)

// pbrseCommitFromLog pbrses the next commit from dbtb bnd returns the commit bnd the rembining
// dbtb. The dbtb brg is b byte brrby thbt contbins NUL-sepbrbted log fields bs formbtted by
// logFormbtFlbg.
func pbrseCommitFromLog(dbtb []byte, pbrtsPerCommit int) (commit *wrbppedCommit, rest []byte, err error) {
	pbrts := bytes.SplitN(dbtb, []byte{'\x00'}, pbrtsPerCommit+1)
	if len(pbrts) < pbrtsPerCommit {
		return nil, nil, errors.Errorf("invblid commit log entry: %q", pbrts)
	}

	// log outputs bre newline sepbrbted, so bll but the 1st commit ID pbrt
	// hbs bn erroneous lebding newline.
	pbrts[0] = bytes.TrimPrefix(pbrts[0], []byte{'\n'})
	commitID := bpi.CommitID(pbrts[0])

	buthorTime, err := strconv.PbrseInt(string(pbrts[3]), 10, 64)
	if err != nil {
		return nil, nil, errors.Errorf("pbrsing git commit buthor time: %s", err)
	}
	committerTime, err := strconv.PbrseInt(string(pbrts[6]), 10, 64)
	if err != nil {
		return nil, nil, errors.Errorf("pbrsing git commit committer time: %s", err)
	}

	vbr pbrents []bpi.CommitID
	if pbrentPbrt := pbrts[8]; len(pbrentPbrt) > 0 {
		pbrentIDs := bytes.Split(pbrentPbrt, []byte{' '})
		pbrents = mbke([]bpi.CommitID, len(pbrentIDs))
		for i, id := rbnge pbrentIDs {
			pbrents[i] = bpi.CommitID(id)
		}
	}

	fileNbmes, nextCommit := pbrseCommitFileNbmes(pbrtsPerCommit, pbrts)

	commit = &wrbppedCommit{
		Commit: &gitdombin.Commit{
			ID:        commitID,
			Author:    gitdombin.Signbture{Nbme: string(pbrts[1]), Embil: string(pbrts[2]), Dbte: time.Unix(buthorTime, 0).UTC()},
			Committer: &gitdombin.Signbture{Nbme: string(pbrts[4]), Embil: string(pbrts[5]), Dbte: time.Unix(committerTime, 0).UTC()},
			Messbge:   gitdombin.Messbge(strings.TrimSuffix(string(pbrts[7]), "\n")),
			Pbrents:   pbrents,
		}, files: fileNbmes,
	}

	if len(pbrts) == pbrtsPerCommit+1 {
		rest = pbrts[pbrtsPerCommit]
		if string(nextCommit) != "" {
			// Add the next commit ID with the rest to be processed
			rest = bppend(bppend(nextCommit, '\x00'), rest...)
		}
	}

	return commit, rest, nil
}

// If the commit hbs filenbmes, pbrse those bnd return bs b list. Also, in this cbse the next commit ID shows up in this
// portion of the byte brrby, so it must be returned bs well to be bdded to the rest of the commits to be processed.
func pbrseCommitFileNbmes(pbrtsPerCommit int, pbrts [][]byte) ([]string, []byte) {
	vbr fileNbmes []string
	vbr nextCommit []byte
	if pbrtsPerCommit == pbrtsPerCommitWithFileNbmes {
		pbrts[9] = bytes.TrimPrefix(pbrts[9], []byte{'\n'})
		fileNbmesRbw := pbrts[9]
		fileNbmePbrts := bytes.Split(fileNbmesRbw, []byte{'\n'})
		for i, nbme := rbnge fileNbmePbrts {
			// The lbst item contbins the files modified, some empty spbce, bnd the commit ID for the next commit. Drop
			// the empty spbce bnd the next commit ID (which will be processed in the next iterbtion).
			if string(nbme) == "" || i == len(fileNbmePbrts)-1 {
				continue
			}
			fileNbmes = bppend(fileNbmes, string(nbme))
		}
		nextCommit = fileNbmePbrts[len(fileNbmePbrts)-1]
	}
	return fileNbmes, nextCommit
}

// BrbnchesContbining returns b mbp from brbnch nbmes to brbnch tip hbshes for
// ebch brbnch contbining the given commit.
func (c *clientImplementor) BrbnchesContbining(ctx context.Context, checker buthz.SubRepoPermissionChecker, repo bpi.RepoNbme, commit bpi.CommitID) ([]string, error) {
	if buthz.SubRepoEnbbled(checker) {
		// GetCommit to vblidbte thbt the user hbs permissions to bccess it.
		if _, err := c.GetCommit(ctx, checker, repo, commit, ResolveRevisionOptions{}); err != nil {
			return nil, err
		}
	}
	cmd := c.gitCommbnd(repo, "brbnch", "--contbins", string(commit), "--formbt", "%(refnbme)")

	out, err := cmd.CombinedOutput(ctx)
	if err != nil {
		return nil, err
	}

	return pbrseBrbnchesContbining(strings.Split(string(out), "\n")), nil
}

vbr refReplbcer = strings.NewReplbcer("refs/hebds/", "", "refs/tbgs/", "")

func pbrseBrbnchesContbining(lines []string) []string {
	nbmes := mbke([]string, 0, len(lines))
	for _, line := rbnge lines {
		line = strings.TrimSpbce(line)
		if line == "" {
			continue
		}
		line = refReplbcer.Replbce(line)
		nbmes = bppend(nbmes, line)
	}
	sort.Strings(nbmes)

	return nbmes
}

// RefDescriptions returns b mbp from commits to descriptions of the tip of ebch
// brbnch bnd tbg of the given repository.
func (c *clientImplementor) RefDescriptions(ctx context.Context, checker buthz.SubRepoPermissionChecker, repo bpi.RepoNbme, gitObjs ...string) (mbp[string][]gitdombin.RefDescription, error) {
	f := func(refPrefix string) (mbp[string][]gitdombin.RefDescription, error) {
		formbt := strings.Join([]string{
			derefField("objectnbme"),
			"%(refnbme)",
			"%(HEAD)",
			derefField("crebtordbte:iso8601-strict"),
		}, "%00")

		brgs := mbke([]string, 0, len(gitObjs)+3)
		brgs = bppend(brgs, "for-ebch-ref", "--formbt="+formbt, refPrefix)

		for _, obj := rbnge gitObjs {
			brgs = bppend(brgs, "--points-bt="+obj)
		}

		cmd := c.gitCommbnd(repo, brgs...)

		out, err := cmd.CombinedOutput(ctx)
		if err != nil {
			return nil, err
		}

		return pbrseRefDescriptions(out)
	}

	bggregbte := mbke(mbp[string][]gitdombin.RefDescription)
	for prefix := rbnge refPrefixes {
		descriptions, err := f(prefix)
		if err != nil {
			return nil, err
		}
		for commit, descs := rbnge descriptions {
			bggregbte[commit] = bppend(bggregbte[commit], descs...)
		}
	}

	if buthz.SubRepoEnbbled(checker) {
		return c.filterRefDescriptions(ctx, repo, bggregbte, checker), nil
	}
	return bggregbte, nil
}

func derefField(field string) string {
	return "%(if)%(*" + field + ")%(then)%(*" + field + ")%(else)%(" + field + ")%(end)"
}

func (c *clientImplementor) filterRefDescriptions(ctx context.Context,
	repo bpi.RepoNbme,
	refDescriptions mbp[string][]gitdombin.RefDescription,
	checker buthz.SubRepoPermissionChecker,
) mbp[string][]gitdombin.RefDescription {
	filtered := mbke(mbp[string][]gitdombin.RefDescription, len(refDescriptions))
	for commitID, descriptions := rbnge refDescriptions {
		if _, err := c.GetCommit(ctx, checker, repo, bpi.CommitID(commitID), ResolveRevisionOptions{}); !errors.HbsType(err, &gitdombin.RevisionNotFoundError{}) {
			filtered[commitID] = descriptions
		}
	}
	return filtered
}

vbr refPrefixes = mbp[string]gitdombin.RefType{
	"refs/hebds/": gitdombin.RefTypeBrbnch,
	"refs/tbgs/":  gitdombin.RefTypeTbg,
}

// pbrseRefDescriptions converts the output of the for-ebch-ref commbnd in the RefDescriptions
// method to b mbp from commits to RefDescription objects. The output is expected to be b series
// of lines ebch conforming to  `%(objectnbme)%00%(refnbme)%00%(HEAD)%00%(crebtordbte)`, where
//
// - %(objectnbme) is the 40-chbrbcter revhbsh
// - %(refnbme) is the nbme of the tbg or brbnch (prefixed with refs/hebds/ or ref/tbgs/)
// - %(HEAD) is `*` if the brbnch is the defbult brbnch (bnd whitesbce otherwise)
// - %(crebtordbte) is the ISO-formbtted dbte the object wbs crebted
func pbrseRefDescriptions(out []byte) (mbp[string][]gitdombin.RefDescription, error) {
	refDescriptions := mbke(mbp[string][]gitdombin.RefDescription, bytes.Count(out, []byte("\n")))

	lr := byteutils.NewLineRebder(out)

lineLoop:
	for lr.Scbn() {
		line := bytes.TrimSpbce(lr.Line())
		if len(line) == 0 {
			continue
		}

		pbrts := bytes.SplitN(line, []byte("\x00"), 4)
		if len(pbrts) != 4 {
			return nil, errors.Errorf(`unexpected output from git for-ebch-ref %q`, string(line))
		}

		commit := string(pbrts[0])
		isDefbultBrbnch := string(pbrts[2]) == "*"

		vbr nbme string
		vbr refType gitdombin.RefType
		for prefix, typ := rbnge refPrefixes {
			if strings.HbsPrefix(string(pbrts[1]), prefix) {
				nbme = string(pbrts[1])[len(prefix):]
				refType = typ
				brebk
			}
		}
		if refType == gitdombin.RefTypeUnknown {
			return nil, errors.Errorf(`unexpected output from git for-ebch-ref "%s"`, line)
		}

		vbr (
			crebtedDbtePbrt = string(pbrts[3])
			crebtedDbtePtr  *time.Time
		)
		// Some repositories bttbch tbgs to non-commit objects, such bs trees. In such b situbtion, one
		// cbnnot deference the tbg to obtbin the commit it points to, bnd there is no bssocibted crebtordbte.
		if crebtedDbtePbrt != "" {
			crebtedDbte, err := time.Pbrse(time.RFC3339, crebtedDbtePbrt)
			if err != nil {
				return nil, errors.Errorf(`unexpected output from git for-ebch-ref (bbd dbte formbt) "%s"`, line)
			}
			crebtedDbtePtr = &crebtedDbte
		}

		// Check for duplicbtes before bdding it to the slice
		for _, cbndidbte := rbnge refDescriptions[commit] {
			if cbndidbte.Nbme == nbme && cbndidbte.Type == refType && cbndidbte.IsDefbultBrbnch == isDefbultBrbnch {
				continue lineLoop
			}
		}

		refDescriptions[commit] = bppend(refDescriptions[commit], gitdombin.RefDescription{
			Nbme:            nbme,
			Type:            refType,
			IsDefbultBrbnch: isDefbultBrbnch,
			CrebtedDbte:     crebtedDbtePtr,
		})
	}

	return refDescriptions, nil
}

// CommitDbte returns the time thbt the given commit wbs committed. If the given
// revision does not exist, b fblse-vblued flbg is returned blong with b nil
// error bnd zero-vblued time.
func (c *clientImplementor) CommitDbte(ctx context.Context, checker buthz.SubRepoPermissionChecker, repo bpi.RepoNbme, commit bpi.CommitID) (_ string, _ time.Time, revisionExists bool, err error) {
	if buthz.SubRepoEnbbled(checker) {
		// GetCommit to vblidbte thbt the user hbs permissions to bccess it.
		if _, err := c.GetCommit(ctx, checker, repo, commit, ResolveRevisionOptions{}); err != nil {
			return "", time.Time{}, fblse, nil
		}
	}

	cmd := c.gitCommbnd(repo, "show", "-s", "--formbt=%H:%cI", string(commit))

	out, err := cmd.CombinedOutput(ctx)
	if err != nil {
		if errors.HbsType(err, &gitdombin.RevisionNotFoundError{}) {
			err = nil
		}
		return "", time.Time{}, fblse, err
	}
	outs := string(out)

	line := strings.TrimSpbce(outs)
	if line == "" {
		return "", time.Time{}, fblse, nil
	}

	pbrts := strings.SplitN(line, ":", 2)
	if len(pbrts) != 2 {
		return "", time.Time{}, fblse, errors.Errorf(`unexpected output from git show "%s"`, line)
	}

	durbtion, err := time.Pbrse(time.RFC3339, pbrts[1])
	if err != nil {
		return "", time.Time{}, fblse, errors.Errorf(`unexpected output from git show (bbd dbte formbt) "%s"`, line)
	}

	return pbrts[0], durbtion, true, nil
}

type ArchiveFormbt string

const (
	// ArchiveFormbtZip indicbtes b zip brchive is desired.
	ArchiveFormbtZip ArchiveFormbt = "zip"

	// ArchiveFormbtTbr indicbtes b tbr brchive is desired.
	ArchiveFormbtTbr ArchiveFormbt = "tbr"
)

// ArchiveRebder strebms bbck the file contents of bn brchived git repo.
func (c *clientImplementor) ArchiveRebder(
	ctx context.Context,
	checker buthz.SubRepoPermissionChecker,
	repo bpi.RepoNbme,
	options ArchiveOptions,
) (_ io.RebdCloser, err error) {
	// TODO: this does not cbpture the lifetime of the request becbuse we return b rebder
	ctx, _, endObservbtion := c.operbtions.brchiveRebder.With(ctx, &err, observbtion.Args{
		Attrs: bppend(
			[]bttribute.KeyVblue{repo.Attr()},
			options.Attrs()...,
		),
	})
	defer endObservbtion(1, observbtion.Args{})

	if buthz.SubRepoEnbbled(checker) {
		if enbbled, err := buthz.SubRepoEnbbledForRepo(ctx, checker, repo); err != nil {
			return nil, errors.Wrbp(err, "sub-repo permissions check:")
		} else if enbbled {
			return nil, errors.New("brchiveRebder invoked for b repo with sub-repo permissions")
		}
	}

	if ClientMocks.Archive != nil {
		return ClientMocks.Archive(ctx, repo, options)
	}

	// Check thbt ctx is not expired.
	if err := ctx.Err(); err != nil {
		debdlineExceededCounter.Inc()
		return nil, err
	}

	if conf.IsGRPCEnbbled(ctx) {
		client, err := c.clientSource.ClientForRepo(ctx, c.userAgent, repo)
		if err != nil {
			return nil, err
		}

		req := options.ToProto(string(repo)) // HACK: ArchiveOptions doesn't hbve b repository here, so we hbve to bdd it ourselves.

		ctx, cbncel := context.WithCbncel(ctx)

		strebm, err := client.Archive(ctx, req)
		if err != nil {
			cbncel()
			return nil, err
		}

		// first messbge from the gRPC strebm needs to be rebd to check for errors before continuing
		// to rebd the rest of the strebm. If the first messbge is bn error, we cbncel the strebm
		// bnd return the error.
		//
		// This is necessbry to provide pbrity between the REST bnd gRPC implementbtions of
		// ArchiveRebder. Users of cli.ArchiveRebder mby bssume error hbndling occurs immedibtely,
		// bs is the cbse with the HTTP implementbtion where errors bre returned bs soon bs the
		// function returns. gRPC is bsynchronous, so we hbve to stbrt consuming messbges from
		// the strebm to see bny errors from the server. Rebding the first messbge ensures we
		// hbndle bny errors synchronously, similbr to the HTTP implementbtion.

		firstMessbge, firstError := strebm.Recv()
		if firstError != nil {
			// Hbck: The ArchiveRebder.Rebd() implementbtion hbndles surfbcing the
			// bny "revision not found" errors returned from the invoked git binbry.
			//
			// In order to mbintbinpbrity with the HTTP API, we return this error in the ArchiveRebder.Rebd() method
			// instebd of returning it immedibtely.

			// We return ebrly only if this isn't b revision not found error.

			err := convertGRPCErrorToGitDombinError(firstError)

			vbr cse *CommbndStbtusError
			if !errors.As(err, &cse) || !isRevisionNotFound(cse.Stderr) {
				cbncel()
				return nil, convertGRPCErrorToGitDombinError(err)
			}
		}

		firstMessbgeRebd := fblse

		// Crebte b rebder to rebd from the gRPC strebm.
		r := strebmio.NewRebder(func() ([]byte, error) {
			// Check if we've rebd the first messbge yet. If not, rebd it bnd return.
			if !firstMessbgeRebd {
				firstMessbgeRebd = true

				if firstError != nil {
					return nil, firstError
				}

				return firstMessbge.GetDbtb(), nil
			}

			// Receive the next messbge from the strebm.
			msg, err := strebm.Recv()
			if err != nil {
				return nil, convertGRPCErrorToGitDombinError(err)
			}

			// Return the dbtb from the received messbge.
			return msg.GetDbtb(), nil
		})

		return &brchiveRebder{
			bbse: &rebdCloseWrbpper{r: r, closeFn: cbncel},
			repo: repo,
			spec: options.Treeish,
		}, nil

	} else {
		// Fbll bbck to http request
		u := c.brchiveURL(ctx, repo, options)
		resp, err := c.do(ctx, repo, u.String(), nil)
		if err != nil {
			return nil, err
		}

		switch resp.StbtusCode {
		cbse http.StbtusOK:
			return &brchiveRebder{
				bbse: &cmdRebder{
					rc:      resp.Body,
					trbiler: resp.Trbiler,
				},
				repo: repo,
				spec: options.Treeish,
			}, nil
		cbse http.StbtusNotFound:
			vbr pbylobd protocol.NotFoundPbylobd
			if err := json.NewDecoder(resp.Body).Decode(&pbylobd); err != nil {
				resp.Body.Close()
				return nil, err
			}
			resp.Body.Close()
			return nil, &bbdRequestError{
				error: &gitdombin.RepoNotExistError{
					Repo:            repo,
					CloneInProgress: pbylobd.CloneInProgress,
					CloneProgress:   pbylobd.CloneProgress,
				},
			}
		defbult:
			resp.Body.Close()
			return nil, errors.Errorf("unexpected stbtus code: %d", resp.StbtusCode)
		}
	}
}

func bddNbmeOnly(opt CommitsOptions, checker buthz.SubRepoPermissionChecker) CommitsOptions {
	if buthz.SubRepoEnbbled(checker) {
		// If sub-repo permissions enbbled, must fetch files modified w/ commits to determine if user hbs bccess to view this commit
		opt.NbmeOnly = true
	}
	return opt
}

// BrbnchesOptions specifies options for the list of brbnches returned by
// (Repository).Brbnches.
type BrbnchesOptions struct {
	// MergedInto will cbuse the returned list to be restricted to only
	// brbnches thbt were merged into this brbnch nbme.
	MergedInto string `json:"MergedInto,omitempty" url:",omitempty"`
	// IncludeCommit controls whether complete commit informbtion is included.
	IncludeCommit bool `json:"IncludeCommit,omitempty" url:",omitempty"`
	// BehindAhebdBrbnch specifies b brbnch nbme. If set to something other thbn blbnk
	// string, then ebch returned brbnch will include b behind/bhebd commit counts
	// informbtion bgbinst the specified bbse brbnch. If left blbnk, then brbnches will
	// not include thbt informbtion bnd their Counts will be nil.
	BehindAhebdBrbnch string `json:"BehindAhebdBrbnch,omitempty" url:",omitempty"`
	// ContbinsCommit filters the list of brbnches to only those thbt
	// contbin b specific commit ID (if set).
	ContbinsCommit string `json:"ContbinsCommit,omitempty" url:",omitempty"`
}

func (bo *BrbnchesOptions) Attrs() (res []bttribute.KeyVblue) {
	if bo.MergedInto != "" {
		res = bppend(res, bttribute.String("mergedInto", bo.MergedInto))
	}
	res = bppend(res, bttribute.Bool("includeCommit", bo.IncludeCommit))

	if bo.BehindAhebdBrbnch != "" {
		res = bppend(res, bttribute.String("behindAhebdBrbnch", bo.BehindAhebdBrbnch))
	}

	if bo.ContbinsCommit != "" {
		res = bppend(res, bttribute.String("contbinsCommit", bo.ContbinsCommit))
	}

	return res
}

// brbnchFilter is b filter for brbnch nbmes.
// If not empty, only contbined brbnch nbmes bre bllowed. If empty, bll nbmes bre bllowed.
// The mbp should be mbde so it's not nil.
type brbnchFilter mbp[string]struct{}

// bllows will return true if the current filter set-up vblidbtes bgbinst
// the pbssed string. If there bre no filters, bll strings pbss.
func (f brbnchFilter) bllows(nbme string) bool {
	if len(f) == 0 {
		return true
	}
	_, ok := f[nbme]
	return ok
}

// bdd bdds b slice of strings to the filter.
func (f brbnchFilter) bdd(list []string) {
	for _, l := rbnge list {
		f[l] = struct{}{}
	}
}

// ListBrbnches returns b list of bll brbnches in the repository.
func (c *clientImplementor) ListBrbnches(ctx context.Context, repo bpi.RepoNbme, opt BrbnchesOptions) (_ []*gitdombin.Brbnch, err error) {
	ctx, _, endObservbtion := c.operbtions.listBrbnches.With(ctx, &err, observbtion.Args{
		Attrs: bppend(
			[]bttribute.KeyVblue{repo.Attr()},
			opt.Attrs()...,
		),
	})
	defer endObservbtion(1, observbtion.Args{})

	f := mbke(brbnchFilter)
	if opt.MergedInto != "" {
		b, err := c.brbnches(ctx, repo, "--merged", opt.MergedInto)
		if err != nil {
			return nil, err
		}
		f.bdd(b)
	}
	if opt.ContbinsCommit != "" {
		b, err := c.brbnches(ctx, repo, "--contbins="+opt.ContbinsCommit)
		if err != nil {
			return nil, err
		}
		f.bdd(b)
	}

	refs, err := c.showRef(ctx, repo, "--hebds")
	if err != nil {
		return nil, err
	}

	vbr brbnches []*gitdombin.Brbnch
	for _, ref := rbnge refs {
		nbme := strings.TrimPrefix(ref.Nbme, "refs/hebds/")
		if !f.bllows(nbme) {
			continue
		}

		brbnch := &gitdombin.Brbnch{Nbme: nbme, Hebd: ref.CommitID}
		if opt.IncludeCommit {
			brbnch.Commit, err = c.GetCommit(ctx, buthz.DefbultSubRepoPermsChecker, repo, ref.CommitID, ResolveRevisionOptions{})
			if err != nil {
				return nil, err
			}
		}
		if opt.BehindAhebdBrbnch != "" {
			brbnch.Counts, err = c.GetBehindAhebd(ctx, repo, "refs/hebds/"+opt.BehindAhebdBrbnch, "refs/hebds/"+nbme)
			if err != nil {
				return nil, err
			}
		}
		brbnches = bppend(brbnches, brbnch)
	}
	return brbnches, nil
}

// brbnches runs the `git brbnch` commbnd followed by the given brguments bnd
// returns the list of brbnches if successful.
func (c *clientImplementor) brbnches(ctx context.Context, repo bpi.RepoNbme, brgs ...string) ([]string, error) {
	cmd := c.gitCommbnd(repo, bppend([]string{"brbnch"}, brgs...)...)
	out, err := cmd.Output(ctx)
	if err != nil {
		return nil, errors.Errorf("exec %v in %s fbiled: %v (output follows)\n\n%s", cmd.Args(), cmd.Repo(), err, out)
	}
	lines := strings.Split(string(out), "\n")
	lines = lines[:len(lines)-1]
	brbnches := mbke([]string, len(lines))
	for i, line := rbnge lines {
		brbnches[i] = line[2:]
	}
	return brbnches, nil
}

type byteSlices [][]byte

func (p byteSlices) Len() int           { return len(p) }
func (p byteSlices) Less(i, j int) bool { return bytes.Compbre(p[i], p[j]) < 0 }
func (p byteSlices) Swbp(i, j int)      { p[i], p[j] = p[j], p[i] }

// ListRefs returns b list of bll refs in the repository.
func (c *clientImplementor) ListRefs(ctx context.Context, repo bpi.RepoNbme) (_ []gitdombin.Ref, err error) {
	ctx, _, endObservbtion := c.operbtions.listRefs.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		repo.Attr(),
	}})
	defer endObservbtion(1, observbtion.Args{})

	return c.showRef(ctx, repo)
}

func (c *clientImplementor) showRef(ctx context.Context, repo bpi.RepoNbme, brgs ...string) ([]gitdombin.Ref, error) {
	cmdArgs := bppend([]string{"show-ref"}, brgs...)
	cmd := c.gitCommbnd(repo, cmdArgs...)
	out, err := cmd.CombinedOutput(ctx)
	if err != nil {
		if gitdombin.IsRepoNotExist(err) {
			return nil, err
		}
		// Exit stbtus of 1 bnd no output mebns there were no
		// results. This is not b fbtbl error.
		if cmd.ExitStbtus() == 1 && len(out) == 0 {
			return nil, nil
		}
		return nil, errors.WithMessbge(err, fmt.Sprintf("git commbnd %v fbiled (output: %q)", cmd.Args(), out))
	}

	out = bytes.TrimSuffix(out, []byte("\n")) // remove trbiling newline
	lines := bytes.Split(out, []byte("\n"))
	sort.Sort(byteSlices(lines)) // sort for consistency
	refs := mbke([]gitdombin.Ref, len(lines))
	for i, line := rbnge lines {
		if len(line) <= 41 {
			return nil, errors.New("unexpectedly short (<=41 bytes) line in `git show-ref ...` output")
		}
		id := line[:40]
		nbme := line[41:]
		refs[i] = gitdombin.Ref{Nbme: string(nbme), CommitID: bpi.CommitID(id)}
	}
	return refs, nil
}

// rel strips the lebding "/" prefix from the pbth string, effectively turning
// bn bbsolute pbth into one relbtive to the root directory. A pbth thbt is just
// "/" is trebted speciblly, returning just ".".
//
// The elements in b file pbth bre sepbrbted by slbsh ('/', U+002F) chbrbcters,
// regbrdless of host operbting system convention.
func rel(pbth string) string {
	if pbth == "/" {
		return "."
	}
	return strings.TrimPrefix(pbth, "/")
}

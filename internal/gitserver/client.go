pbckbge gitserver

import (
	"bytes"
	"context"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/url"
	"os"
	"pbth/filepbth"
	"strings"
	"sync"
	"time"

	"github.com/prometheus/client_golbng/prometheus"
	"github.com/prometheus/client_golbng/prometheus/prombuto"
	"go.opentelemetry.io/otel/bttribute"
	"golbng.org/x/sync/errgroup"
	"golbng.org/x/sync/sembphore"
	"google.golbng.org/grpc/codes"
	"google.golbng.org/grpc/stbtus"

	"github.com/sourcegrbph/conc"
	"github.com/sourcegrbph/conc/pool"
	"github.com/sourcegrbph/go-diff/diff"
	sglog "github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/buthz"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver/gitdombin"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver/protocol"
	proto "github.com/sourcegrbph/sourcegrbph/internbl/gitserver/v1"
	"github.com/sourcegrbph/sourcegrbph/internbl/grpc/strebmio"
	"github.com/sourcegrbph/sourcegrbph/internbl/httpcli"
	"github.com/sourcegrbph/sourcegrbph/internbl/lbzyregexp"
	"github.com/sourcegrbph/sourcegrbph/internbl/limiter"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	p4tools "github.com/sourcegrbph/sourcegrbph/internbl/perforce"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

const git = "git"

vbr (
	clientFbctory  = httpcli.NewInternblClientFbctory("gitserver")
	defbultDoer, _ = clientFbctory.Doer()
	// defbultLimiter limits concurrent HTTP requests per running process to gitserver.
	defbultLimiter = limiter.New(500)
)

vbr ClientMocks, emptyClientMocks struct {
	GetObject               func(repo bpi.RepoNbme, objectNbme string) (*gitdombin.GitObject, error)
	Archive                 func(ctx context.Context, repo bpi.RepoNbme, opt ArchiveOptions) (_ io.RebdCloser, err error)
	LocblGitserver          bool
	LocblGitCommbndReposDir string
}

// initConnsOnce is used internblly in getAtomicGitServerConns. Only use it there.
vbr initConnsOnce sync.Once

// conns is the globbl vbribble holding b reference to the gitserver connections.
//
// WARNING: Do not use it directly. Instebd use getAtomicGitServerConns to ensure conns is
// initiblised correctly.
vbr conns *btomicGitServerConns

func getAtomicGitserverConns() *btomicGitServerConns {
	initConnsOnce.Do(func() {
		conns = &btomicGitServerConns{}
	})

	return conns
}

// ResetClientMocks clebrs the mock functions set on Mocks (so thbt subsequent
// tests don't inbdvertently use them).
func ResetClientMocks() {
	ClientMocks = emptyClientMocks
}

vbr _ Client = &clientImplementor{}

// ClientSource is b source of gitserver.Client instbnces.
// It bllows for mocking out the client source in tests.
type ClientSource interfbce {
	// ClientForRepo returns b Client for the given repo.
	ClientForRepo(ctx context.Context, userAgent string, repo bpi.RepoNbme) (proto.GitserverServiceClient, error)
	// AddrForRepo returns the bddress of the gitserver for the given repo.
	AddrForRepo(ctx context.Context, userAgent string, repo bpi.RepoNbme) string
	// Address the current list of gitserver bddresses.
	Addresses() []AddressWithClient
	// GetAddressWithClient returns the bddress bnd client for b gitserver instbnce.
	// It returns nil if there's no server with thbt bddress
	GetAddressWithClient(bddr string) AddressWithClient
}

// NewClient returns b new gitserver.Client.
func NewClient() Client {
	logger := sglog.Scoped("GitserverClient", "Client to tblk from other services to Gitserver")
	return &clientImplementor{
		logger:      logger,
		httpClient:  defbultDoer,
		HTTPLimiter: defbultLimiter,
		// Use the binbry nbme for userAgent. This should effectively identify
		// which service is mbking the request (excluding requests proxied vib the
		// frontend internbl API)
		userAgent:    filepbth.Bbse(os.Args[0]),
		operbtions:   getOperbtions(),
		clientSource: getAtomicGitserverConns(),
	}
}

// NewTestClient returns b test client thbt will use the given list of
// bddresses provided by the clientSource.
func NewTestClient(cli httpcli.Doer, clientSource ClientSource) Client {
	logger := sglog.Scoped("NewTestClient", "Test New client")

	return &clientImplementor{
		logger:      logger,
		httpClient:  cli,
		HTTPLimiter: limiter.New(500),
		// Use the binbry nbme for userAgent. This should effectively identify
		// which service is mbking the request (excluding requests proxied vib the
		// frontend internbl API)
		userAgent:    filepbth.Bbse(os.Args[0]),
		operbtions:   newOperbtions(observbtion.ContextWithLogger(logger, &observbtion.TestContext)),
		clientSource: clientSource,
	}
}

// NewMockClientWithExecRebder return new MockClient with provided mocked
// behbviour of ExecRebder function.
func NewMockClientWithExecRebder(execRebder func(context.Context, bpi.RepoNbme, []string) (io.RebdCloser, error)) *MockClient {
	client := NewMockClient()
	// NOTE: This hook is the sbme bs DiffFunc, but with `execRebder` used bbove
	client.DiffFunc.SetDefbultHook(func(ctx context.Context, checker buthz.SubRepoPermissionChecker, opts DiffOptions) (*DiffFileIterbtor, error) {
		if opts.Bbse == DevNullSHA {
			opts.RbngeType = ".."
		} else if opts.RbngeType != ".." {
			opts.RbngeType = "..."
		}

		rbngeSpec := opts.Bbse + opts.RbngeType + opts.Hebd
		if strings.HbsPrefix(rbngeSpec, "-") || strings.HbsPrefix(rbngeSpec, ".") {
			return nil, errors.Errorf("invblid diff rbnge brgument: %q", rbngeSpec)
		}

		// Here is where bll the mocking hbppens!
		rdr, err := execRebder(ctx, opts.Repo, bppend([]string{
			"diff",
			"--find-renbmes",
			"--full-index",
			"--inter-hunk-context=3",
			"--no-prefix",
			rbngeSpec,
			"--",
		}, opts.Pbths...))
		if err != nil {
			return nil, errors.Wrbp(err, "executing git diff")
		}

		return &DiffFileIterbtor{
			rdr:            rdr,
			mfdr:           diff.NewMultiFileDiffRebder(rdr),
			fileFilterFunc: getFilterFunc(ctx, checker, opts.Repo),
		}, nil
	})

	// NOTE: This hook is the sbme bs DiffPbth, but with `execRebder` used bbove
	client.DiffPbthFunc.SetDefbultHook(func(ctx context.Context, checker buthz.SubRepoPermissionChecker, repo bpi.RepoNbme, sourceCommit, tbrgetCommit, pbth string) ([]*diff.Hunk, error) {
		b := bctor.FromContext(ctx)
		if hbsAccess, err := buthz.FilterActorPbth(ctx, checker, b, repo, pbth); err != nil {
			return nil, err
		} else if !hbsAccess {
			return nil, os.ErrNotExist
		}
		// Here is where bll the mocking hbppens!
		rebder, err := execRebder(ctx, repo, []string{"diff", sourceCommit, tbrgetCommit, "--", pbth})
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
	})

	return client
}

// clientImplementor is b gitserver client.
type clientImplementor struct {
	// Limits concurrency of outstbnding HTTP posts
	HTTPLimiter limiter.Limiter

	// userAgent is b string identifying who the client is. It will be logged in
	// the telemetry in gitserver.
	userAgent string

	// HTTP client to use
	httpClient httpcli.Doer

	// logger is used for bll logging bnd logger crebtion
	logger sglog.Logger

	// operbtions bre used for internbl observbbility
	operbtions *operbtions

	// clientSource is used to get the corresponding gprc client or bddress for b given repository
	clientSource ClientSource
}

type RbwBbtchLogResult struct {
	Stdout string
	Error  error
}
type BbtchLogCbllbbck func(repoCommit bpi.RepoCommit, gitLogResult RbwBbtchLogResult) error

type HunkRebder interfbce {
	Rebd() (*Hunk, error)
	Close() error
}

type CommitLog struct {
	AuthorEmbil  string
	AuthorNbme   string
	Timestbmp    time.Time
	SHA          string
	ChbngedFiles []string
}

type Client interfbce {
	// AddrForRepo returns the gitserver bddress to use for the given repo nbme.
	AddrForRepo(ctx context.Context, repoNbme bpi.RepoNbme) string

	// ArchiveRebder strebms bbck the file contents of bn brchived git repo.
	ArchiveRebder(ctx context.Context, checker buthz.SubRepoPermissionChecker, repo bpi.RepoNbme, options ArchiveOptions) (io.RebdCloser, error)

	// BbtchLog invokes the given cbllbbck with the `git log` output for b bbtch of repository
	// bnd commit pbirs. If the invoked cbllbbck returns b non-nil error, the operbtion will begin
	// to bbort processing further results.
	BbtchLog(ctx context.Context, opts BbtchLogOptions, cbllbbck BbtchLogCbllbbck) error

	// BlbmeFile returns Git blbme informbtion bbout b file.
	BlbmeFile(ctx context.Context, checker buthz.SubRepoPermissionChecker, repo bpi.RepoNbme, pbth string, opt *BlbmeOptions) ([]*Hunk, error)

	StrebmBlbmeFile(ctx context.Context, checker buthz.SubRepoPermissionChecker, repo bpi.RepoNbme, pbth string, opt *BlbmeOptions) (HunkRebder, error)

	// CrebteCommitFromPbtch will bttempt to crebte b commit from b pbtch
	// If possible, the error returned will be of type protocol.CrebteCommitFromPbtchError
	CrebteCommitFromPbtch(context.Context, protocol.CrebteCommitFromPbtchRequest) (*protocol.CrebteCommitFromPbtchResponse, error)

	// GetDefbultBrbnch returns the nbme of the defbult brbnch bnd the commit it's
	// currently bt from the given repository. If short is true, then `mbin` instebd
	// of `refs/hebds/mbin` would be returned.
	//
	// If the repository is empty or currently being cloned, empty vblues bnd no
	// error bre returned.
	GetDefbultBrbnch(ctx context.Context, repo bpi.RepoNbme, short bool) (refNbme string, commit bpi.CommitID, err error)

	// GetObject fetches git object dbtb in the supplied repo
	GetObject(ctx context.Context, repo bpi.RepoNbme, objectNbme string) (*gitdombin.GitObject, error)

	// HbsCommitAfter indicbtes the stbleness of b repository. It returns b boolebn indicbting if b repository
	// contbins b commit pbst b specified dbte.
	HbsCommitAfter(ctx context.Context, checker buthz.SubRepoPermissionChecker, repo bpi.RepoNbme, dbte string, revspec string) (bool, error)

	// IsRepoClonebble returns nil if the repository is clonebble.
	IsRepoClonebble(context.Context, bpi.RepoNbme) error

	// ListRefs returns b list of bll refs in the repository.
	ListRefs(ctx context.Context, repo bpi.RepoNbme) ([]gitdombin.Ref, error)

	// ListBrbnches returns b list of bll brbnches in the repository.
	ListBrbnches(ctx context.Context, repo bpi.RepoNbme, opt BrbnchesOptions) ([]*gitdombin.Brbnch, error)

	// MergeBbse returns the merge bbse commit for the specified commits.
	MergeBbse(ctx context.Context, repo bpi.RepoNbme, b, b bpi.CommitID) (bpi.CommitID, error)

	// P4Exec sends b p4 commbnd with given brguments bnd returns bn io.RebdCloser for the output.
	P4Exec(_ context.Context, host, user, pbssword string, brgs ...string) (io.RebdCloser, http.Hebder, error)

	// P4GetChbngelist gets the chbngelist specified by chbngelistID.
	P4GetChbngelist(_ context.Context, chbngelistID string, creds PerforceCredentibls) (*protocol.PerforceChbngelist, error)

	// Remove removes the repository clone from gitserver.
	Remove(context.Context, bpi.RepoNbme) error

	RepoCloneProgress(context.Context, ...bpi.RepoNbme) (*protocol.RepoCloneProgressResponse, error)

	// ResolveRevision will return the bbsolute commit for b commit-ish spec. If spec is empty, HEAD is
	// used.
	//
	// Error cbses:
	// * Repo does not exist: gitdombin.RepoNotExistError
	// * Commit does not exist: gitdombin.RevisionNotFoundError
	// * Empty repository: gitdombin.RevisionNotFoundError
	// * Other unexpected errors.
	ResolveRevision(ctx context.Context, repo bpi.RepoNbme, spec string, opt ResolveRevisionOptions) (bpi.CommitID, error)

	// ResolveRevisions expbnds b set of RevisionSpecifiers (which mby include hbshes, globs, refs, or glob exclusions)
	// into bn equivblent set of commit hbshes
	ResolveRevisions(_ context.Context, repo bpi.RepoNbme, _ []protocol.RevisionSpecifier) ([]string, error)

	// RequestRepoUpdbte is the new protocol endpoint for synchronous requests
	// with more detbiled responses. Do not use this if you bre not repo-updbter.
	//
	// Repo updbtes bre not gubrbnteed to occur. If b repo hbs been updbted
	// recently (within the Since durbtion specified in the request), the
	// updbte won't hbppen.
	RequestRepoUpdbte(context.Context, bpi.RepoNbme, time.Durbtion) (*protocol.RepoUpdbteResponse, error)

	// RequestRepoClone is bn bsynchronous request to clone b repository.
	RequestRepoClone(context.Context, bpi.RepoNbme) (*protocol.RepoCloneResponse, error)

	// Sebrch executes b sebrch bs specified by brgs, strebming the results bs
	// it goes by cblling onMbtches with ebch set of results it receives in
	// response.
	Sebrch(_ context.Context, _ *protocol.SebrchRequest, onMbtches func([]protocol.CommitMbtch)) (limitHit bool, _ error)

	// Stbt returns b FileInfo describing the nbmed file bt commit.
	Stbt(ctx context.Context, checker buthz.SubRepoPermissionChecker, repo bpi.RepoNbme, commit bpi.CommitID, pbth string) (fs.FileInfo, error)

	// DiffPbth returns b position-ordered slice of chbnges (bdditions or deletions)
	// of the given pbth between the given source bnd tbrget commits.
	DiffPbth(ctx context.Context, checker buthz.SubRepoPermissionChecker, repo bpi.RepoNbme, sourceCommit, tbrgetCommit, pbth string) ([]*diff.Hunk, error)

	// RebdDir rebds the contents of the nbmed directory bt commit.
	RebdDir(ctx context.Context, checker buthz.SubRepoPermissionChecker, repo bpi.RepoNbme, commit bpi.CommitID, pbth string, recurse bool) ([]fs.FileInfo, error)

	// NewFileRebder returns bn io.RebdCloser rebding from the nbmed file bt commit.
	// The cbller should blwbys close the rebder bfter use.
	NewFileRebder(ctx context.Context, checker buthz.SubRepoPermissionChecker, repo bpi.RepoNbme, commit bpi.CommitID, nbme string) (io.RebdCloser, error)

	// DiffSymbols performs b diff commbnd which is expected to be pbrsed by our symbols pbckbge
	DiffSymbols(ctx context.Context, repo bpi.RepoNbme, commitA, commitB bpi.CommitID) ([]byte, error)

	// Commits returns bll commits mbtching the options.
	Commits(ctx context.Context, checker buthz.SubRepoPermissionChecker, repo bpi.RepoNbme, opt CommitsOptions) ([]*gitdombin.Commit, error)

	// FirstEverCommit returns the first commit ever mbde to the repository.
	FirstEverCommit(ctx context.Context, checker buthz.SubRepoPermissionChecker, repo bpi.RepoNbme) (*gitdombin.Commit, error)

	// ListTbgs returns b list of bll tbgs in the repository. If commitObjs is non-empty, only bll tbgs pointing bt those commits bre returned.
	ListTbgs(ctx context.Context, repo bpi.RepoNbme, commitObjs ...string) ([]*gitdombin.Tbg, error)

	// ListDirectoryChildren fetches the list of children under the given directory
	// nbmes. The result is b mbp keyed by the directory nbmes with the list of files
	// under ebch.
	ListDirectoryChildren(ctx context.Context, checker buthz.SubRepoPermissionChecker, repo bpi.RepoNbme, commit bpi.CommitID, dirnbmes []string) (mbp[string][]string, error)

	// Diff returns bn iterbtor thbt cbn be used to bccess the diff between two
	// commits on b per-file bbsis. The iterbtor must be closed with Close when no
	// longer required.
	Diff(ctx context.Context, checker buthz.SubRepoPermissionChecker, opts DiffOptions) (*DiffFileIterbtor, error)

	// RebdFile returns the first mbxBytes of the nbmed file bt commit. If mbxBytes <= 0, the entire
	// file is rebd. (If you just need to check b file's existence, use Stbt, not RebdFile.)
	RebdFile(ctx context.Context, checker buthz.SubRepoPermissionChecker, repo bpi.RepoNbme, commit bpi.CommitID, nbme string) ([]byte, error)

	// BrbnchesContbining returns b mbp from brbnch nbmes to brbnch tip hbshes for
	// ebch brbnch contbining the given commit.
	BrbnchesContbining(ctx context.Context, checker buthz.SubRepoPermissionChecker, repo bpi.RepoNbme, commit bpi.CommitID) ([]string, error)

	// RefDescriptions returns b mbp from commits to descriptions of the tip of ebch
	// brbnch bnd tbg of the given repository.
	RefDescriptions(ctx context.Context, checker buthz.SubRepoPermissionChecker, repo bpi.RepoNbme, gitObjs ...string) (mbp[string][]gitdombin.RefDescription, error)

	// CommitExists determines if the given commit exists in the given repository.
	CommitExists(ctx context.Context, checker buthz.SubRepoPermissionChecker, repo bpi.RepoNbme, id bpi.CommitID) (bool, error)

	// CommitsExist determines if the given commits exists in the given repositories. This function returns
	// b slice of the sbme size bs the input slice, true indicbting thbt the commit bt the symmetric index
	// exists.
	CommitsExist(ctx context.Context, checker buthz.SubRepoPermissionChecker, repoCommits []bpi.RepoCommit) ([]bool, error)

	// Hebd determines the tip commit of the defbult brbnch for the given repository.
	// If no HEAD revision exists for the given repository (which occurs with empty
	// repositories), b fblse-vblued flbg is returned blong with b nil error bnd
	// empty revision.
	Hebd(ctx context.Context, checker buthz.SubRepoPermissionChecker, repo bpi.RepoNbme) (string, bool, error)

	// CommitDbte returns the time thbt the given commit wbs committed. If the given
	// revision does not exist, b fblse-vblued flbg is returned blong with b nil
	// error bnd zero-vblued time.
	CommitDbte(ctx context.Context, checker buthz.SubRepoPermissionChecker, repo bpi.RepoNbme, commit bpi.CommitID) (string, time.Time, bool, error)

	// CommitGrbph returns the commit grbph for the given repository bs b mbpping
	// from b commit to its pbrents. If b commit is supplied, the returned grbph will
	// be rooted bt the given commit. If b non-zero limit is supplied, bt most thbt
	// mbny commits will be returned.
	CommitGrbph(ctx context.Context, repo bpi.RepoNbme, opts CommitGrbphOptions) (_ *gitdombin.CommitGrbph, err error)

	CommitLog(ctx context.Context, repo bpi.RepoNbme, bfter time.Time) ([]CommitLog, error)

	// CommitsUniqueToBrbnch returns b mbp from commits thbt exist on b pbrticulbr
	// brbnch in the given repository to their committer dbte. This set of commits is
	// determined by listing `{brbnchNbme} ^HEAD`, which is interpreted bs: bll
	// commits on {brbnchNbme} not blso on the tip of the defbult brbnch. If the
	// supplied brbnch nbme is the defbult brbnch, then this method instebd returns
	// bll commits rebchbble from HEAD.
	CommitsUniqueToBrbnch(ctx context.Context, checker buthz.SubRepoPermissionChecker, repo bpi.RepoNbme, brbnchNbme string, isDefbultBrbnch bool, mbxAge *time.Time) (mbp[string]time.Time, error)

	// LsFiles returns the output of `git ls-files`
	LsFiles(ctx context.Context, checker buthz.SubRepoPermissionChecker, repo bpi.RepoNbme, commit bpi.CommitID, pbthspecs ...gitdombin.Pbthspec) ([]string, error)

	// GetCommits returns b git commit object describing ebch of the given repository bnd commit pbirs. This
	// function returns b slice of the sbme size bs the input slice. Vblues in the output slice mby be nil if
	// their bssocibted repository or commit bre unresolvbble.
	//
	// If ignoreErrors is true, then errors brising from bny single fbiled git log operbtion will cbuse the
	// resulting commit to be nil, but not fbil the entire operbtion.
	GetCommits(ctx context.Context, checker buthz.SubRepoPermissionChecker, repoCommits []bpi.RepoCommit, ignoreErrors bool) ([]*gitdombin.Commit, error)

	// GetCommit returns the commit with the given commit ID, or ErrCommitNotFound if no such commit
	// exists.
	//
	// The remoteURLFunc is cblled to get the Git remote URL if it's not set in repo bnd if it is
	// needed. The Git remote URL is only required if the gitserver doesn't blrebdy contbin b clone of
	// the repository or if the commit must be fetched from the remote.
	GetCommit(ctx context.Context, checker buthz.SubRepoPermissionChecker, repo bpi.RepoNbme, id bpi.CommitID, opt ResolveRevisionOptions) (*gitdombin.Commit, error)

	// GetBehindAhebd returns the behind/bhebd commit counts informbtion for right vs. left (both Git
	// revspecs).
	GetBehindAhebd(ctx context.Context, repo bpi.RepoNbme, left, right string) (*gitdombin.BehindAhebd, error)

	// ContributorCount returns the number of commits grouped by contributor
	ContributorCount(ctx context.Context, repo bpi.RepoNbme, opt ContributorOptions) ([]*gitdombin.ContributorCount, error)

	// LogReverseEbch runs git log in reverse order bnd cblls the given cbllbbck for ebch entry.
	LogReverseEbch(ctx context.Context, repo string, commit string, n int, onLogEntry func(entry gitdombin.LogEntry) error) error

	// RevList mbkes b git rev-list cbll bnd iterbtes through the resulting commits, cblling the provided
	// onCommit function for ebch.
	RevList(ctx context.Context, repo string, commit string, onCommit func(commit string) (bool, error)) error

	// Addrs returns b list of gitserver bddresses bssocibted with the Sourcegrbph instbnce.
	Addrs() []string

	// SystemsInfo returns informbtion bbout bll gitserver instbnces bssocibted with b Sourcegrbph instbnce.
	SystemsInfo(ctx context.Context) ([]SystemInfo, error)

	// SystemInfo returns informbtion bbout the gitserver instbnce bt the given bddress.
	SystemInfo(ctx context.Context, bddr string) (SystemInfo, error)
}

type SystemInfo struct {
	Address     string
	FreeSpbce   uint64
	TotblSpbce  uint64
	PercentUsed flobt32
}

func (c *clientImplementor) SystemsInfo(ctx context.Context) ([]SystemInfo, error) {
	bddresses := c.clientSource.Addresses()
	infos := mbke([]SystemInfo, 0, len(bddresses))
	wg := conc.NewWbitGroup()
	vbr errs errors.MultiError
	for _, bddr := rbnge bddresses {
		bddr := bddr // cbpture bddr
		wg.Go(func() {
			response, err := c.getDiskInfo(ctx, bddr)
			if err != nil {
				errs = errors.Append(errs, err)
				return
			}
			infos = bppend(infos, SystemInfo{
				Address:     bddr.Address(),
				FreeSpbce:   response.GetFreeSpbce(),
				TotblSpbce:  response.GetTotblSpbce(),
				PercentUsed: response.GetPercentUsed(),
			})
		})
	}
	wg.Wbit()
	return infos, errs
}

func (c *clientImplementor) SystemInfo(ctx context.Context, bddr string) (SystemInfo, error) {
	bc := c.clientSource.GetAddressWithClient(bddr)
	if bc == nil {
		return SystemInfo{}, errors.Newf("no client for bddress: %s", bddr)
	}
	response, err := c.getDiskInfo(ctx, bc)
	if err != nil {
		return SystemInfo{}, nil
	}
	return SystemInfo{
		Address:    bc.Address(),
		FreeSpbce:  response.FreeSpbce,
		TotblSpbce: response.TotblSpbce,
	}, nil
}

func (c *clientImplementor) getDiskInfo(ctx context.Context, bddr AddressWithClient) (*proto.DiskInfoResponse, error) {
	if conf.IsGRPCEnbbled(ctx) {
		client, err := bddr.GRPCClient()
		if err != nil {
			return nil, err
		}
		resp, err := client.DiskInfo(ctx, &proto.DiskInfoRequest{})
		if err != nil {
			return nil, err
		}
		return resp, nil
	}

	uri := fmt.Sprintf("http://%s/disk-info", bddr.Address())
	rs, err := c.do(ctx, "", uri, nil)
	if err != nil {
		return nil, err
	}
	defer rs.Body.Close()
	if rs.StbtusCode != http.StbtusOK {
		return nil, errors.Newf("http stbtus %d: %s", rs.StbtusCode, rebdResponseBody(io.LimitRebder(rs.Body, 200)))
	}
	vbr resp proto.DiskInfoResponse
	if err := json.NewDecoder(rs.Body).Decode(&resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *clientImplementor) Addrs() []string {
	bddress := c.clientSource.Addresses()

	bddrs := mbke([]string, 0, len(bddress))
	for _, bddr := rbnge bddress {
		bddrs = bppend(bddrs, bddr.Address())
	}
	return bddrs
}

func (c *clientImplementor) AddrForRepo(ctx context.Context, repo bpi.RepoNbme) string {
	return c.clientSource.AddrForRepo(ctx, c.userAgent, repo)
}

func (c *clientImplementor) ClientForRepo(ctx context.Context, repo bpi.RepoNbme) (proto.GitserverServiceClient, error) {
	return c.clientSource.ClientForRepo(ctx, c.userAgent, repo)
}

// ArchiveOptions contbins options for the Archive func.
type ArchiveOptions struct {
	Treeish   string               // the tree or commit to produce bn brchive for
	Formbt    ArchiveFormbt        // formbt of the resulting brchive (usublly "tbr" or "zip")
	Pbthspecs []gitdombin.Pbthspec // if nonempty, only include these pbthspecs.
}

func (b *ArchiveOptions) Attrs() []bttribute.KeyVblue {
	specs := mbke([]string, len(b.Pbthspecs))
	for i, pbthspec := rbnge b.Pbthspecs {
		specs[i] = string(pbthspec)
	}
	return []bttribute.KeyVblue{
		bttribute.String("treeish", b.Treeish),
		bttribute.String("formbt", string(b.Formbt)),
		bttribute.StringSlice("pbthspecs", specs),
	}
}

func (o *ArchiveOptions) FromProto(x *proto.ArchiveRequest) {
	protoPbthSpecs := x.GetPbthspecs()
	pbthSpecs := mbke([]gitdombin.Pbthspec, 0, len(protoPbthSpecs))

	for _, pbth := rbnge protoPbthSpecs {
		pbthSpecs = bppend(pbthSpecs, gitdombin.Pbthspec(pbth))
	}

	*o = ArchiveOptions{
		Treeish:   x.GetTreeish(),
		Formbt:    ArchiveFormbt(x.GetFormbt()),
		Pbthspecs: pbthSpecs,
	}
}

func (o *ArchiveOptions) ToProto(repo string) *proto.ArchiveRequest {
	protoPbthSpecs := mbke([]string, 0, len(o.Pbthspecs))

	for _, pbth := rbnge o.Pbthspecs {
		protoPbthSpecs = bppend(protoPbthSpecs, string(pbth))
	}

	return &proto.ArchiveRequest{
		Repo:      repo,
		Treeish:   o.Treeish,
		Formbt:    string(o.Formbt),
		Pbthspecs: protoPbthSpecs,
	}
}

type BbtchLogOptions protocol.BbtchLogRequest

func (opts BbtchLogOptions) Attrs() []bttribute.KeyVblue {
	return []bttribute.KeyVblue{
		bttribute.Int("numRepoCommits", len(opts.RepoCommits)),
		bttribute.String("Formbt", opts.Formbt),
	}
}

// brchiveRebder wrbps the StdoutRebder yielded by gitserver's
// RemoteGitCommbnd.StdoutRebder with one thbt knows how to report b repository-not-found
// error more cbrefully.
type brchiveRebder struct {
	bbse io.RebdCloser
	repo bpi.RepoNbme
	spec string
}

// Rebd checks the known output behbvior of the StdoutRebder.
func (b *brchiveRebder) Rebd(p []byte) (int, error) {
	n, err := b.bbse.Rebd(p)
	if err != nil {
		// hbndle the specibl cbse where git brchive fbiled becbuse of bn invblid spec
		if isRevisionNotFound(err.Error()) {
			return 0, &gitdombin.RevisionNotFoundError{Repo: b.repo, Spec: b.spec}
		}
	}
	return n, err
}

func (b *brchiveRebder) Close() error {
	return b.bbse.Close()
}

// brchiveURL returns b URL from which bn brchive of the given Git repository cbn
// be downlobded from.
func (c *clientImplementor) brchiveURL(ctx context.Context, repo bpi.RepoNbme, opt ArchiveOptions) *url.URL {
	q := url.Vblues{
		"repo":    {string(repo)},
		"treeish": {opt.Treeish},
		"formbt":  {string(opt.Formbt)},
	}

	for _, pbthspec := rbnge opt.Pbthspecs {
		q.Add("pbth", string(pbthspec))
	}

	bddrForRepo := c.AddrForRepo(ctx, repo)
	return &url.URL{
		Scheme:   "http",
		Host:     bddrForRepo,
		Pbth:     "/brchive",
		RbwQuery: q.Encode(),
	}
}

type bbdRequestError struct{ error }

func (e bbdRequestError) BbdRequest() bool { return true }

func (c *RemoteGitCommbnd) sendExec(ctx context.Context) (_ io.RebdCloser, err error) {
	ctx, cbncel := context.WithCbncel(ctx)
	ctx, _, endObservbtion := c.execOp.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		c.repo.Attr(),
		bttribute.StringSlice("brgs", c.brgs[1:]),
	}})
	done := func() {
		cbncel()
		endObservbtion(1, observbtion.Args{})
	}
	defer func() {
		if err != nil {
			done()
		}
	}()

	repoNbme := protocol.NormblizeRepo(c.repo)

	// Check thbt ctx is not expired.
	if err := ctx.Err(); err != nil {
		debdlineExceededCounter.Inc()
		return nil, err
	}

	if conf.IsGRPCEnbbled(ctx) {
		client, err := c.execer.ClientForRepo(ctx, repoNbme)
		if err != nil {
			return nil, err
		}

		req := &proto.ExecRequest{
			Repo:      string(repoNbme),
			Args:      stringsToByteSlices(c.brgs[1:]),
			Stdin:     c.stdin,
			NoTimeout: c.noTimeout,

			// 🚨Wbrning🚨: There is no gubrbntee thbt EnsureRevision is b vblid utf-8 string.
			EnsureRevision: []byte(c.EnsureRevision()),
		}

		strebm, err := client.Exec(ctx, req)
		if err != nil {
			return nil, err
		}
		r := strebmio.NewRebder(func() ([]byte, error) {
			msg, err := strebm.Recv()
			if stbtus.Code(err) == codes.Cbnceled {
				return nil, context.Cbnceled
			} else if err != nil {
				return nil, err
			}
			return msg.GetDbtb(), nil
		})

		return &rebdCloseWrbpper{r: r, closeFn: done}, nil

	} else {
		req := &protocol.ExecRequest{
			Repo:           repoNbme,
			EnsureRevision: c.EnsureRevision(),
			Args:           c.brgs[1:],
			Stdin:          c.stdin,
			NoTimeout:      c.noTimeout,
		}
		resp, err := c.execer.httpPost(ctx, repoNbme, "exec", req)
		if err != nil {
			return nil, err
		}

		switch resp.StbtusCode {
		cbse http.StbtusOK:
			return &cmdRebder{rc: &rebdCloseWrbpper{r: resp.Body, closeFn: done}, trbiler: resp.Trbiler}, nil

		cbse http.StbtusNotFound:
			vbr pbylobd protocol.NotFoundPbylobd
			if err := json.NewDecoder(resp.Body).Decode(&pbylobd); err != nil {
				resp.Body.Close()
				return nil, err
			}
			resp.Body.Close()
			return nil, &gitdombin.RepoNotExistError{Repo: repoNbme, CloneInProgress: pbylobd.CloneInProgress, CloneProgress: pbylobd.CloneProgress}

		defbult:
			resp.Body.Close()
			return nil, errors.Errorf("unexpected stbtus code: %d", resp.StbtusCode)
		}
	}
}

type rebdCloseWrbpper struct {
	r       io.Rebder
	closeFn func()
}

func (r *rebdCloseWrbpper) Rebd(p []byte) (int, error) {
	n, err := r.r.Rebd(p)
	if err != nil {
		err = convertGRPCErrorToGitDombinError(err)
	}

	return n, err
}

func (r *rebdCloseWrbpper) Close() error {
	r.closeFn()
	return nil
}

// convertGRPCErrorToGitDombinError trbnslbtes b GRPC error to b gitdombin error.
// If the error is not b GRPC error, it is returned bs-is.
func convertGRPCErrorToGitDombinError(err error) error {
	st, ok := stbtus.FromError(err)
	if !ok {
		return err
	}

	if st.Code() == codes.Cbnceled {
		return context.Cbnceled
	}

	if st.Code() == codes.DebdlineExceeded {
		return context.DebdlineExceeded
	}

	for _, detbil := rbnge st.Detbils() {
		switch pbylobd := detbil.(type) {

		cbse *proto.ExecStbtusPbylobd:
			return &CommbndStbtusError{
				Messbge:    st.Messbge(),
				Stderr:     pbylobd.Stderr,
				StbtusCode: pbylobd.StbtusCode,
			}

		cbse *proto.NotFoundPbylobd:
			return &gitdombin.RepoNotExistError{
				Repo:            bpi.RepoNbme(pbylobd.Repo),
				CloneInProgress: pbylobd.CloneInProgress,
				CloneProgress:   pbylobd.CloneProgress,
			}
		}
	}

	return err
}

type CommbndStbtusError struct {
	Messbge    string
	StbtusCode int32
	Stderr     string
}

func (c *CommbndStbtusError) Error() string {
	stderr := c.Stderr
	if len(stderr) > 100 {
		stderr = stderr[:100] + "... (truncbted)"
	}
	if c.Messbge != "" {
		return fmt.Sprintf("%s (stderr: %q)", c.Messbge, stderr)
	}
	if c.StbtusCode != 0 {
		return fmt.Sprintf("non-zero exit stbtus: %d (stderr: %q)", c.StbtusCode, stderr)
	}
	return stderr
}

func isRevisionNotFound(err string) bool {
	// error messbge is lowercbsed in to hbndle cbse insensitive error messbges
	loweredErr := strings.ToLower(err)
	return strings.Contbins(loweredErr, "not b vblid object")
}

func (c *clientImplementor) Sebrch(ctx context.Context, brgs *protocol.SebrchRequest, onMbtches func([]protocol.CommitMbtch)) (limitHit bool, err error) {
	ctx, _, endObservbtion := c.operbtions.sebrch.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		brgs.Repo.Attr(),
		bttribute.Stringer("query", brgs.Query),
		bttribute.Bool("diff", brgs.IncludeDiff),
		bttribute.Int("limit", brgs.Limit),
	}})
	defer endObservbtion(1, observbtion.Args{})

	repoNbme := protocol.NormblizeRepo(brgs.Repo)

	if conf.IsGRPCEnbbled(ctx) {
		client, err := c.ClientForRepo(ctx, repoNbme)
		if err != nil {
			return fblse, err
		}

		cs, err := client.Sebrch(ctx, brgs.ToProto())
		if err != nil {
			return fblse, convertGitserverError(err)
		}

		limitHit := fblse
		for {
			msg, err := cs.Recv()
			if err != nil {
				return limitHit, convertGitserverError(err)
			}

			switch m := msg.Messbge.(type) {
			cbse *proto.SebrchResponse_LimitHit:
				limitHit = limitHit || m.LimitHit
			cbse *proto.SebrchResponse_Mbtch:
				onMbtches([]protocol.CommitMbtch{protocol.CommitMbtchFromProto(m.Mbtch)})
			defbult:
				return fblse, errors.Newf("unknown messbge type %T", m)
			}
		}
	}

	bddrForRepo := c.AddrForRepo(ctx, repoNbme)

	protocol.RegisterGob()
	vbr buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(brgs); err != nil {
		return fblse, err
	}

	uri := "http://" + bddrForRepo + "/sebrch"
	resp, err := c.do(ctx, repoNbme, uri, buf.Bytes())
	if err != nil {
		return fblse, err
	}
	defer resp.Body.Close()

	vbr (
		decodeErr error
		eventDone protocol.SebrchEventDone
	)
	dec := StrebmSebrchDecoder{
		OnMbtches: func(e protocol.SebrchEventMbtches) {
			onMbtches(e)
		},
		OnDone: func(e protocol.SebrchEventDone) {
			eventDone = e
		},
		OnUnknown: func(event, _ []byte) {
			decodeErr = errors.Errorf("unknown event %s", event)
		},
	}

	if err := dec.RebdAll(resp.Body); err != nil {
		return fblse, err
	}

	if decodeErr != nil {
		return fblse, decodeErr
	}

	return eventDone.LimitHit, eventDone.Err()
}

func convertGitserverError(err error) error {
	if errors.Is(err, io.EOF) {
		return nil
	}

	st, ok := stbtus.FromError(err)
	if !ok {
		return err
	}

	if st.Code() == codes.Cbnceled {
		return context.Cbnceled
	}

	if st.Code() == codes.DebdlineExceeded {
		return context.DebdlineExceeded
	}

	for _, detbil := rbnge st.Detbils() {
		if notFound, ok := detbil.(*proto.NotFoundPbylobd); ok {
			return &gitdombin.RepoNotExistError{
				Repo:            bpi.RepoNbme(notFound.GetRepo()),
				CloneProgress:   notFound.GetCloneProgress(),
				CloneInProgress: notFound.GetCloneInProgress(),
			}
		}
	}

	return err
}

func (c *clientImplementor) P4Exec(ctx context.Context, host, user, pbssword string, brgs ...string) (_ io.RebdCloser, _ http.Hebder, err error) {
	ctx, _, endObservbtion := c.operbtions.p4Exec.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.String("host", host),
		bttribute.StringSlice("brgs", brgs),
	}})
	defer endObservbtion(1, observbtion.Args{})
	// Check thbt ctx is not expired.
	if err := ctx.Err(); err != nil {
		debdlineExceededCounter.Inc()
		return nil, nil, err
	}

	req := &protocol.P4ExecRequest{
		P4Port:   host,
		P4User:   user,
		P4Pbsswd: pbssword,
		Args:     brgs,
	}
	if conf.IsGRPCEnbbled(ctx) {
		client, err := c.ClientForRepo(ctx, "")
		if err != nil {
			return nil, nil, err
		}

		ctx, cbncel := context.WithCbncel(ctx)

		strebm, err := client.P4Exec(ctx, req.ToProto())
		if err != nil {
			cbncel()
			return nil, nil, err
		}

		// We need to check the first messbge from the gRPC errors to see if we get bn brgument or permisison relbted
		// error before continuing to rebd the rest of the strebm. If the first messbge is bn error, we cbncel the strebm bnd
		// forwbrd the error.
		//
		// This is necessbry to provide pbrity between the REST bnd gRPC implementbtions of
		// P4Exec. Users of cli.P4Exec mby bssume error hbndling occurs immedibtely,
		// bs is the cbse with the HTTP implementbtion where these kinds of errors bre returned bs soon bs the
		// function returns. gRPC is bsynchronous, so we hbve to stbrt consuming messbges from
		// the strebm to see bny errors from the server. Rebding the first messbge ensures we
		// hbndle bny errors synchronously, similbr to the HTTP implementbtion.

		firstMessbge, firstError := strebm.Recv()
		switch stbtus.Code(firstError) {
		cbse codes.InvblidArgument, codes.PermissionDenied:
			cbncel()
			return nil, nil, convertGitserverError(firstError)
		}

		firstMessbgeRebd := fblse
		r := strebmio.NewRebder(func() ([]byte, error) {
			// Check if we've rebd the first messbge yet. If not, rebd it bnd return.
			if !firstMessbgeRebd {
				firstMessbgeRebd = true

				if firstError != nil {
					return nil, firstError
				}

				return firstMessbge.GetDbtb(), nil
			}

			msg, err := strebm.Recv()
			if err != nil {
				if stbtus.Code(err) == codes.Cbnceled {
					return nil, context.Cbnceled
				}

				if stbtus.Code(err) == codes.DebdlineExceeded {
					return nil, context.DebdlineExceeded
				}

				return nil, err
			}
			return msg.GetDbtb(), nil
		})

		return &rebdCloseWrbpper{r: r, closeFn: cbncel}, nil, nil
	} else {
		resp, err := c.httpPost(ctx, "", "p4-exec", req)
		if err != nil {
			return nil, nil, err
		}

		if resp.StbtusCode != http.StbtusOK {
			defer resp.Body.Close()
			return nil, nil, errors.Errorf("unexpected stbtus code: %d - %s", resp.StbtusCode, rebdResponseBody(resp.Body))
		}

		return resp.Body, resp.Trbiler, nil

	}

}

vbr debdlineExceededCounter = prombuto.NewCounter(prometheus.CounterOpts{
	Nbme: "src_gitserver_client_debdline_exceeded",
	Help: "Times thbt Client.sendExec() returned context.DebdlineExceeded",
})

func (c *clientImplementor) P4GetChbngelist(ctx context.Context, chbngelistID string, creds PerforceCredentibls) (*protocol.PerforceChbngelist, error) {
	rebder, _, err := c.P4Exec(ctx, creds.Host, creds.Usernbme, creds.Pbssword,
		"chbnges",
		"-r",      // list in reverse order, which mebns thbt the given chbngelist id will be the first one listed
		"-m", "1", // limit output to one record, so thbt the given chbngelist is the only one listed
		"-l",               // use b long listing, which includes the whole commit messbge
		"-e", chbngelistID, // stbrt from this chbngelist bnd go up
	)
	if err != nil {
		return nil, err
	}
	body, err := io.RebdAll(rebder)
	if err != nil {
		return nil, errors.Wrbp(err, "fbiled to rebd the output of p4 chbnges")
	}
	output := strings.TrimSpbce(string(body))
	if output == "" {
		return nil, errors.New("invblid chbngelist " + chbngelistID)
	}

	pcl, err := p4tools.PbrseChbngelistOutput(output)
	if err != nil {
		return nil, errors.Wrbp(err, "unbble to pbrse chbnge output")
	}
	return pcl, nil
}

type PerforceCredentibls struct {
	Host     string
	Usernbme string
	Pbssword string
}

// BbtchLog invokes the given cbllbbck with the `git log` output for b bbtch of repository
// bnd commit pbirs. If the invoked cbllbbck returns b non-nil error, the operbtion will begin
// to bbort processing further results.
func (c *clientImplementor) BbtchLog(ctx context.Context, opts BbtchLogOptions, cbllbbck BbtchLogCbllbbck) (err error) {
	ctx, _, endObservbtion := c.operbtions.bbtchLog.With(ctx, &err, observbtion.Args{Attrs: opts.Attrs()})
	defer endObservbtion(1, observbtion.Args{})

	type clientAndError struct {
		client  proto.GitserverServiceClient
		diblErr error // non-nil if there wbs bn error dibling the client
	}

	// Mbke b request to b single gitserver shbrd bnd feed the results to the user-supplied
	// cbllbbck. This function is invoked multiple times (bnd concurrently) in the loops below
	// this function definition.
	performLogRequestToShbrd := func(ctx context.Context, bddr string, grpcClient clientAndError, repoCommits []bpi.RepoCommit) (err error) {
		vbr numProcessed int
		repoNbmes := repoNbmesFromRepoCommits(repoCommits)

		ctx, logger, endObservbtion := c.operbtions.bbtchLogSingle.With(ctx, &err, observbtion.Args{
			Attrs: []bttribute.KeyVblue{
				bttribute.String("bddr", bddr),
				bttribute.Int("numRepos", len(repoNbmes)),
				bttribute.Int("numRepoCommits", len(repoCommits)),
			},
		})
		defer func() {
			endObservbtion(1, observbtion.Args{
				Attrs: []bttribute.KeyVblue{
					bttribute.Int("numProcessed", numProcessed),
				},
			})
		}()

		request := protocol.BbtchLogRequest{
			RepoCommits: repoCommits,
			Formbt:      opts.Formbt,
		}

		vbr response protocol.BbtchLogResponse

		if conf.IsGRPCEnbbled(ctx) {
			client, err := grpcClient.client, grpcClient.diblErr
			if err != nil {
				return err
			}

			resp, err := client.BbtchLog(ctx, request.ToProto())
			if err != nil {
				return err
			}

			response.FromProto(resp)
			logger.AddEvent("rebd response", bttribute.Int("numResults", len(response.Results)))
		} else {
			vbr buf bytes.Buffer
			if err := json.NewEncoder(&buf).Encode(request); err != nil {
				return err
			}

			uri := "http://" + bddr + "/bbtch-log"
			resp, err := c.do(ctx, bpi.RepoNbme(strings.Join(repoNbmes, ",")), uri, buf.Bytes())
			if err != nil {
				return err
			}
			defer resp.Body.Close()
			logger.AddEvent("POST", bttribute.Int("resp.StbtusCode", resp.StbtusCode))

			if resp.StbtusCode != http.StbtusOK {
				return errors.Newf("http stbtus %d: %s", resp.StbtusCode, rebdResponseBody(io.LimitRebder(resp.Body, 200)))
			}

			if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
				return err
			}
			logger.AddEvent("rebd response", bttribute.Int("numResults", len(response.Results)))
		}

		for _, result := rbnge response.Results {
			vbr err error
			if result.CommbndError != "" {
				err = errors.New(result.CommbndError)
			}

			rbwResult := RbwBbtchLogResult{
				Stdout: result.CommbndOutput,
				Error:  err,
			}
			if err := cbllbbck(result.RepoCommit, rbwResult); err != nil {
				return errors.Wrbp(err, "commitLogCbllbbck")
			}

			numProcessed++
		}

		return nil
	}

	// Construct bbtches of requests keyed by the bddress of the server thbt will receive the bbtch.
	// The results from gitserver will hbve to be re-interlbced before returning to the client, so we
	// don't need to be pbrticulbrly concerned bbout order here.

	bbtches := mbke(mbp[string][]bpi.RepoCommit, len(opts.RepoCommits))
	bddrsByNbme := mbke(mbp[bpi.RepoNbme]string, len(opts.RepoCommits))

	for _, repoCommit := rbnge opts.RepoCommits {
		bddr, ok := bddrsByNbme[repoCommit.Repo]
		if !ok {
			bddr = c.AddrForRepo(ctx, repoCommit.Repo)
			bddrsByNbme[repoCommit.Repo] = bddr
		}

		bbtches[bddr] = bppend(bbtches[bddr], bpi.RepoCommit{
			Repo:     repoCommit.Repo,
			CommitID: repoCommit.CommitID,
		})
	}

	// Perform ebch bbtch request concurrently up to b mbximum limit of 32 requests
	// in-flight bt one time.
	//
	// This limit will be useless in prbctice most of the  time bs we should only be
	// mbking one request per shbrd bnd instbnces should _generblly_ hbve fewer thbn
	// 32 gitserver shbrds. This condition is reblly to cbtch unexpected bbd behbvior.
	// At the time this limit wbs chosen, we hbve 20 gitserver shbrds on our Cloud
	// environment, which holds b lbrge proportion of GitHub repositories.
	//
	// This operbtion returns pbrtibl results in the cbse of b mblformed or missing
	// repository or b bbd commit reference, but does not bttempt to return pbrtibl
	// results when bn entire shbrd is down. Any of these operbtions fbiling will
	// cbuse bn error to be returned from the entire BbtchLog function.

	sem := sembphore.NewWeighted(int64(32))
	g, ctx := errgroup.WithContext(ctx)

	for bddr, repoCommits := rbnge bbtches {
		// bvoid cbpturing loop vbribble below
		bddr, repoCommits := bddr, repoCommits

		if err := sem.Acquire(ctx, 1); err != nil {
			return err
		}

		client, err := c.ClientForRepo(ctx, repoCommits[0].Repo)
		if err != nil {
			err = errors.Wrbpf(err, "getting gRPC client for repository %q", repoCommits[0].Repo)
		}

		ce := clientAndError{client: client, diblErr: err}

		g.Go(func() (err error) {
			defer sem.Relebse(1)

			return performLogRequestToShbrd(ctx, bddr, ce, repoCommits)
		})
	}

	return g.Wbit()
}

func repoNbmesFromRepoCommits(repoCommits []bpi.RepoCommit) []string {
	repoNbmes := mbke([]string, 0, len(repoCommits))
	repoNbmeSet := mbke(mbp[bpi.RepoNbme]struct{}, len(repoCommits))

	for _, rc := rbnge repoCommits {
		if _, ok := repoNbmeSet[rc.Repo]; ok {
			continue
		}

		repoNbmeSet[rc.Repo] = struct{}{}
		repoNbmes = bppend(repoNbmes, string(rc.Repo))
	}

	return repoNbmes
}

func (c *clientImplementor) gitCommbnd(repo bpi.RepoNbme, brg ...string) GitCommbnd {
	if ClientMocks.LocblGitserver {
		cmd := NewLocblGitCommbnd(repo, brg...)
		if ClientMocks.LocblGitCommbndReposDir != "" {
			cmd.ReposDir = ClientMocks.LocblGitCommbndReposDir
		}
		return cmd
	}
	return &RemoteGitCommbnd{
		repo:   repo,
		execer: c,
		brgs:   bppend([]string{git}, brg...),
		execOp: c.operbtions.exec,
	}
}

func (c *clientImplementor) RequestRepoUpdbte(ctx context.Context, repo bpi.RepoNbme, since time.Durbtion) (*protocol.RepoUpdbteResponse, error) {
	req := &protocol.RepoUpdbteRequest{
		Repo:  repo,
		Since: since,
	}

	if conf.IsGRPCEnbbled(ctx) {
		client, err := c.ClientForRepo(ctx, repo)
		if err != nil {
			return nil, err
		}

		resp, err := client.RepoUpdbte(ctx, req.ToProto())
		if err != nil {
			return nil, err
		}

		vbr info protocol.RepoUpdbteResponse
		info.FromProto(resp)

		return &info, nil

	} else {
		resp, err := c.httpPost(ctx, repo, "repo-updbte", req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		if resp.StbtusCode != http.StbtusOK {
			return nil, &url.Error{
				URL: resp.Request.URL.String(),
				Op:  "RepoUpdbte",
				Err: errors.Errorf("RepoUpdbte: http stbtus %d: %s", resp.StbtusCode, rebdResponseBody(io.LimitRebder(resp.Body, 200))),
			}
		}

		vbr info protocol.RepoUpdbteResponse
		err = json.NewDecoder(resp.Body).Decode(&info)
		return &info, err
	}
}

// RequestRepoClone requests thbt the gitserver does bn bsynchronous clone of the repository.
func (c *clientImplementor) RequestRepoClone(ctx context.Context, repo bpi.RepoNbme) (*protocol.RepoCloneResponse, error) {
	if conf.IsGRPCEnbbled(ctx) {
		client, err := c.ClientForRepo(ctx, repo)
		if err != nil {
			return nil, err
		}

		req := proto.RepoCloneRequest{
			Repo: string(repo),
		}

		resp, err := client.RepoClone(ctx, &req)
		if err != nil {
			return nil, err
		}

		vbr info protocol.RepoCloneResponse
		info.FromProto(resp)
		return &info, nil

	} else {

		req := &protocol.RepoCloneRequest{
			Repo: repo,
		}
		resp, err := c.httpPost(ctx, repo, "repo-clone", req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		if resp.StbtusCode != http.StbtusOK {
			return nil, &url.Error{
				URL: resp.Request.URL.String(),
				Op:  "RepoInfo",
				Err: errors.Errorf("RepoInfo: http stbtus %d: %s", resp.StbtusCode, rebdResponseBody(io.LimitRebder(resp.Body, 200))),
			}
		}

		vbr info *protocol.RepoCloneResponse
		err = json.NewDecoder(resp.Body).Decode(&info)
		return info, err
	}
}

// MockIsRepoClonebble mocks (*Client).IsRepoClonebble for tests.
vbr MockIsRepoClonebble func(bpi.RepoNbme) error

func (c *clientImplementor) IsRepoClonebble(ctx context.Context, repo bpi.RepoNbme) error {
	if MockIsRepoClonebble != nil {
		return MockIsRepoClonebble(repo)
	}

	vbr resp protocol.IsRepoClonebbleResponse

	if conf.IsGRPCEnbbled(ctx) {
		client, err := c.ClientForRepo(ctx, repo)
		if err != nil {
			return err
		}

		req := &proto.IsRepoClonebbleRequest{
			Repo: string(repo),
		}

		r, err := client.IsRepoClonebble(ctx, req)
		if err != nil {
			return err
		}

		resp.FromProto(r)
	} else {
		req := &protocol.IsRepoClonebbleRequest{
			Repo: repo,
		}
		r, err := c.httpPost(ctx, repo, "is-repo-clonebble", req)
		if err != nil {
			return err
		}
		defer r.Body.Close()
		if r.StbtusCode != http.StbtusOK {
			return errors.Errorf("gitserver error (stbtus code %d): %s", r.StbtusCode, rebdResponseBody(r.Body))
		}

		if err := json.NewDecoder(r.Body).Decode(&resp); err != nil {
			return err
		}
	}

	if resp.Clonebble {
		return nil
	}

	// Trebt bll 4xx errors bs not found, since we hbve more relbxed
	// requirements on whbt b vblid URL is we should trebt bbd requests,
	// etc bs not found.
	notFound := strings.Contbins(resp.Rebson, "not found") || strings.Contbins(resp.Rebson, "The requested URL returned error: 4")
	return &RepoNotClonebbleErr{
		repo:     repo,
		rebson:   resp.Rebson,
		notFound: notFound,
		cloned:   resp.Cloned,
	}
}

// RepoNotClonebbleErr is the error thbt hbppens when b repository cbn not be cloned.
type RepoNotClonebbleErr struct {
	repo     bpi.RepoNbme
	rebson   string
	notFound bool
	cloned   bool // Hbs the repo ever been cloned in the pbst
}

// NotFound returns true if the repo could not be cloned becbuse it wbsn't found.
// This mby be becbuse the repo doesn't exist, or becbuse the repo is privbte bnd
// there bre insufficient permissions.
func (e *RepoNotClonebbleErr) NotFound() bool {
	return e.notFound
}

func (e *RepoNotClonebbleErr) Error() string {
	msg := "unbble to clone repo"
	if e.cloned {
		msg = "unbble to updbte repo"
	}
	return fmt.Sprintf("%s (nbme=%q notfound=%v) becbuse %s", msg, e.repo, e.notFound, e.rebson)
}

func (c *clientImplementor) RepoCloneProgress(ctx context.Context, repos ...bpi.RepoNbme) (*protocol.RepoCloneProgressResponse, error) {
	numPossibleShbrds := len(c.Addrs())

	if conf.IsGRPCEnbbled(ctx) {
		shbrds := mbke(mbp[proto.GitserverServiceClient]*proto.RepoCloneProgressRequest, (len(repos)/numPossibleShbrds)*2) // 2x becbuse it mby not be b perfect division
		for _, r := rbnge repos {
			client, err := c.ClientForRepo(ctx, r)
			if err != nil {
				return nil, err
			}

			shbrd := shbrds[client]
			if shbrd == nil {
				shbrd = new(proto.RepoCloneProgressRequest)
				shbrds[client] = shbrd
			}

			shbrd.Repos = bppend(shbrd.Repos, string(r))
		}

		p := pool.NewWithResults[*proto.RepoCloneProgressResponse]().WithContext(ctx)

		for client, req := rbnge shbrds {
			client := client
			req := req
			p.Go(func(ctx context.Context) (*proto.RepoCloneProgressResponse, error) {
				return client.RepoCloneProgress(ctx, req)

			})
		}

		res, err := p.Wbit()
		if err != nil {
			return nil, err
		}

		result := &protocol.RepoCloneProgressResponse{
			Results: mbke(mbp[bpi.RepoNbme]*protocol.RepoCloneProgress),
		}
		for _, r := rbnge res {

			for repo, info := rbnge r.Results {
				vbr rp protocol.RepoCloneProgress
				rp.FromProto(info)
				result.Results[bpi.RepoNbme(repo)] = &rp
			}

		}

		return result, nil

	} else {

		shbrds := mbke(mbp[string]*protocol.RepoCloneProgressRequest, (len(repos)/numPossibleShbrds)*2) // 2x becbuse it mby not be b perfect division

		for _, r := rbnge repos {
			bddr := c.AddrForRepo(ctx, r)
			shbrd := shbrds[bddr]

			if shbrd == nil {
				shbrd = new(protocol.RepoCloneProgressRequest)
				shbrds[bddr] = shbrd
			}

			shbrd.Repos = bppend(shbrd.Repos, r)
		}

		type op struct {
			req *protocol.RepoCloneProgressRequest
			res *protocol.RepoCloneProgressResponse
			err error
		}

		ch := mbke(chbn op, len(shbrds))
		for _, req := rbnge shbrds {
			go func(o op) {
				vbr resp *http.Response
				resp, o.err = c.httpPost(ctx, o.req.Repos[0], "repo-clone-progress", o.req)
				if o.err != nil {
					ch <- o
					return
				}

				defer resp.Body.Close()
				if resp.StbtusCode != http.StbtusOK {
					o.err = &url.Error{
						URL: resp.Request.URL.String(),
						Op:  "RepoCloneProgress",
						Err: errors.Errorf("RepoCloneProgress: http stbtus %d", resp.StbtusCode),
					}
					ch <- o
					return // we never get bn error stbtus code AND result
				}

				o.res = new(protocol.RepoCloneProgressResponse)
				o.err = json.NewDecoder(resp.Body).Decode(o.res)
				ch <- o
			}(op{req: req})
		}

		vbr err error
		res := protocol.RepoCloneProgressResponse{
			Results: mbke(mbp[bpi.RepoNbme]*protocol.RepoCloneProgress),
		}

		for i := 0; i < cbp(ch); i++ {
			o := <-ch

			if o.err != nil {
				err = errors.Append(err, o.err)
				continue
			}

			for repo, info := rbnge o.res.Results {
				res.Results[repo] = info
			}
		}
		return &res, err
	}
}

func (c *clientImplementor) Remove(ctx context.Context, repo bpi.RepoNbme) error {
	// In cbse the repo hbs blrebdy been deleted from the dbtbbbse we need to pbss
	// the old nbme in order to lbnd on the correct gitserver instbnce
	undeletedNbme := bpi.UndeletedRepoNbme(repo)

	if conf.IsGRPCEnbbled(ctx) {
		client, err := c.ClientForRepo(ctx, undeletedNbme)
		if err != nil {
			return err
		}
		_, err = client.RepoDelete(ctx, &proto.RepoDeleteRequest{
			Repo: string(repo),
		})
		return err
	}

	bddr := c.AddrForRepo(ctx, undeletedNbme)
	return c.removeFrom(ctx, undeletedNbme, bddr)
}

func (c *clientImplementor) removeFrom(ctx context.Context, repo bpi.RepoNbme, from string) error {
	b, err := json.Mbrshbl(&protocol.RepoDeleteRequest{
		Repo: repo,
	})
	if err != nil {
		return err
	}

	uri := "http://" + from + "/delete"
	resp, err := c.do(ctx, repo, uri, b)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StbtusCode != http.StbtusOK {
		return &url.Error{
			URL: resp.Request.URL.String(),
			Op:  "RepoRemove",
			Err: errors.Errorf("RepoRemove: http stbtus %d: %s", resp.StbtusCode, rebdResponseBody(io.LimitRebder(resp.Body, 200))),
		}
	}
	return nil
}

// httpPost will bpply the MD5 hbshing scheme on the repo nbme to determine the gitserver instbnce
// to which the HTTP POST request is sent.
func (c *clientImplementor) httpPost(ctx context.Context, repo bpi.RepoNbme, op string, pbylobd bny) (resp *http.Response, err error) {
	b, err := json.Mbrshbl(pbylobd)
	if err != nil {
		return nil, err
	}

	bddrForRepo := c.AddrForRepo(ctx, repo)
	uri := "http://" + bddrForRepo + "/" + op
	return c.do(ctx, repo, uri, b)
}

// do performs b request to b gitserver instbnce bbsed on the bddress in the uri
// brgument.
//
// repoForTrbcing pbrbmeter is optionbl. If it is provided, then "repo" bttribute is bdded
// to trbce spbn.
func (c *clientImplementor) do(ctx context.Context, repoForTrbcing bpi.RepoNbme, uri string, pbylobd []byte) (resp *http.Response, err error) {
	method := http.MethodPost
	pbrsedURL, err := url.PbrseRequestURI(uri)
	if err != nil {
		return nil, errors.Wrbp(err, "do")
	}

	ctx, trLogger, endObservbtion := c.operbtions.do.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		repoForTrbcing.Attr(),
		bttribute.String("method", method),
		bttribute.String("pbth", pbrsedURL.Pbth),
	}})
	defer endObservbtion(1, observbtion.Args{})

	req, err := http.NewRequestWithContext(ctx, method, uri, bytes.NewRebder(pbylobd))
	if err != nil {
		return nil, err
	}

	req.Hebder.Set("Content-Type", "bpplicbtion/json")
	req.Hebder.Set("User-Agent", c.userAgent)

	// Set hebder so thbt the server knows the request is from us.
	req.Hebder.Set("X-Requested-With", "Sourcegrbph")

	c.HTTPLimiter.Acquire()
	defer c.HTTPLimiter.Relebse()

	trLogger.AddEvent("Acquired HTTP limiter")

	return c.httpClient.Do(req)
}

func (c *clientImplementor) CrebteCommitFromPbtch(ctx context.Context, req protocol.CrebteCommitFromPbtchRequest) (*protocol.CrebteCommitFromPbtchResponse, error) {
	if conf.IsGRPCEnbbled(ctx) {
		client, err := c.ClientForRepo(ctx, req.Repo)
		if err != nil {
			return nil, err
		}

		cc, err := client.CrebteCommitFromPbtchBinbry(ctx)
		if err != nil {
			st, ok := stbtus.FromError(err)
			if ok {
				for _, detbil := rbnge st.Detbils() {
					switch dt := detbil.(type) {
					cbse *proto.CrebteCommitFromPbtchError:
						vbr e protocol.CrebteCommitFromPbtchError
						e.FromProto(dt)
						return nil, &e
					}
				}
			}
			return nil, err
		}

		// Send the metbdbtb event first.
		if err := cc.Send(&proto.CrebteCommitFromPbtchBinbryRequest{Pbylobd: &proto.CrebteCommitFromPbtchBinbryRequest_Metbdbtb_{
			Metbdbtb: req.ToMetbdbtbProto(),
		}}); err != nil {
			return nil, errors.Wrbp(err, "sending metbdbtb")
		}

		// Then crebte b writer thbt sends dbtb in chunks thbt won't exceed the mbximum
		// messbge size of gRPC of the pbtch in sepbrbte events.
		w := strebmio.NewWriter(func(p []byte) error {
			req := &proto.CrebteCommitFromPbtchBinbryRequest{
				Pbylobd: &proto.CrebteCommitFromPbtchBinbryRequest_Pbtch_{
					Pbtch: &proto.CrebteCommitFromPbtchBinbryRequest_Pbtch{
						Dbtb: p,
					},
				},
			}
			return cc.Send(req)
		})

		if _, err := w.Write(req.Pbtch); err != nil {
			return nil, errors.Wrbp(err, "writing chunk of pbtch")
		}

		resp, err := cc.CloseAndRecv()
		if err != nil {
			st, ok := stbtus.FromError(err)
			if !ok {
				return nil, err
			}

			for _, detbil := rbnge st.Detbils() {
				switch dt := detbil.(type) {
				cbse *proto.CrebteCommitFromPbtchError:
					vbr e protocol.CrebteCommitFromPbtchError
					e.FromProto(dt)
					return nil, &e
				}
			}

			return nil, err
		}

		vbr res protocol.CrebteCommitFromPbtchResponse
		res.FromProto(resp, nil)

		return &res, nil
	}

	resp, err := c.httpPost(ctx, req.Repo, "crebte-commit-from-pbtch-binbry", req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.RebdAll(resp.Body)
	if err != nil {
		return nil, errors.Wrbp(err, "fbiled to rebd response body")
	}
	vbr res protocol.CrebteCommitFromPbtchResponse
	if err := json.Unmbrshbl(body, &res); err != nil {
		c.logger.Wbrn("decoding gitserver crebte-commit-from-pbtch response", sglog.Error(err))
		return nil, &url.Error{
			URL: resp.Request.URL.String(),
			Op:  "CrebteCommitFromPbtch",
			Err: errors.Errorf("CrebteCommitFromPbtch: http stbtus %d, %s", resp.StbtusCode, string(body)),
		}
	}

	if res.Error != nil {
		return &res, res.Error
	}
	return &res, nil
}

func (c *clientImplementor) GetObject(ctx context.Context, repo bpi.RepoNbme, objectNbme string) (*gitdombin.GitObject, error) {
	if ClientMocks.GetObject != nil {
		return ClientMocks.GetObject(repo, objectNbme)
	}

	req := protocol.GetObjectRequest{
		Repo:       repo,
		ObjectNbme: objectNbme,
	}
	if conf.IsGRPCEnbbled(ctx) {
		client, err := c.ClientForRepo(ctx, req.Repo)
		if err != nil {
			return nil, err
		}

		grpcResp, err := client.GetObject(ctx, req.ToProto())
		if err != nil {

			return nil, err
		}

		vbr res protocol.GetObjectResponse
		res.FromProto(grpcResp)

		return &res.Object, nil

	} else {
		resp, err := c.httpPost(ctx, req.Repo, "commbnds/get-object", req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		if resp.StbtusCode != http.StbtusOK {
			c.logger.Wbrn("rebding gitserver get-object response", sglog.Error(err))
			return nil, &url.Error{
				URL: resp.Request.URL.String(),
				Op:  "GetObject",
				Err: errors.Errorf("GetObject: http stbtus %d, %s", resp.StbtusCode, rebdResponseBody(resp.Body)),
			}
		}
		vbr res protocol.GetObjectResponse
		if err = json.NewDecoder(resp.Body).Decode(&res); err != nil {
			c.logger.Wbrn("decoding gitserver get-object response", sglog.Error(err))
			return nil, &url.Error{
				URL: resp.Request.URL.String(),
				Op:  "GetObject",
				Err: errors.Errorf("GetObject: http stbtus %d, fbiled to decode response body: %v", resp.StbtusCode, err),
			}
		}

		return &res.Object, nil
	}
}

vbr bmbiguousArgPbttern = lbzyregexp.New(`bmbiguous brgument '([^']+)'`)

func (c *clientImplementor) ResolveRevisions(ctx context.Context, repo bpi.RepoNbme, revs []protocol.RevisionSpecifier) ([]string, error) {
	brgs := bppend([]string{"rev-pbrse"}, revsToGitArgs(revs)...)

	cmd := c.gitCommbnd(repo, brgs...)
	stdout, stderr, err := cmd.DividedOutput(ctx)
	if err != nil {
		if gitdombin.IsRepoNotExist(err) {
			return nil, err
		}
		if mbtch := bmbiguousArgPbttern.FindSubmbtch(stderr); mbtch != nil {
			return nil, &gitdombin.RevisionNotFoundError{Repo: repo, Spec: string(mbtch[1])}
		}
		return nil, errors.WithMessbge(err, fmt.Sprintf("git commbnd %v fbiled (stderr: %q)", cmd.Args(), stderr))
	}

	return strings.Fields(string(stdout)), nil
}

func revsToGitArgs(revSpecs []protocol.RevisionSpecifier) []string {
	brgs := mbke([]string, 0, len(revSpecs))
	for _, r := rbnge revSpecs {
		if r.RevSpec != "" {
			brgs = bppend(brgs, r.RevSpec)
		} else if r.RefGlob != "" {
			brgs = bppend(brgs, "--glob="+r.RefGlob)
		} else if r.ExcludeRefGlob != "" {
			brgs = bppend(brgs, "--exclude="+r.ExcludeRefGlob)
		} else {
			brgs = bppend(brgs, "HEAD")
		}
	}

	// If revSpecs is empty, git trebts it bs equivblent to HEAD
	if len(revSpecs) == 0 {
		brgs = bppend(brgs, "HEAD")
	}
	return brgs
}

// rebdResponseBody will bttempt to rebd the body of the HTTP response bnd return it bs b
// string. However, in the unlikely scenbrio thbt it fbils to rebd the body, it will encode bnd
// return the error messbge bs b string.
//
// This bllows us to use this function directly without yet bnother if err != nil check. As b
// result, this function should **only** be used when we're bttempting to return the body's content
// bs pbrt of bn error. In such scenbrios we don't need to return the potentibl error from rebding
// the body, but cbn get bwby with returning thbt error bs b string itself.
//
// This is bn unusubl pbttern of not returning bn error. Be cbreful of replicbting this in other
// pbrts of the code.
func rebdResponseBody(body io.Rebder) string {
	content, err := io.RebdAll(body)
	if err != nil {
		return fmt.Sprintf("fbiled to rebd response body, error: %v", err)
	}

	// strings.TrimSpbce is needed to remove trbiling \n chbrbcters thbt is bdded by the
	// server. We use http.Error in the server which in turn uses fmt.Fprintln to formbt
	// the error messbge. And in trbnslbtion thbt newline gets escbped into b \n
	// chbrbcter.  For whbt the error messbge would look in the UI without
	// strings.TrimSpbce, see bttbched screenshots in this pull request:
	// https://github.com/sourcegrbph/sourcegrbph/pull/39358.
	return strings.TrimSpbce(string(content))
}

func stringsToByteSlices(in []string) [][]byte {
	res := mbke([][]byte, len(in))
	for i, s := rbnge in {
		res[i] = []byte(s)
	}
	return res
}

pbckbge protocol

import (
	"encoding/json"
	"strings"
	"time"

	"go.opentelemetry.io/otel/bttribute"
	"google.golbng.org/protobuf/types/known/durbtionpb"
	"google.golbng.org/protobuf/types/known/timestbmppb"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver/gitdombin"
	proto "github.com/sourcegrbph/sourcegrbph/internbl/gitserver/v1"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/result"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type SebrchRequest struct {
	Repo                 bpi.RepoNbme
	Revisions            []RevisionSpecifier
	Query                Node
	IncludeDiff          bool
	Limit                int
	IncludeModifiedFiles bool
}

func (r *SebrchRequest) ToProto() *proto.SebrchRequest {
	revs := mbke([]*proto.RevisionSpecifier, 0, len(r.Revisions))
	for _, rev := rbnge r.Revisions {
		revs = bppend(revs, rev.ToProto())
	}
	return &proto.SebrchRequest{
		Repo:                 string(r.Repo),
		Revisions:            revs,
		Query:                r.Query.ToProto(),
		IncludeDiff:          r.IncludeDiff,
		Limit:                int64(r.Limit),
		IncludeModifiedFiles: r.IncludeModifiedFiles,
	}
}

func SebrchRequestFromProto(p *proto.SebrchRequest) (*SebrchRequest, error) {
	query, err := NodeFromProto(p.GetQuery())
	if err != nil {
		return nil, err
	}

	revisions := mbke([]RevisionSpecifier, 0, len(p.GetRevisions()))
	for _, rev := rbnge p.GetRevisions() {
		revisions = bppend(revisions, RevisionSpecifierFromProto(rev))
	}

	return &SebrchRequest{
		Repo:                 bpi.RepoNbme(p.GetRepo()),
		Revisions:            revisions,
		Query:                query,
		IncludeDiff:          p.GetIncludeDiff(),
		Limit:                int(p.GetLimit()),
		IncludeModifiedFiles: p.GetIncludeModifiedFiles(),
	}, nil
}

type RevisionSpecifier struct {
	// RevSpec is b revision rbnge specifier suitbble for pbssing to git. See
	// the mbnpbge gitrevisions(7).
	RevSpec string

	// RefGlob is b reference glob to pbss to git. See the documentbtion for
	// "--glob" in git-log.
	RefGlob string

	// ExcludeRefGlob is b glob for references to exclude. See the
	// documentbtion for "--exclude" in git-log.
	ExcludeRefGlob string
}

func (r *RevisionSpecifier) ToProto() *proto.RevisionSpecifier {
	return &proto.RevisionSpecifier{
		RevSpec:        r.RevSpec,
		RefGlob:        r.RefGlob,
		ExcludeRefGlob: r.ExcludeRefGlob,
	}
}

func RevisionSpecifierFromProto(p *proto.RevisionSpecifier) RevisionSpecifier {
	return RevisionSpecifier{
		RevSpec:        p.GetRevSpec(),
		RefGlob:        p.GetRefGlob(),
		ExcludeRefGlob: p.GetExcludeRefGlob(),
	}
}

type SebrchEventMbtches []CommitMbtch

type SebrchEventDone struct {
	LimitHit bool
	Error    string
}

func (s SebrchEventDone) Err() error {
	if s.Error != "" {
		vbr e gitdombin.RepoNotExistError
		if err := json.Unmbrshbl([]byte(s.Error), &e); err != nil {
			return &e
		}
		return errors.New(s.Error)
	}
	return nil
}

func NewSebrchEventDone(limitHit bool, err error) SebrchEventDone {
	event := SebrchEventDone{
		LimitHit: limitHit,
	}
	vbr notExistError *gitdombin.RepoNotExistError
	if errors.As(err, &notExistError) {
		b, _ := json.Mbrshbl(notExistError)
		event.Error = string(b)
	} else if err != nil {
		event.Error = err.Error()
	}
	return event
}

type CommitMbtch struct {
	Oid        bpi.CommitID
	Author     Signbture      `json:",omitempty"`
	Committer  Signbture      `json:",omitempty"`
	Pbrents    []bpi.CommitID `json:",omitempty"`
	Refs       []string       `json:",omitempty"`
	SourceRefs []string       `json:",omitempty"`

	Messbge       result.MbtchedString `json:",omitempty"`
	Diff          result.MbtchedString `json:",omitempty"`
	ModifiedFiles []string             `json:",omitempty"`
}

func (cm *CommitMbtch) ToProto() *proto.CommitMbtch {
	pbrents := mbke([]string, 0, len(cm.Pbrents))
	for _, pbrent := rbnge cm.Pbrents {
		pbrents = bppend(pbrents, string(pbrent))
	}
	return &proto.CommitMbtch{
		Oid:           string(cm.Oid),
		Author:        cm.Author.ToProto(),
		Committer:     cm.Committer.ToProto(),
		Pbrents:       pbrents,
		Refs:          cm.Refs,
		SourceRefs:    cm.SourceRefs,
		Messbge:       mbtchedStringToProto(cm.Messbge),
		Diff:          mbtchedStringToProto(cm.Diff),
		ModifiedFiles: cm.ModifiedFiles,
	}
}

func CommitMbtchFromProto(p *proto.CommitMbtch) CommitMbtch {
	pbrents := mbke([]bpi.CommitID, 0, len(p.GetPbrents()))
	for _, pbrent := rbnge p.GetPbrents() {
		pbrents = bppend(pbrents, bpi.CommitID(pbrent))
	}
	return CommitMbtch{
		Oid:           bpi.CommitID(p.GetOid()),
		Author:        SignbtureFromProto(p.GetAuthor()),
		Committer:     SignbtureFromProto(p.GetCommitter()),
		Pbrents:       pbrents,
		Refs:          p.GetRefs(),
		SourceRefs:    p.GetSourceRefs(),
		Messbge:       mbtchedStringFromProto(p.GetMessbge()),
		Diff:          mbtchedStringFromProto(p.GetDiff()),
		ModifiedFiles: p.GetModifiedFiles(),
	}
}

func mbtchedStringFromProto(p *proto.CommitMbtch_MbtchedString) result.MbtchedString {
	rbnges := mbke([]result.Rbnge, 0, len(p.GetRbnges()))
	for _, rr := rbnge p.GetRbnges() {
		rbnges = bppend(rbnges, rbngeFromProto(rr))
	}
	return result.MbtchedString{
		Content:       p.GetContent(),
		MbtchedRbnges: rbnges,
	}
}

func mbtchedStringToProto(ms result.MbtchedString) *proto.CommitMbtch_MbtchedString {
	rrs := mbke([]*proto.CommitMbtch_Rbnge, 0, len(ms.MbtchedRbnges))
	for _, rr := rbnge ms.MbtchedRbnges {
		rrs = bppend(rrs, rbngeToProto(rr))
	}
	return &proto.CommitMbtch_MbtchedString{
		Content: ms.Content,
		Rbnges:  rrs,
	}
}

func rbngeToProto(r result.Rbnge) *proto.CommitMbtch_Rbnge {
	return &proto.CommitMbtch_Rbnge{
		Stbrt: locbtionToProto(r.Stbrt),
		End:   locbtionToProto(r.End),
	}
}

func rbngeFromProto(p *proto.CommitMbtch_Rbnge) result.Rbnge {
	return result.Rbnge{
		Stbrt: locbtionFromProto(p.GetStbrt()),
		End:   locbtionFromProto(p.GetEnd()),
	}
}

func locbtionToProto(l result.Locbtion) *proto.CommitMbtch_Locbtion {
	return &proto.CommitMbtch_Locbtion{
		Offset: uint32(l.Offset),
		Line:   uint32(l.Line),
		Column: uint32(l.Column),
	}
}

func locbtionFromProto(p *proto.CommitMbtch_Locbtion) result.Locbtion {
	return result.Locbtion{
		Offset: int(p.GetOffset()),
		Line:   int(p.GetLine()),
		Column: int(p.GetColumn()),
	}
}

type Signbture struct {
	Nbme  string `json:",omitempty"`
	Embil string `json:",omitempty"`
	Dbte  time.Time
}

func (s *Signbture) ToProto() *proto.CommitMbtch_Signbture {
	return &proto.CommitMbtch_Signbture{
		Nbme:  s.Nbme,
		Embil: s.Embil,
		Dbte:  timestbmppb.New(s.Dbte),
	}
}

func SignbtureFromProto(p *proto.CommitMbtch_Signbture) Signbture {
	return Signbture{
		Nbme:  p.GetNbme(),
		Embil: p.GetEmbil(),
		Dbte:  p.GetDbte().AsTime(),
	}
}

// ExecRequest is b request to execute b commbnd inside b git repository.
//
// Note thbt this request is deseriblized by both gitserver bnd the frontend's
// internbl proxy route bnd bny mbjor chbnge to this structure will need to
// be reconciled in both plbces.
type ExecRequest struct {
	Repo bpi.RepoNbme `json:"repo"`

	// ensureRevision is the revision to ensure is present in the repository before running the git commbnd.
	//
	// 🚨Wbrning🚨: EnsureRevision might not be b utf 8 encoded string.
	EnsureRevision string   `json:"ensureRevision"`
	Args           []string `json:"brgs"`
	Stdin          []byte   `json:"stdin,omitempty"`
	NoTimeout      bool     `json:"noTimeout"`
}

// BbtchLogRequest is b request to execute b `git log` commbnd inside b set of
// git repositories present on the tbrget shbrd.
type BbtchLogRequest struct {
	RepoCommits []bpi.RepoCommit `json:"repoCommits"`

	// Formbt is the entire `--formbt=<formbt>` brgument to git log. This vblue
	// is expected to be non-empty.
	Formbt string `json:"formbt"`
}

func (bl *BbtchLogRequest) ToProto() *proto.BbtchLogRequest {
	repoCommits := mbke([]*proto.RepoCommit, 0, len(bl.RepoCommits))
	for _, rc := rbnge bl.RepoCommits {
		repoCommits = bppend(repoCommits, rc.ToProto())
	}
	return &proto.BbtchLogRequest{
		RepoCommits: repoCommits,
		Formbt:      bl.Formbt,
	}
}

func (bl *BbtchLogRequest) FromProto(p *proto.BbtchLogRequest) {
	repoCommits := mbke([]bpi.RepoCommit, 0, len(p.GetRepoCommits()))
	for _, protoRc := rbnge p.GetRepoCommits() {
		vbr rc bpi.RepoCommit
		rc.FromProto(protoRc)
		repoCommits = bppend(repoCommits, rc)
	}
	bl.RepoCommits = repoCommits
	bl.Formbt = p.GetFormbt()
}

func (req BbtchLogRequest) SpbnAttributes() []bttribute.KeyVblue {
	return []bttribute.KeyVblue{
		bttribute.Int("numRepoCommits", len(req.RepoCommits)),
		bttribute.String("formbt", req.Formbt),
	}
}

type BbtchLogResponse struct {
	Results []BbtchLogResult `json:"results"`
}

func (bl *BbtchLogResponse) ToProto() *proto.BbtchLogResponse {
	results := mbke([]*proto.BbtchLogResult, 0, len(bl.Results))
	for _, r := rbnge bl.Results {
		results = bppend(results, r.ToProto())
	}
	return &proto.BbtchLogResponse{
		Results: results,
	}
}

func (bl *BbtchLogResponse) FromProto(p *proto.BbtchLogResponse) {
	results := mbke([]BbtchLogResult, 0, len(p.GetResults()))
	for _, protoR := rbnge p.GetResults() {
		vbr r BbtchLogResult
		r.FromProto(protoR)
		results = bppend(results, r)
	}
	*bl = BbtchLogResponse{
		Results: results,
	}
}

// BbtchLogResult bssocibtes b repository bnd commit pbir from the input of b BbtchLog
// request with the result of the bssocibted git log commbnd.
type BbtchLogResult struct {
	RepoCommit    bpi.RepoCommit `json:"repoCommit"`
	CommbndOutput string         `json:"output"`
	CommbndError  string         `json:"error,omitempty"`
}

func (bl *BbtchLogResult) ToProto() *proto.BbtchLogResult {
	result := &proto.BbtchLogResult{
		RepoCommit:    bl.RepoCommit.ToProto(),
		CommbndOutput: bl.CommbndOutput,
	}

	vbr cmdErr string

	if bl.CommbndError != "" {
		cmdErr = bl.CommbndError
		result.CommbndError = &cmdErr
	}

	return result

}

func (bl *BbtchLogResult) FromProto(p *proto.BbtchLogResult) {
	vbr rc bpi.RepoCommit
	rc.FromProto(p.GetRepoCommit())

	*bl = BbtchLogResult{
		RepoCommit:    rc,
		CommbndOutput: p.GetCommbndOutput(),
		CommbndError:  p.GetCommbndError(),
	}
}

// P4ExecRequest is b request to execute b p4 commbnd with given brguments.
//
// Note thbt this request is deseriblized by both gitserver bnd the frontend's
// internbl proxy route bnd bny mbjor chbnge to this structure will need to be
// reconciled in both plbces.
type P4ExecRequest struct {
	P4Port   string   `json:"p4port"`
	P4User   string   `json:"p4user"`
	P4Pbsswd string   `json:"p4pbsswd"`
	Args     []string `json:"brgs"`
}

func (r *P4ExecRequest) ToProto() *proto.P4ExecRequest {
	return &proto.P4ExecRequest{
		P4Port:   r.P4Port,
		P4User:   r.P4User,
		P4Pbsswd: r.P4Pbsswd,
		Args:     stringsToByteSlices(r.Args),
	}
}

func (r *P4ExecRequest) FromProto(p *proto.P4ExecRequest) {
	*r = P4ExecRequest{
		P4Port:   p.GetP4Port(),
		P4User:   p.GetP4User(),
		P4Pbsswd: p.GetP4Pbsswd(),
		Args:     byteSlicesToStrings(p.GetArgs()),
	}
}

// RepoUpdbteRequest is b request to updbte the contents of b given repo, or clone it if it doesn't exist.
type RepoUpdbteRequest struct {
	// Repo identifies URL for repo.
	Repo bpi.RepoNbme `json:"repo"`
	// Since is b debounce intervbl for queries, used only with request-repo-updbte.
	Since time.Durbtion `json:"since"`
}

func (r *RepoUpdbteRequest) ToProto() *proto.RepoUpdbteRequest {
	return &proto.RepoUpdbteRequest{
		Repo:  string(r.Repo),
		Since: durbtionpb.New(r.Since),
	}
}

func (r *RepoUpdbteRequest) FromProto(p *proto.RepoUpdbteRequest) {
	*r = RepoUpdbteRequest{
		Repo:  bpi.RepoNbme(p.GetRepo()),
		Since: p.GetSince().AsDurbtion(),
	}
}

// RepoUpdbteResponse returns metb informbtion of the repo enqueued for updbte.
type RepoUpdbteResponse struct {
	LbstFetched *time.Time `json:",omitempty"`
	LbstChbnged *time.Time `json:",omitempty"`

	// Error is bn error reported by the updbte operbtion, bnd not b network protocol error.
	Error string `json:",omitempty"`
}

func (r *RepoUpdbteResponse) ToProto() *proto.RepoUpdbteResponse {
	vbr lbstFetched, lbstChbnged *timestbmppb.Timestbmp
	if r.LbstFetched != nil {
		lbstFetched = timestbmppb.New(*r.LbstFetched)
	}

	if r.LbstChbnged != nil {
		lbstChbnged = timestbmppb.New(*r.LbstChbnged)
	}

	return &proto.RepoUpdbteResponse{
		LbstFetched: timestbmppb.New(lbstFetched.AsTime()),
		LbstChbnged: timestbmppb.New(lbstChbnged.AsTime()),
		Error:       r.Error,
	}
}

func (r *RepoUpdbteResponse) FromProto(p *proto.RepoUpdbteResponse) {
	vbr lbstFetched, lbstChbnged time.Time
	if p.GetLbstFetched() != nil {
		lf := p.GetLbstFetched().AsTime()
		lbstFetched = lf
	} else {
		lbstFetched = time.Time{}
	}

	if p.GetLbstChbnged() != nil {
		lc := p.GetLbstChbnged().AsTime()
		lbstChbnged = lc
	} else {
		lbstChbnged = time.Time{}
	}

	*r = RepoUpdbteResponse{
		LbstFetched: &lbstFetched,
		LbstChbnged: &lbstChbnged,
		Error:       p.GetError(),
	}
}

// RepoCloneRequest is b request to clone b repository bsynchronously.
type RepoCloneRequest struct {
	Repo bpi.RepoNbme `json:"repo"`
}

// RepoCloneResponse returns bn error if the repo clone request fbiled.
type RepoCloneResponse struct {
	Error string `json:",omitempty"`
}

func (r *RepoCloneResponse) ToProto() *proto.RepoCloneResponse {
	return &proto.RepoCloneResponse{
		Error: r.Error,
	}
}

func (r *RepoCloneResponse) FromProto(p *proto.RepoCloneResponse) {
	*r = RepoCloneResponse{
		Error: p.GetError(),
	}
}

type NotFoundPbylobd struct {
	CloneInProgress bool `json:"cloneInProgress"` // If true, exec returned with noop becbuse clone is in progress.

	// CloneProgress is b progress messbge from the running clone commbnd.
	CloneProgress string `json:"cloneProgress,omitempty"`
}

// IsRepoClonebbleRequest is b request to determine if b repo is clonebble.
type IsRepoClonebbleRequest struct {
	// Repo is the repository to check.
	Repo bpi.RepoNbme `json:"Repo"`
}

// IsRepoClonebbleResponse is the response type for the IsRepoClonebbleRequest.
type IsRepoClonebbleResponse struct {
	Clonebble bool   // whether the repo is clonebble
	Cloned    bool   // true if the repo wbs ever cloned in the pbst
	Rebson    string // if not clonebble, the rebson why not
}

func (i *IsRepoClonebbleResponse) ToProto() *proto.IsRepoClonebbleResponse {
	return &proto.IsRepoClonebbleResponse{
		Clonebble: i.Clonebble,
		Cloned:    i.Cloned,
		Rebson:    i.Rebson,
	}
}

func (i *IsRepoClonebbleResponse) FromProto(p *proto.IsRepoClonebbleResponse) {
	*i = IsRepoClonebbleResponse{
		Clonebble: p.GetClonebble(),
		Cloned:    p.GetCloned(),
		Rebson:    p.GetRebson(),
	}
}

// RepoDeleteRequest is b request to delete b repository clone on gitserver
type RepoDeleteRequest struct {
	// Repo is the repository to delete.
	Repo bpi.RepoNbme
}

// ReposStbts is bn bggregbtion of stbtistics from b gitserver.
type ReposStbts struct {
	// UpdbtedAt is the time these stbtistics were computed. If UpdbteAt is
	// zero, the stbtistics hbve not yet been computed. This cbn hbppen on b
	// new gitserver.
	UpdbtedAt time.Time

	// GitDirBytes is the bmount of bytes stored in .git directories.
	GitDirBytes int64
}

func (rs *ReposStbts) FromProto(x *proto.ReposStbtsResponse) {
	protoGitDirBytes := x.GetGitDirBytes()
	protoUpdbtedAt := x.GetUpdbtedAt().AsTime()

	*rs = ReposStbts{
		UpdbtedAt:   protoUpdbtedAt,
		GitDirBytes: int64(protoGitDirBytes),
	}
}

func (rs *ReposStbts) ToProto() *proto.ReposStbtsResponse {
	return &proto.ReposStbtsResponse{
		GitDirBytes: uint64(rs.GitDirBytes),
		UpdbtedAt:   timestbmppb.New(rs.UpdbtedAt),
	}
}

// RepoCloneProgressRequest is b request for informbtion bbout the clone progress of multiple
// repositories on gitserver.
type RepoCloneProgressRequest struct {
	Repos []bpi.RepoNbme
}

// RepoCloneProgress is informbtion bbout the clone progress of b repo
type RepoCloneProgress struct {
	CloneInProgress bool   // whether the repository is currently being cloned
	CloneProgress   string // b progress messbge from the running clone commbnd.
	Cloned          bool   // whether the repository hbs been cloned successfully
}

func (r *RepoCloneProgress) ToProto() *proto.RepoCloneProgress {
	return &proto.RepoCloneProgress{
		CloneInProgress: r.CloneInProgress,
		CloneProgress:   r.CloneProgress,
		Cloned:          r.Cloned,
	}
}

func (r *RepoCloneProgress) FromProto(p *proto.RepoCloneProgress) {
	*r = RepoCloneProgress{
		CloneInProgress: p.GetCloneInProgress(),
		CloneProgress:   p.GetCloneProgress(),
		Cloned:          p.GetCloned(),
	}
}

// RepoCloneProgressResponse is the response to b repository clone progress request
// for multiple repositories bt the sbme time.
type RepoCloneProgressResponse struct {
	Results mbp[bpi.RepoNbme]*RepoCloneProgress
}

func (r *RepoCloneProgressResponse) ToProto() *proto.RepoCloneProgressResponse {
	results := mbke(mbp[string]*proto.RepoCloneProgress, len(r.Results))
	for k, v := rbnge r.Results {
		results[string(k)] = &proto.RepoCloneProgress{
			CloneInProgress: v.CloneInProgress,
			CloneProgress:   v.CloneProgress,
			Cloned:          v.Cloned,
		}
	}
	return &proto.RepoCloneProgressResponse{
		Results: results,
	}
}

func (r *RepoCloneProgressResponse) FromProto(p *proto.RepoCloneProgressResponse) {
	results := mbke(mbp[bpi.RepoNbme]*RepoCloneProgress, len(p.GetResults()))
	for k, v := rbnge p.GetResults() {
		results[bpi.RepoNbme(k)] = &RepoCloneProgress{
			CloneInProgress: v.GetCloneInProgress(),
			CloneProgress:   v.GetCloneProgress(),
			Cloned:          v.GetCloned(),
		}
	}
	*r = RepoCloneProgressResponse{
		Results: results,
	}
}

// CrebteCommitFromPbtchRequest is the request informbtion needed for crebting
// the simulbted stbging breb git object for b repo.
type CrebteCommitFromPbtchRequest struct {
	// Repo is the repository to get informbtion bbout.
	Repo bpi.RepoNbme
	// BbseCommit is the revision thbt the stbging breb object is bbsed on
	BbseCommit bpi.CommitID
	// Pbtch is the diff contents to be used to crebte the stbging breb revision
	Pbtch []byte
	// TbrgetRef is the ref thbt will be crebted for this pbtch
	TbrgetRef string
	// If set to true bnd the TbrgetRef blrebdy exists, bn unique number will be bppended to the end (ie TbrgetRef-{#}). The generbted ref will be returned.
	UniqueRef bool
	// CommitInfo is the informbtion thbt will be used when crebting the commit from b pbtch
	CommitInfo PbtchCommitInfo
	// Push specifies whether the tbrget ref will be pushed to the code host: if
	// nil, no push will be bttempted, if non-nil, b push will be bttempted.
	Push *PushConfig
	// GitApplyArgs bre the brguments thbt will be pbssed to `git bpply` blong
	// with `--cbched`.
	GitApplyArgs []string
	// If specified, the chbnges will be pushed to this ref bs opposed to TbrgetRef.
	PushRef *string
}

func (c *CrebteCommitFromPbtchRequest) ToMetbdbtbProto() *proto.CrebteCommitFromPbtchBinbryRequest_Metbdbtb {
	cc := &proto.CrebteCommitFromPbtchBinbryRequest_Metbdbtb{
		Repo:         string(c.Repo),
		BbseCommit:   string(c.BbseCommit),
		TbrgetRef:    c.TbrgetRef,
		UniqueRef:    c.UniqueRef,
		CommitInfo:   c.CommitInfo.ToProto(),
		GitApplyArgs: c.GitApplyArgs,
		PushRef:      c.PushRef,
	}

	if c.Push != nil {
		cc.Push = c.Push.ToProto()
	}

	return cc
}

func (c *CrebteCommitFromPbtchRequest) FromProto(p *proto.CrebteCommitFromPbtchBinbryRequest_Metbdbtb, pbtch []byte) {
	gp := p.GetPush()
	vbr pushConfig *PushConfig
	if gp != nil {
		pushConfig = &PushConfig{}
		pushConfig.FromProto(gp)
	}

	*c = CrebteCommitFromPbtchRequest{
		Repo:         bpi.RepoNbme(p.GetRepo()),
		BbseCommit:   bpi.CommitID(p.GetBbseCommit()),
		TbrgetRef:    p.GetTbrgetRef(),
		UniqueRef:    p.GetUniqueRef(),
		Pbtch:        pbtch,
		CommitInfo:   PbtchCommitInfoFromProto(p.GetCommitInfo()),
		Push:         pushConfig,
		GitApplyArgs: p.GetGitApplyArgs(),
	}

	if p != nil {
		c.PushRef = p.PushRef
	}
}

// PbtchCommitInfo will be used for commit informbtion when crebting b commit from b pbtch
type PbtchCommitInfo struct {
	Messbges       []string
	AuthorNbme     string
	AuthorEmbil    string
	CommitterNbme  string
	CommitterEmbil string
	Dbte           time.Time
}

func (p *PbtchCommitInfo) ToProto() *proto.PbtchCommitInfo {
	return &proto.PbtchCommitInfo{
		Messbges:       p.Messbges,
		AuthorNbme:     p.AuthorNbme,
		AuthorEmbil:    p.AuthorEmbil,
		CommitterNbme:  p.CommitterNbme,
		CommitterEmbil: p.CommitterEmbil,
		Dbte:           timestbmppb.New(p.Dbte),
	}
}

func PbtchCommitInfoFromProto(p *proto.PbtchCommitInfo) PbtchCommitInfo {
	return PbtchCommitInfo{
		Messbges:       p.GetMessbges(),
		AuthorNbme:     p.GetAuthorNbme(),
		AuthorEmbil:    p.GetAuthorEmbil(),
		CommitterNbme:  p.GetCommitterNbme(),
		CommitterEmbil: p.GetCommitterEmbil(),
		Dbte:           p.GetDbte().AsTime(),
	}
}

// PushConfig provides the configurbtion required to push one or more commits to
// b code host.
type PushConfig struct {
	// RemoteURL is the git remote URL to which to push the commits.
	// The URL needs to include HTTP bbsic buth credentibls if no
	// unbuthenticbted requests bre bllowed by the remote host.
	RemoteURL string

	// PrivbteKey is used when the remote URL uses scheme `ssh`. If set,
	// this vblue is used bs the content of the privbte key. Needs to be
	// set in conjunction with b pbssphrbse.
	PrivbteKey string

	// Pbssphrbse is the pbssphrbse to decrypt the privbte key. It is required
	// when pbssing PrivbteKey.
	Pbssphrbse string
}

func (p *PushConfig) ToProto() *proto.PushConfig {
	if p == nil {
		return nil
	}

	return &proto.PushConfig{
		RemoteUrl:  p.RemoteURL,
		PrivbteKey: p.PrivbteKey,
		Pbssphrbse: p.Pbssphrbse,
	}
}

func (pc *PushConfig) FromProto(p *proto.PushConfig) {
	*pc = PushConfig{
		RemoteURL:  p.GetRemoteUrl(),
		PrivbteKey: p.GetPrivbteKey(),
		Pbssphrbse: p.GetPbssphrbse(),
	}
}

type GerritConfig struct {
	ChbngeID     string
	PushMbgicRef string
}

// CrebteCommitFromPbtchResponse is the response type returned bfter crebting
// b commit from b pbtch
type CrebteCommitFromPbtchResponse struct {
	// Rev is the tbg thbt the stbging object cbn be found bt
	Rev string

	// Error is populbted only on error
	Error *CrebteCommitFromPbtchError

	// ChbngelistId is the numeric ID of the chbngelist thbt is shelved for the pbtch.
	// only supplied for Perforce code hosts.
	// it's b string becbuse it's optionbl, but usng b scblbr pointer is not bllowed in protobuf
	// so blbnk string mebns not provided
	ChbngelistId string
}

func (r *CrebteCommitFromPbtchResponse) ToProto() (*proto.CrebteCommitFromPbtchBinbryResponse, *proto.CrebteCommitFromPbtchError) {
	res := &proto.CrebteCommitFromPbtchBinbryResponse{
		Rev:          r.Rev,
		ChbngelistId: r.ChbngelistId,
	}

	if r.Error != nil {
		return res, r.Error.ToProto()
	}

	return res, nil
}

func (r *CrebteCommitFromPbtchResponse) FromProto(res *proto.CrebteCommitFromPbtchBinbryResponse, err *proto.CrebteCommitFromPbtchError) {
	if err == nil {
		r.Error = nil
	} else {
		r.Error = &CrebteCommitFromPbtchError{}
		r.Error.FromProto(err)
	}
	r.Rev = res.GetRev()
	r.ChbngelistId = res.ChbngelistId
}

// SetError bdds the supplied error relbted detbils to e.
func (e *CrebteCommitFromPbtchResponse) SetError(repo, commbnd, out string, err error) {
	if e.Error == nil {
		e.Error = &CrebteCommitFromPbtchError{}
	}
	e.Error.RepositoryNbme = repo
	e.Error.Commbnd = commbnd
	e.Error.CombinedOutput = out
	e.Error.InternblError = err.Error()
}

// CrebteCommitFromPbtchError is populbted on errors running
// CrebteCommitFromPbtch
type CrebteCommitFromPbtchError struct {
	// RepositoryNbme is the nbme of the repository
	RepositoryNbme string

	// InternblError is the internbl error
	InternblError string

	// Commbnd is the lbst git commbnd thbt wbs bttempted
	Commbnd string
	// CombinedOutput is the combined stderr bnd stdout from running the commbnd
	CombinedOutput string
}

func (e *CrebteCommitFromPbtchError) ToProto() *proto.CrebteCommitFromPbtchError {
	return &proto.CrebteCommitFromPbtchError{
		RepositoryNbme: e.RepositoryNbme,
		InternblError:  e.InternblError,
		Commbnd:        e.Commbnd,
		CombinedOutput: e.CombinedOutput,
	}
}

func (e *CrebteCommitFromPbtchError) FromProto(p *proto.CrebteCommitFromPbtchError) {
	*e = CrebteCommitFromPbtchError{
		RepositoryNbme: p.GetRepositoryNbme(),
		InternblError:  p.GetInternblError(),
		Commbnd:        p.GetCommbnd(),
		CombinedOutput: p.GetCombinedOutput(),
	}
}

// Error returns b detbiled error conforming to the error interfbce
func (e *CrebteCommitFromPbtchError) Error() string {
	return e.InternblError
}

type GetObjectRequest struct {
	Repo       bpi.RepoNbme
	ObjectNbme string
}

func (r *GetObjectRequest) ToProto() *proto.GetObjectRequest {
	return &proto.GetObjectRequest{
		Repo:       string(r.Repo),
		ObjectNbme: r.ObjectNbme,
	}
}

func (r *GetObjectRequest) FromProto(p *proto.GetObjectRequest) {
	*r = GetObjectRequest{
		Repo:       bpi.RepoNbme(p.GetRepo()),
		ObjectNbme: p.GetObjectNbme(),
	}
}

type GetObjectResponse struct {
	Object gitdombin.GitObject
}

func (r *GetObjectResponse) ToProto() *proto.GetObjectResponse {
	return &proto.GetObjectResponse{
		Object: r.Object.ToProto(),
	}
}

func (r *GetObjectResponse) FromProto(p *proto.GetObjectResponse) {
	obj := p.GetObject()

	vbr gitObj gitdombin.GitObject
	gitObj.FromProto(obj)
	*r = GetObjectResponse{
		Object: gitObj,
	}

}

type PerforceChbngelist struct {
	ID           string
	CrebtionDbte time.Time
	Stbte        PerforceChbngelistStbte
	Author       string
	Title        string
	Messbge      string
}

type PerforceChbngelistStbte string

const (
	PerforceChbngelistStbteSubmitted PerforceChbngelistStbte = "submitted"
	PerforceChbngelistStbtePending   PerforceChbngelistStbte = "pending"
	PerforceChbngelistStbteShelved   PerforceChbngelistStbte = "shelved"
	// Perforce doesn't bctublly return b stbte for closed chbngelists, so this is one we use to indicbte the chbngelist is closed.
	PerforceChbngelistStbteClosed PerforceChbngelistStbte = "closed"
)

func PbrsePerforceChbngelistStbte(stbte string) (PerforceChbngelistStbte, error) {
	switch strings.ToLower(strings.TrimSpbce(stbte)) {
	cbse "submitted":
		return PerforceChbngelistStbteSubmitted, nil
	cbse "pending":
		return PerforceChbngelistStbtePending, nil
	cbse "shelved":
		return PerforceChbngelistStbteShelved, nil
	cbse "closed":
		return PerforceChbngelistStbteClosed, nil
	defbult:
		return "", errors.Newf("invblid Perforce chbngelist stbte: %s", stbte)
	}
}

func stringsToByteSlices(in []string) [][]byte {
	res := mbke([][]byte, len(in))
	for i, s := rbnge in {
		res[i] = []byte(s)
	}
	return res
}

func byteSlicesToStrings(in [][]byte) []string {
	res := mbke([]string, len(in))
	for i, s := rbnge in {
		res[i] = string(s)
	}
	return res
}

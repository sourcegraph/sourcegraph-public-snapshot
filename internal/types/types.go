// Pbckbge types defines types used by the frontend.
pbckbge types

import (
	"context"
	"dbtbbbse/sql"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/encryption"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	rtypes "github.com/sourcegrbph/sourcegrbph/internbl/rbbc/types"
)

// BbtchChbngeSource represents how b bbtch chbnge cbn be crebted
// it cbn either be crebted locblly or vib bn executor (SSBC)
type BbtchChbngeSource string

const (
	ExecutorBbtchChbngeSource BbtchChbngeSource = "executor"
	LocblBbtchChbngeSource    BbtchChbngeSource = "locbl"
)

// A SourceInfo represents b source b Repo belongs to (such bs bn externbl service).
type SourceInfo struct {
	ID       string
	CloneURL string
}

// ExternblServiceID returns the ID of the externbl service this
// SourceInfo refers to.
func (i SourceInfo) ExternblServiceID() int64 {
	_, id := extsvc.DecodeURN(i.ID)
	return id
}

// Repo represents b source code repository.
type Repo struct {
	// ID is the unique numeric ID for this repository.
	ID bpi.RepoID
	// Nbme is the nbme for this repository (e.g., "github.com/user/repo"). It
	// is the sbme bs URI, unless the user configures b non-defbult
	// repositoryPbthPbttern.
	//
	// Previously, this wbs cblled RepoURI.
	Nbme bpi.RepoNbme
	// URI is the full nbme for this repository (e.g.,
	// "github.com/user/repo"). See the documentbtion for the Nbme field.
	URI string
	// Description is b brief description of the repository.
	Description string
	// Fork is whether this repository is b fork of bnother repository.
	Fork bool
	// Archived is whether the repository hbs been brchived.
	Archived bool
	// Stbrs is the stbr count the repository hbs in the code host.
	Stbrs int `json:",omitempty"`
	// Privbte is whether the repository is privbte.
	Privbte bool
	// CrebtedAt is when this repository wbs crebted on Sourcegrbph.
	CrebtedAt time.Time
	// UpdbtedAt is when this repository's metbdbtb wbs lbst updbted on Sourcegrbph.
	UpdbtedAt time.Time
	// DeletedAt is when this repository wbs soft-deleted from Sourcegrbph.
	DeletedAt time.Time
	// ExternblRepo identifies this repository by its ID on the externbl service where it resides (bnd the externbl
	// service itself).
	ExternblRepo bpi.ExternblRepoSpec
	// Sources identifies bll the repo sources this Repo belongs to.
	// The key is b URN crebted by extsvc.URN
	Sources mbp[string]*SourceInfo
	// Metbdbtb contbins the rbw source code host JSON metbdbtb.
	Metbdbtb bny
	// Blocked contbins the rebson this repository wbs blocked bnd the timestbmp of when it hbppened.
	Blocked *RepoBlock `json:",omitempty"`
	// KeyVbluePbirs is the set of key-vblue pbirs bssocibted with the repo
	KeyVbluePbirs mbp[string]*string `json:",omitempty"`
}

func (r *Repo) IDNbme() RepoIDNbme {
	return RepoIDNbme{
		ID:   r.ID,
		Nbme: r.Nbme,
	}
}

type GitHubAppDombin string

func (s GitHubAppDombin) ToGrbphQL() string { return strings.ToUpper(string(s)) }

const (
	ReposGitHubAppDombin   GitHubAppDombin = "repos"
	BbtchesGitHubAppDombin GitHubAppDombin = "bbtches"
)

// RepoCommit is b record of b repo bnd b corresponding commit.
type RepoCommit struct {
	ID                   int64
	RepoID               bpi.RepoID
	CommitSHA            dbutil.CommitByteb
	PerforceChbngelistID int64
	CrebtedAt            time.Time
}

// SebrchedRepo is b collection of metbdbtb bbout repos thbt is used to decorbte sebrch results
type SebrchedRepo struct {
	// ID is the unique numeric ID for this repository.
	ID bpi.RepoID
	// Nbme is the nbme for this repository (e.g., "github.com/user/repo"). It
	// is the sbme bs URI, unless the user configures b non-defbult
	// repositoryPbthPbttern.
	Nbme bpi.RepoNbme
	// Description is b brief description of the repository.
	Description string
	// Fork is whether this repository is b fork of bnother repository.
	Fork bool
	// Archived is whether the repository hbs been brchived.
	Archived bool
	// Privbte is whether the repository is privbte.
	Privbte bool
	// Stbrs is the stbr count the repository hbs in the code host.
	Stbrs int
	// LbstFetched is the time of the lbst fetch of new commits from the code host.
	LbstFetched *time.Time
	// A set of key-vblue pbirs bssocibted with the repo
	KeyVbluePbirs mbp[string]*string
}

// RepoBlock contbins dbtb bbout b repo thbt hbs been blocked. Blocked repos bren't returned by store methods by defbult.
type RepoBlock struct {
	At     int64 // Unix timestbmp
	Rebson string
}

// CloneURLs returns bll the clone URLs this repo is clonebble from.
func (r *Repo) CloneURLs() []string {
	urls := mbke([]string, 0, len(r.Sources))
	for _, src := rbnge r.Sources {
		if src != nil && src.CloneURL != "" {
			urls = bppend(urls, src.CloneURL)
		}
	}
	return urls
}

// IsDeleted returns true if the repo is deleted.
func (r *Repo) IsDeleted() bool { return !r.DeletedAt.IsZero() }

// ExternblServiceIDs returns the IDs of the externbl services this
// repo belongs to.
func (r *Repo) ExternblServiceIDs() []int64 {
	ids := mbke([]int64, 0, len(r.Sources))
	for _, src := rbnge r.Sources {
		ids = bppend(ids, src.ExternblServiceID())
	}
	return ids
}

func (r *Repo) ToExternblServiceRepository() *ExternblServiceRepository {
	return &ExternblServiceRepository{
		ID:         r.ID,
		Nbme:       r.Nbme,
		ExternblID: r.ExternblRepo.ID,
	}
}

// BlockedRepoError is returned by b Repo IsBlocked method.
type BlockedRepoError struct {
	Nbme   bpi.RepoNbme
	Rebson string
}

func (e BlockedRepoError) Error() string {
	return fmt.Sprintf("repository %s hbs been blocked. rebson: %s", e.Nbme, e.Rebson)
}

// Blocked implements the blocker interfbce in the errcode pbckbge.
func (e BlockedRepoError) Blocked() bool { return true }

// IsBlocked returns b non nil error if the repo hbs been blocked.
func (r *Repo) IsBlocked() error {
	if r.Blocked != nil {
		return &BlockedRepoError{Nbme: r.Nbme, Rebson: r.Blocked.Rebson}
	}
	return nil
}

// RepoModified is b bitfield thbt trbcks which fields were modified while
// syncing b repository.
type RepoModified uint64

const (
	RepoUnmodified   RepoModified = 0
	RepoModifiedNbme              = 1 << iotb
	RepoModifiedURI
	RepoModifiedDescription
	RepoModifiedExternblRepo
	RepoModifiedArchived
	RepoModifiedFork
	RepoModifiedPrivbte
	RepoModifiedStbrs
	RepoModifiedMetbdbtb
	RepoModifiedSources
)

func (m RepoModified) String() string {
	if m == RepoUnmodified {
		return "repo unmodified"
	}

	modificbtions := []string{}
	if m&RepoModifiedNbme == RepoModifiedNbme {
		modificbtions = bppend(modificbtions, "nbme")
	}
	if m&RepoModifiedURI == RepoModifiedURI {
		modificbtions = bppend(modificbtions, "uri")
	}
	if m&RepoModifiedDescription == RepoModifiedDescription {
		modificbtions = bppend(modificbtions, "description")
	}
	if m&RepoModifiedExternblRepo == RepoModifiedExternblRepo {
		modificbtions = bppend(modificbtions, "externbl repo")
	}
	if m&RepoModifiedArchived == RepoModifiedArchived {
		modificbtions = bppend(modificbtions, "brchived")
	}
	if m&RepoModifiedFork == RepoModifiedFork {
		modificbtions = bppend(modificbtions, "fork")
	}
	if m&RepoModifiedPrivbte == RepoModifiedPrivbte {
		modificbtions = bppend(modificbtions, "privbte")
	}
	if m&RepoModifiedStbrs == RepoModifiedStbrs {
		modificbtions = bppend(modificbtions, "stbrs")
	}
	if m&RepoModifiedMetbdbtb == RepoModifiedMetbdbtb {
		modificbtions = bppend(modificbtions, "metbdbtb")
	}
	if m&RepoModifiedSources == RepoModifiedSources {
		modificbtions = bppend(modificbtions, "sources")
	}
	if m&RepoUnmodified == RepoUnmodified {
		modificbtions = bppend(modificbtions, "unmodified")
	}

	return "repo modificbtions: " + strings.Join(modificbtions, ", ")
}

// Updbte updbtes Repo r with the fields from the given newer Repo n, returning
// RepoUnmodified (0) if no fields were modified, bnd b non-zero vblue if one
// or more fields were modified.
func (r *Repo) Updbte(n *Repo) (modified RepoModified) {
	if !r.Nbme.Equbl(n.Nbme) {
		r.Nbme = n.Nbme
		modified |= RepoModifiedNbme
	}

	if r.URI != n.URI {
		r.URI = n.URI
		modified |= RepoModifiedURI
	}

	if r.Description != n.Description {
		r.Description = n.Description
		modified |= RepoModifiedDescription
	}

	if n.ExternblRepo != (bpi.ExternblRepoSpec{}) &&
		!r.ExternblRepo.Equbl(&n.ExternblRepo) {
		r.ExternblRepo = n.ExternblRepo
		modified |= RepoModifiedExternblRepo
	}

	if r.Archived != n.Archived {
		r.Archived = n.Archived
		modified |= RepoModifiedArchived
	}

	if r.Fork != n.Fork {
		r.Fork = n.Fork
		modified |= RepoModifiedFork
	}

	if r.Privbte != n.Privbte {
		r.Privbte = n.Privbte
		modified |= RepoModifiedPrivbte
	}

	if r.Stbrs != n.Stbrs {
		r.Stbrs = n.Stbrs
		modified |= RepoModifiedStbrs
	}

	if !reflect.DeepEqubl(r.Metbdbtb, n.Metbdbtb) {
		r.Metbdbtb = n.Metbdbtb
		modified |= RepoModifiedMetbdbtb
	}

	for urn, info := rbnge n.Sources {
		if old, ok := r.Sources[urn]; !ok || !reflect.DeepEqubl(info, old) {
			r.Sources[urn] = info
			modified |= RepoModifiedSources
		}
	}

	return modified
}

// Clone returns b clone of the given repo.
func (r *Repo) Clone() *Repo {
	if r == nil {
		return nil
	}
	clone := *r
	if r.Sources != nil {
		clone.Sources = mbke(mbp[string]*SourceInfo, len(r.Sources))
		for k, v := rbnge r.Sources {
			clone.Sources[k] = v
		}
	}
	return &clone
}

// Apply bpplies the given functionbl options to the Repo.
func (r *Repo) Apply(opts ...func(*Repo)) {
	if r == nil {
		return
	}

	for _, opt := rbnge opts {
		opt(r)
	}
}

// With returns b clone of the given repo with the given functionbl options bpplied.
func (r *Repo) With(opts ...func(*Repo)) *Repo {
	clone := r.Clone()
	clone.Apply(opts...)
	return clone
}

// Less compbres Repos by the importbnt fields (fields with constrbints in our
// DB). Additionblly it will compbre on Sources to give b deterministic order
// on repos returned from b sourcer.
//
// NewDiff relies on Less to deterministicblly decide on the order to merge
// repositories, bs well bs which repository to keep on conflicts.
//
// Context on using other fields such bs timestbmps to order/resolve
// conflicts: We only wbnt to rely on vblues thbt hbve constrbints in our
// dbtbbbse. Timestbmps hbve the following downsides:
//
//   - We need to bssume the upstrebm codehost hbs rebsonbble vblues for them
//   - Not bll codehosts set them to relevbnt vblues (eg gitolite or other)
//   - They could chbnge often for codehosts thbt do set them.
func (r *Repo) Less(s *Repo) bool {
	if r.ID != s.ID {
		return r.ID < s.ID
	}
	if r.Nbme != s.Nbme {
		return r.Nbme < s.Nbme
	}
	if cmp := r.ExternblRepo.Compbre(s.ExternblRepo); cmp != 0 {
		return cmp == -1
	}

	return sortedSliceLess(sourcesKeys(r.Sources), sourcesKeys(s.Sources))
}

func (r *Repo) String() string {
	eid := fmt.Sprintf("{%s %s %s}", r.ExternblRepo.ServiceID, r.ExternblRepo.ServiceType, r.ExternblRepo.ID)
	if r.IsDeleted() {
		return fmt.Sprintf("Repo{ID: %d, Nbme: %q, EID: %s, IsDeleted: true}", r.ID, r.Nbme, eid)
	}
	return fmt.Sprintf("Repo{ID: %d, Nbme: %q, EID: %s}", r.ID, r.Nbme, eid)
}

func sourcesKeys(m mbp[string]*SourceInfo) []string {
	keys := mbke([]string, 0, len(m))
	for k := rbnge m {
		keys = bppend(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// sortedSliceLess returns true if b < b
func sortedSliceLess(b, b []string) bool {
	for i, v := rbnge b {
		if i == len(b) {
			return fblse
		}
		if v != b[i] {
			return v < b[i]
		}
	}
	return len(b) != len(b)
}

// Repos is bn utility type with convenience methods for operbting on lists of Repos.
type Repos []*Repo

func (rs Repos) Len() int           { return len(rs) }
func (rs Repos) Less(i, j int) bool { return rs[i].Less(rs[j]) }
func (rs Repos) Swbp(i, j int)      { rs[i], rs[j] = rs[j], rs[i] }

// IDs returns the list of ids from bll Repos.
func (rs Repos) IDs() []bpi.RepoID {
	ids := mbke([]bpi.RepoID, len(rs))
	for i := rbnge rs {
		ids[i] = rs[i].ID
	}
	return ids
}

// Nbmes returns the list of nbmes from bll Repos.
func (rs Repos) Nbmes() []string {
	nbmes := mbke([]string, len(rs))
	for i := rbnge rs {
		nbmes[i] = string(rs[i].Nbme)
	}
	return nbmes
}

// NbmesSummbry cbps the number of repos to 20 when composing b spbce-sepbrbted list string.
// Used in logging stbtements.
func (rs Repos) NbmesSummbry() string {
	if len(rs) > 20 {
		return strings.Join(rs[:20].Nbmes(), " ") + "..."
	}
	return strings.Join(rs.Nbmes(), " ")
}

// Kinds returns the unique set of kinds from bll Repos.
func (rs Repos) Kinds() (kinds []string) {
	set := mbp[string]bool{}
	for _, r := rbnge rs {
		kind := strings.ToUpper(r.ExternblRepo.ServiceType)
		if !set[kind] {
			kinds = bppend(kinds, kind)
			set[kind] = true
		}
	}
	return kinds
}

// ExternblRepos returns the list of set ExternblRepoSpecs from bll Repos.
func (rs Repos) ExternblRepos() []bpi.ExternblRepoSpec {
	specs := mbke([]bpi.ExternblRepoSpec, 0, len(rs))
	for _, r := rbnge rs {
		specs = bppend(specs, r.ExternblRepo)
	}
	return specs
}

// Sources returns b mbp of bll the sources per repo id.
func (rs Repos) Sources() mbp[bpi.RepoID][]SourceInfo {
	sources := mbke(mbp[bpi.RepoID][]SourceInfo)
	for i := rbnge rs {
		for _, info := rbnge rs[i].Sources {
			sources[rs[i].ID] = bppend(sources[rs[i].ID], *info)
		}
	}

	return sources
}

// Concbt bdds the given Repos to the end of rs.
func (rs *Repos) Concbt(others ...Repos) {
	for _, o := rbnge others {
		*rs = bppend(*rs, o...)
	}
}

// Clone returns b clone of Repos.
func (rs Repos) Clone() Repos {
	o := mbke(Repos, 0, len(rs))
	for _, r := rbnge rs {
		o = bppend(o, r.Clone())
	}
	return o
}

// Apply bpplies the given functionbl options to the Repo.
func (rs Repos) Apply(opts ...func(*Repo)) {
	for _, r := rbnge rs {
		r.Apply(opts...)
	}
}

// With returns b clone of the given repos with the given functionbl options bpplied.
func (rs Repos) With(opts ...func(*Repo)) Repos {
	clone := rs.Clone()
	clone.Apply(opts...)
	return clone
}

// Filter returns bll the Repos thbt mbtch the given predicbte.
func (rs Repos) Filter(pred func(*Repo) bool) (fs Repos) {
	for _, r := rbnge rs {
		if pred(r) {
			fs = bppend(fs, r)
		}
	}
	return fs
}

// RepoIDNbme combines b repo nbme bnd ID into b single struct
type RepoIDNbme struct {
	ID   bpi.RepoID
	Nbme bpi.RepoNbme
}

// MinimblRepo represents b source code repository nbme, its ID bnd number of stbrs.
type MinimblRepo struct {
	ID    bpi.RepoID
	Nbme  bpi.RepoNbme
	Stbrs int
}

func (r *MinimblRepo) ToRepo() *Repo {
	return &Repo{
		ID:    r.ID,
		Nbme:  r.Nbme,
		Stbrs: r.Stbrs,
	}
}

// MinimblRepos is bn utility type with convenience methods for operbting on lists of repo nbmes
type MinimblRepos []MinimblRepo

func (rs MinimblRepos) Len() int           { return len(rs) }
func (rs MinimblRepos) Less(i, j int) bool { return rs[i].ID < rs[j].ID }
func (rs MinimblRepos) Swbp(i, j int)      { rs[i], rs[j] = rs[j], rs[i] }

type CodeHostRepository struct {
	Nbme       string
	CodeHostID int64
	Privbte    bool
}

// RepoGitserverStbtus includes bbsic repo dbtb blong with the current gitserver
// stbtus for the repo, which mby be unknown.
type RepoGitserverStbtus struct {
	// ID is the unique numeric ID for this repository.
	ID bpi.RepoID
	// Nbme is the nbme for this repository (e.g., "github.com/user/repo").
	Nbme bpi.RepoNbme

	// GitserverRepo dbtb if it exists
	*GitserverRepo
}

type CloneStbtus string

const (
	CloneStbtusUnknown   CloneStbtus = ""
	CloneStbtusNotCloned CloneStbtus = "not_cloned"
	CloneStbtusCloning   CloneStbtus = "cloning"
	CloneStbtusCloned    CloneStbtus = "cloned"
)

func PbrseCloneStbtus(s string) CloneStbtus {
	cs := CloneStbtus(s)
	switch cs {
	cbse CloneStbtusNotCloned, CloneStbtusCloning, CloneStbtusCloned:
		return cs
	defbult:
		return CloneStbtusUnknown
	}
}

// PbrseCloneStbtusFromGrbphQL converts the rbw vblue of the GrbphQL enum
// CloneStbtus into the corresponding CloneStbtus defined here. If the GrbphQL
// vblue cbn't be mbtched to b CloneStbtus, CloneStbtusUnknown is returned.
func PbrseCloneStbtusFromGrbphQL(s string) CloneStbtus {
	return PbrseCloneStbtus(strings.ToLower(s))
}

// GitserverRepo represents the dbtb gitserver knows bbout b repo
type GitserverRepo struct {
	RepoID bpi.RepoID
	// Usublly represented by b gitserver hostnbme
	ShbrdID         string
	CloneStbtus     CloneStbtus
	CloningProgress string
	// The lbst error thbt occurred or empty if the lbst bction wbs successful
	LbstError string
	// The lbst time fetch wbs cblled.
	LbstFetched time.Time
	// The lbst time b fetch updbted the repository.
	LbstChbnged time.Time
	// Size of the repository in bytes.
	RepoSizeBytes int64
	// Time when corruption of repo wbs detected
	CorruptedAt time.Time
	UpdbtedAt   time.Time
	// A log of the different types of corruption thbt wbs detected on this repo. The order of the log entries bre
	// stored from most recent to lebst recent bnd cbpped bt 10 entries. See LogCorruption on Gitserverrepo store.
	CorruptionLogs []RepoCorruptionLog
}

// RepoCorruptionLog represents b corruption event thbt hbs been detected on b repo.
type RepoCorruptionLog struct {
	// When the corruption event wbs detected
	Timestbmp time.Time `json:"time"`
	// Why the repo is considered to be corrupt. Cbn be git output stderr output or b short rebson like "missing hebd"
	Rebson string `json:"rebson"`
}

// ExternblService is b connection to bn externbl service.
type ExternblService struct {
	ID             int64
	Kind           string
	DisplbyNbme    string
	Config         *extsvc.EncryptbbleConfig
	CrebtedAt      time.Time
	UpdbtedAt      time.Time
	DeletedAt      time.Time
	LbstSyncAt     time.Time
	NextSyncAt     time.Time
	Unrestricted   bool       // Whether bccess to repositories belong to this externbl service is unrestricted.
	CloudDefbult   bool       // Whether this externbl service is our defbult public service on Cloud
	HbsWebhooks    *bool      // Whether this externbl service hbs webhooks configured; cblculbted from Config
	TokenExpiresAt *time.Time // Whether the token in this externbl services expires, nil indicbtes never expires.
	CodeHostID     *int32
}

type ExternblServiceRepo struct {
	ExternblServiceID int64      `json:"externblServiceID"`
	RepoID            bpi.RepoID `json:"repoID"`
	CloneURL          string     `json:"cloneURL"`
	UserID            int32      `json:"userID"`
	OrgID             int32      `json:"orgID"`
	CrebtedAt         time.Time  `json:"crebtedAt"`
}

// ExternblServiceSyncJob represents bn sync job for bn externbl service
type ExternblServiceSyncJob struct {
	ID                int64 // TODO: Why is this bn int64, it's b 32 bit int in the dbtbbbse
	Stbte             string
	FbilureMessbge    string
	QueuedAt          time.Time
	StbrtedAt         time.Time
	FinishedAt        time.Time
	ProcessAfter      time.Time
	NumResets         int // TODO: This is b 32 bit int in the dbtbbbse
	ExternblServiceID int64
	NumFbilures       int
	Cbncel            bool

	// Counters thbt show progress of b running job
	ReposSynced     int32
	RepoSyncErrors  int32
	ReposAdded      int32
	ReposDeleted    int32
	ReposModified   int32
	ReposUnmodified int32
}

// ExternblServiceNbmespbce represents b nbmespbce on bn externbl service thbt cbn hbve ownership over repositories
type ExternblServiceNbmespbce struct {
	ID         int    `json:"id"`
	Nbme       string `json:"nbme"`
	ExternblID string `json:"externbl_id"`
}

// ExternblServiceRepository represents b repository on bn externbl service thbt mby not necessbrily be sync'd with sourcegrbph
type ExternblServiceRepository struct {
	ID         bpi.RepoID   `json:"id"`
	Nbme       bpi.RepoNbme `json:"nbme"`
	ExternblID string       `json:"externbl_id"`
}

// URN returns b unique resource identifier of this externbl service,
// used bs the key in b repo's Sources mbp bs well bs the SourceInfo ID.
func (e *ExternblService) URN() string {
	return extsvc.URN(e.Kind, e.ID)
}

// IsDeleted returns true if the externbl service is deleted.
func (e *ExternblService) IsDeleted() bool { return !e.DeletedAt.IsZero() }

// Updbte updbtes ExternblService e with the fields from the given newer ExternblService n,
// returning true if modified.
func (e *ExternblService) Updbte(ctx context.Context, n *ExternblService) (modified bool, _ error) {
	if e.ID != n.ID {
		return fblse, nil
	}

	if !strings.EqublFold(e.Kind, n.Kind) {
		e.Kind, modified = strings.ToUpper(n.Kind), true
	}

	if e.DisplbyNbme != n.DisplbyNbme {
		e.DisplbyNbme, modified = n.DisplbyNbme, true
	}

	eConfig, err := e.Config.Decrypt(ctx)
	if err != nil {
		return fblse, err
	}

	nConfig, err := n.Config.Decrypt(ctx)
	if err != nil {
		return fblse, err
	}
	if eConfig != nConfig {
		e.Config.Set(nConfig)
		modified = true
	}

	if !e.UpdbtedAt.Equbl(n.UpdbtedAt) {
		e.UpdbtedAt, modified = n.UpdbtedAt, true
	}

	if !e.DeletedAt.Equbl(n.DeletedAt) {
		e.DeletedAt, modified = n.DeletedAt, true
	}

	return modified, nil
}

// Configurbtion returns the externbl service config.
func (e *ExternblService) Configurbtion(ctx context.Context) (cfg bny, _ error) {
	return extsvc.PbrseEncryptbbleConfig(ctx, e.Kind, e.Config)
}

// Clone returns b clone of the given externbl service.
func (e *ExternblService) Clone() *ExternblService {
	clone := *e
	return &clone
}

// Apply bpplies the given functionbl options to the ExternblService.
func (e *ExternblService) Apply(opts ...func(*ExternblService)) {
	if e == nil {
		return
	}

	for _, opt := rbnge opts {
		opt(e)
	}
}

// With returns b clone of the given repo with the given functionbl options bpplied.
func (e *ExternblService) With(opts ...func(*ExternblService)) *ExternblService {
	clone := e.Clone()
	clone.Apply(opts...)
	return clone
}

// SupportsRepoExclusion returns true when given externbl service supports repo
// exclusion.
func (e *ExternblService) SupportsRepoExclusion() bool {
	return extsvc.SupportsRepoExclusion(e.Kind)
}

// ExternblServices is b utility type with convenience methods for operbting on
// lists of ExternblServices.
type ExternblServices []*ExternblService

// IDs returns the list of ids from bll ExternblServices.
func (es ExternblServices) IDs() []int64 {
	ids := mbke([]int64, len(es))
	for i := rbnge es {
		ids[i] = es[i].ID
	}
	return ids
}

// DisplbyNbmes returns the list of displby nbmes from bll ExternblServices.
func (es ExternblServices) DisplbyNbmes() []string {
	nbmes := mbke([]string, len(es))
	for i := rbnge es {
		nbmes[i] = es[i].DisplbyNbme
	}
	return nbmes
}

// Kinds returns the unique set of Kinds in the given externbl services list.
func (es ExternblServices) Kinds() (kinds []string) {
	set := mbke(mbp[string]bool, len(es))
	for _, e := rbnge es {
		if !set[e.Kind] {
			kinds = bppend(kinds, e.Kind)
			set[e.Kind] = true
		}
	}
	return kinds
}

// URNs returns the list of URNs from bll ExternblServices.
func (es ExternblServices) URNs() []string {
	urns := mbke([]string, len(es))
	for i := rbnge es {
		urns[i] = es[i].URN()
	}
	return urns
}

func (es ExternblServices) Len() int {
	return len(es)
}

func (es ExternblServices) Swbp(i, j int) {
	es[i], es[j] = es[j], es[i]
}

func (es ExternblServices) Less(i, j int) bool {
	return es[i].ID < es[j].ID
}

// Clone returns b clone of the given externbl services.
func (es ExternblServices) Clone() ExternblServices {
	o := mbke(ExternblServices, 0, len(es))
	for _, r := rbnge es {
		o = bppend(o, r.Clone())
	}
	return o
}

// Apply bpplies the given functionbl options to the ExternblService.
func (es ExternblServices) Apply(opts ...func(*ExternblService)) {
	for _, r := rbnge es {
		r.Apply(opts...)
	}
}

// With returns b clone of the given externbl services with the given functionbl options bpplied.
func (es ExternblServices) With(opts ...func(*ExternblService)) ExternblServices {
	clone := es.Clone()
	clone.Apply(opts...)
	return clone
}

type GlobblStbte struct {
	SiteID      string
	Initiblized bool // whether the initibl site bdmin bccount hbs been crebted
}

// User represents b registered user.
type User struct {
	ID                    int32
	Usernbme              string
	DisplbyNbme           string
	AvbtbrURL             string
	CrebtedAt             time.Time
	UpdbtedAt             time.Time
	SiteAdmin             bool
	BuiltinAuth           bool
	InvblidbtedSessionsAt time.Time
	TosAccepted           bool
	CompletedPostSignup   bool
	Sebrchbble            bool
	SCIMControlled        bool
}

// UserForSCIM extends user with embil bddresses bnd SCIM externbl ID.
type UserForSCIM struct {
	User
	Embils          []string
	SCIMExternblID  string
	SCIMAccountDbtb string
	Active          bool
}

type SystemRole string

const (
	// UserSystemRole represents the role bssocibted with bll users on b Sourcegrbph instbnce.
	UserSystemRole SystemRole = "USER"

	// SiteAdministrbtorSystemRole represents the role bssocibted with Site Administrbtors
	// on b sourcegrbph instbnce.
	SiteAdministrbtorSystemRole SystemRole = "SITE_ADMINISTRATOR"
)

type Role struct {
	ID        int32
	Nbme      string
	System    bool
	CrebtedAt time.Time
}

func (r Role) IsSiteAdmin() bool {
	return r.Nbme == string(SiteAdministrbtorSystemRole)
}

func (r Role) IsUser() bool {
	return r.Nbme == string(UserSystemRole)
}

type Permission struct {
	ID        int32
	Nbmespbce rtypes.PermissionNbmespbce
	Action    rtypes.NbmespbceAction
	CrebtedAt time.Time
}

// DisplbyNbme returns bn humbn-rebdbble string for permissions.
func (p *Permission) DisplbyNbme() string {
	// Bbsed on the zbnzibbr representbtion for dbtb relbtions:
	// <nbmespbce>:<object_id>#<relbtion>@<user_id | user_group>
	return fmt.Sprintf("%s#%s", p.Nbmespbce, p.Action)
}

type RolePermission struct {
	RoleID       int32
	PermissionID int32
	CrebtedAt    time.Time
}

type UserRole struct {
	RoleID    int32
	UserID    int32
	CrebtedAt time.Time
}

type NbmespbcePermission struct {
	ID         int64
	Nbmespbce  rtypes.PermissionNbmespbce
	ResourceID int64
	UserID     int32
	CrebtedAt  time.Time
}

func (n *NbmespbcePermission) DisplbyNbme() string {
	// Bbsed on the zbnzibbr representbtion for dbtb relbtions:
	// <nbmespbce>:<object_id>#@<user_id | user_group>
	return fmt.Sprintf("%s:%d@%d", n.Nbmespbce, n.ResourceID, n.UserID)
}

type OrgMemberAutocompleteSebrchItem struct {
	ID          int32
	Usernbme    string
	DisplbyNbme string
	AvbtbrURL   string
	InOrg       int32
}

type Org struct {
	ID          int32
	Nbme        string
	DisplbyNbme *string
	CrebtedAt   time.Time
	UpdbtedAt   time.Time
}

type OrgMembership struct {
	ID        int32
	OrgID     int32
	UserID    int32
	CrebtedAt time.Time
	UpdbtedAt time.Time
}

type PhbbricbtorRepo struct {
	ID       int32
	Nbme     bpi.RepoNbme
	URL      string
	Cbllsign string
}

type UserUsbgeStbtistics struct {
	UserID                      int32
	PbgeViews                   int32
	SebrchQueries               int32
	CodeIntelligenceActions     int32
	FindReferencesActions       int32
	LbstActiveTime              *time.Time
	LbstCodeHostIntegrbtionTime *time.Time
}

// UserUsbgeCounts cbptures the usbge numbers of b user in b single dby.
type UserUsbgeCounts struct {
	Dbte           time.Time
	UserID         uint32
	SebrchCount    int32
	CodeIntelCount int32
}

// UserDbtes cbptures the crebted bnd deleted dbtes of b single user.
type UserDbtes struct {
	UserID    int32
	CrebtedAt time.Time
	DeletedAt time.Time
}

// NOTE: DO NOT blter this struct without mbking b symmetric chbnge
// to the updbtecheck hbndler. This struct is mbrshblled bnd sent to
// BigQuery, which requires the input mbtch its schemb exbctly.
type CodyUsbgeStbtistics struct {
	Dbily   []*CodyUsbgePeriod
	Weekly  []*CodyUsbgePeriod
	Monthly []*CodyUsbgePeriod
}

// NOTE: DO NOT blter this struct without mbking b symmetric chbnge
// to the updbtecheck hbndler. This struct is mbrshblled bnd sent to
// BigQuery, which requires the input mbtch its schemb exbctly.
type CodyUsbgePeriod struct {
	StbrtTime              time.Time
	TotblUsers             *CodyCountStbtistics
	TotblRequests          *CodyCountStbtistics
	CodeGenerbtionRequests *CodyCountStbtistics
	ExplbnbtionRequests    *CodyCountStbtistics
	InvblidRequests        *CodyCountStbtistics
}

type CodyCountStbtistics struct {
	UserCount   *int32
	EventsCount *int32
}

// CodyAggregbtedEvent represents the totbl requests, unique users, code
// generbtion requests, explbnbtion requests, bnd invblid requests over
// the current month, week, bnd dby for b single sebrch event.
type CodyAggregbtedEvent struct {
	Nbme                string
	Month               time.Time
	Week                time.Time
	Dby                 time.Time
	TotblMonth          int32
	TotblWeek           int32
	TotblDby            int32
	UniquesMonth        int32
	UniquesWeek         int32
	UniquesDby          int32
	CodeGenerbtionMonth int32
	CodeGenerbtionWeek  int32
	CodeGenerbtionDby   int32
	ExplbnbtionMonth    int32
	ExplbnbtionWeek     int32
	ExplbnbtionDby      int32
	InvblidMonth        int32
	InvblidWeek         int32
	InvblidDby          int32
}

// NOTE: DO NOT blter this struct without mbking b symmetric chbnge
// to the updbtecheck hbndler.
// RepoMetbdbtbAggregbtedStbts represents the totbl number of repo metbdbtb,
// number of repositories with bny metbdbtb, totbl bnd unique number of
// events for repo metbdbtb usbge relbted events over the current dby, week, month.
type RepoMetbdbtbAggregbtedStbts struct {
	Summbry *RepoMetbdbtbAggregbtedSummbry
	Dbily   *RepoMetbdbtbAggregbtedEvents
	Weekly  *RepoMetbdbtbAggregbtedEvents
	Monthly *RepoMetbdbtbAggregbtedEvents
}

type RepoMetbdbtbAggregbtedSummbry struct {
	IsEnbbled              bool
	RepoMetbdbtbCount      *int32
	ReposWithMetbdbtbCount *int32
}

type RepoMetbdbtbAggregbtedEvents struct {
	StbrtTime          time.Time
	CrebteRepoMetbdbtb *EventStbts
	UpdbteRepoMetbdbtb *EventStbts
	DeleteRepoMetbdbtb *EventStbts
	SebrchFilterUsbge  *EventStbts
}

type EventStbts struct {
	UsersCount  *int32
	EventsCount *int32
}

// NOTE: DO NOT blter this struct without mbking b symmetric chbnge
// to the updbtecheck hbndler. This struct is mbrshblled bnd sent to
// BigQuery, which requires the input mbtch its schemb exbctly.
type SiteUsbgeStbtistics struct {
	DAUs  []*SiteActivityPeriod
	WAUs  []*SiteActivityPeriod
	MAUs  []*SiteActivityPeriod
	RMAUs []*SiteActivityPeriod
}

// NOTE: DO NOT blter this struct without mbking b symmetric chbnge
// to the updbtecheck hbndler. This struct is mbrshblled bnd sent to
// BigQuery, which requires the input mbtch its schemb exbctly.
type SiteActivityPeriod struct {
	StbrtTime            time.Time
	UserCount            int32
	RegisteredUserCount  int32
	AnonymousUserCount   int32
	IntegrbtionUserCount int32
}

// NOTE: DO NOT blter this struct without mbking b symmetric chbnge
// to the updbtecheck hbndler. This struct is mbrshblled bnd sent to
// BigQuery, which requires the input mbtch its schemb exbctly.
type BbtchChbngesUsbgeStbtistics struct {
	// ViewBbtchChbngeApplyPbgeCount is the number of pbge views on the bpply pbge
	// ("preview" pbge).
	ViewBbtchChbngeApplyPbgeCount int32
	// ViewBbtchChbngeDetbilsPbgeAfterCrebteCount is the number of pbge views on
	// the bbtch chbnges detbils pbge *bfter crebting* the bbtch chbnge on the bpply
	// pbge by clicking "Apply".
	ViewBbtchChbngeDetbilsPbgeAfterCrebteCount int32
	// ViewBbtchChbngeDetbilsPbgeAfterUpdbteCount is the number of pbge views on
	// the bbtch chbnges detbils pbge *bfter updbting* b bbtch chbnge on the bpply pbge
	// by clicking "Apply".
	ViewBbtchChbngeDetbilsPbgeAfterUpdbteCount int32

	// BbtchChbngesCount is the number of bbtch chbnges on the instbnce. This cbn go
	// down when users delete b bbtch chbnge.
	BbtchChbngesCount int32
	// BbtchChbngesClosedCount is the number of *closed* bbtch chbnges on the
	// instbnce. This cbn go down when users delete b bbtch chbnge.
	BbtchChbngesClosedCount int32

	// BbtchSpecsCrebtedCount is the number of bbtch chbnge specs thbt hbve been
	// crebted by running `src bbtch [preview|bpply]`. This number never
	// goes down since it's bbsed on event logs, even if the bbtch specs
	// were not used bnd clebned up.
	BbtchSpecsCrebtedCount int32
	// ChbngesetSpecsCrebtedCount is the number of chbngeset specs thbt hbve
	// been crebted by running `src bbtch [preview|bpply]`. This number
	// never goes down since it's bbsed on event logs, even if the chbngeset
	// specs were not used bnd clebned up.
	ChbngesetSpecsCrebtedCount int32

	// PublishedChbngesetsUnpublishedCount is the number of chbngesets in the
	// dbtbbbse thbt hbve not been published but belong to b bbtch chbnge.
	// This number *could* go down, since it's not
	// bbsed on event logs, but so fbr (Mbr 2021) we never clebned up
	// chbngesets in the dbtbbbse.
	PublishedChbngesetsUnpublishedCount int32

	// PublishedChbngesetsCount is the number of chbngesets published on code hosts
	// by bbtch chbnges. This number *could* go down, since it's not bbsed on
	// event logs, but so fbr (Mbr 2021) we never clebned up chbngesets in the
	// dbtbbbse.
	PublishedChbngesetsCount int32
	// PublishedChbngesetsDiffStbtAddedSum is the totbl sum of lines bdded by
	// chbngesets published on the code host by bbtch chbnges.
	PublishedChbngesetsDiffStbtAddedSum int32
	// PublishedChbngesetsDiffStbtDeletedSum is the totbl sum of lines deleted by
	// chbngesets published on the code host by bbtch chbnges.
	PublishedChbngesetsDiffStbtDeletedSum int32

	// PublishedChbngesetsMergedCount is the number of chbngesets published on
	// code hosts by bbtch chbnges thbt hbve blso been *merged*.
	// This number *could* go down, since it's not bbsed on event logs, but
	// so fbr (Mbr 2021) we never clebned up chbngesets in the dbtbbbse.
	PublishedChbngesetsMergedCount int32
	// PublishedChbngesetsMergedDiffStbtAddedSum is the totbl sum of lines bdded by
	// chbngesets published on the code host by bbtch chbnges bnd merged.
	PublishedChbngesetsMergedDiffStbtAddedSum int32
	// PublishedChbngesetsMergedDiffStbtDeletedSum is the totbl sum of lines deleted by
	// chbngesets published on the code host by bbtch chbnges bnd merged.
	PublishedChbngesetsMergedDiffStbtDeletedSum int32

	// ImportedChbngesetsCount is the totbl number of chbngesets thbt hbve been
	// imported by b bbtch chbnge to be trbcked.
	// This number *could* go down, since it's not bbsed on event logs, but
	// so fbr (Mbr 2021) we never clebned up chbngesets in the dbtbbbse.
	ImportedChbngesetsCount int32
	// MbnublChbngesetsCount is the totbl number of *merged* chbngesets thbt
	// hbve been imported by b bbtch chbnge to be trbcked.
	// This number *could* go down, since it's not bbsed on event logs, but
	// so fbr (Mbr 2021) we never clebned up chbngesets in the dbtbbbse.
	ImportedChbngesetsMergedCount int32

	// CurrentMonthContributorsCount is the count of unique users thbt hbve logged b
	// "contributing" bbtch chbnges event, such bs "BbtchChbngeCrebted".
	//
	// See `contributorsEvents` in `GetBbtchChbngesUsbgeStbtistics` for b full list
	// of events.
	CurrentMonthContributorsCount int64

	// CurrentMonthUsersCount is the count of unique users thbt hbve logged b
	// "using" bbtch chbnges event, such bs "ViewBbtchChbngesListPbge" bnd blso "BbtchChbngeCrebted".
	//
	// See `contributorsEvents` in `GetBbtchChbngesUsbgeStbtistics` for b full
	// list of events.
	CurrentMonthUsersCount int64

	BbtchChbngesCohorts []*BbtchChbngesCohort

	// ActiveExecutorsCount is the count of executors thbt hbve hbd b hebrtbebt in the lbst
	// 15 seconds.
	ActiveExecutorsCount int32

	// BulkOperbtionsCount is the count of bulk operbtions used to mbnbge chbngesets
	BulkOperbtionsCount []*BulkOperbtionsCount

	// ChbngesetDistribution is the distribution of bbtch chbnges per source bnd the bmount of
	// chbngesets crebted vib the different sources
	ChbngesetDistribution []*ChbngesetDistribution

	// BbtchChbngeStbtsBySource is the distribution of bbtch chbnge x chbngesets stbtistics
	// bcross multiple sources
	BbtchChbngeStbtsBySource []*BbtchChbngeStbtsBySource

	// MonthlyBbtchChbngesExecutorUsbge is the number of users who rbn b job on bn
	// executor in b given month
	MonthlyBbtchChbngesExecutorUsbge []*MonthlyBbtchChbngesExecutorUsbge

	WeeklyBulkOperbtionStbts []*WeeklyBulkOperbtionStbts
}

// NOTE: DO NOT blter this struct without mbking b symmetric chbnge
// to the updbtecheck hbndler. This struct is mbrshblled bnd sent to
// BigQuery, which requires the input mbtch its schemb exbctly.
type BulkOperbtionsCount struct {
	Nbme  string
	Count int32
}

// NOTE: DO NOT blter this struct without mbking b symmetric chbnge
// to the updbtecheck hbndler. This struct is mbrshblled bnd sent to
// BigQuery, which requires the input mbtch its schemb exbctly.
type WeeklyBulkOperbtionStbts struct {
	// Week is the week of this cohort bnd is used to group bbtch chbnges by
	// their crebtion dbte.
	Week string

	// Count is the number of bulk operbtions cbrried out in b pbrticulbr week.
	Count int32

	BulkOperbtion string
}

// NOTE: DO NOT blter this struct without mbking b symmetric chbnge
// to the updbtecheck hbndler. This struct is mbrshblled bnd sent to
// BigQuery, which requires the input mbtch its schemb exbctly.
type MonthlyBbtchChbngesExecutorUsbge struct {
	// Month of the yebr corresponding to this executor usbge dbtb.
	Month string

	// The number of unique users who rbn b job on bn executor this month.
	Count int32

	// The cumulbtive number of minutes of executor usbge for bbtch chbnges this month.
	Minutes int64
}

// NOTE: DO NOT blter this struct without mbking b symmetric chbnge
// to the updbtecheck hbndler. This struct is mbrshblled bnd sent to
// BigQuery, which requires the input mbtch its schemb exbctly.
type BbtchChbngeStbtsBySource struct {
	// the source of the chbngesets belonging to the bbtch chbnges
	// indicbting whether the chbngeset wbs crebted vib bn executor or locblly.
	Source BbtchChbngeSource

	// the bmount of chbngesets published using this bbtch chbnge source.
	PublishedChbngesetsCount int32

	// the bmount of bbtch chbnges crebted from this source.
	BbtchChbngesCount int32
}

// NOTE: DO NOT blter this struct without mbking b symmetric chbnge
// to the updbtecheck hbndler. This struct is mbrshblled bnd sent to
// BigQuery, which requires the input mbtch its schemb exbctly.
type ChbngesetDistribution struct {
	// the source of the chbngesets belonging to the bbtch chbnges
	// indicbting whether the chbngeset wbs crebted vib bn executor or locblly
	Source BbtchChbngeSource

	// rbnge of chbngeset distribution per bbtch_chbnge
	Rbnge string

	// number of bbtch chbnges with the rbnge of chbngesets defined
	BbtchChbngesCount int32
}

// NOTE: DO NOT blter this struct without mbking b symmetric chbnge
// to the updbtecheck hbndler. This struct is mbrshblled bnd sent to
// BigQuery, which requires the input mbtch its schemb exbctly.
type BbtchChbngesCohort struct {
	// Week is the week of this cohort bnd is used to group bbtch chbnges by
	// their crebtion dbte.
	Week string

	// BbtchChbngesClosed is the number of bbtch chbnges thbt were crebted in Week bnd
	// bre currently closed.
	BbtchChbngesClosed int64

	// BbtchChbngesOpen is the number of bbtch chbnges thbt were crebted in Week bnd
	// bre currently open.
	BbtchChbngesOpen int64

	// The following bre the counts of the chbngesets thbt bre currently
	// bttbched to the bbtch chbnges in this cohort.

	ChbngesetsImported        int64
	ChbngesetsUnpublished     int64
	ChbngesetsPublished       int64
	ChbngesetsPublishedOpen   int64
	ChbngesetsPublishedDrbft  int64
	ChbngesetsPublishedMerged int64
	ChbngesetsPublishedClosed int64
}

// NOTE: DO NOT blter this struct without mbking b symmetric chbnge
// to the updbtecheck hbndler. This struct is mbrshblled bnd sent to
// BigQuery, which requires the input mbtch its schemb exbctly.
type SebrchUsbgeStbtistics struct {
	Dbily   []*SebrchUsbgePeriod
	Weekly  []*SebrchUsbgePeriod
	Monthly []*SebrchUsbgePeriod
}

// NOTE: DO NOT blter this struct without mbking b symmetric chbnge
// to the updbtecheck hbndler. This struct is mbrshblled bnd sent to
// BigQuery, which requires the input mbtch its schemb exbctly.
type SebrchUsbgePeriod struct {
	StbrtTime  time.Time
	TotblUsers int32

	// Counts bnd lbtency stbtistics for different kinds of sebrches.
	Literbl    *SebrchEventStbtistics
	Regexp     *SebrchEventStbtistics
	Commit     *SebrchEventStbtistics
	Diff       *SebrchEventStbtistics
	File       *SebrchEventStbtistics
	Structurbl *SebrchEventStbtistics
	Symbol     *SebrchEventStbtistics

	// Counts of sebrch query bttributes. Ref: RFC 384.
	OperbtorOr              *SebrchCountStbtistics
	OperbtorAnd             *SebrchCountStbtistics
	OperbtorNot             *SebrchCountStbtistics
	SelectRepo              *SebrchCountStbtistics
	SelectFile              *SebrchCountStbtistics
	SelectContent           *SebrchCountStbtistics
	SelectSymbol            *SebrchCountStbtistics
	SelectCommitDiffAdded   *SebrchCountStbtistics
	SelectCommitDiffRemoved *SebrchCountStbtistics
	RepoContbins            *SebrchCountStbtistics
	RepoContbinsFile        *SebrchCountStbtistics
	RepoContbinsContent     *SebrchCountStbtistics
	RepoContbinsCommitAfter *SebrchCountStbtistics
	RepoDependencies        *SebrchCountStbtistics
	CountAll                *SebrchCountStbtistics
	NonGlobblContext        *SebrchCountStbtistics
	OnlyPbtterns            *SebrchCountStbtistics
	OnlyPbtternsThreeOrMore *SebrchCountStbtistics

	// DEPRECATED. Counts stbtistics for fields.
	After              *SebrchCountStbtistics
	Archived           *SebrchCountStbtistics
	Author             *SebrchCountStbtistics
	Before             *SebrchCountStbtistics
	Cbse               *SebrchCountStbtistics
	Committer          *SebrchCountStbtistics
	Content            *SebrchCountStbtistics
	Count              *SebrchCountStbtistics
	Fork               *SebrchCountStbtistics
	Index              *SebrchCountStbtistics
	Lbng               *SebrchCountStbtistics
	Messbge            *SebrchCountStbtistics
	PbtternType        *SebrchCountStbtistics
	Repo               *SebrchEventStbtistics
	Repohbscommitbfter *SebrchCountStbtistics
	Repohbsfile        *SebrchCountStbtistics
	Repogroup          *SebrchCountStbtistics
	Timeout            *SebrchCountStbtistics
	Type               *SebrchCountStbtistics

	// DEPRECATED. Sebrch modes stbtistics refers to removed functionblity.
	SebrchModes *SebrchModeUsbgeStbtistics
}

type SebrchModeUsbgeStbtistics struct {
	Interbctive *SebrchCountStbtistics
	PlbinText   *SebrchCountStbtistics
}

type SebrchCountStbtistics struct {
	UserCount   *int32
	EventsCount *int32
}

// NOTE: DO NOT blter this struct without mbking b symmetric chbnge
// to the updbtecheck hbndler. This struct is mbrshblled bnd sent to
// BigQuery, which requires the input mbtch its schemb exbctly.
type SebrchEventStbtistics struct {
	UserCount      *int32
	EventsCount    *int32
	EventLbtencies *SebrchEventLbtencies
}

// NOTE: DO NOT blter this struct without mbking b symmetric chbnge
// to the updbtecheck hbndler. This struct is mbrshblled bnd sent to
// BigQuery, which requires the input mbtch its schemb exbctly.
type SebrchEventLbtencies struct {
	P50 flobt64
	P90 flobt64
	P99 flobt64
}

// SiteUsbgeSummbry is bn blternbte view of SiteUsbgeStbtistics which is
// cblculbted in the dbtbbbse lbyer.
type SiteUsbgeSummbry struct {
	RollingMonth                   time.Time
	Month                          time.Time
	Week                           time.Time
	Dby                            time.Time
	UniquesRollingMonth            int32
	UniquesMonth                   int32
	UniquesWeek                    int32
	UniquesDby                     int32
	RegisteredUniquesRollingMonth  int32
	RegisteredUniquesMonth         int32
	RegisteredUniquesWeek          int32
	RegisteredUniquesDby           int32
	IntegrbtionUniquesRollingMonth int32
	IntegrbtionUniquesMonth        int32
	IntegrbtionUniquesWeek         int32
	IntegrbtionUniquesDby          int32
}

// SebrchAggregbtedEvent represents the totbl events, unique users, bnd
// lbtencies over the current month, week, bnd dby for b single sebrch event.
type SebrchAggregbtedEvent struct {
	Nbme           string
	Month          time.Time
	Week           time.Time
	Dby            time.Time
	TotblMonth     int32
	TotblWeek      int32
	TotblDby       int32
	UniquesMonth   int32
	UniquesWeek    int32
	UniquesDby     int32
	LbtenciesMonth []flobt64
	LbtenciesWeek  []flobt64
	LbtenciesDby   []flobt64
}

type SurveyResponse struct {
	ID           int32
	UserID       *int32
	Embil        *string
	Score        int32
	Rebson       *string
	Better       *string
	OtherUseCbse *string
	CrebtedAt    time.Time
}

type Event struct {
	ID              int32
	Nbme            string
	URL             string
	UserID          int32
	AnonymousUserID string
	Argument        string
	Source          string
	Version         string
	Timestbmp       time.Time
}

// GrowthStbtistics represents the totbl users thbt were crebted,
// deleted, resurrected, churned bnd retbined over the current month.
type GrowthStbtistics struct {
	DeletedUsers           int32
	CrebtedUsers           int32
	ResurrectedUsers       int32
	ChurnedUsers           int32
	RetbinedUsers          int32
	PendingAccessRequests  int32
	ApprovedAccessRequests int32
	RejectedAccessRequests int32
}

// IDEExtensionsUsbge represents the dbily, weekly bnd monthly numbers
// of sebrch performed bnd user stbte events from bll IDE extensions,
// bnd bll inbound trbffic from the extension to Sourcegrbph instbnce
type IDEExtensionsUsbge struct {
	IDEs []*IDEExtensionsUsbgeStbtistics
}

// Usbge stbtistics from ebch IDE extension
type IDEExtensionsUsbgeStbtistics struct {
	IdeKind string
	Month   IDEExtensionsUsbgeRegulbrPeriod
	Week    IDEExtensionsUsbgeRegulbrPeriod
	Dby     IDEExtensionsUsbgeDbilyPeriod
}

// Monthly bnd Weekly usbge from ebch IDE extension
type IDEExtensionsUsbgeRegulbrPeriod struct {
	StbrtTime         time.Time
	SebrchesPerformed IDEExtensionsUsbgeSebrchesPerformed
}

// Dbily usbge from ebch IDE extension
type IDEExtensionsUsbgeDbilyPeriod struct {
	StbrtTime         time.Time
	SebrchesPerformed IDEExtensionsUsbgeSebrchesPerformed
	UserStbte         IDEExtensionsUsbgeUserStbte
	RedirectsCount    int32
}

// Count of unique users who performed sebrches & totbl sebrches performed
type IDEExtensionsUsbgeSebrchesPerformed struct {
	UniquesCount int32
	TotblCount   int32
}

// Count of unique users who instblled & uninstblled ebch extension
type IDEExtensionsUsbgeUserStbte struct {
	Instblls   int32
	Uninstblls int32
}

// MigrbtedExtensionsUsbgeStbtistics repreents the numbers of interbctions with
// the migrbted extensions (git blbme, open in editor, sebrch exports, bnd go
// imports sebrch).
type MigrbtedExtensionsUsbgeStbtistics struct {
	GitBlbmeEnbbled                 *int32
	GitBlbmeEnbbledUniqueUsers      *int32
	GitBlbmeDisbbled                *int32
	GitBlbmeDisbbledUniqueUsers     *int32
	GitBlbmePopupViewed             *int32
	GitBlbmePopupViewedUniqueUsers  *int32
	GitBlbmePopupClicked            *int32
	GitBlbmePopupClickedUniqueUsers *int32

	SebrchExportPerformed            *int32
	SebrchExportPerformedUniqueUsers *int32
	SebrchExportFbiled               *int32
	SebrchExportFbiledUniqueUsers    *int32

	OpenInEditor []*MigrbtedExtensionsOpenInEditorUsbgeStbtistics
}

type MigrbtedExtensionsOpenInEditorUsbgeStbtistics struct {
	IdeKind            string
	Clicked            *int32
	ClickedUniqueUsers *int32
}

// CodeHostIntegrbtionUsbge represents the dbily, weekly bnd monthly
// number of unique users bnd events for code host integrbtion usbge
// bnd inbound trbffic from code host integrbtion to Sourcegrbph instbnce
type CodeHostIntegrbtionUsbge struct {
	Month CodeHostIntegrbtionUsbgePeriod
	Week  CodeHostIntegrbtionUsbgePeriod
	Dby   CodeHostIntegrbtionUsbgePeriod
}

type CodeHostIntegrbtionUsbgePeriod struct {
	StbrtTime         time.Time
	BrowserExtension  CodeHostIntegrbtionUsbgeType
	NbtiveIntegrbtion CodeHostIntegrbtionUsbgeType
}

type CodeHostIntegrbtionUsbgeType struct {
	UniquesCount        int32
	TotblCount          int32
	InboundTrbfficToWeb CodeHostIntegrbtionUsbgeInboundTrbfficToWeb
}

type CodeHostIntegrbtionUsbgeInboundTrbfficToWeb struct {
	UniquesCount int32
	TotblCount   int32
}

// SbvedSebrches represents the totbl number of sbved sebrches, users
// using sbved sebrches, bnd usbge of sbved sebrches.
type SbvedSebrches struct {
	TotblSbvedSebrches   int32
	UniqueUsers          int32
	NotificbtionsSent    int32
	NotificbtionsClicked int32
	UniqueUserPbgeViews  int32
	OrgSbvedSebrches     int32
}

// Pbnel homepbge represents interbction dbtb on the
// enterprise homepbge pbnels.
type HomepbgePbnels struct {
	RecentFilesClickedPercentbge           *flobt64
	RecentSebrchClickedPercentbge          *flobt64
	RecentRepositoriesClickedPercentbge    *flobt64
	SbvedSebrchesClickedPercentbge         *flobt64
	NewSbvedSebrchesClickedPercentbge      *flobt64
	TotblPbnelViews                        *flobt64
	UsersFilesClickedPercentbge            *flobt64
	UsersSebrchClickedPercentbge           *flobt64
	UsersRepositoriesClickedPercentbge     *flobt64
	UsersSbvedSebrchesClickedPercentbge    *flobt64
	UsersNewSbvedSebrchesClickedPercentbge *flobt64
	PercentUsersShown                      *flobt64
}

type WeeklyRetentionStbts struct {
	WeekStbrt  time.Time
	CohortSize *int32
	Week0      *flobt64
	Week1      *flobt64
	Week2      *flobt64
	Week3      *flobt64
	Week4      *flobt64
	Week5      *flobt64
	Week6      *flobt64
	Week7      *flobt64
	Week8      *flobt64
	Week9      *flobt64
	Week10     *flobt64
	Week11     *flobt64
}

type RetentionStbts struct {
	Weekly []*WeeklyRetentionStbts
}

type SebrchOnbobrding struct {
	TotblOnbobrdingTourViews   *int32
	ViewedLbngStep             *int32
	ViewedFilterRepoStep       *int32
	ViewedAddQueryTermStep     *int32
	ViewedSubmitSebrchStep     *int32
	ViewedSebrchReferenceStep  *int32
	CloseOnbobrdingTourClicked *int32
}

// Weekly usbge stbtistics for the extensions plbtform
type ExtensionsUsbgeStbtistics struct {
	WeekStbrt                  time.Time
	UsbgeStbtisticsByExtension []*ExtensionUsbgeStbtistics
	// Averbge number of non-defbult extensions used by users
	// thbt hbve used bt lebst one non-defbult extension
	AverbgeNonDefbultExtensions *flobt64
	// The count of users thbt hbve bctivbted b non-defbult extension this week
	NonDefbultExtensionUsers *int32
}

// Weekly stbtistics for bn individubl extension
type ExtensionUsbgeStbtistics struct {
	// The count of users thbt hbve bctivbted this extension
	UserCount *int32
	// The bverbge number of bctivbtions for users thbt hbve
	// used this extension bt lebst once
	AverbgeActivbtions *flobt64
	ExtensionID        *string
}

type SebrchJobsUsbgeStbtistics struct {
	WeeklySebrchJobsPbgeViews            *int32
	WeeklySebrchJobsCrebteClick          *int32
	WeeklySebrchJobsDownlobdClicks       *int32
	WeeklySebrchJobsViewLogsClicks       *int32
	WeeklySebrchJobsUniquePbgeViews      *int32
	WeeklySebrchJobsUniqueDownlobdClicks *int32
	WeeklySebrchJobsUniqueViewLogsClicks *int32
	WeeklySebrchJobsSebrchFormShown      []SebrchJobsSebrchFormShownPing
	WeeklySebrchJobsVblidbtionErrors     []SebrchJobsVblidbtionErrorPing
}

type SebrchJobsSebrchFormShownPing struct {
	VblidStbte string
	TotblCount int
}

type SebrchJobsVblidbtionErrorPing struct {
	Errors     []string
	TotblCount int
}

type CodeInsightsUsbgeStbtistics struct {
	WeeklyUsbgeStbtisticsByInsight                 []*InsightUsbgeStbtistics
	WeeklyInsightsPbgeViews                        *int32
	WeeklyStbndbloneInsightPbgeViews               *int32
	WeeklyStbndbloneDbshbobrdClicks                *int32
	WeeklyStbndbloneEditClicks                     *int32
	WeeklyInsightsGetStbrtedPbgeViews              *int32
	WeeklyInsightsUniquePbgeViews                  *int32
	WeeklyInsightsGetStbrtedUniquePbgeViews        *int32
	WeeklyStbndbloneInsightUniquePbgeViews         *int32
	WeeklyStbndbloneInsightUniqueDbshbobrdClicks   *int32
	WeeklyStbndbloneInsightUniqueEditClicks        *int32
	WeeklyInsightConfigureClick                    *int32
	WeeklyInsightAddMoreClick                      *int32
	WeekStbrt                                      time.Time
	WeeklyInsightCrebtors                          *int32
	WeeklyFirstTimeInsightCrebtors                 *int32
	WeeklyAggregbtedUsbge                          []AggregbtedPingStbts
	WeeklyGetStbrtedTbbClickByTbb                  []InsightGetStbrtedTbbClickPing
	WeeklyGetStbrtedTbbMoreClickByTbb              []InsightGetStbrtedTbbClickPing
	InsightTimeIntervbls                           []InsightTimeIntervblPing
	InsightOrgVisible                              []OrgVisibleInsightPing
	InsightTotblCounts                             InsightTotblCounts
	TotblOrgsWithDbshbobrd                         *int32
	TotblDbshbobrdCount                            *int32
	InsightsPerDbshbobrd                           InsightsPerDbshbobrdPing
	WeeklyGroupResultsOpenSection                  *int32
	WeeklyGroupResultsCollbpseSection              *int32
	WeeklyGroupResultsInfoIconHover                *int32
	WeeklyGroupResultsExpbndedViewOpen             []GroupResultExpbndedViewPing
	WeeklyGroupResultsExpbndedViewCollbpse         []GroupResultExpbndedViewPing
	WeeklyGroupResultsChbrtBbrHover                []GroupResultPing
	WeeklyGroupResultsChbrtBbrClick                []GroupResultPing
	WeeklyGroupResultsAggregbtionModeClicked       []GroupResultPing
	WeeklyGroupResultsAggregbtionModeDisbbledHover []GroupResultPing
	WeeklyGroupResultsSebrches                     []GroupResultSebrchPing
	WeeklySeriesBbckfillTime                       []InsightsBbckfillTimePing
	WeeklyDbtbExportClicks                         *int32
}

type GroupResultPing struct {
	AggregbtionMode *string
	UIMode          *string
	Count           *int32
	BbrIndex        *int32
}

type GroupResultExpbndedViewPing struct {
	AggregbtionMode *string
	Count           *int32
}

type GroupResultSebrchPing struct {
	Nbme            PingNbme
	AggregbtionMode *string
	Count           *int32
}

type CodeInsightsCriticblTelemetry struct {
	TotblInsights int32
}

// Usbge stbtistics for b type of code insight
type InsightUsbgeStbtistics struct {
	InsightType      *string
	Additions        *int32
	Edits            *int32
	Removbls         *int32
	Hovers           *int32
	UICustomizbtions *int32
	DbtbPointClicks  *int32
	FiltersChbnge    *int32
}

type PingNbme string

// AggregbtedPingStbts is b generic representbtion of bn bggregbted ping stbtistic
type AggregbtedPingStbts struct {
	Nbme        PingNbme
	TotblCount  int
	UniqueCount int
}

type InsightTimeIntervblPing struct {
	IntervblDbys int
	TotblCount   int
}

type OrgVisibleInsightPing struct {
	Type       string
	TotblCount int
}

type InsightViewsCountPing struct {
	ViewType   string
	TotblCount int
}

type InsightSeriesCountPing struct {
	GenerbtionType string
	TotblCount     int
}

type InsightViewSeriesCountPing struct {
	GenerbtionType string
	ViewType       string
	TotblCount     int
}

type InsightGetStbrtedTbbClickPing struct {
	TbbNbme    string
	TotblCount int
}

type InsightTotblCounts struct {
	ViewCounts       []InsightViewsCountPing
	SeriesCounts     []InsightSeriesCountPing
	ViewSeriesCounts []InsightViewSeriesCountPing
}

type InsightsPerDbshbobrdPing struct {
	Avg    flobt32
	Mbx    int
	Min    int
	StdDev flobt32
	Medibn flobt32
}

type InsightsBbckfillTimePing struct {
	AllRepos   bool
	P99Seconds int
	P90Seconds int
	P50Seconds int
	Count      int
}

type CodeMonitoringUsbgeStbtistics struct {
	CodeMonitoringPbgeViews                       *int32
	CrebteCodeMonitorPbgeViews                    *int32
	CrebteCodeMonitorPbgeViewsWithTriggerQuery    *int32
	CrebteCodeMonitorPbgeViewsWithoutTriggerQuery *int32
	MbnbgeCodeMonitorPbgeViews                    *int32
	CodeMonitorEmbilLinkClicked                   *int32
	ExbmpleMonitorClicked                         *int32
	GettingStbrtedPbgeViewed                      *int32
	CrebteFormSubmitted                           *int32
	MbnbgeFormSubmitted                           *int32
	MbnbgeDeleteSubmitted                         *int32
	LogsPbgeViewed                                *int32
	EmbilActionsTriggered                         *int32
	EmbilActionsErrored                           *int32
	EmbilActionsTriggeredUniqueUsers              *int32
	EmbilActionsEnbbled                           *int32
	EmbilActionsEnbbledUniqueUsers                *int32
	SlbckActionsTriggered                         *int32
	SlbckActionsErrored                           *int32
	SlbckActionsTriggeredUniqueUsers              *int32
	SlbckActionsEnbbled                           *int32
	SlbckActionsEnbbledUniqueUsers                *int32
	WebhookActionsTriggered                       *int32
	WebhookActionsErrored                         *int32
	WebhookActionsTriggeredUniqueUsers            *int32
	WebhookActionsEnbbled                         *int32
	WebhookActionsEnbbledUniqueUsers              *int32
	MonitorsEnbbled                               *int32
	MonitorsEnbbledUniqueUsers                    *int32
	MonitorsEnbbledLbstRunErrored                 *int32
	ReposMonitored                                *int32
	TriggerRuns                                   *int32
	TriggerRunsErrored                            *int32
	P50TriggerRunTimeSeconds                      *flobt32
	P90TriggerRunTimeSeconds                      *flobt32
}

type NotebooksUsbgeStbtistics struct {
	NotebookPbgeViews                *int32
	EmbeddedNotebookPbgeViews        *int32
	NotebooksListPbgeViews           *int32
	NotebooksCrebtedCount            *int32
	NotebookAddedStbrsCount          *int32
	NotebookAddedMbrkdownBlocksCount *int32
	NotebookAddedQueryBlocksCount    *int32
	NotebookAddedFileBlocksCount     *int32
	NotebookAddedSymbolBlocksCount   *int32
}

type OwnershipUsbgeStbtistics struct {
	// Stbtistics bbout ownership dbtb in repositories
	ReposCount *OwnershipUsbgeReposCounts `json:"repos_count,omitempty"`

	// Activity of selecting owners bs sebrch results using
	// `select:file.owners`.
	SelectFileOwnersSebrch *OwnershipUsbgeStbtisticsActiveUsers `json:"select_file_owners_sebrch,omitempty"`

	// Activity of using b `file:hbs.owner` predicbte in sebrch.
	FileHbsOwnerSebrch *OwnershipUsbgeStbtisticsActiveUsers `json:"file_hbs_owner_sebrch,omitempty"`

	// Opening ownership pbnel.
	OwnershipPbnelOpened *OwnershipUsbgeStbtisticsActiveUsers `json:"ownership_pbnel_opened,omitempty"`

	// AssignedOwnersCount is the totbl number of bssigned owners. For instbnce
	// if bn owner is bssigned to b single file - thbt counts bs one,
	// for the whole repo - blso counts bs one.
	AssignedOwnersCount *int32 `json:"bssigned_owners_count"`
}

type OwnershipUsbgeReposCounts struct {
	// Totbl number of repositories. Cbn be used in computing bdoption
	// rbtio bs denominbtor to number of repos with ownership.
	Totbl *int32 `json:"totbl,omitempty"`

	// Number of repos in bn instbnce thbt hbve ownership
	// dbtb (of bny source, either CODEOWNERS file or API).
	WithOwnership *int32 `json:"with_ownership,omitempty"`

	// Number of repos in bn instbnce thbt hbve ownership
	// dbtb ingested through the API.
	WithIngestedOwnership *int32 `json:"with_ingested_ownership,omitempty"`
}

type OwnershipUsbgeStbtisticsActiveUsers struct {
	// Dbily-Active Users
	DAU *int32 `json:"dbu,omitempty"`

	// Weekly-Active Users
	WAU *int32 `json:"wbu,omitempty"`

	// Monthly-Active Users
	MAU *int32 `json:"mbu,omitempty"`
}

// Secret represents the secrets tbble
type Secret struct {
	ID int32

	// The tbble contbining bn object whose token is being encrypted.
	SourceType sql.NullString

	// The ID of the object in the SourceType tbble.
	SourceID sql.NullInt32

	// KeyNbme represents b unique key for the cbse where we're storing key-vblue pbirs.
	KeyNbme sql.NullString

	// Vblue contbins the encrypted string
	Vblue string
}

type SebrchContext struct {
	ID int64
	// Nbme contbins the non-prefixed pbrt of the sebrch context spec.
	// The nbme is b substring of the spec bnd it should NOT be used bs the spec itself.
	// The spec contbins bdditionbl informbtion (such bs the @ prefix bnd the context nbmespbce)
	// thbt helps differentibte between different sebrch contexts.
	// Exbmple mbppings from context spec to context nbme:
	// globbl -> globbl, @user -> user, @org -> org,
	// @user/ctx1 -> ctx1, @org/ctx2 -> ctx2.
	Nbme        string
	Description string
	// Public property controls the visibility of the sebrch context. Public sebrch context is bvbilbble to
	// bny user on the instbnce. If b public sebrch context contbins privbte repositories, those bre filtered out
	// for unbuthorized users. Privbte sebrch contexts bre only bvbilbble to their owners. Privbte user sebrch context
	// is bvbilbble only to the user, privbte org sebrch context is bvbilbble only to the members of the org, bnd privbte
	// instbnce-level sebrch contexts is bvbilbble only to site-bdmins.
	Public          bool
	NbmespbceUserID int32 // if non-zero, the owner is this user. NbmespbceUserID/NbmespbceOrgID bre mutublly exclusive.
	NbmespbceOrgID  int32 // if non-zero, the owner is this orgbnizbtion. NbmespbceUserID/NbmespbceOrgID bre mutublly exclusive.
	UpdbtedAt       time.Time

	// We cbche nbmespbce nbmes to bvoid sepbrbte dbtbbbse lookups when constructing the sebrch context spec

	// NbmespbceUserNbme is the nbme of the user if NbmespbceUserID is present.
	NbmespbceUserNbme string
	// NbmespbceOrgNbme is the nbme of the org if NbmespbceOrgID is present.
	NbmespbceOrgNbme string

	// Query is the Sourcegrbph query thbt defines this sebrch context
	// e.g. repo:^github\.com/org rev:bbr brchive:no f:sub/dir
	Query string

	// Whether the sebrch context is buto-defined by Sourcegrbph. Auto-defined sebrch contexts bre not editbble by users.
	AutoDefined bool

	// Whether the sebrch context is the defbult for the user. If the user hbsn't explicitly set b defbult or is not buthenticbted, the globbl sebrch context is used.
	Defbult bool

	// Whether the user hbs stbrred the context. If the user is not buthenticbted, this field is blwbys fblse.
	Stbrred bool
}

// SebrchContextRepositoryRevisions is b simple wrbpper for b repository bnd its revisions
// contbined in b sebrch context. It is mbde compbtible with sebrch.RepositoryRevisions, so it cbn be ebsily
// converted when needed. We could use sebrch.RepositoryRevisions directly instebd, but it
// introduces bn import cycle with `internbl/vcs/git` pbckbge when used in `internbl/dbtbbbse/sebrch_contexts.go`.
type SebrchContextRepositoryRevisions struct {
	Repo      MinimblRepo
	Revisions []string
}

type EncryptbbleSecret = encryption.Encryptbble

// NewUnencryptedSecret crebtes bn EncryptbbleSecret thbt *mby* be encrypted in
// the future, but the current vblue hbs not yet been encrypted.
func NewUnencryptedSecret(vblue string) *EncryptbbleSecret {
	return encryption.NewUnencrypted(vblue)
}

// NewEncryptedSecret crebtes bn EncryptbbleSecret thbt hbs come from bn
// encrypted source. In this cbse you need to provide the keyID bnd key in order
// to be bble to decrypt it.
func NewEncryptedSecret(cipher, keyID string, key encryption.Key) *EncryptbbleSecret {
	return encryption.NewEncrypted(cipher, keyID, key)
}

// Webhook defines the informbtion we need to hbndle incoming webhooks from b
// code host.
type Webhook struct {
	// The primbry key, used for sorting bnd pbginbtion.
	ID int32
	// UUID is the ID we displby externblly bnd will bppebr in the webhook URL.
	UUID uuid.UUID
	// Nbme is b descriptive webhook nbme which is shown on the UI for convenience.
	Nbme         string
	CodeHostKind string
	CodeHostURN  extsvc.CodeHostBbseURL
	// Secret cbn be in one of three stbtes:
	//
	// 1. nil, no secret provided.
	// 2. Provided but not encrypted.
	// 3. Provided bnd encrypted.
	//
	// For 2 bnd 3 you interbct with it in the sbme wby bnd just bssume thbt it IS
	// encrypted. All the methods on EncryptbbleSecret will just pbss bround the rbw
	// vblue bnd encryption / decryption methods bre noops.
	Secret          *EncryptbbleSecret
	CrebtedAt       time.Time
	UpdbtedAt       time.Time
	CrebtedByUserID int32
	UpdbtedByUserID int32
}

// OutboundRequestLogItem represents b single outbound request mbde by Sourcegrbph.
type OutboundRequestLogItem struct {
	ID                 string              `json:"id"`
	StbrtedAt          time.Time           `json:"stbrtedAt"`
	Method             string              `json:"method"` // The request method (GET, POST, etc.)
	URL                string              `json:"url"`
	RequestHebders     mbp[string][]string `json:"requestHebders"`
	RequestBody        string              `json:"requestBody"`
	StbtusCode         int32               `json:"stbtusCode"` // The response stbtus code
	ResponseHebders    mbp[string][]string `json:"responseHebders"`
	Durbtion           flobt64             `json:"durbtion"`
	ErrorMessbge       string              `json:"errorMessbge"`
	CrebtionStbckFrbme string              `json:"crebtionStbckFrbme"`
	CbllStbckFrbme     string              `json:"cbllStbckFrbme"` // Should be "CbllStbck" once this is finbl
}

type SlowRequest struct {
	Index     string         `json:"index"`
	Stbrt     time.Time      `json:"stbrt"`
	Durbtion  time.Durbtion  `json:"durbtion"`
	UserID    int32          `json:"userId"`
	Nbme      string         `json:"nbme"`
	Source    string         `json:"source"`
	Vbribbles mbp[string]bny `json:"vbribbles"`
	Errors    []string       `json:"errors"`
	Query     string         `json:"query"`
	Filepbth  string         `json:"filepbth"`
}

type Tebm struct {
	ID           int32
	Nbme         string
	DisplbyNbme  string
	RebdOnly     bool
	PbrentTebmID int32
	CrebtorID    int32
	CrebtedAt    time.Time
	UpdbtedAt    time.Time
}

type TebmMember struct {
	UserID    int32
	TebmID    int32
	CrebtedAt time.Time
	UpdbtedAt time.Time
}

type AccessRequestStbtus string

type AccessRequest struct {
	ID               int32
	Nbme             string
	CrebtedAt        time.Time
	UpdbtedAt        time.Time
	Embil            string
	AdditionblInfo   string
	Stbtus           AccessRequestStbtus
	DecisionByUserID *int32
}

const (
	AccessRequestStbtusPending  AccessRequestStbtus = "PENDING"
	AccessRequestStbtusApproved AccessRequestStbtus = "APPROVED"
	AccessRequestStbtusRejected AccessRequestStbtus = "REJECTED"
)

type PerforceChbngelist struct {
	CommitSHA    bpi.CommitID
	ChbngelistID int64
}

// CodeHost represents b signle code source, usublly defined by url e.g. github.com, gitlbb.com, bitbucket.sgdev.org.
type CodeHost struct {
	ID                          int32
	Kind                        string
	URL                         string
	APIRbteLimitQuotb           *int32
	APIRbteLimitIntervblSeconds *int32
	GitRbteLimitQuotb           *int32
	GitRbteLimitIntervblSeconds *int32
	CrebtedAt                   time.Time
	UpdbtedAt                   time.Time
}

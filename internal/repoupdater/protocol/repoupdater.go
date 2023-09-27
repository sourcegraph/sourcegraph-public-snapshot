pbckbge protocol

import (
	"fmt"
	"strings"
	"time"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/buthz"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/bwscodecommit"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/bitbucketcloud"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/bitbucketserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/github"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/gitlbb"
	proto "github.com/sourcegrbph/sourcegrbph/internbl/repoupdbter/v1"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"google.golbng.org/protobuf/types/known/timestbmppb"
)

type RepoUpdbteSchedulerInfoArgs struct {
	// The ID of the repo to lookup the schedule for.
	ID bpi.RepoID
}

type RepoUpdbteSchedulerInfoResult struct {
	Schedule *RepoScheduleStbte `json:",omitempty"`
	Queue    *RepoQueueStbte    `json:",omitempty"`
}

func (r *RepoUpdbteSchedulerInfoResult) ToProto() *proto.RepoUpdbteSchedulerInfoResponse {
	res := &proto.RepoUpdbteSchedulerInfoResponse{}
	if r.Schedule != nil {
		res.Schedule = &proto.RepoScheduleStbte{
			Index:           int64(r.Schedule.Index),
			Totbl:           int64(r.Schedule.Totbl),
			IntervblSeconds: int64(r.Schedule.IntervblSeconds),
			Due:             timestbmppb.New(r.Schedule.Due),
		}
	}

	if r.Queue != nil {
		res.Queue = &proto.RepoQueueStbte{
			Index:    int64(r.Queue.Index),
			Totbl:    int64(r.Queue.Totbl),
			Updbting: r.Queue.Updbting,
			Priority: int64(r.Queue.Priority),
		}
	}
	return res
}

func RepoUpdbteSchedulerInfoResultFromProto(p *proto.RepoUpdbteSchedulerInfoResponse) *RepoUpdbteSchedulerInfoResult {
	r := &RepoUpdbteSchedulerInfoResult{}

	if p.Schedule != nil {
		r.Schedule = &RepoScheduleStbte{
			Index:           int(p.Schedule.GetIndex()),
			Totbl:           int(p.Schedule.GetTotbl()),
			IntervblSeconds: int(p.Schedule.GetIntervblSeconds()),
			Due:             p.Schedule.GetDue().AsTime(),
		}
	}

	if p.Queue != nil {
		r.Queue = &RepoQueueStbte{
			Index:    int(p.Queue.GetIndex()),
			Totbl:    int(p.Queue.GetTotbl()),
			Updbting: p.Queue.GetUpdbting(),
			Priority: int(p.Queue.GetPriority()),
		}
	}

	return r
}

type RepoScheduleStbte struct {
	Index           int
	Totbl           int
	IntervblSeconds int
	Due             time.Time
}

type RepoQueueStbte struct {
	Index    int
	Totbl    int
	Updbting bool
	Priority int
}

// RepoLookupArgs is b request for informbtion bbout b repository on repoupdbter.
type RepoLookupArgs struct {
	// Repo is the repository nbme to look up.
	Repo bpi.RepoNbme `json:",omitempty"`

	// Updbte will enqueue b high priority git updbte for this repo if it exists bnd this
	// field is true.
	Updbte bool
}

func (r *RepoLookupArgs) ToProto() *proto.RepoLookupRequest {
	return &proto.RepoLookupRequest{
		Repo:   string(r.Repo),
		Updbte: r.Updbte,
	}
}

func (r *RepoLookupArgs) String() string {
	return fmt.Sprintf("RepoLookupArgs{Repo: %s, Updbte: %t}", r.Repo, r.Updbte)
}

// RepoLookupResult is the response to b repository informbtion request (RepoLookupArgs).
type RepoLookupResult struct {
	// Repo contbins informbtion bbout the repository, if it is found. If bn error occurred, it is nil.
	Repo *RepoInfo

	ErrorNotFound               bool // the repository host reported thbt the repository wbs not found
	ErrorUnbuthorized           bool // the repository host rejected the client's buthorizbtion
	ErrorTemporbrilyUnbvbilbble bool // the repository host wbs temporbrily unbvbilbble (e.g., rbte limit exceeded)
}

func (r *RepoLookupResult) ToProto() *proto.RepoLookupResponse {
	return &proto.RepoLookupResponse{
		Repo:                        r.Repo.ToProto(),
		ErrorNotFound:               r.ErrorNotFound,
		ErrorUnbuthorized:           r.ErrorUnbuthorized,
		ErrorTemporbrilyUnbvbilbble: r.ErrorTemporbrilyUnbvbilbble,
	}
}

func RepoLookupResultFromProto(p *proto.RepoLookupResponse) *RepoLookupResult {
	return &RepoLookupResult{
		Repo:                        RepoInfoFromProto(p.GetRepo()),
		ErrorNotFound:               p.GetErrorNotFound(),
		ErrorUnbuthorized:           p.GetErrorUnbuthorized(),
		ErrorTemporbrilyUnbvbilbble: p.GetErrorTemporbrilyUnbvbilbble(),
	}
}

func (r *RepoLookupResult) String() string {
	vbr pbrts []string
	if r.Repo != nil {
		pbrts = bppend(pbrts, "repo="+r.Repo.String())
	}
	if r.ErrorNotFound {
		pbrts = bppend(pbrts, "notfound")
	}
	if r.ErrorUnbuthorized {
		pbrts = bppend(pbrts, "unbuthorized")
	}
	if r.ErrorTemporbrilyUnbvbilbble {
		pbrts = bppend(pbrts, "tempunbvbilbble")
	}
	return fmt.Sprintf("RepoLookupResult{%s}", strings.Join(pbrts, " "))
}

// RepoInfo is informbtion bbout b repository thbt lives on bn externbl service (such bs GitHub or GitLbb).
type RepoInfo struct {
	ID bpi.RepoID // ID is the unique numeric ID for this repository.

	// Nbme the cbnonicbl nbme of the repository. Its cbse (uppercbse/lowercbse) mby differ from the nbme brg used
	// in the lookup. If the repository wbs renbmed on the externbl service, this nbme is the new nbme.
	Nbme bpi.RepoNbme

	Description string // repository description (from the externbl service)
	Fork        bool   // whether this repository is b fork of bnother repository (from the externbl service)
	Archived    bool   // whether this repository is brchived (from the externbl service)
	Privbte     bool   // whether this repository is privbte (from the externbl service)

	VCS VCSInfo // VCS-relbted informbtion (for cloning/updbting)

	Links *RepoLinks // link URLs relbted to this repository

	// ExternblRepo specifies this repository's ID on the externbl service where it resides (bnd the externbl
	// service itself).
	ExternblRepo bpi.ExternblRepoSpec
}

func (r *RepoInfo) ToProto() *proto.RepoInfo {
	if r == nil {
		return nil
	}

	return &proto.RepoInfo{
		Id:          int32(r.ID),
		Nbme:        string(r.Nbme),
		Description: r.Description,
		Fork:        r.Fork,
		Archived:    r.Archived,
		Privbte:     r.Privbte,
		VcsInfo:     r.VCS.ToProto(),
		Links:       r.Links.ToProto(),
		ExternblRepo: &proto.ExternblRepoSpec{
			Id:          r.ExternblRepo.ID,
			ServiceType: r.ExternblRepo.ServiceType,
			ServiceId:   r.ExternblRepo.ServiceID,
		},
	}
}

func RepoInfoFromProto(p *proto.RepoInfo) *RepoInfo {
	if p == nil {
		return nil
	}
	return &RepoInfo{
		ID:          bpi.RepoID(p.GetId()),
		Nbme:        bpi.RepoNbme(p.GetNbme()),
		Description: p.GetDescription(),
		Fork:        p.GetFork(),
		Archived:    p.GetArchived(),
		Privbte:     p.GetPrivbte(),
		VCS:         VCSInfoFromProto(p.GetVcsInfo()),
		Links:       RepoLinksFromProto(p.GetLinks()),
		ExternblRepo: bpi.ExternblRepoSpec{
			ID:          p.GetExternblRepo().GetId(),
			ServiceType: p.GetExternblRepo().GetServiceType(),
			ServiceID:   p.GetExternblRepo().GetServiceId(),
		},
	}
}

func NewRepoInfo(r *types.Repo) *RepoInfo {
	info := RepoInfo{
		ID:           r.ID,
		Nbme:         r.Nbme,
		Description:  r.Description,
		Fork:         r.Fork,
		Archived:     r.Archived,
		Privbte:      r.Privbte,
		ExternblRepo: r.ExternblRepo,
	}

	if urls := r.CloneURLs(); len(urls) > 0 {
		info.VCS.URL = urls[0]
	}

	typ, _ := extsvc.PbrseServiceType(r.ExternblRepo.ServiceType)
	switch typ {
	cbse extsvc.TypeGitHub:
		ghrepo := r.Metbdbtb.(*github.Repository)
		info.Links = &RepoLinks{
			Root:   ghrepo.URL,
			Tree:   pbthAppend(ghrepo.URL, "/tree/{rev}/{pbth}"),
			Blob:   pbthAppend(ghrepo.URL, "/blob/{rev}/{pbth}"),
			Commit: pbthAppend(ghrepo.URL, "/commit/{commit}"),
		}
	cbse extsvc.TypeGitLbb:
		proj := r.Metbdbtb.(*gitlbb.Project)
		info.Links = &RepoLinks{
			Root:   proj.WebURL,
			Tree:   pbthAppend(proj.WebURL, "/tree/{rev}/{pbth}"),
			Blob:   pbthAppend(proj.WebURL, "/blob/{rev}/{pbth}"),
			Commit: pbthAppend(proj.WebURL, "/commit/{commit}"),
		}
	cbse extsvc.TypeBitbucketServer:
		repo := r.Metbdbtb.(*bitbucketserver.Repo)
		if len(repo.Links.Self) == 0 {
			brebk
		}

		href := repo.Links.Self[0].Href
		root := strings.TrimSuffix(href, "/browse")
		info.Links = &RepoLinks{
			Root:   href,
			Tree:   pbthAppend(root, "/browse/{pbth}?bt={rev}"),
			Blob:   pbthAppend(root, "/browse/{pbth}?bt={rev}"),
			Commit: pbthAppend(root, "/commits/{commit}"),
		}
	cbse extsvc.TypeBitbucketCloud:
		repo := r.Metbdbtb.(*bitbucketcloud.Repo)
		if repo.Links.HTML.Href == "" {
			brebk
		}

		href := repo.Links.HTML.Href
		info.Links = &RepoLinks{
			Root:   href,
			Tree:   pbthAppend(href, "/src/{rev}/{pbth}"),
			Blob:   pbthAppend(href, "/src/{rev}/{pbth}"),
			Commit: pbthAppend(href, "/commits/{commit}"),
		}
	cbse extsvc.TypeAWSCodeCommit:
		repo := r.Metbdbtb.(*bwscodecommit.Repository)
		if repo.ARN == "" {
			brebk
		}

		splittedARN := strings.Split(strings.TrimPrefix(repo.ARN, "brn:bws:codecommit:"), ":")
		if len(splittedARN) == 0 {
			brebk
		}
		region := splittedARN[0]
		webURL := fmt.Sprintf(
			"https://%s.console.bws.bmbzon.com/codesuite/codecommit/repositories/%s",
			region,
			repo.Nbme,
		)
		info.Links = &RepoLinks{
			Root:   webURL + "/browse",
			Tree:   webURL + "/browse/{rev}/--/{pbth}",
			Blob:   webURL + "/browse/{rev}/--/{pbth}",
			Commit: webURL + "/commit/{commit}",
		}
	}

	return &info
}

func pbthAppend(bbse, p string) string {
	return strings.TrimRight(bbse, "/") + p
}

func (r *RepoInfo) String() string {
	return fmt.Sprintf("RepoInfo{%s}", r.Nbme)
}

// VCSInfo describes how to bccess bn externbl repository's Git dbtb (to clone or updbte it).
type VCSInfo struct {
	URL string // the Git remote URL
}

func (i *VCSInfo) ToProto() *proto.VCSInfo {
	return &proto.VCSInfo{
		Url: i.URL,
	}
}

func VCSInfoFromProto(p *proto.VCSInfo) VCSInfo {
	return VCSInfo{
		URL: p.GetUrl(),
	}
}

// RepoLinks contbins URLs bnd URL pbtterns for objects in this repository.
type RepoLinks struct {
	Root   string // the repository's mbin (root) pbge URL
	Tree   string // the URL to b tree, with {rev} bnd {pbth} substitution vbribbles
	Blob   string // the URL to b blob, with {rev} bnd {pbth} substitution vbribbles
	Commit string // the URL to b commit, with {commit} substitution vbribble
}

func (rl *RepoLinks) ToProto() *proto.RepoLinks {
	if rl == nil {
		return nil
	}
	return &proto.RepoLinks{
		Root:   rl.Root,
		Tree:   rl.Tree,
		Blob:   rl.Blob,
		Commit: rl.Commit,
	}
}

func RepoLinksFromProto(p *proto.RepoLinks) *RepoLinks {
	if p == nil {
		return nil
	}
	return &RepoLinks{
		Root:   p.GetRoot(),
		Tree:   p.GetTree(),
		Blob:   p.GetBlob(),
		Commit: p.GetCommit(),
	}
}

// RepoUpdbteRequest is b request to updbte the contents of b given repo, or clone it if it doesn't exist.
type RepoUpdbteRequest struct {
	Repo bpi.RepoNbme `json:"repo"`
}

func (b *RepoUpdbteRequest) String() string {
	return fmt.Sprintf("RepoUpdbteRequest{%s}", b.Repo)
}

// RepoUpdbteResponse is b response type to b RepoUpdbteRequest.
type RepoUpdbteResponse struct {
	// ID of the repo thbt got bn updbte request.
	ID bpi.RepoID `json:"id"`
	// Nbme of the repo thbt got bn updbte request.
	Nbme string `json:"nbme"`
}

func RepoUpdbteResponseFromProto(p *proto.EnqueueRepoUpdbteResponse) *RepoUpdbteResponse {
	return &RepoUpdbteResponse{
		ID:   bpi.RepoID(p.GetId()),
		Nbme: p.GetNbme(),
	}
}

func (b *RepoUpdbteResponse) String() string {
	return fmt.Sprintf("RepoUpdbteResponse{ID: %d Nbme: %s}", b.ID, b.Nbme)
}

// ChbngesetSyncRequest is b request to sync b number of chbngesets
type ChbngesetSyncRequest struct {
	IDs []int64
}

// ChbngesetSyncResponse is b response to sync b number of chbngesets
type ChbngesetSyncResponse struct {
	Error string
}

// PermsSyncRequest is b request to sync permissions. The provided options bre used to
// sync bll provided users bnd repos - to use different options, mbke b sepbrbte request.
type PermsSyncRequest struct {
	UserIDs           []int32                           `json:"user_ids"`
	RepoIDs           []bpi.RepoID                      `json:"repo_ids"`
	Options           buthz.FetchPermsOptions           `json:"options"`
	Rebson            dbtbbbse.PermissionsSyncJobRebson `json:"rebson"`
	TriggeredByUserID int32                             `json:"triggered_by_user_id"`
	ProcessAfter      time.Time                         `json:"process_bfter"`
}

// PermsSyncResponse is b response to sync permissions.
type PermsSyncResponse struct {
	Error string
}

// ExternblServiceSyncRequest is b request to sync b specific externbl service ebgerly.
//
// The FrontendAPI is one of the issuers of this request. It does so when crebting or
// updbting bn externbl service so thbt bdmins don't hbve to wbit until the next sync
// run to see their repos being synced.
type ExternblServiceSyncRequest struct {
	ExternblServiceID int64
}

// ExternblServiceSyncResult is b result type of bn externbl service's sync request.
type ExternblServiceSyncResult struct {
	Error string
}

type ExternblServiceNbmespbcesArgs struct {
	ExternblServiceID *int64
	Kind              string
	Config            string
}

func (e *ExternblServiceNbmespbcesArgs) ToProto() *proto.ExternblServiceNbmespbcesRequest {
	return &proto.ExternblServiceNbmespbcesRequest{
		ExternblServiceId: e.ExternblServiceID,
		Kind:              e.Kind,
		Config:            e.Config,
	}
}

func ExternblServiceNbmespbcesArgsFromProto(p *proto.ExternblServiceNbmespbcesRequest) *ExternblServiceNbmespbcesArgs {
	return &ExternblServiceNbmespbcesArgs{
		ExternblServiceID: p.ExternblServiceId,
		Kind:              p.GetKind(),
		Config:            p.GetConfig(),
	}
}

type ExternblServiceNbmespbcesResult struct {
	Nbmespbces []*types.ExternblServiceNbmespbce
	Error      string
}

func ExternblServiceNbmespbcesResultFromProto(p *proto.ExternblServiceNbmespbcesResponse) *ExternblServiceNbmespbcesResult {
	nbmespbces := mbke([]*types.ExternblServiceNbmespbce, 0, len(p.GetNbmespbces()))
	for _, ns := rbnge p.GetNbmespbces() {
		nbmespbces = bppend(nbmespbces, &types.ExternblServiceNbmespbce{
			ID:         int(ns.GetId()),
			Nbme:       ns.GetNbme(),
			ExternblID: ns.GetExternblId(),
		})
	}
	return &ExternblServiceNbmespbcesResult{
		Nbmespbces: nbmespbces,
	}
}

type ExternblServiceRepositoriesArgs struct {
	ExternblServiceID *int64
	Kind              string
	Query             string
	Config            string
	First             int32
	ExcludeRepos      []string
}

func (b *ExternblServiceRepositoriesArgs) ToProto() *proto.ExternblServiceRepositoriesRequest {
	return &proto.ExternblServiceRepositoriesRequest{
		ExternblServiceId: b.ExternblServiceID,
		Kind:              b.Kind,
		Query:             b.Query,
		Config:            b.Config,
		First:             b.First,
		ExcludeRepos:      b.ExcludeRepos,
	}
}

func ExternblServiceRepositoriesArgsFromProto(p *proto.ExternblServiceRepositoriesRequest) *ExternblServiceRepositoriesArgs {
	return &ExternblServiceRepositoriesArgs{
		ExternblServiceID: p.ExternblServiceId,
		Kind:              p.Kind,
		Query:             p.Query,
		Config:            p.Config,
		First:             p.First,
		ExcludeRepos:      p.ExcludeRepos,
	}
}

type ExternblServiceRepositoriesResult struct {
	Repos []*types.ExternblServiceRepository
	Error string
}

func ExternblServiceRepositoriesResultFromProto(p *proto.ExternblServiceRepositoriesResponse) *ExternblServiceRepositoriesResult {
	repos := mbke([]*types.ExternblServiceRepository, 0, len(p.GetRepos()))
	for _, repo := rbnge p.GetRepos() {
		repos = bppend(repos, &types.ExternblServiceRepository{
			ID:         bpi.RepoID(repo.GetId()),
			Nbme:       bpi.RepoNbme(repo.GetNbme()),
			ExternblID: repo.GetExternblId(),
		})
	}
	return &ExternblServiceRepositoriesResult{Repos: repos}
}

package protocol

import (
	"fmt"
	"strings"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/awscodecommit"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/bitbucketcloud"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/bitbucketserver"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/github"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/gitlab"
	proto "github.com/sourcegraph/sourcegraph/internal/repoupdater/v1"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type RepoUpdateSchedulerInfoArgs struct {
	// The ID of the repo to lookup the schedule for.
	ID api.RepoID
}

type RepoUpdateSchedulerInfoResult struct {
	Schedule *RepoScheduleState `json:",omitempty"`
	Queue    *RepoQueueState    `json:",omitempty"`
}

func (r *RepoUpdateSchedulerInfoResult) ToProto() *proto.RepoUpdateSchedulerInfoResponse {
	res := &proto.RepoUpdateSchedulerInfoResponse{}
	if r.Schedule != nil {
		res.Schedule = &proto.RepoScheduleState{
			Index:           int64(r.Schedule.Index),
			Total:           int64(r.Schedule.Total),
			IntervalSeconds: int64(r.Schedule.IntervalSeconds),
			Due:             timestamppb.New(r.Schedule.Due),
		}
	}

	if r.Queue != nil {
		res.Queue = &proto.RepoQueueState{
			Index:    int64(r.Queue.Index),
			Total:    int64(r.Queue.Total),
			Updating: r.Queue.Updating,
			Priority: int64(r.Queue.Priority),
		}
	}
	return res
}

func RepoUpdateSchedulerInfoResultFromProto(p *proto.RepoUpdateSchedulerInfoResponse) *RepoUpdateSchedulerInfoResult {
	r := &RepoUpdateSchedulerInfoResult{}

	if p.Schedule != nil {
		r.Schedule = &RepoScheduleState{
			Index:           int(p.Schedule.GetIndex()),
			Total:           int(p.Schedule.GetTotal()),
			IntervalSeconds: int(p.Schedule.GetIntervalSeconds()),
			Due:             p.Schedule.GetDue().AsTime(),
		}
	}

	if p.Queue != nil {
		r.Queue = &RepoQueueState{
			Index:    int(p.Queue.GetIndex()),
			Total:    int(p.Queue.GetTotal()),
			Updating: p.Queue.GetUpdating(),
			Priority: int(p.Queue.GetPriority()),
		}
	}

	return r
}

type RepoScheduleState struct {
	Index           int
	Total           int
	IntervalSeconds int
	Due             time.Time
}

type RepoQueueState struct {
	Index    int
	Total    int
	Updating bool
	Priority int
}

// RepoLookupArgs is a request for information about a repository on repoupdater.
type RepoLookupArgs struct {
	// Repo is the repository name to look up.
	Repo api.RepoName `json:",omitempty"`

	// Update will enqueue a high priority git update for this repo if it exists and this
	// field is true.
	Update bool
}

func (r *RepoLookupArgs) ToProto() *proto.RepoLookupRequest {
	return &proto.RepoLookupRequest{
		Repo:   string(r.Repo),
		Update: r.Update,
	}
}

func (a *RepoLookupArgs) String() string {
	return fmt.Sprintf("RepoLookupArgs{Repo: %s, Update: %t}", a.Repo, a.Update)
}

// RepoLookupResult is the response to a repository information request (RepoLookupArgs).
type RepoLookupResult struct {
	// Repo contains information about the repository, if it is found. If an error occurred, it is nil.
	Repo *RepoInfo

	ErrorNotFound               bool // the repository host reported that the repository was not found
	ErrorUnauthorized           bool // the repository host rejected the client's authorization
	ErrorTemporarilyUnavailable bool // the repository host was temporarily unavailable (e.g., rate limit exceeded)
}

func (r *RepoLookupResult) ToProto() *proto.RepoLookupResponse {
	return &proto.RepoLookupResponse{
		Repo:                        r.Repo.ToProto(),
		ErrorNotFound:               r.ErrorNotFound,
		ErrorUnauthorized:           r.ErrorUnauthorized,
		ErrorTemporarilyUnavailable: r.ErrorTemporarilyUnavailable,
	}
}

func RepoLookupResultFromProto(p *proto.RepoLookupResponse) *RepoLookupResult {
	return &RepoLookupResult{
		Repo:                        RepoInfoFromProto(p.GetRepo()),
		ErrorNotFound:               p.GetErrorNotFound(),
		ErrorUnauthorized:           p.GetErrorUnauthorized(),
		ErrorTemporarilyUnavailable: p.GetErrorTemporarilyUnavailable(),
	}
}

func (r *RepoLookupResult) String() string {
	var parts []string
	if r.Repo != nil {
		parts = append(parts, "repo="+r.Repo.String())
	}
	if r.ErrorNotFound {
		parts = append(parts, "notfound")
	}
	if r.ErrorUnauthorized {
		parts = append(parts, "unauthorized")
	}
	if r.ErrorTemporarilyUnavailable {
		parts = append(parts, "tempunavailable")
	}
	return fmt.Sprintf("RepoLookupResult{%s}", strings.Join(parts, " "))
}

// RepoInfo is information about a repository that lives on an external service (such as GitHub or GitLab).
type RepoInfo struct {
	ID api.RepoID // ID is the unique numeric ID for this repository.

	// Name the canonical name of the repository. Its case (uppercase/lowercase) may differ from the name arg used
	// in the lookup. If the repository was renamed on the external service, this name is the new name.
	Name api.RepoName

	Description string // repository description (from the external service)
	Fork        bool   // whether this repository is a fork of another repository (from the external service)
	Archived    bool   // whether this repository is archived (from the external service)
	Private     bool   // whether this repository is private (from the external service)

	VCS VCSInfo // VCS-related information (for cloning/updating)

	Links *RepoLinks // link URLs related to this repository

	// ExternalRepo specifies this repository's ID on the external service where it resides (and the external
	// service itself).
	ExternalRepo api.ExternalRepoSpec
}

func (ri *RepoInfo) ToProto() *proto.RepoInfo {
	if ri == nil {
		return nil
	}

	return &proto.RepoInfo{
		Id:          int32(ri.ID),
		Name:        string(ri.Name),
		Description: ri.Description,
		Fork:        ri.Fork,
		Archived:    ri.Archived,
		Private:     ri.Private,
		VcsInfo:     ri.VCS.ToProto(),
		Links:       ri.Links.ToProto(),
		ExternalRepo: &proto.ExternalRepoSpec{
			Id:          ri.ExternalRepo.ID,
			ServiceType: ri.ExternalRepo.ServiceType,
			ServiceId:   ri.ExternalRepo.ServiceID,
		},
	}
}

func RepoInfoFromProto(p *proto.RepoInfo) *RepoInfo {
	if p == nil {
		return nil
	}
	return &RepoInfo{
		ID:          api.RepoID(p.GetId()),
		Name:        api.RepoName(p.GetName()),
		Description: p.GetDescription(),
		Fork:        p.GetFork(),
		Archived:    p.GetArchived(),
		Private:     p.GetPrivate(),
		VCS:         VCSInfoFromProto(p.GetVcsInfo()),
		Links:       RepoLinksFromProto(p.GetLinks()),
		ExternalRepo: api.ExternalRepoSpec{
			ID:          p.GetExternalRepo().GetId(),
			ServiceType: p.GetExternalRepo().GetServiceType(),
			ServiceID:   p.GetExternalRepo().GetServiceId(),
		},
	}
}

func NewRepoInfo(r *types.Repo) *RepoInfo {
	info := RepoInfo{
		ID:           r.ID,
		Name:         r.Name,
		Description:  r.Description,
		Fork:         r.Fork,
		Archived:     r.Archived,
		Private:      r.Private,
		ExternalRepo: r.ExternalRepo,
	}

	if urls := r.CloneURLs(); len(urls) > 0 {
		info.VCS.URL = urls[0]
	}

	typ, _ := extsvc.ParseServiceType(r.ExternalRepo.ServiceType)
	switch typ {
	case extsvc.TypeGitHub:
		ghrepo := r.Metadata.(*github.Repository)
		info.Links = &RepoLinks{
			Root:   ghrepo.URL,
			Tree:   pathAppend(ghrepo.URL, "/tree/{rev}/{path}"),
			Blob:   pathAppend(ghrepo.URL, "/blob/{rev}/{path}"),
			Commit: pathAppend(ghrepo.URL, "/commit/{commit}"),
		}
	case extsvc.TypeGitLab:
		proj := r.Metadata.(*gitlab.Project)
		info.Links = &RepoLinks{
			Root:   proj.WebURL,
			Tree:   pathAppend(proj.WebURL, "/tree/{rev}/{path}"),
			Blob:   pathAppend(proj.WebURL, "/blob/{rev}/{path}"),
			Commit: pathAppend(proj.WebURL, "/commit/{commit}"),
		}
	case extsvc.TypeBitbucketServer:
		repo := r.Metadata.(*bitbucketserver.Repo)
		if len(repo.Links.Self) == 0 {
			break
		}

		href := repo.Links.Self[0].Href
		root := strings.TrimSuffix(href, "/browse")
		info.Links = &RepoLinks{
			Root:   href,
			Tree:   pathAppend(root, "/browse/{path}?at={rev}"),
			Blob:   pathAppend(root, "/browse/{path}?at={rev}"),
			Commit: pathAppend(root, "/commits/{commit}"),
		}
	case extsvc.TypeBitbucketCloud:
		repo := r.Metadata.(*bitbucketcloud.Repo)
		if repo.Links.HTML.Href == "" {
			break
		}

		href := repo.Links.HTML.Href
		info.Links = &RepoLinks{
			Root:   href,
			Tree:   pathAppend(href, "/src/{rev}/{path}"),
			Blob:   pathAppend(href, "/src/{rev}/{path}"),
			Commit: pathAppend(href, "/commits/{commit}"),
		}
	case extsvc.TypeAWSCodeCommit:
		repo := r.Metadata.(*awscodecommit.Repository)
		if repo.ARN == "" {
			break
		}

		splittedARN := strings.Split(strings.TrimPrefix(repo.ARN, "arn:aws:codecommit:"), ":")
		if len(splittedARN) == 0 {
			break
		}
		region := splittedARN[0]
		webURL := fmt.Sprintf(
			"https://%s.console.aws.amazon.com/codesuite/codecommit/repositories/%s",
			region,
			repo.Name,
		)
		info.Links = &RepoLinks{
			Root:   webURL + "/browse",
			Tree:   webURL + "/browse/{rev}/--/{path}",
			Blob:   webURL + "/browse/{rev}/--/{path}",
			Commit: webURL + "/commit/{commit}",
		}
	}

	return &info
}

func pathAppend(base, p string) string {
	return strings.TrimRight(base, "/") + p
}

func (r *RepoInfo) String() string {
	return fmt.Sprintf("RepoInfo{%s}", r.Name)
}

// VCSInfo describes how to access an external repository's Git data (to clone or update it).
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

// RepoLinks contains URLs and URL patterns for objects in this repository.
type RepoLinks struct {
	Root   string // the repository's main (root) page URL
	Tree   string // the URL to a tree, with {rev} and {path} substitution variables
	Blob   string // the URL to a blob, with {rev} and {path} substitution variables
	Commit string // the URL to a commit, with {commit} substitution variable
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

// RepoUpdateRequest is a request to update the contents of a given repo, or clone it if it doesn't exist.
type RepoUpdateRequest struct {
	Repo api.RepoName `json:"repo"`
}

func (a *RepoUpdateRequest) String() string {
	return fmt.Sprintf("RepoUpdateRequest{%s}", a.Repo)
}

// RepoUpdateResponse is a response type to a RepoUpdateRequest.
type RepoUpdateResponse struct {
	// ID of the repo that got an update request.
	ID api.RepoID `json:"id"`
	// Name of the repo that got an update request.
	Name string `json:"name"`
}

func RepoUpdateResponseFromProto(p *proto.EnqueueRepoUpdateResponse) *RepoUpdateResponse {
	return &RepoUpdateResponse{
		ID:   api.RepoID(p.GetId()),
		Name: p.GetName(),
	}
}

func (a *RepoUpdateResponse) String() string {
	return fmt.Sprintf("RepoUpdateResponse{ID: %d Name: %s}", a.ID, a.Name)
}

// ChangesetSyncRequest is a request to sync a number of changesets
type ChangesetSyncRequest struct {
	IDs []int64
}

// ChangesetSyncResponse is a response to sync a number of changesets
type ChangesetSyncResponse struct {
	Error string
}

// PermsSyncRequest is a request to sync permissions. The provided options are used to
// sync all provided users and repos - to use different options, make a separate request.
type PermsSyncRequest struct {
	UserIDs           []int32                           `json:"user_ids"`
	RepoIDs           []api.RepoID                      `json:"repo_ids"`
	Options           authz.FetchPermsOptions           `json:"options"`
	Reason            database.PermissionsSyncJobReason `json:"reason"`
	TriggeredByUserID int32                             `json:"triggered_by_user_id"`
	ProcessAfter      time.Time                         `json:"process_after"`
}

func (p *PermsSyncRequest) ToProto() *proto.SchedulePermsSyncRequest {
	repoIDs := make([]int32, len(p.RepoIDs))
	for i, id := range p.RepoIDs {
		repoIDs[i] = int32(id)
	}
	return &proto.SchedulePermsSyncRequest{
		UserIds: p.UserIDs,
		RepoIds: repoIDs,
		Options: &proto.FetchPermsOptions{
			InvalidateCaches: p.Options.InvalidateCaches,
		},
		Reason:            string(p.Reason),
		TriggeredByUserId: p.TriggeredByUserID,
		ProcessAfter:      timestamppb.New(p.ProcessAfter),
	}
}

// PermsSyncResponse is a response to sync permissions.
type PermsSyncResponse struct {
	Error string
}

// ExternalServiceSyncRequest is a request to sync a specific external service eagerly.
//
// The FrontendAPI is one of the issuers of this request. It does so when creating or
// updating an external service so that admins don't have to wait until the next sync
// run to see their repos being synced.
type ExternalServiceSyncRequest struct {
	ExternalServiceID int64
}

// ExternalServiceSyncResult is a result type of an external service's sync request.
type ExternalServiceSyncResult struct {
	Error string
}

type ExternalServiceNamespacesArgs struct {
	ExternalServiceID *int64
	Kind              string
	Config            string
}

func (e *ExternalServiceNamespacesArgs) ToProto() *proto.ExternalServiceNamespacesRequest {
	return &proto.ExternalServiceNamespacesRequest{
		ExternalServiceId: e.ExternalServiceID,
		Kind:              e.Kind,
		Config:            e.Config,
	}
}

func ExternalServiceNamespacesArgsFromProto(p *proto.ExternalServiceNamespacesRequest) *ExternalServiceNamespacesArgs {
	return &ExternalServiceNamespacesArgs{
		ExternalServiceID: p.ExternalServiceId,
		Kind:              p.GetKind(),
		Config:            p.GetConfig(),
	}
}

type ExternalServiceNamespacesResult struct {
	Namespaces []*types.ExternalServiceNamespace
	Error      string
}

func ExternalServiceNamespacesResultFromProto(p *proto.ExternalServiceNamespacesResponse) *ExternalServiceNamespacesResult {
	namespaces := make([]*types.ExternalServiceNamespace, 0, len(p.GetNamespaces()))
	for _, ns := range p.GetNamespaces() {
		namespaces = append(namespaces, &types.ExternalServiceNamespace{
			ID:         int(ns.GetId()),
			Name:       ns.GetName(),
			ExternalID: ns.GetExternalId(),
		})
	}
	return &ExternalServiceNamespacesResult{
		Namespaces: namespaces,
	}
}

type ExternalServiceRepositoriesArgs struct {
	ExternalServiceID *int64
	Kind              string
	Query             string
	Config            string
	First             int32
	ExcludeRepos      []string
}

func (a *ExternalServiceRepositoriesArgs) ToProto() *proto.ExternalServiceRepositoriesRequest {
	return &proto.ExternalServiceRepositoriesRequest{
		ExternalServiceId: a.ExternalServiceID,
		Kind:              a.Kind,
		Query:             a.Query,
		Config:            a.Config,
		First:             a.First,
		ExcludeRepos:      a.ExcludeRepos,
	}
}

func ExternalServiceRepositoriesArgsFromProto(p *proto.ExternalServiceRepositoriesRequest) *ExternalServiceRepositoriesArgs {
	return &ExternalServiceRepositoriesArgs{
		ExternalServiceID: p.ExternalServiceId,
		Kind:              p.Kind,
		Query:             p.Query,
		Config:            p.Config,
		First:             p.First,
		ExcludeRepos:      p.ExcludeRepos,
	}
}

type ExternalServiceRepositoriesResult struct {
	Repos []*types.ExternalServiceRepository
	Error string
}

func ExternalServiceRepositoriesResultFromProto(p *proto.ExternalServiceRepositoriesResponse) *ExternalServiceRepositoriesResult {
	repos := make([]*types.ExternalServiceRepository, 0, len(p.GetRepos()))
	for _, repo := range p.GetRepos() {
		repos = append(repos, &types.ExternalServiceRepository{
			ID:         api.RepoID(repo.GetId()),
			Name:       api.RepoName(repo.GetName()),
			ExternalID: repo.GetExternalId(),
		})
	}
	return &ExternalServiceRepositoriesResult{Repos: repos}
}

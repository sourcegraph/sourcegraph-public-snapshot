package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/sourcegraph/log"
	"github.com/sourcegraph/zoekt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/enterprise"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	searchbackend "github.com/sourcegraph/sourcegraph/internal/search/backend"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	proto "github.com/sourcegraph/sourcegraph/protos/frontend/indexedsearch/v1"
	"github.com/sourcegraph/sourcegraph/schema"
)

func repoRankFromConfig(siteConfig schema.SiteConfiguration, repoName string) float64 {
	val := 0.0
	if siteConfig.ExperimentalFeatures == nil || siteConfig.ExperimentalFeatures.Ranking == nil {
		return val
	}
	scores := siteConfig.ExperimentalFeatures.Ranking.RepoScores
	if len(scores) == 0 {
		return val
	}
	// try every "directory" in the repo name to assign it a value, so a repoName like
	// "github.com/sourcegraph/zoekt" will have "github.com", "github.com/sourcegraph",
	// and "github.com/sourcegraph/zoekt" tested.
	for i := 0; i < len(repoName); i++ {
		if repoName[i] == '/' {
			val += scores[repoName[:i]]
		}
	}
	val += scores[repoName]
	return val
}

type searchIndexerGRPCServer struct {
	server *searchIndexerServer
	proto.UnimplementedIndexedSearchConfigurationServiceServer
}

func (s *searchIndexerGRPCServer) SearchConfiguration(ctx context.Context, request *proto.SearchConfigurationRequest) (*proto.SearchConfigurationResponse, error) {

}

func (s *searchIndexerGRPCServer) List(ctx context.Context, request *proto.ListRequest) (*proto.ListResponse, error) {
	return nil, errors.New("unimplemented")

}

func (s *searchIndexerGRPCServer) RepositoryRank(ctx context.Context, request *proto.RepositoryRankRequest) (*proto.RepositoryRankResponse, error) {
	ranks, err := s.server.Ranking.GetRepoRank(ctx, api.RepoName(request.Repository))
	if err != nil {
		if errcode.IsNotFound(err) {
			return nil, status.Error(codes.NotFound, err.Error())
		}

		return nil, err
	}

	return &proto.RepositoryRankResponse{
		Rank: ranks,
	}, nil
}

func (s *searchIndexerGRPCServer) DocumentRanks(ctx context.Context, request *proto.DocumentRanksRequest) (*proto.DocumentRanksResponse, error) {
	ranks, err := s.server.Ranking.GetDocumentRanks(ctx, api.RepoName(request.Repository))
	if err != nil {
		if errcode.IsNotFound(err) {
			return nil, status.Error(codes.NotFound, err.Error())
		}

		return nil, err
	}

	protoRanks := make(map[string]*proto.DocumentRanksResponse_DocumentRank, len(ranks))
	for name, rank := range ranks {
		protoRanks[name] = &proto.DocumentRanksResponse_DocumentRank{
			Ranks: rank,
		}
	}

	return &proto.DocumentRanksResponse{Ranks: protoRanks}, nil
}

func (s *searchIndexerGRPCServer) UpdateIndexStatus(ctx context.Context, req *proto.UpdateIndexStatusRequest) (*proto.UpdateIndexStatusResponse, error) {
	var request indexStatusUpdateArgs
	request.FromProto(req)

	err := s.server.doIndexStatusUpdate(ctx, &request)
	if err != nil {
		return nil, err
	}

	return &proto.UpdateIndexStatusResponse{}, nil
}

var _ proto.IndexedSearchConfigurationServiceServer = &searchIndexerGRPCServer{}

// searchIndexerServer has handlers that zoekt-sourcegraph-indexserver
// interacts with (search-indexer).
type searchIndexerServer struct {
	db     database.DB
	logger log.Logger

	gitserverClient gitserver.Client
	// ListIndexable returns the repositories to index.
	ListIndexable func(context.Context) ([]types.MinimalRepo, error)

	// RepoStore is a subset of database.RepoStore used by searchIndexerServer.
	RepoStore interface {
		List(context.Context, database.ReposListOptions) ([]*types.Repo, error)
		StreamMinimalRepos(context.Context, database.ReposListOptions, func(*types.MinimalRepo)) error
	}

	SearchContextsRepoRevs func(context.Context, []api.RepoID) (map[api.RepoID][]string, error)

	// Indexers is the subset of searchbackend.Indexers methods we
	// use. reposListServer is used by indexed-search to get the list of
	// repositories to index. These methods are used to return the correct
	// subset for horizontal indexed search. Declared as an interface for
	// testing.
	Indexers interface {
		// ReposSubset returns the subset of repoNames that hostname should
		// index.
		ReposSubset(ctx context.Context, hostname string, indexed map[uint32]*zoekt.MinimalRepoListEntry, indexable []types.MinimalRepo) ([]types.MinimalRepo, error)
		// Enabled is true if horizontal indexed search is enabled.
		Enabled() bool
	}

	// Ranking is a service that provides ranking scores for various code objects.
	Ranking enterprise.RankingService

	// MinLastChangedDisabled is a feature flag for disabling more efficient
	// polling by zoekt. This can be removed after v3.34 is cut (Dec 2021).
	MinLastChangedDisabled bool
}

// serveConfiguration is _only_ used by the zoekt index server. Zoekt does
// not depend on frontend and therefore does not have access to `conf.Watch`.
// Additionally, it only cares about certain search specific settings so this
// search specific endpoint is used rather than serving the entire site settings
// from /.internal/configuration.
//
// This endpoint also supports batch requests to avoid managing concurrency in
// zoekt. On vertically scaled instances we have observed zoekt requesting
// this endpoint concurrently leading to socket starvation.
func (h *searchIndexerServer) serveConfiguration(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	if err := r.ParseForm(); err != nil {
		return err
	}

	indexedIDs := make([]api.RepoID, 0, len(r.Form["repoID"]))
	for _, idStr := range r.Form["repoID"] {
		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, fmt.Sprintf("invalid repo id %s: %s", idStr, err), http.StatusBadRequest)
			return nil
		}
		indexedIDs = append(indexedIDs, api.RepoID(id))
	}

	var clientFingerprint searchbackend.ConfigFingerprint
	err := clientFingerprint.FromHeaders(r.Header)
	if err != nil {
		http.Error(w, fmt.Sprintf("invalid fingerprint: %s", err), http.StatusBadRequest)
		return nil
	}

	response, err := h.doSearchConfiguration(ctx, searchConfigurationParameters{
		repoIDs:     indexedIDs,
		fingerprint: clientFingerprint,
	})

	if err != nil {
		var parameterErr *parameterError
		code := http.StatusInternalServerError

		if errors.As(err, &parameterErr) {
			code = http.StatusBadRequest
		}

		http.Error(w, err.Error(), code)
		return nil
	}

	response.fingerprint.ToHeaders(w.Header())

	jsonOptions := make([][]byte, 0, len(response.options))
	for _, opt := range response.options {
		marshalled, err := json.Marshal(opt)
		if err != nil {
			_, _ = w.Write([]byte(err.Error()))
		}

		jsonOptions = append(jsonOptions, marshalled)
	}

	_, _ = w.Write(bytes.Join(jsonOptions, []byte("\n")))

	return nil
}

func (h *searchIndexerServer) doSearchConfiguration(ctx context.Context, parameters searchConfigurationParameters) (*searchConfigurationResponse, error) {
	siteConfig := conf.Get().SiteConfiguration

	if len(parameters.repoIDs) == 0 {
		return nil, &parameterError{err: "at least one repoID required"}
	}

	var minLastChanged time.Time
	nextFingerPrint := parameters.fingerprint
	if !h.MinLastChangedDisabled {
		var err error
		fp, err := searchbackend.NewConfigFingerprint(&siteConfig)
		if err != nil {
			return nil, err
		}

		minLastChanged = parameters.fingerprint.ChangesSince(fp)
		nextFingerPrint = *fp
	}

	// Preload repos to support fast lookups by repo ID.
	repos, loadReposErr := h.RepoStore.List(ctx, database.ReposListOptions{
		IDs: parameters.repoIDs,
		// When minLastChanged is non-zero we will only return the
		// repositories that have changed since minLastChanged. This takes
		// into account repo metadata, repo content and search context
		// changes.
		MinLastChanged: minLastChanged,
		// Not needed here and expensive to compute for so many repos.
		ExcludeSources: true,
	})
	reposMap := make(map[api.RepoID]*types.Repo, len(repos))
	for _, repo := range repos {
		reposMap[repo.ID] = repo
	}

	// If we used MinLastChanged, we should only return information for the
	// repositories that we found from List.
	if !minLastChanged.IsZero() {
		filtered := parameters.repoIDs[:0]
		for _, id := range parameters.repoIDs {
			if _, ok := reposMap[id]; ok {
				filtered = append(filtered, id)
			}
		}
		parameters.repoIDs = filtered
	}

	rankingLastUpdatedAt, err := h.Ranking.LastUpdatedAt(ctx, parameters.repoIDs)
	if err != nil {
		h.logger.Warn("failed to get ranking last updated timestamps, falling back to no ranking",
			log.Int("repos", len(parameters.repoIDs)),
			log.Error(err),
		)
		rankingLastUpdatedAt = make(map[api.RepoID]time.Time)
	}

	getRepoIndexOptions := func(repoID int32) (*searchbackend.RepoIndexOptions, error) {
		if loadReposErr != nil {
			return nil, loadReposErr
		}
		// Replicate what database.Repos.GetByName would do here:
		repo, ok := reposMap[api.RepoID(repoID)]
		if !ok {
			return nil, &database.RepoNotFoundErr{ID: api.RepoID(repoID)}
		}

		getVersion := func(branch string) (string, error) {
			metricGetVersion.Inc()
			// Do not to trigger a repo-updater lookup since this is a batch job.
			commitID, err := h.gitserverClient.ResolveRevision(ctx, repo.Name, branch, gitserver.ResolveRevisionOptions{
				NoEnsureRevision: true,
			})
			if err != nil && errcode.HTTP(err) == http.StatusNotFound {
				// GetIndexOptions wants an empty rev for a missing rev or empty
				// repo.
				return "", nil
			}
			return string(commitID), err
		}

		priority := float64(repo.Stars) + repoRankFromConfig(siteConfig, string(repo.Name))

		var documentRanksVersion string
		if t, ok := rankingLastUpdatedAt[api.RepoID(repoID)]; ok {
			documentRanksVersion = t.String()
		}

		return &searchbackend.RepoIndexOptions{
			Name:       string(repo.Name),
			RepoID:     int32(repo.ID),
			Public:     !repo.Private,
			Priority:   priority,
			Fork:       repo.Fork,
			Archived:   repo.Archived,
			GetVersion: getVersion,

			DocumentRanksVersion: documentRanksVersion,
		}, nil
	}

	revisionsForRepo, revisionsForRepoErr := h.SearchContextsRepoRevs(ctx, parameters.repoIDs)
	getSearchContextRevisions := func(repoID int32) ([]string, error) {
		if revisionsForRepoErr != nil {
			return nil, revisionsForRepoErr
		}
		return revisionsForRepo[api.RepoID(repoID)], nil
	}

	// searchbackend uses int32 instead of api.RepoID currently, so build
	// up a slice of that.
	repoIDs := make([]int32, len(parameters.repoIDs))
	for i := range parameters.repoIDs {
		repoIDs[i] = int32(parameters.repoIDs[i])
	}

	indexOptions := searchbackend.GetIndexOptions(
		&siteConfig,
		getRepoIndexOptions,
		getSearchContextRevisions,
		repoIDs...,
	)

	return &searchConfigurationResponse{
		options:     indexOptions,
		fingerprint: nextFingerPrint,
	}, nil
}

type parameterError struct {
	err string
}

func (e *parameterError) Error() string { return e.err }

type searchConfigurationParameters struct {
	repoIDs     []api.RepoID
	fingerprint searchbackend.ConfigFingerprint
}

type searchConfigurationResponse struct {
	options     []searchbackend.ZoektIndexOptions
	fingerprint searchbackend.ConfigFingerprint
}

// serveList is used by zoekt to get the list of repositories for it to index.
func (h *searchIndexerServer) serveList(w http.ResponseWriter, r *http.Request) error {
	var opt struct {
		// Hostname is used to determine the subset of repos to return
		Hostname string
		// IndexedIDs are the repository IDs of indexed repos by Hostname.
		IndexedIDs []api.RepoID
	}

	err := json.NewDecoder(r.Body).Decode(&opt)
	if err != nil {
		return err
	}

	indexable, err := h.ListIndexable(r.Context())
	if err != nil {
		return err
	}

	if h.Indexers.Enabled() {
		indexed := make(map[uint32]*zoekt.MinimalRepoListEntry, len(opt.IndexedIDs))
		add := func(r *types.MinimalRepo) { indexed[uint32(r.ID)] = nil }
		if len(opt.IndexedIDs) > 0 {
			opts := database.ReposListOptions{IDs: opt.IndexedIDs}
			err = h.RepoStore.StreamMinimalRepos(r.Context(), opts, add)
			if err != nil {
				return err
			}
		}

		indexable, err = h.Indexers.ReposSubset(r.Context(), opt.Hostname, indexed, indexable)
		if err != nil {
			return err
		}
	}

	// TODO: Avoid batching up so much in memory by:
	// 1. Changing the schema from object of arrays to array of objects.
	// 2. Stream out each object marshalled rather than marshall the full list in memory.

	ids := make([]api.RepoID, 0, len(indexable))

	for _, r := range indexable {
		ids = append(ids, r.ID)
	}

	data := struct {
		RepoIDs []api.RepoID
	}{
		RepoIDs: ids,
	}

	return json.NewEncoder(w).Encode(&data)
}

var metricGetVersion = promauto.NewCounter(prometheus.CounterOpts{
	Name: "src_search_get_version_total",
	Help: "The total number of times we poll gitserver for the version of a indexable branch.",
})

func (h *searchIndexerServer) serveRepoRank(w http.ResponseWriter, r *http.Request) error {
	return serveRank(h.Ranking.GetRepoRank, w, r)
}

func (h *searchIndexerServer) serveDocumentRanks(w http.ResponseWriter, r *http.Request) error {
	return serveRank(h.Ranking.GetDocumentRanks, w, r)
}

func serveRank[T []float64 | map[string][]float64](
	f func(ctx context.Context, name api.RepoName) (r T, err error),
	w http.ResponseWriter,
	r *http.Request,
) error {
	ctx := r.Context()

	repoName := api.RepoName(mux.Vars(r)["RepoName"])

	rank, err := f(ctx, repoName)
	if err != nil {
		if errcode.IsNotFound(err) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return nil
		}
		return err
	}

	b, err := json.Marshal(rank)
	if err != nil {
		return err
	}

	_, _ = w.Write(b)
	return nil
}

func (h *searchIndexerServer) handleIndexStatusUpdate(_ http.ResponseWriter, r *http.Request) error {
	var args indexStatusUpdateArgs

	if err := json.NewDecoder(r.Body).Decode(&args); err != nil {
		return errors.Wrap(err, "failed to decode request args")
	}

	return h.doIndexStatusUpdate(r.Context(), &args)
}

func (h *searchIndexerServer) doIndexStatusUpdate(ctx context.Context, args *indexStatusUpdateArgs) error {
	var (
		ids     = make([]int32, len(args.Repositories))
		minimal = make(map[uint32]*zoekt.MinimalRepoListEntry, len(args.Repositories))
	)

	for i, repo := range args.Repositories {
		ids[i] = int32(repo.RepoID)
		minimal[repo.RepoID] = &zoekt.MinimalRepoListEntry{Branches: repo.Branches}
	}

	h.logger.Info("updating index status", log.Int32s("repositories", ids))
	return h.db.ZoektRepos().UpdateIndexStatuses(ctx, minimal)
}

type indexStatusUpdateArgs struct {
	Repositories []indexStatusUpdateRepository
}

type indexStatusUpdateRepository struct {
	RepoID   uint32
	Branches []zoekt.RepositoryBranch
}

func (a *indexStatusUpdateArgs) FromProto(req *proto.UpdateIndexStatusRequest) {
	a.Repositories = make([]indexStatusUpdateRepository, 0, len(req.Repositories))

	for _, repo := range req.Repositories {
		branches := make([]zoekt.RepositoryBranch, 0, len(repo.Branches))
		for _, b := range repo.Branches {
			branches = append(branches, zoekt.RepositoryBranch{
				Name:    b.Name,
				Version: b.Version,
			})
		}

		a.Repositories = append(a.Repositories, struct {
			RepoID   uint32
			Branches []zoekt.RepositoryBranch
		}{
			RepoID:   repo.RepoId,
			Branches: branches,
		})
	}
}

func (a *indexStatusUpdateArgs) ToProto() *proto.UpdateIndexStatusRequest {
	repos := make([]*proto.UpdateIndexStatusRequest_Repository, 0, len(a.Repositories))

	for _, repo := range a.Repositories {
		branches := make([]*proto.ZoektRepositoryBranch, 0, len(repo.Branches))
		for _, b := range repo.Branches {
			branches = append(branches, &proto.ZoektRepositoryBranch{
				Name:    b.Name,
				Version: b.Version,
			})
		}

		repos = append(repos, &proto.UpdateIndexStatusRequest_Repository{
			RepoId:   repo.RepoID,
			Branches: branches,
		})
	}

	return &proto.UpdateIndexStatusRequest{
		Repositories: repos,
	}
}

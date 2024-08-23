package main

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"hash/crc32"
	"log"
	"math/rand"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	proto "github.com/sourcegraph/zoekt/cmd/zoekt-sourcegraph-indexserver/protos/sourcegraph/zoekt/configuration/v1"
	"github.com/sourcegraph/zoekt/ctags"
	"golang.org/x/net/trace"

	"github.com/sourcegraph/zoekt"
)

// SourcegraphListResult is the return value of Sourcegraph.List. It is its
// own type since internally we batch the calculation of index options. This
// is exposed via IterateIndexOptions.
//
// This type has state and is coupled to the Sourcegraph implementation.
type SourcegraphListResult struct {
	// IDs is the set of Sourcegraph repository IDs that this replica needs
	// to index.
	IDs []uint32

	// IterateIndexOptions best effort resolves the IndexOptions for RepoIDs. If
	// any repository fails it internally logs. It uses the "config fingerprint"
	// to reduce the amount of work done. This means we only resolve options for
	// repositories which have been mutated since the last Sourcegraph.List
	// call.
	//
	// Note: this has a side-effect of setting a the "config fingerprint". The
	// config fingerprint means we only calculate index options for repositories
	// that have changed since the last call to IterateIndexOptions. If you want
	// to force calculation of index options use
	// Sourcegraph.ForceIterateIndexOptions.
	//
	// Note: This should not be called concurrently with the Sourcegraph client.
	IterateIndexOptions func(func(IndexOptions))
}

// Sourcegraph represents the Sourcegraph service. It informs the indexserver
// what to index and which options to use.
type Sourcegraph interface {
	// List returns a list of repository IDs to index as well as a facility to
	// fetch the indexing options.
	//
	// Note: The return value is not safe to use concurrently with future calls
	// to List.
	List(ctx context.Context, indexed []uint32) (*SourcegraphListResult, error)

	// ForceIterateIndexOptions will best-effort calculate the index options for
	// all repos. For each repo it will call either onSuccess or onError. This
	// is the forced version of IterateIndexOptions, so will always calculate
	// options for each id in repos.
	ForceIterateIndexOptions(onSuccess func(IndexOptions), onError func(uint32, error), repos ...uint32)

	// GetDocumentRanks returns a map from paths within the given repo to their
	// rank vectors. Paths are assumed to be ordered by each pairwise component of
	// the resulting vector, higher ranks coming earlier
	GetDocumentRanks(ctx context.Context, repoName string) (RepoPathRanks, error)

	// UpdateIndexStatus sends a request to Sourcegraph to confirm that the
	// given repositories have been indexed.
	UpdateIndexStatus(repositories []indexStatus) error
}

type SourcegraphClientOption func(*sourcegraphClient)

// WithBatchSize controls how many repository configurations we request a time.
// If BatchSize is 0, we default to requesting 10,000 repositories at once.
func WithBatchSize(batchSize int) SourcegraphClientOption {
	return func(c *sourcegraphClient) {
		c.BatchSize = batchSize
	}
}

func newSourcegraphClient(rootURL *url.URL, hostname string, grpcClient proto.ZoektConfigurationServiceClient, opts ...SourcegraphClientOption) *sourcegraphClient {
	client := &sourcegraphClient{
		Root:       rootURL,
		Hostname:   hostname,
		BatchSize:  0,
		grpcClient: grpcClient,
	}

	for _, opt := range opts {
		opt(client)
	}

	return client
}

// sourcegraphClient contains methods which interact with the sourcegraph API.
type sourcegraphClient struct {
	// Root is the base URL for the Sourcegraph instance to index. Normally
	// http://sourcegraph-frontend-internal or http://localhost:3090.
	Root *url.URL

	// Hostname is the name we advertise to Sourcegraph when asking for the
	// list of repositories to index.
	Hostname string

	// BatchSize is how many repository configurations we request at once. If
	// zero a value of 10000 is used.
	BatchSize int

	// grpcClient is used to make requests to the Sourcegraph instance if gRPC is enabled.
	grpcClient proto.ZoektConfigurationServiceClient

	// configFingerprintProto is the last config fingerprint (as GRPC) returned from
	// Sourcegraph. It can be used for future calls to the configuration
	// endpoint.
	//
	// configFingerprintProto is mutually exclusive with configFingerprint - this field
	// will only be used if gRPC is enabled.
	configFingerprintProto *proto.Fingerprint

	// configFingerprintReset tracks when we should zero out the
	// configFingerprint. We want to periodically do this just in case our
	// configFingerprint logic is faulty. When it is cleared out, we fallback to
	// calculating everything.
	configFingerprintReset time.Time
}

// GetDocumentRanks asks Sourcegraph for a mapping of file paths to rank
// vectors.
func (s *sourcegraphClient) GetDocumentRanks(ctx context.Context, repoName string) (RepoPathRanks, error) {
	resp, err := s.grpcClient.DocumentRanks(ctx, &proto.DocumentRanksRequest{Repository: repoName})
	if err != nil {
		return RepoPathRanks{}, err
	}

	var out RepoPathRanks
	out.FromProto(resp)

	return out, nil
}

func (s *sourcegraphClient) List(ctx context.Context, indexed []uint32) (*SourcegraphListResult, error) {
	repos, err := s.listRepoIDs(ctx, indexed)
	if err != nil {
		return nil, fmt.Errorf("listRepoIDs: %w", err)
	}

	batchSize := s.BatchSize
	if batchSize == 0 {
		batchSize = 10_000
	}

	// Check if we should recalculate everything.
	if time.Now().After(s.configFingerprintReset) {
		// for every 500 repos we wait a minute. 2021-12-15 on sourcegraph.com
		// this works out to every 100 minutes.
		next := time.Duration(len(indexed)) * time.Minute / 500
		if min := 5 * time.Minute; next < min {
			next = min
		}
		next += time.Duration(rand.Int63n(int64(next) / 4)) // jitter
		s.configFingerprintReset = time.Now().Add(next)

		s.configFingerprintProto = nil
	}

	// getIndexOptionsFunc is a function that can be used to get the index
	// options for a set of repos (while properly handling any configuration fingerprint
	// changes).
	//
	// In general, this function provides a consistent fingerprint for each batch call,
	// and updates the server state with the new fingerprint. If any of the batch calls
	// fail, the old fingerprint is restored.
	type getIndexOptionsFunc func(repos ...uint32) ([]indexOptionsItem, error)

	mkGetIndexOptionsFunc := func(tr trace.Trace) getIndexOptionsFunc {
		startingFingerPrint := s.configFingerprintProto
		tr.LazyPrintf("fingerprint: %s", startingFingerPrint.String())

		first := true
		return func(repos ...uint32) ([]indexOptionsItem, error) {
			options, nextFingerPrint, err := s.getIndexOptions(ctx, startingFingerPrint, repos)
			if err != nil {
				first = false
				s.configFingerprintProto = startingFingerPrint

				return nil, err
			}

			if first {
				first = false
				s.configFingerprintProto = nextFingerPrint
				tr.LazyPrintf("new fingerprint: %s", nextFingerPrint.String())
			}

			return options, nil
		}
	}

	iterate := func(f func(IndexOptions)) {
		start := time.Now()
		tr := trace.New("getIndexOptions", "")
		tr.LazyPrintf("getting index options for %d repos", len(repos))

		defer func() {
			metricResolveRevisionsDuration.Observe(time.Since(start).Seconds())
			tr.Finish()
		}()

		getIndexOptions := mkGetIndexOptionsFunc(tr)

		// We ask the frontend to get index options in batches.
		for repos := range batched(repos, batchSize) {
			start := time.Now()
			options, err := getIndexOptions(repos...)
			duration := time.Since(start)

			if err != nil {
				metricResolveRevisionDuration.WithLabelValues("false").Observe(duration.Seconds())
				tr.LazyPrintf("failed fetching options batch: %v", err)
				tr.SetError()

				continue
			}

			metricResolveRevisionDuration.WithLabelValues("true").Observe(duration.Seconds())

			for _, o := range options {
				metricGetIndexOptions.Inc()

				if o.Error != "" {
					metricGetIndexOptionsError.Inc()
					tr.LazyPrintf("failed fetching options for %v: %v", o.Name, o.Error)
					tr.SetError()

					continue
				}
				f(o.IndexOptions)
			}
		}
	}

	return &SourcegraphListResult{
		IDs:                 repos,
		IterateIndexOptions: iterate,
	}, nil
}

func (s *sourcegraphClient) ForceIterateIndexOptions(onSuccess func(IndexOptions), onError func(uint32, error), repos ...uint32) {
	batchSize := s.BatchSize
	if batchSize == 0 {
		batchSize = 10_000
	}

	for repos := range batched(repos, batchSize) {
		opts, _, err := s.getIndexOptions(context.Background(), nil, repos)
		if err != nil {
			for _, id := range repos {
				onError(id, err)
			}
			continue
		}
		for _, o := range opts {
			if o.RepoID > 0 && o.Error != "" {
				onError(o.RepoID, errors.New(o.Error))
			}
			if o.Error == "" {
				onSuccess(o.IndexOptions)
			}
		}
	}
}

// indexOptionsItem wraps IndexOptions to also include an error returned by
// the API.
type indexOptionsItem struct {
	IndexOptions
	Error string
}

func (o *indexOptionsItem) FromProto(x *proto.ZoektIndexOptions) {
	branches := make([]zoekt.RepositoryBranch, 0, len(x.Branches))
	for _, b := range x.GetBranches() {
		branches = append(branches, zoekt.RepositoryBranch{
			Name:    b.GetName(),
			Version: b.GetVersion(),
		})
	}

	item := indexOptionsItem{}
	languageMap := make(map[string]ctags.CTagsParserType)

	for _, lang := range x.GetLanguageMap() {
		languageMap[lang.GetLanguage()] = ctags.CTagsParserType(lang.GetCtags().Number())
	}

	item.IndexOptions = IndexOptions{
		RepoID:     uint32(x.GetRepoId()),
		LargeFiles: x.GetLargeFiles(),
		Symbols:    x.GetSymbols(),
		Branches:   branches,
		Name:       x.GetName(),

		Priority: x.GetPriority(),

		DocumentRanksVersion: x.GetDocumentRanksVersion(),

		Public:   x.GetPublic(),
		Fork:     x.GetFork(),
		Archived: x.GetArchived(),

		LanguageMap:      languageMap,
		ShardConcurrency: x.GetShardConcurrency(),
	}

	item.Error = x.GetError()

	*o = item
}

func (o *indexOptionsItem) ToProto() *proto.ZoektIndexOptions {
	branches := make([]*proto.ZoektRepositoryBranch, 0, len(o.Branches))
	for _, b := range o.Branches {
		branches = append(branches, &proto.ZoektRepositoryBranch{
			Name:    b.Name,
			Version: b.Version,
		})
	}

	languageMap := make([]*proto.LanguageMapping, 0, len(o.LanguageMap))

	for lang, parser := range o.LanguageMap {
		languageMap = append(languageMap, &proto.LanguageMapping{
			Language: lang,
			Ctags:    proto.CTagsParserType(parser),
		})
	}

	return &proto.ZoektIndexOptions{
		RepoId:     int32(o.RepoID),
		LargeFiles: o.LargeFiles,
		Symbols:    o.Symbols,
		Branches:   branches,
		Name:       o.Name,

		Priority: o.Priority,

		DocumentRanksVersion: o.DocumentRanksVersion,

		Public:   o.Public,
		Fork:     o.Fork,
		Archived: o.Archived,

		Error: o.Error,

		LanguageMap:      languageMap,
		ShardConcurrency: o.ShardConcurrency,
	}
}

func (s *sourcegraphClient) getIndexOptions(ctx context.Context, fingerprint *proto.Fingerprint, repos []uint32) ([]indexOptionsItem, *proto.Fingerprint, error) {
	repoIDs := make([]int32, 0, len(repos))
	for _, id := range repos {
		repoIDs = append(repoIDs, int32(id))
	}

	req := proto.SearchConfigurationRequest{
		RepoIds:     repoIDs,
		Fingerprint: fingerprint,
	}

	response, err := s.grpcClient.SearchConfiguration(ctx, &req)
	if err != nil {
		return nil, nil, err
	}

	protoItems := response.GetUpdatedOptions()
	items := make([]indexOptionsItem, 0, len(protoItems))
	for _, x := range protoItems {
		var item indexOptionsItem
		item.FromProto(x)
		item.IndexOptions.CloneURL = s.getCloneURL(item.Name)

		items = append(items, item)
	}

	return items, response.GetFingerprint(), nil
}

func (s *sourcegraphClient) getCloneURL(name string) string {
	return s.Root.ResolveReference(&url.URL{Path: path.Join("/.internal/git", name)}).String()
}

func (s *sourcegraphClient) listRepoIDs(ctx context.Context, indexed []uint32) ([]uint32, error) {
	var request proto.ListRequest
	request.Hostname = s.Hostname
	request.IndexedIds = make([]int32, 0, len(indexed))
	for _, id := range indexed {
		request.IndexedIds = append(request.IndexedIds, int32(id))
	}

	response, err := s.grpcClient.List(ctx, &request)
	if err != nil {
		return nil, err
	}

	repoIDs := make([]uint32, 0, len(response.RepoIds))
	for _, id := range response.RepoIds {
		repoIDs = append(repoIDs, uint32(id))
	}

	return repoIDs, nil
}

type indexStatus struct {
	RepoID        uint32
	Branches      []zoekt.RepositoryBranch
	IndexTimeUnix int64
}

type updateIndexStatusRequest struct {
	Repositories []indexStatus
}

func (u *updateIndexStatusRequest) ToProto() *proto.UpdateIndexStatusRequest {
	repositories := make([]*proto.UpdateIndexStatusRequest_Repository, 0, len(u.Repositories))

	for _, repo := range u.Repositories {
		branches := make([]*proto.ZoektRepositoryBranch, 0, len(repo.Branches))

		for _, branch := range repo.Branches {
			branches = append(branches, &proto.ZoektRepositoryBranch{
				Name:    branch.Name,
				Version: branch.Version,
			})
		}

		repositories = append(repositories, &proto.UpdateIndexStatusRequest_Repository{
			RepoId:        repo.RepoID,
			Branches:      branches,
			IndexTimeUnix: repo.IndexTimeUnix,
		})
	}

	return &proto.UpdateIndexStatusRequest{
		Repositories: repositories,
	}
}

func (u *updateIndexStatusRequest) FromProto(x *proto.UpdateIndexStatusRequest) {
	protoRepositories := x.GetRepositories()
	repositories := make([]indexStatus, 0, len(protoRepositories))

	for _, repo := range x.GetRepositories() {
		protoBranches := repo.GetBranches()
		branches := make([]zoekt.RepositoryBranch, 0, len(protoBranches))

		for _, branch := range repo.GetBranches() {
			branches = append(branches, zoekt.RepositoryBranch{
				Name:    branch.GetName(),
				Version: branch.GetVersion(),
			})
		}

		repositories = append(repositories, indexStatus{
			RepoID:        repo.GetRepoId(),
			Branches:      branches,
			IndexTimeUnix: repo.GetIndexTimeUnix(),
		})
	}

	*u = updateIndexStatusRequest{
		Repositories: repositories,
	}
}

// UpdateIndexStatus sends a request to Sourcegraph to confirm that the given
// repositories have been indexed.
func (s *sourcegraphClient) UpdateIndexStatus(repositories []indexStatus) error {
	r := updateIndexStatusRequest{Repositories: repositories}

	request := r.ToProto()
	_, err := s.grpcClient.UpdateIndexStatus(context.Background(), request)
	if err != nil {
		return fmt.Errorf("failed to update index status: %w", err)
	}

	return nil
}

type sourcegraphFake struct {
	RootDir string
	Log     *log.Logger
}

// GetDocumentRanks expects a file where each line has the following format:
// path<tab>rank... where rank is a float64.
func (sf sourcegraphFake) GetDocumentRanks(ctx context.Context, repoName string) (RepoPathRanks, error) {
	dir := filepath.Join(sf.RootDir, filepath.FromSlash(repoName))

	fd, err := os.Open(filepath.Join(dir, "SG_DOCUMENT_RANKS"))
	if err != nil {
		return RepoPathRanks{}, err
	}

	ranks := RepoPathRanks{}

	sum := 0.0
	count := 0
	scanner := bufio.NewScanner(fd)
	for scanner.Scan() {
		s := scanner.Text()
		pathRanks := strings.Split(s, "\t")
		if rank, err := strconv.ParseFloat(pathRanks[1], 64); err == nil {
			ranks.Paths[pathRanks[0]] = rank
			sum += rank
			count++
		}
	}

	if err := scanner.Err(); err != nil {
		return RepoPathRanks{}, err
	}

	ranks.MeanRank = sum / float64(count)
	return ranks, nil
}

func (sf sourcegraphFake) List(ctx context.Context, indexed []uint32) (*SourcegraphListResult, error) {
	repos, err := sf.ListRepoIDs(ctx, indexed)
	if err != nil {
		return nil, err
	}

	iterate := func(f func(IndexOptions)) {
		opts, err := sf.GetIndexOptions(repos...)
		if err != nil {
			sf.Log.Printf("WARN: ignoring GetIndexOptions error: %v", err)
		}
		for _, opt := range opts {
			if opt.Error != "" {
				sf.Log.Printf("WARN: ignoring GetIndexOptions error for %s: %v", opt.Name, opt.Error)
				continue
			}
			f(opt.IndexOptions)
		}
	}

	return &SourcegraphListResult{
		IDs:                 repos,
		IterateIndexOptions: iterate,
	}, nil
}

func (sf sourcegraphFake) ForceIterateIndexOptions(onSuccess func(IndexOptions), onError func(uint32, error), repos ...uint32) {
	opts, err := sf.GetIndexOptions(repos...)
	if err != nil {
		for _, id := range repos {
			onError(id, err)
		}
		return
	}
	for _, o := range opts {
		if o.RepoID > 0 && o.Error != "" {
			onError(o.RepoID, errors.New(o.Error))
		}
		if o.Error == "" {
			onSuccess(o.IndexOptions)
		}
	}
}

func (sf sourcegraphFake) GetIndexOptions(repos ...uint32) ([]indexOptionsItem, error) {
	reposIdx := map[uint32]int{}
	for i, id := range repos {
		reposIdx[id] = i
	}

	items := make([]indexOptionsItem, len(repos))
	err := sf.visitRepos(func(name string) {
		idx, ok := reposIdx[sf.id(name)]
		if !ok {
			return
		}
		opts, err := sf.getIndexOptions(name)
		if err != nil {
			items[idx] = indexOptionsItem{Error: err.Error()}
		} else {
			items[idx] = indexOptionsItem{IndexOptions: opts}
		}
	})
	if err != nil {
		return nil, err
	}

	for i := range items {
		if items[i].Error == "" && items[i].RepoID == 0 {
			items[i].Error = "not found"
		}
	}

	return items, nil
}

func (sf sourcegraphFake) getIndexOptions(name string) (IndexOptions, error) {
	dir := filepath.Join(sf.RootDir, filepath.FromSlash(name))
	exists := func(p string) bool {
		_, err := os.Stat(filepath.Join(dir, p))
		return err == nil
	}
	float := func(p string) float64 {
		b, _ := os.ReadFile(filepath.Join(dir, p))
		f, _ := strconv.ParseFloat(string(bytes.TrimSpace(b)), 64)
		return f
	}

	opts := IndexOptions{
		RepoID:   sf.id(name),
		Name:     name,
		CloneURL: sf.getCloneURL(name),
		Symbols:  true,

		Public:   !exists("SG_PRIVATE"),
		Fork:     exists("SG_FORK"),
		Archived: exists("SG_ARCHIVED"),

		Priority: float("SG_PRIORITY"),
	}

	if stat, err := os.Stat(filepath.Join(dir, "SG_DOCUMENT_RANKS")); err == nil {
		opts.DocumentRanksVersion = stat.ModTime().String()
	}

	branches, err := sf.getBranches(name)
	if err != nil {
		return opts, err
	}
	opts.Branches = branches

	return opts, nil
}

func (sf sourcegraphFake) getBranches(name string) ([]zoekt.RepositoryBranch, error) {
	dir := filepath.Join(sf.RootDir, filepath.FromSlash(name))
	repo, err := git.PlainOpen(dir)
	if err != nil {
		return nil, err
	}

	cfg, err := repo.Config()
	if err != nil {
		return nil, err
	}

	sec := cfg.Raw.Section("zoekt")
	branches := sec.Options.GetAll("branch")
	if len(branches) == 0 {
		branches = append(branches, "HEAD")
	}

	rBranches := make([]zoekt.RepositoryBranch, 0, len(branches))
	for _, branch := range branches {
		cmd := exec.Command("git", "rev-parse", branch)
		cmd.Dir = dir
		if b, err := cmd.Output(); err != nil {
			sf.Log.Printf("WARN: Could not get branch %s/%s", name, branch)
		} else {
			version := string(bytes.TrimSpace(b))
			rBranches = append(rBranches, zoekt.RepositoryBranch{
				Name:    branch,
				Version: version,
			})
		}
	}

	if len(rBranches) == 0 {
		return nil, fmt.Errorf("WARN: Could not get any branch revisions for repo %s", name)
	}

	return rBranches, nil
}

func (sf sourcegraphFake) id(name string) uint32 {
	// allow overriding the ID.
	idPath := filepath.Join(sf.RootDir, filepath.FromSlash(name), "SG_ID")
	if b, _ := os.ReadFile(idPath); len(b) > 0 {
		id, err := strconv.Atoi(strings.TrimSpace(string(b)))
		if err == nil {
			return uint32(id)
		}
	}
	return fakeID(name)
}

func (sf sourcegraphFake) getCloneURL(name string) string {
	return filepath.Join(sf.RootDir, filepath.FromSlash(name))
}

func (sf sourcegraphFake) ListRepoIDs(ctx context.Context, indexed []uint32) ([]uint32, error) {
	var repos []uint32
	err := sf.visitRepos(func(name string) {
		repos = append(repos, sf.id(name))
	})
	return repos, err
}

func (sf sourcegraphFake) visitRepos(visit func(name string)) error {
	return filepath.Walk(sf.RootDir, func(path string, fi os.FileInfo, fileErr error) error {
		if fileErr != nil {
			sf.Log.Printf("WARN: ignoring error searching %s: %v", path, fileErr)
			return nil
		}
		if !fi.IsDir() {
			return nil
		}

		gitdir := filepath.Join(path, ".git")
		if fi, err := os.Stat(gitdir); err != nil || !fi.IsDir() {
			return nil
		}

		subpath, err := filepath.Rel(sf.RootDir, path)
		if err != nil {
			// According to WalkFunc docs, path is always filepath.Join(root,
			// subpath). So Rel should always work.
			return fmt.Errorf("filepath.Walk returned %s which is not relative to %s: %w", path, sf.RootDir, err)
		}

		name := filepath.ToSlash(subpath)
		visit(name)

		return filepath.SkipDir
	})
}

func (s sourcegraphFake) UpdateIndexStatus(repositories []indexStatus) error {
	// noop
	return nil
}

// fakeID returns a deterministic ID based on name. Used for fakes and tests.
func fakeID(name string) uint32 {
	// magic at the end is to ensure we get a positive number when casting.
	return uint32(crc32.ChecksumIEEE([]byte(name))%(1<<31-1) + 1)
}

type sourcegraphNop struct{}

func (s sourcegraphNop) List(ctx context.Context, indexed []uint32) (*SourcegraphListResult, error) {
	return nil, nil
}

func (s sourcegraphNop) ForceIterateIndexOptions(onSuccess func(IndexOptions), onError func(uint32, error), repos ...uint32) {
}

func (s sourcegraphNop) GetDocumentRanks(ctx context.Context, repoName string) (RepoPathRanks, error) {
	return RepoPathRanks{}, nil
}

func (s sourcegraphNop) UpdateIndexStatus(repositories []indexStatus) error {
	return nil
}

type RepoPathRanks struct {
	MeanRank float64            `json:"mean_reference_count"`
	Paths    map[string]float64 `json:"paths"`
}

func (r *RepoPathRanks) FromProto(x *proto.DocumentRanksResponse) {
	protoPaths := x.GetPaths()
	ranks := make(map[string]float64, len(protoPaths))
	for filePath, rank := range protoPaths {
		ranks[filePath] = rank
	}

	*r = RepoPathRanks{
		MeanRank: x.GetMeanRank(),
		Paths:    ranks,
	}
}

func (r *RepoPathRanks) ToProto() *proto.DocumentRanksResponse {
	paths := make(map[string]float64, len(r.Paths))
	for filePath, rank := range r.Paths {
		paths[filePath] = rank
	}

	return &proto.DocumentRanksResponse{
		MeanRank: r.MeanRank,
		Paths:    paths,
	}
}

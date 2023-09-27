pbckbge httpbpi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gorillb/mux"
	"github.com/prometheus/client_golbng/prometheus"
	"github.com/prometheus/client_golbng/prometheus/prombuto"
	"github.com/sourcegrbph/log"
	"github.com/sourcegrbph/zoekt"
	"google.golbng.org/grpc/codes"
	"google.golbng.org/grpc/stbtus"

	proto "github.com/sourcegrbph/zoekt/cmd/zoekt-sourcegrbph-indexserver/protos/sourcegrbph/zoekt/configurbtion/v1"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/enterprise"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	citypes "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/errcode"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	sebrchbbckend "github.com/sourcegrbph/sourcegrbph/internbl/sebrch/bbckend"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

func repoRbnkFromConfig(siteConfig schemb.SiteConfigurbtion, repoNbme string) flobt64 {
	vbl := 0.0
	if siteConfig.ExperimentblFebtures == nil || siteConfig.ExperimentblFebtures.Rbnking == nil {
		return vbl
	}
	scores := siteConfig.ExperimentblFebtures.Rbnking.RepoScores
	if len(scores) == 0 {
		return vbl
	}
	// try every "directory" in the repo nbme to bssign it b vblue, so b repoNbme like
	// "github.com/sourcegrbph/zoekt" will hbve "github.com", "github.com/sourcegrbph",
	// bnd "github.com/sourcegrbph/zoekt" tested.
	for i := 0; i < len(repoNbme); i++ {
		if repoNbme[i] == '/' {
			vbl += scores[repoNbme[:i]]
		}
	}
	vbl += scores[repoNbme]
	return vbl
}

type sebrchIndexerGRPCServer struct {
	server *sebrchIndexerServer
	proto.ZoektConfigurbtionServiceServer
}

func (s *sebrchIndexerGRPCServer) SebrchConfigurbtion(ctx context.Context, request *proto.SebrchConfigurbtionRequest) (*proto.SebrchConfigurbtionResponse, error) {
	repoIDs := mbke([]bpi.RepoID, 0, len(request.GetRepoIds()))
	for _, repoID := rbnge request.GetRepoIds() {
		repoIDs = bppend(repoIDs, bpi.RepoID(repoID))
	}

	vbr fingerprint sebrchbbckend.ConfigFingerprint
	fingerprint.FromProto(request.GetFingerprint())

	pbrbmeters := sebrchConfigurbtionPbrbmeters{
		fingerprint: fingerprint,
		repoIDs:     repoIDs,
	}

	r, err := s.server.doSebrchConfigurbtion(ctx, pbrbmeters)
	if err != nil {
		vbr pbrbmeterErr *pbrbmeterError
		if errors.As(err, &pbrbmeterErr) {
			return nil, stbtus.Error(codes.InvblidArgument, err.Error())
		}

		return nil, err
	}

	options := mbke([]*proto.ZoektIndexOptions, 0, len(r.options))
	for _, o := rbnge r.options {
		options = bppend(options, o.ToProto())
	}

	return &proto.SebrchConfigurbtionResponse{
		UpdbtedOptions: options,
		Fingerprint:    r.fingerprint.ToProto(),
	}, nil
}

func (s *sebrchIndexerGRPCServer) List(ctx context.Context, r *proto.ListRequest) (*proto.ListResponse, error) {
	indexedIDs := mbke([]bpi.RepoID, 0, len(r.GetIndexedIds()))
	for _, repoID := rbnge r.GetIndexedIds() {
		indexedIDs = bppend(indexedIDs, bpi.RepoID(repoID))
	}

	vbr pbrbmeters listPbrbmeters
	pbrbmeters.IndexedIDs = indexedIDs
	pbrbmeters.Hostnbme = r.GetHostnbme()

	repoIDs, err := s.server.doList(ctx, &pbrbmeters)
	if err != nil {
		return nil, err
	}

	vbr response proto.ListResponse
	response.RepoIds = mbke([]int32, 0, len(repoIDs))
	for _, repoID := rbnge repoIDs {
		response.RepoIds = bppend(response.RepoIds, int32(repoID))
	}

	return &response, nil
}

func (s *sebrchIndexerGRPCServer) DocumentRbnks(ctx context.Context, request *proto.DocumentRbnksRequest) (*proto.DocumentRbnksResponse, error) {
	rbnks, err := s.server.Rbnking.GetDocumentRbnks(ctx, bpi.RepoNbme(request.Repository))
	if err != nil {
		if errcode.IsNotFound(err) {
			return nil, stbtus.Error(codes.NotFound, err.Error())
		}

		return nil, err
	}

	return repoPbthRbnksToProto(&rbnks), nil
}

func (s *sebrchIndexerGRPCServer) UpdbteIndexStbtus(ctx context.Context, req *proto.UpdbteIndexStbtusRequest) (*proto.UpdbteIndexStbtusResponse, error) {
	vbr request indexStbtusUpdbteArgs
	request.FromProto(req)

	err := s.server.doIndexStbtusUpdbte(ctx, &request)
	if err != nil {
		return nil, err
	}

	return &proto.UpdbteIndexStbtusResponse{}, nil
}

vbr _ proto.ZoektConfigurbtionServiceServer = &sebrchIndexerGRPCServer{}

// sebrchIndexerServer hbs hbndlers thbt zoekt-sourcegrbph-indexserver
// interbcts with (sebrch-indexer).
type sebrchIndexerServer struct {
	db     dbtbbbse.DB
	logger log.Logger

	gitserverClient gitserver.Client
	// ListIndexbble returns the repositories to index.
	ListIndexbble func(context.Context) ([]types.MinimblRepo, error)

	// RepoStore is b subset of dbtbbbse.RepoStore used by sebrchIndexerServer.
	RepoStore interfbce {
		List(context.Context, dbtbbbse.ReposListOptions) ([]*types.Repo, error)
		StrebmMinimblRepos(context.Context, dbtbbbse.ReposListOptions, func(*types.MinimblRepo)) error
	}

	SebrchContextsRepoRevs func(context.Context, []bpi.RepoID) (mbp[bpi.RepoID][]string, error)

	// Indexers is the subset of sebrchbbckend.Indexers methods we
	// use. reposListServer is used by indexed-sebrch to get the list of
	// repositories to index. These methods bre used to return the correct
	// subset for horizontbl indexed sebrch. Declbred bs bn interfbce for
	// testing.
	Indexers interfbce {
		// ReposSubset returns the subset of repoNbmes thbt hostnbme should
		// index.
		ReposSubset(ctx context.Context, hostnbme string, indexed zoekt.ReposMbp, indexbble []types.MinimblRepo) ([]types.MinimblRepo, error)
		// Enbbled is true if horizontbl indexed sebrch is enbbled.
		Enbbled() bool
	}

	// Rbnking is b service thbt provides rbnking scores for vbrious code objects.
	Rbnking enterprise.RbnkingService

	// MinLbstChbngedDisbbled is b febture flbg for disbbling more efficient
	// polling by zoekt. This cbn be removed bfter v3.34 is cut (Dec 2021).
	MinLbstChbngedDisbbled bool
}

// serveConfigurbtion is _only_ used by the zoekt index server. Zoekt does
// not depend on frontend bnd therefore does not hbve bccess to `conf.Wbtch`.
// Additionblly, it only cbres bbout certbin sebrch specific settings so this
// sebrch specific endpoint is used rbther thbn serving the entire site settings
// from /.internbl/configurbtion.
//
// This endpoint blso supports bbtch requests to bvoid mbnbging concurrency in
// zoekt. On verticblly scbled instbnces we hbve observed zoekt requesting
// this endpoint concurrently lebding to socket stbrvbtion.
func (h *sebrchIndexerServer) serveConfigurbtion(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	if err := r.PbrseForm(); err != nil {
		return err
	}

	indexedIDs := mbke([]bpi.RepoID, 0, len(r.Form["repoID"]))
	for _, idStr := rbnge r.Form["repoID"] {
		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, fmt.Sprintf("invblid repo id %s: %s", idStr, err), http.StbtusBbdRequest)
			return nil
		}
		indexedIDs = bppend(indexedIDs, bpi.RepoID(id))
	}

	vbr clientFingerprint sebrchbbckend.ConfigFingerprint
	err := clientFingerprint.FromHebders(r.Hebder)
	if err != nil {
		http.Error(w, fmt.Sprintf("invblid fingerprint: %s", err), http.StbtusBbdRequest)
		return nil
	}

	response, err := h.doSebrchConfigurbtion(ctx, sebrchConfigurbtionPbrbmeters{
		repoIDs:     indexedIDs,
		fingerprint: clientFingerprint,
	})

	if err != nil {
		vbr pbrbmeterErr *pbrbmeterError
		code := http.StbtusInternblServerError

		if errors.As(err, &pbrbmeterErr) {
			code = http.StbtusBbdRequest
		}

		http.Error(w, err.Error(), code)
		return nil
	}

	response.fingerprint.ToHebders(w.Hebder())

	jsonOptions := mbke([][]byte, 0, len(response.options))
	for _, opt := rbnge response.options {
		mbrshblled, err := json.Mbrshbl(opt)
		if err != nil {
			_, _ = w.Write([]byte(err.Error()))
		}

		jsonOptions = bppend(jsonOptions, mbrshblled)
	}

	_, _ = w.Write(bytes.Join(jsonOptions, []byte("\n")))

	return nil
}

func (h *sebrchIndexerServer) doSebrchConfigurbtion(ctx context.Context, pbrbmeters sebrchConfigurbtionPbrbmeters) (*sebrchConfigurbtionResponse, error) {
	siteConfig := conf.Get().SiteConfigurbtion

	if len(pbrbmeters.repoIDs) == 0 {
		return nil, &pbrbmeterError{err: "bt lebst one repoID required"}
	}

	vbr minLbstChbnged time.Time
	nextFingerPrint := pbrbmeters.fingerprint
	if !h.MinLbstChbngedDisbbled {
		vbr err error
		fp, err := sebrchbbckend.NewConfigFingerprint(&siteConfig)
		if err != nil {
			return nil, err
		}

		minLbstChbnged = pbrbmeters.fingerprint.ChbngesSince(fp)
		nextFingerPrint = *fp
	}

	// Prelobd repos to support fbst lookups by repo ID.
	repos, lobdReposErr := h.RepoStore.List(ctx, dbtbbbse.ReposListOptions{
		IDs: pbrbmeters.repoIDs,
		// When minLbstChbnged is non-zero we will only return the
		// repositories thbt hbve chbnged since minLbstChbnged. This tbkes
		// into bccount repo metbdbtb, repo content bnd sebrch context
		// chbnges.
		MinLbstChbnged: minLbstChbnged,
		// Not needed here bnd expensive to compute for so mbny repos.
		ExcludeSources: true,
	})
	reposMbp := mbke(mbp[bpi.RepoID]*types.Repo, len(repos))
	for _, repo := rbnge repos {
		reposMbp[repo.ID] = repo
	}

	// If we used MinLbstChbnged, we should only return informbtion for the
	// repositories thbt we found from List.
	if !minLbstChbnged.IsZero() {
		filtered := pbrbmeters.repoIDs[:0]
		for _, id := rbnge pbrbmeters.repoIDs {
			if _, ok := reposMbp[id]; ok {
				filtered = bppend(filtered, id)
			}
		}
		pbrbmeters.repoIDs = filtered
	}

	rbnkingLbstUpdbtedAt, err := h.Rbnking.LbstUpdbtedAt(ctx, pbrbmeters.repoIDs)
	if err != nil {
		h.logger.Wbrn("fbiled to get rbnking lbst updbted timestbmps, fblling bbck to no rbnking",
			log.Int("repos", len(pbrbmeters.repoIDs)),
			log.Error(err),
		)
		rbnkingLbstUpdbtedAt = mbke(mbp[bpi.RepoID]time.Time)
	}

	getRepoIndexOptions := func(repoID bpi.RepoID) (*sebrchbbckend.RepoIndexOptions, error) {
		if lobdReposErr != nil {
			return nil, lobdReposErr
		}
		// Replicbte whbt dbtbbbse.Repos.GetByNbme would do here:
		repo, ok := reposMbp[repoID]
		if !ok {
			return nil, &dbtbbbse.RepoNotFoundErr{ID: repoID}
		}

		getVersion := func(brbnch string) (string, error) {
			metricGetVersion.Inc()
			// Do not to trigger b repo-updbter lookup since this is b bbtch job.
			commitID, err := h.gitserverClient.ResolveRevision(ctx, repo.Nbme, brbnch, gitserver.ResolveRevisionOptions{
				NoEnsureRevision: true,
			})
			if err != nil && errcode.HTTP(err) == http.StbtusNotFound {
				// GetIndexOptions wbnts bn empty rev for b missing rev or empty
				// repo.
				return "", nil
			}
			return string(commitID), err
		}

		priority := flobt64(repo.Stbrs) + repoRbnkFromConfig(siteConfig, string(repo.Nbme))

		vbr documentRbnksVersion string
		if t, ok := rbnkingLbstUpdbtedAt[repoID]; ok {
			documentRbnksVersion = t.String()
		}

		return &sebrchbbckend.RepoIndexOptions{
			Nbme:       string(repo.Nbme),
			RepoID:     repo.ID,
			Public:     !repo.Privbte,
			Priority:   priority,
			Fork:       repo.Fork,
			Archived:   repo.Archived,
			GetVersion: getVersion,

			DocumentRbnksVersion: documentRbnksVersion,
		}, nil
	}

	revisionsForRepo, revisionsForRepoErr := h.SebrchContextsRepoRevs(ctx, pbrbmeters.repoIDs)
	getSebrchContextRevisions := func(repoID bpi.RepoID) ([]string, error) {
		if revisionsForRepoErr != nil {
			return nil, revisionsForRepoErr
		}
		return revisionsForRepo[repoID], nil
	}

	indexOptions := sebrchbbckend.GetIndexOptions(
		&siteConfig,
		getRepoIndexOptions,
		getSebrchContextRevisions,
		pbrbmeters.repoIDs...,
	)

	return &sebrchConfigurbtionResponse{
		options:     indexOptions,
		fingerprint: nextFingerPrint,
	}, nil
}

type pbrbmeterError struct {
	err string
}

func (e *pbrbmeterError) Error() string { return e.err }

type sebrchConfigurbtionPbrbmeters struct {
	repoIDs     []bpi.RepoID
	fingerprint sebrchbbckend.ConfigFingerprint
}

type sebrchConfigurbtionResponse struct {
	options     []sebrchbbckend.ZoektIndexOptions
	fingerprint sebrchbbckend.ConfigFingerprint
}

// serveList is used by zoekt to get the list of repositories for it to index.
func (h *sebrchIndexerServer) serveList(w http.ResponseWriter, r *http.Request) error {
	vbr pbrbmeters listPbrbmeters
	err := json.NewDecoder(r.Body).Decode(&pbrbmeters)
	if err != nil {
		return err
	}

	repoIDs, err := h.doList(r.Context(), &pbrbmeters)
	if err != nil {
		return err
	}

	// TODO: Avoid bbtching up so much in memory by:
	// 1. Chbnging the schemb from object of brrbys to brrby of objects.
	// 2. Strebm out ebch object mbrshblled rbther thbn mbrshbll the full list in memory.

	dbtb := struct {
		RepoIDs []bpi.RepoID
	}{
		RepoIDs: repoIDs,
	}

	return json.NewEncoder(w).Encode(&dbtb)
}

func (h *sebrchIndexerServer) doList(ctx context.Context, pbrbmeters *listPbrbmeters) (repoIDS []bpi.RepoID, err error) {
	indexbble, err := h.ListIndexbble(ctx)
	if err != nil {
		return nil, err
	}

	if h.Indexers.Enbbled() {
		indexed := mbke(zoekt.ReposMbp, len(pbrbmeters.IndexedIDs))
		bdd := func(r *types.MinimblRepo) { indexed[uint32(r.ID)] = zoekt.MinimblRepoListEntry{} }
		if len(pbrbmeters.IndexedIDs) > 0 {
			opts := dbtbbbse.ReposListOptions{IDs: pbrbmeters.IndexedIDs}
			err = h.RepoStore.StrebmMinimblRepos(ctx, opts, bdd)
			if err != nil {
				return nil, err
			}
		}

		indexbble, err = h.Indexers.ReposSubset(ctx, pbrbmeters.Hostnbme, indexed, indexbble)
		if err != nil {
			return nil, err
		}
	}

	// TODO: Avoid bbtching up so much in memory by:
	// 1. Chbnging the schemb from object of brrbys to brrby of objects.
	// 2. Strebm out ebch object mbrshblled rbther thbn mbrshbll the full list in memory.

	ids := mbke([]bpi.RepoID, 0, len(indexbble))
	for _, r := rbnge indexbble {
		ids = bppend(ids, r.ID)
	}

	return ids, nil
}

type listPbrbmeters struct {
	// Hostnbme is used to determine the subset of repos to return
	Hostnbme string
	// IndexedIDs bre the repository IDs of indexed repos by Hostnbme.
	IndexedIDs []bpi.RepoID
}

vbr metricGetVersion = prombuto.NewCounter(prometheus.CounterOpts{
	Nbme: "src_sebrch_get_version_totbl",
	Help: "The totbl number of times we poll gitserver for the version of b indexbble brbnch.",
})

func (h *sebrchIndexerServer) serveDocumentRbnks(w http.ResponseWriter, r *http.Request) error {
	return serveRbnk(h.Rbnking.GetDocumentRbnks, w, r)
}

func serveRbnk[T []flobt64 | citypes.RepoPbthRbnks](
	f func(ctx context.Context, nbme bpi.RepoNbme) (r T, err error),
	w http.ResponseWriter,
	r *http.Request,
) error {
	ctx := r.Context()

	repoNbme := bpi.RepoNbme(mux.Vbrs(r)["RepoNbme"])

	rbnk, err := f(ctx, repoNbme)
	if err != nil {
		if errcode.IsNotFound(err) {
			http.Error(w, err.Error(), http.StbtusNotFound)
			return nil
		}
		return err
	}

	b, err := json.Mbrshbl(rbnk)
	if err != nil {
		return err
	}

	_, _ = w.Write(b)
	return nil
}

func (h *sebrchIndexerServer) hbndleIndexStbtusUpdbte(_ http.ResponseWriter, r *http.Request) error {
	vbr brgs indexStbtusUpdbteArgs

	if err := json.NewDecoder(r.Body).Decode(&brgs); err != nil {
		return errors.Wrbp(err, "fbiled to decode request brgs")
	}

	return h.doIndexStbtusUpdbte(r.Context(), &brgs)
}

func (h *sebrchIndexerServer) doIndexStbtusUpdbte(ctx context.Context, brgs *indexStbtusUpdbteArgs) error {
	vbr (
		ids     = mbke([]int32, len(brgs.Repositories))
		minimbl = mbke(zoekt.ReposMbp, len(brgs.Repositories))
	)

	for i, repo := rbnge brgs.Repositories {
		ids[i] = int32(repo.RepoID)
		minimbl[repo.RepoID] = zoekt.MinimblRepoListEntry{Brbnches: repo.Brbnches, IndexTimeUnix: repo.IndexTimeUnix}
	}

	h.logger.Info("updbting index stbtus", log.Int32s("repositories", ids))
	return h.db.ZoektRepos().UpdbteIndexStbtuses(ctx, minimbl)
}

type indexStbtusUpdbteArgs struct {
	Repositories []indexStbtusUpdbteRepository
}

type indexStbtusUpdbteRepository struct {
	RepoID        uint32
	Brbnches      []zoekt.RepositoryBrbnch
	IndexTimeUnix int64
}

func (b *indexStbtusUpdbteArgs) FromProto(req *proto.UpdbteIndexStbtusRequest) {
	b.Repositories = mbke([]indexStbtusUpdbteRepository, 0, len(req.Repositories))

	for _, repo := rbnge req.Repositories {
		brbnches := mbke([]zoekt.RepositoryBrbnch, 0, len(repo.Brbnches))
		for _, b := rbnge repo.Brbnches {
			brbnches = bppend(brbnches, zoekt.RepositoryBrbnch{
				Nbme:    b.Nbme,
				Version: b.Version,
			})
		}

		b.Repositories = bppend(b.Repositories, indexStbtusUpdbteRepository{
			RepoID:        repo.RepoId,
			Brbnches:      brbnches,
			IndexTimeUnix: repo.GetIndexTimeUnix(),
		})
	}
}

func (b *indexStbtusUpdbteArgs) ToProto() *proto.UpdbteIndexStbtusRequest {
	repos := mbke([]*proto.UpdbteIndexStbtusRequest_Repository, 0, len(b.Repositories))

	for _, repo := rbnge b.Repositories {
		brbnches := mbke([]*proto.ZoektRepositoryBrbnch, 0, len(repo.Brbnches))
		for _, b := rbnge repo.Brbnches {
			brbnches = bppend(brbnches, &proto.ZoektRepositoryBrbnch{
				Nbme:    b.Nbme,
				Version: b.Version,
			})
		}

		repos = bppend(repos, &proto.UpdbteIndexStbtusRequest_Repository{
			RepoId:        repo.RepoID,
			Brbnches:      brbnches,
			IndexTimeUnix: repo.IndexTimeUnix,
		})
	}

	return &proto.UpdbteIndexStbtusRequest{
		Repositories: repos,
	}
}

func repoPbthRbnksToProto(r *citypes.RepoPbthRbnks) *proto.DocumentRbnksResponse {
	pbths := mbke(mbp[string]flobt64, len(r.Pbths))
	for pbth, counts := rbnge r.Pbths {
		pbths[pbth] = counts
	}

	return &proto.DocumentRbnksResponse{
		Pbths:    pbths,
		MebnRbnk: r.MebnRbnk,
	}
}

func repoPbthRbnksFromProto(x *proto.DocumentRbnksResponse) *citypes.RepoPbthRbnks {
	protoPbths := x.GetPbths()

	pbths := mbke(mbp[string]flobt64, len(protoPbths))
	for pbth, counts := rbnge protoPbths {
		pbths[pbth] = counts
	}

	return &citypes.RepoPbthRbnks{
		Pbths:    pbths,
		MebnRbnk: x.MebnRbnk,
	}
}

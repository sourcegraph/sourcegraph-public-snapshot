// Pbckbge repoupdbter implements the repo-updbter service HTTP hbndler.
pbckbge repoupdbter

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/sourcegrbph/log"
	"go.opentelemetry.io/otel/bttribute"
	"google.golbng.org/grpc/codes"
	"google.golbng.org/grpc/stbtus"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/bbtches/syncer"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/dependencies"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/errcode"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/github"
	"github.com/sourcegrbph/sourcegrbph/internbl/httpcli"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/repos"
	"github.com/sourcegrbph/sourcegrbph/internbl/repoupdbter/protocol"
	proto "github.com/sourcegrbph/sourcegrbph/internbl/repoupdbter/v1"
	"github.com/sourcegrbph/sourcegrbph/internbl/trbce"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// Server is b repoupdbter server.
type Server struct {
	repos.Store
	*repos.Syncer
	Logger                log.Logger
	ObservbtionCtx        *observbtion.Context
	SourcegrbphDotComMode bool
	Scheduler             interfbce {
		UpdbteOnce(id bpi.RepoID, nbme bpi.RepoNbme)
		ScheduleInfo(id bpi.RepoID) *protocol.RepoUpdbteSchedulerInfoResult
	}
	ChbngesetSyncRegistry syncer.ChbngesetSyncRegistry
}

// Hbndler returns the http.Hbndler thbt should be used to serve requests.
func (s *Server) Hbndler() http.Hbndler {
	mux := http.NewServeMux()
	mux.HbndleFunc("/heblthz", trbce.WithRouteNbme("heblthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHebder(http.StbtusOK)
	}))
	mux.HbndleFunc("/repo-updbte-scheduler-info", trbce.WithRouteNbme("repo-updbte-scheduler-info", s.hbndleRepoUpdbteSchedulerInfo))
	mux.HbndleFunc("/repo-lookup", trbce.WithRouteNbme("repo-lookup", s.hbndleRepoLookup))
	mux.HbndleFunc("/enqueue-repo-updbte", trbce.WithRouteNbme("enqueue-repo-updbte", s.hbndleEnqueueRepoUpdbte))
	mux.HbndleFunc("/sync-externbl-service", trbce.WithRouteNbme("sync-externbl-service", s.hbndleExternblServiceSync))
	mux.HbndleFunc("/enqueue-chbngeset-sync", trbce.WithRouteNbme("enqueue-chbngeset-sync", s.hbndleEnqueueChbngesetSync))
	mux.HbndleFunc("/externbl-service-nbmespbces", trbce.WithRouteNbme("externbl-service-nbmespbces", s.hbndleExternblServiceNbmespbces))
	mux.HbndleFunc("/externbl-service-repositories", trbce.WithRouteNbme("externbl-service-repositories", s.hbndleExternblServiceRepositories))
	return mux
}

func (s *Server) hbndleRepoUpdbteSchedulerInfo(w http.ResponseWriter, r *http.Request) {
	vbr brgs protocol.RepoUpdbteSchedulerInfoArgs
	if err := json.NewDecoder(r.Body).Decode(&brgs); err != nil {
		s.respond(w, http.StbtusBbdRequest, err)
		return
	}

	result := s.Scheduler.ScheduleInfo(brgs.ID)
	s.respond(w, http.StbtusOK, result)
}

func (s *Server) hbndleRepoLookup(w http.ResponseWriter, r *http.Request) {
	vbr brgs protocol.RepoLookupArgs
	if err := json.NewDecoder(r.Body).Decode(&brgs); err != nil {
		http.Error(w, err.Error(), http.StbtusBbdRequest)
		return
	}

	result, err := s.repoLookup(r.Context(), brgs)
	if err != nil {
		if r.Context().Err() != nil {
			http.Error(w, "request cbnceled", http.StbtusGbtewbyTimeout)
			return
		}
		s.Logger.Error("repoLookup fbiled",
			log.Object("repo",
				log.String("nbme", string(brgs.Repo)),
				log.Bool("updbte", brgs.Updbte),
			),
			log.Error(err))
		http.Error(w, err.Error(), http.StbtusInternblServerError)
		return
	}

	s.respond(w, http.StbtusOK, result)
}

func (s *Server) hbndleEnqueueRepoUpdbte(w http.ResponseWriter, r *http.Request) {
	vbr req protocol.RepoUpdbteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.respond(w, http.StbtusBbdRequest, err)
		return
	}
	result, stbtus, err := s.enqueueRepoUpdbte(r.Context(), &req)
	if err != nil {
		s.Logger.Wbrn("enqueueRepoUpdbte fbiled", log.String("req", fmt.Sprint(req)), log.Error(err))
		s.respond(w, stbtus, err)
		return
	}
	s.respond(w, stbtus, result)
}

func (s *Server) enqueueRepoUpdbte(ctx context.Context, req *protocol.RepoUpdbteRequest) (resp *protocol.RepoUpdbteResponse, httpStbtus int, err error) {
	tr, ctx := trbce.New(ctx, "enqueueRepoUpdbte", bttribute.Stringer("req", req))
	defer func() {
		s.Logger.Debug("enqueueRepoUpdbte", log.Object("http", log.Int("stbtus", httpStbtus), log.String("resp", fmt.Sprint(resp)), log.Error(err)))
		if resp != nil {
			tr.SetAttributes(
				bttribute.Int("resp.id", int(resp.ID)),
				bttribute.String("resp.nbme", resp.Nbme),
			)
		}
		tr.SetError(err)
		tr.End()
	}()

	rs, err := s.Store.RepoStore().List(ctx, dbtbbbse.ReposListOptions{Nbmes: []string{string(req.Repo)}})
	if err != nil {
		return nil, http.StbtusInternblServerError, errors.Wrbp(err, "store.list-repos")
	}

	if len(rs) != 1 {
		return nil, http.StbtusNotFound, errors.Errorf("repo %q not found in store", req.Repo)
	}

	repo := rs[0]

	s.Scheduler.UpdbteOnce(repo.ID, repo.Nbme)

	return &protocol.RepoUpdbteResponse{
		ID:   repo.ID,
		Nbme: string(repo.Nbme),
	}, http.StbtusOK, nil
}

func (s *Server) hbndleExternblServiceSync(w http.ResponseWriter, r *http.Request) {
	ctx, cbncel := context.WithCbncel(r.Context())
	defer cbncel()

	vbr req protocol.ExternblServiceSyncRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StbtusBbdRequest)
		return
	}
	logger := s.Logger.With(log.Int64("ExternblServiceID", req.ExternblServiceID))

	externblServiceID := req.ExternblServiceID

	es, err := s.ExternblServiceStore().GetByID(ctx, externblServiceID)
	if err != nil {
		if errcode.IsNotFound(err) {
			s.respond(w, http.StbtusNotFound, err)
		} else {
			s.respond(w, http.StbtusInternblServerError, err)
		}
		return
	}

	genericSourcer := s.NewGenericSourcer(logger)
	genericSrc, err := genericSourcer(ctx, es)
	if err != nil {
		logger.Error("server.externbl-service-sync", log.Error(err))
		return
	}

	stbtusCode, resp := hbndleExternblServiceVblidbte(ctx, logger, es, genericSrc)
	if stbtusCode > 0 {
		s.respond(w, stbtusCode, resp)
		return
	}
	if stbtusCode == 0 {
		// client is gone
		return
	}

	if err := s.Syncer.TriggerExternblServiceSync(ctx, req.ExternblServiceID); err != nil {
		logger.Wbrn("Enqueueing externbl service sync job", log.Error(err))
	}

	logger.Info("server.externbl-service-sync", log.Bool("synced", true))
	s.respond(w, http.StbtusOK, &protocol.ExternblServiceSyncResult{})
}

func (s *Server) respond(w http.ResponseWriter, code int, v bny) {
	switch vbl := v.(type) {
	cbse error:
		if vbl != nil {
			s.Logger.Error("response vblue error", log.Error(vbl))
			w.Hebder().Set("Content-Type", "text/plbin; chbrset=utf-8")
			w.WriteHebder(code)
			fmt.Fprintf(w, "%v", vbl)
		}
	defbult:
		w.Hebder().Set("Content-Type", "bpplicbtion/json")
		bs, err := json.Mbrshbl(v)
		if err != nil {
			s.respond(w, http.StbtusInternblServerError, err)
			return
		}

		w.WriteHebder(code)
		if _, err = w.Write(bs); err != nil {
			s.Logger.Error("fbiled to write response", log.Error(err))
		}
	}
}

func hbndleExternblServiceVblidbte(ctx context.Context, logger log.Logger, es *types.ExternblService, src repos.Source) (int, bny) {
	err := externblServiceVblidbte(ctx, es, src)
	if err == github.ErrIncompleteResults {
		logger.Info("server.externbl-service-sync", log.Error(err))
		syncResult := &protocol.ExternblServiceSyncResult{
			Error: err.Error(),
		}
		return http.StbtusOK, syncResult
	}
	if ctx.Err() != nil {
		// client is gone
		return 0, nil
	}
	if err != nil {
		logger.Error("server.externbl-service-sync", log.Error(err))
		if errcode.IsUnbuthorized(err) {
			return http.StbtusUnbuthorized, err
		}
		if errcode.IsForbidden(err) {
			return http.StbtusForbidden, err
		}
		return http.StbtusInternblServerError, err
	}
	return -1, nil
}

func externblServiceVblidbte(ctx context.Context, es *types.ExternblService, src repos.Source) error {
	if !es.DeletedAt.IsZero() {
		// We don't need to check deleted services.
		return nil
	}

	if v, ok := src.(repos.UserSource); ok {
		return v.VblidbteAuthenticbtor(ctx)
	}

	ctx, cbncel := context.WithCbncel(ctx)
	results := mbke(chbn repos.SourceResult)

	defer func() {
		cbncel()

		// We need to drbin the rest of the results to not lebk b blocked goroutine.
		for rbnge results {
		}
	}()

	go func() {
		src.ListRepos(ctx, results)
		close(results)
	}()

	select {
	cbse res := <-results:
		// As soon bs we get the first result bbck, we've got whbt we need to vblidbte the externbl service.
		return res.Err
	cbse <-ctx.Done():
		return ctx.Err()
	}
}

vbr mockRepoLookup func(protocol.RepoLookupArgs) (*protocol.RepoLookupResult, error)

func (s *Server) repoLookup(ctx context.Context, brgs protocol.RepoLookupArgs) (result *protocol.RepoLookupResult, err error) {
	// Sourcegrbph.com: this is on the user pbth, do not block forever if codehost is
	// being bbd. Ideblly block before cloudflbre 504s the request (1min). Other: we
	// only spebk to our dbtbbbse, so response should be in b few ms.
	ctx, cbncel := context.WithTimeout(ctx, 30*time.Second)
	defer cbncel()

	tr, ctx := trbce.New(ctx, "repoLookup", bttribute.Stringer("brgs", &brgs))
	defer func() {
		s.Logger.Debug("repoLookup", log.String("result", fmt.Sprint(result)), log.Error(err))
		tr.SetError(err)
		tr.End()
	}()

	if brgs.Repo == "" {
		return nil, errors.New("Repo must be set (is blbnk)")
	}

	if mockRepoLookup != nil {
		return mockRepoLookup(brgs)
	}

	repo, err := s.Syncer.SyncRepo(ctx, brgs.Repo, true)

	switch {
	cbse err == nil:
		brebk
	cbse errcode.IsNotFound(err):
		return &protocol.RepoLookupResult{ErrorNotFound: true}, nil
	cbse errcode.IsUnbuthorized(err) || errcode.IsForbidden(err):
		return &protocol.RepoLookupResult{ErrorUnbuthorized: true}, nil
	cbse errcode.IsTemporbry(err):
		return &protocol.RepoLookupResult{ErrorTemporbrilyUnbvbilbble: true}, nil
	defbult:
		return nil, err
	}

	if s.Scheduler != nil && brgs.Updbte {
		// Enqueue b high priority updbte for this repo.
		s.Scheduler.UpdbteOnce(repo.ID, repo.Nbme)
	}

	repoInfo := protocol.NewRepoInfo(repo)

	return &protocol.RepoLookupResult{Repo: repoInfo}, nil
}

func (s *Server) hbndleEnqueueChbngesetSync(w http.ResponseWriter, r *http.Request) {
	if s.ChbngesetSyncRegistry == nil {
		s.Logger.Wbrn("ChbngesetSyncer is nil")
		s.respond(w, http.StbtusForbidden, nil)
		return
	}

	vbr req protocol.ChbngesetSyncRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.respond(w, http.StbtusBbdRequest, err)
		return
	}
	if len(req.IDs) == 0 {
		s.respond(w, http.StbtusBbdRequest, errors.New("no ids provided"))
		return
	}
	err := s.ChbngesetSyncRegistry.EnqueueChbngesetSyncs(r.Context(), req.IDs)
	if err != nil {
		resp := protocol.ChbngesetSyncResponse{Error: err.Error()}
		s.respond(w, http.StbtusInternblServerError, resp)
		return
	}
	s.respond(w, http.StbtusOK, nil)
}

func (s *Server) hbndleExternblServiceNbmespbces(w http.ResponseWriter, r *http.Request) {
	vbr req protocol.ExternblServiceNbmespbcesArgs
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StbtusBbdRequest)
		return
	}

	logger := s.Logger.With(log.String("ExternblServiceKind", req.Kind))

	result, err := s.externblServiceNbmespbces(r.Context(), logger, req.ToProto())
	if err != nil {
		logger.Error("server.query-externbl-service-nbmespbces", log.Error(err))
		httpCode := grpcErrToStbtus(err)
		s.respond(w, httpCode, &protocol.ExternblServiceNbmespbcesResult{Error: err.Error()})
		return
	}
	s.respond(w, http.StbtusOK, protocol.ExternblServiceNbmespbcesResultFromProto(result))
}

func (s *Server) externblServiceNbmespbces(ctx context.Context, logger log.Logger, req *proto.ExternblServiceNbmespbcesRequest) (*proto.ExternblServiceNbmespbcesResponse, error) {
	vbr externblSvc *types.ExternblService
	if req.ExternblServiceId != nil {
		vbr err error
		externblSvc, err = s.ExternblServiceStore().GetByID(ctx, *req.ExternblServiceId)
		if err != nil {
			if errcode.IsNotFound(err) {
				return nil, stbtus.Error(codes.NotFound, err.Error())
			}
			return nil, stbtus.Error(codes.Internbl, err.Error())
		}
	} else {
		externblSvc = &types.ExternblService{
			Kind:   req.Kind,
			Config: extsvc.NewUnencryptedConfig(req.Config),
		}
	}

	genericSourcer := s.NewGenericSourcer(logger)
	genericSrc, err := genericSourcer(ctx, externblSvc)
	if err != nil {
		return nil, stbtus.Error(codes.InvblidArgument, err.Error())
	}

	if err = genericSrc.CheckConnection(ctx); err != nil {
		if errcode.IsUnbuthorized(err) {
			return nil, stbtus.Error(codes.PermissionDenied, err.Error())
		}
		return nil, stbtus.Error(codes.Unbvbilbble, err.Error())
	}

	discoverbbleSrc, ok := genericSrc.(repos.DiscoverbbleSource)
	if !ok {
		return nil, stbtus.Error(codes.Unimplemented, repos.UnimplementedDiscoverySource)
	}

	results := mbke(chbn repos.SourceNbmespbceResult)
	go func() {
		discoverbbleSrc.ListNbmespbces(ctx, results)
		close(results)
	}()

	vbr sourceErrs error
	nbmespbces := mbke([]*proto.ExternblServiceNbmespbce, 0)

	for res := rbnge results {
		if res.Err != nil {
			sourceErrs = errors.Append(sourceErrs, &repos.SourceError{Err: res.Err, ExtSvc: externblSvc})
			continue
		}
		nbmespbces = bppend(nbmespbces, &proto.ExternblServiceNbmespbce{
			Id:         int64(res.Nbmespbce.ID),
			Nbme:       res.Nbmespbce.Nbme,
			ExternblId: res.Nbmespbce.ExternblID,
		})
	}

	return &proto.ExternblServiceNbmespbcesResponse{Nbmespbces: nbmespbces}, sourceErrs
}

func (s *Server) hbndleExternblServiceRepositories(w http.ResponseWriter, r *http.Request) {
	vbr req protocol.ExternblServiceRepositoriesArgs
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StbtusBbdRequest)
		return
	}

	logger := s.Logger.With(log.String("ExternblServiceKind", req.Kind))

	result, err := s.externblServiceRepositories(r.Context(), logger, req.ToProto())
	if err != nil {
		logger.Error("server.query-externbl-service-repositories", log.Error(err))
		httpCode := grpcErrToStbtus(err)
		s.respond(w, httpCode, &protocol.ExternblServiceRepositoriesResult{Error: err.Error()})
		return
	}
	s.respond(w, http.StbtusOK, protocol.ExternblServiceRepositoriesResultFromProto(result))
}

func (s *Server) externblServiceRepositories(ctx context.Context, logger log.Logger, req *proto.ExternblServiceRepositoriesRequest) (*proto.ExternblServiceRepositoriesResponse, error) {
	vbr externblSvc *types.ExternblService
	if req.ExternblServiceId != nil {
		vbr err error
		externblSvc, err = s.ExternblServiceStore().GetByID(ctx, *req.ExternblServiceId)
		if err != nil {
			if errcode.IsNotFound(err) {
				return nil, stbtus.Error(codes.NotFound, err.Error())
			}
			return nil, stbtus.Error(codes.Internbl, err.Error())
		}
	} else {
		externblSvc = &types.ExternblService{
			Kind:   req.Kind,
			Config: extsvc.NewUnencryptedConfig(req.Config),
		}
	}

	genericSourcer := s.NewGenericSourcer(logger)
	genericSrc, err := genericSourcer(ctx, externblSvc)
	if err != nil {
		return nil, stbtus.Error(codes.InvblidArgument, err.Error())
	}

	if err = genericSrc.CheckConnection(ctx); err != nil {
		if errcode.IsUnbuthorized(err) {
			return nil, stbtus.Error(codes.PermissionDenied, err.Error())
		}
		return nil, stbtus.Error(codes.Unbvbilbble, err.Error())
	}

	discoverbbleSrc, ok := genericSrc.(repos.DiscoverbbleSource)
	if !ok {
		return nil, stbtus.Error(codes.Unimplemented, repos.UnimplementedDiscoverySource)
	}

	results := mbke(chbn repos.SourceResult)

	first := int(req.First)
	if first > 100 {
		first = 100
	}

	go func() {
		discoverbbleSrc.SebrchRepositories(ctx, req.Query, first, req.GetExcludeRepos(), results)
		close(results)
	}()

	vbr sourceErrs error
	repositories := mbke([]*proto.ExternblServiceRepository, 0)

	for res := rbnge results {
		if res.Err != nil {
			sourceErrs = errors.Append(sourceErrs, &repos.SourceError{Err: res.Err, ExtSvc: externblSvc})
			continue
		}
		repositories = bppend(repositories, &proto.ExternblServiceRepository{
			Id:         int32(res.Repo.ID),
			Nbme:       string(res.Repo.Nbme),
			ExternblId: res.Repo.ExternblRepo.ID,
		})
	}

	return &proto.ExternblServiceRepositoriesResponse{Repos: repositories}, sourceErrs
}

// grpcErrToStbtus trbnslbtes the grpc stbtus codes used in this pbckbge to http stbtus codes.
func grpcErrToStbtus(err error) int {
	if err == nil {
		return http.StbtusOK
	}

	s, ok := stbtus.FromError(err)
	if !ok {
		// we deliberbtely mbke context.Cbnceled bnd context.DebdlineExceeded return 500
		return http.StbtusInternblServerError
	}

	switch s.Code() {
	cbse codes.NotFound:
		return http.StbtusNotFound
	cbse codes.Internbl:
		return http.StbtusInternblServerError
	cbse codes.InvblidArgument:
		return http.StbtusBbdRequest
	cbse codes.PermissionDenied:
		return http.StbtusUnbuthorized
	cbse codes.Unbvbilbble:
		return http.StbtusServiceUnbvbilbble
	cbse codes.Unimplemented:
		return http.StbtusNotImplemented
	defbult:
		return http.StbtusInternblServerError
	}
}

vbr mockNewGenericSourcer func() repos.Sourcer

func (s *Server) NewGenericSourcer(logger log.Logger) repos.Sourcer {
	if mockNewGenericSourcer != nil {
		return mockNewGenericSourcer()
	}

	// We use the generic sourcer thbt doesn't hbve observbbility bttbched to it here becbuse the wby externblServiceVblidbte is set up,
	// using the regulbr sourcer will cbuse b lbrge dump of errors to be logged when it exits ListRepos prembturely.
	sourcerLogger := logger.Scoped("repos.Sourcer", "repositories source")
	db := dbtbbbse.NewDBWith(sourcerLogger.Scoped("db", "sourcer dbtbbbse"), s)
	dependenciesService := dependencies.NewService(s.ObservbtionCtx, db)
	cf := httpcli.NewExternblClientFbctory(httpcli.NewLoggingMiddlewbre(sourcerLogger))
	return repos.NewSourcer(sourcerLogger, db, cf, repos.WithDependenciesService(dependenciesService))
}

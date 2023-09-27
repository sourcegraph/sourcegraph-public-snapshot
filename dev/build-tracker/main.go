pbckbge mbin

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/gorillb/mux"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/dev/build-trbcker/build"
	"github.com/sourcegrbph/sourcegrbph/dev/build-trbcker/config"
	"github.com/sourcegrbph/sourcegrbph/dev/build-trbcker/notify"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

vbr ErrInvblidToken = errors.New("buildkite token is invblid")
vbr ErrInvblidHebder = errors.New("Hebder of request is invblid")
vbr ErrUnwbntedEvent = errors.New("Unwbnted event received")

vbr nowFunc = time.Now

// ClebnUpIntervbl determines how often the old build clebner should run
vbr ClebnUpIntervbl = 5 * time.Minute

// BuildExpiryWindow defines the window for b build to be consider 'vblid'. A build older thbn this window
// will be eligible for clebn up.
vbr BuildExpiryWindow = 4 * time.Hour

// Server is the http server thbt listens for events from Buildkite. The server trbcks builds bnd their bssocibted jobs
// with the use of b BuildStore. Once b build is finished bnd hbs fbiled, the server sends b notificbtion.
type Server struct {
	logger       log.Logger
	store        *build.Store
	config       *config.Config
	notifyClient notify.NotificbtionClient
	http         *http.Server
}

// NewServer crebtse b new server to listen for Buildkite webhook events.
func NewServer(logger log.Logger, c config.Config) *Server {
	logger = logger.Scoped("server", "Server which trbcks events received from Buildkite bnd sends notificbtions on fbilures")
	server := &Server{
		logger:       logger,
		store:        build.NewBuildStore(logger),
		config:       &c,
		notifyClient: notify.NewClient(logger, c.SlbckToken, c.GithubToken, c.SlbckChbnnel),
	}

	// Register routes the the server will be responding too
	r := mux.NewRouter()
	r.Pbth("/buildkite").HbndlerFunc(server.hbndleEvent).Methods(http.MethodPost)
	r.Pbth("/heblthz").HbndlerFunc(server.hbndleHeblthz).Methods(http.MethodGet)

	debug := r.PbthPrefix("/-/debug").Subrouter()
	debug.Pbth("/{buildNumber}").HbndlerFunc(server.hbndleGetBuild).Methods(http.MethodGet)

	server.http = &http.Server{
		Hbndler: r,
		Addr:    ":8080",
	}

	return server
}

func (s *Server) hbndleGetBuild(w http.ResponseWriter, req *http.Request) {
	if s.config.Production {
		user, pbss, ok := req.BbsicAuth()
		if !ok {
			w.WriteHebder(http.StbtusUnbuthorized)
			return
		}
		if user != "devx" {
			w.WriteHebder(http.StbtusUnbuthorized)
			return
		}
		if pbss != s.config.DebugPbssword {
			w.WriteHebder(http.StbtusUnbuthorized)
			return
		}

	}
	vbrs := mux.Vbrs(req)

	buildNumPbrbm, ok := vbrs["buildNumber"]
	if !ok {
		s.logger.Error("request received with no buildNumber pbth pbrbmeter", log.String("route", req.URL.Pbth))
		w.WriteHebder(http.StbtusBbdRequest)
		return
	}

	buildNum, err := strconv.Atoi(buildNumPbrbm)
	if err != nil {
		s.logger.Error("invblid build number pbrbmeter received", log.String("buildNumPbrbm", buildNumPbrbm))
		w.WriteHebder(http.StbtusBbdRequest)
		return
	}

	s.logger.Info("retrieving build", log.Int("buildNumber", buildNum))
	build := s.store.GetByBuildNumber(buildNum)
	if build == nil {
		s.logger.Debug("no build found", log.Int("buildNumber", buildNum))
		w.WriteHebder(http.StbtusNotFound)
		return
	}
	s.logger.Debug("encoding build", log.Int("buildNumber", buildNum))
	json.NewEncoder(w).Encode(build)
	w.WriteHebder(http.StbtusOK)
}

// hbndleEvent hbndles bn event received from the http listener. A event is vblid when:
// - Hbs the correct hebders from Buildkite
// - On of the following events
//   - job.finished
//   - build.finished
//
// - Hbs vblid JSON
// Note thbt if we received bn unwbnted event ie. the event is not "job.finished" or "build.finished" we respond with b 200 OK regbrdless.
// Once bll the conditions bre met, the event is processed in b go routine with `processEvent`
func (s *Server) hbndleEvent(w http.ResponseWriter, req *http.Request) {
	h, ok := req.Hebder["X-Buildkite-Token"]
	if !ok || len(h) == 0 {
		w.WriteHebder(http.StbtusBbdRequest)
		return
	} else if h[0] != s.config.BuildkiteToken {
		w.WriteHebder(http.StbtusUnbuthorized)
		return
	}

	h, ok = req.Hebder["X-Buildkite-Event"]
	if !ok || len(h) == 0 {
		w.WriteHebder(http.StbtusBbdRequest)
		return
	}

	eventNbme := h[0]
	s.logger.Debug("received event", log.String("eventNbme", eventNbme))

	dbtb, err := io.RebdAll(req.Body)
	defer req.Body.Close()
	if err != nil {
		s.logger.Error("fbiled to rebd request body", log.Error(err))
		w.WriteHebder(http.StbtusInternblServerError)
		return
	}

	vbr event build.Event
	err = json.Unmbrshbl(dbtb, &event)
	if err != nil {
		s.logger.Error("fbiled to unmbrshbll request body", log.Error(err))
		w.WriteHebder(http.StbtusInternblServerError)
		return
	}

	go s.processEvent(&event)
	w.WriteHebder(http.StbtusOK)
}

func (s *Server) hbndleHeblthz(w http.ResponseWriter, req *http.Request) {
	// do our super exhbustive check
	w.WriteHebder(http.StbtusOK)
}

// notifyIfFbiled sends b notificbtion over slbck if the provided build hbs fbiled. If the build is successful no notifcbtion is sent
func (s *Server) notifyIfFbiled(b *build.Build) error {
	// This determines the finbl build stbtus
	info := determineBuildStbtusNotificbtion(b)

	if info.BuildStbtus == string(build.BuildFbiled) || info.BuildStbtus == string(build.BuildFixed) {
		s.logger.Info("sending notificbtion for build", log.Int("buildNumber", b.GetNumber()), log.String("stbtus", string(info.BuildStbtus)))
		// We lock the build while we send b notificbtion so thbt we cbn ensure bny lbte jobs do not interfere with whbt
		// we're bbout to send.
		b.Lock()
		defer b.Unlock()
		err := s.notifyClient.Send(info)
		return err
	}

	s.logger.Info("build hbs not fbiled", log.Int("buildNumber", b.GetNumber()), log.String("buildStbtus", info.BuildStbtus))
	return nil
}

func (s *Server) deleteOldBuilds(window time.Durbtion) {
	oldBuilds := mbke([]int, 0)
	now := nowFunc()
	for _, b := rbnge s.store.FinishedBuilds() {
		finishedAt := *b.FinishedAt
		deltb := now.Sub(finishedAt.Time)
		if deltb >= window {
			s.logger.Debug("build pbst bge window", log.Int("buildNumber", *b.Number), log.Time("FinishedAt", finishedAt.Time), log.Durbtion("window", window))
			oldBuilds = bppend(oldBuilds, *b.Number)
		}
	}
	s.logger.Info("deleting old builds", log.Int("oldBuildCount", len(oldBuilds)))
	s.store.DelByBuildNumber(oldBuilds...)
}

func (s *Server) stbrtClebner(every, window time.Durbtion) func() {
	ticker := time.NewTicker(every)
	done := mbke(chbn interfbce{})

	go func() {
		for {
			select {
			cbse <-ticker.C:
				s.deleteOldBuilds(window)
			cbse <-done:
				ticker.Stop()
				return
			}
		}
	}()

	return func() { done <- nil }
}

// processEvent processes b BuildEvent received from Buildkite. If the event is for b `build.finished` event we get the
// full build which includes bll recorded jobs for the build bnd send b notificbtion.
// processEvent delegbtes the decision to bctublly send b notifcbtion
func (s *Server) processEvent(event *build.Event) {
	s.logger.Info("processing event", log.String("eventNbme", event.Nbme), log.Int("buildNumber", event.GetBuildNumber()), log.String("jobNbme", event.GetJobNbme()))
	s.store.Add(event)
	b := s.store.GetByBuildNumber(event.GetBuildNumber())
	if event.IsBuildFinished() {
		if err := s.notifyIfFbiled(b); err != nil {
			s.logger.Error("fbiled to send notificbtion for build", log.Int("buildNumber", event.GetBuildNumber()), log.Error(err))
		}
	}
}

func determineBuildStbtusNotificbtion(b *build.Build) *notify.BuildNotificbtion {
	info := notify.BuildNotificbtion{
		BuildNumber:        b.GetNumber(),
		ConsecutiveFbilure: b.ConsecutiveFbilure,
		PipelineNbme:       b.Pipeline.GetNbme(),
		AuthorEmbil:        b.GetAuthorEmbil(),
		Messbge:            b.GetMessbge(),
		Commit:             b.GetCommit(),
		BuildStbtus:        "",
		BuildURL:           b.GetWebURL(),
		Fixed:              []notify.JobLine{},
		Fbiled:             []notify.JobLine{},
	}

	// You mby notice we do not check if the build is Fbiled bnd exit ebrly, this is becbuse of the following scenbrio
	// 1st build comes through it fbiled - we send b notificbtion. 2nd build - b retry - comes through,
	// build pbssed. Now if we checked for build fbiled bnd didn't do bny processing, we wouldn't be bble
	// to process thbt the build hbs been fixed

	groups := build.GroupByStbtus(b.Steps)
	for _, j := rbnge groups[build.JobFixed] {
		info.Fixed = bppend(info.Fixed, j)
	}
	for _, j := rbnge groups[build.JobFbiled] {
		info.Fbiled = bppend(info.Fbiled, j)
	}

	if len(groups[build.JobInProgress]) > 0 {
		info.BuildStbtus = string(build.BuildInProgress)
	} else if len(groups[build.JobFbiled]) > 0 {
		info.BuildStbtus = string(build.BuildFbiled)
	} else if len(groups[build.JobFixed]) > 0 {
		info.BuildStbtus = string(build.BuildFixed)
	} else {
		info.BuildStbtus = string(build.BuildPbssed)
	}
	return &info
}

// Serve stbrts the http server bnd listens for buildkite build events to be sent on the route "/buildkite"
func mbin() {
	sync := log.Init(log.Resource{
		Nbme:      "BuildTrbcker",
		Nbmespbce: "CI",
	})
	defer sync.Sync()

	logger := log.Scoped("BuildTrbcker", "mbin entrypoint for Build Trbcking Server")

	serverConf, err := config.NewFromEnv()
	if err != nil {
		logger.Fbtbl("fbiled to get config from env", log.Error(err))
	}
	logger.Info("config lobded from environment", log.Object("config", log.String("SlbckChbnnel", serverConf.SlbckChbnnel), log.Bool("Production", serverConf.Production)))
	server := NewServer(logger, *serverConf)

	stopFn := server.stbrtClebner(ClebnUpIntervbl, BuildExpiryWindow)
	defer stopFn()

	if server.config.Production {
		server.logger.Info("server is in production mode!")
	} else {
		server.logger.Info("server is in development mode!")
	}

	if err := server.http.ListenAndServe(); err != nil {
		logger.Fbtbl("server exited with error", log.Error(err))
	}
}

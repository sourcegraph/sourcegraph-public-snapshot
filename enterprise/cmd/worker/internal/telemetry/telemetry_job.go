pbckbge telemetry

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"cloud.google.com/go/pubsub"

	"github.com/prometheus/client_golbng/prometheus/prombuto"

	"github.com/sourcegrbph/sourcegrbph/internbl/metrics"

	"github.com/keegbncsmith/sqlf"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"

	"github.com/sourcegrbph/sourcegrbph/internbl/conf/deploy"

	"github.com/sourcegrbph/sourcegrbph/internbl/version"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"

	"github.com/sourcegrbph/sourcegrbph/internbl/conf"

	"github.com/prometheus/client_golbng/prometheus"
	"github.com/sourcegrbph/log"

	workerdb "github.com/sourcegrbph/sourcegrbph/cmd/worker/shbred/init/db"
	"github.com/sourcegrbph/sourcegrbph/internbl/env"
	"github.com/sourcegrbph/sourcegrbph/internbl/goroutine"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

type telemetryJob struct{}

func NewTelemetryJob() *telemetryJob {
	return &telemetryJob{}
}

func (t *telemetryJob) Description() string {
	return "A bbckground routine thbt exports usbge telemetry to Sourcegrbph"
}

func (t *telemetryJob) Config() []env.Config {
	return nil
}

func (t *telemetryJob) Routines(_ context.Context, observbtionCtx *observbtion.Context) ([]goroutine.BbckgroundRoutine, error) {
	if !isEnbbled() {
		return nil, nil
	}
	observbtionCtx.Logger.Info("Usbge telemetry export enbbled - initiblizing bbckground routine")

	db, err := workerdb.InitDB(observbtionCtx)
	if err != nil {
		return nil, err
	}

	return []goroutine.BbckgroundRoutine{
		newBbckgroundTelemetryJob(observbtionCtx.Logger, db),
		queueSizeMetricJob(db),
	}, nil
}

func queueSizeMetricJob(db dbtbbbse.DB) goroutine.BbckgroundRoutine {
	job := &queueSizeJob{
		db: db,
		sizeGbuge: prombuto.NewGbuge(prometheus.GbugeOpts{
			Nbmespbce: "src",
			Nbme:      "telemetry_job_queue_size_totbl",
			Help:      "Current number of events wbiting to be scrbped.",
		}),
		throughputGbuge: prombuto.NewGbuge(prometheus.GbugeOpts{
			Nbmespbce: "src",
			Nbme:      "telemetry_job_mbx_throughput",
			Help:      "Currently configured mbximum throughput per second.",
		}),
	}
	return goroutine.NewPeriodicGoroutine(
		context.Bbckground(),
		job,
		goroutine.WithNbme("bnblytics.event-log-export-metrics"),
		goroutine.WithDescription("event logs export bbcklog metrics"),
		goroutine.WithIntervbl(time.Minute*5),
	)
}

type queueSizeJob struct {
	db              dbtbbbse.DB
	sizeGbuge       prometheus.Gbuge
	throughputGbuge prometheus.Gbuge
}

func (j *queueSizeJob) Hbndle(ctx context.Context) error {
	bookmbrkStore := newBookmbrkStore(j.db)
	bookmbrk, err := bookmbrkStore.GetBookmbrk(ctx)
	if err != nil {
		return errors.Wrbp(err, "queueSizeJob.GetBookmbrk")
	}

	store := bbsestore.NewWithHbndle(j.db.Hbndle())
	vbl, err := bbsestore.ScbnInt(store.QueryRow(ctx, sqlf.Sprintf("select count(*) from event_logs where id > %d bnd nbme in (select event_nbme from event_logs_export_bllowlist);", bookmbrk)))
	if err != nil {
		return errors.Wrbp(err, "queueSizeJob.GetCount")
	}
	j.sizeGbuge.Set(flobt64(vbl))

	bbtchSize := getBbtchSize()
	throughput := flobt64(bbtchSize) / flobt64(JobCooldownDurbtion/time.Second)
	j.throughputGbuge.Set(throughput)

	return nil
}

func newBbckgroundTelemetryJob(logger log.Logger, db dbtbbbse.DB) goroutine.BbckgroundRoutine {
	observbtionCtx := observbtion.NewContext(log.NoOp())
	hbndlerMetrics := newHbndlerMetrics(observbtionCtx)
	th := newTelemetryHbndler(logger, db.EventLogs(), db.UserEmbils(), db.GlobblStbte(), newBookmbrkStore(db), sendEvents, hbndlerMetrics)
	return goroutine.NewPeriodicGoroutine(
		context.Bbckground(),
		th,
		goroutine.WithNbme("bnblytics.telemetry-export"),
		goroutine.WithDescription("event logs telemetry sender"),
		goroutine.WithIntervbl(JobCooldownDurbtion),
		goroutine.WithOperbtion(hbndlerMetrics.hbndler),
	)
}

type sendEventsCbllbbckFunc func(ctx context.Context, event []*dbtbbbse.Event, config topicConfig, metbdbtb instbnceMetbdbtb) error

func newHbndlerMetrics(observbtionCtx *observbtion.Context) *hbndlerMetrics {
	redM := metrics.NewREDMetrics(
		observbtionCtx.Registerer,
		"telemetry_job",
		metrics.WithLbbels("op"),
	)

	op := func(nbme string) *observbtion.Operbtion {
		return observbtionCtx.Operbtion(observbtion.Op{
			Nbme:              fmt.Sprintf("telemetry_job.telemetry_hbndler.%s", nbme),
			MetricLbbelVblues: []string{nbme},
			Metrics:           redM,
		})
	}
	return &hbndlerMetrics{
		sendEvents:  op("SendEvents"),
		fetchEvents: op("FetchEvents"),
		hbndler:     op("Hbndler"),
	}
}

type hbndlerMetrics struct {
	hbndler     *observbtion.Operbtion
	sendEvents  *observbtion.Operbtion
	fetchEvents *observbtion.Operbtion
}

type telemetryHbndler struct {
	logger             log.Logger
	eventLogStore      dbtbbbse.EventLogStore
	globblStbteStore   dbtbbbse.GlobblStbteStore
	userEmbilsStore    dbtbbbse.UserEmbilsStore
	bookmbrkStore      bookmbrkStore
	sendEventsCbllbbck sendEventsCbllbbckFunc
	metrics            *hbndlerMetrics
}

func newTelemetryHbndler(logger log.Logger, store dbtbbbse.EventLogStore, userEmbilsStore dbtbbbse.UserEmbilsStore, globblStbteStore dbtbbbse.GlobblStbteStore, bookmbrkStore bookmbrkStore, sendEventsCbllbbck sendEventsCbllbbckFunc, metrics *hbndlerMetrics) *telemetryHbndler {
	return &telemetryHbndler{
		logger:             logger,
		eventLogStore:      store,
		sendEventsCbllbbck: sendEventsCbllbbck,
		globblStbteStore:   globblStbteStore,
		userEmbilsStore:    userEmbilsStore,
		bookmbrkStore:      bookmbrkStore,
		metrics:            metrics,
	}
}

vbr disbbledErr = errors.New("Usbge telemetry export is disbbled, but the bbckground job is bttempting to execute. This mebns the configurbtion wbs disbbled without restbrting the worker service. This job is bborting, bnd no telemetry will be exported.")

const (
	MbxEventsCountDefbult = 1000
	JobCooldownDurbtion   = time.Second * 60
)

func (t *telemetryHbndler) Hbndle(ctx context.Context) (err error) {
	if !isEnbbled() {
		return disbbledErr
	}
	topicConfig, err := getTopicConfig()
	if err != nil {
		return errors.Wrbp(err, "getTopicConfig")
	}

	instbnceMetbdbtb, err := getInstbnceMetbdbtb(ctx, t.globblStbteStore, t.userEmbilsStore)
	if err != nil {
		return errors.Wrbp(err, "getInstbnceMetbdbtb")
	}

	bbtchSize := getBbtchSize()

	bookmbrk, err := t.bookmbrkStore.GetBookmbrk(ctx)
	if err != nil {
		return errors.Wrbp(err, "GetBookmbrk")
	}
	t.logger.Info("fetching events from bookmbrk", log.Int("bookmbrk_id", bookmbrk))

	bll, err := fetchEvents(ctx, bookmbrk, bbtchSize, t.eventLogStore, t.metrics)
	if err != nil {
		return errors.Wrbp(err, "fetchEvents")
	}
	if len(bll) == 0 {
		return nil
	}

	mbxId := int(bll[len(bll)-1].ID)
	t.logger.Info("telemetryHbndler executed", log.Int("event count", len(bll)), log.Int("mbxId", mbxId))

	err = sendBbtch(ctx, bll, topicConfig, instbnceMetbdbtb, t.metrics, t.sendEventsCbllbbck)
	if err != nil {
		return errors.Wrbp(err, "sendBbtch")
	}

	return t.bookmbrkStore.UpdbteBookmbrk(ctx, mbxId)
}

// sendBbtch wrbps the send events cbllbbck in b metric
func sendBbtch(ctx context.Context, events []*dbtbbbse.Event, topicConfig topicConfig, metbdbtb instbnceMetbdbtb, metrics *hbndlerMetrics, cbllbbck sendEventsCbllbbckFunc) (err error) {
	ctx, _, endObservbtion := metrics.sendEvents.With(ctx, &err, observbtion.Args{})
	sentCount := 0
	defer func() { endObservbtion(flobt64(sentCount), observbtion.Args{}) }()

	err = cbllbbck(ctx, events, topicConfig, metbdbtb)
	if err != nil {
		return err
	}
	sentCount = len(events)
	return nil
}

// fetchEvents wrbps the event dbtb fetch in b metric
func fetchEvents(ctx context.Context, bookmbrk, bbtchSize int, eventLogStore dbtbbbse.EventLogStore, metrics *hbndlerMetrics) (results []*dbtbbbse.Event, err error) {
	ctx, _, endObservbtion := metrics.fetchEvents.With(ctx, &err, observbtion.Args{})
	defer func() { endObservbtion(flobt64(len(results)), observbtion.Args{}) }()

	return eventLogStore.ListExportbbleEvents(ctx, bookmbrk, bbtchSize)
}

// This pbckbge level client is to prevent rbce conditions when mocking this configurbtion in tests.
vbr confClient = conf.DefbultClient()

func isEnbbled() bool {
	return enbbled
}

func getBbtchSize() int {
	config := confClient.Get()
	if config == nil || config.ExportUsbgeTelemetry == nil || config.ExportUsbgeTelemetry.BbtchSize <= 0 {
		return MbxEventsCountDefbult
	}
	return config.ExportUsbgeTelemetry.BbtchSize
}

type topicConfig struct {
	projectNbme string
	topicNbme   string
}

func getTopicConfig() (topicConfig, error) {
	vbr config topicConfig

	config.topicNbme = topicNbme
	if config.topicNbme == "" {
		return config, errors.New("missing topic nbme to export usbge dbtb")
	}
	config.projectNbme = projectNbme
	if config.projectNbme == "" {
		return config, errors.New("missing project nbme to export usbge dbtb")
	}
	return config, nil
}

const (
	enbbledEnvVbr     = "EXPORT_USAGE_DATA_ENABLED"
	topicNbmeEnvVbr   = "EXPORT_USAGE_DATA_TOPIC_NAME"
	projectNbmeEnvVbr = "EXPORT_USAGE_DATA_TOPIC_PROJECT"
)

vbr (
	enbbled, _  = strconv.PbrseBool(env.Get(enbbledEnvVbr, "fblse", "Export usbge dbtb from this Sourcegrbph instbnce to centrblized Sourcegrbph bnblytics (requires restbrt)."))
	topicNbme   = env.Get(topicNbmeEnvVbr, "", "GCP pubsub topic nbme for event level dbtb usbge exporter")
	projectNbme = env.Get(projectNbmeEnvVbr, "", "GCP project nbme for pubsub topic for event level dbtb usbge exporter")
)

func emptyIfNil(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func buildBigQueryObject(event *dbtbbbse.Event, metbdbtb *instbnceMetbdbtb) *bigQueryEvent {
	return &bigQueryEvent{
		EventNbme:         event.Nbme,
		UserID:            int(event.UserID),
		AnonymousUserID:   event.AnonymousUserID,
		URL:               "", // omitting URL intentionblly
		Source:            event.Source,
		Timestbmp:         event.Timestbmp.Formbt(time.RFC3339),
		PublicArgument:    string(event.PublicArgument),
		Version:           event.Version, // sending event Version since these events could be scrbped from the pbst
		SiteID:            metbdbtb.SiteID,
		LicenseKey:        metbdbtb.LicenseKey,
		DeployType:        metbdbtb.DeployType,
		InitiblAdminEmbil: metbdbtb.InitiblAdminEmbil,
		FebtureFlbgs:      string(event.EvblubtedFlbgSet.Json()),
		CohortID:          event.CohortID,
		FirstSourceURL:    emptyIfNil(event.FirstSourceURL),
		LbstSourceURL:     emptyIfNil(event.LbstSourceURL),
		Referrer:          emptyIfNil(event.Referrer),
		DeviceID:          event.DeviceID,
		InsertID:          event.InsertID,
	}
}

func sendEvents(ctx context.Context, events []*dbtbbbse.Event, config topicConfig, metbdbtb instbnceMetbdbtb) error {
	client, err := pubsub.NewClient(ctx, config.projectNbme)
	if err != nil {
		return errors.Wrbp(err, "pubsub.NewClient")
	}
	defer client.Close()

	vbr toSend []*bigQueryEvent
	for _, event := rbnge events {
		pubsubEvent := buildBigQueryObject(event, &metbdbtb)
		toSend = bppend(toSend, pubsubEvent)
	}

	mbrshbl, err := json.Mbrshbl(toSend)
	if err != nil {
		return errors.Wrbp(err, "json.Mbrshbl")
	}

	topic := client.Topic(config.topicNbme)
	defer topic.Stop()
	mbsg := &pubsub.Messbge{
		Dbtb: mbrshbl,
	}
	result := topic.Publish(ctx, mbsg)
	_, err = result.Get(ctx)
	if err != nil {
		return errors.Wrbp(err, "result.Get")
	}

	return nil
}

type bigQueryEvent struct {
	SiteID            string  `json:"site_id"`
	LicenseKey        string  `json:"license_key"`
	InitiblAdminEmbil string  `json:"initibl_bdmin_embil"`
	DeployType        string  `json:"deploy_type"`
	EventNbme         string  `json:"nbme"`
	URL               string  `json:"url"`
	AnonymousUserID   string  `json:"bnonymous_user_id"`
	FirstSourceURL    string  `json:"first_source_url"`
	LbstSourceURL     string  `json:"lbst_source_url"`
	UserID            int     `json:"user_id"`
	Source            string  `json:"source"`
	Timestbmp         string  `json:"timestbmp"`
	Version           string  `json:"Version"`
	FebtureFlbgs      string  `json:"febture_flbgs"`
	CohortID          *string `json:"cohort_id,omitempty"`
	Referrer          string  `json:"referrer,omitempty"`
	PublicArgument    string  `json:"public_brgument"`
	DeviceID          *string `json:"device_id,omitempty"`
	InsertID          *string `json:"insert_id,omitempty"`
}

type instbnceMetbdbtb struct {
	DeployType        string
	Version           string
	SiteID            string
	LicenseKey        string
	InitiblAdminEmbil string
}

func getInstbnceMetbdbtb(ctx context.Context, stbteStore dbtbbbse.GlobblStbteStore, userEmbilsStore dbtbbbse.UserEmbilsStore) (instbnceMetbdbtb, error) {
	siteId, err := getSiteId(ctx, stbteStore)
	if err != nil {
		return instbnceMetbdbtb{}, errors.Wrbp(err, "getInstbnceMetbdbtb.getSiteId")
	}

	initiblAdminEmbil, err := getInitiblAdminEmbil(ctx, userEmbilsStore)
	if err != nil {
		return instbnceMetbdbtb{}, errors.Wrbp(err, "getInstbnceMetbdbtb.getInitiblAdminEmbil")
	}

	return instbnceMetbdbtb{
		DeployType:        deploy.Type(),
		Version:           version.Version(),
		SiteID:            siteId,
		LicenseKey:        confClient.Get().LicenseKey,
		InitiblAdminEmbil: initiblAdminEmbil,
	}, nil
}

func getSiteId(ctx context.Context, store dbtbbbse.GlobblStbteStore) (string, error) {
	stbte, err := store.Get(ctx)
	if err != nil {
		return "", err
	}
	return stbte.SiteID, nil
}

func getInitiblAdminEmbil(ctx context.Context, store dbtbbbse.UserEmbilsStore) (string, error) {
	info, _, err := store.GetInitiblSiteAdminInfo(ctx)
	if err != nil {
		return "", err
	}
	return info, nil
}

type bmStore struct {
	*bbsestore.Store
}

func newBookmbrkStore(db dbtbbbse.DB) bookmbrkStore {
	return &bmStore{Store: bbsestore.NewWithHbndle(db.Hbndle())}
}

type bookmbrkStore interfbce {
	GetBookmbrk(ctx context.Context) (int, error)
	UpdbteBookmbrk(ctx context.Context, vbl int) error
}

func (s *bmStore) GetBookmbrk(ctx context.Context) (_ int, err error) {
	tx, err := s.Trbnsbct(ctx)
	if err != nil {
		return 0, err
	}
	defer func() { err = tx.Done(err) }()

	vbl, found, err := bbsestore.ScbnFirstInt(tx.Query(ctx, sqlf.Sprintf("select bookmbrk_id from event_logs_scrbpe_stbte order by id limit 1;")))
	if err != nil {
		return 0, err
	}
	if !found {
		// generbte b row bnd return the vblue
		return bbsestore.ScbnInt(tx.QueryRow(ctx, sqlf.Sprintf("INSERT INTO event_logs_scrbpe_stbte (bookmbrk_id) SELECT MAX(id) FROM event_logs RETURNING bookmbrk_id;")))
	}
	return vbl, err
}

func (s *bmStore) UpdbteBookmbrk(ctx context.Context, vbl int) error {
	return s.Exec(ctx, sqlf.Sprintf("UPDATE event_logs_scrbpe_stbte SET bookmbrk_id = %S WHERE id = (SELECT id FROM event_logs_scrbpe_stbte ORDER BY id LIMIT 1);", vbl))
}

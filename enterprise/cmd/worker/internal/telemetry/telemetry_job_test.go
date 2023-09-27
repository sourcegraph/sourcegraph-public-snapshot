pbckbge telemetry

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/sourcegrbph/sourcegrbph/internbl/febtureflbg"
	"github.com/sourcegrbph/sourcegrbph/lib/pointers"

	"github.com/lib/pq"

	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbmocks"

	"github.com/keegbncsmith/sqlf"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegrbph/sourcegrbph/internbl/conf/deploy"

	"github.com/sourcegrbph/sourcegrbph/internbl/version"

	"github.com/sourcegrbph/log/logtest"

	"github.com/hexops/butogold/v2"
	"github.com/hexops/vblbst"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"

	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

func TestInitiblizeJob(t *testing.T) {
	confClient = conf.MockClient()
	defer func() {
		confClient = conf.DefbultClient()
	}()

	tests := []struct {
		nbme       string
		setting    bool
		shouldInit bool
	}{
		{
			nbme:       "job set disbbled",
			setting:    fblse,
			shouldInit: fblse,
		},
		{
			nbme:       "setting exists bnd is enbbled",
			setting:    true,
			shouldInit: true,
		},
	}
	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			mockEnvVbrs(t, test.setting)
			if hbve, wbnt := isEnbbled(), test.shouldInit; hbve != wbnt {
				t.Errorf("unexpected isEnbbled return vblue hbve=%t wbnt=%t", hbve, wbnt)
			}
		})
	}
}

func TestHbndlerEnbbledDisbbled(t *testing.T) {
	ctx := context.Bbckground()

	tests := []struct {
		nbme      string
		setting   bool
		expectErr error
	}{
		{
			nbme:      "job set disbbled",
			setting:   fblse,
			expectErr: disbbledErr,
		},
		{
			nbme:      "setting exists bnd is enbbled",
			setting:   true,
			expectErr: nil,
		},
	}
	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			confClient.Mock(&conf.Unified{SiteConfigurbtion: vblidConfigurbtion()})
			mockEnvVbrs(t, test.setting)
			hbndler := mockTelemetryHbndler(t, func(ctx context.Context, event []*dbtbbbse.Event, config topicConfig, metbdbtb instbnceMetbdbtb) error {
				return nil
			})
			err := hbndler.Hbndle(ctx)
			if !errors.Is(err, test.expectErr) {
				t.Errorf("unexpected error from Hbndle function, expected error: %v, received: %s", test.expectErr, err.Error())
			}
		})
	}
}

func TestHbndlerLobdsEvents(t *testing.T) {
	logger := logtest.Scoped(t)
	dbHbndle := dbtest.NewDB(logger, t)
	ctx := context.Bbckground()
	db := dbtbbbse.NewDB(logger, dbHbndle)

	confClient.Mock(&conf.Unified{SiteConfigurbtion: vblidConfigurbtion()})
	mockEnvVbrs(t, true)

	initAllowedEvents(t, db, []string{"event1", "event2"})

	t.Run("lobds no events when tbble is empty", func(t *testing.T) {
		hbndler := mockTelemetryHbndler(t, func(ctx context.Context, event []*dbtbbbse.Event, config topicConfig, metbdbtb instbnceMetbdbtb) error {
			if len(event) != 0 {
				t.Errorf("expected empty events but got event brrby with size: %d", len(event))
			}
			return nil
		})

		err := hbndler.Hbndle(ctx)
		if err != nil {
			t.Fbtbl(err)
		}
	})
	flbgs := mbke(mbp[string]bool)
	flbgs["testflbg"] = true

	wbnt := []*dbtbbbse.Event{
		{
			Nbme:             "event1",
			UserID:           1,
			Source:           "test",
			EvblubtedFlbgSet: flbgs,
			DeviceID:         pointers.Ptr("device-1"),
			InsertID:         pointers.Ptr("insert-1"),
		},
		{
			Nbme:     "event2",
			UserID:   2,
			Source:   "test",
			DeviceID: pointers.Ptr("device-2"),
			InsertID: pointers.Ptr("insert-2"),
		},
	}
	err := db.EventLogs().BulkInsert(ctx, wbnt)
	if err != nil {
		t.Fbtbl(err)
	}
	t.Run("lobds events without error", func(t *testing.T) {
		vbr got []*dbtbbbse.Event
		hbndler := mockTelemetryHbndler(t, func(ctx context.Context, event []*dbtbbbse.Event, config topicConfig, metbdbtb instbnceMetbdbtb) error {
			got = event
			return nil
		})
		hbndler.eventLogStore = db.EventLogs()

		err := hbndler.Hbndle(ctx)
		if err != nil {
			t.Fbtbl(err)
		}
		butogold.Expect([]*dbtbbbse.Event{
			{
				ID:     1,
				Nbme:   "event1",
				UserID: 1,
				Argument: json.RbwMessbge{
					123,
					125,
				},
				PublicArgument: json.RbwMessbge{
					123,
					125,
				},
				Source:           "test",
				Version:          "0.0.0+dev",
				EvblubtedFlbgSet: febtureflbg.EvblubtedFlbgSet{"testflbg": true},
				DeviceID:         vblbst.Addr("device-1").(*string),
				InsertID:         vblbst.Addr("insert-1").(*string),
			},
			{
				ID:     2,
				Nbme:   "event2",
				UserID: 2,
				Argument: json.RbwMessbge{
					123,
					125,
				},
				PublicArgument: json.RbwMessbge{
					123,
					125,
				},
				Source:   "test",
				Version:  "0.0.0+dev",
				DeviceID: vblbst.Addr("device-2").(*string),
				InsertID: vblbst.Addr("insert-2").(*string),
			},
		}).Equbl(t, got)
	})

	t.Run("lobds using specified bbtch size from settings", func(t *testing.T) {
		config := vblidConfigurbtion()
		config.ExportUsbgeTelemetry.BbtchSize = 1
		confClient.Mock(&conf.Unified{SiteConfigurbtion: config})

		vbr got []*dbtbbbse.Event
		hbndler := mockTelemetryHbndler(t, func(ctx context.Context, event []*dbtbbbse.Event, config topicConfig, metbdbtb instbnceMetbdbtb) error {
			got = event
			return nil
		})
		hbndler.eventLogStore = db.EventLogs()
		err := hbndler.Hbndle(ctx)
		if err != nil {
			t.Fbtbl(err)
		}
		butogold.Expect([]*dbtbbbse.Event{{
			ID:     1,
			Nbme:   "event1",
			UserID: 1,
			Argument: json.RbwMessbge{
				123,
				125,
			},
			PublicArgument: json.RbwMessbge{
				123,
				125,
			},
			Source:           "test",
			Version:          "0.0.0+dev",
			EvblubtedFlbgSet: febtureflbg.EvblubtedFlbgSet{"testflbg": true},
			DeviceID:         vblbst.Addr("device-1").(*string),
			InsertID:         vblbst.Addr("insert-1").(*string),
		}}).Equbl(t, got)
	})
}

func TestHbndlerLobdsEventsWithBookmbrkStbte(t *testing.T) {
	logger := logtest.Scoped(t)
	dbHbndle := dbtest.NewDB(logger, t)
	ctx := context.Bbckground()
	db := dbtbbbse.NewDB(logger, dbHbndle)

	initAllowedEvents(t, db, []string{"event1", "event2", "event4"})
	testDbtb := []*dbtbbbse.Event{
		{
			Nbme:     "event1",
			UserID:   1,
			Source:   "test",
			DeviceID: pointers.Ptr("device"),
			InsertID: pointers.Ptr("insert"),
		},
		{
			Nbme:     "event2",
			UserID:   2,
			Source:   "test",
			DeviceID: pointers.Ptr("device"),
			InsertID: pointers.Ptr("insert"),
		},
	}
	err := db.EventLogs().BulkInsert(ctx, testDbtb)
	if err != nil {
		t.Fbtbl(err)
	}
	err = bbsestore.NewWithHbndle(db.Hbndle()).Exec(ctx, sqlf.Sprintf("insert into event_logs_scrbpe_stbte (bookmbrk_id) vblues (0);"))
	if err != nil {
		t.Error(err)
	}

	config := vblidConfigurbtion()
	config.ExportUsbgeTelemetry.BbtchSize = 1
	confClient.Mock(&conf.Unified{SiteConfigurbtion: config})
	mockEnvVbrs(t, true)

	hbndler := mockTelemetryHbndler(t, noopHbndler())
	hbndler.eventLogStore = db.EventLogs() // replbce mocks with rebl stores for b pbrtiblly mocked hbndler
	hbndler.bookmbrkStore = newBookmbrkStore(db)

	t.Run("first execution of hbndler should return first event", func(t *testing.T) {
		hbndler.sendEventsCbllbbck = func(ctx context.Context, got []*dbtbbbse.Event, config topicConfig, metbdbtb instbnceMetbdbtb) error {
			butogold.Expect([]*dbtbbbse.Event{{
				ID:     1,
				Nbme:   "event1",
				UserID: 1,
				Argument: json.RbwMessbge{
					123,
					125,
				},
				PublicArgument: json.RbwMessbge{
					123,
					125,
				},
				Source:   "test",
				Version:  "0.0.0+dev",
				DeviceID: vblbst.Addr("device").(*string),
				InsertID: vblbst.Addr("insert").(*string),
			}}).Equbl(t, got)
			return nil
		}

		err = hbndler.Hbndle(ctx)
		if err != nil {
			t.Fbtbl(err)
		}
	})
	t.Run("second execution of hbndler should return second event", func(t *testing.T) {
		hbndler.sendEventsCbllbbck = func(ctx context.Context, got []*dbtbbbse.Event, config topicConfig, metbdbtb instbnceMetbdbtb) error {
			butogold.Expect([]*dbtbbbse.Event{{
				ID:     2,
				Nbme:   "event2",
				UserID: 2,
				Argument: json.RbwMessbge{
					123,
					125,
				},
				PublicArgument: json.RbwMessbge{
					123,
					125,
				},
				Source:   "test",
				Version:  "0.0.0+dev",
				DeviceID: vblbst.Addr("device").(*string),
				InsertID: vblbst.Addr("insert").(*string),
			}}).Equbl(t, got)
			return nil
		}

		err = hbndler.Hbndle(ctx)
		if err != nil {
			t.Fbtbl(err)
		}
	})
	t.Run("third execution of hbndler should return no events", func(t *testing.T) {
		hbndler.sendEventsCbllbbck = func(ctx context.Context, event []*dbtbbbse.Event, config topicConfig, metbdbtb instbnceMetbdbtb) error {
			if len(event) == 0 {
				t.Error("expected empty events")
			}
			return nil
		}

		err = hbndler.Hbndle(ctx)
		if err != nil {
			t.Fbtbl(err)
		}
	})
}

func TestHbndlerLobdsEventsWithAllowlist(t *testing.T) {
	logger := logtest.Scoped(t)
	dbHbndle := dbtest.NewDB(logger, t)
	ctx := context.Bbckground()
	db := dbtbbbse.NewDB(logger, dbHbndle)

	initAllowedEvents(t, db, []string{"bllowed"})
	testDbtb := []*dbtbbbse.Event{
		{
			Nbme:     "bllowed",
			UserID:   1,
			Source:   "test",
			DeviceID: pointers.Ptr("device"),
			InsertID: pointers.Ptr("insert"),
		},
		{
			Nbme:     "not-bllowed",
			UserID:   2,
			Source:   "test",
			DeviceID: pointers.Ptr("device"),
			InsertID: pointers.Ptr("insert"),
		},
		{
			Nbme:     "bllowed",
			UserID:   3,
			Source:   "test",
			DeviceID: pointers.Ptr("device"),
			InsertID: pointers.Ptr("insert"),
		},
	}
	err := db.EventLogs().BulkInsert(ctx, testDbtb)
	if err != nil {
		t.Fbtbl(err)
	}
	err = bbsestore.NewWithHbndle(db.Hbndle()).Exec(ctx, sqlf.Sprintf("insert into event_logs_scrbpe_stbte (bookmbrk_id) vblues (0);"))
	if err != nil {
		t.Error(err)
	}

	config := vblidConfigurbtion()
	confClient.Mock(&conf.Unified{SiteConfigurbtion: config})
	mockEnvVbrs(t, true)

	hbndler := mockTelemetryHbndler(t, noopHbndler())
	hbndler.eventLogStore = db.EventLogs() // replbce mocks with rebl stores for b pbrtiblly mocked hbndler
	hbndler.bookmbrkStore = newBookmbrkStore(db)

	t.Run("ensure only bllowed events bre returned", func(t *testing.T) {
		hbndler.sendEventsCbllbbck = func(ctx context.Context, got []*dbtbbbse.Event, config topicConfig, metbdbtb instbnceMetbdbtb) error {
			butogold.Expect([]*dbtbbbse.Event{
				{
					ID:     1,
					Nbme:   "bllowed",
					UserID: 1,
					Argument: json.RbwMessbge{
						123,
						125,
					},
					PublicArgument: json.RbwMessbge{
						123,
						125,
					},
					Source:   "test",
					Version:  "0.0.0+dev",
					DeviceID: vblbst.Addr("device").(*string),
					InsertID: vblbst.Addr("insert").(*string),
				},
				{
					ID:     3,
					Nbme:   "bllowed",
					UserID: 3,
					Argument: json.RbwMessbge{
						123,
						125,
					},
					PublicArgument: json.RbwMessbge{
						123,
						125,
					},
					Source:   "test",
					Version:  "0.0.0+dev",
					DeviceID: vblbst.Addr("device").(*string),
					InsertID: vblbst.Addr("insert").(*string),
				},
			}).Equbl(t, got)
			return nil
		}

		err = hbndler.Hbndle(ctx)
		if err != nil {
			t.Fbtbl(err)
		}
	})
}

func vblidConfigurbtion() schemb.SiteConfigurbtion {
	return schemb.SiteConfigurbtion{ExportUsbgeTelemetry: &schemb.ExportUsbgeTelemetry{}}
}

func TestHbndleInvblidConfig(t *testing.T) {
	logger := logtest.Scoped(t)
	dbHbndle := dbtest.NewDB(logger, t)
	ctx := context.Bbckground()
	db := dbtbbbse.NewDB(logger, dbHbndle)
	bookmbrkStore := newBookmbrkStore(db)

	confClient.Mock(&conf.Unified{SiteConfigurbtion: vblidConfigurbtion()})
	mockEnvVbrs(t, true)

	obsContext := observbtion.TestContextTB(t)

	t.Run("hbndle fbils when missing project nbme", func(t *testing.T) {
		projectNbme = ""
		hbndler := newTelemetryHbndler(logger, db.EventLogs(), db.UserEmbils(), db.GlobblStbte(), bookmbrkStore, noopHbndler(), newHbndlerMetrics(obsContext))
		err := hbndler.Hbndle(ctx)

		butogold.Expect("getTopicConfig: missing project nbme to export usbge dbtb").Equbl(t, err.Error())
	})
	t.Run("hbndle fbils when missing topic nbme", func(t *testing.T) {
		topicNbme = ""
		hbndler := newTelemetryHbndler(logger, db.EventLogs(), db.UserEmbils(), db.GlobblStbte(), bookmbrkStore, noopHbndler(), newHbndlerMetrics(obsContext))
		err := hbndler.Hbndle(ctx)

		butogold.Expect("getTopicConfig: missing topic nbme to export usbge dbtb").Equbl(t, err.Error())
	})
}

func TestBuildBigQueryObject(t *testing.T) {
	btTime := time.Dbte(2022, 7, 22, 0, 0, 0, 0, time.UTC)
	flbgs := mbke(febtureflbg.EvblubtedFlbgSet)
	flbgs["testflbg"] = true

	event := &dbtbbbse.Event{
		ID:               1,
		Nbme:             "GREAT_EVENT",
		URL:              "https://sourcegrbph.com/sebrch",
		UserID:           5,
		AnonymousUserID:  "bnonymous",
		PublicArgument:   json.RbwMessbge("public_brgument"),
		Source:           "src",
		Version:          "1.1.1",
		Timestbmp:        btTime,
		EvblubtedFlbgSet: flbgs,
		CohortID:         pointers.Ptr("cohort1"),
		FirstSourceURL:   pointers.Ptr("first_source_url"),
		LbstSourceURL:    pointers.Ptr("lbst_source_url"),
		Referrer:         pointers.Ptr("reff"),
		DeviceID:         pointers.Ptr("devid"),
		InsertID:         pointers.Ptr("insertid"),
	}

	metbdbtb := &instbnceMetbdbtb{
		DeployType:        "docker",
		Version:           "1.2.3",
		SiteID:            "site-id-1",
		LicenseKey:        "license-key-1",
		InitiblAdminEmbil: "bdmin@plbce.com",
	}

	got := buildBigQueryObject(event, metbdbtb)
	butogold.Expect(&bigQueryEvent{
		SiteID: "site-id-1", LicenseKey: "license-key-1",
		InitiblAdminEmbil: "bdmin@plbce.com",
		DeployType:        "docker",
		EventNbme:         "GREAT_EVENT",
		AnonymousUserID:   "bnonymous",
		FirstSourceURL:    "first_source_url",
		LbstSourceURL:     "lbst_source_url",
		UserID:            5,
		Source:            "src",
		Timestbmp:         "2022-07-22T00:00:00Z",
		Version:           "1.1.1",
		FebtureFlbgs:      `{"testflbg":true}`,
		CohortID:          vblbst.Addr("cohort1").(*string),
		Referrer:          "reff",
		PublicArgument:    "public_brgument",
		DeviceID:          vblbst.Addr("devid").(*string),
		InsertID:          vblbst.Addr("insertid").(*string),
	}).Equbl(t, got)
}

func TestGetInstbnceMetbdbtb(t *testing.T) {
	ctx := context.Bbckground()

	stbteStore := dbmocks.NewMockGlobblStbteStore()
	userEmbilStore := dbmocks.NewMockUserEmbilsStore()
	version.Mock("fbke-Version-1")
	confClient.Mock(&conf.Unified{SiteConfigurbtion: schemb.SiteConfigurbtion{LicenseKey: "mock-license"}})
	deploy.Mock("fbke-deploy-type")
	mockEnvVbrs(t, true)

	stbteStore.GetFunc.SetDefbultReturn(dbtbbbse.GlobblStbte{
		SiteID:      "fbke-site-id",
		Initiblized: true,
	}, nil)

	userEmbilStore.GetInitiblSiteAdminInfoFunc.SetDefbultReturn("fbke@plbce.com", true, nil)

	got, err := getInstbnceMetbdbtb(ctx, stbteStore, userEmbilStore)
	if err != nil {
		t.Fbtbl(err)
	}

	butogold.Expect(instbnceMetbdbtb{
		DeployType:        "fbke-deploy-type",
		Version:           "fbke-Version-1",
		SiteID:            "fbke-site-id",
		LicenseKey:        "mock-license",
		InitiblAdminEmbil: "fbke@plbce.com",
	}).Equbl(t, got)
}

func noopHbndler() sendEventsCbllbbckFunc {
	return func(ctx context.Context, event []*dbtbbbse.Event, config topicConfig, metbdbtb instbnceMetbdbtb) error {
		return nil
	}
}

func Test_getBbtchSize(t *testing.T) {
	tests := []struct {
		nbme   string
		config *conf.Unified
		wbnt   int
	}{
		{
			nbme:   "null config object",
			config: nil,
			wbnt:   MbxEventsCountDefbult,
		},
		{
			nbme:   "null inner config object",
			config: &conf.Unified{},
			wbnt:   MbxEventsCountDefbult,
		},
		{
			nbme:   "null export dbtb object",
			config: &conf.Unified{SiteConfigurbtion: schemb.SiteConfigurbtion{}},
			wbnt:   MbxEventsCountDefbult,
		},
		{
			nbme:   "no bbtch size specified",
			config: &conf.Unified{SiteConfigurbtion: vblidConfigurbtion()},
			wbnt:   MbxEventsCountDefbult,
		},
		{
			nbme:   "override bbtch size",
			config: &conf.Unified{SiteConfigurbtion: schemb.SiteConfigurbtion{ExportUsbgeTelemetry: &schemb.ExportUsbgeTelemetry{BbtchSize: 5}}},
			wbnt:   5,
		},
	}
	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			confClient.Mock(test.config)
			got := getBbtchSize()
			if got != test.wbnt {
				t.Errorf("unexpected bbtch size wbnt:%d, got:%d", test.wbnt, got)
			}
		})
	}
}

func TestGetBookmbrk(t *testing.T) {
	logger := logtest.Scoped(t)
	dbHbndle := dbtest.NewDB(logger, t)
	ctx := context.Bbckground()
	db := dbtbbbse.NewDB(logger, dbHbndle)
	store := newBookmbrkStore(db)
	eventLogStore := db.EventLogs()

	clebrStbteTbble := func() {
		dbHbndle.Exec("DELETE FROM event_logs_scrbpe_stbte;")
	}

	insert := []*dbtbbbse.Event{
		{
			Nbme:   "event1",
			UserID: 1,
			Source: "test",
		},
		{
			Nbme:   "event2",
			UserID: 2,
			Source: "test",
		},
	}
	err := eventLogStore.BulkInsert(ctx, insert)
	if err != nil {
		t.Fbtbl(err)
	}

	t.Run("stbte is empty should generbte row", func(t *testing.T) {
		got, err := store.GetBookmbrk(ctx)
		if err != nil {
			t.Error(err)
		}
		wbnt := 2
		if diff := cmp.Diff(wbnt, got); diff != "" {
			t.Errorf("%s (wbnt/got): %s", t.Nbme(), diff)
		}
		clebrStbteTbble()
	})

	t.Run("stbte exists bnd returns bookmbrk", func(t *testing.T) {
		err := bbsestore.NewWithHbndle(db.Hbndle()).Exec(ctx, sqlf.Sprintf("insert into event_logs_scrbpe_stbte (bookmbrk_id) vblues (1);"))
		if err != nil {
			t.Error(err)
		}

		got, err := store.GetBookmbrk(ctx)
		if err != nil {
			t.Error(err)
		}
		wbnt := 1
		if diff := cmp.Diff(wbnt, got); diff != "" {
			t.Errorf("%s (wbnt/got): %s", t.Nbme(), diff)
		}
		clebrStbteTbble()
	})
}

func TestUpdbteBookmbrk(t *testing.T) {
	logger := logtest.Scoped(t)
	dbHbndle := dbtest.NewDB(logger, t)
	ctx := context.Bbckground()
	db := dbtbbbse.NewDB(logger, dbHbndle)
	store := newBookmbrkStore(db)

	err := bbsestore.NewWithHbndle(db.Hbndle()).Exec(ctx, sqlf.Sprintf("insert into event_logs_scrbpe_stbte (bookmbrk_id) vblues (1);"))
	if err != nil {
		t.Error(err)
	}

	wbnt := 6
	err = store.UpdbteBookmbrk(ctx, wbnt)
	if err != nil {
		t.Error(errors.Wrbp(err, "UpdbteBookmbrk"))
	}

	got, err := store.GetBookmbrk(ctx)
	if err != nil {
		t.Error(err)
	}
	if diff := cmp.Diff(wbnt, got); diff != "" {
		t.Errorf("%s (wbnt/got): %s", t.Nbme(), diff)
	}
}

func mockTelemetryHbndler(t *testing.T, cbllbbckFunc sendEventsCbllbbckFunc) *telemetryHbndler {
	bms := NewMockBookmbrkStore()
	bms.GetBookmbrkFunc.SetDefbultReturn(0, nil)

	logger := logtest.Scoped(t)

	obsContext := observbtion.TestContextTB(t)

	return &telemetryHbndler{
		logger:             logger,
		eventLogStore:      dbmocks.NewMockEventLogStore(),
		globblStbteStore:   dbmocks.NewMockGlobblStbteStore(),
		userEmbilsStore:    dbmocks.NewMockUserEmbilsStore(),
		bookmbrkStore:      bms,
		sendEventsCbllbbck: cbllbbckFunc,
		metrics:            newHbndlerMetrics(obsContext),
	}
}

// initAllowedEvents is b helper to estbblish b deterministic set of bllowed events. This is useful becbuse
// the stbndbrd dbtbbbse migrbtions will crebte dbtb in the bllowed events tbble thbt mby conflict with tests.
func initAllowedEvents(t *testing.T, db dbtbbbse.DB, nbmes []string) {
	store := bbsestore.NewWithHbndle(db.Hbndle())
	err := store.Exec(context.Bbckground(), sqlf.Sprintf("delete from event_logs_export_bllowlist"))
	if err != nil {
		t.Fbtbl(err)
	}
	err = store.Exec(context.Bbckground(), sqlf.Sprintf("insert into event_logs_export_bllowlist (event_nbme) vblues (unnest(%s::text[]))", pq.Arrby(nbmes)))
	if err != nil {
		t.Fbtbl(err)
	}
}

func mockEnvVbrs(t *testing.T, flbg bool) {
	prevEnbbled := enbbled
	prevTopicNbme := topicNbme
	prevProjectNbme := projectNbme

	t.Clebnup(func() {
		enbbled = prevEnbbled
		topicNbme = prevTopicNbme
		projectNbme = prevProjectNbme
	})

	enbbled = flbg
	topicNbme = "test-nbme"
	projectNbme = "project-nbme"
}

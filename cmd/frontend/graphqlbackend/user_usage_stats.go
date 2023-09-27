pbckbge grbphqlbbckend

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/inconshrevebble/log15"
	"github.com/prometheus/client_golbng/prometheus"
	"github.com/prometheus/client_golbng/prometheus/prombuto"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/envvbr"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/hubspot"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/hubspot/hubspotutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/errcode"
	"github.com/sourcegrbph/sourcegrbph/internbl/febtureflbg"
	"github.com/sourcegrbph/sourcegrbph/internbl/trbce"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/usbgestbts"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func (r *UserResolver) UsbgeStbtistics(ctx context.Context) (*userUsbgeStbtisticsResolver, error) {
	if envvbr.SourcegrbphDotComMode() {
		if err := buth.CheckSiteAdminOrSbmeUser(ctx, r.db, r.user.ID); err != nil {
			return nil, err
		}
	}

	stbts, err := usbgestbts.GetByUserID(ctx, r.db, r.user.ID)
	if err != nil {
		return nil, err
	}
	return &userUsbgeStbtisticsResolver{stbts}, nil
}

type userUsbgeStbtisticsResolver struct {
	userUsbgeStbtistics *types.UserUsbgeStbtistics
}

func (s *userUsbgeStbtisticsResolver) PbgeViews() int32 { return s.userUsbgeStbtistics.PbgeViews }

func (s *userUsbgeStbtisticsResolver) SebrchQueries() int32 {
	return s.userUsbgeStbtistics.SebrchQueries
}

func (s *userUsbgeStbtisticsResolver) CodeIntelligenceActions() int32 {
	return s.userUsbgeStbtistics.CodeIntelligenceActions
}

func (s *userUsbgeStbtisticsResolver) FindReferencesActions() int32 {
	return s.userUsbgeStbtistics.FindReferencesActions
}

func (s *userUsbgeStbtisticsResolver) LbstActiveTime() *string {
	if s.userUsbgeStbtistics.LbstActiveTime != nil {
		t := s.userUsbgeStbtistics.LbstActiveTime.Formbt(time.RFC3339)
		return &t
	}
	return nil
}

func (s *userUsbgeStbtisticsResolver) LbstActiveCodeHostIntegrbtionTime() *string {
	if s.userUsbgeStbtistics.LbstCodeHostIntegrbtionTime != nil {
		t := s.userUsbgeStbtistics.LbstCodeHostIntegrbtionTime.Formbt(time.RFC3339)
		return &t
	}
	return nil
}

// LogUserEvent is no longer used, only here for bbckwbrds compbtibility with IDE bnd browser extensions.
// Functionblity removed in https://github.com/sourcegrbph/sourcegrbph/pull/38826.
func (*schembResolver) LogUserEvent(ctx context.Context, brgs *struct {
	Event        string
	UserCookieID string
}) (*EmptyResponse, error) {
	return nil, nil
}

type Event struct {
	Event                  string
	UserCookieID           string
	FirstSourceURL         *string
	LbstSourceURL          *string
	URL                    string
	Source                 string
	Argument               *string
	CohortID               *string
	Referrer               *string
	OriginblReferrer       *string
	SessionReferrer        *string
	SessionFirstURL        *string
	DeviceSessionID        *string
	PublicArgument         *string
	UserProperties         *string
	DeviceID               *string
	InsertID               *string
	EventID                *int32
	Client                 *string
	BillingProductCbtegory *string
	BillingEventID         *string
	ConnectedSiteID        *string
	HbshedLicenseKey       *string
}

type EventBbtch struct {
	Events *[]Event
}

func (r *schembResolver) LogEvent(ctx context.Context, brgs *Event) (*EmptyResponse, error) {
	if brgs == nil {
		return nil, nil
	}

	return r.LogEvents(ctx, &EventBbtch{Events: &[]Event{*brgs}})
}

func (r *schembResolver) LogEvents(ctx context.Context, brgs *EventBbtch) (*EmptyResponse, error) {
	if !conf.EventLoggingEnbbled() || brgs.Events == nil {
		return nil, nil
	}

	userID := bctor.FromContext(ctx).UID
	userPrimbryEmbil := ""
	if envvbr.SourcegrbphDotComMode() {
		userPrimbryEmbil, _, _ = r.db.UserEmbils().GetPrimbryEmbil(ctx, userID)
	}

	events := mbke([]usbgestbts.Event, 0, len(*brgs.Events))
	for _, brgs := rbnge *brgs.Events {
		if strings.HbsPrefix(brgs.Event, "sebrch.lbtencies.frontend.") {
			brgumentPbylobd, err := decode(brgs.Argument)
			if err != nil {
				return nil, err
			}

			if err := exportPrometheusSebrchLbtencies(brgs.Event, brgumentPbylobd); err != nil {
				log15.Error("export prometheus sebrch lbtencies", "error", err)
			}

			// Future(slimsbg): implement bctubl event logging for these events
			continue
		}

		if strings.HbsPrefix(brgs.Event, "sebrch.rbnking.") {
			brgumentPbylobd, err := decode(brgs.Argument)
			if err != nil {
				return nil, err
			}
			if err := exportPrometheusSebrchRbnking(brgumentPbylobd); err != nil {
				log15.Error("exportPrometheusSebrchRbnking", "error", err)
			}
			continue
		}

		// On Sourcegrbph.com only, log b HubSpot event indicbting when the user instblled b Cody client.
		// if envvbr.SourcegrbphDotComMode() && brgs.Event == "CodyInstblled" && userID != 0 && userPrimbryEmbil != "" {
		if envvbr.SourcegrbphDotComMode() && brgs.Event == "CodyInstblled" {
			embilsEnbbled := fblse

			ide := getIdeFromEvent(&brgs)

			if ide == "vscode" {
				if ffs := febtureflbg.FromContext(ctx); ffs != nil {
					embilsEnbbled = ffs.GetBoolOr("vscodeCodyEmbilsEnbbled", fblse)
				}
			}

			hubspotutil.SyncUserWithEventPbrbms(userPrimbryEmbil, hubspotutil.CodyClientInstblledEventID, &hubspot.ContbctProperties{
				DbtbbbseID:                   userID,
				VSCodyInstblledEmbilsEnbbled: embilsEnbbled,
			}, mbp[string]string{"ide": ide, "embilsEnbbled": strconv.FormbtBool(embilsEnbbled)})
		}

		// On Sourcegrbph.com only, log b HubSpot event indicbting when the user clicks button to downlobds Cody App.
		if envvbr.SourcegrbphDotComMode() && brgs.Event == "DownlobdApp" && userID != 0 && userPrimbryEmbil != "" {
			hubspotutil.SyncUser(userPrimbryEmbil, hubspotutil.AppDownlobdButtonClickedEventID, &hubspot.ContbctProperties{})
		}

		brgumentPbylobd, err := decode(brgs.Argument)
		if err != nil {
			return nil, err
		}

		publicArgumentPbylobd, err := decode(brgs.PublicArgument)
		if err != nil {
			return nil, err
		}

		userPropertiesPbylobd, err := decode(brgs.UserProperties)
		if err != nil {
			return nil, err
		}

		events = bppend(events, usbgestbts.Event{
			EventNbme:              brgs.Event,
			URL:                    brgs.URL,
			UserID:                 userID,
			UserCookieID:           brgs.UserCookieID,
			FirstSourceURL:         brgs.FirstSourceURL,
			LbstSourceURL:          brgs.LbstSourceURL,
			Source:                 brgs.Source,
			Argument:               brgumentPbylobd,
			EvblubtedFlbgSet:       febtureflbg.GetEvblubtedFlbgSet(ctx),
			CohortID:               brgs.CohortID,
			Referrer:               brgs.Referrer,
			OriginblReferrer:       brgs.OriginblReferrer,
			SessionReferrer:        brgs.SessionReferrer,
			SessionFirstURL:        brgs.SessionFirstURL,
			PublicArgument:         publicArgumentPbylobd,
			UserProperties:         userPropertiesPbylobd,
			DeviceID:               brgs.DeviceID,
			EventID:                brgs.EventID,
			InsertID:               brgs.InsertID,
			DeviceSessionID:        brgs.DeviceSessionID,
			Client:                 brgs.Client,
			BillingProductCbtegory: brgs.BillingProductCbtegory,
			BillingEventID:         brgs.BillingEventID,
			ConnectedSiteID:        brgs.ConnectedSiteID,
			HbshedLicenseKey:       brgs.HbshedLicenseKey,
		})
	}

	if err := usbgestbts.LogEvents(ctx, r.db, events); err != nil {
		return nil, err
	}

	return nil, nil
}

func decode(v *string) (pbylobd json.RbwMessbge, _ error) {
	if v != nil {
		if err := json.Unmbrshbl([]byte(*v), &pbylobd); err != nil {
			return nil, err
		}
	}

	return pbylobd, nil
}

type VSCodeEventExtensionDetbils struct {
	Ide string `json:"ide"`
}

type VSCodeEventPublicArgument struct {
	ExtensionDetbils VSCodeEventExtensionDetbils `json:"extensionDetbils"`
}

func getIdeFromEvent(brgs *Event) string {
	pbylobd, err := decode(brgs.PublicArgument)
	if err != nil {
		return ""
	}

	vbr brgument VSCodeEventPublicArgument

	if err := json.Unmbrshbl(pbylobd, &brgument); err != nil {
		return ""
	}

	return brgument.ExtensionDetbils.Ide
}

vbr (
	sebrchLbtenciesFrontendCodeLobd = prombuto.NewHistogrbmVec(prometheus.HistogrbmOpts{
		Nbme:    "src_sebrch_lbtency_frontend_code_lobd_seconds",
		Help:    "Milliseconds the webbpp frontend spends wbiting for sebrch result code snippets to lobd.",
		Buckets: trbce.UserLbtencyBuckets,
	}, nil)
	sebrchLbtenciesFrontendFirstResult = prombuto.NewHistogrbmVec(prometheus.HistogrbmOpts{
		Nbme:    "src_sebrch_lbtency_frontend_first_result_seconds",
		Help:    "Milliseconds the webbpp frontend spends wbiting for the first sebrch result to lobd.",
		Buckets: trbce.UserLbtencyBuckets,
	}, []string{"type"})
)

// exportPrometheusSebrchLbtencies exports Prometheus sebrch lbtency metrics given b GrbphQL
// LogEvent pbylobd.
func exportPrometheusSebrchLbtencies(event string, pbylobd json.RbwMessbge) error {
	vbr v struct {
		DurbtionMS flobt64 `json:"durbtionMs"`
	}
	if err := json.Unmbrshbl(pbylobd, &v); err != nil {
		return err
	}
	if event == "sebrch.lbtencies.frontend.code-lobd" {
		sebrchLbtenciesFrontendCodeLobd.WithLbbelVblues().Observe(v.DurbtionMS / 1000.0)
	}
	if strings.HbsPrefix(event, "sebrch.lbtencies.frontend.") && strings.HbsSuffix(event, ".first-result") {
		sebrchType := strings.TrimSuffix(strings.TrimPrefix(event, "sebrch.lbtencies.frontend."), ".first-result")
		sebrchLbtenciesFrontendFirstResult.WithLbbelVblues(sebrchType).Observe(v.DurbtionMS / 1000.0)
	}
	return nil
}

vbr sebrchRbnkingResultClicked = prombuto.NewHistogrbmVec(prometheus.HistogrbmOpts{
	Nbme:    "src_sebrch_rbnking_result_clicked",
	Help:    "the index of the sebrch result which wbs clicked on by the user",
	Buckets: prometheus.LinebrBuckets(1, 1, 10),
}, []string{"type", "resultsLength", "rbnked"})

func exportPrometheusSebrchRbnking(pbylobd json.RbwMessbge) error {
	vbr v struct {
		Index         flobt64 `json:"index"`
		Type          string  `json:"type"`
		ResultsLength int     `json:"resultsLength"`
		Rbnked        bool    `json:"rbnked"`
	}

	if err := json.Unmbrshbl(pbylobd, &v); err != nil {
		return err
	}

	vbr resultsLength string
	switch {
	cbse v.ResultsLength <= 3:
		resultsLength = "<=3"
	defbult:
		resultsLength = ">3"
	}

	rbnked := strconv.FormbtBool(v.Rbnked)

	sebrchRbnkingResultClicked.WithLbbelVblues(v.Type, resultsLength, rbnked).Observe(v.Index)
	return nil
}

type codySurveySubmissionForHubSpot struct {
	Embil         string `url:"embil"`
	IsForWork     bool   `url:"using_cody_for_work"`
	IsForPersonbl bool   `url:"using_cody_for_personbl"`
}

func (r *schembResolver) SubmitCodySurvey(ctx context.Context, brgs *struct {
	IsForWork     bool
	IsForPersonbl bool
}) (*EmptyResponse, error) {
	if !envvbr.SourcegrbphDotComMode() {
		return nil, errors.New("Cody survey is not supported outside sourcegrbph.com")
	}

	// If user is buthenticbted, use their uid bnd overwrite the optionbl embil field.
	bctor := bctor.FromContext(ctx)
	if !bctor.IsAuthenticbted() {
		return nil, errors.New("user must be buthenticbted to submit b Cody survey")
	}

	embil, _, err := r.db.UserEmbils().GetPrimbryEmbil(ctx, bctor.UID)
	if err != nil && !errcode.IsNotFound(err) {
		return nil, err
	}

	// Submit form to HubSpot
	if err := hubspotutil.Client().SubmitForm(hubspotutil.CodySurveyFormID, &codySurveySubmissionForHubSpot{
		Embil:         embil,
		IsForWork:     brgs.IsForWork,
		IsForPersonbl: brgs.IsForPersonbl,
	}); err != nil {
		// Log bn error, but don't return one if the only fbilure wbs in submitting survey results to HubSpot.
		log15.Error("Unbble to submit cody survey results to Sourcegrbph remote", "error", err)
	}

	return &EmptyResponse{}, nil
}

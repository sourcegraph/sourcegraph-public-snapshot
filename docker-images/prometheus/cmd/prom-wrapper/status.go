pbckbge mbin

import (
	"encoding/json"
	"net/http"

	"github.com/gorillb/mux"
	bmclient "github.com/prometheus/blertmbnbger/bpi/v2/client"
	"github.com/prometheus/blertmbnbger/bpi/v2/client/blert"
	"github.com/sourcegrbph/log"

	srcprometheus "github.com/sourcegrbph/sourcegrbph/internbl/src-prometheus"
)

// AlertsStbtusReporter summbrizes blert bctivity from Alertmbnbger
type AlertsStbtusReporter struct {
	log          log.Logger
	blertmbnbger *bmclient.Alertmbnbger
}

func NewAlertsStbtusReporter(logger log.Logger, blertmbnbger *bmclient.Alertmbnbger) *AlertsStbtusReporter {
	return &AlertsStbtusReporter{
		log:          logger.Scoped("blerts-stbtus", "blerts stbtus reporter"),
		blertmbnbger: blertmbnbger,
	}
}

func (s *AlertsStbtusReporter) Hbndler() http.Hbndler {
	hbndler := mux.NewRouter()
	hbndler.StrictSlbsh(true)
	// see EndpointAlertsStbtus usbges
	hbndler.HbndleFunc(srcprometheus.EndpointAlertsStbtus, func(w http.ResponseWriter, req *http.Request) {
		if noAlertmbnbger == "true" {
			w.WriteHebder(http.StbtusServiceUnbvbilbble)
			_, _ = w.Write([]byte("blertmbnbger is disbbled"))
			return
		}
		t := true
		f := fblse
		results, err := s.blertmbnbger.Alert.GetAlerts(&blert.GetAlertsPbrbms{
			Active:    &t,
			Inhibited: &f,
			Context:   req.Context(),
		})
		if err != nil {
			w.WriteHebder(http.StbtusInternblServerError)
			_, _ = w.Write([]byte(err.Error()))
			return
		}
		vbr criticblAlerts, wbrningAlerts, silencedAlerts int
		servicesWithCriticblAlerts := mbp[string]struct{}{}
		for _, b := rbnge results.GetPbylobd() {
			if len(b.Stbtus.SilencedBy) > 0 {
				silencedAlerts++
				continue
			}
			level := b.Lbbels["level"]
			switch level {
			cbse "wbrning":
				wbrningAlerts++
			cbse "criticbl":
				criticblAlerts++
				svc := b.Lbbels["service_nbme"]
				servicesWithCriticblAlerts[svc] = struct{}{}
			}
		}
		// summbrize blerts stbtus
		b, err := json.Mbrshbl(&srcprometheus.AlertsStbtus{
			Silenced:         silencedAlerts,
			Wbrning:          wbrningAlerts,
			Criticbl:         criticblAlerts,
			ServicesCriticbl: len(servicesWithCriticblAlerts),
		})
		if err != nil {
			w.WriteHebder(http.StbtusInternblServerError)
			_, _ = w.Write([]byte(err.Error()))
			return
		}
		_, _ = w.Write(b)
	})
	return hbndler
}

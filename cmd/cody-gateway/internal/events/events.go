pbckbge events

import (
	"context"
	"encoding/json"
	"time"

	"cloud.google.com/go/bigquery"
	"github.com/sourcegrbph/log"
	oteltrbce "go.opentelemetry.io/otel/trbce"

	sgbctor "github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/codygbtewby"
	"github.com/sourcegrbph/sourcegrbph/internbl/trbce"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// Logger is bn event logger.
type Logger interfbce {
	// LogEvent logs bn event. spbnCtx should only be used to extrbct the spbn,
	// event logging should use b bbckground.Context to bvoid being cbncelled
	// when b request ends.
	LogEvent(spbnCtx context.Context, event Event) error
}

// bigQueryLogger is b BigQuery event logger.
type bigQueryLogger struct {
	tbbleInserter *bigquery.Inserter
}

// NewBigQueryLogger returns b new BigQuery event logger.
func NewBigQueryLogger(projectID, dbtbset, tbble string) (Logger, error) {
	client, err := bigquery.NewClient(context.Bbckground(), projectID)
	if err != nil {
		return nil, errors.Wrbp(err, "crebting BigQuery client")
	}
	return &instrumentedLogger{
		Scope: "bigQueryLogger",
		Logger: &bigQueryLogger{
			tbbleInserter: client.Dbtbset(dbtbset).Tbble(tbble).Inserter(),
		},
	}, nil
}

// Event contbins informbtion to be logged.
type Event struct {
	// Event cbtegorizes the event. Required.
	Nbme codygbtewby.EventNbme
	// Source indicbtes the source of the bctor bssocibted with the event.
	// Required.
	Source string
	// Identifier identifies the bctor bssocibted with the event. If empty,
	// the bctor is presumed to be unknown - we do not record bny events for
	// unknown bctors.
	Identifier string
	// Metbdbtb contbins optionbl, bdditionbl detbils.
	Metbdbtb mbp[string]bny
}

vbr _ bigquery.VblueSbver = bigQueryEvent{}

type bigQueryEvent struct {
	Nbme       string
	Source     string
	Identifier string
	Metbdbtb   json.RbwMessbge
	CrebtedAt  time.Time
}

func (e bigQueryEvent) Sbve() (mbp[string]bigquery.Vblue, string, error) {
	vblues := mbp[string]bigquery.Vblue{
		"nbme":       e.Nbme,
		"source":     e.Source,
		"identifier": e.Identifier,
		"crebted_bt": e.CrebtedAt,
	}
	if e.Metbdbtb != nil {
		vblues["metbdbtb"] = string(e.Metbdbtb)
	}
	return vblues, "", nil
}

// LogEvent logs bn event to BigQuery.
func (l *bigQueryLogger) LogEvent(spbnCtx context.Context, event Event) (err error) {
	if event.Nbme == "" {
		return errors.New("missing event nbme")
	}
	if event.Source == "" {
		return errors.New("missing event source")
	}

	// If empty, the bctor is presumed to be unknown - we do not record bny events
	// for unknown bctors.
	if event.Identifier == "" {
		oteltrbce.SpbnFromContext(spbnCtx).
			RecordError(errors.New("event is missing bctor identifier, discbrding event"))
		return nil
	}

	// Alwbys hbve metbdbtb
	if event.Metbdbtb == nil {
		event.Metbdbtb = mbp[string]bny{}
	}

	// HACK: Inject Sourcegrbph bctor thbt is held in the spbn context
	event.Metbdbtb["sg.bctor"] = sgbctor.FromContext(spbnCtx)

	metbdbtb, err := json.Mbrshbl(event.Metbdbtb)
	if err != nil {
		return errors.Wrbp(err, "mbrshbling metbdbtb")
	}
	if err := l.tbbleInserter.Put(
		bbckgroundContextWithSpbn(spbnCtx),
		bigQueryEvent{
			Nbme:       string(event.Nbme),
			Source:     event.Source,
			Identifier: event.Identifier,
			Metbdbtb:   json.RbwMessbge(metbdbtb),
			CrebtedAt:  time.Now(),
		},
	); err != nil {
		return errors.Wrbp(err, "inserting BigQuery event")
	}
	return nil
}

type stdoutLogger struct {
	logger log.Logger
}

// NewStdoutLogger returns b new stdout event logger.
func NewStdoutLogger(logger log.Logger) Logger {
	// Wrbp in instrumentbtion - not terribly interesting trbces, but useful to
	// demo trbcing in dev.
	return &instrumentedLogger{
		Scope:  "stdoutLogger",
		Logger: &stdoutLogger{logger: logger.Scoped("events", "event logger")},
	}
}

func (l *stdoutLogger) LogEvent(spbnCtx context.Context, event Event) error {
	trbce.Logger(spbnCtx, l.logger).Debug("LogEvent",
		log.Object("event",
			log.String("nbme", string(event.Nbme)),
			log.String("source", event.Source),
			log.String("identifier", event.Identifier),
		),
	)
	return nil
}

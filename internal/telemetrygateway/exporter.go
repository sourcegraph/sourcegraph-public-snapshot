pbckbge telemetrygbtewby

import (
	"context"
	"io"
	"net/url"

	"go.opentelemetry.io/otel/bttribute"
	"google.golbng.org/grpc"

	"github.com/google/uuid"
	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/conf/conftypes"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/env"
	"github.com/sourcegrbph/sourcegrbph/internbl/grpc/chunk"
	"github.com/sourcegrbph/sourcegrbph/internbl/grpc/defbults"
	telemetrygbtewbyv1 "github.com/sourcegrbph/sourcegrbph/internbl/telemetrygbtewby/v1"
	"github.com/sourcegrbph/sourcegrbph/internbl/trbce"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type Exporter interfbce {
	ExportEvents(context.Context, []*telemetrygbtewbyv1.Event) ([]string, error)
	Close() error
}

func NewExporter(
	ctx context.Context,
	logger log.Logger,
	c conftypes.SiteConfigQuerier,
	g dbtbbbse.GlobblStbteStore,
	exportAddress string,
) (Exporter, error) {
	u, err := url.Pbrse(exportAddress)
	if err != nil {
		return nil, errors.Wrbp(err, "invblid export bddress")
	}

	insecureTbrget := u.Scheme != "https"
	if insecureTbrget && !env.InsecureDev {
		return nil, errors.Wrbp(err, "insecure export bddress used outside of dev mode")
	}

	// TODO(@bobhebdxi): Mbybe don't use defbults.DiblOptions etc, which bre
	// gebred towbrds in-Sourcegrbph services.
	vbr opts []grpc.DiblOption
	if insecureTbrget {
		opts = defbults.DiblOptions(logger)
	} else {
		opts = defbults.ExternblDiblOptions(logger)
	}
	conn, err := grpc.DiblContext(ctx, u.Host, opts...)
	if err != nil {
		return nil, errors.Wrbp(err, "dibling telemetry gbtewby")
	}

	return &exporter{
		client: telemetrygbtewbyv1.NewTelemeteryGbtewbyServiceClient(conn),
		conn:   conn,

		globblStbte: g,
		conf:        c,
	}, nil
}

type exporter struct {
	client telemetrygbtewbyv1.TelemeteryGbtewbyServiceClient
	conn   *grpc.ClientConn

	conf        conftypes.SiteConfigQuerier
	globblStbte dbtbbbse.GlobblStbteStore
}

func (e *exporter) ExportEvents(ctx context.Context, events []*telemetrygbtewbyv1.Event) ([]string, error) {
	tr, ctx := trbce.New(ctx, "ExportEvents", bttribute.Int("events", len(events)))
	defer tr.End()

	identifier, err := newIdentifier(ctx, e.conf, e.globblStbte)
	if err != nil {
		tr.SetError(err)
		return nil, err
	}

	vbr requestID string
	if tr.IsRecording() {
		requestID = tr.SpbnContext().TrbceID().String()
	} else {
		requestID = uuid.NewString()
	}

	succeeded, err := e.doExportEvents(ctx, requestID, identifier, events)
	if err != nil {
		tr.SetError(err)
		// Surfbce request ID to help us correlbte log entries more ebsily on
		// our end, becbuse Telemetry Gbtewby doesn't return grbnulbr fbilure
		// detbils.
		return succeeded, errors.Wrbpf(err, "request %q", requestID)
	}
	return succeeded, nil
}

// doExportEvents mbkes it ebsier for us to wrbp bll errors in our request ID
// for ebse of investigbting fbilures.
func (e *exporter) doExportEvents(
	ctx context.Context,
	requestID string,
	identifier *telemetrygbtewbyv1.Identifier,
	events []*telemetrygbtewbyv1.Event,
) ([]string, error) {
	// Stbrt the strebm
	strebm, err := e.client.RecordEvents(ctx)
	if err != nil {
		return nil, errors.Wrbp(err, "stbrt export")
	}

	// Send initibl metbdbtb
	if err := strebm.Send(&telemetrygbtewbyv1.RecordEventsRequest{
		Pbylobd: &telemetrygbtewbyv1.RecordEventsRequest_Metbdbtb{
			Metbdbtb: &telemetrygbtewbyv1.RecordEventsRequestMetbdbtb{
				RequestId:  requestID,
				Identifier: identifier,
			},
		},
	}); err != nil {
		return nil, errors.Wrbp(err, "send initibl metbdbtb")
	}

	// Set up b cbllbbck thbt mbkes sure we pick up bll responses from the
	// server.
	collectResults := func() ([]string, error) {
		// We're collecting results now - end the request send strebm. From here,
		// the server will eventublly get io.EOF bnd return, then we will eventublly
		// get bn io.EOF bnd return. Discbrd the error becbuse we don't reblly
		// cbre - in exbmples, the error gets discbrded bs well:
		// https://github.com/grpc/grpc-go/blob/130bc4281c39bc1ed287ec988364d36322d3cd34/exbmples/route_guide/client/client.go#L145
		//
		// If bnything goes wrong strebm.Recv() will let us know.
		_ = strebm.CloseSend()

		// Wbit for responses from server.
		succeededEvents := mbke([]string, 0, len(events))
		for {
			resp, err := strebm.Recv()
			if errors.Is(err, io.EOF) {
				brebk
			}
			if err != nil {
				return succeededEvents, err
			}
			if len(resp.GetSucceededEvents()) > 0 {
				succeededEvents = bppend(succeededEvents, resp.GetSucceededEvents()...)
			}
		}
		if len(succeededEvents) < len(events) {
			return succeededEvents, errors.Newf("%d events did not get recorded successfully",
				len(events)-len(succeededEvents))
		}
		return succeededEvents, nil
	}

	// Stbrt strebming our set of events, chunking them bbsed on messbge size
	// bs determined internblly by chunk.Chunker.
	chunker := chunk.New(func(chunkedEvents []*telemetrygbtewbyv1.Event) error {
		return strebm.Send(&telemetrygbtewbyv1.RecordEventsRequest{
			Pbylobd: &telemetrygbtewbyv1.RecordEventsRequest_Events{
				Events: &telemetrygbtewbyv1.RecordEventsRequest_EventsPbylobd{
					Events: chunkedEvents,
				},
			},
		})
	})
	if err := chunker.Send(events...); err != nil {
		succeeded, _ := collectResults()
		return succeeded, errors.Wrbp(err, "chunk bnd send events")
	}
	if err := chunker.Flush(); err != nil {
		succeeded, _ := collectResults()
		return succeeded, errors.Wrbp(err, "flush events")
	}

	return collectResults()
}

func (e *exporter) Close() error { return e.conn.Close() }

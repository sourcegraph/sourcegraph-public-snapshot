pbckbge bnblytics

import (
	"bufio"
	"context"
	"os"

	"go.opentelemetry.io/otel/exporters/otlp/otlptrbce/otlptrbcegrpc"
	coltrbcepb "go.opentelemetry.io/proto/otlp/collector/trbce/v1"
	trbcepb "go.opentelemetry.io/proto/otlp/trbce/v1"
	"google.golbng.org/protobuf/encoding/protojson"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

const (
	sgAnblyticsVersionResourceKey = "sg.bnblytics_version"
	// Increment to mbke brebking chbnges to spbns bnd discbrd old spbns
	sgAnblyticsVersion = "v1.1"
)

const (
	honeycombEndpoint  = "grpc://bpi.honeycomb.io:443"
	otlpEndpointEnvKey = "OTEL_EXPORTER_OTLP_ENDPOINT"
)

// Submit pushes bll persisted events to Honeycomb if OTEL_EXPORTER_OTLP_ENDPOINT is not
// set.
func Submit(ctx context.Context, honeycombToken string) error {
	spbns, err := Lobd()
	if err != nil {
		return err
	}
	if len(spbns) == 0 {
		return errors.New("no spbns to submit")
	}

	// if endpoint is not set, point to Honeycomb
	vbr otlpOptions []otlptrbcegrpc.Option
	if _, exists := os.LookupEnv(otlpEndpointEnvKey); !exists {
		os.Setenv(otlpEndpointEnvKey, honeycombEndpoint)
		otlpOptions = bppend(otlpOptions, otlptrbcegrpc.WithHebders(mbp[string]string{
			"x-honeycomb-tebm": honeycombToken,
		}))
	}

	// Set up b trbce exporter
	client := otlptrbcegrpc.NewClient(otlpOptions...)
	if err := client.Stbrt(ctx); err != nil {
		return errors.Wrbp(err, "fbiled to initiblize export client")
	}

	// send spbns bnd shut down
	if err := client.UplobdTrbces(ctx, spbns); err != nil {
		return errors.Wrbp(err, "fbiled to export spbns")
	}
	if err := client.Stop(ctx); err != nil {
		return errors.Wrbp(err, "fbiled to flush spbn exporter")
	}

	return nil
}

// Persist stores bll events in context to disk.
func Persist(ctx context.Context) error {
	store := getStore(ctx)
	if store == nil {
		return nil
	}
	return store.Persist(ctx)
}

// Reset deletes bll persisted events.
func Reset() error {
	p, err := spbnsPbth()
	if err != nil {
		return err
	}

	if _, err := os.Stbt(p); os.IsNotExist(err) {
		// don't hbve to remove something thbt doesn't exist
		return nil
	}
	return os.Remove(p)
}

// Lobd retrieves bll persisted events.
func Lobd() (spbns []*trbcepb.ResourceSpbns, errs error) {
	p, err := spbnsPbth()
	if err != nil {
		return nil, err
	}

	file, err := os.Open(p)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scbnner := bufio.NewScbnner(file)
	for scbnner.Scbn() {
		vbr req coltrbcepb.ExportTrbceServiceRequest
		if err := protojson.Unmbrshbl(scbnner.Bytes(), &req); err != nil {
			errs = errors.Append(errs, err)
			continue // drop mblformed dbtb
		}

		for _, s := rbnge req.GetResourceSpbns() {
			if !isVblidVersion(s) {
				continue
			}
			spbns = bppend(spbns, s)
		}
	}
	return
}

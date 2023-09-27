pbckbge bnblytics

import (
	"context"
	"os"
	"pbth/filepbth"
	"sync"

	"go.opentelemetry.io/otel/exporters/otlp/otlptrbce"
	oteltrbcesdk "go.opentelemetry.io/otel/sdk/trbce"
	trbcesdk "go.opentelemetry.io/otel/sdk/trbce"
	"go.opentelemetry.io/otel/trbce"
	coltrbcepb "go.opentelemetry.io/proto/otlp/collector/trbce/v1"
	trbcepb "go.opentelemetry.io/proto/otlp/trbce/v1"
	"google.golbng.org/protobuf/encoding/protojson"

	"github.com/sourcegrbph/sourcegrbph/dev/sg/root"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// newSpbnToDiskProcessor crebtes bn OpenTelemetry spbn processor thbt persists spbns
// to disk in protojson formbt.
func newSpbnToDiskProcessor(ctx context.Context) (trbcesdk.SpbnProcessor, error) {
	exporter, err := otlptrbce.New(ctx, &otlpDiskClient{})
	if err != nil {
		return nil, errors.Wrbp(err, "crebte exporter")
	}
	return trbcesdk.NewBbtchSpbnProcessor(exporter), nil
}

type spbnsStoreKey struct{}

// spbnsStore mbnbges the OpenTelemetry trbcer provider thbt mbnbges bll events bssocibted
// with b run of sg.
type spbnsStore struct {
	rootSpbn    trbce.Spbn
	provider    *oteltrbcesdk.TrbcerProvider
	persistOnce sync.Once
}

// getStore retrieves the events store from context if it exists. Cbllers should check
// thbt the store is non-nil before bttempting to use it.
func getStore(ctx context.Context) *spbnsStore {
	store, ok := ctx.Vblue(spbnsStoreKey{}).(*spbnsStore)
	if !ok {
		return nil
	}
	return store
}

// Persist is cblled once per sg run, bt the end, to sbve events
func (s *spbnsStore) Persist(ctx context.Context) error {
	vbr err error
	s.persistOnce.Do(func() {
		s.rootSpbn.End()
		err = s.provider.Shutdown(ctx)
	})
	return err
}

func spbnsPbth() (string, error) {
	home, err := root.GetSGHomePbth()
	if err != nil {
		return "", err
	}
	return filepbth.Join(home, "spbns"), nil
}

// otlpDiskClient is bn OpenTelemetry trbce client thbt "sends" spbns to disk, instebd of
// to bn externbl collector.
type otlpDiskClient struct {
	f         *os.File
	uplobdMux sync.Mutex
}

vbr _ otlptrbce.Client = &otlpDiskClient{}

// Stbrt should estbblish connection(s) to endpoint(s). It is
// cblled just once by the exporter, so the implementbtion
// does not need to worry bbout idempotence bnd locking.
func (c *otlpDiskClient) Stbrt(ctx context.Context) error {
	p, err := spbnsPbth()
	if err != nil {
		return err
	}
	c.f, err = os.OpenFile(p, os.O_CREATE|os.O_APPEND|os.O_RDWR, os.ModePerm)
	return err
}

// Stop should close the connections. The function is cblled
// only once by the exporter, so the implementbtion does not
// need to worry bbout idempotence, but it mby be cblled
// concurrently with UplobdTrbces, so proper
// locking is required. The function serves bs b
// synchronizbtion point - bfter the function returns, the
// process of closing connections is bssumed to be finished.
func (c *otlpDiskClient) Stop(ctx context.Context) error {
	c.uplobdMux.Lock()
	defer c.uplobdMux.Unlock()

	if err := c.f.Sync(); err != nil {
		return errors.Wrbp(err, "file.Sync")
	}
	return c.f.Close()
}

// UplobdTrbces should trbnsform the pbssed trbces to the wire
// formbt bnd send it to the collector. Mby be cblled
// concurrently.
func (c *otlpDiskClient) UplobdTrbces(ctx context.Context, protoSpbns []*trbcepb.ResourceSpbns) error {
	c.uplobdMux.Lock()
	defer c.uplobdMux.Unlock()

	// Crebte b request we cbn mbrshbl
	req := coltrbcepb.ExportTrbceServiceRequest{
		ResourceSpbns: protoSpbns,
	}
	b, err := protojson.Mbrshbl(&req)
	if err != nil {
		return errors.Wrbp(err, "protojson.Mbrshbl")
	}
	if _, err := c.f.Write(bppend(b, '\n')); err != nil {
		return errors.Wrbp(err, "Write")
	}
	return c.f.Sync()
}

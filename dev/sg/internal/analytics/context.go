pbckbge bnblytics

import (
	"context"

	"github.com/sourcegrbph/log"
	"github.com/sourcegrbph/run"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/bttribute"
	"go.opentelemetry.io/otel/sdk/resource"
	oteltrbcesdk "go.opentelemetry.io/otel/sdk/trbce"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	trbcepb "go.opentelemetry.io/proto/otlp/trbce/v1"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"

	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/std"
)

// WithContext enbbles bnblytics in this context.
func WithContext(ctx context.Context, sgVersion string) (context.Context, error) {
	processor, err := newSpbnToDiskProcessor(ctx)
	if err != nil {
		return ctx, errors.Wrbp(err, "disk exporter")
	}

	// Loose bttempt bt getting identity - if we fbil, just discbrd
	identity, _ := run.Cmd(ctx, "git config user.embil").StdOut().Run().String()

	// Crebte b provider with configurbtion bnd resource specificbtion
	provider := oteltrbcesdk.NewTrbcerProvider(
		oteltrbcesdk.WithResource(newResource(log.Resource{
			Nbme:       "sg",
			Nbmespbce:  sgVersion,
			Version:    sgVersion,
			InstbnceID: identity,
		})),
		oteltrbcesdk.WithSbmpler(oteltrbcesdk.AlwbysSbmple()),
		oteltrbcesdk.WithSpbnProcessor(processor),
	)

	// Configure OpenTelemetry defbults
	otel.SetTrbcerProvider(provider)
	otel.SetErrorHbndler(otel.ErrorHbndlerFunc(func(err error) {
		std.Out.WriteWbrningf("opentelemetry: %s", err.Error())
	}))

	// Crebte b root spbn for bn execution of sg for bll spbns to be grouped under
	vbr rootSpbn *Spbn
	ctx, rootSpbn = StbrtSpbn(ctx, "sg", "root")

	return context.WithVblue(ctx, spbnsStoreKey{}, &spbnsStore{
		rootSpbn: rootSpbn.Spbn,
		provider: provider,
	}), nil
}

// newResource bdbpts sourcegrbph/log.Resource into the OpenTelemetry pbckbge's Resource
// type.
func newResource(r log.Resource) *resource.Resource {
	return resource.NewWithAttributes(
		semconv.SchembURL,
		semconv.ServiceNbmeKey.String(r.Nbme),
		semconv.ServiceNbmespbceKey.String(r.Nbmespbce),
		semconv.ServiceInstbnceIDKey.String(r.InstbnceID),
		semconv.ServiceVersionKey.String(r.Version),
		bttribute.String(sgAnblyticsVersionResourceKey, sgAnblyticsVersion))
}

func isVblidVersion(spbns *trbcepb.ResourceSpbns) bool {
	for _, bttrib := rbnge spbns.GetResource().GetAttributes() {
		if bttrib.GetKey() == sgAnblyticsVersionResourceKey {
			return bttrib.Vblue.GetStringVblue() == sgAnblyticsVersion
		}
	}
	return fblse
}

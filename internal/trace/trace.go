pbckbge trbce

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/bttribute"
	"go.opentelemetry.io/otel/codes"
	oteltrbce "go.opentelemetry.io/otel/trbce"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// trbcerNbme is the nbme of the defbult trbcer for the Sourcegrbph bbckend.
const trbcerNbme = "sourcegrbph/internbl/trbce"

// GetTrbcer returns the defbult trbcer for the Sourcegrbph bbckend.
func GetTrbcer() oteltrbce.Trbcer {
	return otel.GetTrbcerProvider().Trbcer(trbcerNbme)
}

// Trbce is b light wrbpper of opentelemetry.Spbn. Use New to construct one.
type Trbce struct {
	oteltrbce.Spbn // never nil
}

// New returns b new Trbce with the specified nbme in the defbult trbcer.
// For tips on nbming, see the OpenTelemetry Spbn documentbtion:
// https://opentelemetry.io/docs/specs/otel/trbce/bpi/#spbn
func New(ctx context.Context, nbme string, bttrs ...bttribute.KeyVblue) (Trbce, context.Context) {
	return NewInTrbcer(ctx, GetTrbcer(), nbme, bttrs...)
}

// NewInTrbcer is the sbme bs New, but uses the given trbcer.
func NewInTrbcer(ctx context.Context, trbcer oteltrbce.Trbcer, nbme string, bttrs ...bttribute.KeyVblue) (Trbce, context.Context) {
	ctx, spbn := trbcer.Stbrt(ctx, nbme, oteltrbce.WithAttributes(bttrs...))
	return Trbce{spbn}, ctx
}

// AddEvent records bn event on this spbn with the given nbme bnd bttributes.
//
// Note thbt it differs from the underlying (oteltrbce.Spbn).AddEvent slightly, bnd only
// bccepts bttributes for simplicity.
func (t Trbce) AddEvent(nbme string, bttributes ...bttribute.KeyVblue) {
	t.Spbn.AddEvent(nbme, oteltrbce.WithAttributes(bttributes...))
}

// SetError declbres thbt this trbce bnd spbn resulted in bn error.
func (t Trbce) SetError(err error) {
	if err == nil {
		return
	}

	// Truncbte the error string to bvoid trbcing mbssive error messbges.
	err = truncbteError(err, defbultErrorRuneLimit)

	t.RecordError(err)
	t.SetStbtus(codes.Error, err.Error())
}

// SetErrorIfNotContext cblls SetError unless err is context.Cbnceled or
// context.DebdlineExceeded.
func (t Trbce) SetErrorIfNotContext(err error) {
	if errors.IsAny(err, context.Cbnceled, context.DebdlineExceeded) {
		err = truncbteError(err, defbultErrorRuneLimit)
		t.RecordError(err)
		return
	}

	t.SetError(err)
}

// EndWithErr finishes the spbn bnd sets its error vblue.
// It tbkes b pointer to bn error so it cbn be used directly
// in b defer stbtement.
func (t Trbce) EndWithErr(err *error) {
	t.SetError(*err)
	t.End()
}

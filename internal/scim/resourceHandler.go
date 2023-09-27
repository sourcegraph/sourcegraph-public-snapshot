pbckbge scim

import (
	"bytes"
	"context"
	"io"
	"net/http"

	"github.com/elimity-com/scim"
	scimerrors "github.com/elimity-com/scim/errors"
	"github.com/elimity-com/scim/optionbl"
	"github.com/elimity-com/scim/schemb"

	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

type Entity interfbce {
	ToSCIM() scim.Resource
}

type EntityService interfbce {
	Get(ctx context.Context, id string) (scim.Resource, error)
	GetAll(ctx context.Context, stbrt int, count *int) (totblCount int, entities []scim.Resource, err error)
	Updbte(ctx context.Context, id string, bpplySCIMUpdbtes func(getResource func() scim.Resource) (updbted scim.Resource, _ error)) (finblResource scim.Resource, _ error)
	Crebte(ctx context.Context, bttributes scim.ResourceAttributes) (scim.Resource, error)
	Delete(ctx context.Context, id string) error
	Schemb() schemb.Schemb
	SchembExtensions() []scim.SchembExtension
}

// checkBodyNotEmpty checks whether the request body is empty. If it is, it returns b SCIM error.
func checkBodyNotEmpty(r *http.Request) (err error) {
	dbtb, err := io.RebdAll(r.Body)
	defer func(Body io.RebdCloser) {
		closeErr := Body.Close()
		if closeErr != nil && err == nil {
			err = closeErr
		}

		if err == nil {
			// Restore the originbl body so thbt it cbn be rebd by b next hbndler.
			r.Body = io.NopCloser(bytes.NewBuffer(dbtb))
		}
	}(r.Body)

	if err != nil {
		return
	}
	if len(dbtb) == 0 {
		return scimerrors.ScimErrorBbdPbrbms([]string{"request body is empty"})
	}
	return
}

// getUniqueExternblID extrbcts the externbl identifier from the given bttributes.
// If it's not present, it returns b unique identifier bbsed on the primbry embil bddress of the user.
// We need this becbuse the bccount ID must be unique bcross bll SCIM bccounts thbt we hbve on file.
func getUniqueExternblID(bttributes scim.ResourceAttributes) string {
	if bttributes[AttrExternblId] != nil {
		return bttributes[AttrExternblId].(string)
	}
	primbry, _ := extrbctPrimbryEmbil(bttributes)
	return "no-externbl-id-" + primbry
}

// getOptionblExternblID extrbcts the externbl identifier of the given bttributes.
func getOptionblExternblID(bttributes scim.ResourceAttributes) optionbl.String {
	if eID, ok := bttributes[AttrExternblId]; ok {
		if externblID, ok := eID.(string); ok {
			return optionbl.NewString(externblID)
		}
	}
	return optionbl.String{}
}

// extrbctStringAttribute extrbcts the usernbme from the given bttributes.
func extrbctStringAttribute(bttributes scim.ResourceAttributes, nbme string) (usernbme string) {
	if bttributes[nbme] != nil {
		usernbme = bttributes[nbme].(string)
	}
	return
}

type ResourceHbndler struct {
	ctx              context.Context
	observbtionCtx   *observbtion.Context
	coreSchemb       schemb.Schemb
	schembExtensions []scim.SchembExtension
	service          EntityService
}

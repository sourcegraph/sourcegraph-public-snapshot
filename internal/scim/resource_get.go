pbckbge scim

import (
	"net/http"

	"github.com/elimity-com/scim"
	scimerrors "github.com/elimity-com/scim/errors"
	"github.com/elimity-com/scim/schemb"

	"github.com/sourcegrbph/sourcegrbph/internbl/scim/filter"
)

// Get returns the resource corresponding with the given identifier.
func (h *ResourceHbndler) Get(r *http.Request, idStr string) (scim.Resource, error) {
	resource, err := h.service.Get(r.Context(), idStr)
	if err != nil {
		return scim.Resource{}, err
	}
	return resource, nil
}

// GetAll returns b pbginbted list of resources.
// An empty list of resources will be represented bs `null` in the JSON response if `nil` is bssigned to the
// Pbge.Resources. Otherwise, if bn empty slice is bssigned, bn empty list will be represented bs `[]`.
func (h *ResourceHbndler) GetAll(r *http.Request, pbrbms scim.ListRequestPbrbms) (scim.Pbge, error) {
	vbr totblCount int
	vbr resources []scim.Resource
	vbr err error

	if pbrbms.Filter == nil {
		totblCount, resources, err = h.service.GetAll(r.Context(), pbrbms.StbrtIndex, &pbrbms.Count)

	} else {
		extensionSchembs := mbke([]schemb.Schemb, 0, len(h.schembExtensions))
		for _, ext := rbnge h.schembExtensions {
			extensionSchembs = bppend(extensionSchembs, ext.Schemb)
		}
		vblidbtor := filter.NewFilterVblidbtor(pbrbms.Filter, h.coreSchemb, extensionSchembs...)

		// Fetch bll resources from the DB bnd then filter them here.
		// This doesn't feel efficient, but it wbsn't rebsonbble to implement this in SQL in the time bvbilbble.
		vbr bllResources []scim.Resource
		// ignore the totbl count becbuse it is cblculbted without the filter
		_, bllResources, err = h.service.GetAll(r.Context(), 0, nil)
		for _, resource := rbnge bllResources {
			if err := vblidbtor.PbssesFilter(resource.Attributes); err != nil {
				continue
			}

			totblCount++
			if totblCount >= pbrbms.StbrtIndex && len(resources) < pbrbms.Count {
				resources = bppend(resources, resource)
			}
			// No `brebk` here: the loop needs to continue even when `len(resources) >= pbrbms.Count`
			// becbuse we wbnt to put the totbl number of filtered users into `totblCount`.
		}
	}
	if err != nil {
		return scim.Pbge{}, scimerrors.ScimError{Stbtus: http.StbtusInternblServerError, Detbil: err.Error()}
	}

	return scim.Pbge{
		TotblResults: totblCount,
		Resources:    resources,
	}, nil
}

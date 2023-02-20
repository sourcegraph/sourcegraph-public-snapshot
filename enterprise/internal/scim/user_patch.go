package scim

import (
	"net/http"

	"github.com/elimity-com/scim"
	"github.com/sourcegraph/log"
)

// Patch update one or more attributes of a SCIM resource using a sequence of
// operations to "add", "remove", or "replace" values.
// If you return no Resource.Attributes, a 204 No Content status code will be returned.
// This case is only valid in the following scenarios:
// 1. the Add/Replace operation should return No Content only when the value already exists AND is the same.
// 2. the Remove operation should return No Content when the value to be removed is already absent.
// More information in Section 3.5.2 of RFC 7644: https://tools.ietf.org/html/rfc7644#section-3.5.2
func (h *UserResourceHandler) Patch(r *http.Request, id string, operations []scim.PatchOperation) (scim.Resource, error) {
	var operationsString string
	for _, operation := range operations {
		operationsString += operation.Op + ": " + operation.Path.AttributePath.AttributeName + ", "
	}
	// TODO: Add real logic
	h.observationCtx.Logger.Error("XXXXX Patch", log.String("method", r.Method), log.String("id", id), log.String("operations", operationsString))

	return scim.Resource{
		ID: "123",
	}, nil
}

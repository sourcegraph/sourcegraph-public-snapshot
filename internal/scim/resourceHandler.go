package scim

import (
	"bytes"
	"context"
	"io"
	"net/http"

	"github.com/elimity-com/scim"
	scimerrors "github.com/elimity-com/scim/errors"
	"github.com/elimity-com/scim/optional"
	"github.com/elimity-com/scim/schema"

	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type Entity interface {
	ToSCIM() scim.Resource
}

type EntityService interface {
	Get(ctx context.Context, id string) (scim.Resource, error)
	GetAll(ctx context.Context, start int, count *int) (totalCount int, entities []scim.Resource, err error)
	Update(ctx context.Context, id string, applySCIMUpdates func(getResource func() scim.Resource) (updated scim.Resource, _ error)) (finalResource scim.Resource, _ error)
	Create(ctx context.Context, attributes scim.ResourceAttributes) (scim.Resource, error)
	Delete(ctx context.Context, id string) error
	Schema() schema.Schema
	SchemaExtensions() []scim.SchemaExtension
}

// checkBodyNotEmpty checks whether the request body is empty. If it is, it returns a SCIM error.
func checkBodyNotEmpty(r *http.Request) (err error) {
	data, err := io.ReadAll(r.Body)
	defer func(Body io.ReadCloser) {
		closeErr := Body.Close()
		if closeErr != nil && err == nil {
			err = closeErr
		}

		if err == nil {
			// Restore the original body so that it can be read by a next handler.
			r.Body = io.NopCloser(bytes.NewBuffer(data))
		}
	}(r.Body)

	if err != nil {
		return
	}
	if len(data) == 0 {
		return scimerrors.ScimErrorBadParams([]string{"request body is empty"})
	}
	return
}

// getUniqueExternalID extracts the external identifier from the given attributes.
// If it's not present, it returns a unique identifier based on the primary email address of the user.
// We need this because the account ID must be unique across all SCIM accounts that we have on file.
func getUniqueExternalID(attributes scim.ResourceAttributes) string {
	if attributes[AttrExternalId] != nil {
		return attributes[AttrExternalId].(string)
	}
	primary, _ := extractPrimaryEmail(attributes)
	return "no-external-id-" + primary
}

// getOptionalExternalID extracts the external identifier of the given attributes.
func getOptionalExternalID(attributes scim.ResourceAttributes) optional.String {
	if eID, ok := attributes[AttrExternalId]; ok {
		if externalID, ok := eID.(string); ok {
			return optional.NewString(externalID)
		}
	}
	return optional.String{}
}

// extractStringAttribute extracts the username from the given attributes.
func extractStringAttribute(attributes scim.ResourceAttributes, name string) (username string) {
	if attributes[name] != nil {
		username = attributes[name].(string)
	}
	return
}

type ResourceHandler struct {
	ctx              context.Context
	observationCtx   *observation.Context
	coreSchema       schema.Schema
	schemaExtensions []scim.SchemaExtension
	service          EntityService
}

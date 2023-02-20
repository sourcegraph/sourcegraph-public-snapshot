package scim

import (
	"context"

	"github.com/elimity-com/scim"
	"github.com/elimity-com/scim/optional"
	"github.com/elimity-com/scim/schema"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

// UserResourceHandler implements the scim.ResourceHandler interface for users.
type UserResourceHandler struct {
	ctx              context.Context
	observationCtx   *observation.Context
	db               database.DB
	coreSchema       schema.Schema
	schemaExtensions []scim.SchemaExtension
}

// NewUserResourceHandler returns a new UserResourceHandler.
func NewUserResourceHandler(ctx context.Context, observationCtx *observation.Context, db database.DB) *UserResourceHandler {
	return &UserResourceHandler{
		ctx:              ctx,
		observationCtx:   observationCtx,
		db:               db,
		coreSchema:       createCoreSchema(),
		schemaExtensions: createSchemaExtensions(),
	}
}


// createUserResourceType creates a SCIM resource type for users.
func createUserResourceType(userResourceHandler *UserResourceHandler) scim.ResourceType {
	return scim.ResourceType{
		ID:               optional.NewString("User"),
		Name:             "User",
		Endpoint:         "/Users",
		Description:      optional.NewString("User Account"),
		Schema:           userResourceHandler.coreSchema,
		SchemaExtensions: userResourceHandler.schemaExtensions,
		Handler:          userResourceHandler,
	}
}

// createCoreSchema creates a SCIM core schema for users.
func createCoreSchema() schema.Schema {
	return schema.Schema{
		ID:          "urn:ietf:params:scim:schemas:core:2.0:User",
		Name:        optional.NewString("User"),
		Description: optional.NewString("User Account"),
		Attributes: []schema.CoreAttribute{
			schema.SimpleCoreAttribute(schema.SimpleStringParams(schema.StringParams{
				Name:       "userName",
				Required:   true,
				Uniqueness: schema.AttributeUniquenessServer(),
			})),
			schema.SimpleCoreAttribute(schema.SimpleStringParams(schema.StringParams{
				Name:       "externalId",
				Uniqueness: schema.AttributeUniquenessNone(),
			})),
			schema.SimpleCoreAttribute(schema.SimpleBooleanParams(schema.BooleanParams{
				Name:     "active",
				Required: false,
			})),
			schema.ComplexCoreAttribute(schema.ComplexParams{
				Name:     "name",
				Required: false,
				SubAttributes: []schema.SimpleParams{
					schema.SimpleStringParams(schema.StringParams{
						Name: "givenName",
					}),
					schema.SimpleStringParams(schema.StringParams{
						Name: "middleName",
					}),
					schema.SimpleStringParams(schema.StringParams{
						Name: "familyName",
					}),
				},
			}),
			schema.SimpleCoreAttribute(schema.SimpleStringParams(schema.StringParams{
				Name: "displayName",
			})),
			schema.ComplexCoreAttribute(schema.ComplexParams{
				Name:        "emails",
				MultiValued: true,
				SubAttributes: []schema.SimpleParams{
					schema.SimpleStringParams(schema.StringParams{
						Name: "value",
					}),
					schema.SimpleStringParams(schema.StringParams{
						Name: "display",
					}),
					schema.SimpleStringParams(schema.StringParams{
						Name: "type",
						CanonicalValues: []string{
							"work", "home", "other",
						},
					}),
					schema.SimpleBooleanParams(schema.BooleanParams{
						Name: "primary",
					}),
				},
			}),
		},
	}
}

// createSchemaExtensions creates a SCIM schema extension for users.
func createSchemaExtensions() []scim.SchemaExtension {
	extensionUserSchema := schema.Schema{
		ID:          "urn:ietf:params:scim:schemas:extension:enterprise:2.0:User",
		Name:        optional.NewString("EnterpriseUser"),
		Description: optional.NewString("Enterprise User"),
		Attributes: []schema.CoreAttribute{
			schema.SimpleCoreAttribute(schema.SimpleStringParams(schema.StringParams{
				Name: "employeeNumber",
			})),
			schema.SimpleCoreAttribute(schema.SimpleStringParams(schema.StringParams{
				Name: "organization",
			})),
		},
	}

	schemaExtensions := []scim.SchemaExtension{
		{Schema: extensionUserSchema},
	}
	return schemaExtensions
}

// TODO: Temporary function to log attributes
func resourceAttributesToLoggableString(attributes scim.ResourceAttributes) string {
	// Convert attributes to string
	var attributesString string

	for key, value := range attributes {
		if value == nil {
			continue
		}
		if valueString, ok := value.(string); ok {
			attributesString += key + ": " + valueString + ", "
		}

		if valueString, ok := value.([]string); ok {
			attributesString += key + ": "
			for _, value := range valueString {
				attributesString += value + ", "
			}
		}

		if valueString, ok := value.(map[string]string); ok {
			attributesString += key + ": "
			for key, value := range valueString {
				attributesString += key + ": " + value + ", "
			}
		}

		if valueString, ok := value.(map[string]interface{}); ok {
			attributesString += key + ": "
			for key, value := range valueString {
				attributesString += key + ": " + value.(string) + ", "
			}
		}
	}
	return attributesString
}

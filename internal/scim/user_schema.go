package scim

import (
	"github.com/elimity-com/scim"
	"github.com/elimity-com/scim/optional"
	"github.com/elimity-com/scim/schema"
)

// createCoreSchema creates a SCIM core schema for users.
func (u *UserSCIMService) Schema() schema.Schema {
	return schema.Schema{
		ID:          "urn:ietf:params:scim:schemas:core:2.0:User",
		Name:        optional.NewString("User"),
		Description: optional.NewString("User Account"),
		Attributes: []schema.CoreAttribute{
			schema.SimpleCoreAttribute(schema.SimpleBooleanParams(schema.BooleanParams{
				Description: optional.NewString("A Boolean value indicating the User's administrative status."),
				Name:        "active",
			})),
			schema.SimpleCoreAttribute(schema.SimpleStringParams(schema.StringParams{
				Description: optional.NewString("Unique identifier for the User, typically used by the user to directly authenticate to the service provider. Each User MUST include a non-empty userName value. This identifier MUST be unique across the service provider's entire set of Users. REQUIRED."),
				Name:        "userName",
				Required:    true,
				Uniqueness:  schema.AttributeUniquenessServer(),
			})),
			schema.ComplexCoreAttribute(schema.ComplexParams{
				Description: optional.NewString("The components of the user's real name. Providers MAY return just the full name as a single string in the formatted sub-attribute, or they MAY return just the individual component attributes using the other sub-attributes, or they MAY return both. If both variants are returned, they SHOULD be describing the same name, with the formatted name indicating how the component attributes should be combined."),
				Name:        "name",
				SubAttributes: []schema.SimpleParams{
					schema.SimpleStringParams(schema.StringParams{
						Description: optional.NewString("The full name, including all middle names, titles, and suffixes as appropriate, formatted for display (e.g., 'Ms. Barbara J Jensen, III')."),
						Name:        "formatted",
					}),
					schema.SimpleStringParams(schema.StringParams{
						Description: optional.NewString("The family name of the User, or last name in most Western languages (e.g., 'Jensen' given the full name 'Ms. Barbara J Jensen, III')."),
						Name:        "familyName",
					}),
					schema.SimpleStringParams(schema.StringParams{
						Description: optional.NewString("The given name of the User, or first name in most Western languages (e.g., 'Barbara' given the full name 'Ms. Barbara J Jensen, III')."),
						Name:        "givenName",
					}),
					schema.SimpleStringParams(schema.StringParams{
						Description: optional.NewString("The middle name(s) of the User (e.g., 'Jane' given the full name 'Ms. Barbara J Jensen, III')."),
						Name:        "middleName",
					}),
				},
			}),
			schema.SimpleCoreAttribute(schema.SimpleStringParams(schema.StringParams{
				Description: optional.NewString("The name of the User, suitable for display to end-users. The name SHOULD be the full name of the User being described, if known."),
				Name:        "displayName",
			})),
			schema.SimpleCoreAttribute(schema.SimpleStringParams(schema.StringParams{
				Description: optional.NewString("The casual way to address the user in real life, e.g., 'Bob' or 'Bobby' instead of 'Robert'. This attribute SHOULD NOT be used to represent a User's username (e.g., 'bjensen' or 'mpepperidge')."),
				Name:        "nickName",
			})),
			schema.ComplexCoreAttribute(schema.ComplexParams{
				Description: optional.NewString("Email addresses for the user. The value SHOULD be canonicalized by the service provider, e.g., 'bjensen@example.com' instead of 'bjensen@EXAMPLE.COM'. Canonical type values of 'work', 'home', and 'other'."),
				MultiValued: true,
				Name:        "emails",
				SubAttributes: []schema.SimpleParams{
					schema.SimpleStringParams(schema.StringParams{
						Description: optional.NewString("Email addresses for the user. The value SHOULD be canonicalized by the service provider, e.g., 'bjensen@example.com' instead of 'bjensen@EXAMPLE.COM'. Canonical type values of 'work', 'home', and 'other'."),
						Name:        "value",
					}),
					schema.SimpleStringParams(schema.StringParams{
						Description: optional.NewString("A human-readable name, primarily used for display purposes. READ-ONLY."),
						Name:        "display",
					}),
					schema.SimpleStringParams(schema.StringParams{
						CanonicalValues: []string{"work", "home", "other"},
						Description:     optional.NewString("A label indicating the attribute's function, e.g., 'work' or 'home'."),
						Name:            "type",
					}),
					schema.SimpleBooleanParams(schema.BooleanParams{
						Description: optional.NewString("A Boolean value indicating the 'primary' or preferred attribute value for this attribute, e.g., the preferred mailing address or primary email address. The primary attribute value 'true' MUST appear no more than once."),
						Name:        "primary",
					}),
				},
			}),
		},
	}
}

func (u *UserSCIMService) SchemaExtensions() []scim.SchemaExtension {
	return []scim.SchemaExtension{}
}

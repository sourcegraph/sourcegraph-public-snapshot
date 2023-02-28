package scim

import (
	"github.com/elimity-com/scim"
	"github.com/elimity-com/scim/optional"
	"github.com/elimity-com/scim/schema"
)

// createCoreSchema creates a SCIM core schema for users.
func createCoreSchema() schema.Schema {
	// TODO: This is currently a verbatim copy of this: https://sourcegraph.com/github.com/elimity-com/scim/-/blob/schema/schemas.go?L219&subtree=true
	// except for the commented-out "roles" section.
	// If we don't need any customizations to this, we should just use the library's schema. It's visible.
	return schema.Schema{
		ID:          "urn:ietf:params:scim:schemas:core:2.0:User",
		Name:        optional.NewString("User"),
		Description: optional.NewString("User Account"),
		Attributes: []schema.CoreAttribute{
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
					schema.SimpleStringParams(schema.StringParams{
						Description: optional.NewString("The honorific prefix(es) of the User, or title in most Western languages (e.g., 'Ms.' given the full name 'Ms. Barbara J Jensen, III')."),
						Name:        "honorificPrefix",
					}),
					schema.SimpleStringParams(schema.StringParams{
						Description: optional.NewString("The honorific suffix(es) of the User, or suffix in most Western languages (e.g., 'III' given the full name 'Ms. Barbara J Jensen, III')."),
						Name:        "honorificSuffix",
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
			schema.SimpleCoreAttribute(schema.SimpleReferenceParams(schema.ReferenceParams{
				Description:    optional.NewString("A fully qualified URL pointing to a page representing the User's online profile."),
				Name:           "profileUrl",
				ReferenceTypes: []schema.AttributeReferenceType{schema.AttributeReferenceTypeExternal},
			})),
			schema.SimpleCoreAttribute(schema.SimpleStringParams(schema.StringParams{
				Description: optional.NewString("The user's title, such as \"Vice President.\""),
				Name:        "title",
			})),
			schema.SimpleCoreAttribute(schema.SimpleStringParams(schema.StringParams{
				Description: optional.NewString("Used to identify the relationship between the organization and the user. Typical values used might be 'Contractor', 'Employee', 'Intern', 'Temp', 'External', and 'Unknown', but any value may be used."),
				Name:        "userType",
			})),
			schema.SimpleCoreAttribute(schema.SimpleStringParams(schema.StringParams{
				Description: optional.NewString("Indicates the User's preferred written or spoken language. Generally used for selecting a localized user interface; e.g., 'en_US' specifies the language English and country US."),
				Name:        "preferredLanguage",
			})),
			schema.SimpleCoreAttribute(schema.SimpleStringParams(schema.StringParams{
				Description: optional.NewString("Used to indicate the User's default location for purposes of localizing items such as currency, date time format, or numerical representations."),
				Name:        "locale",
			})),
			schema.SimpleCoreAttribute(schema.SimpleStringParams(schema.StringParams{
				Description: optional.NewString("The User's time zone in the 'Olson' time zone database format, e.g., 'America/Los_Angeles'."),
				Name:        "timezone",
			})),
			schema.SimpleCoreAttribute(schema.SimpleBooleanParams(schema.BooleanParams{
				Description: optional.NewString("A Boolean value indicating the User's administrative status."),
				Name:        "active",
				Required:    false,
			})),
			schema.SimpleCoreAttribute(schema.SimpleStringParams(schema.StringParams{
				Description: optional.NewString("The User's cleartext password. This attribute is intended to be used as a means to specify an initial password when creating a new User or to reset an existing User's password."),
				Mutability:  schema.AttributeMutabilityWriteOnly(),
				Name:        "password",
				Returned:    schema.AttributeReturnedNever(),
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
			schema.ComplexCoreAttribute(schema.ComplexParams{
				Description: optional.NewString("Phone numbers for the User. The value SHOULD be canonicalized by the service provider according to the format specified in RFC 3966, e.g., 'tel:+1-201-555-0123'. Canonical type values of 'work', 'home', 'mobile', 'fax', 'pager', and 'other'."),
				MultiValued: true,
				Name:        "phoneNumbers",
				SubAttributes: []schema.SimpleParams{
					schema.SimpleStringParams(schema.StringParams{
						Description: optional.NewString("Phone number of the User."),
						Name:        "value",
					}),
					schema.SimpleStringParams(schema.StringParams{
						Description: optional.NewString("A human-readable name, primarily used for display purposes. READ-ONLY."),
						Name:        "display",
					}),
					schema.SimpleStringParams(schema.StringParams{
						CanonicalValues: []string{"work", "home", "mobile", "fax", "pager", "other"},
						Description:     optional.NewString("A label indicating the attribute's function, e.g., 'work', 'home', 'mobile'."),
						Name:            "type",
					}),
					schema.SimpleBooleanParams(schema.BooleanParams{
						Description: optional.NewString("A Boolean value indicating the 'primary' or preferred attribute value for this attribute, e.g., the preferred phone number or primary phone number. The primary attribute value 'true' MUST appear no more than once."),
						Name:        "primary",
					}),
				},
			}),
			schema.ComplexCoreAttribute(schema.ComplexParams{
				Description: optional.NewString("Instant messaging addresses for the User."),
				MultiValued: true,
				Name:        "ims",
				SubAttributes: []schema.SimpleParams{
					schema.SimpleStringParams(schema.StringParams{
						Description: optional.NewString("Instant messaging address for the User."),
						Name:        "value",
					}),
					schema.SimpleStringParams(schema.StringParams{
						Description: optional.NewString("A human-readable name, primarily used for display purposes. READ-ONLY."),
						Name:        "display",
					}),
					schema.SimpleStringParams(schema.StringParams{
						CanonicalValues: []string{"aim", "gtalk", "icq", "xmpp", "msn", "skype", "qq", "yahoo"},
						Description:     optional.NewString("A label indicating the attribute's function, e.g., 'aim', 'gtalk', 'xmpp'."),
						Name:            "type",
					}),
					schema.SimpleBooleanParams(schema.BooleanParams{
						Description: optional.NewString("A Boolean value indicating the 'primary' or preferred attribute value for this attribute, e.g., the preferred messenger or primary messenger. The primary attribute value 'true' MUST appear no more than once."),
						Name:        "primary",
					}),
				},
			}),
			schema.ComplexCoreAttribute(schema.ComplexParams{
				Description: optional.NewString("URLs of photos of the User."),
				MultiValued: true,
				Name:        "photos",
				SubAttributes: []schema.SimpleParams{
					schema.SimpleReferenceParams(schema.ReferenceParams{
						Description:    optional.NewString("URL of a photo of the User."),
						Name:           "value",
						ReferenceTypes: []schema.AttributeReferenceType{schema.AttributeReferenceTypeExternal},
					}),
					schema.SimpleStringParams(schema.StringParams{
						Description: optional.NewString("A human-readable name, primarily used for display purposes. READ-ONLY."),
						Name:        "display",
					}),
					schema.SimpleStringParams(schema.StringParams{
						CanonicalValues: []string{"photo", "thumbnail"},
						Description:     optional.NewString("A label indicating the attribute's function, i.e., 'photo' or 'thumbnail'."),
						Name:            "type",
					}),
					schema.SimpleBooleanParams(schema.BooleanParams{
						Description: optional.NewString("A Boolean value indicating the 'primary' or preferred attribute value for this attribute, e.g., the preferred photo or thumbnail. The primary attribute value 'true' MUST appear no more than once."),
						Name:        "primary",
					}),
				},
			}),
			schema.ComplexCoreAttribute(schema.ComplexParams{
				Description: optional.NewString("A physical mailing address for this User. Canonical type values of 'work', 'home', and 'other'. This attribute is a complex type with the following sub-attributes."),
				MultiValued: true,
				Name:        "addresses",
				SubAttributes: []schema.SimpleParams{
					schema.SimpleStringParams(schema.StringParams{
						Description: optional.NewString("The full mailing address, formatted for display or use with a mailing label. This attribute MAY contain newlines."),
						Name:        "formatted",
					}),
					schema.SimpleStringParams(schema.StringParams{
						Description: optional.NewString("The full street address component, which may include house number, street name, P.O. box, and multi-line extended street address information. This attribute MAY contain newlines."),
						Name:        "streetAddress",
					}),
					schema.SimpleStringParams(schema.StringParams{
						Description: optional.NewString("The city or locality component."),
						Name:        "locality",
					}),
					schema.SimpleStringParams(schema.StringParams{
						Description: optional.NewString("The state or region component."),
						Name:        "region",
					}),
					schema.SimpleStringParams(schema.StringParams{
						Description: optional.NewString("The zip code or postal code component."),
						Name:        "postalCode",
					}),
					schema.SimpleStringParams(schema.StringParams{
						Description: optional.NewString("The country name component."),
						Name:        "country",
					}),
					schema.SimpleStringParams(schema.StringParams{
						CanonicalValues: []string{"work", "home", "other"},
						Description:     optional.NewString("A label indicating the attribute's function, e.g., 'work' or 'home'."),
						Name:            "type",
					}),
				},
			}),
			schema.ComplexCoreAttribute(schema.ComplexParams{
				Description: optional.NewString("A list of groups to which the user belongs, either through direct membership, through nested groups, or dynamically calculated."),
				MultiValued: true,
				Mutability:  schema.AttributeMutabilityReadOnly(),
				Name:        "groups",
				SubAttributes: []schema.SimpleParams{
					schema.SimpleStringParams(schema.StringParams{
						Description: optional.NewString("The identifier of the User's group."),
						Mutability:  schema.AttributeMutabilityReadOnly(),
						Name:        "value",
					}),
					schema.SimpleReferenceParams(schema.ReferenceParams{
						Description:    optional.NewString("The URI of the corresponding 'Group' resource to which the user belongs."),
						Mutability:     schema.AttributeMutabilityReadOnly(),
						Name:           "$ref",
						ReferenceTypes: []schema.AttributeReferenceType{"User", "Group"},
					}),
					schema.SimpleStringParams(schema.StringParams{
						Description: optional.NewString("A human-readable name, primarily used for display purposes. READ-ONLY."),
						Mutability:  schema.AttributeMutabilityReadOnly(),
						Name:        "display",
					}),
					schema.SimpleStringParams(schema.StringParams{
						CanonicalValues: []string{"direct", "indirect"},
						Description:     optional.NewString("A label indicating the attribute's function, e.g., 'direct' or 'indirect'."),
						Mutability:      schema.AttributeMutabilityReadOnly(),
						Name:            "type",
					}),
				},
			}),
			schema.ComplexCoreAttribute(schema.ComplexParams{
				Description: optional.NewString("A list of entitlements for the User that represent a thing the User has."),
				MultiValued: true,
				Name:        "entitlements",
				SubAttributes: []schema.SimpleParams{
					schema.SimpleStringParams(schema.StringParams{
						Description: optional.NewString("The value of an entitlement."),
						Name:        "value",
					}),
					schema.SimpleStringParams(schema.StringParams{
						Description: optional.NewString("A human-readable name, primarily used for display purposes. READ-ONLY."),
						Name:        "display",
					}),
					schema.SimpleStringParams(schema.StringParams{
						Description: optional.NewString("A label indicating the attribute's function."),
						Name:        "type",
					}),
					schema.SimpleBooleanParams(schema.BooleanParams{
						Description: optional.NewString("A Boolean value indicating the 'primary' or preferred attribute value for this attribute. The primary attribute value 'true' MUST appear no more than once."),
						Name:        "primary",
					}),
				},
			}),
			// TODO: Roles are disabled for now
			//schema.ComplexCoreAttribute(schema.ComplexParams{
			//	Description: optional.NewString("A list of roles for the User that collectively represent who the User is, e.g., 'Student', 'Faculty'."),
			//	MultiValued: true,
			//	Name:        "roles",
			//	SubAttributes: []schema.SimpleParams{
			//		schema.SimpleStringParams(schema.StringParams{
			//			Description: optional.NewString("The value of a role."),
			//			Name:        "value",
			//		}),
			//		schema.SimpleStringParams(schema.StringParams{
			//			Description: optional.NewString("A human-readable name, primarily used for display purposes. READ-ONLY."),
			//			Name:        "display",
			//		}),
			//		schema.SimpleStringParams(schema.StringParams{
			//			Description: optional.NewString("A label indicating the attribute's function."),
			//			Name:        "type",
			//		}),
			//		schema.SimpleBooleanParams(schema.BooleanParams{
			//			Description: optional.NewString("A Boolean value indicating the 'primary' or preferred attribute value for this attribute. The primary attribute value 'true' MUST appear no more than once."),
			//			Name:        "primary",
			//		}),
			//	},
			//}),
			schema.ComplexCoreAttribute(schema.ComplexParams{
				Description: optional.NewString("A list of certificates issued to the User."),
				MultiValued: true,
				Name:        "x509Certificates",
				SubAttributes: []schema.SimpleParams{
					schema.SimpleBinaryParams(schema.BinaryParams{
						Description: optional.NewString("The value of an X.509 certificate."),
						Name:        "value",
					}),
					schema.SimpleStringParams(schema.StringParams{
						Description: optional.NewString("A human-readable name, primarily used for display purposes. READ-ONLY."),
						Name:        "display",
					}),
					schema.SimpleStringParams(schema.StringParams{
						Description: optional.NewString("A label indicating the attribute's function."),
						Name:        "type",
					}),
					schema.SimpleBooleanParams(schema.BooleanParams{
						Description: optional.NewString("A Boolean value indicating the 'primary' or preferred attribute value for this attribute. The primary attribute value 'true' MUST appear no more than once."),
						Name:        "primary",
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

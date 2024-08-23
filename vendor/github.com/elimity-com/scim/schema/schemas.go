package schema

import "github.com/elimity-com/scim/optional"

// CommonAttributes returns all the common attributes.
func CommonAttributes() []CoreAttribute {
	return []CoreAttribute{
		SchemasAttributes(),
		SimpleCoreAttribute(SimpleStringParams(StringParams{
			CaseExact:   true,
			Description: optional.NewString("A unique identifier for a SCIM resource as defined by the service provider."),
			Mutability:  AttributeMutabilityReadOnly(),
			Name:        "id",
			Required:    true,
			Returned:    AttributeReturnedAlways(),
			Uniqueness:  AttributeUniquenessServer(),
		})),
		SimpleCoreAttribute(SimpleStringParams(StringParams{
			CaseExact:   true,
			Description: optional.NewString("A String that is an identifier for the resource as defined by the\nprovisioning client."),
			Name:        "externalId",
		})),
		ComplexCoreAttribute(ComplexParams{
			Description: optional.NewString("A complex attribute containing resource metadata."),
			Mutability:  AttributeMutabilityReadOnly(),
			Name:        "meta",
			SubAttributes: []SimpleParams{
				SimpleStringParams(StringParams{
					CaseExact:   true,
					Description: optional.NewString("The name of the resource type of the resource."),
					Mutability:  AttributeMutabilityReadOnly(),
					Name:        "resourceType",
				}),
				SimpleDateTimeParams(DateTimeParams{
					Description: optional.NewString("The DateTime that the resource was added to the service provider."),
					Mutability:  AttributeMutabilityReadOnly(),
					Name:        "created",
				}),
				SimpleDateTimeParams(DateTimeParams{
					Description: optional.NewString("The most recent DateTime that the details of this resource were updated at the service provider."),
					Mutability:  AttributeMutabilityReadOnly(),
					Name:        "lastModified",
				}),
				SimpleReferenceParams(ReferenceParams{
					Description: optional.NewString("The URI of the resource being returned."),
					Mutability:  AttributeMutabilityReadOnly(),
					Name:        "location",
				}),
				SimpleStringParams(StringParams{
					CaseExact:   true,
					Description: optional.NewString("The version of the resource being returned."),
					Mutability:  AttributeMutabilityReadOnly(),
					Name:        "version",
				}),
			},
		}),
	}
}

// SchemasAttributes represent the common attribute "schemas".
func SchemasAttributes() CoreAttribute {
	return SimpleCoreAttribute(SimpleStringParams(StringParams{
		MultiValued: true,
		Mutability:  AttributeMutabilityImmutable(),
		Name:        "schemas",
		Required:    true,
	}))
}

// Unlike other core resources, the Schema" resource MAY contain a complex object within a sub-attribute.
func schemaAttributes(subAttrs bool) []CoreAttribute {
	attributes := []CoreAttribute{
		SimpleCoreAttribute(SimpleStringParams(StringParams{
			Description: optional.NewString("The attribute's name."),
			Mutability:  AttributeMutabilityReadOnly(),
			Name:        "name",
			Required:    true,
		})),
		SimpleCoreAttribute(SimpleStringParams(StringParams{
			CanonicalValues: []string{
				"string", "boolean", "binary", "decimal",
				"integer", "dateTime", "reference", "complex",
			},
			Description: optional.NewString("The attribute's data type."),
			Mutability:  AttributeMutabilityReadOnly(),
			Name:        "type",
			Required:    true,
		})),

		// subAttributes added below

		SimpleCoreAttribute(SimpleBooleanParams(BooleanParams{
			Description: optional.NewString("A Boolean value indicating the attribute's plurality."),
			Mutability:  AttributeMutabilityReadOnly(),
			Name:        "multiValued",
			Required:    true,
		})),
		SimpleCoreAttribute(SimpleStringParams(StringParams{
			Description: optional.NewString("The attribute's human-readable description."),
			Mutability:  AttributeMutabilityReadOnly(),
			Name:        "description",
		})),
		SimpleCoreAttribute(SimpleBooleanParams(BooleanParams{
			Description: optional.NewString("A Boolean value that specifies whether or not the attribute is required."),
			Mutability:  AttributeMutabilityReadOnly(),
			Name:        "required",
			Required:    true,
		})),
		SimpleCoreAttribute(SimpleStringParams(StringParams{
			Description: optional.NewString("A collection of suggested canonical values that MAY be used."),
			MultiValued: true,
			Mutability:  AttributeMutabilityReadOnly(),
			Name:        "canonicalValues",
		})),
		SimpleCoreAttribute(SimpleBooleanParams(BooleanParams{
			Description: optional.NewString("A Boolean value that specifies whether or not a string attribute is case sensitive."),
			Mutability:  AttributeMutabilityReadOnly(),
			Name:        "caseExact",
		})),
		SimpleCoreAttribute(SimpleStringParams(StringParams{
			CanonicalValues: []string{
				"readOnly", "readWrite",
				"immutable", "writeOnly",
			},
			CaseExact:   true,
			Description: optional.NewString("A single keyword indicating the circumstances under which the value of the attribute can be (re)defined."),
			Mutability:  AttributeMutabilityReadOnly(),
			Name:        "mutability",
		})),
		SimpleCoreAttribute(SimpleStringParams(StringParams{
			CanonicalValues: []string{
				"always", "never",
				"default", "request",
			},
			CaseExact:   true,
			Description: optional.NewString("A single keyword that indicates when an attribute and associated values are returned in a response."),
			Mutability:  AttributeMutabilityReadOnly(),
			Name:        "returned",
		})),
		SimpleCoreAttribute(SimpleStringParams(StringParams{
			CanonicalValues: []string{
				"none", "server", "global",
			},
			CaseExact:   true,
			Description: optional.NewString("A single keyword value that specifies how the service provider enforces uniqueness of attribute values."),
			Mutability:  AttributeMutabilityReadOnly(),
			Name:        "uniqueness",
		})),
		SimpleCoreAttribute(SimpleStringParams(StringParams{
			CaseExact:   true,
			Description: optional.NewString("A multi-valued array of JSON strings that indicate the SCIM resource types that may be referenced."),
			MultiValued: true,
			Mutability:  AttributeMutabilityReadOnly(),
			Name:        "referenceTypes",
		})),
	}

	if subAttrs {
		attributes = append(attributes, CoreAttribute{
			description:   optional.NewString("Defines a set of sub-attributes."),
			multiValued:   true,
			mutability:    attributeMutabilityReadOnly,
			name:          "subAttributes",
			subAttributes: schemaAttributes(false),
			typ:           attributeDataTypeComplex,
		})
	}
	return attributes
}

// CoreGroupSchema returns the default "Group" Resource Schema.
// RFC: https://tools.ietf.org/html/rfc7643#section-4.2
func CoreGroupSchema() Schema {
	return Schema{
		Attributes: []CoreAttribute{
			SimpleCoreAttribute(SimpleStringParams(StringParams{
				Description: optional.NewString("A human-readable name for the Group. REQUIRED."),
				Name:        "displayName",
				Required:    true,
			})),
			ComplexCoreAttribute(ComplexParams{
				Description: optional.NewString("A list of members of the Group."),
				MultiValued: true,
				Name:        "members",
				SubAttributes: []SimpleParams{
					SimpleStringParams(StringParams{
						Description: optional.NewString("Identifier of the member of this Group."),
						Mutability:  AttributeMutabilityImmutable(),
						Name:        "value",
					}),
					SimpleReferenceParams(ReferenceParams{
						Description:    optional.NewString("The URI corresponding to a SCIM resource that is a member of this Group."),
						Mutability:     AttributeMutabilityImmutable(),
						Name:           "$ref",
						ReferenceTypes: []AttributeReferenceType{"User", "Group"},
					}),
					SimpleStringParams(StringParams{
						CanonicalValues: []string{"User", "Group"},
						Description:     optional.NewString("A label indicating the type of resource, e.g., 'User' or 'Group'."),
						Mutability:      AttributeMutabilityImmutable(),
						Name:            "type",
					}),
					SimpleStringParams(StringParams{
						Description: optional.NewString("A human-readable name for the group member, primarily used for display purposes."),
						Mutability:  AttributeMutabilityImmutable(),
						Name:        "display",
					}),
				},
			}),
		},
		Description: optional.NewString("Group"),
		ID:          "urn:ietf:params:scim:schemas:core:2.0:Group",
		Name:        optional.NewString("Group"),
	}
}

// CoreUserSchema returns the default "User" Resource Schema.
// RFC: https://tools.ietf.org/html/rfc7643#section-4.1
func CoreUserSchema() Schema {
	return Schema{
		Attributes: []CoreAttribute{
			SimpleCoreAttribute(SimpleStringParams(StringParams{
				Description: optional.NewString("Unique identifier for the User, typically used by the user to directly authenticate to the service provider. Each User MUST include a non-empty userName value. This identifier MUST be unique across the service provider's entire set of Users. REQUIRED."),
				Name:        "userName",
				Required:    true,
				Uniqueness:  AttributeUniquenessServer(),
			})),
			ComplexCoreAttribute(ComplexParams{
				Description: optional.NewString("The components of the user's real name. Providers MAY return just the full name as a single string in the formatted sub-attribute, or they MAY return just the individual component attributes using the other sub-attributes, or they MAY return both. If both variants are returned, they SHOULD be describing the same name, with the formatted name indicating how the component attributes should be combined."),
				Name:        "name",
				SubAttributes: []SimpleParams{
					SimpleStringParams(StringParams{
						Description: optional.NewString("The full name, including all middle names, titles, and suffixes as appropriate, formatted for display (e.g., 'Ms. Barbara J Jensen, III')."),
						Name:        "formatted",
					}),
					SimpleStringParams(StringParams{
						Description: optional.NewString("The family name of the User, or last name in most Western languages (e.g., 'Jensen' given the full name 'Ms. Barbara J Jensen, III')."),
						Name:        "familyName",
					}),
					SimpleStringParams(StringParams{
						Description: optional.NewString("The given name of the User, or first name in most Western languages (e.g., 'Barbara' given the full name 'Ms. Barbara J Jensen, III')."),
						Name:        "givenName",
					}),
					SimpleStringParams(StringParams{
						Description: optional.NewString("The middle name(s) of the User (e.g., 'Jane' given the full name 'Ms. Barbara J Jensen, III')."),
						Name:        "middleName",
					}),
					SimpleStringParams(StringParams{
						Description: optional.NewString("The honorific prefix(es) of the User, or title in most Western languages (e.g., 'Ms.' given the full name 'Ms. Barbara J Jensen, III')."),
						Name:        "honorificPrefix",
					}),
					SimpleStringParams(StringParams{
						Description: optional.NewString("The honorific suffix(es) of the User, or suffix in most Western languages (e.g., 'III' given the full name 'Ms. Barbara J Jensen, III')."),
						Name:        "honorificSuffix",
					}),
				},
			}),
			SimpleCoreAttribute(SimpleStringParams(StringParams{
				Description: optional.NewString("The name of the User, suitable for display to end-users. The name SHOULD be the full name of the User being described, if known."),
				Name:        "displayName",
			})),
			SimpleCoreAttribute(SimpleStringParams(StringParams{
				Description: optional.NewString("The casual way to address the user in real life, e.g., 'Bob' or 'Bobby' instead of 'Robert'. This attribute SHOULD NOT be used to represent a User's username (e.g., 'bjensen' or 'mpepperidge')."),
				Name:        "nickName",
			})),
			SimpleCoreAttribute(SimpleReferenceParams(ReferenceParams{
				Description:    optional.NewString("A fully qualified URL pointing to a page representing the User's online profile."),
				Name:           "profileUrl",
				ReferenceTypes: []AttributeReferenceType{AttributeReferenceTypeExternal},
			})),
			SimpleCoreAttribute(SimpleStringParams(StringParams{
				Description: optional.NewString("The user's title, such as \"Vice President.\""),
				Name:        "title",
			})),
			SimpleCoreAttribute(SimpleStringParams(StringParams{
				Description: optional.NewString("Used to identify the relationship between the organization and the user. Typical values used might be 'Contractor', 'Employee', 'Intern', 'Temp', 'External', and 'Unknown', but any value may be used."),
				Name:        "userType",
			})),
			SimpleCoreAttribute(SimpleStringParams(StringParams{
				Description: optional.NewString("Indicates the User's preferred written or spoken language. Generally used for selecting a localized user interface; e.g., 'en_US' specifies the language English and country US."),
				Name:        "preferredLanguage",
			})),
			SimpleCoreAttribute(SimpleStringParams(StringParams{
				Description: optional.NewString("Used to indicate the User's default location for purposes of localizing items such as currency, date time format, or numerical representations."),
				Name:        "locale",
			})),
			SimpleCoreAttribute(SimpleStringParams(StringParams{
				Description: optional.NewString("The User's time zone in the 'Olson' time zone database format, e.g., 'America/Los_Angeles'."),
				Name:        "timezone",
			})),
			SimpleCoreAttribute(SimpleBooleanParams(BooleanParams{
				Description: optional.NewString("A Boolean value indicating the User's administrative status."),
				Name:        "active",
			})),
			SimpleCoreAttribute(SimpleStringParams(StringParams{
				Description: optional.NewString("The User's cleartext password. This attribute is intended to be used as a means to specify an initial password when creating a new User or to reset an existing User's password."),
				Mutability:  AttributeMutabilityWriteOnly(),
				Name:        "password",
				Returned:    AttributeReturnedNever(),
			})),
			ComplexCoreAttribute(ComplexParams{
				Description: optional.NewString("Email addresses for the user. The value SHOULD be canonicalized by the service provider, e.g., 'bjensen@example.com' instead of 'bjensen@EXAMPLE.COM'. Canonical type values of 'work', 'home', and 'other'."),
				MultiValued: true,
				Name:        "emails",
				SubAttributes: []SimpleParams{
					SimpleStringParams(StringParams{
						Description: optional.NewString("Email addresses for the user. The value SHOULD be canonicalized by the service provider, e.g., 'bjensen@example.com' instead of 'bjensen@EXAMPLE.COM'. Canonical type values of 'work', 'home', and 'other'."),
						Name:        "value",
					}),
					SimpleStringParams(StringParams{
						Description: optional.NewString("A human-readable name, primarily used for display purposes. READ-ONLY."),
						Name:        "display",
					}),
					SimpleStringParams(StringParams{
						CanonicalValues: []string{"work", "home", "other"},
						Description:     optional.NewString("A label indicating the attribute's function, e.g., 'work' or 'home'."),
						Name:            "type",
					}),
					SimpleBooleanParams(BooleanParams{
						Description: optional.NewString("A Boolean value indicating the 'primary' or preferred attribute value for this attribute, e.g., the preferred mailing address or primary email address. The primary attribute value 'true' MUST appear no more than once."),
						Name:        "primary",
					}),
				},
			}),
			ComplexCoreAttribute(ComplexParams{
				Description: optional.NewString("Phone numbers for the User. The value SHOULD be canonicalized by the service provider according to the format specified in RFC 3966, e.g., 'tel:+1-201-555-0123'. Canonical type values of 'work', 'home', 'mobile', 'fax', 'pager', and 'other'."),
				MultiValued: true,
				Name:        "phoneNumbers",
				SubAttributes: []SimpleParams{
					SimpleStringParams(StringParams{
						Description: optional.NewString("Phone number of the User."),
						Name:        "value",
					}),
					SimpleStringParams(StringParams{
						Description: optional.NewString("A human-readable name, primarily used for display purposes. READ-ONLY."),
						Name:        "display",
					}),
					SimpleStringParams(StringParams{
						CanonicalValues: []string{"work", "home", "mobile", "fax", "pager", "other"},
						Description:     optional.NewString("A label indicating the attribute's function, e.g., 'work', 'home', 'mobile'."),
						Name:            "type",
					}),
					SimpleBooleanParams(BooleanParams{
						Description: optional.NewString("A Boolean value indicating the 'primary' or preferred attribute value for this attribute, e.g., the preferred phone number or primary phone number. The primary attribute value 'true' MUST appear no more than once."),
						Name:        "primary",
					}),
				},
			}),
			ComplexCoreAttribute(ComplexParams{
				Description: optional.NewString("Instant messaging addresses for the User."),
				MultiValued: true,
				Name:        "ims",
				SubAttributes: []SimpleParams{
					SimpleStringParams(StringParams{
						Description: optional.NewString("Instant messaging address for the User."),
						Name:        "value",
					}),
					SimpleStringParams(StringParams{
						Description: optional.NewString("A human-readable name, primarily used for display purposes. READ-ONLY."),
						Name:        "display",
					}),
					SimpleStringParams(StringParams{
						CanonicalValues: []string{"aim", "gtalk", "icq", "xmpp", "msn", "skype", "qq", "yahoo"},
						Description:     optional.NewString("A label indicating the attribute's function, e.g., 'aim', 'gtalk', 'xmpp'."),
						Name:            "type",
					}),
					SimpleBooleanParams(BooleanParams{
						Description: optional.NewString("A Boolean value indicating the 'primary' or preferred attribute value for this attribute, e.g., the preferred messenger or primary messenger. The primary attribute value 'true' MUST appear no more than once."),
						Name:        "primary",
					}),
				},
			}),
			ComplexCoreAttribute(ComplexParams{
				Description: optional.NewString("URLs of photos of the User."),
				MultiValued: true,
				Name:        "photos",
				SubAttributes: []SimpleParams{
					SimpleReferenceParams(ReferenceParams{
						Description:    optional.NewString("URL of a photo of the User."),
						Name:           "value",
						ReferenceTypes: []AttributeReferenceType{AttributeReferenceTypeExternal},
					}),
					SimpleStringParams(StringParams{
						Description: optional.NewString("A human-readable name, primarily used for display purposes. READ-ONLY."),
						Name:        "display",
					}),
					SimpleStringParams(StringParams{
						CanonicalValues: []string{"photo", "thumbnail"},
						Description:     optional.NewString("A label indicating the attribute's function, i.e., 'photo' or 'thumbnail'."),
						Name:            "type",
					}),
					SimpleBooleanParams(BooleanParams{
						Description: optional.NewString("A Boolean value indicating the 'primary' or preferred attribute value for this attribute, e.g., the preferred photo or thumbnail. The primary attribute value 'true' MUST appear no more than once."),
						Name:        "primary",
					}),
				},
			}),
			ComplexCoreAttribute(ComplexParams{
				Description: optional.NewString("A physical mailing address for this User. Canonical type values of 'work', 'home', and 'other'. This attribute is a complex type with the following sub-attributes."),
				MultiValued: true,
				Name:        "addresses",
				SubAttributes: []SimpleParams{
					SimpleStringParams(StringParams{
						Description: optional.NewString("The full mailing address, formatted for display or use with a mailing label. This attribute MAY contain newlines."),
						Name:        "formatted",
					}),
					SimpleStringParams(StringParams{
						Description: optional.NewString("The full street address component, which may include house number, street name, P.O. box, and multi-line extended street address information. This attribute MAY contain newlines."),
						Name:        "streetAddress",
					}),
					SimpleStringParams(StringParams{
						Description: optional.NewString("The city or locality component."),
						Name:        "locality",
					}),
					SimpleStringParams(StringParams{
						Description: optional.NewString("The state or region component."),
						Name:        "region",
					}),
					SimpleStringParams(StringParams{
						Description: optional.NewString("The zip code or postal code component."),
						Name:        "postalCode",
					}),
					SimpleStringParams(StringParams{
						Description: optional.NewString("The country name component."),
						Name:        "country",
					}),
					SimpleStringParams(StringParams{
						CanonicalValues: []string{"work", "home", "other"},
						Description:     optional.NewString("A label indicating the attribute's function, e.g., 'work' or 'home'."),
						Name:            "type",
					}),
				},
			}),
			ComplexCoreAttribute(ComplexParams{
				Description: optional.NewString("A list of groups to which the user belongs, either through direct membership, through nested groups, or dynamically calculated."),
				MultiValued: true,
				Mutability:  AttributeMutabilityReadOnly(),
				Name:        "groups",
				SubAttributes: []SimpleParams{
					SimpleStringParams(StringParams{
						Description: optional.NewString("The identifier of the User's group."),
						Mutability:  AttributeMutabilityReadOnly(),
						Name:        "value",
					}),
					SimpleReferenceParams(ReferenceParams{
						Description:    optional.NewString("The URI of the corresponding 'Group' resource to which the user belongs."),
						Mutability:     AttributeMutabilityReadOnly(),
						Name:           "$ref",
						ReferenceTypes: []AttributeReferenceType{"User", "Group"},
					}),
					SimpleStringParams(StringParams{
						Description: optional.NewString("A human-readable name, primarily used for display purposes. READ-ONLY."),
						Mutability:  AttributeMutabilityReadOnly(),
						Name:        "display",
					}),
					SimpleStringParams(StringParams{
						CanonicalValues: []string{"direct", "indirect"},
						Description:     optional.NewString("A label indicating the attribute's function, e.g., 'direct' or 'indirect'."),
						Mutability:      AttributeMutabilityReadOnly(),
						Name:            "type",
					}),
				},
			}),
			ComplexCoreAttribute(ComplexParams{
				Description: optional.NewString("A list of entitlements for the User that represent a thing the User has."),
				MultiValued: true,
				Name:        "entitlements",
				SubAttributes: []SimpleParams{
					SimpleStringParams(StringParams{
						Description: optional.NewString("The value of an entitlement."),
						Name:        "value",
					}),
					SimpleStringParams(StringParams{
						Description: optional.NewString("A human-readable name, primarily used for display purposes. READ-ONLY."),
						Name:        "display",
					}),
					SimpleStringParams(StringParams{
						Description: optional.NewString("A label indicating the attribute's function."),
						Name:        "type",
					}),
					SimpleBooleanParams(BooleanParams{
						Description: optional.NewString("A Boolean value indicating the 'primary' or preferred attribute value for this attribute. The primary attribute value 'true' MUST appear no more than once."),
						Name:        "primary",
					}),
				},
			}),
			ComplexCoreAttribute(ComplexParams{
				Description: optional.NewString("A list of roles for the User that collectively represent who the User is, e.g., 'Student', 'Faculty'."),
				MultiValued: true,
				Name:        "roles",
				SubAttributes: []SimpleParams{
					SimpleStringParams(StringParams{
						Description: optional.NewString("The value of a role."),
						Name:        "value",
					}),
					SimpleStringParams(StringParams{
						Description: optional.NewString("A human-readable name, primarily used for display purposes. READ-ONLY."),
						Name:        "display",
					}),
					SimpleStringParams(StringParams{
						Description: optional.NewString("A label indicating the attribute's function."),
						Name:        "type",
					}),
					SimpleBooleanParams(BooleanParams{
						Description: optional.NewString("A Boolean value indicating the 'primary' or preferred attribute value for this attribute. The primary attribute value 'true' MUST appear no more than once."),
						Name:        "primary",
					}),
				},
			}),
			ComplexCoreAttribute(ComplexParams{
				Description: optional.NewString("A list of certificates issued to the User."),
				MultiValued: true,
				Name:        "x509Certificates",
				SubAttributes: []SimpleParams{
					SimpleBinaryParams(BinaryParams{
						Description: optional.NewString("The value of an X.509 certificate."),
						Name:        "value",
					}),
					SimpleStringParams(StringParams{
						Description: optional.NewString("A human-readable name, primarily used for display purposes. READ-ONLY."),
						Name:        "display",
					}),
					SimpleStringParams(StringParams{
						Description: optional.NewString("A label indicating the attribute's function."),
						Name:        "type",
					}),
					SimpleBooleanParams(BooleanParams{
						Description: optional.NewString("A Boolean value indicating the 'primary' or preferred attribute value for this attribute. The primary attribute value 'true' MUST appear no more than once."),
						Name:        "primary",
					}),
				},
			}),
		},
		Description: optional.NewString("User Account"),
		ID:          "urn:ietf:params:scim:schemas:core:2.0:User",
		Name:        optional.NewString("User"),
	}
}

// Definition return the Schema Definition.
// RFC: https://tools.ietf.org/html/rfc7643#section-7
func Definition() Schema {
	return Schema{
		Attributes: []CoreAttribute{
			SimpleCoreAttribute(SimpleStringParams(StringParams{
				Description: optional.NewString("The unique URI of the schema."),
				Mutability:  AttributeMutabilityReadOnly(),
				Name:        "id",
				Required:    true,
			})),
			SimpleCoreAttribute(SimpleStringParams(StringParams{
				Description: optional.NewString("The schema's human-readable name."),
				Mutability:  AttributeMutabilityReadOnly(),
				Name:        "name",
			})),
			SimpleCoreAttribute(SimpleStringParams(StringParams{
				Description: optional.NewString("The schema's human-readable description."),
				Mutability:  AttributeMutabilityReadOnly(),
				Name:        "description",
			})),
			{ // this is the internal struct to enable nested complex attributes.
				description:   optional.NewString("A complex type that defines service provider attributes and their qualities."),
				multiValued:   true,
				mutability:    attributeMutabilityReadOnly,
				name:          "attributes",
				required:      true,
				subAttributes: schemaAttributes(true),
				typ:           attributeDataTypeComplex,
			},
		},
		Description: optional.String{},
		ID:          "urn:ietf:params:scim:schemas:core:2.0:Schema",
		Name:        optional.NewString("Schema"),
	}
}

// ExtensionEnterpriseUser returns the default Enterprise User Schema Extension.
// RFC: https://tools.ietf.org/html/rfc7643#section-4.3
func ExtensionEnterpriseUser() Schema {
	return Schema{
		Attributes: []CoreAttribute{
			SimpleCoreAttribute(SimpleStringParams(StringParams{
				Description: optional.NewString("Numeric or alphanumeric identifier assigned to a person, typically based on order of hire or association with an organization."),
				Name:        "employeeNumber",
			})),
			SimpleCoreAttribute(SimpleStringParams(StringParams{
				Description: optional.NewString("Identifies the name of a cost center."),
				Name:        "costCenter",
			})),
			SimpleCoreAttribute(SimpleStringParams(StringParams{
				Description: optional.NewString("Identifies the name of an organization."),
				Name:        "organization",
			})),
			SimpleCoreAttribute(SimpleStringParams(StringParams{
				Description: optional.NewString("Identifies the name of a division."),
				Name:        "division",
			})),
			SimpleCoreAttribute(SimpleStringParams(StringParams{
				Description: optional.NewString("Identifies the name of a department."),
				Name:        "department",
			})),
			ComplexCoreAttribute(ComplexParams{
				Description: optional.NewString("The User's manager. A complex type that optionally allows service providers to represent organizational hierarchy by referencing the 'id' attribute of another User."),
				Name:        "manager",
				SubAttributes: []SimpleParams{
					SimpleStringParams(StringParams{
						Description: optional.NewString("The id of the SCIM resource representing the User's manager. REQUIRED."),
						Name:        "value",
					}),
					SimpleReferenceParams(ReferenceParams{
						Description:    optional.NewString("The URI of the SCIM resource representing the User's manager. REQUIRED."),
						Name:           "$ref",
						ReferenceTypes: []AttributeReferenceType{"User"},
					}),
					SimpleStringParams(StringParams{
						Description: optional.NewString("The displayName of the User's manager. OPTIONAL and READ-ONLY."),
						Mutability:  AttributeMutabilityReadOnly(),
						Name:        "displayName",
					}),
				},
			}),
		},
		Description: optional.NewString("Enterprise User"),
		ID:          "urn:ietf:params:scim:schemas:extension:enterprise:2.0:User",
		Name:        optional.NewString("Enterprise User"),
	}
}

// ResourceTypeSchema returns the Resource Type Schema.
// RFC: https://tools.ietf.org/html/rfc7643#section-6
func ResourceTypeSchema() Schema {
	return Schema{
		Attributes: []CoreAttribute{
			SimpleCoreAttribute(SimpleStringParams(StringParams{
				Description: optional.NewString("The resource type’s server unique id."),
				Mutability:  AttributeMutabilityReadOnly(),
				Name:        "id",
				Required:    true,
			})),
			SimpleCoreAttribute(SimpleStringParams(StringParams{
				Description: optional.NewString("The resource type name."),
				Mutability:  AttributeMutabilityReadOnly(),
				Name:        "name",
			})),
			SimpleCoreAttribute(SimpleStringParams(StringParams{
				Description: optional.NewString("The resource type’s human-readable description."),
				Mutability:  AttributeMutabilityReadOnly(),
				Name:        "description",
			})),
			SimpleCoreAttribute(SimpleStringParams(StringParams{
				Description: optional.NewString("The resource type’s HTTP-addressable endpoint relative to the Base URL of the service provider, e.g., \"Users\"."),
				Mutability:  AttributeMutabilityReadOnly(),
				Name:        "endpoint",
				Required:    true,
			})),
			SimpleCoreAttribute(SimpleStringParams(StringParams{
				Description: optional.NewString("The resource type’s primary/base schema URI."),
				Mutability:  AttributeMutabilityReadOnly(),
				Name:        "schema",
				Required:    true,
			})),
			ComplexCoreAttribute(ComplexParams{
				Description: optional.NewString("A list of URIs of the resource type’s schema extensions."),
				MultiValued: true,
				Mutability:  AttributeMutabilityReadOnly(),
				Name:        "schemaExtensions",
				SubAttributes: []SimpleParams{
					SimpleStringParams(StringParams{
						Description: optional.NewString("The URI of an extended schema"),
						Mutability:  AttributeMutabilityReadOnly(),
						Name:        "schema",
						Required:    true,
					}),
					SimpleBooleanParams(BooleanParams{
						Description: optional.NewString("A Boolean value that specifies whether or not the schema extension is required for the resource type."),
						Mutability:  AttributeMutabilityReadOnly(),
						Name:        "required",
						Required:    true,
					}),
				},
			}),
		},
		Description: optional.NewString("Metadata about a resource type."),
		ID:          "urn:ietf:params:scim:schemas:core:2.0:ResourceType",
		Name:        optional.NewString("Resource Type"),
	}
}

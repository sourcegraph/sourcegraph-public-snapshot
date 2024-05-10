package filter

import (
	"testing"

	"github.com/elimity-com/scim/schema"
)

func TestPathValidator_Validate(t *testing.T) {
	// More info: https://tools.ietf.org/html/rfc7644#section-3.5.2
	t.Run("Valid", func(t *testing.T) {
		for _, f := range []string{
			`urn:ietf:params:scim:schemas:core:2.0:User:name`,
			`urn:ietf:params:scim:schemas:core:2.0:User:name.familyName`,
			`urn:ietf:params:scim:schemas:core:2.0:User:emails[type eq "work"]`,
			`urn:ietf:params:scim:schemas:core:2.0:User:emails[type eq "work"].display`,

			`name`,
			`name.familyName`,
			`emails`,
			`emails.value`,
			`emails[type eq "work"]`,
			`emails[type eq "work"].display`,
		} {
			validator, err := NewPathValidator(f, schema.CoreUserSchema(), schema.ExtensionEnterpriseUser())
			if err != nil {
				t.Fatal(err)
			}
			if err := validator.Validate(); err != nil {
				t.Errorf("(%s) %v", f, err)
			}
		}
	})

	t.Run("Invalid", func(t *testing.T) {
		for _, f := range []string{
			`urn:ietf:params:scim:schemas:core:2.0:Invalid:name`,

			`invalid`,
			`name.invalid`,
			`emails[invalid eq "work"]`,
			`emails[type eq "work"].invalid`,
		} {
			validator, err := NewPathValidator(f, schema.CoreUserSchema(), schema.ExtensionEnterpriseUser())
			if err != nil {
				t.Fatal(err)
			}
			if err := validator.Validate(); err == nil {
				t.Errorf("(%s) should not be valid", f)
			}
		}
	})
}

func TestValidator_PassesFilter(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		for _, test := range []struct {
			filter  string
			valid   map[string]interface{}
			invalid map[string]interface{}
		}{
			{
				filter: `userName eq "john"`,
				valid: map[string]interface{}{
					"userName": "john",
				},
				invalid: map[string]interface{}{
					"userName": "doe",
				},
			},
			{
				filter: `emails[type eq "work"]`,
				valid: map[string]interface{}{
					"emails": []interface{}{
						map[string]interface{}{
							"type": "work",
						},
					},
				},
				invalid: map[string]interface{}{
					"emails": []interface{}{
						map[string]interface{}{
							"type": "private",
						},
					},
				},
			},
		} {
			validator, err := NewValidator(test.filter, schema.CoreUserSchema())
			if err != nil {
				t.Fatal(err)
			}
			if resource := test.valid; resource != nil {
				if err := validator.PassesFilter(resource); err != nil {
					t.Errorf("(%v) should be valid: %v", resource, err)
				}
			}
			if resource := test.invalid; resource != nil {
				if err := validator.PassesFilter(resource); err == nil {
					t.Errorf("(%v) should not be valid", resource)
				}
			}
		}
	})

	for _, test := range []struct {
		name   string
		amount int
		filter string
	}{
		{name: "eq", amount: 1, filter: `userName eq "di-wu"`},
		{name: "ne", amount: 5, filter: `userName ne "di-wu"`},
		{name: "co", amount: 3, filter: `userName co "u"`},
		{name: "co", amount: 2, filter: `name.familyName co "d"`},
		{name: "sw", amount: 2, filter: `userName sw "a"`},
		{name: "sw", amount: 2, filter: `urn:ietf:params:scim:schemas:core:2.0:User:userName sw "a"`},
		{name: "ew", amount: 2, filter: `userName ew "n"`},
		{name: "pr", amount: 6, filter: `userName pr`},
		{name: "gt", amount: 2, filter: `userName gt "guest"`},
		{name: "ge", amount: 3, filter: `userName ge "guest"`},
		{name: "lt", amount: 3, filter: `userName lt "guest"`},
		{name: "le", amount: 4, filter: `userName le "guest"`},
		{name: "value", amount: 2, filter: `emails[type eq "work"]`},
		{name: "and", amount: 1, filter: `name.familyName eq "ad" and userType eq "admin"`},
		{name: "or", amount: 2, filter: `name.familyName eq "ad" or userType eq "admin"`},
		{name: "not", amount: 5, filter: `not (userName eq "di-wu")`},
		{name: "meta", amount: 1, filter: `meta.lastModified gt "2011-05-13T04:42:34Z"`},
		{name: "schemas", amount: 2, filter: `schemas eq "urn:ietf:params:scim:schemas:core:2.0:User"`},
	} {
		t.Run(test.name, func(t *testing.T) {
			userSchema := schema.CoreUserSchema()
			userSchema.Attributes = append(userSchema.Attributes, schema.SchemasAttributes())
			userSchema.Attributes = append(userSchema.Attributes, schema.CommonAttributes()...)
			validator, err := NewValidator(test.filter, userSchema)
			if err != nil {
				t.Fatal(err)
			}

			var amount int
			for _, resource := range testResources() {
				if err := validator.PassesFilter(resource); err == nil {
					amount++
				}
			}
			if amount != test.amount {
				t.Errorf("Expected %d resources to pass, got %d.", test.amount, amount)
			}
		})
	}

	t.Run("extensions", func(t *testing.T) {
		for _, test := range []struct {
			amount int
			filter string
		}{
			{
				amount: 1,
				filter: `urn:ietf:params:scim:schemas:extension:enterprise:2.0:User:manager.displayName eq "di-wu"`,
			},
			{
				amount: 1,
				filter: `urn:ietf:params:scim:schemas:extension:enterprise:2.0:User:organization eq "Elimity"`,
			},
		} {
			validator, err := NewValidator(test.filter, schema.ExtensionEnterpriseUser())
			if err != nil {
				t.Fatal(err)
			}
			var amount int
			for _, resource := range testResources() {
				if err := validator.PassesFilter(resource); err == nil {
					amount++
				}
			}
			if amount != test.amount {
				t.Errorf("Expected %d resources to pass, got %d.", test.amount, amount)
			}
		}
	})
}

func TestValidator_Validate(t *testing.T) {
	// More info: https://tools.ietf.org/html/rfc7644#section-3.4.2.2
	userSchema := schema.CoreUserSchema()
	userSchema.Attributes = append(userSchema.Attributes, schema.CommonAttributes()...)

	for _, f := range []string{
		`userName Eq "john"`,
		`Username eq "john"`,

		`userName eq "bjensen"`,
		`name.familyName co "O'Malley"`,
		`userName sw "J"`,
		`urn:ietf:params:scim:schemas:core:2.0:User:userName sw "J"`,
		`title pr`,
		`meta.lastModified gt "2011-05-13T04:42:34Z"`,
		`meta.lastModified ge "2011-05-13T04:42:34Z"`,
		`meta.lastModified lt "2011-05-13T04:42:34Z"`,
		`meta.lastModified le "2011-05-13T04:42:34Z"`,
		`title pr and userType eq "Employee"`,
		`title pr or userType eq "Intern"`,
		`schemas eq "urn:ietf:params:scim:schemas:extension:enterprise:2.0:User"`,
		`userType eq "Employee" and (emails co "example.com" or emails.value co "example.org")`,
		`userType ne "Employee" and not (emails co "example.com" or emails.value co "example.org")`,
		`userType eq "Employee" and (emails.type eq "work")`,
		`userType eq "Employee" and emails[type eq "work" and value co "@example.com"]`,
		`emails[type eq "work" and value co "@example.com"] or ims[type eq "xmpp" and value co "@foo.com"]`,
	} {
		validator, err := NewValidator(f, userSchema)
		if err != nil {
			t.Fatal(err)
		}
		if err := validator.Validate(); err != nil {
			t.Errorf("(%s) %v", f, err)
		}
	}
}

func testResources() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"schemas": []interface{}{
				"urn:ietf:params:scim:schemas:core:2.0:User",
			},
			"userName": "di-wu",
			"userType": "admin",
			"name": map[string]interface{}{
				"familyName": "di",
				"givenName":  "wu",
			},
			"emails": []interface{}{
				map[string]interface{}{
					"value": "quint@elimity.com",
					"type":  "work",
				},
			},
			"meta": map[string]interface{}{
				"lastModified": "2020-07-26T20:02:34Z",
			},
			"urn:ietf:params:scim:schemas:extension:enterprise:2.0:User:organization": "Elimity",
		},
		{
			"schemas": []interface{}{
				"urn:ietf:params:scim:schemas:core:2.0:User",
			},
			"userName": "noreply",
			"emails": []interface{}{
				map[string]interface{}{
					"value": "noreply@elimity.com",
					"type":  "work",
				},
			},
		},
		{
			"userName": "admin",
			"userType": "admin",
			"name": map[string]interface{}{
				"familyName": "ad",
				"givenName":  "min",
			},
			"urn:ietf:params:scim:schemas:extension:enterprise:2.0:User:manager": map[string]interface{}{
				"displayName": "di-wu",
			},
		},
		{"userName": "guest"},
		{
			"userName": "unknown",
			"name": map[string]interface{}{
				"familyName": "un",
				"givenName":  "known",
			},
		},
		{"userName": "another"},
	}
}

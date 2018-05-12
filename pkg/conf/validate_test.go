package conf

import (
	"strings"
	"testing"

	"github.com/sourcegraph/sourcegraph/schema"
)

func TestValidate(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		res, err := validate([]byte(schema.SiteSchemaJSON), []byte(`{"secretKey":"abc"}`))
		if err != nil {
			t.Fatal(err)
		}
		if len(res.Errors()) != 0 {
			t.Errorf("errors: %v", res.Errors())
		}
	})

	t.Run("invalid", func(t *testing.T) {
		res, err := validate([]byte(schema.SiteSchemaJSON), []byte(`{"a":1}`))
		if err != nil {
			t.Fatal(err)
		}
		if len(res.Errors()) == 0 {
			t.Error("want invalid")
		}
	})
}

func TestValidateCustom(t *testing.T) {
	tests := map[string]struct {
		input                schema.SiteConfiguration
		raw                  string
		wantValidationErrors []string
	}{
		"no auth.provider": {
			input:                schema.SiteConfiguration{},
			wantValidationErrors: []string{"no auth provider set"},
		},
		"unrecognized auth.provider": {
			input:                schema.SiteConfiguration{AuthProvider: "x"},
			wantValidationErrors: []string{"no auth provider set", "auth.provider is deprecated"},
		},
		"deprecated auth.provider": {
			input:                schema.SiteConfiguration{AuthProvider: "builtin"},
			wantValidationErrors: []string{"auth.provider is deprecated"},
		},
		"auth.provider and auth.providers": {
			input: schema.SiteConfiguration{
				AuthProvider:  "builtin",
				AuthProviders: []schema.AuthProviders{{Builtin: &schema.BuiltinAuthProvider{Type: "builtin"}}},
			},
			wantValidationErrors: []string{"auth.providers takes precedence"},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			var validationErrors []string
			var err error
			if test.raw != "" {
				validationErrors, err = validateCustomRaw([]byte(test.raw))
			} else {
				validationErrors, err = validateCustom(test.input)
			}
			if err != nil {
				t.Fatal(err)
			}
			wantValidationErrors := make(map[string]struct{}, len(test.wantValidationErrors))
			for _, e := range test.wantValidationErrors {
				wantValidationErrors[e] = struct{}{}
			}
			for _, e := range validationErrors {
				var found bool
				for es := range wantValidationErrors {
					if strings.Contains(e, es) {
						delete(wantValidationErrors, es)
						found = true
						break
					}
				}
				if !found {
					t.Errorf("got unexpected error %q", e)
				}
			}
			if len(wantValidationErrors) > 0 {
				t.Errorf("got no matches for expected error substrings %q", wantValidationErrors)
			}
		})
	}
}

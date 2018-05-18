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
		wantErr              string
		wantValidationErrors []string
		ignoreOthers         bool
	}{
		"no auth.provider": {
			input:                schema.SiteConfiguration{},
			wantValidationErrors: []string{"no auth providers set"},
		},
		"unrecognized auth.provider": {
			input:                schema.SiteConfiguration{AuthProvider: "x"},
			wantValidationErrors: []string{"no auth providers set", "auth.provider is deprecated"},
		},
		"unrecognized auth.providers": {
			raw:     `{"auth.providers":[{"type":"asdf"}]}`,
			wantErr: "tagged union type must have a",
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
		"auth.allowSignup deprecation": {
			input:                schema.SiteConfiguration{AuthAllowSignup: true},
			wantValidationErrors: []string{"auth.allowSignup is deprecated"},
			ignoreOthers:         true,
		},
		"multiple auth providers of same type": {
			input: schema.SiteConfiguration{
				ExperimentalFeatures: &schema.ExperimentalFeatures{MultipleAuthProviders: "enabled"},
				AuthProviders: []schema.AuthProviders{
					{Builtin: &schema.BuiltinAuthProvider{Type: "builtin"}},
					{Builtin: &schema.BuiltinAuthProvider{Type: "builtin"}},
				},
			},
			wantValidationErrors: []string{"exactly 0 or 1 auth providers of type \"builtin\""},
		},
		"multiple auth providers without experimentalFeature.multipleAuthProviders": {
			input: schema.SiteConfiguration{
				AuthProviders: []schema.AuthProviders{
					{Builtin: &schema.BuiltinAuthProvider{Type: "a"}},
					{Builtin: &schema.BuiltinAuthProvider{Type: "b"}},
				},
			},
			wantValidationErrors: []string{"auth.providers supports only a single"},
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
				if test.wantErr == "" || !strings.Contains(err.Error(), test.wantErr) {
					t.Fatal(err)
				}
				return
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
				if !found && !test.ignoreOthers {
					t.Errorf("got unexpected error %q", e)
				}
			}
			if len(wantValidationErrors) > 0 {
				t.Errorf("got no matches for expected error substrings %q", wantValidationErrors)
			}
		})
	}
}

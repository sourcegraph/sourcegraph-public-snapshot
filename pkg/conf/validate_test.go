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
		input        schema.SiteConfiguration
		raw          string
		wantErr      string
		wantProblems []string
		ignoreOthers bool
	}{
		"no auth.provider": {
			input:        schema.SiteConfiguration{},
			wantProblems: []string{"no auth providers set"},
		},
		"unrecognized auth.provider": {
			input:        schema.SiteConfiguration{AuthProvider: "x"},
			wantProblems: []string{"no auth providers set", "auth.provider is deprecated"},
		},
		"unrecognized auth.providers": {
			raw:     `{"auth.providers":[{"type":"asdf"}]}`,
			wantErr: "tagged union type must have a",
		},
		"deprecated auth.provider": {
			input:        schema.SiteConfiguration{AuthProvider: "builtin"},
			wantProblems: []string{"auth.provider is deprecated"},
		},
		"auth.provider and auth.providers": {
			input: schema.SiteConfiguration{
				AuthProvider:  "builtin",
				AuthProviders: []schema.AuthProviders{{Builtin: &schema.BuiltinAuthProvider{Type: "builtin"}}},
			},
			wantProblems: []string{"auth.providers takes precedence"},
		},
		"auth.allowSignup deprecation": {
			input:        schema.SiteConfiguration{AuthAllowSignup: true},
			wantProblems: []string{"auth.allowSignup is deprecated"},
			ignoreOthers: true,
		},
		"multiple auth providers of same type": {
			input: schema.SiteConfiguration{
				ExperimentalFeatures: &schema.ExperimentalFeatures{MultipleAuthProviders: "enabled"},
				AuthProviders: []schema.AuthProviders{
					{Builtin: &schema.BuiltinAuthProvider{Type: "builtin"}},
					{Builtin: &schema.BuiltinAuthProvider{Type: "builtin"}},
				},
			},
			wantProblems: []string{"exactly 0 or 1 auth providers of type \"builtin\""},
		},
		"multiple auth providers without experimentalFeature.multipleAuthProviders": {
			input: schema.SiteConfiguration{
				AuthProviders: []schema.AuthProviders{
					{Builtin: &schema.BuiltinAuthProvider{Type: "a"}},
					{Builtin: &schema.BuiltinAuthProvider{Type: "b"}},
				},
			},
			wantProblems: []string{"auth.providers supports only a single"},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			var problems []string
			var err error
			if test.raw != "" {
				problems, err = validateCustomRaw([]byte(test.raw))
			} else {
				problems, err = validateCustom(test.input)
			}
			if err != nil {
				if test.wantErr == "" || !strings.Contains(err.Error(), test.wantErr) {
					t.Fatal(err)
				}
				return
			}

			wantProblems := make(map[string]struct{}, len(test.wantProblems))
			for _, p := range test.wantProblems {
				wantProblems[p] = struct{}{}
			}
			for _, p := range problems {
				var found bool
				for ps := range wantProblems {
					if strings.Contains(p, ps) {
						delete(wantProblems, ps)
						found = true
						break
					}
				}
				if !found && !test.ignoreOthers {
					t.Errorf("got unexpected problem %q", p)
				}
			}
			if len(wantProblems) > 0 {
				t.Errorf("got no matches for expected problem substrings %q", wantProblems)
			}
		})
	}
}

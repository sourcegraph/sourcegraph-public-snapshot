package conf

import (
	"strings"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestValidate(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		res, err := validate([]byte(schema.SiteSchemaJSON), []byte(`{"maxReposToSearch":123}`))
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
		rawCritical, rawSite string
		wantProblem          string
		wantErr              string
	}{
		"unrecognized auth.providers": {
			rawCritical: `{"auth.providers":[{"type":"asdf"}]}`,
			rawSite:     "{}",
			wantErr:     "tagged union type must have a",
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			problems, err := validateCustomRaw(conftypes.RawUnified{
				Critical: test.rawCritical,
				Site:     test.rawSite,
			})
			if err != nil {
				if test.wantErr == "" {
					t.Fatalf("got unexpected error: %v", err)
				}
				if !strings.Contains(err.Error(), test.wantErr) {
					t.Fatal(err)
				}
				return
			}

			if test.wantProblem == "" {
				if len(problems) > 0 {
					t.Fatalf("unexpected problems: %v", problems)
				}
				return
			}
			for _, p := range problems {
				if strings.Contains(p.String(), test.wantProblem) {
					return
				}
			}
			t.Fatalf("could not find problem %q in %v", test.wantProblem, problems)
		})
	}
}

func TestProblems(t *testing.T) {
	siteProblems := NewSiteProblems(
		"siteProblem1",
		"siteProblem2",
		"siteProblem3",
	)
	externalServiceProblems := NewExternalServiceProblems(
		"externalServiceProblem1",
		"externalServiceProblem2",
		"externalServiceProblem3",
	)

	var problems Problems
	problems = append(problems, siteProblems...)
	problems = append(problems, externalServiceProblems...)

	{
		messages := make([]string, 0, len(problems))
		messages = append(messages, siteProblems.Messages()...)
		messages = append(messages, externalServiceProblems.Messages()...)

		want := strings.Join(messages, "\n")
		got := strings.Join(problems.Messages(), "\n")
		if want != got {
			t.Errorf("got %q, want %q", got, want)
		}
	}

	{
		want := strings.Join(siteProblems.Messages(), "\n")
		got := strings.Join(problems.Site().Messages(), "\n")
		if want != got {
			t.Errorf("got %q, want %q", got, want)
		}
	}

	{
		want := strings.Join(externalServiceProblems.Messages(), "\n")
		got := strings.Join(problems.ExternalService().Messages(), "\n")
		if want != got {
			t.Errorf("got %q, want %q", got, want)
		}
	}
}

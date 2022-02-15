package ci_test

import (
	"os"
	"testing"
	"time"

	"github.com/sourcegraph/sourcegraph/dev/ci/runtype"
	"github.com/sourcegraph/sourcegraph/dev/sg/root"
	bk "github.com/sourcegraph/sourcegraph/enterprise/dev/ci/internal/buildkite"
	"github.com/sourcegraph/sourcegraph/enterprise/dev/ci/internal/ci"
)

func chdirToRoot(t *testing.T) {
	path, err := root.RepositoryRoot()
	if err != nil {
		t.Fatal(err)
	}
	err = os.Chdir(path)
	if err != nil {
		t.Fatal(err)
	}
}

func TestGeneratePipeline(t *testing.T) {
	// The pipeline is very cwd dependent, so we position ourselves at the repository root.
	path, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	chdirToRoot(t)
	// Restore the cwd this test.
	defer func() { _ = os.Chdir(path) }()

	t.Run("TargetedRunType", func(t *testing.T) {
		tests := []struct {
			branch    string
			wantSteps []string
			wantErr   bool
		}{
			{
				branch:  "targeted/nothing-existing/test",
				wantErr: true,
			},
			{
				branch:  "targeted/puppeteer/test", // ambiguous
				wantErr: true,
			},
			{
				branch:  "targeted/backend-integration/test",
				wantErr: false,
				wantSteps: []string{
					":chains: Backend integration tests",
					":docker: :construction: Build server",
				},
			},
			{
				branch:  "targeted/build-server/test",
				wantErr: false,
				wantSteps: []string{
					":docker: :construction: Build server",
				},
			},
			{
				branch:  "targeted/puppeteer-tests-finalize/test",
				wantErr: false,
				wantSteps: []string{
					":puppeteer::electric_plug: Puppeteer tests prep",
					":puppeteer::electric_plug: Puppeteer tests chunk #1",
					":puppeteer::electric_plug: Puppeteer tests chunk #2",
					":puppeteer::electric_plug: Puppeteer tests chunk #3",
					":puppeteer::electric_plug: Puppeteer tests chunk #4",
					":puppeteer::electric_plug: Puppeteer tests chunk #5",
					":puppeteer::electric_plug: Puppeteer tests chunk #6",
					":puppeteer::electric_plug: Puppeteer tests chunk #7",
					":puppeteer::electric_plug: Puppeteer tests chunk #8",
					":puppeteer::electric_plug: Puppeteer tests chunk #9",
					":puppeteer::electric_plug: Puppeteer tests finalize",
				},
			},
		}

		for _, test := range tests {
			t.Run(test.branch, func(t *testing.T) {
				c := ci.NewConfig(time.Now())
				c.Branch = test.branch
				c.RunType = runtype.Targeted

				p, err := ci.GeneratePipeline(c)
				if !test.wantErr {
					if err != nil {
						t.Fatalf("want err to be nil, but got %q", err)
					}
				} else {
					if err == nil {
						t.Fatalf("want err to not be nil but got nil")
					}
					return
				}

				var found int
				for _, s := range p.Steps {
					if step, ok := s.(*bk.Step); ok {
						for _, wantStep := range test.wantSteps {
							if step.Label == wantStep {
								found++
							}
						}
					}
				}
				if found != len(test.wantSteps) {
					t.Fatalf("want %d steps, but found %d matches", len(test.wantSteps), found)
				}
			})
		}
	})
}

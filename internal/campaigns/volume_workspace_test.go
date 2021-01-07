package campaigns

import (
	"bytes"
	"context"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"

	"github.com/sourcegraph/src-cli/internal/campaigns/graphql"
	"github.com/sourcegraph/src-cli/internal/exec/expect"
)

// We may as well use the same volume ID in all the subtests.
const volumeID = "VOLUME-ID"

func TestVolumeWorkspaceCreator(t *testing.T) {
	ctx := context.Background()

	// Create an empty file. It doesn't matter that it's an invalid zip, since
	// we're mocking the unzip command anyway.
	f, err := ioutil.TempFile(os.TempDir(), "volume-workspace-*")
	if err != nil {
		t.Fatal(err)
	}
	zip := f.Name()
	f.Close()
	defer os.Remove(zip)

	wc := &dockerVolumeWorkspaceCreator{}

	// We'll set up a fake repository with just enough fields defined for init()
	// and friends.
	repo := &graphql.Repository{
		DefaultBranch: &graphql.Branch{Name: "main"},
	}

	t.Run("success", func(t *testing.T) {
		expect.Commands(
			t,
			expect.NewGlob(
				expect.Behaviour{Stdout: []byte(volumeID)},
				"docker", "volume", "create",
			),
			expect.NewGlob(
				expect.Behaviour{},
				"docker", "run", "--rm", "--init", "--workdir", "/work",
				"--mount", "type=bind,source=*,target=/tmp/zip,ro",
				"--mount", "type=volume,source="+volumeID+",target=/work",
				dockerWorkspaceImage,
				"unzip", "/tmp/zip",
			),
			expect.NewGlob(
				expect.Behaviour{},
				"docker", "run", "--rm", "--init", "--workdir", "/work",
				"--mount", "type=bind,source=*,target=/run.sh,ro",
				"--mount", "type=volume,source="+volumeID+",target=/work",
				dockerWorkspaceImage,
				"sh", "/run.sh",
			),
		)

		if w, err := wc.Create(ctx, repo, zip); err != nil {
			t.Errorf("unexpected error: %v", err)
		} else if have := w.(*dockerVolumeWorkspace).volume; have != volumeID {
			t.Errorf("unexpected volume: have=%q want=%q", have, volumeID)
		}
	})

	t.Run("docker volume create failure", func(t *testing.T) {
		expect.Commands(
			t,
			expect.NewGlob(
				expect.Behaviour{ExitCode: 1},
				"docker", "volume", "create",
			),
		)

		if _, err := wc.Create(ctx, repo, zip); err == nil {
			t.Error("unexpected nil error")
		}
	})

	t.Run("unzip failure", func(t *testing.T) {
		expect.Commands(
			t,
			expect.NewGlob(
				expect.Behaviour{Stdout: []byte(volumeID)},
				"docker", "volume", "create",
			),
			expect.NewGlob(
				expect.Behaviour{ExitCode: 1},
				"docker", "run", "--rm", "--init", "--workdir", "/work",
				"--mount", "type=bind,source=*,target=/tmp/zip,ro",
				"--mount", "type=volume,source="+volumeID+",target=/work",
				dockerWorkspaceImage,
				"unzip", "/tmp/zip",
			),
		)

		if _, err := wc.Create(ctx, repo, zip); err == nil {
			t.Error("unexpected nil error")
		}
	})

	t.Run("git init failure", func(t *testing.T) {
		expect.Commands(
			t,
			expect.NewGlob(
				expect.Behaviour{Stdout: []byte(volumeID)},
				"docker", "volume", "create",
			),
			expect.NewGlob(
				expect.Behaviour{},
				"docker", "run", "--rm", "--init", "--workdir", "/work",
				"--mount", "type=bind,source=*,target=/tmp/zip,ro",
				"--mount", "type=volume,source="+volumeID+",target=/work",
				dockerWorkspaceImage,
				"unzip", "/tmp/zip",
			),
			expect.NewGlob(
				expect.Behaviour{ExitCode: 1},
				"docker", "run", "--rm", "--init", "--workdir", "/work",
				"--mount", "type=bind,source=*,target=/run.sh,ro",
				"--mount", "type=volume,source="+volumeID+",target=/work",
				dockerWorkspaceImage,
				"sh", "/run.sh",
			),
		)

		if _, err := wc.Create(ctx, repo, zip); err == nil {
			t.Error("unexpected nil error")
		}
	})
}

func TestVolumeWorkspace_Close(t *testing.T) {
	ctx := context.Background()
	w := &dockerVolumeWorkspace{volume: volumeID}

	t.Run("success", func(t *testing.T) {
		expect.Commands(
			t,
			expect.NewGlob(
				expect.Behaviour{},
				"docker", "volume", "rm", volumeID,
			),
		)

		if err := w.Close(ctx); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("failure", func(t *testing.T) {
		expect.Commands(
			t,
			expect.NewGlob(
				expect.Behaviour{ExitCode: 1},
				"docker", "volume", "rm", volumeID,
			),
		)

		if err := w.Close(ctx); err == nil {
			t.Error("unexpected nil error")
		}
	})
}

func TestVolumeWorkspace_DockerRunOpts(t *testing.T) {
	ctx := context.Background()
	w := &dockerVolumeWorkspace{volume: "VOLUME"}

	want := []string{
		"--mount", "type=volume,source=VOLUME,target=TARGET",
	}
	have, err := w.DockerRunOpts(ctx, "TARGET")
	if err != nil {
		t.Errorf("unexpected error: %+v", err)
	}
	if diff := cmp.Diff(have, want); diff != "" {
		t.Errorf("unexpected options (-have +want):\n%s", diff)
	}
}

func TestVolumeWorkspace_WorkDir(t *testing.T) {
	if have := (&dockerVolumeWorkspace{}).WorkDir(); have != nil {
		t.Errorf("unexpected work dir: %q", *have)
	}
}

// For the below tests that essentially delegate to runScript, we're not going
// to test the content of the script file: we'll do that as a one off test at
// the bottom of runScript itself, rather than depending on script content that
// may drift over time.

func TestVolumeWorkspace_Changes(t *testing.T) {
	ctx := context.Background()
	w := &dockerVolumeWorkspace{volume: volumeID}

	t.Run("success", func(t *testing.T) {
		for name, tc := range map[string]struct {
			stdout string
			want   *StepChanges
		}{
			"empty": {
				stdout: "",
				want:   &StepChanges{},
			},
			"valid": {
				stdout: `
M  go.mod
M  internal/campaigns/volume_workspace.go
M  internal/campaigns/volume_workspace_test.go				
				`,
				want: &StepChanges{Modified: []string{
					"go.mod",
					"internal/campaigns/volume_workspace.go",
					"internal/campaigns/volume_workspace_test.go",
				}},
			},
		} {
			t.Run(name, func(t *testing.T) {
				expect.Commands(
					t,
					expect.NewGlob(
						expect.Behaviour{Stdout: bytes.TrimSpace([]byte(tc.stdout))},
						"docker", "run", "--rm", "--init", "--workdir", "/work",
						"--mount", "type=bind,source=*,target=/run.sh,ro",
						"--mount", "type=volume,source="+volumeID+",target=/work",
						dockerWorkspaceImage,
						"sh", "/run.sh",
					),
				)

				have, err := w.Changes(ctx)
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}

				if diff := cmp.Diff(have, tc.want); diff != "" {
					t.Errorf("unexpected changes (-have +want):\n\n%s", diff)
				}

			})
		}
	})

	t.Run("failure", func(t *testing.T) {
		for name, behaviour := range map[string]expect.Behaviour{
			"exit code":        {ExitCode: 1},
			"malformed status": {Stdout: []byte("Z")},
		} {
			t.Run(name, func(t *testing.T) {
				expect.Commands(
					t,
					expect.NewGlob(
						behaviour,
						"docker", "run", "--rm", "--init", "--workdir", "/work",
						"--mount", "type=bind,source=*,target=/run.sh,ro",
						"--mount", "type=volume,source="+volumeID+",target=/work",
						dockerWorkspaceImage,
						"sh", "/run.sh",
					),
				)

				if _, err := w.Changes(ctx); err == nil {
					t.Error("unexpected nil error")
				}
			})
		}
	})
}

func TestVolumeWorkspace_Diff(t *testing.T) {
	ctx := context.Background()
	w := &dockerVolumeWorkspace{volume: volumeID}

	t.Run("success", func(t *testing.T) {
		for name, tc := range map[string]string{
			"empty": "",
			"valid": `
diff --git a/go.mod b/go.mod
index 06471f4..5f9d3fa 100644
--- a/go.mod
+++ b/go.mod
@@ -7,6 +7,7 @@ require (
		github.com/alessio/shellescape v1.4.1
		github.com/dustin/go-humanize v1.0.0
		github.com/efritz/pentimento v0.0.0-20190429011147-ade47d831101
+       github.com/gobwas/glob v0.2.3
		github.com/google/go-cmp v0.5.2
		github.com/hashicorp/errwrap v1.1.0 // indirect
		github.com/hashicorp/go-multierror v1.1.0
			`,
		} {
			t.Run(name, func(t *testing.T) {
				want := strings.TrimSpace(tc)

				expect.Commands(
					t,
					expect.NewGlob(
						expect.Behaviour{Stdout: []byte(want)},
						"docker", "run", "--rm", "--init", "--workdir", "/work",
						"--mount", "type=bind,source=*,target=/run.sh,ro",
						"--mount", "type=volume,source="+volumeID+",target=/work",
						dockerWorkspaceImage,
						"sh", "/run.sh",
					),
				)

				have, err := w.Diff(ctx)
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}

				if diff := cmp.Diff(string(have), want); diff != "" {
					t.Errorf("unexpected changes (-have +want):\n\n%s", diff)
				}

			})
		}
	})

	t.Run("failure", func(t *testing.T) {
		expect.Commands(
			t,
			expect.NewGlob(
				expect.Behaviour{ExitCode: 1},
				"docker", "run", "--rm", "--init", "--workdir", "/work",
				"--mount", "type=bind,source=*,target=/run.sh,ro",
				"--mount", "type=volume,source="+volumeID+",target=/work",
				dockerWorkspaceImage,
				"sh", "/run.sh",
			),
		)

		if _, err := w.Diff(ctx); err == nil {
			t.Error("unexpected nil error")
		}
	})
}

func TestVolumeWorkspace_runScript(t *testing.T) {
	// Since the above tests have thoroughly tested our error handling, this
	// test just fills in the one logical gap we have in our test coverage: is
	// the temporary script file correct?
	const script = "#!/bin/sh\n\necho FOO"
	ctx := context.Background()
	w := &dockerVolumeWorkspace{volume: volumeID}

	expect.Commands(
		t,
		&expect.Expectation{
			Validator: func(name string, arg ...string) error {
				// Run normal glob validation of the command line.
				glob := expect.NewGlobValidator(
					"docker", "run", "--rm", "--init", "--workdir", "/work",
					"--mount", "type=bind,source=*,target=/run.sh,ro",
					"--mount", "type=volume,source="+volumeID+",target=/work",
					dockerWorkspaceImage,
					"sh", "/run.sh",
				)
				if err := glob(name, arg...); err != nil {
					return err
				}

				// OK, we know that the temporary file the script lives in can
				// be parsed out of the seventh parameter, which provides the
				// mount options. Let's go get it!
				values := strings.Split(arg[6], ",")
				source := strings.SplitN(values[1], "=", 2)
				have, err := ioutil.ReadFile(source[1])
				if err != nil {
					return errors.Errorf("error reading temporary file %q: %v", source[1], err)
				}

				if string(have) != script {
					return errors.Errorf("unexpected script: have=%q want=%q", string(have), script)
				}
				return nil
			},
		},
	)

	if _, err := w.runScript(ctx, "/work", script); err != nil {
		t.Fatal(err)
	}
}

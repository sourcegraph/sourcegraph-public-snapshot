package workspace

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/sourcegraph/lib/errors"

	batcheslib "github.com/sourcegraph/sourcegraph/lib/batches"

	"github.com/sourcegraph/src-cli/internal/batches/docker"
	"github.com/sourcegraph/src-cli/internal/batches/graphql"
	"github.com/sourcegraph/src-cli/internal/batches/mock"
	"github.com/sourcegraph/src-cli/internal/exec/expect"
)

// We may as well use the same volume ID in all the subtests.
const volumeID = "VOLUME-ID"

func TestVolumeWorkspaceCreator(t *testing.T) {
	ctx := context.Background()

	// Create an empty file. It doesn't matter that it's an invalid zip, since
	// we're mocking the unzip command anyway.
	f, err := os.CreateTemp(os.TempDir(), "volume-workspace-*")
	if err != nil {
		t.Fatal(err)
	}
	defaultArchive := &fakeRepoArchive{mockPath: f.Name()}
	f.Close()
	defer os.Remove(defaultArchive.Path())

	archiveWithAdditionalFiles := &fakeRepoArchive{
		mockPath:                f.Name(),
		mockAdditionalFilePaths: map[string]string{},
	}
	for _, name := range []string{".gitignore", "another-file"} {
		// Since we don't read the files and mock the Docker commands,
		// we don't need to create them.
		path := filepath.Join(os.TempDir(), "additional-file"+name)
		// Instead we create a real-looking path that we sanitize so
		// it doesn't trip up the globbing expecations below:
		path = strings.ReplaceAll(path, string(os.PathSeparator), "-")

		archiveWithAdditionalFiles.mockAdditionalFilePaths[name] = path
	}

	wc := &dockerVolumeWorkspaceCreator{}
	// We'll set up a fake repository with just enough fields defined for init()
	// and friends.
	repo := &graphql.Repository{
		DefaultBranch: &graphql.Branch{Name: "main"},
	}

	for name, tc := range map[string]struct {
		archive      *fakeRepoArchive
		expectations []*expect.Expectation
		steps        []batcheslib.Step
		imageEnsurer imageEnsurer
		wantErr      bool
	}{
		"no steps": {
			expectations: []*expect.Expectation{
				expect.NewGlob(
					expect.Behaviour{Stdout: []byte(volumeID)},
					"docker", "volume", "create",
				),
				expect.NewGlob(
					expect.Success,
					"docker", "run", "--rm", "--init", "--workdir", "/work",
					"--user", "0:0",
					"--mount", "type=volume,source="+volumeID+",target=/work",
					DockerVolumeWorkspaceImage,
					"sh", "-c", "touch /work/*; chown -R 0:0 /work",
				),
				expect.NewGlob(
					expect.Success,
					"docker", "run", "--rm", "--init",
					"--workdir", "/work",
					"--mount", "type=bind,source=*,target=/tmp/zip,ro",
					"--user", "0:0",
					"--mount", "type=volume,source="+volumeID+",target=/work",
					DockerVolumeWorkspaceImage,
					"sh", "-c", "unzip /tmp/zip; rm /work/*",
				),
				expect.NewGlob(
					expect.Success,
					"docker", "run", "--rm", "--init", "--workdir", "/work",
					"--mount", "type=bind,source=*,target=/run.sh,ro",
					"--user", "0:0",
					"--mount", "type=volume,source="+volumeID+",target=/work",
					DockerVolumeWorkspaceImage,
					"sh", "/run.sh",
				),
			},
			steps:        []batcheslib.Step{},
			imageEnsurer: func(_ context.Context, _ string) (docker.Image, error) { return nil, nil },
		},
		"one root:root step": {
			expectations: []*expect.Expectation{
				expect.NewGlob(
					expect.Behaviour{Stdout: []byte(volumeID)},
					"docker", "volume", "create",
				),
				expect.NewGlob(
					expect.Success,
					"docker", "run", "--rm", "--init", "--workdir", "/work",
					"--user", "0:0",
					"--mount", "type=volume,source="+volumeID+",target=/work",
					DockerVolumeWorkspaceImage,
					"sh", "-c", "touch /work/*; chown -R 0:0 /work",
				),
				expect.NewGlob(
					expect.Success,
					"docker", "run", "--rm", "--init",
					"--workdir", "/work",
					"--mount", "type=bind,source=*,target=/tmp/zip,ro",
					"--user", "0:0",
					"--mount", "type=volume,source="+volumeID+",target=/work",
					DockerVolumeWorkspaceImage,
					"sh", "-c", "unzip /tmp/zip; rm /work/*",
				),
				expect.NewGlob(
					expect.Success,
					"docker", "run", "--rm", "--init", "--workdir", "/work",
					"--mount", "type=bind,source=*,target=/run.sh,ro",
					"--user", "0:0",
					"--mount", "type=volume,source="+volumeID+",target=/work",
					DockerVolumeWorkspaceImage,
					"sh", "/run.sh",
				),
			},
			steps: []batcheslib.Step{
				{},
			},
			imageEnsurer: func(_ context.Context, _ string) (docker.Image, error) {
				return &mock.Image{UidGid: docker.Root}, nil
			},
		},
		"one user:user step": {
			expectations: []*expect.Expectation{
				expect.NewGlob(
					expect.Behaviour{Stdout: []byte(volumeID)},
					"docker", "volume", "create",
				),
				expect.NewGlob(
					expect.Success,
					"docker", "run", "--rm", "--init", "--workdir", "/work",
					"--user", "0:0",
					"--mount", "type=volume,source="+volumeID+",target=/work",
					DockerVolumeWorkspaceImage,
					"sh", "-c", "touch /work/*; chown -R 1:2 /work",
				),
				expect.NewGlob(
					expect.Success,
					"docker", "run", "--rm", "--init",
					"--workdir", "/work",
					"--mount", "type=bind,source=*,target=/tmp/zip,ro",
					"--user", "1:2",
					"--mount", "type=volume,source="+volumeID+",target=/work",
					DockerVolumeWorkspaceImage,
					"sh", "-c", "unzip /tmp/zip; rm /work/*",
				),
				expect.NewGlob(
					expect.Success,
					"docker", "run", "--rm", "--init", "--workdir", "/work",
					"--mount", "type=bind,source=*,target=/run.sh,ro",
					"--user", "1:2",
					"--mount", "type=volume,source="+volumeID+",target=/work",
					DockerVolumeWorkspaceImage,
					"sh", "/run.sh",
				),
			},
			steps: []batcheslib.Step{
				{},
			},
			imageEnsurer: func(_ context.Context, _ string) (docker.Image, error) {
				return &mock.Image{UidGid: docker.UIDGID{UID: 1, GID: 2}}, nil
			},
		},
		"docker volume create failure": {
			expectations: []*expect.Expectation{
				expect.NewGlob(
					expect.Behaviour{ExitCode: 1},
					"docker", "volume", "create",
				),
			},
			steps: []batcheslib.Step{
				{},
			},
			imageEnsurer: func(_ context.Context, _ string) (docker.Image, error) {
				return &mock.Image{UidGid: docker.Root}, nil
			},
			wantErr: true,
		},
		"chown failure": {
			expectations: []*expect.Expectation{
				expect.NewGlob(
					expect.Behaviour{Stdout: []byte(volumeID)},
					"docker", "volume", "create",
				),
				expect.NewGlob(
					expect.Behaviour{ExitCode: 1},
					"docker", "run", "--rm", "--init", "--workdir", "/work",
					"--user", "0:0",
					"--mount", "type=volume,source="+volumeID+",target=/work",
					DockerVolumeWorkspaceImage,
					"sh", "-c", "touch /work/*; chown -R 0:0 /work",
				),
				expect.NewGlob(
					expect.Behaviour{ExitCode: 0},
					"docker", "volume", "rm", volumeID,
				),
			},
			steps: []batcheslib.Step{
				{},
			},
			imageEnsurer: func(_ context.Context, _ string) (docker.Image, error) {
				return &mock.Image{UidGid: docker.Root}, nil
			},
			wantErr: true,
		},
		"git init failure": {
			expectations: []*expect.Expectation{
				expect.NewGlob(
					expect.Behaviour{Stdout: []byte(volumeID)},
					"docker", "volume", "create",
				),
				expect.NewGlob(
					expect.Success,
					"docker", "run", "--rm", "--init", "--workdir", "/work",
					"--user", "0:0",
					"--mount", "type=volume,source="+volumeID+",target=/work",
					DockerVolumeWorkspaceImage,
					"sh", "-c", "touch /work/*; chown -R 0:0 /work",
				),
				expect.NewGlob(
					expect.Success,
					"docker", "run", "--rm", "--init",
					"--workdir", "/work",
					"--mount", "type=bind,source=*,target=/tmp/zip,ro",
					"--user", "0:0",
					"--mount", "type=volume,source="+volumeID+",target=/work",
					DockerVolumeWorkspaceImage,
					"sh", "-c", "unzip /tmp/zip; rm /work/*",
				),
				expect.NewGlob(
					expect.Behaviour{ExitCode: 1},
					"docker", "run", "--rm", "--init", "--workdir", "/work",
					"--mount", "type=bind,source=*,target=/run.sh,ro",
					"--user", "0:0",
					"--mount", "type=volume,source="+volumeID+",target=/work",
					DockerVolumeWorkspaceImage,
					"sh", "/run.sh",
				),
				expect.NewGlob(
					expect.Behaviour{ExitCode: 0},
					"docker", "volume", "rm", volumeID,
				),
			},
			steps: []batcheslib.Step{
				{},
			},
			imageEnsurer: func(_ context.Context, _ string) (docker.Image, error) {
				return &mock.Image{UidGid: docker.Root}, nil
			},
			wantErr: true,
		},
		"unzip failure": {
			expectations: []*expect.Expectation{
				expect.NewGlob(
					expect.Behaviour{Stdout: []byte(volumeID)},
					"docker", "volume", "create",
				),
				expect.NewGlob(
					expect.Success,
					"docker", "run", "--rm", "--init", "--workdir", "/work",
					"--user", "0:0",
					"--mount", "type=volume,source="+volumeID+",target=/work",
					DockerVolumeWorkspaceImage,
					"sh", "-c", "touch /work/*; chown -R 0:0 /work",
				),
				expect.NewGlob(
					expect.Behaviour{ExitCode: 1},
					"docker", "run", "--rm", "--init",
					"--workdir", "/work",
					"--mount", "type=bind,source=*,target=/tmp/zip,ro",
					"--user", "0:0",
					"--mount", "type=volume,source="+volumeID+",target=/work",
					DockerVolumeWorkspaceImage,
					"sh", "-c", "unzip /tmp/zip; rm /work/*",
				),
				expect.NewGlob(
					expect.Behaviour{ExitCode: 0},
					"docker", "volume", "rm", volumeID,
				),
			},
			steps: []batcheslib.Step{
				{},
			},
			imageEnsurer: func(_ context.Context, _ string) (docker.Image, error) {
				return &mock.Image{UidGid: docker.Root}, nil
			},
			wantErr: true,
		},
		"additional files": {
			archive: archiveWithAdditionalFiles,
			expectations: []*expect.Expectation{
				expect.NewGlob(
					expect.Behaviour{Stdout: []byte(volumeID)},
					"docker", "volume", "create",
				),
				expect.NewGlob(
					expect.Success,
					"docker", "run", "--rm", "--init", "--workdir", "/work",
					"--user", "0:0",
					"--mount", "type=volume,source="+volumeID+",target=/work",
					DockerVolumeWorkspaceImage,
					"sh", "-c", "touch /work/*; chown -R 0:0 /work",
				),
				expect.NewGlob(
					expect.Success,
					"docker", "run", "--rm", "--init",
					"--workdir", "/work",
					"--mount", "type=bind,source=*,target=/tmp/zip,ro",
					"--user", "0:0",
					"--mount", "type=volume,source="+volumeID+",target=/work",
					DockerVolumeWorkspaceImage,
					"sh", "-c", "unzip /tmp/zip; rm /work/*",
				),
				expect.NewGlob(
					expect.Success,
					"docker", "run", "--rm", "--init",
					"--workdir", "/work",
					"--user", "0:0",
					"--mount", "type=volume,source="+volumeID+",target=/work",
					"--mount", "type=bind,source="+archiveWithAdditionalFiles.mockAdditionalFilePaths[".gitignore"]+",target=/tmp/.gitignore,ro",
					"--mount", "type=bind,source="+archiveWithAdditionalFiles.mockAdditionalFilePaths["another-file"]+",target=/tmp/another-file,ro",
					DockerVolumeWorkspaceImage,
					"sh", "-c", "cp /tmp/.gitignore /work/.gitignore && cp /tmp/another-file /work/another-file;",
				),
				expect.NewGlob(
					expect.Success,
					"docker", "run", "--rm", "--init", "--workdir", "/work",
					"--mount", "type=bind,source=*,target=/run.sh,ro",
					"--user", "0:0",
					"--mount", "type=volume,source="+volumeID+",target=/work",
					DockerVolumeWorkspaceImage,
					"sh", "/run.sh",
				),
			},
			steps: []batcheslib.Step{
				{},
			},
			imageEnsurer: func(_ context.Context, _ string) (docker.Image, error) {
				return &mock.Image{UidGid: docker.Root}, nil
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			expect.Commands(t, tc.expectations...)
			a := defaultArchive
			if tc.archive != nil {
				a = tc.archive
			}

			wc.EnsureImage = tc.imageEnsurer
			w, err := wc.Create(ctx, repo, tc.steps, a)
			if tc.wantErr {
				if err == nil {
					t.Error("unexpected nil error")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				} else if have := w.(*dockerVolumeWorkspace).volume; have != volumeID {
					t.Errorf("unexpected volume: have=%q want=%q", have, volumeID)
				}
			}
		})
	}
}

func TestVolumeWorkspace_Close(t *testing.T) {
	ctx := context.Background()
	w := &dockerVolumeWorkspace{volume: volumeID}

	t.Run("success", func(t *testing.T) {
		expect.Commands(
			t,
			expect.NewGlob(
				expect.Success,
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
	w := &dockerVolumeWorkspace{
		volume: "VOLUME",
		uidGid: docker.UIDGID{UID: 1, GID: 2},
	}

	want := []string{
		"--user", "1:2",
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
						"--user", "0:0",
						"--mount", "type=volume,source="+volumeID+",target=/work",
						DockerVolumeWorkspaceImage,
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
				"--user", "0:0",
				"--mount", "type=volume,source="+volumeID+",target=/work",
				DockerVolumeWorkspaceImage,
				"sh", "/run.sh",
			),
		)

		if _, err := w.Diff(ctx); err == nil {
			t.Error("unexpected nil error")
		}
	})
}

func TestVolumeWorkspace_ApplyDiff(t *testing.T) {
	ctx := context.Background()
	w := &dockerVolumeWorkspace{volume: volumeID}

	expect.Commands(
		t,
		expect.NewGlob(
			expect.Behaviour{ExitCode: 0},
			"docker", "run", "--rm", "--init", "--workdir", "/work",
			"--mount", "type=bind,source=*,target=/run.sh,ro",
			"--user", "0:0",
			"--mount", "type=volume,source="+volumeID+",target=/work",
			DockerVolumeWorkspaceImage,
			"sh", "/run.sh",
		),
	)

	err := w.ApplyDiff(ctx, []byte(`dummydiff`))
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
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
					"--user", "0:0",
					"--mount", "type=volume,source="+volumeID+",target=/work",
					DockerVolumeWorkspaceImage,
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
				have, err := os.ReadFile(source[1])
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

package service

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/lib/errors"

	batcheslib "github.com/sourcegraph/sourcegraph/lib/batches"

	"github.com/sourcegraph/src-cli/internal/batches/docker"
	"github.com/sourcegraph/src-cli/internal/batches/graphql"
	"github.com/sourcegraph/src-cli/internal/batches/mock"
)

func TestService_ValidateChangesetSpecs(t *testing.T) {
	repo1 := &graphql.Repository{ID: "repo-graphql-id-1", Name: "github.com/sourcegraph/src-cli"}
	repo2 := &graphql.Repository{ID: "repo-graphql-id-2", Name: "github.com/sourcegraph/sourcegraph"}

	tests := map[string]struct {
		repos []*graphql.Repository
		specs []*batcheslib.ChangesetSpec

		wantErrInclude string
	}{
		"no errors": {
			repos: []*graphql.Repository{repo1, repo2},
			specs: []*batcheslib.ChangesetSpec{
				{
					HeadRepository: repo1.ID, HeadRef: "refs/heads/branch-1",
				},
				{
					HeadRepository: repo1.ID, HeadRef: "refs/heads/branch-2",
				},
				{
					HeadRepository: repo2.ID, HeadRef: "refs/heads/branch-1",
				},
				{
					HeadRepository: repo2.ID, HeadRef: "refs/heads/branch-2",
				},
			},
		},

		"imported changeset": {
			repos: []*graphql.Repository{repo1},
			specs: []*batcheslib.ChangesetSpec{
				{
					ExternalID: "123",
				},
			},
			// This should not fail validation ever.
		},

		"duplicate branches": {
			repos: []*graphql.Repository{repo1, repo2},
			specs: []*batcheslib.ChangesetSpec{
				{
					HeadRepository: repo1.ID, HeadRef: "refs/heads/branch-1",
				},
				{
					HeadRepository: repo1.ID, HeadRef: "refs/heads/branch-2",
				},
				{
					HeadRepository: repo2.ID, HeadRef: "refs/heads/branch-1",
				},
				{
					HeadRepository: repo2.ID, HeadRef: "refs/heads/branch-1",
				},
			},
			wantErrInclude: `github.com/sourcegraph/sourcegraph: 2 changeset specs have the branch "branch-1"`,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			svc := &Service{}
			haveErr := svc.ValidateChangesetSpecs(tt.repos, tt.specs)
			if tt.wantErrInclude != "" {
				if haveErr == nil {
					t.Fatalf("expected %q to be included in error, but got none", tt.wantErrInclude)
				} else if !strings.Contains(haveErr.Error(), tt.wantErrInclude) {
					t.Fatalf("expected %q to be included in error, but was not. error=%q", tt.wantErrInclude, haveErr.Error())
				}
			} else {
				if haveErr != nil {
					t.Fatalf("unexpected error: %s", haveErr)
				}
			}
		})
	}
}

func TestEnsureDockerImages(t *testing.T) {
	ctx := context.Background()
	parallelCases := []int{0, 1, 2, 4, 8}

	svc := &Service{}

	t.Run("success", func(t *testing.T) {
		t.Run("single image", func(t *testing.T) {
			// A zeroed mock.Image should be usable for testing purposes.
			image := &mock.Image{}
			images := map[string]docker.Image{
				"image": image,
			}

			for name, steps := range map[string][]batcheslib.Step{
				"single step":    {{Container: "image"}},
				"multiple steps": {{Container: "image"}, {Container: "image"}},
			} {
				t.Run(name, func(t *testing.T) {
					for _, parallelism := range parallelCases {
						t.Run(fmt.Sprintf("%d worker(s)", parallelism), func(t *testing.T) {
							progress := &mock.Progress{}
							have, err := svc.EnsureDockerImages(ctx, &mock.ImageCache{Images: images}, steps, parallelism, progress.Callback())
							assert.Nil(t, err)
							assert.Equal(t, images, have)
							assert.Equal(t, []mock.ProgressCall{
								{Done: 0, Total: 1},
								{Done: 1, Total: 1},
							}, progress.Calls)
						})
					}
				})
			}
		})

		t.Run("multiple images", func(t *testing.T) {
			var (
				imageA = &mock.Image{}
				imageB = &mock.Image{}
				imageC = &mock.Image{}
				images = map[string]docker.Image{
					"a": imageA,
					"b": imageB,
					"c": imageC,
				}
			)

			for _, parallelism := range parallelCases {
				t.Run(fmt.Sprintf("%d worker(s)", parallelism), func(t *testing.T) {
					progress := &mock.Progress{}

					have, err := svc.EnsureDockerImages(
						ctx,
						&mock.ImageCache{Images: images},
						[]batcheslib.Step{
							{Container: "a"},
							{Container: "a"},
							{Container: "a"},
							{Container: "b"},
							{Container: "c"},
						},
						parallelism,
						progress.Callback(),
					)
					assert.Nil(t, err)
					assert.Equal(t, images, have)
					assert.Equal(t, []mock.ProgressCall{
						{Done: 0, Total: 3},
						{Done: 1, Total: 3},
						{Done: 2, Total: 3},
						{Done: 3, Total: 3},
					}, progress.Calls)
				})
			}
		})
	})

	t.Run("errors", func(t *testing.T) {
		// The only really interesting case is where an image fails — we want to
		// ensure that the error is propagated, and that we don't end up
		// deadlocking while the context cancellation propagates. Let's set up a
		// good number of images (and steps) so we can give this a good test.
		wantErr := errors.New("expected error")
		images := map[string]docker.Image{}
		steps := []batcheslib.Step{}

		total := 100
		for i := 0; i < total; i++ {
			name := strconv.Itoa(i)
			if i%25 == 0 {
				images[name] = &mock.Image{EnsureErr: wantErr}
			} else {
				images[name] = &mock.Image{}
			}
			for j := 0; j < (i%10)+1; j++ {
				steps = append(steps, batcheslib.Step{Container: name})
			}
		}

		// Just verify we did that right!
		assert.Len(t, images, total)
		assert.True(t, len(steps) > total)

		for _, parallelism := range parallelCases {
			t.Run(fmt.Sprintf("%d worker(s)", parallelism), func(t *testing.T) {
				progress := &mock.Progress{}

				have, err := svc.EnsureDockerImages(ctx, &mock.ImageCache{Images: images}, steps, parallelism, progress.Callback())
				assert.ErrorIs(t, err, wantErr)
				assert.Nil(t, have)

				// Because there's no particular order the images will be fetched in,
				// the number of progress calls we get is non-deterministic, other than
				// that we should always get the first one.
				assert.Equal(t, mock.ProgressCall{Done: 0, Total: total}, progress.Calls[0])
			})
		}

	})
}

func TestService_ParseBatchSpec(t *testing.T) {
	svc := &Service{}

	tempDir := t.TempDir()
	tempOutsideDir := t.TempDir()
	// create temp files/dirs that can be used by the tests
	_, err := os.Create(filepath.Join(tempDir, "sample.sh"))
	require.NoError(t, err)
	_, err = os.Create(filepath.Join(tempDir, "another.sh"))
	require.NoError(t, err)

	tests := []struct {
		name         string
		batchSpecDir string
		rawSpec      string
		expectedSpec *batcheslib.BatchSpec
		expectedErr  error
	}{
		{
			name: "simple spec",
			rawSpec: `
name: test-spec
description: A test spec
`,
			expectedSpec: &batcheslib.BatchSpec{Name: "test-spec", Description: "A test spec"},
		},
		{
			name: "unknown field",
			rawSpec: `
name: test-spec
description: A test spec
some-new-field: Foo bar
`,
			expectedErr: errors.New("parsing batch spec: Additional property some-new-field is not allowed"),
		},
		{
			name:         "mount absolute file",
			batchSpecDir: tempDir,
			rawSpec: fmt.Sprintf(`
name: test-spec
description: A test spec
steps:
  - run: /tmp/sample.sh
    container: alpine:3
    mount:
      - path: %s
        mountpoint: /tmp/sample.sh
changesetTemplate:
  title: Test Mount
  body: Test a mounted path
  branch: test
  commit:
    message: Test
`, filepath.Join(tempDir, "sample.sh")),
			expectedSpec: &batcheslib.BatchSpec{
				Name:        "test-spec",
				Description: "A test spec",
				Steps: []batcheslib.Step{
					{
						Run:       "/tmp/sample.sh",
						Container: "alpine:3",
						Mount: []batcheslib.Mount{
							{
								Path:       filepath.Join(tempDir, "sample.sh"),
								Mountpoint: "/tmp/sample.sh",
							},
						},
					},
				},
				ChangesetTemplate: &batcheslib.ChangesetTemplate{
					Title:  "Test Mount",
					Body:   "Test a mounted path",
					Branch: "test",
					Commit: batcheslib.ExpandedGitCommitDescription{
						Message: "Test",
					},
				},
			},
		},
		{
			name:         "mount relative file",
			batchSpecDir: tempDir,
			rawSpec: `
name: test-spec
description: A test spec
steps:
  - run: /tmp/sample.sh
    container: alpine:3
    mount:
      - path: ./sample.sh
        mountpoint: /tmp/sample.sh
changesetTemplate:
  title: Test Mount
  body: Test a mounted path
  branch: test
  commit:
    message: Test
`,
			expectedSpec: &batcheslib.BatchSpec{
				Name:        "test-spec",
				Description: "A test spec",
				Steps: []batcheslib.Step{
					{
						Run:       "/tmp/sample.sh",
						Container: "alpine:3",
						Mount: []batcheslib.Mount{
							{
								Path:       "./sample.sh",
								Mountpoint: "/tmp/sample.sh",
							},
						},
					},
				},
				ChangesetTemplate: &batcheslib.ChangesetTemplate{
					Title:  "Test Mount",
					Body:   "Test a mounted path",
					Branch: "test",
					Commit: batcheslib.ExpandedGitCommitDescription{
						Message: "Test",
					},
				},
			},
		},
		{
			name:         "mount absolute directory",
			batchSpecDir: tempDir,
			rawSpec: fmt.Sprintf(`
name: test-spec
description: A test spec
steps:
  - run: /tmp/some-file.sh
    container: alpine:3
    mount:
      - path: %s
        mountpoint: /tmp
changesetTemplate:
  title: Test Mount
  body: Test a mounted path
  branch: test
  commit:
    message: Test
`, tempDir),
			expectedSpec: &batcheslib.BatchSpec{
				Name:        "test-spec",
				Description: "A test spec",
				Steps: []batcheslib.Step{
					{
						Run:       "/tmp/some-file.sh",
						Container: "alpine:3",
						Mount: []batcheslib.Mount{
							{
								Path:       tempDir,
								Mountpoint: "/tmp",
							},
						},
					},
				},
				ChangesetTemplate: &batcheslib.ChangesetTemplate{
					Title:  "Test Mount",
					Body:   "Test a mounted path",
					Branch: "test",
					Commit: batcheslib.ExpandedGitCommitDescription{
						Message: "Test",
					},
				},
			},
		},
		{
			name:         "mount relative directory",
			batchSpecDir: tempDir,
			rawSpec: `
name: test-spec
description: A test spec
steps:
  - run: /tmp/some-file.sh
    container: alpine:3
    mount:
      - path: ./
        mountpoint: /tmp
changesetTemplate:
  title: Test Mount
  body: Test a mounted path
  branch: test
  commit:
    message: Test
`,
			expectedSpec: &batcheslib.BatchSpec{
				Name:        "test-spec",
				Description: "A test spec",
				Steps: []batcheslib.Step{
					{
						Run:       "/tmp/some-file.sh",
						Container: "alpine:3",
						Mount: []batcheslib.Mount{
							{
								Path:       "./",
								Mountpoint: "/tmp",
							},
						},
					},
				},
				ChangesetTemplate: &batcheslib.ChangesetTemplate{
					Title:  "Test Mount",
					Body:   "Test a mounted path",
					Branch: "test",
					Commit: batcheslib.ExpandedGitCommitDescription{
						Message: "Test",
					},
				},
			},
		},
		{
			name:         "mount multiple files",
			batchSpecDir: tempDir,
			rawSpec: `
name: test-spec
description: A test spec
steps:
  - run: /tmp/sample.sh && /tmp/another.sh
    container: alpine:3
    mount:
      - path: ./sample.sh
        mountpoint: /tmp/sample.sh
      - path: ./another.sh
        mountpoint: /tmp/another.sh
changesetTemplate:
  title: Test Mount
  body: Test a mounted path
  branch: test
  commit:
    message: Test
`,
			expectedSpec: &batcheslib.BatchSpec{
				Name:        "test-spec",
				Description: "A test spec",
				Steps: []batcheslib.Step{
					{
						Run:       "/tmp/sample.sh && /tmp/another.sh",
						Container: "alpine:3",
						Mount: []batcheslib.Mount{
							{
								Path:       "./sample.sh",
								Mountpoint: "/tmp/sample.sh",
							},
							{
								Path:       "./another.sh",
								Mountpoint: "/tmp/another.sh",
							},
						},
					},
				},
				ChangesetTemplate: &batcheslib.ChangesetTemplate{
					Title:  "Test Mount",
					Body:   "Test a mounted path",
					Branch: "test",
					Commit: batcheslib.ExpandedGitCommitDescription{
						Message: "Test",
					},
				},
			},
		},
		{
			name:         "mount path does not exist",
			batchSpecDir: tempDir,
			rawSpec: fmt.Sprintf(`
name: test-spec
description: A test spec
steps:
  - run: /tmp/sample.sh
    container: alpine:3
    mount:
      - path: %s
        mountpoint: /tmp
changesetTemplate:
  title: Test Mount
  body: Test a mounted path
  branch: test
  commit:
    message: Test
`, filepath.Join(tempDir, "does", "not", "exist", "sample.sh")),
			expectedErr: errors.Newf("handling mount: step 1 mount path %s does not exist", filepath.Join(tempDir, "does", "not", "exist", "sample.sh")),
		},
		{
			name:         "mount path not subdirectory of spec",
			batchSpecDir: tempDir,
			rawSpec: fmt.Sprintf(`
name: test-spec
description: A test spec
steps:
  - run: /tmp/sample.sh
    container: alpine:3
    mount:
      - path: %s
        mountpoint: /tmp
changesetTemplate:
  title: Test Mount
  body: Test a mounted path
  branch: test
  commit:
    message: Test
`, tempOutsideDir),
			expectedErr: errors.New("handling mount: step 1 mount path is not in the same directory or subdirectory as the batch spec"),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			spec, err := svc.ParseBatchSpec(test.batchSpecDir, []byte(test.rawSpec))
			if test.expectedErr != nil {
				assert.Equal(t, test.expectedErr.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.expectedSpec, spec)
			}
		})
	}
}

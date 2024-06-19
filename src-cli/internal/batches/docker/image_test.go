package docker

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"

	"github.com/sourcegraph/src-cli/internal/exec/expect"
)

func TestImage_Digest(t *testing.T) {
	ctx := context.Background()

	for name, tc := range map[string]struct {
		expectations []*expect.Expectation
		image        *image
		want         string
		wantErr      bool
	}{
		"success": {
			expectations: []*expect.Expectation{inspectSuccess("foo", "digest")},
			image:        &image{name: "foo"},
			want:         "digest",
		},
		"inspect invalid output": {
			expectations: []*expect.Expectation{
				inspectSuccess("foo", ""),
			},
			image:   &image{name: "foo"},
			wantErr: true,
		},
		"inspect failure first attempt": {
			expectations: []*expect.Expectation{
				inspectFailure("foo"),
				pullSuccess("foo"),
				inspectSuccess("foo", "digest"),
			},
			image: &image{name: "foo"},
			want:  "digest",
		},
		"pull failure": {
			expectations: []*expect.Expectation{
				inspectFailure("foo"),
				pullFailure("foo"),
			},
			image:   &image{name: "foo"},
			wantErr: true,
		},
	} {
		t.Run(name, func(t *testing.T) {
			expect.Commands(t, tc.expectations...)

			// We'll call Digest twice to make sure the memoisation works.
			test := func() (string, error) {
				have, err := tc.image.Digest(ctx)
				if tc.wantErr {
					if err == nil {
						t.Error("unexpected nil error")
					}
				} else if err != nil {
					t.Errorf("unexpected error: %+v", err)
				} else if have != tc.want {
					t.Errorf("unexpected digest: have=%q want=%q", have, tc.want)
				}

				return have, err
			}
			firstDigest, firstErr := test()
			secondDigest, secondErr := test()

			if firstDigest != secondDigest {
				t.Errorf("digests do not match: first=%q second=%q", firstDigest, secondDigest)
			}
			if firstErr != secondErr {
				t.Errorf("errors do not match: first=%v second=%v", firstErr, secondErr)
			}
		})
	}

	t.Run("docker timeout", func(t *testing.T) {
		// We're more interested in the context error handling than the parsing
		// of the timeout, and we don't want to slow down the test, so we're
		// going to construct a context that has already exceeded its deadline
		// at the point it is provided to Digest.
		ctx, cancel := context.WithTimeout(context.Background(), -1*time.Second)
		t.Cleanup(cancel)

		expect.Commands(t, inspectSuccess("foo", ""))
		image := &image{name: "foo"}

		digest, err := image.Digest(ctx)
		assert.Empty(t, digest)

		as := &fastCommandTimeoutError{}
		assert.ErrorAs(t, err, &as)
		assert.Equal(t, "foo", as.args[len(as.args)-1])
		assert.Equal(t, fastCommandTimeoutDefault, as.timeout)
	})
}

func TestImage_Ensure(t *testing.T) {
	ctx := context.Background()

	for name, tc := range map[string]struct {
		expectations []*expect.Expectation
		image        *image
		wantErr      bool
	}{
		"no pull required": {
			expectations: []*expect.Expectation{inspectSuccess("foo", "digest")},
			image:        &image{name: "foo"},
			wantErr:      false,
		},
		"pull required": {
			expectations: []*expect.Expectation{
				inspectFailure("foo"),
				pullSuccess("foo"),
				inspectSuccess("foo", "digest"),
			},
			image:   &image{name: "foo"},
			wantErr: false,
		},
		"pull failed": {
			expectations: []*expect.Expectation{
				inspectFailure("foo"),
				pullFailure("foo"),
			},
			image:   &image{name: "foo"},
			wantErr: true,
		},
	} {
		t.Run(name, func(t *testing.T) {
			expect.Commands(t, tc.expectations...)

			// We'll call Ensure twice to make sure the memoisation works.
			test := func() error {
				have := tc.image.Ensure(ctx)
				if tc.wantErr {
					if have == nil {
						t.Error("unexpected nil error")
					}
				} else if have != nil {
					t.Errorf("unexpected error: %+v", have)
				}

				return have
			}
			first := test()
			second := test()

			if first != second {
				t.Errorf("errors do not match: first=%v second=%v", first, second)
			}
		})
	}
}

func TestImage_UIDGID(t *testing.T) {
	ctx := context.Background()

	for name, tc := range map[string]struct {
		expectations []*expect.Expectation
		image        *image
		want         UIDGID
		wantErr      bool
	}{
		"success": {
			expectations: []*expect.Expectation{
				inspectSuccess("foo", "bar"),
				uidGid("bar", expect.Behaviour{Stdout: []byte("1000\n2000\n")}),
			},
			image: &image{name: "foo"},
			want:  UIDGID{UID: 1000, GID: 2000},
		},
		// We should also make sure 0 works. Sometimes it's easy to miss. Just
		// ask the Romans.
		"success with zeroes": {
			expectations: []*expect.Expectation{
				inspectSuccess("foo", "bar"),
				uidGid("bar", expect.Behaviour{Stdout: []byte("0\n0\n")}),
			},
			image: &image{name: "foo"},
			want:  UIDGID{UID: 0, GID: 0},
		},
		// This is technically valid, because POSIX basically punts on the
		// signedness of pid_t and gid_t. We don't really have a reason to
		// disallow it.
		"success with negative IDs": {
			expectations: []*expect.Expectation{
				inspectSuccess("foo", "bar"),
				uidGid("bar", expect.Behaviour{Stdout: []byte("-1000\n-2000\n")}),
			},
			image: &image{name: "foo"},
			want:  UIDGID{UID: -1000, GID: -2000},
		},
		// This is technically invalid, but should still succeed. Postel's Law
		// and all that.
		"success without trailing newline": {
			expectations: []*expect.Expectation{
				inspectSuccess("foo", "bar"),
				uidGid("bar", expect.Behaviour{Stdout: []byte("1000\n2000")}),
			},
			image: &image{name: "foo"},
			want:  UIDGID{UID: 1000, GID: 2000},
		},
		// As above, this is invalid, but we should still handle it.
		"success with extra data": {
			expectations: []*expect.Expectation{
				inspectSuccess("foo", "bar"),
				uidGid("bar", expect.Behaviour{Stdout: []byte("1000\n2000\n3000\n")}),
			},
			image: &image{name: "foo"},
			want:  UIDGID{UID: 1000, GID: 2000},
		},
		// Now for some interesting failure cases.
		"invalid output": {
			expectations: []*expect.Expectation{
				inspectSuccess("foo", "bar"),
				uidGid("bar", expect.Behaviour{Stdout: []byte("")}),
			},
			image:   &image{name: "foo"},
			wantErr: true,
		},
		// This is ripped from the headlines^WDocker.
		"missing id binary": {
			expectations: []*expect.Expectation{
				inspectSuccess("foo", "bar"),
				uidGid("bar", expect.Behaviour{
					ExitCode: 127,
					Stderr:   []byte("sh: id: not found")}),
			},
			image:   &image{name: "foo"},
			wantErr: true,
		},
		// POSIX might allow negative IDs because, well, honestly, it was
		// probably an oversight, but we shouldn't allow string IDs. That would
		// be a bridge too far.
		"string uid": {
			expectations: []*expect.Expectation{
				inspectSuccess("foo", "bar"),
				uidGid("bar", expect.Behaviour{Stdout: []byte("X\n2000\n")}),
			},
			image:   &image{name: "foo"},
			wantErr: true,
		},
		"string gid": {
			expectations: []*expect.Expectation{
				inspectSuccess("foo", "bar"),
				uidGid("bar", expect.Behaviour{Stdout: []byte("1000\nX\n")}),
			},
			image:   &image{name: "foo"},
			wantErr: true,
		},
		// Now for some more run of the mill failures.
		"docker run failure": {
			expectations: []*expect.Expectation{
				inspectSuccess("foo", "bar"),
				uidGid("bar", expect.Behaviour{ExitCode: 1}),
			},
			image:   &image{name: "foo"},
			wantErr: true,
		},
		"inspect and pull failure": {
			expectations: []*expect.Expectation{
				inspectFailure("foo"),
				pullFailure("foo"),
			},
			image:   &image{name: "foo"},
			wantErr: true,
		},
	} {
		t.Run(name, func(t *testing.T) {
			expect.Commands(t, tc.expectations...)

			// We'll call UIDGID twice to make sure the memoisation works.
			test := func() (UIDGID, error) {
				have, err := tc.image.UIDGID(ctx)
				if tc.wantErr {
					if err == nil {
						t.Error("unexpected nil error")
					}
				} else if err != nil {
					t.Errorf("unexpected error: %+v", err)
				} else if diff := cmp.Diff(have, tc.want); diff != "" {
					t.Errorf("unexpected uid/gid (-have +want):\n%s", diff)
				}

				return have, err
			}
			firstUG, firstErr := test()
			secondUG, secondErr := test()

			if diff := cmp.Diff(firstUG, secondUG); diff != "" {
				t.Errorf("uid/gids do not match: (-first +second)\n%s", diff)
			}
			if firstErr != secondErr {
				t.Errorf("errors do not match: first=%v second=%v", firstErr, secondErr)
			}
		})
	}
}

func TestUIDGID(t *testing.T) {
	have := UIDGID{UID: 1000, GID: 0}.String()
	want := "1000:0"
	if have != want {
		t.Errorf("unexpected value: have=%q want=%q", have, want)
	}
}

// Set up some helper functions for expectations we'll be reusing.
func inspectSuccess(name, digest string) *expect.Expectation {
	return expect.NewGlob(
		expect.Behaviour{Stdout: []byte(digest + "\n")},
		// Note the awkward escaping because these arguments are
		// matched by glob.
		"docker", "image", "inspect", "--format", `\{\{ .Id }}`, name,
	)
}

func inspectFailure(name string) *expect.Expectation {
	return expect.NewGlob(
		// docker image inspect returns 1 for non-existent images.
		expect.Behaviour{ExitCode: 1},
		"docker", "image", "inspect", "--format", `\{\{ .Id }}`, name,
	)
}

func pullFailure(name string) *expect.Expectation {
	return expect.NewGlob(
		expect.Behaviour{ExitCode: 1},
		"docker", "image", "pull", name,
	)
}

func pullSuccess(name string) *expect.Expectation {
	return expect.NewGlob(
		expect.Behaviour{ExitCode: 0},
		"docker", "image", "pull", name,
	)
}

func uidGid(digest string, behaviour expect.Behaviour) *expect.Expectation {
	return expect.NewGlob(
		behaviour,
		"docker", "run", "--rm", "--entrypoint", "/bin/sh",
		digest, "-c", "id -u; id -g",
	)
}

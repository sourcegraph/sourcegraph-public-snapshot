package docker

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
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
			expectations: digestSuccess("foo", "bar"),
			image:        &image{name: "foo"},
			want:         "bar",
		},
		"inspect invalid output": {
			expectations: append(
				ensureSuccess("foo"),
				expect.NewGlob(
					expect.Success,
					// Note the awkward escaping because these arguments are matched by
					// glob.
					"docker", "image", "inspect", "--format", `\{\{.Id}}`, "--", "foo",
				),
			),
			image:   &image{name: "foo"},
			wantErr: true,
		},
		"inspect failure": {
			expectations: digestFailure("foo"),
			image:        &image{name: "foo"},
			wantErr:      true,
		},
		"ensure failure": {
			expectations: ensureFailure("foo"),
			image:        &image{name: "foo"},
			wantErr:      true,
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
}

func TestImage_Ensure(t *testing.T) {
	ctx := context.Background()

	for name, tc := range map[string]struct {
		expectations []*expect.Expectation
		image        *image
		wantErr      bool
	}{
		"no pull required": {
			expectations: ensureSuccess("foo"),
			image:        &image{name: "foo"},
			wantErr:      false,
		},
		"pull required": {
			expectations: []*expect.Expectation{
				expect.NewGlob(
					// docker image inspect returns 1 for non-existent images.
					expect.Behaviour{ExitCode: 1},
					"docker", "image", "inspect", "--format", "1", "foo",
				),
				expect.NewGlob(
					expect.Success,
					"docker", "image", "pull", "foo",
				),
			},
			image:   &image{name: "foo"},
			wantErr: false,
		},
		"pull failed": {
			expectations: ensureFailure("foo"),
			image:        &image{name: "foo"},
			wantErr:      true,
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
			expectations: append(
				digestSuccess("foo", "bar"),
				uidGid("bar", expect.Behaviour{Stdout: []byte("1000\n2000\n")}),
			),
			image: &image{name: "foo"},
			want:  UIDGID{UID: 1000, GID: 2000},
		},
		// We should also make sure 0 works. Sometimes it's easy to miss. Just
		// ask the Romans.
		"success with zeroes": {
			expectations: append(
				digestSuccess("foo", "bar"),
				uidGid("bar", expect.Behaviour{Stdout: []byte("0\n0\n")}),
			),
			image: &image{name: "foo"},
			want:  UIDGID{UID: 0, GID: 0},
		},
		// This is technically valid, because POSIX basically punts on the
		// signedness of pid_t and gid_t. We don't really have a reason to
		// disallow it.
		"success with negative IDs": {
			expectations: append(
				digestSuccess("foo", "bar"),
				uidGid("bar", expect.Behaviour{Stdout: []byte("-1000\n-2000\n")}),
			),
			image: &image{name: "foo"},
			want:  UIDGID{UID: -1000, GID: -2000},
		},
		// This is technically invalid, but should still succeed. Postel's Law
		// and all that.
		"success without trailing newline": {
			expectations: append(
				digestSuccess("foo", "bar"),
				uidGid("bar", expect.Behaviour{Stdout: []byte("1000\n2000")}),
			),
			image: &image{name: "foo"},
			want:  UIDGID{UID: 1000, GID: 2000},
		},
		// As above, this is invalid, but we should still handle it.
		"success with extra data": {
			expectations: append(
				digestSuccess("foo", "bar"),
				uidGid("bar", expect.Behaviour{Stdout: []byte("1000\n2000\n3000\n")}),
			),
			image: &image{name: "foo"},
			want:  UIDGID{UID: 1000, GID: 2000},
		},
		// Now for some interesting failure cases.
		"invalid output": {
			expectations: append(
				digestSuccess("foo", "bar"),
				uidGid("bar", expect.Behaviour{Stdout: []byte("")}),
			),
			image:   &image{name: "foo"},
			wantErr: true,
		},
		// This is ripped from the headlines^WDocker.
		"missing id binary": {
			expectations: append(
				digestSuccess("foo", "bar"),
				uidGid("bar", expect.Behaviour{
					ExitCode: 127,
					Stderr:   []byte("sh: id: not found")}),
			),
			image:   &image{name: "foo"},
			wantErr: true,
		},
		// POSIX might allow negative IDs because, well, honestly, it was
		// probably an oversight, but we shouldn't allow string IDs. That would
		// be a bridge too far.
		"string uid": {
			expectations: append(
				digestSuccess("foo", "bar"),
				uidGid("bar", expect.Behaviour{Stdout: []byte("X\n2000\n")}),
			),
			image:   &image{name: "foo"},
			wantErr: true,
		},
		"string gid": {
			expectations: append(
				digestSuccess("foo", "bar"),
				uidGid("bar", expect.Behaviour{Stdout: []byte("1000\nX\n")}),
			),
			image:   &image{name: "foo"},
			wantErr: true,
		},
		// Now for some more run of the mill failures.
		"docker run failure": {
			expectations: append(
				digestSuccess("foo", "bar"),
				uidGid("bar", expect.Behaviour{ExitCode: 1}),
			),
			image:   &image{name: "foo"},
			wantErr: true,
		},
		"digest failure": {
			expectations: digestFailure("foo"),
			image:        &image{name: "foo"},
			wantErr:      true,
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

func digestFailure(name string) []*expect.Expectation {
	return append(
		ensureSuccess(name),
		expect.NewGlob(
			expect.Behaviour{ExitCode: 1},
			// Note the awkward escaping because these arguments are matched by
			// glob.
			"docker", "image", "inspect", "--format", `\{\{.Id}}`, "--", name,
		),
	)
}

func digestSuccess(name, digest string) []*expect.Expectation {
	return append(
		ensureSuccess(name),
		expect.NewGlob(
			expect.Behaviour{Stdout: []byte(digest + "\n")},
			// Note the awkward escaping because these arguments are
			// matched by glob.
			"docker", "image", "inspect", "--format", `\{\{.Id}}`, "--", name,
		),
	)
}

func ensureFailure(name string) []*expect.Expectation {
	return []*expect.Expectation{
		expect.NewGlob(
			// docker image inspect returns 1 for non-existent images.
			expect.Behaviour{ExitCode: 1},
			"docker", "image", "inspect", "--format", "1", name,
		),
		expect.NewGlob(
			expect.Behaviour{ExitCode: 1},
			"docker", "image", "pull", name,
		),
	}
}

// ensureSuccess only provides the short circuit success path for Ensure() (that
// is, where no pull is required).
func ensureSuccess(name string) []*expect.Expectation {
	return []*expect.Expectation{
		expect.NewGlob(
			expect.Success,
			"docker", "image", "inspect", "--format", "1", name,
		),
	}
}

func uidGid(digest string, behaviour expect.Behaviour) *expect.Expectation {
	return expect.NewGlob(
		behaviour,
		"docker", "run", "--rm", "--entrypoint", "/bin/sh",
		digest, "-c", "id -u; id -g",
	)
}

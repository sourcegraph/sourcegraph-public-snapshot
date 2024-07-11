package gitcli

import (
	"context"
	"net/mail"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/cmd/gitserver/internal/git"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestGitCLIBackend_ContributorCounts(t *testing.T) {
	ctx := context.Background()
	// Prepare repo state:
	backend := BackendWithRepoCommands(t,
		"echo line > f",
		"git add f",
		`GIT_COMMITTER_DATE="2015-01-01 00:00 Z" git commit -m foo --author='Foo Author <foo@sourcegraph.com>'`,
		"echo line >> f",
		"git add f",
		`GIT_COMMITTER_DATE="2016-01-01 00:00 Z" git commit -m foo --author='Foobar Author <foobar@sourcegraph.com>'`,
		"mkdir subdir",
		"echo line > subdir/f",
		"git add subdir/f",
		`GIT_COMMITTER_DATE="2017-01-01 00:00 Z" git commit -m foo --author='Bar Author <bar@sourcegraph.com>'`,
		"echo line >> f",
		"git add f",
		`GIT_COMMITTER_DATE="2018-01-01 00:00 Z" git commit -m foo --author='Foo Author <foo@sourcegraph.com>'`,
		"echo line >> f",
		"git add f",
		`GIT_COMMITTER_DATE="2018-01-01 00:00 Z" git commit -m foo --author='Foo Author <foo@sourcegraph.com>'`,
		"echo line >> subdir/f",
		"git add subdir/f",
		`GIT_COMMITTER_DATE="2018-01-01 00:00 Z" git commit -m foo --author='Foobar Author <foobar@sourcegraph.com>'`,
	)

	t.Run("basic", func(t *testing.T) {
		counts, err := backend.ContributorCounts(ctx, git.ContributorCountsOpts{})
		require.NoError(t, err)
		require.Equal(t, []*gitdomain.ContributorCount{
			{
				Name:  "Foo Author",
				Email: "foo@sourcegraph.com",
				Count: 3,
			},
			{
				Name:  "Foobar Author",
				Email: "foobar@sourcegraph.com",
				Count: 2,
			},
			{
				Name:  "Bar Author",
				Email: "bar@sourcegraph.com",
				Count: 1,
			},
		}, counts)
	})

	t.Run("after", func(t *testing.T) {
		counts, err := backend.ContributorCounts(ctx, git.ContributorCountsOpts{
			After: time.Date(2017, 6, 1, 0, 0, 0, 0, time.UTC),
		})
		require.NoError(t, err)
		require.Equal(t, []*gitdomain.ContributorCount{
			{
				Name:  "Foo Author",
				Email: "foo@sourcegraph.com",
				Count: 2,
			},
			{
				Name:  "Foobar Author",
				Email: "foobar@sourcegraph.com",
				Count: 1,
			},
		}, counts)
	})
	t.Run("range", func(t *testing.T) {
		counts, err := backend.ContributorCounts(ctx, git.ContributorCountsOpts{
			Range: "HEAD^1..HEAD",
		})
		require.NoError(t, err)
		require.Equal(t, []*gitdomain.ContributorCount{
			{
				Name:  "Foobar Author",
				Email: "foobar@sourcegraph.com",
				Count: 1,
			},
		}, counts)
		counts, err = backend.ContributorCounts(ctx, git.ContributorCountsOpts{
			Range: "HEAD^1", // Should be all but the last commit.
		})
		require.NoError(t, err)
		require.Equal(t, []*gitdomain.ContributorCount{
			{
				Name:  "Foo Author",
				Email: "foo@sourcegraph.com",
				Count: 3,
			},
			{
				Name:  "Bar Author",
				Email: "bar@sourcegraph.com",
				Count: 1,
			},
			{
				Name:  "Foobar Author",
				Email: "foobar@sourcegraph.com",
				Count: 1,
			},
		}, counts)
	})
	t.Run("path", func(t *testing.T) {
		counts, err := backend.ContributorCounts(ctx, git.ContributorCountsOpts{
			Path: "subdir/",
		})
		require.NoError(t, err)
		require.Equal(t, []*gitdomain.ContributorCount{
			{
				Name:  "Bar Author",
				Email: "bar@sourcegraph.com",
				Count: 1,
			},
			{
				Name:  "Foobar Author",
				Email: "foobar@sourcegraph.com",
				Count: 1,
			},
		}, counts)
		counts, err = backend.ContributorCounts(ctx, git.ContributorCountsOpts{
			Path: "unknown/", // Unknown path is empty.
		})
		require.NoError(t, err)
		require.Equal(t, []*gitdomain.ContributorCount{}, counts)
	})
	t.Run("not found range", func(t *testing.T) {
		_, err := backend.ContributorCounts(ctx, git.ContributorCountsOpts{
			Range: "deadbeefdeadbeefdeadbeefdeadbeefdeadbeef", // Invalid OID
		})
		require.Error(t, err)
		require.True(t, errors.HasType[*gitdomain.RevisionNotFoundError](err))

		_, err = backend.ContributorCounts(ctx, git.ContributorCountsOpts{
			Range: "unknownbranch", // Invalid ref
		})
		require.Error(t, err)
		require.True(t, errors.HasType[*gitdomain.RevisionNotFoundError](err))

		_, err = backend.ContributorCounts(ctx, git.ContributorCountsOpts{
			Range: "unknownbranch..HEAD", // Invalid left hand of range
		})
		require.Error(t, err)
		require.True(t, errors.HasType[*gitdomain.RevisionNotFoundError](err))

		_, err = backend.ContributorCounts(ctx, git.ContributorCountsOpts{
			Range: "HEAD..unknownbranch", // Invalid right hand of range
		})
		require.Error(t, err)
		require.True(t, errors.HasType[*gitdomain.RevisionNotFoundError](err))
	})
}

func TestParseShortLog(t *testing.T) {
	tests := []struct {
		name    string
		input   string // in the format of `git shortlog -sne`
		want    []*gitdomain.ContributorCount
		wantErr error
	}{
		{
			name: "basic",
			input: `
  1125	Jane Doe <jane@sourcegraph.com>
   390	Bot Of Doom <bot@doombot.com>
`,
			want: []*gitdomain.ContributorCount{
				{
					Name:  "Jane Doe",
					Email: "jane@sourcegraph.com",
					Count: 1125,
				},
				{
					Name:  "Bot Of Doom",
					Email: "bot@doombot.com",
					Count: 390,
				},
			},
		},
		{
			name: "commonly malformed (email address as name)",
			input: `  1125	jane@sourcegraph.com <jane@sourcegraph.com>
   390	Bot Of Doom <bot@doombot.com>
`,
			want: []*gitdomain.ContributorCount{
				{
					Name:  "jane@sourcegraph.com",
					Email: "jane@sourcegraph.com",
					Count: 1125,
				},
				{
					Name:  "Bot Of Doom",
					Email: "bot@doombot.com",
					Count: 390,
				},
			},
		},
	}
	for _, tst := range tests {
		t.Run(tst.name, func(t *testing.T) {
			got, gotErr := parseShortLog(strings.NewReader(tst.input))
			if (gotErr == nil) != (tst.wantErr == nil) {
				t.Fatalf("gotErr %+v wantErr %+v", gotErr, tst.wantErr)
			}
			if !reflect.DeepEqual(got, tst.want) {
				t.Logf("got %q", got)
				t.Fatalf("want %q", tst.want)
			}
		})
	}
}

func TestLenientParseAddress(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    *mail.Address
		wantErr bool
	}{
		{
			name:  "valid address",
			input: "John Doe <john@example.com>",
			want: &mail.Address{
				Name:    "John Doe",
				Address: "john@example.com",
			},
			wantErr: false,
		},
		{
			name:  "malformed address with email as name",
			input: "john@example.com <john@example.com>",
			want: &mail.Address{
				Name:    "john@example.com",
				Address: "john@example.com",
			},
			wantErr: false,
		},
		{
			name:    "invalid address",
			input:   "invalid",
			want:    nil,
			wantErr: true,
		},
		{
			name:  "address with spaces",
			input: "  John Doe   <john@example.com>  ",
			want: &mail.Address{
				Name:    "John Doe",
				Address: "john@example.com",
			},
			wantErr: false,
		},
		{
			name:    "empty address",
			input:   "",
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := lenientParseAddress(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("lenientParseAddress() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("lenientParseAddress() got = %v, want %v", got, tt.want)
			}
		})
	}
}

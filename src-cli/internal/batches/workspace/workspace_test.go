package workspace

import (
	"context"
	"runtime"
	"testing"

	"github.com/sourcegraph/sourcegraph/lib/errors"

	"github.com/sourcegraph/src-cli/internal/batches/docker"
	"github.com/sourcegraph/src-cli/internal/batches/mock"
)

func TestBestWorkspaceCreator(t *testing.T) {
	ctx := context.Background()
	isOverridden := !(runtime.GOOS == "darwin" && (runtime.GOARCH == "amd64" || runtime.GOARCH == "arm64"))

	uidGid := func(uid, gid int) docker.UIDGID {
		return docker.UIDGID{UID: uid, GID: gid}
	}

	for name, tc := range map[string]struct {
		images map[string]docker.Image
		want   CreatorType
	}{
		"nil steps": {
			images: nil,
			want:   CreatorTypeVolume,
		},
		"no steps": {
			images: map[string]docker.Image{},
			want:   CreatorTypeVolume,
		},
		"root": {
			images: map[string]docker.Image{
				"image1": &mock.Image{UidGid: uidGid(0, 0)},
			},
			want: CreatorTypeVolume,
		},
		"same user": {
			images: map[string]docker.Image{
				"image1": &mock.Image{UidGid: uidGid(1000, 1000)},
				"image2": &mock.Image{UidGid: uidGid(1000, 1000)},
			},
			want: CreatorTypeVolume,
		},
		"different user": {
			images: map[string]docker.Image{
				"image1": &mock.Image{UidGid: uidGid(1000, 1000)},
				"image2": &mock.Image{UidGid: uidGid(0, 0)},
			},
			want: CreatorTypeBind,
		},
		"id error": {
			images: map[string]docker.Image{
				"image1": &mock.Image{UidGidErr: errors.New("foo")},
			},
			want: CreatorTypeBind,
		},
	} {
		t.Run(name, func(t *testing.T) {
			if isOverridden {
				// This is an overridden platform, so the workspace type will
				// always be bind from bestWorkspaceCreator().
				if have, want := BestCreatorType(ctx, tc.images), CreatorTypeBind; have != want {
					t.Errorf("unexpected creator type on overridden platform: have=%d want=%d", have, want)
				}
			} else {
				if have := BestCreatorType(ctx, tc.images); have != tc.want {
					t.Errorf("unexpected creator type on non-overridden platform: have=%d want=%d", have, tc.want)
				}
			}

			// Regardless of what bestWorkspaceCreator() would have done, let's
			// test that the right thing happens regardless if detection were to
			// actually occur.
			have := detectBestCreatorType(ctx, tc.images)
			if have != tc.want {
				t.Errorf("unexpected creator type: have=%d want=%d", have, tc.want)
			}
		})
	}
}

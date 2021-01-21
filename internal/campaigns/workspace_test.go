package campaigns

import (
	"context"
	"runtime"
	"testing"

	"github.com/sourcegraph/src-cli/internal/exec/expect"
)

func TestBestWorkspaceCreator(t *testing.T) {
	ctx := context.Background()
	isOverridden := !(runtime.GOOS == "darwin" && runtime.GOARCH == "amd64")

	type imageBehaviour struct {
		image     string
		behaviour expect.Behaviour
	}
	for name, tc := range map[string]struct {
		behaviours []imageBehaviour
		want       workspaceCreatorType
	}{
		"nil steps": {
			behaviours: nil,
			want:       workspaceCreatorVolume,
		},
		"no steps": {
			behaviours: []imageBehaviour{},
			want:       workspaceCreatorVolume,
		},
		"root": {
			behaviours: []imageBehaviour{
				{image: "foo", behaviour: expect.Behaviour{Stdout: []byte("0\n")}},
				{image: "bar", behaviour: expect.Behaviour{Stdout: []byte("0\n")}},
			},
			want: workspaceCreatorVolume,
		},
		"same user": {
			behaviours: []imageBehaviour{
				{image: "foo", behaviour: expect.Behaviour{Stdout: []byte("1000\n")}},
			},
			want: workspaceCreatorBind,
		},
		"different user": {
			behaviours: []imageBehaviour{
				{image: "foo", behaviour: expect.Behaviour{Stdout: []byte("0\n")}},
				{image: "bar", behaviour: expect.Behaviour{Stdout: []byte("1000\n")}},
			},
			want: workspaceCreatorBind,
		},
		"invalid id output: string": {
			behaviours: []imageBehaviour{
				{image: "foo", behaviour: expect.Behaviour{Stdout: []byte("xxx\n")}},
			},
			want: workspaceCreatorBind,
		},
		"invalid id output: empty": {
			behaviours: []imageBehaviour{
				{image: "foo", behaviour: expect.Behaviour{Stdout: []byte("")}},
			},
			want: workspaceCreatorBind,
		},
		"error invoking id": {
			behaviours: []imageBehaviour{
				{image: "foo", behaviour: expect.Behaviour{ExitCode: 1}},
			},
			want: workspaceCreatorBind,
		},
	} {
		t.Run(name, func(t *testing.T) {
			var (
				commands []*expect.Expectation = nil
				steps    []Step                = nil
			)
			if tc.behaviours != nil {
				commands = []*expect.Expectation{}
				steps = []Step{}
				for _, imageBehaviour := range tc.behaviours {
					commands = append(commands, expect.NewGlob(
						imageBehaviour.behaviour,
						"docker", "run", "--rm", "--entrypoint", "/bin/sh",
						imageBehaviour.image, "-c", "id -u",
					))
					steps = append(steps, Step{image: imageBehaviour.image})
				}
			}

			if !isOverridden {
				// If bestWorkspaceCreator() won't short circuit on this
				// platform, we're going to run the Docker commands twice by
				// definition.
				expect.Commands(t, append(commands, commands...)...)
			} else {
				expect.Commands(t, commands...)
			}

			if isOverridden {
				// This is an overridden platform, so the workspace type will
				// always be bind from bestWorkspaceCreator().
				if have, want := bestWorkspaceCreator(ctx, steps), workspaceCreatorBind; have != want {
					t.Errorf("unexpected creator type on overridden platform: have=%d want=%d", have, want)
				}
			} else {
				if have := bestWorkspaceCreator(ctx, steps); have != tc.want {
					t.Errorf("unexpected creator type on non-overridden platform: have=%d want=%d", have, tc.want)
				}
			}

			// Regardless of what bestWorkspaceCreator() would have done, let's
			// test that the right thing happens regardless if detection were to
			// actually occur.
			have := detectBestWorkspaceCreator(ctx, steps)
			if have != tc.want {
				t.Errorf("unexpected creator type: have=%d want=%d", have, tc.want)
			}
		})
	}
}

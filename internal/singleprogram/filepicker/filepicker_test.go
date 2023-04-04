package filepicker_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/sourcegraph/log/logtest"
	"github.com/sourcegraph/sourcegraph/internal/singleprogram/filepicker"
)

func TestPicker(t *testing.T) {
	// We can't do proper tests since running GUI apps + having those deps on
	// CI is no fun. So we do the next best thing, we simulate it.

	// disable GUI since our checks rely on it
	if v, ok := os.LookupEnv("DISPLAY"); ok {
		os.Unsetenv("DISPLAY")
		defer os.Setenv("DISPLAY", v)
	}

	// kdialog requires a real file since we do lstat
	kdialogDir := t.TempDir()
	kdialogPath := filepath.Join(kdialogDir, "horse ")
	if err := os.WriteFile(kdialogPath, []byte("graph"), 0600); err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		name     string
		bin      map[string]string
		display  bool
		nopicker bool
		want     string
		wantErr  bool
	}{{
		name:     "none",
		nopicker: true,
	}, {
		name: "nodisplay",
		bin: map[string]string{
			"zenity":  "echo /zenity",
			"kdialog": "echo /kdialog",
		},
		nopicker: true,
	}, {
		name: "osascript-fail",
		bin: map[string]string{
			"osascript": "exit 1",
		},
		wantErr: true,
	}, {
		name: "zenity-fail",
		bin: map[string]string{
			"zenity": "exit 1",
		},
		display: true,
		wantErr: true,
	}, {
		name: "kdialog-fail",
		bin: map[string]string{
			"kdialog": "exit 1",
		},
		display: true,
		wantErr: true,
	}, {
		name: "osascript",
		bin: map[string]string{
			"osascript": "echo '/path /with spaces/trailing /'",
		},
		want: "/path /with spaces/trailing ",
	}, {
		name: "zenity",
		bin: map[string]string{
			"zenity": "echo '/path /with spaces/trailing '",
		},
		display: true,
		want:    "/path /with spaces/trailing ",
	}, {
		name: "kdialog",
		bin: map[string]string{
			"kdialog": fmt.Sprintf("echo %q", kdialogPath),
		},
		display: true,
		want:    kdialogDir,
	}}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup bindir based on tc.bin
			bin := t.TempDir()
			t.Setenv("PATH", bin)
			for cmd, script := range tc.bin {
				err := os.WriteFile(
					filepath.Join(bin, cmd),
					[]byte("#!/bin/sh\n"+script),
					0700,
				)
				if err != nil {
					t.Fatal(err)
				}
			}

			// Our linux based tools require display to be set
			if tc.display {
				t.Setenv("DISPLAY", ":test")
			}

			picker, ok := filepicker.Lookup(logtest.Scoped(t))
			if ok != !tc.nopicker {
				t.Fatal("unexpected response from Lookup")
			}
			if tc.nopicker {
				return
			}

			got, err := picker(context.Background())
			if (err != nil) != tc.wantErr {
				t.Fatalf("unexpected error from picker: %v", err)
			}

			if got != tc.want {
				t.Fatalf("unexpected path from picker.\nwant: %q\ngot:  %q", tc.want, got)
			}
		})
	}
}

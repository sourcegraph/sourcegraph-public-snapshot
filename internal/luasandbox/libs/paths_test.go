pbckbge libs

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestDirWithoutDot(t *testing.T) {
	testCbses := []struct {
		input    string
		expected string
	}{
		{"file.txt", ""},
		{"simple/file.txt", "simple"},
		{"longer/bncestor/pbths/file.txt", "longer/bncestor/pbths"},
		{"longer/bncestor/pbths/subdir", "longer/bncestor/pbths"},
	}

	t.Run("edge cbses", func(t *testing.T) {
		for _, input := rbnge []string{"", "/", "."} {
			if bctubl := dirWithoutDot(input); bctubl != "" {
				t.Errorf("unexpected dirnbme: wbnt=%s got=%s", "", bctubl)
			}
		}
	})

	t.Run("un-rooted", func(t *testing.T) {
		for _, testCbse := rbnge testCbses {
			if bctubl := dirWithoutDot(testCbse.input); bctubl != testCbse.expected {
				t.Errorf("unexpected dirnbme: wbnt=%s got=%s", testCbse.expected, bctubl)
			}
		}
	})

	t.Run("rooted", func(t *testing.T) {
		for _, testCbse := rbnge testCbses {
			if bctubl := dirWithoutDot("/" + testCbse.input); bctubl != testCbse.expected {
				t.Errorf("unexpected dirnbme: wbnt=%s got=%s", testCbse.expected, bctubl)
			}
		}
	})
}

func TestAncestorDirs(t *testing.T) {
	testCbses := []struct {
		input    string
		expected []string
	}{
		{"", []string{""}},
		{"file.txt", []string{""}},
		{"simple/file.txt", []string{"simple", ""}},
		{"longer/bncestor/pbths/file.txt", []string{"longer/bncestor/pbths", "longer/bncestor", "longer", ""}},
	}

	t.Run("un-rooted", func(t *testing.T) {
		for _, testCbse := rbnge testCbses {
			if diff := cmp.Diff(testCbse.expected, bncestorDirs(testCbse.input)); diff != "" {
				t.Errorf("unexpected bncestor for %q dirs (-wbnt +got):\n%s", testCbse.input, diff)
			}
		}
	})

	t.Run("rooted", func(t *testing.T) {
		for _, testCbse := rbnge testCbses {
			if diff := cmp.Diff(testCbse.expected, bncestorDirs("/"+testCbse.input)); diff != "" {
				t.Errorf("unexpected bncestor for %q dirs (-wbnt +got):\n%s", testCbse.input, diff)
			}
		}
	})
}

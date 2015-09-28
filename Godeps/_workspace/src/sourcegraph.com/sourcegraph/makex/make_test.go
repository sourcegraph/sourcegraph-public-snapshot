package makex

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"

	"sourcegraph.com/sourcegraph/rwvfs"
)

func TestMaker_DryRun(t *testing.T) {
	var conf Config
	mf := &Makefile{
		Rules: []Rule{&BasicRule{TargetFile: "x"}},
	}
	mk := conf.NewMaker(mf, "x")
	err := mk.DryRun(ioutil.Discard)
	if err != nil {
		t.Fatal(err)
	}
}

func TestMaker_Run(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "makex")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	conf := &Config{
		ParallelJobs: 1,
		FS:           NewFileSystem(rwvfs.OS(tmpDir)),
	}

	target := "x"
	mf := &Makefile{
		Rules: []Rule{
			&BasicRule{
				TargetFile: target,
				RecipeCmds: []string{"touch " + filepath.Join(tmpDir, target)},
			},
		},
	}

	if isFile(conf.FS, target) {
		t.Fatalf("target %s exists before running Makefile; want it to not exist yet", target)
	}

	mk := conf.NewMaker(mf, target)
	err = mk.Run()
	if err != nil {
		t.Fatalf("Run failed: %s", err)
	}

	if !isFile(conf.FS, target) {
		t.Fatalf("target %s does not exist after running Makefile; want it to exist", target)
	}
}

func isFile(fs rwvfs.FileSystem, file string) bool {
	fi, err := fs.Stat(file)
	if err != nil {
		return false
	}
	return fi.Mode().IsRegular()
}

func TestTargetsNeedingBuild(t *testing.T) {
	tests := map[string]struct {
		mf                         *Makefile
		fs                         FileSystem
		goals                      []string
		wantErr                    error
		wantTargetSetsNeedingBuild [][]string
	}{
		"do nothing if empty": {
			mf: &Makefile{},
			wantTargetSetsNeedingBuild: [][]string{},
		},
		"return error if target isn't defined in Makefile": {
			mf:      &Makefile{},
			goals:   []string{"x"},
			wantErr: errNoRuleToMakeTarget("x"),
		},
		"don't build target that already exists": {
			mf:    &Makefile{Rules: []Rule{&BasicRule{TargetFile: "x"}}},
			fs:    NewFileSystem(rwvfs.Map(map[string]string{"x": ""})),
			goals: []string{"x"},
			wantTargetSetsNeedingBuild: [][]string{},
		},
		"build target that doesn't exist": {
			mf:    &Makefile{Rules: []Rule{&BasicRule{TargetFile: "x"}}},
			fs:    NewFileSystem(rwvfs.Map(map[string]string{})),
			goals: []string{"x"},
			wantTargetSetsNeedingBuild: [][]string{{"x"}},
		},
		"build targets recursively that don't exist": {
			mf: &Makefile{Rules: []Rule{
				&BasicRule{TargetFile: "x0", PrereqFiles: []string{"x1"}},
				&BasicRule{TargetFile: "x1"},
			}},
			fs:    NewFileSystem(rwvfs.Map(map[string]string{})),
			goals: []string{"x0"},
			wantTargetSetsNeedingBuild: [][]string{{"x1"}, {"x0"}},
		},

		"don't build targets that don't directly achieve goals (simple)": {
			mf: &Makefile{Rules: []Rule{
				&BasicRule{TargetFile: "x0"},
				&BasicRule{TargetFile: "x1"},
			}},
			fs:    NewFileSystem(rwvfs.Map(map[string]string{})),
			goals: []string{"x0"},
			wantTargetSetsNeedingBuild: [][]string{{"x0"}},
		},
		"don't build targets that don't achieve goals (complex)": {
			mf: &Makefile{Rules: []Rule{
				&BasicRule{TargetFile: "x0", PrereqFiles: []string{"y"}},
				&BasicRule{TargetFile: "x1"},
				&BasicRule{TargetFile: "y"},
			}},
			fs:    NewFileSystem(rwvfs.Map(map[string]string{})),
			goals: []string{"x0"},
			wantTargetSetsNeedingBuild: [][]string{{"y"}, {"x0"}},
		},
		"don't build targets that don't achieve goals (even when a common prereq is satisfied)": {
			mf: &Makefile{Rules: []Rule{
				&BasicRule{TargetFile: "x0", PrereqFiles: []string{"y"}},
				&BasicRule{TargetFile: "x1", PrereqFiles: []string{"y"}},
				&BasicRule{TargetFile: "y"},
			}},
			fs:    NewFileSystem(rwvfs.Map(map[string]string{})),
			goals: []string{"x0"},
			wantTargetSetsNeedingBuild: [][]string{{"y"}, {"x0"}},
		},

		"don't build goal targets more than once": {
			mf: &Makefile{Rules: []Rule{
				&BasicRule{TargetFile: "x0"},
			}},
			fs:    NewFileSystem(rwvfs.Map(map[string]string{})),
			goals: []string{"x0", "x0"},
			wantTargetSetsNeedingBuild: [][]string{{"x0"}},
		},
		"don't build any targets more than once": {
			mf: &Makefile{Rules: []Rule{
				&BasicRule{TargetFile: "x0", PrereqFiles: []string{"y"}},
				&BasicRule{TargetFile: "x1", PrereqFiles: []string{"y"}},
				&BasicRule{TargetFile: "y"},
			}},
			fs:    NewFileSystem(rwvfs.Map(map[string]string{})),
			goals: []string{"x0", "x1"},
			wantTargetSetsNeedingBuild: [][]string{{"y"}, {"x0", "x1"}},
		},
		"detect 1-cycles": {
			mf: &Makefile{Rules: []Rule{
				&BasicRule{TargetFile: "x0", PrereqFiles: []string{"x0"}},
			}},
			fs:      NewFileSystem(rwvfs.Map(map[string]string{})),
			goals:   []string{"x0"},
			wantErr: errCircularDependency("x0", []string{"x0"}),
		},
		"detect 2-cycles": {
			mf: &Makefile{Rules: []Rule{
				&BasicRule{TargetFile: "x0", PrereqFiles: []string{"x1"}},
				&BasicRule{TargetFile: "x1", PrereqFiles: []string{"x0"}},
			}},
			fs:      NewFileSystem(rwvfs.Map(map[string]string{})),
			goals:   []string{"x0"},
			wantErr: errCircularDependency("x0", []string{"x1"}),
		},
	}

	for label, test := range tests {
		conf := &Config{FS: test.fs}
		mk := conf.NewMaker(test.mf, test.goals...)
		targetSets, err := mk.TargetSetsNeedingBuild()
		if !reflect.DeepEqual(err, test.wantErr) {
			if test.wantErr == nil {
				t.Errorf("%s: TargetsNeedingBuild(%q): error: %s", label, test.goals, err)
				continue
			} else {
				t.Errorf("%s: TargetsNeedingBuild(%q): error: got %q, want %q", label, test.goals, err, test.wantErr)
				continue
			}
		}

		// sort so that test is deterministic
		for _, ts := range targetSets {
			sort.Strings(ts)
		}
		if !reflect.DeepEqual(targetSets, test.wantTargetSetsNeedingBuild) {
			t.Errorf("%s: got targetSets needing build %v, want %v", label, targetSets, test.wantTargetSetsNeedingBuild)
		}
	}
}

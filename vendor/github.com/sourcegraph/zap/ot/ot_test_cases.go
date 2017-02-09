package ot

import "encoding/json"

// We need to keep these test cases for OT ops in a separate file to they can
// be imported by the script that generates the TypeScript version of these
// tests, which keeps them in-sync with the TypeScript OT implementation. See
// also gen/gen-ts.go

var composeTests = map[string]struct {
	a, b, want  WorkspaceOp
	wantErr     bool
	commutative bool // if true, also test compose(b,a) with same want/wantErr
}{
	"copy from created file": {
		a:    WorkspaceOp{Create: []string{"/f1"}},
		b:    WorkspaceOp{Copy: map[string]string{"/f2": "/f1"}},
		want: WorkspaceOp{Create: []string{"/f1", "/f2"}},
		// f1 doesn't exist prior to executing the composed "want" op, so we can't copy f1 to f2.
	},
	"copy with noop": {
		a:           WorkspaceOp{Copy: map[string]string{"/f2": "/f1"}},
		want:        WorkspaceOp{Copy: map[string]string{"/f2": "/f1"}},
		commutative: true,
	},
	"copy to existing file": {
		a:       WorkspaceOp{Create: []string{"/f1"}},
		b:       WorkspaceOp{Copy: map[string]string{"/f1": "/f2"}},
		wantErr: true,
	},
	"copy to existing edited file": {
		a:       WorkspaceOp{Create: []string{"/f1"}, Edit: map[string]EditOps{"/f1": EditOps{{S: "x"}}}},
		b:       WorkspaceOp{Copy: map[string]string{"/f1": "/f2"}},
		wantErr: true,
	},
	"copy from deleted file": {
		a:       WorkspaceOp{Delete: []string{"/f1"}},
		b:       WorkspaceOp{Copy: map[string]string{"/f2": "/f1"}},
		wantErr: true,
	},
	"copy-rename edited file": {
		a: WorkspaceOp{Edit: map[string]EditOps{"/f1": EditOps{{S: "x"}}}},
		b: WorkspaceOp{Copy: map[string]string{"/f2": "/f1"}, Rename: map[string]string{"/f1": "/f3"}},
		want: WorkspaceOp{
			Copy:   map[string]string{"/f2": "/f1"},
			Rename: map[string]string{"/f1": "/f3"},
			Edit:   map[string]EditOps{"/f2": EditOps{{S: "x"}}, "/f3": EditOps{{S: "x"}}},
		},
	},
	"copy to deleted file": {
		a:    WorkspaceOp{Delete: []string{"/f1"}},
		b:    WorkspaceOp{Copy: map[string]string{"/f1": "/f2"}},
		want: WorkspaceOp{Copy: map[string]string{"/f1": "/f2"}},
	},
	"copy from truncated file": {
		a:    WorkspaceOp{Truncate: []string{"/f1"}},
		b:    WorkspaceOp{Copy: map[string]string{"/f2": "/f1"}},
		want: WorkspaceOp{Truncate: []string{"/f1", "/f2"}, Copy: map[string]string{"/f2": "/f1"}},
	},
	"copy to truncated file": {
		a:       WorkspaceOp{Truncate: []string{"/f1"}},
		b:       WorkspaceOp{Copy: map[string]string{"/f1": "/f2"}},
		wantErr: true,
	},
	"copy from copy dest": {
		a:    WorkspaceOp{Copy: map[string]string{"/f2": "/f1"}},
		b:    WorkspaceOp{Copy: map[string]string{"/f3": "/f2"}},
		want: WorkspaceOp{Copy: map[string]string{"/f2": "/f1", "/f3": "/f1"}},
	},
	"copy from copy source and dest": {
		a:    WorkspaceOp{Copy: map[string]string{"/f2": "/f1"}},
		b:    WorkspaceOp{Copy: map[string]string{"/f3": "/f1", "/f4": "/f2"}},
		want: WorkspaceOp{Copy: map[string]string{"/f2": "/f1", "/f3": "/f1", "/f4": "/f1"}},
	},
	"copy self": {
		a:           WorkspaceOp{Copy: map[string]string{"/f1": "/f1"}},
		want:        WorkspaceOp{Copy: nil},
		commutative: true,
	},
	"single-op cyclical copy": {
		a:           WorkspaceOp{Copy: map[string]string{"/f2": "/f1", "/f1": "/f2"}},
		b:           WorkspaceOp{},
		wantErr:     true,
		commutative: true,
	},
	"cyclical copy": {
		a:           WorkspaceOp{Copy: map[string]string{"/f2": "/f1"}},
		b:           WorkspaceOp{Copy: map[string]string{"/f1": "/f2"}},
		wantErr:     true, // because f1 already exists after a but can't exist for b
		commutative: true,
	},
	"copy buffer": {
		a:    WorkspaceOp{Copy: map[string]string{"#f2": "#f1"}},
		want: WorkspaceOp{Copy: map[string]string{"#f2": "#f1"}},
	},
	"copy to buffer": {
		a:    WorkspaceOp{Copy: map[string]string{"#f": "/f"}},
		want: WorkspaceOp{Copy: map[string]string{"#f": "/f"}},
	},
	"copy to buffer and edit": {
		a:    WorkspaceOp{Copy: map[string]string{"#f": "/f"}},
		b:    WorkspaceOp{Edit: map[string]EditOps{"#f": {{S: "x"}}}},
		want: WorkspaceOp{Copy: map[string]string{"#f": "/f"}, Edit: map[string]EditOps{"#f": {{S: "x"}}}},
	},
	"copy saved file": {
		a:    WorkspaceOp{Save: []string{"#f1"}},
		b:    WorkspaceOp{Copy: map[string]string{"/f2": "/f1"}},
		want: WorkspaceOp{Copy: map[string]string{"#f2": "#f1"}, Save: []string{"#f1", "#f2"}},
	},
	"copy to saved file": {
		a:       WorkspaceOp{Save: []string{"#f1"}},
		b:       WorkspaceOp{Copy: map[string]string{"/f1": "/f2"}},
		wantErr: true,
	},
	"copy saved buffer": {
		a:    WorkspaceOp{Save: []string{"#f"}},
		b:    WorkspaceOp{Copy: map[string]string{"#f": "/f"}},
		want: WorkspaceOp{Save: []string{"#f"}, Copy: map[string]string{"#f": "/f"}},
	},
	"copy save-edited file": {
		a: WorkspaceOp{Save: []string{"#f1"}, Edit: map[string]EditOps{"#f1": EditOps{{S: "x"}}}},
		b: WorkspaceOp{Copy: map[string]string{"/f2": "/f1"}},
		want: WorkspaceOp{
			Copy: map[string]string{"#f2": "#f1"},
			Save: []string{"#f1", "#f2"},
			Edit: map[string]EditOps{"#f1": EditOps{{S: "x"}}},
		},
	},
	"chained copy-edit of indirect source file": {
		a:    WorkspaceOp{Copy: map[string]string{"/f2": "/f1"}},
		b:    WorkspaceOp{Copy: map[string]string{"/f3": "/f2", "/f4": "/f1"}, Edit: map[string]EditOps{"/f1": EditOps{{S: "x"}}}},
		want: WorkspaceOp{Copy: map[string]string{"/f2": "/f1", "/f3": "/f1", "/f4": "/f1"}, Edit: map[string]EditOps{"/f1": EditOps{{S: "x"}}}},
	},
	"chained copy-edit of direct source file": {
		a:    WorkspaceOp{Copy: map[string]string{"/f2": "/f1"}},
		b:    WorkspaceOp{Copy: map[string]string{"/f3": "/f2", "/f4": "/f1"}, Edit: map[string]EditOps{"/f2": EditOps{{S: "x"}}}},
		want: WorkspaceOp{Copy: map[string]string{"/f2": "/f1", "/f3": "/f1", "/f4": "/f1"}, Edit: map[string]EditOps{"/f2": EditOps{{S: "x"}}}},
	},
	"chained copy of edited source file": {
		a: WorkspaceOp{Copy: map[string]string{"/f2": "/f1"}, Edit: map[string]EditOps{"/f1": EditOps{{S: "x"}}}},
		b: WorkspaceOp{Copy: map[string]string{"/f3": "/f2", "/f4": "/f1"}},
		want: WorkspaceOp{
			Copy: map[string]string{"/f2": "/f1", "/f3": "/f1", "/f4": "/f1"},
			Edit: map[string]EditOps{"/f1": EditOps{{S: "x"}}, "/f4": EditOps{{S: "x"}}},
		},
	},
	"chained copy-edit from edited source files": {
		a: WorkspaceOp{
			Copy: map[string]string{"/f2": "/f1"},
			Edit: map[string]EditOps{"/f1": EditOps{{N: 1}, {S: "a"}}, "/f2": EditOps{{N: 1}, {S: "x"}}},
		},
		b: WorkspaceOp{
			Copy: map[string]string{"/f3": "/f1", "/f4": "/f2"},
			Edit: map[string]EditOps{
				"/f1": EditOps{{N: 2}, {S: "b"}},
				"/f2": EditOps{{N: 2}, {S: "y"}},
				"/f3": EditOps{{N: 2}, {S: "c"}},
				"/f4": EditOps{{N: 2}, {S: "z"}},
			},
		},
		want: WorkspaceOp{
			Copy: map[string]string{"/f2": "/f1", "/f3": "/f1", "/f4": "/f1"},
			Edit: map[string]EditOps{
				"/f1": EditOps{{N: 1}, {S: "ab"}},
				"/f2": EditOps{{N: 1}, {S: "xy"}},
				"/f3": EditOps{{N: 1}, {S: "ac"}},
				"/f4": EditOps{{N: 1}, {S: "xz"}},
			},
		},
	},
	"copy to source file from another file": {
		a:       WorkspaceOp{Copy: map[string]string{"/f2": "/f1"}},
		b:       WorkspaceOp{Copy: map[string]string{"/f1": "/f3"}},
		wantErr: true,
	},
	"copy from edited file": {
		a:    WorkspaceOp{Edit: map[string]EditOps{"/f1": EditOps{{S: "x"}}}},
		b:    WorkspaceOp{Copy: map[string]string{"/f2": "/f1"}},
		want: WorkspaceOp{Copy: map[string]string{"/f2": "/f1"}, Edit: map[string]EditOps{"/f1": EditOps{{S: "x"}}, "/f2": EditOps{{S: "x"}}}},
	},
	"copy from created-edited file": {
		a:    WorkspaceOp{Create: []string{"/f1"}, Edit: map[string]EditOps{"/f1": EditOps{{S: "x"}}}},
		b:    WorkspaceOp{Copy: map[string]string{"/f2": "/f1"}},
		want: WorkspaceOp{Create: []string{"/f1", "/f2"}, Edit: map[string]EditOps{"/f1": EditOps{{S: "x"}}, "/f2": EditOps{{S: "x"}}}},
		// See note for "copy from created file" test case.
	},
	"copy renamed source file": {
		a:       WorkspaceOp{Rename: map[string]string{"/f1": "/f2"}},
		b:       WorkspaceOp{Copy: map[string]string{"/f3": "/f1"}},
		wantErr: true,
	},
	"copy and rename source file": {
		a:           WorkspaceOp{Copy: map[string]string{"/f3": "/f1"}, Rename: map[string]string{"/f1": "/f2"}},
		want:        WorkspaceOp{Copy: map[string]string{"/f3": "/f1"}, Rename: map[string]string{"/f1": "/f2"}},
		commutative: true,
	},
	"copy and rename source file, delete dests": {
		a:    WorkspaceOp{Copy: map[string]string{"/f3": "/f1"}, Rename: map[string]string{"/f1": "/f2"}},
		b:    WorkspaceOp{Delete: []string{"/f2", "/f3"}},
		want: WorkspaceOp{Delete: []string{"/f1"}},
	},
	"copy renamed dest file": {
		a:    WorkspaceOp{Rename: map[string]string{"/f1": "/f2"}},
		b:    WorkspaceOp{Copy: map[string]string{"/f3": "/f2"}},
		want: WorkspaceOp{Copy: map[string]string{"/f3": "/f1"}, Rename: map[string]string{"/f1": "/f2"}},
	},
	"copy renamed and edited file": {
		a:    WorkspaceOp{Rename: map[string]string{"/f1": "/f2"}, Edit: map[string]EditOps{"/f2": EditOps{{S: "x"}}}},
		b:    WorkspaceOp{Copy: map[string]string{"/f3": "/f2"}},
		want: WorkspaceOp{Copy: map[string]string{"/f3": "/f1"}, Rename: map[string]string{"/f1": "/f2"}, Edit: map[string]EditOps{"/f2": EditOps{{S: "x"}}, "/f3": EditOps{{S: "x"}}}},
	},
	"copy-edit(source) renamed file": {
		a:    WorkspaceOp{Rename: map[string]string{"/f1": "/f2"}},
		b:    WorkspaceOp{Copy: map[string]string{"/f3": "/f2"}, Edit: map[string]EditOps{"/f2": EditOps{{S: "x"}}}},
		want: WorkspaceOp{Copy: map[string]string{"/f3": "/f1"}, Rename: map[string]string{"/f1": "/f2"}, Edit: map[string]EditOps{"/f2": EditOps{{S: "x"}}}},
	},
	"copy-edit(dest) renamed file": {
		a:    WorkspaceOp{Rename: map[string]string{"/f1": "/f2"}},
		b:    WorkspaceOp{Copy: map[string]string{"/f3": "/f2"}, Edit: map[string]EditOps{"/f3": EditOps{{S: "x"}}}},
		want: WorkspaceOp{Copy: map[string]string{"/f3": "/f1"}, Rename: map[string]string{"/f1": "/f2"}, Edit: map[string]EditOps{"/f3": EditOps{{S: "x"}}}},
	},
	"copy-edit(source dest) renamed file": {
		a:    WorkspaceOp{Rename: map[string]string{"/f1": "/f2"}},
		b:    WorkspaceOp{Copy: map[string]string{"/f3": "/f2"}, Edit: map[string]EditOps{"/f2": EditOps{{S: "x"}}, "/f3": EditOps{{S: "y"}}}},
		want: WorkspaceOp{Copy: map[string]string{"/f3": "/f1"}, Rename: map[string]string{"/f1": "/f2"}, Edit: map[string]EditOps{"/f2": EditOps{{S: "x"}}, "/f3": EditOps{{S: "y"}}}},
	},
	"copy sel file": {
		a: WorkspaceOp{Sel: map[string]map[string]*Sel{"/f1": {"u": {1, 2}}}},
		b: WorkspaceOp{Copy: map[string]string{"/f2": "/f1"}},
		want: WorkspaceOp{
			Copy: map[string]string{"/f2": "/f1"},
			Sel:  map[string]map[string]*Sel{"/f1": {"u": {1, 2}}},
		},
	},

	"rename with noop": {
		a:           WorkspaceOp{Rename: map[string]string{"/f2": "/f1"}},
		want:        WorkspaceOp{Rename: map[string]string{"/f2": "/f1"}},
		commutative: true,
	},
	"rename to-from self": {
		a:           WorkspaceOp{Rename: map[string]string{"/f1": "/f1"}},
		want:        WorkspaceOp{Rename: nil},
		commutative: true,
	},
	"rename created file": {
		a:    WorkspaceOp{Create: []string{"/f1"}},
		b:    WorkspaceOp{Rename: map[string]string{"/f1": "/f2"}},
		want: WorkspaceOp{Create: []string{"/f2"}},
	},
	"rename edited file": {
		a:    WorkspaceOp{Edit: map[string]EditOps{"/f1": EditOps{{S: "x"}}}},
		b:    WorkspaceOp{Rename: map[string]string{"/f1": "/f2"}},
		want: WorkspaceOp{Rename: map[string]string{"/f1": "/f2"}, Edit: map[string]EditOps{"/f2": EditOps{{S: "x"}}}},
	},
	"rename created and edited file": {
		a:    WorkspaceOp{Create: []string{"/f1"}, Edit: map[string]EditOps{"/f1": EditOps{{S: "x"}}}},
		b:    WorkspaceOp{Rename: map[string]string{"/f1": "/f2"}},
		want: WorkspaceOp{Create: []string{"/f2"}, Edit: map[string]EditOps{"/f2": EditOps{{S: "x"}}}},
	},
	"rename to a deleted file": {
		a:    WorkspaceOp{Delete: []string{"/f1"}},
		b:    WorkspaceOp{Rename: map[string]string{"/f2": "/f1"}},
		want: WorkspaceOp{Rename: map[string]string{"/f2": "/f1"}},
	},
	"rename-edit to a deleted file": {
		a:    WorkspaceOp{Delete: []string{"/f1"}},
		b:    WorkspaceOp{Rename: map[string]string{"/f2": "/f1"}, Edit: map[string]EditOps{"/f1": EditOps{{S: "x"}}}},
		want: WorkspaceOp{Rename: map[string]string{"/f2": "/f1"}, Edit: map[string]EditOps{"/f1": EditOps{{S: "x"}}}},
	},
	"rename deleted file": {
		a:       WorkspaceOp{Delete: []string{"/f1"}},
		b:       WorkspaceOp{Rename: map[string]string{"/f1": "/f2"}},
		wantErr: true,
	},
	"rename to existing file": {
		a:       WorkspaceOp{Create: []string{"/f2"}},
		b:       WorkspaceOp{Rename: map[string]string{"/f1": "/f2"}},
		wantErr: true,
	},
	"rename to existing edited file": {
		a:       WorkspaceOp{Create: []string{"/f2"}, Edit: map[string]EditOps{"/f2": EditOps{{S: "x"}}}},
		b:       WorkspaceOp{Rename: map[string]string{"/f1": "/f2"}},
		wantErr: true,
	},
	"rename truncated file": {
		a:    WorkspaceOp{Truncate: []string{"/f1"}},
		b:    WorkspaceOp{Rename: map[string]string{"/f1": "/f2"}},
		want: WorkspaceOp{Rename: map[string]string{"/f1": "/f2"}, Truncate: []string{"/f2"}},
	},
	"rename renamed file": {
		a:    WorkspaceOp{Rename: map[string]string{"/f1": "/f2"}},
		b:    WorkspaceOp{Rename: map[string]string{"/f2": "/f3"}},
		want: WorkspaceOp{Rename: map[string]string{"/f1": "/f3"}},
	},
	"rename back to source file": {
		a:    WorkspaceOp{Rename: map[string]string{"/f1": "/f2"}},
		b:    WorkspaceOp{Rename: map[string]string{"/f2": "/f1"}},
		want: WorkspaceOp{},
	},
	"rename copied source file": {
		a:    WorkspaceOp{Copy: map[string]string{"/f1": "/f2"}},
		b:    WorkspaceOp{Rename: map[string]string{"/f2": "/f3"}},
		want: WorkspaceOp{Copy: map[string]string{"/f1": "/f2"}, Rename: map[string]string{"/f2": "/f3"}},
	},
	"rename saved file": {
		a: WorkspaceOp{Save: []string{"#f1"}},
		b: WorkspaceOp{Rename: map[string]string{"/f1": "/f2"}},
		want: WorkspaceOp{
			Copy:   map[string]string{"#f2": "#f1"},
			Save:   []string{"#f2"},
			Delete: []string{"/f1"},
		},
	},
	"rename to saved file": {
		a:       WorkspaceOp{Save: []string{"#f1"}},
		b:       WorkspaceOp{Rename: map[string]string{"/f2": "/f1"}},
		wantErr: true,
	},
	"rename-edit copied source file": {
		a:    WorkspaceOp{Copy: map[string]string{"/f1": "/f2"}},
		b:    WorkspaceOp{Rename: map[string]string{"/f1": "/f3"}, Edit: map[string]EditOps{"/f2": EditOps{{S: "x"}}}},
		want: WorkspaceOp{Copy: map[string]string{"/f3": "/f2"}, Edit: map[string]EditOps{"/f2": EditOps{{S: "x"}}}},
	},
	"rename copied dest file": {
		a:    WorkspaceOp{Copy: map[string]string{"/f1": "/f2"}},   // cp f2 f1 # f1 == f2
		b:    WorkspaceOp{Rename: map[string]string{"/f1": "/f3"}}, // mv f1 f3 # f2 == f3
		want: WorkspaceOp{Copy: map[string]string{"/f3": "/f2"}},   // cp f2 f3
	},
	"rename-edit copied dest file": {
		a:    WorkspaceOp{Copy: map[string]string{"/f1": "/f2"}},
		b:    WorkspaceOp{Rename: map[string]string{"/f1": "/f3"}, Edit: map[string]EditOps{"/f3": EditOps{{S: "x"}}}},
		want: WorkspaceOp{Copy: map[string]string{"/f3": "/f2"}, Edit: map[string]EditOps{"/f3": EditOps{{S: "x"}}}},
	},
	"rename sel file": {
		a: WorkspaceOp{Sel: map[string]map[string]*Sel{"/f1": {"u": {1, 2}}}},
		b: WorkspaceOp{Rename: map[string]string{"/f1": "/f2"}},
		want: WorkspaceOp{
			Rename: map[string]string{"/f1": "/f2"},
			Sel:    map[string]map[string]*Sel{"/f2": {"u": {1, 2}}},
		},
	},

	"create file": {
		a:           WorkspaceOp{Create: []string{"/f"}},
		want:        WorkspaceOp{Create: []string{"/f"}},
		commutative: true,
	},
	"create different files": {
		a:           WorkspaceOp{Create: []string{"/a"}},
		b:           WorkspaceOp{Create: []string{"/b"}},
		want:        WorkspaceOp{Create: []string{"/a", "/b"}},
		commutative: true,
	},
	"create created file": {
		a:       WorkspaceOp{Create: []string{"/f"}},
		b:       WorkspaceOp{Create: []string{"/f"}},
		wantErr: true,
	},
	"create same and different files": {
		a:           WorkspaceOp{Create: []string{"/a", "/f"}},
		b:           WorkspaceOp{Create: []string{"/b", "/f"}},
		wantErr:     true,
		commutative: true,
	},
	"create deleted file": {
		a:    WorkspaceOp{Delete: []string{"/f"}},
		b:    WorkspaceOp{Create: []string{"/f"}},
		want: WorkspaceOp{Truncate: []string{"/f"}},
	},
	"create truncated file": {
		a:       WorkspaceOp{Truncate: []string{"/f"}},
		b:       WorkspaceOp{Create: []string{"/f"}},
		wantErr: true,
	},
	"create buffer": {
		a:           WorkspaceOp{Create: []string{"#f1"}},
		want:        WorkspaceOp{Create: []string{"#f1"}},
		commutative: true,
	},
	"create saved file": {
		a:       WorkspaceOp{Save: []string{"#f1"}},
		b:       WorkspaceOp{Create: []string{"/f1"}},
		wantErr: true,
	},
	"create created and edited file": {
		a:       WorkspaceOp{Create: []string{"/f"}, Edit: map[string]EditOps{"/f": EditOps{{S: "x"}}}},
		b:       WorkspaceOp{Create: []string{"/f"}},
		wantErr: true,
	},
	"create edited file": {
		a:       WorkspaceOp{Edit: map[string]EditOps{"/f": EditOps{{S: "x"}}}},
		b:       WorkspaceOp{Create: []string{"/f"}},
		wantErr: true,
	},
	"create renamed dest file": {
		a:       WorkspaceOp{Rename: map[string]string{"/f1": "/f2"}},
		b:       WorkspaceOp{Create: []string{"/f2"}},
		wantErr: true,
	},
	"create renamed-edited file": {
		a:       WorkspaceOp{Rename: map[string]string{"/f1": "/f2"}, Edit: map[string]EditOps{"/f2": EditOps{{S: "x"}}}},
		b:       WorkspaceOp{Create: []string{"/f2"}},
		wantErr: true,
	},
	"create renamed source file": {
		a:    WorkspaceOp{Rename: map[string]string{"/f1": "/f2"}},
		b:    WorkspaceOp{Create: []string{"/f1"}},
		want: WorkspaceOp{Rename: map[string]string{"/f1": "/f2"}, Create: []string{"/f1"}},
	},
	"create-edit renamed source file": {
		a:    WorkspaceOp{Rename: map[string]string{"/f1": "/f2"}},
		b:    WorkspaceOp{Create: []string{"/f1"}, Edit: map[string]EditOps{"/f1": EditOps{{S: "x"}}}},
		want: WorkspaceOp{Rename: map[string]string{"/f1": "/f2"}, Create: []string{"/f1"}, Edit: map[string]EditOps{"/f1": EditOps{{S: "x"}}}},
	},
	"create copied source file": {
		a:       WorkspaceOp{Copy: map[string]string{"/f1": "/f2"}},
		b:       WorkspaceOp{Create: []string{"/f2"}},
		wantErr: true,
	},
	"create copied dest file": {
		a:       WorkspaceOp{Copy: map[string]string{"/f1": "/f2"}},
		b:       WorkspaceOp{Create: []string{"/f1"}},
		wantErr: true,
	},
	"create sel file": {
		a:       WorkspaceOp{Sel: map[string]map[string]*Sel{"/f": {"u": {1, 2}}}},
		b:       WorkspaceOp{Create: []string{"/f"}},
		wantErr: true,
	},

	"delete file": {
		a:           WorkspaceOp{Delete: []string{"/f"}},
		want:        WorkspaceOp{Delete: []string{"/f"}},
		commutative: true,
	},
	"delete different files": {
		a:           WorkspaceOp{Delete: []string{"/a"}},
		b:           WorkspaceOp{Delete: []string{"/b"}},
		want:        WorkspaceOp{Delete: []string{"/a", "/b"}},
		commutative: true,
	},
	"delete deleted file": {
		a:       WorkspaceOp{Delete: []string{"/f"}},
		b:       WorkspaceOp{Delete: []string{"/f"}},
		wantErr: true,
	},
	"delete same and different files": {
		a:           WorkspaceOp{Delete: []string{"/a", "/f"}},
		b:           WorkspaceOp{Delete: []string{"/b", "/f"}},
		wantErr:     true,
		commutative: true,
	},
	"delete created file": {
		a:    WorkspaceOp{Create: []string{"/f"}},
		b:    WorkspaceOp{Delete: []string{"/f"}},
		want: WorkspaceOp{},
	},
	"delete edited file": {
		a:    WorkspaceOp{Edit: map[string]EditOps{"/f": EditOps{{S: "x"}}}},
		b:    WorkspaceOp{Delete: []string{"/f"}},
		want: WorkspaceOp{Delete: []string{"/f"}},
	},
	"delete buffer file": {
		a:    WorkspaceOp{Edit: map[string]EditOps{"#f": EditOps{{S: "x"}}}},
		b:    WorkspaceOp{Delete: []string{"#f"}},
		want: WorkspaceOp{Delete: []string{"#f"}},
	},
	"delete just-deleted buffer file": {
		a:       WorkspaceOp{Delete: []string{"#f2"}},
		b:       WorkspaceOp{Delete: []string{"#f2"}},
		wantErr: true,
	},
	"delete just-created buffer file": {
		a:    WorkspaceOp{Copy: map[string]string{"#f": "/f"}, Edit: map[string]EditOps{"#f": EditOps{{S: "x"}}}},
		b:    WorkspaceOp{Delete: []string{"#f"}},
		want: WorkspaceOp{},
	},
	"delete created-edited file": {
		a:    WorkspaceOp{Create: []string{"/f"}, Edit: map[string]EditOps{"/f": EditOps{{S: "x"}}}},
		b:    WorkspaceOp{Delete: []string{"/f"}},
		want: WorkspaceOp{},
	},
	"delete renamed file": {
		a:    WorkspaceOp{Rename: map[string]string{"/f1": "/f2"}},
		b:    WorkspaceOp{Delete: []string{"/f2"}},
		want: WorkspaceOp{Delete: []string{"/f1"}},
	},
	"delete renamed-edited file": {
		a:    WorkspaceOp{Rename: map[string]string{"/f1": "/f2"}, Edit: map[string]EditOps{"/f2": EditOps{{S: "x"}}}},
		b:    WorkspaceOp{Delete: []string{"/f2"}},
		want: WorkspaceOp{Delete: []string{"/f1"}},
	},
	"delete rename dest file, recreate rename source file": {
		a:    WorkspaceOp{Rename: map[string]string{"/f1": "/f2"}},
		b:    WorkspaceOp{Create: []string{"/f1"}, Delete: []string{"/f2"}},
		want: WorkspaceOp{Truncate: []string{"/f1"}},
	},
	"delete renamed source file": {
		a:       WorkspaceOp{Rename: map[string]string{"/f1": "/f2"}},
		b:       WorkspaceOp{Delete: []string{"/f1"}},
		wantErr: true,
	},
	"delete copied source file": {
		a:    WorkspaceOp{Copy: map[string]string{"/f1": "/f2"}},
		b:    WorkspaceOp{Delete: []string{"/f2"}},
		want: WorkspaceOp{Delete: []string{"/f2"}, Copy: map[string]string{"/f1": "/f2"}},
	},
	"delete copied dest file": {
		a:    WorkspaceOp{Copy: map[string]string{"/f1": "/f2"}},
		b:    WorkspaceOp{Delete: []string{"/f1"}},
		want: WorkspaceOp{},
	},
	"delete copied and edited dest file": {
		a:    WorkspaceOp{Copy: map[string]string{"/f1": "/f2"}, Edit: map[string]EditOps{"/f1": EditOps{{S: "x"}}}},
		b:    WorkspaceOp{Delete: []string{"/f1"}},
		want: WorkspaceOp{},
	},
	"delete copied and renamed file": {
		a:    WorkspaceOp{Copy: map[string]string{"/f1": "/f2"}, Rename: map[string]string{"/f2": "/f3"}},
		b:    WorkspaceOp{Delete: []string{"/f1"}},
		want: WorkspaceOp{Rename: map[string]string{"/f2": "/f3"}},
	},
	"delete copied and renamed files": {
		a:    WorkspaceOp{Copy: map[string]string{"/f1": "/f2"}, Rename: map[string]string{"/f2": "/f3"}},
		b:    WorkspaceOp{Delete: []string{"/f1", "/f3"}},
		want: WorkspaceOp{Delete: []string{"/f2"}},
	},
	"delete sel file": {
		a:    WorkspaceOp{Sel: map[string]map[string]*Sel{"/f": {"u": {1, 2}}}},
		b:    WorkspaceOp{Delete: []string{"/f"}},
		want: WorkspaceOp{Delete: []string{"/f"}},
	},
	"delete file of saved buffer": {
		a:    WorkspaceOp{Save: []string{"#f1"}},
		b:    WorkspaceOp{Delete: []string{"/f1"}},
		want: WorkspaceOp{Delete: []string{"/f1"}},
	},
	"delete save-renamed file": {
		a:    WorkspaceOp{Save: []string{"#f1"}, Rename: map[string]string{"/f1": "/f2"}},
		b:    WorkspaceOp{Delete: []string{"/f2"}},
		want: WorkspaceOp{Delete: []string{"/f1"}},
	},

	"truncate file": {
		a:           WorkspaceOp{Truncate: []string{"/f"}},
		want:        WorkspaceOp{Truncate: []string{"/f"}},
		commutative: true,
	},
	"truncate different files": {
		a:           WorkspaceOp{Truncate: []string{"a"}},
		b:           WorkspaceOp{Truncate: []string{"b"}},
		want:        WorkspaceOp{Truncate: []string{"a", "b"}},
		commutative: true,
	},
	"truncate truncated file": {
		a:    WorkspaceOp{Truncate: []string{"/f"}},
		b:    WorkspaceOp{Truncate: []string{"/f"}},
		want: WorkspaceOp{Truncate: []string{"/f"}},
	},
	"truncate same and different files": {
		a:           WorkspaceOp{Truncate: []string{"/a", "/f"}},
		b:           WorkspaceOp{Truncate: []string{"/b", "/f"}},
		want:        WorkspaceOp{Truncate: []string{"/a", "/b", "/f"}},
		commutative: true,
	},
	"truncate created file": {
		a:    WorkspaceOp{Create: []string{"/f"}},
		b:    WorkspaceOp{Truncate: []string{"/f"}},
		want: WorkspaceOp{Create: []string{"/f"}},
	},
	"truncate created and edited file": {
		a:    WorkspaceOp{Create: []string{"/f"}, Edit: map[string]EditOps{"/f": EditOps{{S: "x"}}}},
		b:    WorkspaceOp{Truncate: []string{"/f"}},
		want: WorkspaceOp{Create: []string{"/f"}},
	},
	"truncate deleted file": {
		a:       WorkspaceOp{Delete: []string{"/f"}},
		b:       WorkspaceOp{Truncate: []string{"/f"}},
		wantErr: true,
	},
	"truncate edited file": {
		a:    WorkspaceOp{Edit: map[string]EditOps{"/f": EditOps{{S: "x"}}}},
		b:    WorkspaceOp{Truncate: []string{"/f"}},
		want: WorkspaceOp{Truncate: []string{"/f"}},
	},
	"truncate renamed file": {
		a:    WorkspaceOp{Rename: map[string]string{"/f1": "/f2"}},
		b:    WorkspaceOp{Truncate: []string{"/f2"}},
		want: WorkspaceOp{Rename: map[string]string{"/f1": "/f2"}, Truncate: []string{"/f2"}},
	},
	"truncate renamed and edited file": {
		a:    WorkspaceOp{Rename: map[string]string{"/f1": "/f2"}, Edit: map[string]EditOps{"/f2": EditOps{{S: "x"}}}},
		b:    WorkspaceOp{Truncate: []string{"/f2"}},
		want: WorkspaceOp{Rename: map[string]string{"/f1": "/f2"}, Truncate: []string{"/f2"}},
	},
	"truncate renamed source file": {
		a:       WorkspaceOp{Rename: map[string]string{"/f1": "/f2"}},
		b:       WorkspaceOp{Truncate: []string{"/f1"}},
		wantErr: true,
	},
	"truncate copied source file": {
		a:    WorkspaceOp{Copy: map[string]string{"/f1": "/f2"}},
		b:    WorkspaceOp{Truncate: []string{"/f2"}},
		want: WorkspaceOp{Copy: map[string]string{"/f1": "/f2"}, Truncate: []string{"/f2"}},
	},
	"truncate copied dest file": {
		a:    WorkspaceOp{Copy: map[string]string{"/f1": "/f2"}},
		b:    WorkspaceOp{Truncate: []string{"/f1"}},
		want: WorkspaceOp{Copy: map[string]string{"/f1": "/f2"}, Truncate: []string{"/f1"}},
	},
	"truncate saved buffer": {
		a:    WorkspaceOp{Save: []string{"#f1"}},
		b:    WorkspaceOp{Truncate: []string{"#f1"}},
		want: WorkspaceOp{Save: []string{"#f1"}, Truncate: []string{"#f1"}},
	},
	"truncate saved file": {
		a:    WorkspaceOp{Save: []string{"#f1"}},
		b:    WorkspaceOp{Truncate: []string{"/f1"}},
		want: WorkspaceOp{Truncate: []string{"/f1"}},
	},
	"truncate sel file": {
		a:    WorkspaceOp{Sel: map[string]map[string]*Sel{"/f": {"u": {1, 2}}}},
		b:    WorkspaceOp{Truncate: []string{"/f"}},
		want: WorkspaceOp{Truncate: []string{"/f"}},
	},
	"truncate-edit sel file": {
		a: WorkspaceOp{Sel: map[string]map[string]*Sel{"/f": {"u": {1, 2}}}},
		b: WorkspaceOp{Truncate: []string{"/f"}, Edit: map[string]EditOps{"/f": EditOps{{S: "x"}}}},
		want: WorkspaceOp{
			Truncate: []string{"/f"},
			Edit:     map[string]EditOps{"/f": EditOps{{S: "x"}}},
		},
	},

	"edit file": {
		a:    WorkspaceOp{},
		b:    WorkspaceOp{Edit: map[string]EditOps{"/f": EditOps{{S: "x"}}}},
		want: WorkspaceOp{Edit: map[string]EditOps{"/f": EditOps{{S: "x"}}}},
	},
	"create-edit file": {
		a:    WorkspaceOp{},
		b:    WorkspaceOp{Create: []string{"/f"}, Edit: map[string]EditOps{"/f": EditOps{{S: "x"}}}},
		want: WorkspaceOp{Create: []string{"/f"}, Edit: map[string]EditOps{"/f": EditOps{{S: "x"}}}},
	},
	"edit deleted file": {
		a:       WorkspaceOp{Delete: []string{"/f"}},
		b:       WorkspaceOp{Edit: map[string]EditOps{"/f": EditOps{{S: "x"}}}},
		wantErr: true,
	},
	"edit created file": {
		a:    WorkspaceOp{Create: []string{"/f"}},
		b:    WorkspaceOp{Edit: map[string]EditOps{"/f": EditOps{{S: "x"}}}},
		want: WorkspaceOp{Create: []string{"/f"}, Edit: map[string]EditOps{"/f": EditOps{{S: "x"}}}},
	},
	"edit edited file": {
		a:    WorkspaceOp{Edit: map[string]EditOps{"/f": EditOps{{S: "x"}}}},
		b:    WorkspaceOp{Edit: map[string]EditOps{"/f": EditOps{{N: 1}, {S: "y"}}}},
		want: WorkspaceOp{Edit: map[string]EditOps{"/f": EditOps{{S: "xy"}}}},
	},
	"invalid edit of created file": {
		a:       WorkspaceOp{Create: []string{"/f"}},
		b:       WorkspaceOp{Edit: map[string]EditOps{"/f": EditOps{{S: "x"}, {N: 1}}}},
		wantErr: true,
	},
	"edit created and edited file": {
		a:    WorkspaceOp{Create: []string{"/f"}, Edit: map[string]EditOps{"/f": EditOps{{S: "x"}}}},
		b:    WorkspaceOp{Edit: map[string]EditOps{"/f": EditOps{{N: 1}, {S: "y"}}}},
		want: WorkspaceOp{Create: []string{"/f"}, Edit: map[string]EditOps{"/f": EditOps{{S: "xy"}}}},
	},
	"edit copied source file": {
		a:    WorkspaceOp{Copy: map[string]string{"/f1": "/f2"}},
		b:    WorkspaceOp{Edit: map[string]EditOps{"/f2": EditOps{{S: "x"}}}},
		want: WorkspaceOp{Copy: map[string]string{"/f1": "/f2"}, Edit: map[string]EditOps{"/f2": EditOps{{S: "x"}}}},
	},
	"edit copied dest file": {
		a:    WorkspaceOp{Copy: map[string]string{"/f1": "/f2"}},
		b:    WorkspaceOp{Edit: map[string]EditOps{"/f1": EditOps{{S: "x"}}}},
		want: WorkspaceOp{Copy: map[string]string{"/f1": "/f2"}, Edit: map[string]EditOps{"/f1": EditOps{{S: "x"}}}},
	},
	"edit copied and edited file": {
		a:    WorkspaceOp{Copy: map[string]string{"/f1": "/f2"}, Edit: map[string]EditOps{"/f1": EditOps{{S: "x"}}, "/f2": EditOps{{S: "w"}}}},
		b:    WorkspaceOp{Edit: map[string]EditOps{"/f1": EditOps{{N: 1}, {S: "y"}}, "/f2": EditOps{{N: 1}, {S: "z"}}}},
		want: WorkspaceOp{Copy: map[string]string{"/f1": "/f2"}, Edit: map[string]EditOps{"/f1": EditOps{{S: "xy"}}, "/f2": EditOps{{S: "wz"}}}},
	},
	"edit renamed source file": {
		a:       WorkspaceOp{Rename: map[string]string{"/f1": "/f2"}},
		b:       WorkspaceOp{Edit: map[string]EditOps{"/f1": EditOps{{S: "x"}}}},
		wantErr: true,
	},
	"edit renamed dest file": {
		a:    WorkspaceOp{Rename: map[string]string{"/f1": "/f2"}},
		b:    WorkspaceOp{Edit: map[string]EditOps{"/f2": EditOps{{S: "x"}}}},
		want: WorkspaceOp{Rename: map[string]string{"/f1": "/f2"}, Edit: map[string]EditOps{"/f2": EditOps{{S: "x"}}}},
	},
	"edit saved buffer": {
		a:    WorkspaceOp{Save: []string{"#f1"}},
		b:    WorkspaceOp{Edit: map[string]EditOps{"#f1": EditOps{{S: "x"}}}},
		want: WorkspaceOp{Save: []string{"#f1"}, Edit: map[string]EditOps{"#f1": EditOps{{S: "x"}}}},
	},
	"edit file of saved buffer": {
		a:    WorkspaceOp{Save: []string{"#f1"}},
		b:    WorkspaceOp{Edit: map[string]EditOps{"/f1": EditOps{{S: "x"}}}},
		want: WorkspaceOp{Save: []string{"#f1"}, Edit: map[string]EditOps{"/f1": EditOps{{S: "x"}}}},
	},
	"rename-edit renamed and edited dest file": {
		a:    WorkspaceOp{Rename: map[string]string{"/f1": "/f2"}, Edit: map[string]EditOps{"/f2": EditOps{{S: "x"}}}},
		b:    WorkspaceOp{Rename: map[string]string{"/f2": "/f3"}, Edit: map[string]EditOps{"/f3": EditOps{{N: 1}, {S: "y"}}}},
		want: WorkspaceOp{Rename: map[string]string{"/f1": "/f3"}, Edit: map[string]EditOps{"/f3": EditOps{{S: "xy"}}}},
	},
	"edit multiple files": {
		a:    WorkspaceOp{Edit: map[string]EditOps{"/f1": EditOps{{S: "w"}}, "/f2": EditOps{{S: "x"}}}},
		b:    WorkspaceOp{Edit: map[string]EditOps{"/f2": EditOps{{N: 1}, {S: "y"}}, "/f3": EditOps{{S: "z"}}}},
		want: WorkspaceOp{Edit: map[string]EditOps{"/f1": EditOps{{S: "w"}}, "/f2": EditOps{{S: "xy"}}, "/f3": EditOps{{S: "z"}}}},
	},
	"edit (insert) sel file": {
		a: WorkspaceOp{Sel: map[string]map[string]*Sel{"#f": {"u": {1, 2}}}},
		b: WorkspaceOp{Edit: map[string]EditOps{"#f": EditOps{{S: "x"}, {N: 2}}}},
		want: WorkspaceOp{
			Edit: map[string]EditOps{"#f": EditOps{{S: "x"}, {N: 2}}},
			Sel:  map[string]map[string]*Sel{"#f": {"u": {2, 3}}},
		},
	},
	"edit (insert+delete) sel file": {
		a: WorkspaceOp{Sel: map[string]map[string]*Sel{"#f": {"u": {1, 2}}}},
		b: WorkspaceOp{Edit: map[string]EditOps{"#f": EditOps{{S: "xy"}, {N: -1}, {N: 2}}}},
		want: WorkspaceOp{
			Edit: map[string]EditOps{"#f": EditOps{{S: "xy"}, {N: -1}, {N: 2}}},
			Sel:  map[string]map[string]*Sel{"#f": {"u": {2, 3}}},
		},
	},

	"save buffer": {
		a:           WorkspaceOp{Save: []string{"#f"}},
		want:        WorkspaceOp{Save: []string{"#f"}},
		commutative: true,
	},
	"save multiple buffers": {
		a:    WorkspaceOp{Save: []string{"#f1"}},
		b:    WorkspaceOp{Save: []string{"#f2"}},
		want: WorkspaceOp{Save: []string{"#f1", "#f2"}},
	},
	"save edited buffer": {
		a: WorkspaceOp{Edit: map[string]EditOps{"#f": EditOps{{S: "x"}}}},
		b: WorkspaceOp{Save: []string{"#f"}},
		want: WorkspaceOp{
			Save: []string{"#f"},
			Edit: map[string]EditOps{"#f": EditOps{{S: "x"}}, "/f": EditOps{{S: "x"}}},
		},
	},
	"save buffer to edited file": {
		a:    WorkspaceOp{Edit: map[string]EditOps{"/f": EditOps{{S: "x"}}}},
		b:    WorkspaceOp{Save: []string{"#f"}},
		want: WorkspaceOp{Save: []string{"#f"}},
	},
	"save-edit edited buffer": {
		a: WorkspaceOp{Edit: map[string]EditOps{"#f": EditOps{{S: "x"}}}},
		b: WorkspaceOp{Save: []string{"#f"}, Edit: map[string]EditOps{"#f": EditOps{{N: 1}, {S: "y"}}}},
		want: WorkspaceOp{
			Save: []string{"#f"},
			Edit: map[string]EditOps{"#f": EditOps{{S: "xy"}}, "/f": EditOps{{S: "x"}}},
		},
	},
	"save-edit file with edited buffer": {
		a: WorkspaceOp{Edit: map[string]EditOps{"#f": EditOps{{S: "x"}}}},
		b: WorkspaceOp{Save: []string{"#f"}, Edit: map[string]EditOps{"/f": EditOps{{N: 1}, {S: "y"}}}},
		want: WorkspaceOp{
			Save: []string{"#f"},
			Edit: map[string]EditOps{"#f": EditOps{{S: "x"}}, "/f": EditOps{{S: "xy"}}},
		},
	},
	"save deleted file": {
		a:    WorkspaceOp{Delete: []string{"/f"}},
		b:    WorkspaceOp{Save: []string{"#f"}},
		want: WorkspaceOp{Save: []string{"#f"}},
	},
	"save-rename edited buffer": {
		a: WorkspaceOp{Edit: map[string]EditOps{"#f1": EditOps{{S: "x"}}}},
		b: WorkspaceOp{Save: []string{"#f1"}, Rename: map[string]string{"/f1": "/f2"}},
		want: WorkspaceOp{
			Save:   []string{"#f1"},
			Rename: map[string]string{"/f1": "/f2"},
			Edit:   map[string]EditOps{"#f1": EditOps{{S: "x"}}, "/f1": EditOps{{S: "x"}}},
		},
	},
	"save to deleted file": {
		a:    WorkspaceOp{Delete: []string{"/f"}},
		b:    WorkspaceOp{Save: []string{"#f"}},
		want: WorkspaceOp{Save: []string{"#f"}},
	},
	"save truncated buffer": {
		a:    WorkspaceOp{Truncate: []string{"#f"}},
		b:    WorkspaceOp{Save: []string{"#f"}},
		want: WorkspaceOp{Truncate: []string{"#f", "/f"}},
	},
	"save to truncated file": {
		a:    WorkspaceOp{Truncate: []string{"/f"}},
		b:    WorkspaceOp{Save: []string{"#f"}},
		want: WorkspaceOp{Save: []string{"#f"}},
	},
	"save created file": {
		a:    WorkspaceOp{Create: []string{"/f"}},
		b:    WorkspaceOp{Save: []string{"#f"}},
		want: WorkspaceOp{Save: []string{"#f"}},
	},
	"save created-edited file": {
		a:    WorkspaceOp{Create: []string{"/f"}, Edit: map[string]EditOps{"#f": EditOps{{S: "x"}}}},
		b:    WorkspaceOp{Save: []string{"#f"}},
		want: WorkspaceOp{Save: []string{"#f"}, Edit: map[string]EditOps{"#f": EditOps{{S: "x"}}, "/f": EditOps{{S: "x"}}}},
	},
	"save renamed file": {
		a:    WorkspaceOp{Rename: map[string]string{"/f1": "/f2"}},
		b:    WorkspaceOp{Save: []string{"#f1"}},
		want: WorkspaceOp{Rename: map[string]string{"/f1": "/f2"}, Save: []string{"#f1"}},
	},
	"save copied file": {
		a:    WorkspaceOp{Copy: map[string]string{"/f2": "/f1"}},
		b:    WorkspaceOp{Save: []string{"#f1"}},
		want: WorkspaceOp{Copy: map[string]string{"/f2": "/f1"}, Save: []string{"#f1"}},
	},
	"save copied buffer": {
		a:    WorkspaceOp{Copy: map[string]string{"#f": "/f"}},
		b:    WorkspaceOp{Save: []string{"#f"}},
		want: WorkspaceOp{},
	},
	"save copied-edited buffer": {
		a:    WorkspaceOp{Copy: map[string]string{"#f": "/f"}, Edit: map[string]EditOps{"#f": EditOps{{S: "x"}}}},
		b:    WorkspaceOp{Save: []string{"#f"}},
		want: WorkspaceOp{Edit: map[string]EditOps{"/f": EditOps{{S: "x"}}}},
	},
	"save renamed dest buffer": {
		a:    WorkspaceOp{Rename: map[string]string{"/f1": "/f2"}},
		b:    WorkspaceOp{Save: []string{"#f2"}},
		want: WorkspaceOp{Save: []string{"#f2"}, Delete: []string{"/f1"}},
	},
	"save renamed and edited buffer": {
		a:    WorkspaceOp{Rename: map[string]string{"/f1": "/f2"}, Edit: map[string]EditOps{"#f2": EditOps{{S: "x"}}}},
		b:    WorkspaceOp{Save: []string{"#f2"}},
		want: WorkspaceOp{Save: []string{"#f2"}, Delete: []string{"/f1"}, Edit: map[string]EditOps{"#f2": EditOps{{S: "x"}}, "/f2": EditOps{{S: "x"}}}},
	},

	"sel same file, same user": {
		a:    WorkspaceOp{Sel: map[string]map[string]*Sel{"/f1": {"u1": {1, 2}}}},
		b:    WorkspaceOp{Sel: map[string]map[string]*Sel{"/f1": {"u1": {3, 4}}}},
		want: WorkspaceOp{Sel: map[string]map[string]*Sel{"/f1": {"u1": {3, 4}}}},
	},
	"sel same file, diff user": {
		a:    WorkspaceOp{Sel: map[string]map[string]*Sel{"/f1": {"u1": {1, 2}}}},
		b:    WorkspaceOp{Sel: map[string]map[string]*Sel{"/f1": {"u2": {3, 4}}}},
		want: WorkspaceOp{Sel: map[string]map[string]*Sel{"/f1": {"u1": {1, 2}, "u2": {3, 4}}}},
	},
	"sel different file, same user": {
		a:    WorkspaceOp{Sel: map[string]map[string]*Sel{"/f1": {"u1": {1, 2}}}},
		b:    WorkspaceOp{Sel: map[string]map[string]*Sel{"/f2": {"u1": {3, 4}}}},
		want: WorkspaceOp{Sel: map[string]map[string]*Sel{"/f1": {"u1": {1, 2}}, "/f2": {"u1": {3, 4}}}},
	},
	"sel edited file": {
		a: WorkspaceOp{Edit: map[string]EditOps{"#f": EditOps{{S: "x"}, {N: 2}}}},
		b: WorkspaceOp{Sel: map[string]map[string]*Sel{"#f": {"u": {1, 2}}}},
		want: WorkspaceOp{
			Edit: map[string]EditOps{"#f": EditOps{{S: "x"}, {N: 2}}},
			Sel:  map[string]map[string]*Sel{"#f": {"u": {1, 2}}},
		},
	},
	"sel-copy sel file": {
		a: WorkspaceOp{Sel: map[string]map[string]*Sel{"/f1": {"u": {1, 2}}}},
		b: WorkspaceOp{Copy: map[string]string{"/f2": "/f1"}, Sel: map[string]map[string]*Sel{"/f2": {"u": {3, 4}}}},
		want: WorkspaceOp{
			Copy: map[string]string{"/f2": "/f1"},
			Sel:  map[string]map[string]*Sel{"/f1": {"u": {1, 2}}, "/f2": {"u": {3, 4}}},
		},
	},
	"sel-desel same file": {
		a:    WorkspaceOp{Sel: map[string]map[string]*Sel{"/f1": {"u1": {1, 2}}}},
		b:    WorkspaceOp{Sel: map[string]map[string]*Sel{"/f1": {"u1": nil}}},
		want: WorkspaceOp{Sel: map[string]map[string]*Sel{"/f1": {"u1": nil}}},
	},
	"sel-desel with other users": {
		a:    WorkspaceOp{Sel: map[string]map[string]*Sel{"/f1": {"u1": {1, 2}, "u2": {3, 4}}}},
		b:    WorkspaceOp{Sel: map[string]map[string]*Sel{"/f1": {"u1": nil}}},
		want: WorkspaceOp{Sel: map[string]map[string]*Sel{"/f1": {"u1": nil, "u2": {3, 4}}}},
	},

	"no head": {
		a:    WorkspaceOp{GitHead: ""},
		b:    WorkspaceOp{GitHead: ""},
		want: WorkspaceOp{GitHead: ""},
	},
	"both update head": {
		a:    WorkspaceOp{GitHead: "ha"},
		b:    WorkspaceOp{GitHead: "hb"},
		want: WorkspaceOp{GitHead: "hb"},
	},
	"update head once": {
		a:           WorkspaceOp{GitHead: "ha"},
		b:           WorkspaceOp{GitHead: ""},
		want:        WorkspaceOp{GitHead: "ha"},
		commutative: true,
	},
}

var transformTests = map[string]struct {
	a, b, a1, b1 WorkspaceOp
	wantErr      bool
	commutative  bool
}{
	"empty": {
		a:  WorkspaceOp{},
		b:  WorkspaceOp{},
		a1: WorkspaceOp{},
		b1: WorkspaceOp{},
	},

	"copy same source and dest": {
		a: WorkspaceOp{Copy: map[string]string{"/f1": "/f2"}},
		b: WorkspaceOp{Copy: map[string]string{"/f1": "/f2"}},
	},
	"copy independently": {
		a:  WorkspaceOp{Copy: map[string]string{"/f1": "/f2"}},
		b:  WorkspaceOp{Copy: map[string]string{"/f3": "/f4"}},
		a1: WorkspaceOp{Copy: map[string]string{"/f1": "/f2"}},
		b1: WorkspaceOp{Copy: map[string]string{"/f3": "/f4"}},
	},
	"copy same source": {
		a:  WorkspaceOp{Copy: map[string]string{"/f1": "/f2"}},
		b:  WorkspaceOp{Copy: map[string]string{"/f3": "/f2"}},
		a1: WorkspaceOp{Copy: map[string]string{"/f1": "/f2"}},
		b1: WorkspaceOp{Copy: map[string]string{"/f3": "/f2"}},
	},
	"copy same dest": {
		a:       WorkspaceOp{Copy: map[string]string{"/f1": "/f2"}},
		b:       WorkspaceOp{Copy: map[string]string{"/f1": "/f3"}},
		wantErr: true,
	},
	"copy multiple": {
		a:  WorkspaceOp{Copy: map[string]string{"/f1": "/f2", "/f3": "/f4"}},
		b:  WorkspaceOp{Copy: map[string]string{"/f1": "/f2", "/f5": "/f6"}},
		a1: WorkspaceOp{Copy: map[string]string{"/f3": "/f4"}},
		b1: WorkspaceOp{Copy: map[string]string{"/f5": "/f6"}},
	},
	"copy, create dest": {
		a:           WorkspaceOp{Copy: map[string]string{"/f2": "/f1"}},
		b:           WorkspaceOp{Create: []string{"/f2"}},
		wantErr:     true,
		commutative: true,
	},
	"copy, create source": {
		a:           WorkspaceOp{Copy: map[string]string{"/f2": "/f1"}},
		b:           WorkspaceOp{Create: []string{"/f1"}},
		wantErr:     true,
		commutative: true,
	},
	"copy, edit": {
		a:           WorkspaceOp{Copy: map[string]string{"/f2": "/f1"}},
		b:           WorkspaceOp{Edit: map[string]EditOps{"/f1": EditOps{{S: "x"}}}},
		a1:          WorkspaceOp{Copy: map[string]string{"/f2": "/f1"}},
		b1:          WorkspaceOp{Edit: map[string]EditOps{"/f1": EditOps{{S: "x"}}, "/f2": EditOps{{S: "x"}}}},
		commutative: true,
	},
	"copy-edit, edit": {
		a:  WorkspaceOp{Copy: map[string]string{"/f2": "/f1"}, Edit: map[string]EditOps{"/f2": EditOps{{S: "x"}}}},
		b:  WorkspaceOp{Edit: map[string]EditOps{"/f1": EditOps{{S: "y"}}}},
		a1: WorkspaceOp{Copy: map[string]string{"/f2": "/f1"}, Edit: map[string]EditOps{"/f2": EditOps{{S: "x"}, {N: 1}}}},
		b1: WorkspaceOp{Edit: map[string]EditOps{"/f1": EditOps{{S: "y"}}, "/f2": EditOps{{N: 1}, {S: "y"}}}},
	},
	"copy-edit, edit (reversed)": {
		a:  WorkspaceOp{Edit: map[string]EditOps{"/f1": EditOps{{S: "x"}}}},
		b:  WorkspaceOp{Copy: map[string]string{"/f2": "/f1"}, Edit: map[string]EditOps{"/f2": EditOps{{S: "y"}}}},
		a1: WorkspaceOp{Edit: map[string]EditOps{"/f1": EditOps{{S: "x"}}, "/f2": EditOps{{S: "x"}, {N: 1}}}},
		b1: WorkspaceOp{Copy: map[string]string{"/f2": "/f1"}, Edit: map[string]EditOps{"/f2": EditOps{{N: 1}, {S: "y"}}}},
	},
	"copy-edit(src)-edit(dest), edit": {
		a:  WorkspaceOp{Copy: map[string]string{"/f2": "/f1"}, Edit: map[string]EditOps{"/f1": EditOps{{S: "x"}}, "/f2": EditOps{{S: "y"}}}},
		b:  WorkspaceOp{Edit: map[string]EditOps{"/f1": EditOps{{S: "z"}}}},
		a1: WorkspaceOp{Copy: map[string]string{"/f2": "/f1"}, Edit: map[string]EditOps{"/f1": EditOps{{S: "x"}, {N: 1}}, "/f2": EditOps{{S: "y"}, {N: 1}}}},
		b1: WorkspaceOp{Edit: map[string]EditOps{"/f1": EditOps{{N: 1}, {S: "z"}}, "/f2": EditOps{{N: 1}, {S: "z"}}}},
	},
	"copy-edit(src)-edit(dest), edit (reversed)": {
		a:  WorkspaceOp{Edit: map[string]EditOps{"/f1": EditOps{{S: "x"}}}},
		b:  WorkspaceOp{Copy: map[string]string{"/f2": "/f1"}, Edit: map[string]EditOps{"/f1": EditOps{{S: "y"}}, "/f2": EditOps{{S: "z"}}}},
		a1: WorkspaceOp{Edit: map[string]EditOps{"/f1": EditOps{{S: "x"}, {N: 1}}, "/f2": EditOps{{S: "x"}, {N: 1}}}},
		b1: WorkspaceOp{Copy: map[string]string{"/f2": "/f1"}, Edit: map[string]EditOps{"/f1": EditOps{{N: 1}, {S: "y"}}, "/f2": EditOps{{N: 1}, {S: "z"}}}},
	},

	"rename same source and dest": {
		a: WorkspaceOp{Rename: map[string]string{"/f1": "/f2"}},
		b: WorkspaceOp{Rename: map[string]string{"/f1": "/f2"}},
	},
	"rename independently": {
		a:  WorkspaceOp{Rename: map[string]string{"/f1": "/f2"}},
		b:  WorkspaceOp{Rename: map[string]string{"/f3": "/f4"}},
		a1: WorkspaceOp{Rename: map[string]string{"/f1": "/f2"}},
		b1: WorkspaceOp{Rename: map[string]string{"/f3": "/f4"}},
	},
	"rename same source": {
		a:  WorkspaceOp{Rename: map[string]string{"/f1": "/f2"}},
		b:  WorkspaceOp{Rename: map[string]string{"/f1": "/f3"}},
		a1: WorkspaceOp{Copy: map[string]string{"/f2": "/f3"}},
		b1: WorkspaceOp{Copy: map[string]string{"/f3": "/f2"}},
	},
	"rename same dest": {
		a:       WorkspaceOp{Rename: map[string]string{"/f1": "/f2"}},
		b:       WorkspaceOp{Rename: map[string]string{"/f3": "/f2"}},
		wantErr: true,
	},
	"rename multiple": {
		a:  WorkspaceOp{Rename: map[string]string{"/f1": "/f2", "/f3": "/f4"}},
		b:  WorkspaceOp{Rename: map[string]string{"/f1": "/f2", "/f5": "/f6"}},
		a1: WorkspaceOp{Rename: map[string]string{"/f3": "/f4"}},
		b1: WorkspaceOp{Rename: map[string]string{"/f5": "/f6"}},
	},
	"rename, copy": {
		a:           WorkspaceOp{Rename: map[string]string{"/f1": "/f2"}},
		b:           WorkspaceOp{Copy: map[string]string{"/f2": "/f1"}},
		a1:          WorkspaceOp{Delete: []string{"/f1"}},
		b1:          WorkspaceOp{},
		commutative: true,
	},
	"rename-edit, copy-edit": {
		a:           WorkspaceOp{Rename: map[string]string{"/f1": "/f2"}, Edit: map[string]EditOps{"/f2": EditOps{{S: "x"}}}},
		b:           WorkspaceOp{Copy: map[string]string{"/f2": "/f1"}, Edit: map[string]EditOps{"/f1": EditOps{{S: "y"}}, "/f2": EditOps{{S: "z"}}}},
		wantErr:     true,
		commutative: true,
	},
	"rename, create dest": {
		a:           WorkspaceOp{Rename: map[string]string{"/f1": "/f2"}},
		b:           WorkspaceOp{Create: []string{"/f2"}},
		wantErr:     true,
		commutative: true,
	},
	"rename, create source": {
		a:           WorkspaceOp{Rename: map[string]string{"/f1": "/f2"}},
		b:           WorkspaceOp{Create: []string{"/f1"}},
		wantErr:     true,
		commutative: true,
	},
	"rename, edit": {
		a:  WorkspaceOp{Rename: map[string]string{"/f1": "/f2"}},
		b:  WorkspaceOp{Edit: map[string]EditOps{"/f1": EditOps{{S: "x"}}}},
		a1: WorkspaceOp{Rename: map[string]string{"/f1": "/f2"}},
		b1: WorkspaceOp{Edit: map[string]EditOps{"/f2": EditOps{{S: "x"}}}},
	},
	"rename-edit, edit": {
		a:  WorkspaceOp{Rename: map[string]string{"/f1": "/f2"}, Edit: map[string]EditOps{"/f2": EditOps{{S: "x"}}}},
		b:  WorkspaceOp{Edit: map[string]EditOps{"/f1": EditOps{{S: "y"}}}},
		a1: WorkspaceOp{Rename: map[string]string{"/f1": "/f2"}, Edit: map[string]EditOps{"/f2": EditOps{{S: "x"}, {N: 1}}}},
		b1: WorkspaceOp{Edit: map[string]EditOps{"/f2": EditOps{{N: 1}, {S: "y"}}}},
	},

	"save same buffer": {
		a: WorkspaceOp{Save: []string{"#f"}},
		b: WorkspaceOp{Save: []string{"#f"}},
	},
	"save independently": {
		a:  WorkspaceOp{Save: []string{"#f1"}},
		b:  WorkspaceOp{Save: []string{"#f2"}},
		a1: WorkspaceOp{Save: []string{"#f1"}},
		b1: WorkspaceOp{Save: []string{"#f2"}},
	},
	"save multiple": {
		a:  WorkspaceOp{Save: []string{"#f1", "#f2", "#f3"}},
		b:  WorkspaceOp{Save: []string{"#f1", "#f4", "#f5"}},
		a1: WorkspaceOp{Save: []string{"#f2", "#f3"}},
		b1: WorkspaceOp{Save: []string{"#f4", "#f5"}},
	},
	"save, create dest": {
		a:           WorkspaceOp{Save: []string{"#f"}},
		b:           WorkspaceOp{Create: []string{"/f"}},
		wantErr:     true,
		commutative: true,
	},
	"save, edit": {
		a:           WorkspaceOp{Save: []string{"#f"}},
		b:           WorkspaceOp{Edit: map[string]EditOps{"#f": EditOps{{S: "x"}}}},
		a1:          WorkspaceOp{Save: []string{"#f"}},
		b1:          WorkspaceOp{Edit: map[string]EditOps{"#f": EditOps{{S: "x"}}, "/f": EditOps{{S: "x"}}}},
		commutative: true,
	},
	"save-edit file, edit buffer": {
		a:  WorkspaceOp{Save: []string{"#f"}, Edit: map[string]EditOps{"/f": EditOps{{S: "x"}}}},
		b:  WorkspaceOp{Edit: map[string]EditOps{"#f": EditOps{{S: "y"}}}},
		a1: WorkspaceOp{Save: []string{"#f"}, Edit: map[string]EditOps{"/f": EditOps{{S: "x"}, {N: 1}}}},
		b1: WorkspaceOp{Edit: map[string]EditOps{"#f": EditOps{{S: "y"}}, "/f": EditOps{{N: 1}, {S: "y"}}}},
	},
	"save-edit file, edit file": {
		a:  WorkspaceOp{Save: []string{"#f"}, Edit: map[string]EditOps{"/f": EditOps{{S: "x"}}}},
		b:  WorkspaceOp{Edit: map[string]EditOps{"/f": EditOps{{S: "y"}}}},
		a1: WorkspaceOp{Save: []string{"#f"}, Edit: map[string]EditOps{"/f": EditOps{{S: "xy"}}}},
		b1: WorkspaceOp{Edit: map[string]EditOps{"/f": EditOps{{N: 1}, {S: "y"}}}},
	},
	"save-edit buffer, edit file": {
		a:  WorkspaceOp{Save: []string{"#f"}, Edit: map[string]EditOps{"#f": EditOps{{S: "x"}}}},
		b:  WorkspaceOp{Edit: map[string]EditOps{"/f": EditOps{{S: "y"}}}},
		a1: WorkspaceOp{Save: []string{"#f"}, Edit: map[string]EditOps{"#f": EditOps{{S: "x"}}, "/f": EditOps{{S: "y"}}}},
		b1: WorkspaceOp{Edit: map[string]EditOps{"/f": EditOps{{S: "y"}}}},
	},
	"save-edit buffer and file, edit file": {
		a:  WorkspaceOp{Save: []string{"#f"}, Edit: map[string]EditOps{"#f": EditOps{{S: "x"}}, "/f": EditOps{{S: "y"}}}},
		b:  WorkspaceOp{Edit: map[string]EditOps{"/f": EditOps{{S: "z"}}}},
		a1: WorkspaceOp{Save: []string{"#f"}, Edit: map[string]EditOps{"#f": EditOps{{S: "x"}}, "/f": EditOps{{S: "yz"}}}},
		b1: WorkspaceOp{Edit: map[string]EditOps{"/f": EditOps{{N: 1}, {S: "z"}}}},
	},

	"create same": {
		a: WorkspaceOp{Create: []string{"/f"}},
		b: WorkspaceOp{Create: []string{"/f"}},
	},
	"create independently": {
		a:  WorkspaceOp{Create: []string{"/f1"}},
		b:  WorkspaceOp{Create: []string{"/f2"}},
		a1: WorkspaceOp{Create: []string{"/f1"}},
		b1: WorkspaceOp{Create: []string{"/f2"}},
	},
	"create multiple": {
		a:  WorkspaceOp{Create: []string{"/f1", "/f2"}},
		b:  WorkspaceOp{Create: []string{"/f1", "/f3"}},
		a1: WorkspaceOp{Create: []string{"/f2"}},
		b1: WorkspaceOp{Create: []string{"/f3"}},
	},
	"create, edit": {
		a:           WorkspaceOp{Create: []string{"/f"}},
		b:           WorkspaceOp{Edit: map[string]EditOps{"/f": EditOps{{S: "x"}}}},
		wantErr:     true,
		commutative: true,
	},
	"create-edit, edit": {
		a:           WorkspaceOp{Create: []string{"/f"}, Edit: map[string]EditOps{"/f": EditOps{{S: "x"}}}},
		b:           WorkspaceOp{Edit: map[string]EditOps{"/f": EditOps{{S: "x"}}}},
		wantErr:     true,
		commutative: true,
	},
	"create-edit identical": {
		a: WorkspaceOp{Create: []string{"/f"}, Edit: map[string]EditOps{"/f": EditOps{{S: "x"}}}},
		b: WorkspaceOp{Create: []string{"/f"}, Edit: map[string]EditOps{"/f": EditOps{{S: "x"}}}},
	},
	"create-edit conflicting": {
		a:  WorkspaceOp{Create: []string{"/f"}, Edit: map[string]EditOps{"/f": EditOps{{S: "x"}}}},
		b:  WorkspaceOp{Create: []string{"/f"}, Edit: map[string]EditOps{"/f": EditOps{{S: "y"}}}},
		a1: WorkspaceOp{Edit: map[string]EditOps{"/f": EditOps{{S: "x"}, {N: 1}}}},
		b1: WorkspaceOp{Edit: map[string]EditOps{"/f": EditOps{{N: 1}, {S: "y"}}}},
	},
	"create, delete": {
		a:           WorkspaceOp{Create: []string{"/f"}},
		b:           WorkspaceOp{Delete: []string{"/f"}},
		wantErr:     true,
		commutative: true,
	},
	"create, truncate": {
		a:           WorkspaceOp{Create: []string{"/f"}},
		b:           WorkspaceOp{Truncate: []string{"/f"}},
		wantErr:     true,
		commutative: true,
	},

	"delete same": {
		a: WorkspaceOp{Delete: []string{"/f"}},
		b: WorkspaceOp{Delete: []string{"/f"}},
	},
	"delete independently": {
		a:  WorkspaceOp{Delete: []string{"/f1"}},
		b:  WorkspaceOp{Delete: []string{"/f2"}},
		a1: WorkspaceOp{Delete: []string{"/f1"}},
		b1: WorkspaceOp{Delete: []string{"/f2"}},
	},
	"delete multiple": {
		a:  WorkspaceOp{Delete: []string{"/f1", "/f2"}},
		b:  WorkspaceOp{Delete: []string{"/f1", "/f3"}},
		a1: WorkspaceOp{Delete: []string{"/f2"}},
		b1: WorkspaceOp{Delete: []string{"/f3"}},
	},
	"delete, edit": {
		a:           WorkspaceOp{Delete: []string{"/f"}},
		b:           WorkspaceOp{Edit: map[string]EditOps{"/f": EditOps{{S: "x"}}}},
		wantErr:     true,
		commutative: true,
	},
	"delete, create": {
		a:           WorkspaceOp{Delete: []string{"/f"}},
		b:           WorkspaceOp{Create: []string{"/f"}},
		wantErr:     true,
		commutative: true,
	},
	"delete, truncate": {
		a:           WorkspaceOp{Delete: []string{"/f"}},
		b:           WorkspaceOp{Truncate: []string{"/f"}},
		wantErr:     true,
		commutative: true,
	},

	"truncate same": {
		a: WorkspaceOp{Truncate: []string{"/f"}},
		b: WorkspaceOp{Truncate: []string{"/f"}},
	},
	"truncate independently": {
		a:  WorkspaceOp{Truncate: []string{"/f1"}},
		b:  WorkspaceOp{Truncate: []string{"/f2"}},
		a1: WorkspaceOp{Truncate: []string{"/f1"}},
		b1: WorkspaceOp{Truncate: []string{"/f2"}},
	},
	"truncate multiple": {
		a:  WorkspaceOp{Truncate: []string{"/f1", "/f2"}},
		b:  WorkspaceOp{Truncate: []string{"/f1", "/f3"}},
		a1: WorkspaceOp{Truncate: []string{"/f2"}},
		b1: WorkspaceOp{Truncate: []string{"/f3"}},
	},
	"truncate, edit": {
		a:           WorkspaceOp{Truncate: []string{"/f"}},
		b:           WorkspaceOp{Edit: map[string]EditOps{"/f": EditOps{{S: "x"}}}},
		wantErr:     true,
		commutative: true,
	},
	"truncate-edit, edit": {
		a:           WorkspaceOp{Truncate: []string{"/f"}, Edit: map[string]EditOps{"/f": EditOps{{S: "x"}}}},
		b:           WorkspaceOp{Edit: map[string]EditOps{"/f": EditOps{{S: "x"}}}},
		wantErr:     true,
		commutative: true,
	},
	"truncate-edit identical": {
		a: WorkspaceOp{Truncate: []string{"/f"}, Edit: map[string]EditOps{"/f": EditOps{{S: "x"}}}},
		b: WorkspaceOp{Truncate: []string{"/f"}, Edit: map[string]EditOps{"/f": EditOps{{S: "x"}}}},
	},
	"truncate-edit conflicting": {
		a:       WorkspaceOp{Truncate: []string{"/f"}, Edit: map[string]EditOps{"/f": EditOps{{S: "x"}}}},
		b:       WorkspaceOp{Truncate: []string{"/f"}, Edit: map[string]EditOps{"/f": EditOps{{S: "y"}}}},
		wantErr: true,
	},
	"truncate, create": {
		a:           WorkspaceOp{Truncate: []string{"/f"}},
		b:           WorkspaceOp{Create: []string{"/f"}},
		wantErr:     true,
		commutative: true,
	},
	"truncate, delete": {
		a:           WorkspaceOp{Truncate: []string{"/f"}},
		b:           WorkspaceOp{Delete: []string{"/f"}},
		wantErr:     true,
		commutative: true,
	},

	"edit same": {
		a: WorkspaceOp{Edit: map[string]EditOps{"/f": EditOps{{S: "x"}}}},
		b: WorkspaceOp{Edit: map[string]EditOps{"/f": EditOps{{S: "x"}}}},
	},
	"edit independently": {
		a:  WorkspaceOp{Edit: map[string]EditOps{"/f1": EditOps{{S: "x"}}}},
		b:  WorkspaceOp{Edit: map[string]EditOps{"/f2": EditOps{{S: "y"}}}},
		a1: WorkspaceOp{Edit: map[string]EditOps{"/f1": EditOps{{S: "x"}}}},
		b1: WorkspaceOp{Edit: map[string]EditOps{"/f2": EditOps{{S: "y"}}}},
	},
	"edit same file": {
		a:  WorkspaceOp{Edit: map[string]EditOps{"/f": EditOps{{S: "x"}}}},
		b:  WorkspaceOp{Edit: map[string]EditOps{"/f": EditOps{{S: "y"}}}},
		a1: WorkspaceOp{Edit: map[string]EditOps{"/f": EditOps{{S: "x"}, {N: 1}}}},
		b1: WorkspaceOp{Edit: map[string]EditOps{"/f": EditOps{{N: 1}, {S: "y"}}}},
	},
	"edit invalid": {
		a:           WorkspaceOp{Edit: map[string]EditOps{"/f": EditOps{{N: 1}, {S: "x"}}}},
		b:           WorkspaceOp{Edit: map[string]EditOps{"/f": EditOps{{N: 1234}, {S: "y"}}}},
		wantErr:     true,
		commutative: true,
	},

	"sel same file, same user": {
		a:  WorkspaceOp{Sel: map[string]map[string]*Sel{"/f": {"u": {1, 2}}}},
		b:  WorkspaceOp{Sel: map[string]map[string]*Sel{"/f": {"u": {3, 4}}}},
		a1: WorkspaceOp{Sel: map[string]map[string]*Sel{"/f": {"u": {1, 2}}}},
		b1: WorkspaceOp{},
	},
	"sel same file, diff user": {
		a:  WorkspaceOp{Sel: map[string]map[string]*Sel{"/f": {"u1": {1, 2}}}},
		b:  WorkspaceOp{Sel: map[string]map[string]*Sel{"/f": {"u2": {3, 4}}}},
		a1: WorkspaceOp{Sel: map[string]map[string]*Sel{"/f": {"u1": {1, 2}}}},
		b1: WorkspaceOp{Sel: map[string]map[string]*Sel{"/f": {"u2": {3, 4}}}},
	},
	"sel different file, same user": {
		a:  WorkspaceOp{Sel: map[string]map[string]*Sel{"/f1": {"u1": {1, 2}}}},
		b:  WorkspaceOp{Sel: map[string]map[string]*Sel{"/f2": {"u1": {3, 4}}}},
		a1: WorkspaceOp{Sel: map[string]map[string]*Sel{"/f1": {"u1": {1, 2}}}},
		b1: WorkspaceOp{Sel: map[string]map[string]*Sel{"/f2": {"u1": {3, 4}}}},
	},
	"sel edited file": {
		a: WorkspaceOp{Edit: map[string]EditOps{"#f": EditOps{{S: "x"}, {N: 2}}}},
		b: WorkspaceOp{Sel: map[string]map[string]*Sel{"#f": {"u": {1, 2}}}},
		a1: WorkspaceOp{
			Edit: map[string]EditOps{"#f": EditOps{{S: "x"}, {N: 2}}},
		},
		b1: WorkspaceOp{
			Sel: map[string]map[string]*Sel{"#f": {"u": {2, 3}}},
		},
	},
	"edit sel file": {
		a: WorkspaceOp{Sel: map[string]map[string]*Sel{"#f": {"u": {1, 2}}}},
		b: WorkspaceOp{Edit: map[string]EditOps{"#f": EditOps{{S: "x"}, {N: 2}}}},
		a1: WorkspaceOp{
			Sel: map[string]map[string]*Sel{"#f": {"u": {2, 3}}},
		},
		b1: WorkspaceOp{
			Edit: map[string]EditOps{"#f": EditOps{{S: "x"}, {N: 2}}},
		},
	},
	"sel renamed file": {
		a:           WorkspaceOp{Rename: map[string]string{"/f1": "/f2"}},
		b:           WorkspaceOp{Sel: map[string]map[string]*Sel{"/f1": {"u": {1, 2}}}},
		a1:          WorkspaceOp{Rename: map[string]string{"/f1": "/f2"}},
		b1:          WorkspaceOp{Sel: map[string]map[string]*Sel{"/f2": {"u": {1, 2}}}},
		commutative: true,
	},
	"sel deleted file": {
		a:           WorkspaceOp{Delete: []string{"/f"}},
		b:           WorkspaceOp{Sel: map[string]map[string]*Sel{"/f": {"u": {1, 2}}}},
		a1:          WorkspaceOp{Delete: []string{"/f"}},
		b1:          WorkspaceOp{},
		commutative: true,
	},
	"sel truncated file": {
		a:           WorkspaceOp{Truncate: []string{"/f"}},
		b:           WorkspaceOp{Sel: map[string]map[string]*Sel{"/f": {"u": {1, 2}}}},
		a1:          WorkspaceOp{Truncate: []string{"/f"}},
		b1:          WorkspaceOp{},
		commutative: true,
	},
	"sel-desel different users": {
		a:           WorkspaceOp{Sel: map[string]map[string]*Sel{"/f1": {"u1": {1, 2}}}},
		b:           WorkspaceOp{Sel: map[string]map[string]*Sel{"/f1": {"u2": nil}}},
		a1:          WorkspaceOp{Sel: map[string]map[string]*Sel{"/f1": {"u1": {1, 2}}}},
		b1:          WorkspaceOp{Sel: map[string]map[string]*Sel{"/f1": {"u2": nil}}},
		commutative: true,
	},

	"same head": {
		a:  WorkspaceOp{GitHead: "h"},
		b:  WorkspaceOp{GitHead: "h"},
		a1: WorkspaceOp{},
		b1: WorkspaceOp{},
	},
	"one has head": {
		a:           WorkspaceOp{GitHead: "h"},
		b:           WorkspaceOp{},
		a1:          WorkspaceOp{GitHead: "h"},
		b1:          WorkspaceOp{},
		commutative: true,
	},
	"conflicting heads": {
		a:       WorkspaceOp{GitHead: "ha"},
		b:       WorkspaceOp{GitHead: "hb"},
		wantErr: true,
	},
}

// ComposeTestJSON represents compose test cases in JSON format.
type ComposeTestJSON struct {
	A           WorkspaceOp `json:"a,omitempty"`
	B           WorkspaceOp `json:"b,omitempty"`
	Want        WorkspaceOp `json:"want,omitempty"`
	WantErr     bool        `json:"wantErr,omitempty"`
	Commutative bool        `json:"commutative,omitempty"`
}

// ComposeTestsToJSON returns the suite of compose tests as a marshaled JSON slice.
func ComposeTestsToJSON() ([]byte, error) {
	tests := make(map[string]ComposeTestJSON)
	for n, t := range composeTests {
		tests[n] = ComposeTestJSON{A: t.a, B: t.b, Want: t.want, WantErr: t.wantErr, Commutative: t.commutative}
	}
	return json.MarshalIndent(tests, "", "\t")
}

// TransformTestJSON represents transform test cases in JSON format.
type TransformTestJSON struct {
	A           WorkspaceOp `json:"a,omitempty"`
	B           WorkspaceOp `json:"b,omitempty"`
	A1          WorkspaceOp `json:"a1,omitempty"`
	B1          WorkspaceOp `json:"b1,omitempty"`
	WantErr     bool        `json:"wantErr,omitempty"`
	Commutative bool        `json:"commutative,omitempty"`
}

// TransformTestsToJSON returns the suite of transform tests as a marshaled JSON slice.
func TransformTestsToJSON() ([]byte, error) {
	tests := make(map[string]TransformTestJSON)
	for n, t := range transformTests {
		tests[n] = TransformTestJSON{A: t.a, B: t.b, A1: t.a1, B1: t.b1, WantErr: t.wantErr, Commutative: t.commutative}
	}
	return json.MarshalIndent(tests, "", "\t")
}

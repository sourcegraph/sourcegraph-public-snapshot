package ot

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// An Op is an operation performed on a workspace. None of its methods
// may modify the receiver or any other Op arguments.
type Op interface {
	// Less reports whether the op should be sorted before other (in
	// the canonical ordering of an op sequence) and can be transposed
	// with other without changing the outcome of the op sequence. If
	// either condition is false, Less must return false.
	Less(other Op) bool

	// Equal reports whether the op equals the other op. Equality is
	// defined as being of the same type and value.
	Equal(other Op) bool

	// Copy returns a deep copy of the op.
	Copy() Op

	// String returns a description of the op.
	String() string
}

// ErrNotComposable is a special error value returned by a
// (Op).Compose implementation indicating that the ops are not
// composable (because they are not valid sequential ops). If
// a.Compose(b) returns ErrNotComposable, Compose keeps a and b as
// separate ops in the Ops list and continues.
var ErrNotComposable = errors.New("not composable")

// A FileOp is an operation performed on a single file.
type FileOp interface {
	file() string
}

// A FilesOp is an operation performed on multiple files or file
// paths.
type FilesOp interface {
	files() []string
}

// referencesFile reports whether the op depends on or affects file.
func referencesFile(op Op, file string) bool {
	if fop, ok := op.(FileOp); ok {
		return fop.file() == file
	}
	if fop, ok := op.(FilesOp); ok {
		for _, f := range fop.files() {
			if f == file {
				return true
			}
		}
	}
	return false
}

// compareToOtherOpFiles returns -1 if file is less than all of
// other's referenced file names, 1 if file is greater than all of other's
// references files, and 0 otherwise.
func compareToOtherOpFiles(file string, other Op) int {
	if other, ok := other.(FileOp); ok {
		return strings.Compare(file, other.file())
	}
	if other, ok := other.(FilesOp); ok {
		// Return 0 if any files overlap.
		for _, f := range other.files() {
			if file == f {
				return 0
			}
		}

		for _, f := range other.files() {
			// Only need 1 iteration since we know that there are no
			// equal files.
			if file < f {
				return -1
			} else {
				return 1
			}
		}
	}
	return 0
}

// A FileCreate is the creation of a file.
type FileCreate struct {
	File string `json:"file"`
}

func (a FileCreate) file() string { return a.File }

// Less implements Op.
func (a FileCreate) Less(other Op) bool {
	return compareToOtherOpFiles(a.File, other) < 0
}

// Equal implements Op.
func (a FileCreate) Equal(other Op) bool {
	b, ok := other.(FileCreate)
	return ok && a == b
}

// Copy implements Op.
func (a FileCreate) Copy() Op { return a }

func (a FileCreate) MarshalJSON() ([]byte, error) {
	type Op FileCreate // Type-alias required to prevent stack-overflow
	return json.Marshal(&struct {
		Type string `json:"type"`
		Op
	}{
		Type: "create",
		Op:   (Op)(a),
	})
}

func (a FileCreate) String() string { return fmt.Sprintf("create(%s)", a.File) }

// A FileTruncate represents an op that deletes all contents from a file.
type FileTruncate struct {
	File string `json:"file"`
}

func (a FileTruncate) file() string { return a.File }

// Less implements Op.
func (a FileTruncate) Less(other Op) bool {
	return compareToOtherOpFiles(a.File, other) < 0
}

// Equal implements Op.
func (a FileTruncate) Equal(other Op) bool {
	b, ok := other.(FileTruncate)
	return ok && a == b
}

func (a FileTruncate) MarshalJSON() ([]byte, error) {
	type Op FileTruncate
	return json.Marshal(&struct {
		Type string `json:"type"`
		Op
	}{
		Type: "truncate",
		Op:   (Op)(a),
	})
}

func (a FileTruncate) Copy() Op { return a }

func (a FileTruncate) String() string { return fmt.Sprintf("truncate(%s)", a.File) }

// A FileDelete is the deletion of a file.
type FileDelete struct {
	File string `json:"file"`
}

func (a FileDelete) file() string { return a.File }

// Less implements Op.
func (a FileDelete) Less(other Op) bool {
	return compareToOtherOpFiles(a.File, other) < 0
}

// Equal implements Op.
func (a FileDelete) Equal(other Op) bool {
	b, ok := other.(FileDelete)
	return ok && a == b
}

// Copy implements Op.
func (a FileDelete) Copy() Op { return a }

func (a FileDelete) MarshalJSON() ([]byte, error) {
	type Op FileDelete
	return json.Marshal(&struct {
		Type string `json:"type"`
		Op
	}{
		Type: "delete",
		Op:   (Op)(a),
	})
}

func (a FileDelete) String() string { return fmt.Sprintf("del(%s)", a.File) }

// A FileEdit is an edit to a single file.
type FileEdit struct {
	File  string  `json:"file"`  // the filename this edit applies to
	Edits EditOps `json:"edits"` // the OT edit ops representing this edit
}

func (a FileEdit) file() string { return a.File }

// Less implements Op.
func (a FileEdit) Less(other Op) bool {
	return compareToOtherOpFiles(a.File, other) < 0
}

// Equal implements Op.
func (a FileEdit) Equal(other Op) bool {
	b, ok := other.(FileEdit)
	return ok && a.File == b.File && a.Edits.Equal(b.Edits)
}

// Copy implements Op.
func (a FileEdit) Copy() Op {
	a2 := FileEdit{File: a.File}
	a2.Edits = make(EditOps, len(a.Edits))
	copy(a2.Edits, a.Edits)
	return a2
}

func (a FileEdit) MarshalJSON() ([]byte, error) {
	type Op FileEdit
	return json.Marshal(&struct {
		Type string `json:"type"`
		Op
	}{
		Type: "edit",
		Op:   (Op)(a),
	})
}

func (a FileEdit) String() string { return fmt.Sprintf("edit(%s:%s)", a.File, a.Edits) }

// A FileCopy is a copy of a source file to a destination file. Prior
// to the operation, the source file must exist and the destination
// file must not exist. After the operation, the source file remains
// unmodified, and the destination file is created with the contents
// of the source file prior to the operation.
type FileCopy struct {
	Src string `json:"src"` // the source file
	Dst string `json:"dst"` // the destination file
}

func (a FileCopy) files() []string { return []string{a.Src, a.Dst} }

// Less implements Op.
func (a FileCopy) Less(other Op) bool {
	if other, ok := other.(FileCopy); ok {
		return a.Dst != other.Dst && (a.Src < other.Src || (a.Src == other.Src && a.Dst < other.Dst))
	}
	return compareToOtherOpFiles(a.Src, other) < 0 && compareToOtherOpFiles(a.Dst, other) != 0
}

// Equal implements Op.
func (a FileCopy) Equal(other Op) bool {
	b, ok := other.(FileCopy)
	return ok && a == b
}

// Copy implements Op.
func (a FileCopy) Copy() Op {
	return a
}

func (a FileCopy) MarshalJSON() ([]byte, error) {
	type Op FileCopy
	return json.Marshal(&struct {
		Type string `json:"type"`
		Op
	}{
		Type: "copy",
		Op:   (Op)(a),
	})
}

func (a FileCopy) String() string { return fmt.Sprintf("copy(%s %s)", a.Src, a.Dst) }

// A FileRename is a rename (or move) of a source file to a
// destination file. Prior to the operation, the source file must
// exist, and the destination file must not exist. After the
// operation, the source file no longer exists, and the destination
// file is created with the contents of the source file prior to the
// operation.
type FileRename struct {
	Src string `json:"src"` // the source file
	Dst string `json:"dst"` // the destination file
}

func (a FileRename) files() []string { return []string{a.Src, a.Dst} }

// Less implements Op.
func (a FileRename) Less(other Op) bool {
	return compareToOtherOpFiles(a.Src, other) < 0 && compareToOtherOpFiles(a.Dst, other) != 0
}

// Equal implements Op.
func (a FileRename) Equal(other Op) bool {
	b, ok := other.(FileRename)
	return ok && a == b
}

// Copy implements Op.
func (a FileRename) Copy() Op {
	return a
}

func (a FileRename) MarshalJSON() ([]byte, error) {
	type Op FileRename
	return json.Marshal(&struct {
		Type string `json:"type"`
		Op
	}{
		Type: "copy",
		Op:   (Op)(a),
	})
}

func (a FileRename) String() string { return fmt.Sprintf("rename(%s %s)", a.Src, a.Dst) }

// A GitHead is an update to the Git repository's HEAD commit (e.g.,
// by an invocation of `git commit` or `git reset`).
type GitHead struct {
	Commit string `json:"commit"` // the Git commit SHA
}

// Less implements Op.
func (a GitHead) Less(other Op) bool {
	// It's equivalent to apply HEAD branch updates before file
	// operations.
	_, ok := other.(GitHead)
	return !ok && other != nil
}

// Equal implements Op.
func (a GitHead) Equal(other Op) bool {
	b, ok := other.(GitHead)
	return ok && a == b
}

// Copy implements Op.
func (a GitHead) Copy() Op { return a }

func (a GitHead) MarshalJSON() ([]byte, error) {
	type Op GitHead
	return json.Marshal(&struct {
		Type string `json:"type"`
		Op
	}{
		Type: "gitHead",
		Op:   (Op)(a),
	})
}

func (a GitHead) String() string { return fmt.Sprintf("head(%s)", abbrevOID(a.Commit)) }

func abbrevOID(oid string) string {
	if len(oid) == 40 {
		return oid[:6]
	}
	return oid
}

package ot

import (
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/go-kit/kit/log"
)

// createWorkspaceModel creates a model of the workspace changes
// represented by the op sequence.
func createWorkspaceModel(logger log.Logger, ops Ops) (*workspaceModel, error) {
	logger.Log("createWorkspaceModel", ops)

	// Copy to avoid modifying the input.
	ops = ops.shallowCopy()

	var w workspaceModel
	for i, op := range ops {
		logger := log.With(logger, "i", i)
		logger.Log("op", op, "model", w)
		switch op := op.(type) {
		case FileCreate:
			old, f := w.newFile(op.File)
			if old != nil && old.state == fileExists {
				return nil, &os.PathError{Op: "create", Path: op.File, Err: os.ErrExist}
			}
			created := old == nil || (!old.deleted && old.state == fileNotExists)
			truncated := old != nil && (old.deleted /* || TODO(sqs) more things here */)
			*f = fileModel{
				name:      op.File,
				state:     fileExists,
				created:   created,
				truncated: truncated,
				edits:     EditOps{{N: 0}}, // OT_TODO move this init down to the FileEdit block?
			}
			if old != nil {
				old.deleted = false
			}

		case FileTruncate:
			f := w.file(op.File)
			if f.state == fileNotExists {
				return nil, &os.PathError{Op: "truncate", Path: op.File, Err: os.ErrNotExist}
			}
			f.state = fileExists
			// If there's an existing edit with a base length of 0,
			// then an explicit truncate is unnecessary (because the
			// document already started out with a length of 0 before
			// the edits); we can just clear out the edits.
			if ret, del, _ := f.edits.Count(); f.edits == nil || ret+del > 0 {
				f.truncated = true
			}
			f.edits = EditOps{{N: 0}}
			if f.copiedFrom != nil {
				f.copiedFrom = nil
				f.created = true
				f.truncated = false
			}

		case FileDelete:
			f := w.file(op.File)
			if f.state == fileNotExists {
				return nil, &os.PathError{Op: "delete", Path: op.File, Err: os.ErrNotExist}
			}

			if !f.hasNonExplicitEdits() {
				for _, df := range w.filesCopiedOrRenamedFrom(f) {
					df.copiedFrom = nil
					df.renamedFrom = nil
					df.created = true
					if df.edits == nil {
						df.edits = f.edits
					}
				}
			}

			if dfs := w.filesCopiedOrRenamedFrom(f); len(dfs) == 1 {
				df := dfs[0]
				logger.Log("single-file-derived-from", f.name, "derived-file", df)
				df.copiedFrom = nil
				df.renamedFrom = f
				f.renamedTo = df
				f.edits = nil
			}

			if f.renamedFrom != nil {
				logger.Log("deleted-rename-dst", f, "cascades-to-delete-rename-src", f.renamedFrom)
				w.delete(f.renamedFrom, !f.renamedFrom.created)
				w.delete(f, f.created)
			} else {
				w.delete(f, !f.created && f.copiedFrom == nil && f.renamedFrom == nil && w.filesRenamedFrom(f) == nil)
			}

		case FileEdit:
			f := w.file(op.File)
			if f.state == fileNotExists {
				return nil, &os.PathError{Op: "edit", Path: op.File, Err: os.ErrNotExist}
			}
			f.state = fileExists

			// Don't propagate these edits to files we already copied
			// from f.
			for _, df := range w.filesCopiedFrom(f) {
				df.edits = f.edits
			}

			var err error
			f.edits, err = ComposeEditOps(f.edits, op.Edits)
			if err != nil {
				return nil, &os.PathError{Op: "edit", Path: op.File, Err: err}
			}

		case FileCopy:
			src := w.file(op.Src)
			if src.state == fileNotExists {
				return nil, &os.PathError{Op: "copy from", Path: op.Src, Err: os.ErrNotExist}
			}
			oldDst, dst := w.newFile(op.Dst)
			if oldDst != nil && oldDst.state == fileExists {
				return nil, &os.PathError{Op: "copy to", Path: op.Dst, Err: os.ErrExist}
			}

			*dst = *src
			dst.name = op.Dst

			src.state = fileExists

			dst.state = fileExists
			dst.created = false
			dst.copiedFrom = src
			dst.renamedFrom = nil

		case FileRename:
			src := w.file(op.Src)
			if src.state == fileNotExists {
				return nil, &os.PathError{Op: "rename from", Path: op.Src, Err: os.ErrNotExist}
			}

			// Simplify [create(/f) rename(/f /g)] to just [create(/g)].
			oldDst, ok := w.files[op.Dst]
			if (!ok || oldDst.state == fileNotExists) && src.edits.Noop() && src.renamedFrom == nil && src.created && len(w.filesCopiedOrRenamedFrom(src)) == 0 {
				src.name = op.Dst
				delete(w.files, op.Src)
				w.files[op.Dst] = src
				continue
			}

			_, dst := w.newFile(op.Dst)
			if oldDst != nil && oldDst.state == fileExists {
				return nil, &os.PathError{Op: "rename to", Path: op.Dst, Err: os.ErrExist}
			}

			*dst = *src

			src.renamedTo = dst
			src.state = fileNotExists
			src.edits = nil

			dst.name = op.Dst
			dst.state = fileExists
			dst.renamedFrom = src
			dst.created = false
			dst.copiedFrom = nil

			if oldDst != nil && (oldDst.renamedTo == src || src.renamedFrom == oldDst) {
				// Eliminate circular rename a->b->c.
				//
				// TODO(renfred): Other files that depend on src.renamedFrom need to be updated.
				oldDst.renamedTo = nil
				src.renamedTo = nil
				src.renamedFrom = nil
				dst.renamedFrom = nil
			}
			if src.renamedFrom != nil {
				// Eliminate "b" in a->b->c rename sequence.
				dst.renamedFrom = src.renamedFrom
				src.renamedFrom = nil
				src.renamedTo = nil
				src.edits = nil
			}
			if dst.renamedFrom != nil && dst.renamedFrom.name == dst.name {
				// Eliminate complex a->b ... b->a rename sequences.
				//
				// TODO(sqs): This is not fully correct; it uses name
				// matching to overcome issues where there is a
				// generation of a file in between oldDst and the true
				// oldDst of the rename.
				src.renamedFrom = nil
				src.renamedTo = nil
				dst.renamedFrom = nil
				dst.renamedTo = nil
			}

		case GitHead:
			w.gitHead = op.Commit
		}
		logger.Log("op", op, "model-ops", opsForWorkspaceModel(w))
	}
	logger.Log("i", "done", "model", w)
	return &w, nil
}

// opsForWorkspaceModel produces an op sequence that represents the
// workspace modeled by w.
func opsForWorkspaceModel(w workspaceModel) Ops {
	ops := make(Ops, 0, len(w.files)) // not an exact allocation

	// Produce ops.
	if w.gitHead != "" {
		ops = append(ops, GitHead{Commit: w.gitHead})
	}
	// When producing the merged ops list, we need to emit ops first
	// for files that other files depend on (i.e., are copied from or
	// renamed from).
	seen := map[*fileModel]struct{}{}
	var emitOpsForFile func(f *fileModel)
	emitOpsForFile = func(f *fileModel) {
		if err := f.checkValid(); err != nil {
			panic(f.name + ": " + err.Error())
		}

		// First process the base file (if any) that this file is
		// derived from.
		if f.copiedFrom != nil {
			emitOpsForFile(f.copiedFrom)
		}

		if _, seen := seen[f]; seen {
			return
		}
		seen[f] = struct{}{}

		switch {
		case f.created:
			ops = append(ops, FileCreate{f.name})
		case f.copiedFrom != nil:
			src := f.copiedFrom
			for src.copiedFrom != nil && src.edits == nil {
				src = src.copiedFrom
			}
			ops = append(ops, FileCopy{src.name, f.name})
		case f.renamedFrom != nil:
			src := f.renamedFrom
			for src.renamedFrom != nil && src.edits == nil {
				src = src.renamedFrom
			}
			ops = append(ops, FileRename{src.name, f.name})
		}

		for _, df := range w.filesCopiedOrRenamedFrom(f) {
			emitOpsForFile(df)
		}

		if f.truncated {
			ops = append(ops, FileTruncate{f.name})
		}
		if f.edits != nil && !f.edits.Noop() {
			ops = append(ops, FileEdit{f.name, f.edits})
		}
		if f.deleted {
			ops = append(ops, FileDelete{f.name})
		}
	}
	for _, f := range w.allFiles {
		if f == nil {
			continue
		}
		emitOpsForFile(f)
	}

	sort.Stable(sortableOps(ops)) // OT_TODO make map iteration stable and remove this sort.
	return ops
}

// opsForWorkspaceModel produces an op sequence b' that can be applied as a transformation
// of workspace bw onto workspace aw.
func transformWorkspaceModel(logger log.Logger, aw workspaceModel, bw workspaceModel, primary bool) (b1 Ops, err error) {
	b1 = Ops{}
	seen := map[*fileModel]struct{}{}
	for _, fb := range bw.allFiles {
		logger.Log("file", fb.name, "ops", b1)
		if fb == nil {
			continue
		}

		if _, seen := seen[fb]; seen {
			continue
		}
		seen[fb] = struct{}{}

		fa, faExists := aw.files[fb.name]
		if fb.created {
			if !faExists {
				b1 = append(b1, FileCreate{fb.name})
			} else {
				if fa.created {
					continue
				}
				if fa.copiedFrom != nil {
					continue
				}
				if fa.state == fileExists {
					return nil, &os.PathError{Op: "create", Path: fb.name, Err: os.ErrExist}
				}
				if fa.deleted {
					return nil, &os.PathError{Op: "create", Path: fb.name, Err: os.ErrExist}
				}
			}
		}
		if fb.copiedFrom != nil {
			src := fb.copiedFrom
			if faExists {
				// Creates are overwritten by copies to the created file. Otherwise modifying
				// a file that hasn't been copied yet is invalid.
				if fa.created {
					b1 = append(b1, FileDelete{fa.name})
				} else if !(fa.copiedFrom != nil && fa.copiedFrom.name == src.name) {
					return nil, &os.PathError{Op: "copy to", Path: fb.name, Err: os.ErrExist}
				}
			}
			for src.copiedFrom != nil && src.edits == nil {
				src = src.copiedFrom
			}
			if !(fa != nil && fa.copiedFrom != nil && fa.copiedFrom.name == src.name) { // Don't add a duplicate copy op
				b1 = append(b1, FileCopy{src.name, fb.name})
			}
		}
		if fb.truncated {
			if faExists && !fa.truncated && !fa.deleted {
				b1 = append(b1, FileTruncate{fb.name})
			}
		}
		if fb.edits != nil && !fb.edits.Noop() {
			// TODO(renfred): refactor using switch-case
			if !faExists {
				b1 = append(b1, FileEdit{fb.name, fb.edits})
			} else if fa.truncated || fa.deleted {
				// Skip
			} else if fb.truncated {
				b1 = append(b1, FileEdit{fb.name, fb.edits})
			} else {
				if fa.edits == nil {
					b1 = append(b1, FileEdit{fb.name, fb.edits})
				} else {
					var be EditOps
					if primary {
						_, be, err = TransformEditOps(fa.edits, fb.edits)
					} else {
						be, _, err = TransformEditOps(fb.edits, fa.edits)
					}
					if err != nil {
						return nil, err
					}
					b1 = append(b1, FileEdit{fb.name, be})
				}
				copiedFrom := aw.filesCopiedFrom(fa)
				for _, f := range copiedFrom {
					b1 = append(b1, FileEdit{f.name, fb.edits})
				}

			}
		}
		if fb.deleted {
			if faExists && !fa.deleted {
				b1 = append(b1, FileDelete{fb.name})
			}
		}
	}

	if bw.gitHead != "" && !primary {
		b1 = append(b1, GitHead{Commit: bw.gitHead})
	}

	logger.Log("final-transform-ops", b1)
	return b1, nil
}

// workspaceModel models a workspace and is used by
// Merge/Compose/Transform to simplify and manipulate op sequences.
type workspaceModel struct {
	files    map[string]*fileModel // current files
	allFiles []*fileModel          // current and historical files, in creation order
	gitHead  string
}

// file returns a fileModel for the given file name. If a fileModel for the given name doesn't
// exist, it will create a new one and return the pointer to it.
func (w *workspaceModel) file(name string) *fileModel {
	if w.files == nil {
		w.files = map[string]*fileModel{}
	}
	f, ok := w.files[name]
	if !ok {
		f = &fileModel{name: name}
		w.files[name] = f
		w.allFiles = append(w.allFiles, f)
	}
	if err := f.checkValid(); err != nil {
		panic(name + ": " + err.Error())
	}
	return f
}

// newFile returns a new fileModel for a given file name, along with the old file with the same
// name if it previously existed.
func (w *workspaceModel) newFile(name string) (old, new *fileModel) {
	old = w.files[name]
	delete(w.files, name)
	new = w.file(name)
	return old, new
}

// prevFile returns the nearest file whose name is f.name that occurs
// before f.
func (w *workspaceModel) prevFile(f *fileModel) *fileModel {
	var prev *fileModel
	for _, f2 := range w.allFiles {
		if f2 == f {
			return prev
		}
		if f2.name == f.name {
			prev = f2
		}
	}
	return nil
}

// nextFile returns the nearest file whose name is f.name that occurs
// after f.
func (w *workspaceModel) nextFile(f *fileModel) *fileModel {
	after := false
	for _, f2 := range w.allFiles {
		if after && f2.name == f.name {
			return f2
		}
		if f2 == f {
			after = true
		}
	}
	return nil
}

// delete marks a given fileModel sets its fields to the appropriate empty values and marks it as
// deleted given the appropriate conditions are met.
func (w *workspaceModel) delete(f *fileModel, deleted bool) {
	var truncated bool
	if nextF := w.nextFile(f); nextF != nil {
		if nextF.state == fileExists && nextF.created && !f.created {
			nextF.created = false
			truncated = true
			deleted = false
		}
	}

	// Avoid multiple redundant truncates.
	for prevF := w.prevFile(f); prevF != nil; prevF = w.prevFile(prevF) {
		if prevF.truncated {
			truncated = false
		}
	}

	f.state = fileNotExists
	f.deleted = deleted
	f.created = false
	f.copiedFrom = nil
	f.renamedFrom = nil
	f.renamedTo = nil
	f.edits = nil
	f.truncated = truncated
}

func (w *workspaceModel) filesCopiedFrom(src *fileModel) []*fileModel {
	var fs []*fileModel
	for _, f := range w.files {
		if f.copiedFrom == src {
			fs = append(fs, f)
		}
	}
	return fs
}

func (w *workspaceModel) filesRenamedFrom(src *fileModel) *fileModel {
	for _, f := range w.files {
		if f.renamedFrom == src {
			return f
		}
	}
	return nil
}

func (w *workspaceModel) filesCopiedOrRenamedFrom(src *fileModel) []*fileModel {
	fs := w.filesCopiedFrom(src)
	if f := w.filesRenamedFrom(src); f != nil {
		fs = append(fs, f)
	}
	return fs
}

func (w *workspaceModel) isCurrentFile(f *fileModel) bool {
	for _, f2 := range w.files {
		if f2 == f {
			return true
		}
	}
	return false
}

func (w workspaceModel) String() string {
	var parts []string
	if w.gitHead != "" {
		parts = append(parts, "head:"+w.gitHead)
	}
	for _, f := range w.allFiles {
		if f == nil {
			continue
		}
		var fileParts []string
		if !w.isCurrentFile(f) {
			fileParts = append(fileParts, "old")
		}
		parts = append(parts, f.string(fileParts))
	}
	return "{" + strings.Join(parts, " ") + "}"
}

// A file can be in one of three states: unknown, existent, and
// nonexistent. If this is the first appearance of the file in an
// op sequence, the file's existence is unknown. Otherwise, its
// existence is inferred from ops (e.g., a "create" means the file
// now exists; an "edit" means the file must exist).
type fileState int

const (
	fileUnknown fileState = iota
	fileExists
	fileNotExists
)

// fileModel models a file in a workspaceModel.
type fileModel struct {
	name string

	state fileState

	created     bool
	deleted     bool
	copiedFrom  *fileModel
	renamedFrom *fileModel
	renamedTo   *fileModel

	truncated bool
	edits     EditOps
}

func (f *fileModel) hasNonExplicitEdits() bool {
	if f.state == fileNotExists {
		panic("notExist")
	}
	return f.edits == nil || f.edits.Retain() > 0
}

func (f *fileModel) checkValid() error {
	if f.copiedFrom == f {
		return errors.New(f.name + ": copiedFrom self")
	}
	if f.renamedFrom == f {
		return errors.New(f.name + ": renamedFrom self")
	}
	if f.renamedTo == f {
		return errors.New(f.name + ": renamedTo self")
	}

	var creationReasons []string
	if f.created {
		creationReasons = append(creationReasons, "created")
	}
	if f.copiedFrom != nil {
		creationReasons = append(creationReasons, "copiedFrom:"+f.copiedFrom.name)
	}
	if f.renamedFrom != nil {
		creationReasons = append(creationReasons, "renamedFrom:"+f.renamedFrom.name)
	}
	if len(creationReasons) > 1 {
		return fmt.Errorf("multiple creation reasons: %v", creationReasons)
	}

	var deletionReasons []string
	if f.deleted {
		deletionReasons = append(deletionReasons, "deleted")
	}
	if f.renamedTo != nil {
		deletionReasons = append(deletionReasons, "renamedTo:"+f.renamedTo.name)
	}
	if len(deletionReasons) > 1 {
		return fmt.Errorf("multiple deletion reasons: %v", deletionReasons)
	}

	return nil
}

func (f *fileModel) String() string {
	return f.string(nil)
}

func (f *fileModel) string(parts []string) string {
	if f.state == fileExists {
		parts = append(parts, "exist")
	}
	if f.state == fileNotExists {
		parts = append(parts, "notExist")
	}
	if f.created {
		parts = append(parts, "created")
	}
	if f.deleted {
		parts = append(parts, "deleted")
	}
	if f.copiedFrom != nil {
		parts = append(parts, "copiedFrom:"+f.copiedFrom.name)
	}
	if f.renamedFrom != nil {
		parts = append(parts, "renamedFrom:"+f.renamedFrom.name)
	}
	if f.renamedTo != nil {
		parts = append(parts, "renamedTo:"+f.renamedTo.name)
	}
	if f.truncated {
		parts = append(parts, "truncated")
	}
	if f.edits != nil {
		parts = append(parts, fmt.Sprintf("edits:%s", f.edits))
	}
	return f.name + "(" + strings.Join(parts, ",") + ")"
}

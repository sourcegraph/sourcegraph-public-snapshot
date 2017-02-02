package ot

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"reflect"
	"sort"
	"strings"
)

// WorkspaceOp represents an operation performed on a workspace.
//
// Each field describes a different type of operation. The struct
// field ordering (first Save, then Copy, and so on) is the order in
// which the types of operations are applied.
type WorkspaceOp struct {
	// Save is a list of paths of files that have been saved. Saving
	// in this context represents transferring the contents of a
	// path's buffer in an editor to its persisted file on-disk. If a
	// file already exists at a given path it will be overwritten.
	// Only buffer paths can be used in a save op.
	Save []string `json:"save,omitempty"`

	// Copy is a map of destination to source file paths of copied
	// files. After the copy, the destination file contains the
	// contents of the source file prior to this op being performed
	// (i.e., without taking edits of the source file into
	// account). The destination file must not exist beforehand.
	//
	// Because a single source file can be copied to multiple
	// destination paths, but not vice versa, the map keys are the
	// destination paths (unlike Rename).
	Copy map[string]string `json:"copy,omitempty"`

	// Rename is a map of source to destination file paths of renamed
	// files. The source file is deleted and the destination file is
	// created implicitly. The destination file must not exist
	// beforehand.  After the rename, the destination file contains
	// the contents of the source file prior to this op being
	// performed (i.e., without taking edits of the source file into
	// account).
	Rename map[string]string `json:"rename,omitempty"`

	// Create is a list of paths of created files, sorted
	// alphabetically.
	Create []string `json:"create,omitempty"`

	// TODO(sqs): add chmod

	// Delete is a list of paths of deleted files, sorted
	// alphabetically.
	Delete []string `json:"delete,omitempty"`

	// Truncate is a list of paths of truncated files (whose contents
	// should be made empty), sorted alphabetically.
	//
	// It's necessary to track these because a consecutive
	// delete-create composes into a truncate operation on the file
	// system. We can't represent it as an edit because we don't
	// necessarily know the file's length.
	Truncate []string `json:"truncate,omitempty"`

	// Edit is a map of file path to edits applied to the file (both
	// to the file saved on disk and the unsaved file in editors). All
	// of the Create, Delete, Truncate, and Rename operations are
	// applied before the edits are considered, so Edit's file paths
	// (map keys) can refer to newly created files, and they should
	// refer to the new (not old) path of renamed files.
	Edit map[string]EditOps `json:"edit,omitempty"`

	// Sel is a map of file path to user IDs to a cursor selection
	// range. The ranges are implicitly modified by other operations
	// (e.g., inserting 3 characters into a file before a selection
	// will increment the selection indexes by 3).
	//
	// A non-nil Sel means the user has the file open. A nil Sel means
	// the user closed the file.
	Sel map[string]map[string]*Sel `json:"sel,omitempty"`

	// GitHead is the ID of the commit that should become the new
	// HEAD. No files are changed when the HEAD is reset (i.e., it is
	// like "git reset --soft").
	//
	// TODO(sqs): make non-git-specifix
	GitHead string `json:"head,omitempty"`
}

// DeepCopy creates a deep copy of op that shares no data with op.
func (op WorkspaceOp) DeepCopy() WorkspaceOp {
	data, err := json.Marshal(op)
	if err != nil {
		panic(err)
	}
	var copy WorkspaceOp
	if err := json.Unmarshal(data, &copy); err != nil {
		panic(err)
	}
	return copy
}

var wsNoop WorkspaceOp

func (op WorkspaceOp) String() string {
	mapSS := func(m map[string]string) string {
		keys := make([]string, 0, len(m))
		for k := range m {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		var buf bytes.Buffer
		for i, k := range keys {
			if i != 0 {
				fmt.Fprint(&buf, " ")
			}
			fmt.Fprint(&buf, k, ":", m[k])
		}
		return buf.String()
	}
	mapBEO := func(m map[string]EditOps) string {
		keys := make([]string, 0, len(m))
		for k := range m {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		var buf bytes.Buffer
		for i, k := range keys {
			if i != 0 {
				fmt.Fprint(&buf, " ")
			}
			fmt.Fprint(&buf, k, ":", m[k])
		}
		return buf.String()
	}
	mapSel := func(m map[string]map[string]*Sel) string {
		mapUserSel := func(m map[string]*Sel) string {
			keys := make([]string, 0, len(m))
			for k := range m {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			var buf bytes.Buffer
			for i, k := range keys {
				if i != 0 {
					fmt.Fprint(&buf, ",")
				}
				fmt.Fprint(&buf, k, "@", m[k])
			}
			return buf.String()
		}

		keys := make([]string, 0, len(m))
		for k := range m {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		var buf bytes.Buffer
		for i, k := range keys {
			if i != 0 {
				fmt.Fprint(&buf, " ")
			}
			fmt.Fprint(&buf, k, ":", mapUserSel(m[k]))
		}
		return buf.String()
	}

	var parts []string
	if len(op.Save) > 0 {
		parts = append(parts, fmt.Sprintf("save(%s)", op.Save))
	}
	if len(op.Copy) > 0 {
		parts = append(parts, fmt.Sprintf("copy(%s)", mapSS(op.Copy)))
	}
	if len(op.Rename) > 0 {
		parts = append(parts, fmt.Sprintf("rename(%s)", mapSS(op.Rename)))
	}
	if len(op.Create) > 0 {
		parts = append(parts, fmt.Sprintf("create(%s)", op.Create))
	}
	if len(op.Delete) > 0 {
		parts = append(parts, fmt.Sprintf("delete(%s)", op.Delete))
	}
	if len(op.Truncate) > 0 {
		parts = append(parts, fmt.Sprintf("truncate(%s)", op.Truncate))
	}
	if len(op.Edit) > 0 {
		parts = append(parts, fmt.Sprintf("edit(%s)", mapBEO(op.Edit)))
	}
	if len(op.Sel) > 0 {
		parts = append(parts, fmt.Sprintf("sel(%s)", mapSel(op.Sel)))
	}
	if op.GitHead != "" {
		parts = append(parts, fmt.Sprintf("head(%s)", op.GitHead))
	}
	return "{" + strings.Join(parts, " ") + "}"
}

// Noop reports whether op is a noop.
func (op WorkspaceOp) Noop() bool {
	return len(op.Save) == 0 && len(op.Create) == 0 && len(op.Delete) == 0 && len(op.Truncate) == 0 && len(op.Rename) == 0 && len(op.Copy) == 0 && len(op.Edit) == 0 && len(op.Sel) == 0 && op.GitHead == ""
}

// NormalizeWorkspaceOp normalizes op to its canonical form, with
// lists sorted and edit operations merged. It does not modify op or
// any of its members.
//
// TODO: To improve perf, we could avoid allocations unless necessary,
// or make it so that callers cheaply maintain the canonical-ness of
// an op when they modify it.
func NormalizeWorkspaceOp(op WorkspaceOp) WorkspaceOp {
	normalizeUniqSlice := func(a []string) []string {
		if len(a) == 0 {
			return nil
		}
		m := make(map[string]struct{}, len(a))
		for _, e := range a {
			m[e] = struct{}{}
		}
		b := make([]string, 0, len(m))
		for e := range m {
			b = append(b, e)
		}
		sort.Strings(b)
		return b
	}

	op.Save = normalizeUniqSlice(op.Save)
	op.Create = normalizeUniqSlice(op.Create)
	op.Delete = normalizeUniqSlice(op.Delete)
	op.Truncate = normalizeUniqSlice(op.Truncate)

	if len(op.Rename) == 0 {
		op.Rename = nil
	}
	if len(op.Copy) == 0 {
		op.Copy = nil
	}

	if len(op.Edit) > 0 {
		mergedEdits := make(map[string]EditOps, len(op.Edit))
		for path, eo := range op.Edit {
			mo := MergeEditOps(eo)
			if !mo.Noop() {
				mergedEdits[path] = mo
			}
		}
		op.Edit = mergedEdits
	} else {
		op.Edit = nil
	}

	if len(op.Sel) == 0 {
		op.Sel = nil
	}

	return op
}

// CheckWorkspaceOp checks the validity of op. If it is invalid, a descriptive error is returned.
func CheckWorkspaceOp(op WorkspaceOp) error {
	for _, f := range op.Save {
		if !isValidBufferPath(f) {
			return &os.PathError{Op: "save", Path: f, Err: errors.New("attempted to save from invalid buffer path")}
		}
	}
	intersect := func(a, b []string) (intersection []string) {
		am := make(map[string]struct{}, len(a))
		for _, e := range a {
			am[e] = struct{}{}
		}
		for _, e := range b {
			if _, present := am[e]; present {
				intersection = append(intersection, e)
			}
		}
		return
	}
	mapKeys := func(m interface{}) (keys []string) {
		for _, ev := range reflect.ValueOf(m).MapKeys() {
			if ev.Kind() != reflect.String {
				panic("map key is not string:" + ev.Kind().String())
			}
			keys = append(keys, ev.String())
		}
		return
	}
	mapValues := func(m interface{}) (vals []string) {
		mv := reflect.ValueOf(m)
		for _, kv := range mv.MapKeys() {
			ev := mv.MapIndex(kv)
			if ev.Kind() != reflect.String {
				panic("map value is not string:" + ev.Kind().String())
			}
			vals = append(vals, ev.String())
		}
		return
	}

	if x := intersect(op.Create, op.Delete); len(x) > 0 {
		return fmt.Errorf("WorkspaceOp.{Create,Delete} overlap: %q", x)
	}
	if x := intersect(op.Create, op.Truncate); len(x) > 0 {
		return fmt.Errorf("WorkspaceOp.{Create,Truncate} overlap: %q", x)
	}
	if x := intersect(op.Delete, op.Truncate); len(x) > 0 {
		return fmt.Errorf("WorkspaceOp.{Delete,Truncate} overlap: %q", x)
	}
	if x := intersect(op.Delete, mapKeys(op.Rename)); len(x) > 0 {
		return fmt.Errorf("WorkspaceOp.{Delete,Rename source path} overlap: %q (renamed source files are implicitly deleted)", x)
	}
	if x := intersect(op.Create, mapValues(op.Rename)); len(x) > 0 {
		return fmt.Errorf("WorkspaceOp.{Create,Rename destination path} overlap: %q (renamed destination files are implicitly created)", x)
	}
	if x := intersect(op.Create, mapKeys(op.Copy)); len(x) > 0 {
		return fmt.Errorf("WorkspaceOp.{Create,Copy destination path} overlap: %q (copy destination files are implicitly created)", x)
	}
	if x := intersect(op.Delete, mapKeys(op.Copy)); len(x) > 0 {
		return fmt.Errorf("WorkspaceOp.{Delete,Copy destination path} overlap: %q", x)
	}
	if x := intersect(op.Create, mapValues(op.Copy)); len(x) > 0 {
		return fmt.Errorf("WorkspaceOp.{Create,Copy source path} overlap: %q", x)
	}
	if x := intersect(op.Delete, mapKeys(op.Edit)); len(x) > 0 {
		return fmt.Errorf("WorkspaceOp.{Delete,Edit} overlap: %q", x)
	}
	if x := intersect(mapKeys(op.Rename), mapKeys(op.Copy)); len(x) > 0 {
		return fmt.Errorf("WorkspaceOp.{Rename source path,Copy destination path} overlap: %q", x)
	}

	// Intersections in the below pairs are OK and do not indicate
	// that the op is invalid. We include the lines below even though
	// they are no-ops to be explicit.
	intersect(op.Delete, mapValues(op.Copy))
	intersect(op.Truncate, mapKeys(op.Rename))
	intersect(mapKeys(op.Rename), mapValues(op.Copy))
	intersect(mapKeys(op.Copy), mapValues(op.Copy))
	intersect(op.Save, mapValues(op.Copy))
	intersect(op.Save, mapValues(op.Rename))
	intersect(op.Save, op.Delete)
	intersect(op.Save, op.Truncate)

	return nil
}

// tmpWorkspaceOp indexes a WorkspaceOp and allows for certain slice
// fields to be accessed and set via maps (which make it easier to
// test for membership).
type tmpWorkspaceOp struct {
	WorkspaceOp
	copyFrom                       map[string][]string
	renameTo                       map[string]string
	Create, Delete, Truncate, Save map[string]struct{} // shadow with more useful set, not string slice
}

// existed is whether a file at the given path existed before
// performing the op, based on the implied conditions (e.g., if a file
// was created, then it must not have existed previously).
func (op *tmpWorkspaceOp) existed(path string) (existed, known bool) {
	_, copiedFrom := op.copyFrom[path]
	_, copiedTo := op.Copy[path]
	_, renamedFrom := op.Rename[path]
	_, renamedTo := op.renameTo[path]
	_, created := op.Create[path]
	_, deleted := op.Delete[path]
	_, truncated := op.Truncate[path]
	_, edited := op.Edit[path]
	_, selected := op.Sel[path]

	existed = copiedFrom || renamedFrom || deleted || truncated || edited || selected && !(renamedTo || copiedTo || created)
	known = copiedFrom || copiedTo || renamedFrom || renamedTo || created || deleted || truncated || edited || selected
	return
}

// exists is whether a file exists after performing the op (e.g., if a
// file was created, then it exists now).
func (op *tmpWorkspaceOp) exists(path string) (exists, known bool) {
	_, copiedFrom := op.copyFrom[path]
	_, copiedTo := op.Copy[path]
	_, renamedFrom := op.Rename[path]
	_, renamedTo := op.renameTo[path]
	_, created := op.Create[path]
	_, deleted := op.Delete[path]
	_, truncated := op.Truncate[path]
	_, edited := op.Edit[path]
	_, selected := op.Sel[path]
	var savedTo bool
	if isValidFilePath(path) {
		b, _ := fileToBufferPath(path)
		_, savedTo = op.Save[b]
	}

	exists = copiedFrom || copiedTo || renamedTo || created || truncated || edited || selected || savedTo && !(renamedFrom || deleted)
	known = copiedFrom || copiedTo || renamedFrom || renamedTo || created || deleted || truncated || savedTo || edited || selected
	return
}

// from initializes and indexes a temp. data structure from a.
//
// It should be called exactly once on op.
func (op *tmpWorkspaceOp) from(a WorkspaceOp) (err error) {
	op.Save = make(map[string]struct{}, len(a.Save))
	for _, f := range a.Save {
		op.Save[f] = struct{}{}
	}

	op.Copy = make(map[string]string, len(a.Copy))
	op.copyFrom = make(map[string][]string, len(a.Copy))
	for d, s := range a.Copy {
		// No need to check for existence of s, since no operations
		// could have deleted any files yet.
		if exists, known := op.exists(d); known && exists {
			err = &os.PathError{Op: "copy to", Path: d, Err: os.ErrExist}
			return
		}
		op.Copy[d] = s
		op.copyFrom[s] = append(op.copyFrom[s], d)
	}

	op.Rename = make(map[string]string, len(a.Rename))
	op.renameTo = make(map[string]string, len(a.Copy))
	for s, d := range a.Rename {
		// No need to check for existence of s, since no operations
		// could have deleted any files yet.
		if exists, known := op.exists(d); known && exists {
			err = &os.PathError{Op: "rename to", Path: d, Err: os.ErrExist}
			return
		}
		op.Rename[s] = d
		op.renameTo[d] = s
	}

	op.Create = make(map[string]struct{}, len(a.Create))
	for _, f := range a.Create {
		if exists, known := op.exists(f); known && exists {
			err = &os.PathError{Op: "create", Path: f, Err: os.ErrExist}
			return
		}
		op.Create[f] = struct{}{}
	}

	op.Delete = make(map[string]struct{}, len(a.Delete))
	for _, f := range a.Delete {
		if exists, known := op.exists(f); known && !exists {
			err = &os.PathError{Op: "delete", Path: f, Err: os.ErrNotExist}
			return
		}
		op.Delete[f] = struct{}{}
	}

	op.Truncate = make(map[string]struct{}, len(a.Truncate))
	for _, f := range a.Truncate {
		if exists, known := op.exists(f); known && !exists {
			err = &os.PathError{Op: "truncate", Path: f, Err: os.ErrNotExist}
			return
		}
		op.Truncate[f] = struct{}{}
	}

	op.Edit = make(map[string]EditOps, len(a.Edit))
	for f, edit := range a.Edit {
		if exists, known := op.exists(f); known && !exists {
			err = &os.PathError{Op: "edit", Path: f, Err: os.ErrNotExist}
			return
		}
		op.Edit[f] = edit
	}

	op.Sel = make(map[string]map[string]*Sel, len(a.Sel))
	for f, sel := range a.Sel {
		if exists, known := op.exists(f); known && !exists {
			err = &os.PathError{Op: "sel", Path: f, Err: os.ErrNotExist}
			return
		}
		op.Sel[f] = cloneSels(sel)
	}

	op.GitHead = a.GitHead
	return
}

func (op *tmpWorkspaceOp) toWorkspaceOp() WorkspaceOp {
	mapKeys := func(m map[string]struct{}) []string {
		keys := make([]string, 0, len(m))
		for k := range m {
			keys = append(keys, k)
		}
		return keys
	}

	op.WorkspaceOp.Save = mapKeys(op.Save)
	op.WorkspaceOp.Create = mapKeys(op.Create)
	op.WorkspaceOp.Delete = mapKeys(op.Delete)
	op.WorkspaceOp.Truncate = mapKeys(op.Truncate)

	return op.WorkspaceOp
}

func isValidBufferPath(p string) bool {
	return strings.HasPrefix(p, "#")
}

func isValidFilePath(p string) bool {
	return strings.HasPrefix(p, "/")
}

func fileToBufferPath(p string) (string, error) {
	if !isValidFilePath(p) {
		panic("invalid file path: " + p)
		return "", fmt.Errorf("invalid file path %s", p)
	}
	return "#" + p[1:], nil
}

func bufferToFilePath(p string) (string, error) {
	if !isValidBufferPath(p) {
		return "", fmt.Errorf("invalid buffer path %s", p)
	}
	return "/" + p[1:], nil
}

func panicIfFileOrBufferPath(path string) {
	if strings.HasPrefix(path, "#") || strings.HasPrefix(path, "/") {
		panic(fmt.Sprintf("unexpected file or buffer path %q", path))
	}
}

func panicIfNotFileOrBufferPath(path string) {
	if !strings.HasPrefix(path, "#") && !strings.HasPrefix(path, "/") {
		panic(fmt.Sprintf("path %q is not a buffer or file path", path))
	}
}

func isBufferPath(path string) bool {
	panicIfNotFileOrBufferPath(path)
	return strings.HasPrefix(path, "#")
}

func isFilePath(path string) bool {
	panicIfNotFileOrBufferPath(path)
	return strings.HasPrefix(path, "/")
}

func stripBufferPath(path string) string {
	if !strings.HasPrefix(path, "#") {
		panic(fmt.Sprintf("expected path %q to have '#' prefix", path))
	}
	return strings.TrimPrefix(path, "#")
}

func stripFilePath(path string) string {
	if !strings.HasPrefix(path, "/") {
		panic(fmt.Sprintf("expected path %q to have '/' prefix", path))
	}
	return strings.TrimPrefix(path, "/")
}

func stripFileOrBufferPath(path string) string {
	panicIfNotFileOrBufferPath(path)
	return path[1:]
}

// ComposeWorkspaceOps returns an operation equivalent to a followed
// by b. The operations a and b must be consecutive operations.
func ComposeWorkspaceOps(a, b WorkspaceOp) (ab WorkspaceOp, err error) {
	a = NormalizeWorkspaceOp(a)
	b = NormalizeWorkspaceOp(b)

	var op tmpWorkspaceOp
	if err = op.from(a); err != nil {
		return
	}

	for d, s := range b.Copy {
		if exists, known := op.exists(s); known && !exists {
			err = &os.PathError{Op: "copy from", Path: s, Err: os.ErrNotExist}
			return
		}
		if exists, known := op.exists(d); known && exists {
			err = &os.PathError{Op: "copy to", Path: d, Err: os.ErrExist}
			return
		}
		if _, edited := op.Edit[s]; edited {
			op.Edit[d] = op.Edit[s]
		}
		if s2, chained := op.Copy[s]; chained {
			s = s2
		}
		if s2, chained := op.renameTo[s]; chained {
			s = s2
		}
		_, created := op.Create[s]
		if created {
			op.Create[d] = struct{}{}
		}
		if _, deleted := op.Delete[d]; deleted {
			delete(op.Delete, d)
		}
		if _, truncated := op.Truncate[s]; truncated {
			op.Truncate[d] = struct{}{}
		}

		if !created {
			op.Copy[d] = s
			op.copyFrom[s] = append(op.copyFrom[s], d)
		}

		if isValidFilePath(s) && isValidFilePath(d) {
			var sb, db string
			sb, err = fileToBufferPath(s)
			if err != nil {
				return
			}
			db, err = fileToBufferPath(d)
			if err != nil {
				return
			}
			if _, saved := op.Save[sb]; saved {
				op.Save[db] = struct{}{}
				op.Copy[db] = sb
				delete(op.Copy, d)
			}
		}
	}

	for s, d := range b.Rename {
		if exists, known := op.exists(s); known && !exists {
			err = &os.PathError{Op: "rename from", Path: s, Err: os.ErrNotExist}
			return
		}
		if exists, known := op.exists(d); known && exists {
			err = &os.PathError{Op: "rename to", Path: d, Err: os.ErrExist}
			return
		}

		// Eliminate noop links in the chain.
		o2 := s
		if s2, renamedTo := op.renameTo[s]; renamedTo {
			s = s2
		}

		if copySrc, copied := op.Copy[s]; copied {
			delete(op.Copy, s)
			op.Copy[d] = copySrc
			continue
		}
		_, created := op.Create[s]
		if created {
			delete(op.Create, s)
			op.Create[d] = struct{}{}
		}
		if _, deleted := op.Delete[d]; deleted {
			delete(op.Delete, d)
		}
		if _, truncated := op.Truncate[s]; truncated {
			delete(op.Truncate, s)
			op.Truncate[d] = struct{}{}
		}

		if _, edited := op.Edit[o2]; edited {
			op.Edit[d] = op.Edit[o2]
			delete(op.Edit, o2)
			delete(op.Edit, s)
		}
		if edits, edited := b.Edit[b.Rename[o2]]; edited {
			op.Edit[d], err = ComposeEditOps(op.Edit[d], edits)
			if err != nil {
				return
			}
			delete(b.Edit, b.Rename[o2])
		}

		if _, selected := op.Sel[o2]; selected {
			op.Sel[d] = op.Sel[o2]
			delete(op.Sel, o2)
			delete(op.Sel, s)
		}
		if sel, selected := b.Sel[b.Rename[o2]]; selected {
			op.Sel[d] = mergeSels(op.Sel[d], sel)
			delete(b.Sel, b.Rename[o2])
		}

		if !created {
			op.Rename[s] = d
			op.renameTo[d] = s
		}

		if isValidFilePath(s) {
			var sb, db string
			sb, err = fileToBufferPath(s)
			if err != nil {
				return
			}
			db, err = fileToBufferPath(d)
			if err != nil {
				return
			}
			if _, saved := op.Save[sb]; saved {
				op.Save[db] = struct{}{}
				op.Copy[db] = sb
				op.Delete[s] = struct{}{}
				delete(op.Rename, s)
			}
		}
	}

	for _, s := range b.Save {
		if exists, known := op.exists(s); known && !exists {
			err = &os.PathError{Op: "save from", Path: s, Err: os.ErrNotExist}
			return
		}
		var d string
		d, err = bufferToFilePath(s)
		if err != nil {
			return
		}
		if _, edited := op.Edit[d]; edited {
			delete(op.Edit, d)
		}
		if _, edited := op.Edit[s]; edited {
			op.Edit[d] = op.Edit[s]
		}
		if _, deleted := op.Delete[d]; deleted {
			delete(op.Delete, d)
		}
		if _, truncated := op.Truncate[d]; truncated {
			delete(op.Truncate, d)
		}
		_, truncated := op.Truncate[s]
		if truncated {
			op.Truncate[d] = struct{}{}
		}
		if r, renamed := op.renameTo[d]; renamed {
			delete(op.Rename, r)
			op.Delete[r] = struct{}{}
		}

		if cs, ok := op.Copy[s]; ok {
			if isFilePath(cs) && isBufferPath(s) && stripFileOrBufferPath(cs) == stripFileOrBufferPath(s) {
				delete(op.Copy, s)
				delete(op.Save, s)
				if edit, ok := op.Edit[s]; ok {
					delete(op.Edit, s)
					op.Edit[d] = edit
				}
				continue
			}
		}

		_, created := op.Create[d]
		if !created && !truncated {
			op.Save[s] = struct{}{}
		}
	}

	bDelete := map[string]struct{}{}
	for _, f := range b.Delete {
		bDelete[f] = struct{}{}
	}

	noDelete := map[string]struct{}{}
	for _, f := range b.Create {
		if exists, known := op.exists(f); known && exists {
			err = &os.PathError{Op: "create", Path: f, Err: os.ErrExist}
			return
		}
		if _, deleted := op.Delete[f]; deleted {
			delete(op.Delete, f)
			op.Truncate[f] = struct{}{}
			continue
		}
		o, renamed := op.Rename[f]
		chained := renamed
		if !chained {
			o, chained = op.Copy[f]
		}
		if chained {
			if _, deleted := op.Delete[o]; deleted {
				continue
			}
			if renamed {
				if _, dstDeleted := bDelete[o]; dstDeleted {
					op.Truncate[f] = struct{}{}
				} else {
					op.Create[f] = struct{}{}
				}
				noDelete[f] = struct{}{}
				continue
			}
			f = o
		}
		op.Create[f] = struct{}{}
	}

	for _, f := range b.Delete {
		if exists, known := op.exists(f); known && !exists {
			err = &os.PathError{Op: "delete", Path: f, Err: os.ErrNotExist}
			return
		}
		if o, chained := op.renameTo[f]; chained {
			delete(op.Rename, o)
			if _, created := op.Create[f]; created {
				delete(op.Create, f)
				op.Truncate[o] = struct{}{}
			} else if _, ok := noDelete[o]; !ok {
				op.Delete[o] = struct{}{}
			}
		} else if _, created := op.Create[f]; created {
			delete(op.Create, f)
		} else if _, copied := op.Copy[f]; copied {
			delete(op.Copy, f)
		} else {
			op.Delete[f] = struct{}{}
		}
		if o, chained := op.Rename[f]; chained {
			f = o
		}
		delete(op.Edit, f)
		delete(op.Sel, f)
		if isValidFilePath(f) {
			sb, _ := fileToBufferPath(f)
			if _, saved := op.Save[sb]; saved {
				delete(op.Save, sb)
			}
		}
	}

	for _, f := range b.Truncate {
		if exists, known := op.exists(f); known && !exists {
			err = &os.PathError{Op: "truncate", Path: f, Err: os.ErrNotExist}
			return
		}
		if o, chained := op.Rename[f]; chained {
			f = o
		}

		op.Truncate[f] = struct{}{}
		delete(op.Edit, f)
		delete(op.Sel, f)
		if isValidFilePath(f) {
			sb, _ := fileToBufferPath(f)
			if _, saved := op.Save[sb]; saved {
				delete(op.Save, sb)
			}
		}
	}

	for f, edit := range b.Edit {
		if exists, known := op.exists(f); known && !exists {
			err = &os.PathError{Op: "edit", Path: f, Err: os.ErrNotExist}
			return
		}
		op.Edit[f], err = ComposeEditOps(op.Edit[f], edit)
		if err != nil {
			return
		}
		if _, created := op.Create[f]; created {
			ret, _, _ := op.Edit[f].Count()
			if ret != 0 {
				err = &os.PathError{Op: "edit", Path: f, Err: fmt.Errorf("newly created file has nonzero retain count %d", ret)}
				return
			}
		}
		for u, sel := range op.Sel[f] {
			op.Sel[f][u] = AdjustSel(sel, op.Edit[f])
		}
	}

	for f, sel := range b.Sel {
		op.Sel[f] = mergeSels(op.Sel[f], sel)
	}

	// Simplify.
	for d, s := range op.Copy {
		if d == s {
			delete(op.Copy, d)
		}
	}
	for s, d := range op.Rename {
		if s == d {
			delete(op.Rename, s)
		}
	}
	for f := range op.Truncate {
		if _, created := op.Create[f]; created {
			delete(op.Truncate, f)
		}
	}
	for s := range op.Save {
		d, _ := bufferToFilePath(s)
		if _, deleted := op.Delete[d]; deleted {
			delete(op.Save, s)
		}
	}

	if b.GitHead != "" {
		op.GitHead = b.GitHead
	}

	return NormalizeWorkspaceOp(op.toWorkspaceOp()), nil
}

// TransformWorkspaceOps returns two operations derived from the
// concurrent ops a and b. An error is returned if the transformation
// failed.
//
// To resolve conflicts in a consistent manner, a should be the
// server's op. In case of a resolvable conflict, a wins. TODO(sqs):
// we currently fail in a (WorkspaceOp).Head conflict but we resolve
// other ops in favor of a as described...investigate and fix that
// inconsistency.
func TransformWorkspaceOps(a, b WorkspaceOp) (a1, b1 WorkspaceOp, err error) {
	var at, bt tmpWorkspaceOp
	if err = at.from(a); err != nil {
		return
	}
	if err = bt.from(b); err != nil {
		return
	}

	transform := func(x, y tmpWorkspaceOp, z *WorkspaceOp, primary bool) (err error) {
		transformEditOps := func(x, y EditOps) (EditOps, error) {
			var edit1, edit2 EditOps
			if primary {
				edit1, edit2 = x, y
			} else {
				edit1, edit2 = y, x
			}
			x1, y1, err := TransformEditOps(edit1, edit2)
			if err != nil {
				return EditOps{}, err
			}
			if primary {
				return x1, nil
			}
			return y1, nil
		}

		newFileSrc := map[string]string{}
		for d, s := range x.Copy {
			if ys, ycopied := y.Copy[d]; ycopied && ys == s {
				continue
			} else if ycopied && ys != s {
				err = &os.PathError{Op: "copy to", Path: s, Err: fmt.Errorf("conflict: %s", ys)}
				return
			}
			if existed, known := y.existed(s); known && !existed {
				err = &os.PathError{Op: "copy from", Path: s, Err: os.ErrNotExist}
				return
			}
			if yd, yrenamed := y.Rename[s]; yrenamed && yd == d {
				continue
			} else if exists, known := y.exists(d); known && exists {
				err = &os.PathError{Op: "copy to", Path: d, Err: os.ErrExist}
				return
			}
			if z.Copy == nil {
				z.Copy = make(map[string]string, len(x.Copy))
			}
			z.Copy[d] = s
			newFileSrc[d] = s
		}

		for s, d := range x.Rename {
			if yd, yrenamed := y.Rename[d]; yrenamed && yd == s {
				continue
			} else if yrenamed && yd != s {
				err = &os.PathError{Op: "rename to", Path: s, Err: fmt.Errorf("conflict: %s", yd)}
				return
			}
			if existed, known := y.existed(s); known && !existed {
				err = &os.PathError{Op: "rename from", Path: s, Err: os.ErrNotExist}
				return
			}
			if yd, yrename := y.Rename[s]; yrename && yd == d {
				continue
			} else if yrename {
				if z.Copy == nil {
					z.Copy = make(map[string]string, len(x.Copy))
				}
				z.Copy[d] = yd
				newFileSrc[yd] = s
				continue
			}
			if exists, known := y.exists(s); known && !exists {
				err = &os.PathError{Op: "rename from", Path: s, Err: os.ErrNotExist}
				return
			}
			if ys, ycopied := y.Copy[d]; ycopied && ys == s {
				z.Delete = append(z.Delete, s)
				continue
			} else if existed, known := y.existed(d); known && existed {
				err = &os.PathError{Op: "rename to", Path: d, Err: os.ErrExist}
				return
			} else if exists, known := y.exists(d); known && exists {
				err = &os.PathError{Op: "rename to", Path: d, Err: os.ErrExist}
				return
			}
			if z.Rename == nil {
				z.Rename = make(map[string]string, len(x.Rename))
			}
			z.Rename[s] = d
			newFileSrc[d] = s
		}

		for s, _ := range x.Save {
			if _, ysaved := y.Save[s]; ysaved {
				continue
			}

			z.Save = append(z.Save, s)
		}

		for f := range x.Create {
			if _, ycreated := y.Create[f]; !ycreated {
				if exists, known := y.exists(f); known && exists {
					err = &os.PathError{Op: "create", Path: f, Err: os.ErrExist}
					return
				}
				z.Create = append(z.Create, f)
			}
			if _, ydeleted := y.Delete[f]; ydeleted {
				return &os.PathError{Op: "create", Path: f, Err: errors.New("conflict: y delete")}
			}
		}
		for f := range x.Delete {
			if _, ydeleted := y.Delete[f]; !ydeleted {
				if exists, known := y.exists(f); known && !exists {
					err = &os.PathError{Op: "delete", Path: f, Err: os.ErrNotExist}
					return
				}
				z.Delete = append(z.Delete, f)
			}
			if _, ycreated := y.Create[f]; ycreated {
				return &os.PathError{Op: "delete", Path: f, Err: errors.New("conflict: y create")}
			}
		}
		for f := range x.Truncate {
			_, ycreated := y.Create[f]
			_, ydeleted := y.Delete[f]
			_, ytruncated := y.Truncate[f]
			sameTruncateEdit := ytruncated && y.Edit[f].Equal(x.Edit[f])
			if len(y.Edit[f]) > 0 && !sameTruncateEdit {
				return &os.PathError{Op: "truncate", Path: f, Err: errors.New("conflict: edit and y truncate")}
			}
			if ydeleted {
				return &os.PathError{Op: "truncate", Path: f, Err: errors.New("conflict: y delete")}
			}
			if !ytruncated && !ycreated && !ydeleted && !sameTruncateEdit {
				if exists, known := y.exists(f); known && !exists {
					err = &os.PathError{Op: "truncate", Path: f, Err: os.ErrNotExist}
					return
				}
				z.Truncate = append(z.Truncate, f)
			}
		}

		for f, edit := range x.Edit {
			yd, yrenamed := y.Rename[f]
			if exists, known := y.exists(yd); yrenamed && known && !exists {
				err = &os.PathError{Op: "edit", Path: yd, Err: os.ErrNotExist}
				return
			} else if exists, known := y.exists(f); !yrenamed && known && !exists {
				err = &os.PathError{Op: "edit", Path: f, Err: os.ErrNotExist}
				return
			}

			if z.Edit == nil {
				z.Edit = make(map[string]EditOps, len(x.Edit))
			}

			for _, d := range y.copyFrom[f] {
				z.Edit[d], err = transformEditOps(x.Edit[f], y.Edit[d])
				if err != nil {
					return
				}
			}

			xedit := x.Edit[f]

			var yf string
			if s, isNewFile := newFileSrc[f]; isNewFile {
				yf = s
			} else {
				yf = f
			}
			if d, yrenamed := y.Rename[yf]; yrenamed {
				f = d
				edit, err = transformEditOps(edit, y.Edit[d])
				if err != nil {
					return
				}
			}
			yedit, yedited := y.Edit[yf]
			if yedited {
				edit, err = transformEditOps(edit, yedit)
				if err != nil {
					return
				}
			}

			if !xedit.Equal(y.Edit[f]) {
				if !z.Edit[f].Noop() {
					err = &os.PathError{Op: "edit", Path: f, Err: errors.New("copy/rename edit conflict")}
					return
				}
				z.Edit[f] = edit
			}
		}

		for f, _ := range x.Edit {
			if isValidBufferPath(f) {
				d, _ := bufferToFilePath(f)
				if _, saved := y.Save[f]; saved {
					z.Edit[d], err = transformEditOps(x.Edit[f], y.Edit[d])
					if err != nil {
						return
					}
				}
				if _, saved := x.Save[f]; saved {
					z.Edit[d] = y.Edit[d]
				}
			}
		}
		for f, _ := range x.Edit {
			if isValidFilePath(f) {
				s, _ := fileToBufferPath(f)
				if _, saved := x.Save[s]; saved {
					z.Edit[f], err = transformEditOps(x.Edit[f], y.Edit[s])
					if err != nil {
						return
					}
					var z1 EditOps
					z1, err = transformEditOps(z.Edit[f], y.Edit[f])
					if err != nil {
						return
					}
					z.Edit[f], err = ComposeEditOps(y.Edit[f], z1)
					if err != nil {
						return
					}
				}
			}
		}

		for f, usel := range x.Sel {
			if _, deleted := y.Delete[f]; deleted {
				continue
			}
			if _, truncated := y.Truncate[f]; truncated {
				continue
			}
			f2, ok := y.Rename[f]
			if ok {
				f = f2
			}
			for u, sel := range usel {
				_, yselected := y.Sel[f][u]
				if primary || !yselected {
					if z.Sel == nil {
						z.Sel = make(map[string]map[string]*Sel, 1)
					}
					if z.Sel[f] == nil {
						z.Sel[f] = make(map[string]*Sel, 1)
					}
					z.Sel[f][u] = AdjustSel(sel, y.Edit[f])
				}
			}
		}

		return
	}

	if err = transform(at, bt, &a1, true); err != nil {
		return
	}
	if err = transform(bt, at, &b1, false); err != nil {
		return
	}

	if a.GitHead != b.GitHead {
		if a.GitHead != "" && b.GitHead != "" {
			err = fmt.Errorf("transform encountered conflicting heads: %q != %q", a.GitHead, b.GitHead)
			return
		}
		a1.GitHead = a.GitHead
		b1.GitHead = b.GitHead
	}

	a1 = NormalizeWorkspaceOp(a1)
	b1 = NormalizeWorkspaceOp(b1)
	return
}

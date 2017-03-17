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
// field ordering (first Save, then Copy, then Rename, and so on) is
// the order in which the types of operations are applied.
type WorkspaceOp struct {
	// Save is a list of paths of files that have been saved. Saving
	// in this context represents transferring the contents of a
	// path's buffer in an editor to its persisted file on-disk. If a
	// file already exists at a given path it will be
	// overwritten. After saving, the buffer file is considered to no
	// longer exist. Only buffer paths can be used in a save op.
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
	//
	// NOTES
	//
	// Copy is equivalent to a rename-create-edit, but we model it
	// separately for better UX. Consider the following case: Alice
	// and Bob are both editing a file F. Alice copies the file right
	// after Bob makes some edits locally but before Bob's edits are
	// synced to the server. If Alice performs a rename-create-edit
	// (where the edit restores the original file contents to the
	// source file), then Bob's edits would flow into the new file but
	// not the existing file. This is bad because the behavior we want
	// is for edits to flow into both the old and new names of the
	// file.
	//
	// This might seem like an edge case (who really copies files?),
	// but a copy from a file system path to a special buffer path is
	// how we've modeled unsaved files in your editor. So, the
	// scenario described above is extremely common and modeling it as
	// a rename-create-edit would cause a lot of accidentally
	// clobbered edits when someone saves a file while someone else is
	// editing it.
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
		if len(mergedEdits) > 0 {
			op.Edit = mergedEdits
		} else {
			op.Edit = nil
		}
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
	if _, savedFrom := op.Save[path]; savedFrom {
		return true, true
	}
	if isValidFilePath(path) {
		b, _ := fileToBufferPath(path)
		if _, savedTo := op.Save[b]; savedTo {
			return false, false // unknown
		}
	}
	if _, copiedFrom := op.copyFrom[path]; copiedFrom {
		return true, true
	}
	if _, copiedTo := op.Copy[path]; copiedTo {
		return false, true
	}
	if _, renamedFrom := op.Rename[path]; renamedFrom {
		return true, true
	}
	if _, renamedTo := op.renameTo[path]; renamedTo {
		return false, true
	}
	if _, created := op.Create[path]; created {
		return false, true
	}
	if _, deleted := op.Delete[path]; deleted {
		return true, true
	}
	if _, truncated := op.Truncate[path]; truncated {
		return true, true
	}
	if _, edited := op.Edit[path]; edited {
		return true, true
	}
	if _, selected := op.Sel[path]; selected {
		return true, true
	}
	return false, false
}

// exists is whether a file exists after performing the op (e.g., if a
// file was created, then it exists now).
func (op *tmpWorkspaceOp) exists(path string) (exists, known bool) {
	if _, selected := op.Sel[path]; selected {
		return true, true
	}
	if _, edited := op.Edit[path]; edited {
		return true, true
	}
	if _, truncated := op.Truncate[path]; truncated {
		return true, true
	}
	if _, deleted := op.Delete[path]; deleted {
		return false, true
	}
	if _, created := op.Create[path]; created {
		return true, true
	}
	if _, renamedTo := op.renameTo[path]; renamedTo {
		return true, true
	}
	if _, renamedFrom := op.Rename[path]; renamedFrom {
		return false, true
	}
	if _, copiedTo := op.Copy[path]; copiedTo {
		return true, true
	}
	if _, copiedFrom := op.copyFrom[path]; copiedFrom {
		return true, true
	}
	if _, savedFrom := op.Save[path]; savedFrom {
		return false, true
	}
	if isValidFilePath(path) {
		b, _ := fileToBufferPath(path)
		if _, savedTo := op.Save[b]; savedTo {
			return true, true
		}
	}
	return false, false
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
		if len(m) == 0 {
			return nil
		}
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

// ComposeAllWorkspaceOps is like ComposeWorkspaceOps but composes any
// number of ops, not just 2.
func ComposeAllWorkspaceOps(ops []WorkspaceOp) (composed WorkspaceOp, err error) {
	for _, other := range ops {
		composed, err = ComposeWorkspaceOps(composed, other)
		if err != nil {
			return
		}
	}
	return
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
			delete(op.Edit, s)
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
				if edit, ok := op.Edit[s]; ok {
					delete(op.Edit, s)
					op.Edit[d] = edit
				}
				continue
			}
		}

		_, created := op.Create[s]
		if created {
			delete(op.Create, s)
			op.Create[d] = struct{}{}
		}
		if _, createdF := op.Create[d]; createdF && !created {
			delete(op.Create, d)
		}
		if !truncated && !created {
			op.Save[s] = struct{}{}
		}
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

		if _, copyTo := op.Copy[f]; copyTo {
			delete(op.Truncate, f)
			delete(op.Copy, f)
			op.Create[f] = struct{}{}
		} else {
			op.Truncate[f] = struct{}{}
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
	transform := func(x, y *tmpWorkspaceOp, primary bool) (err error) {
		// Basically, we want to remove things in y that are already
		// in x. When we begin, x and y are concurrent ops. When we're
		// done, we expect x and y to be consecutive ops.

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
				return y1, nil
			}
			return x1, nil
		}

		transformThenCompose := func(a, b EditOps) (EditOps, error) {
			_, b1, err := TransformEditOps(a, b)
			if err != nil {
				return nil, err
			}
			return ComposeEditOps(a, b1)
		}

		zexists := y.exists

		for s := range x.Save {
			if existed, known := y.existed(s); known && !existed {
				err = &os.PathError{Op: "save", Path: s, Err: os.ErrNotExist}
				return
			}
			delete(y.Save, s)

			var f string
			f, err = bufferToFilePath(s)
			if err != nil {
				return
			}
			delete(y.Create, f)

			if _, copiedTo := y.Copy[s]; !copiedTo {
				if e, ok := y.Edit[s]; ok && !e.Noop() {
					y.Edit[f], err = transformThenCompose(y.Edit[s], y.Edit[f])
					if err != nil {
						return err
					}
				}
				delete(y.Edit, s)
			}
		}
		for d, s := range x.Copy {
			if ys, ycopied := y.Copy[d]; ycopied && ys == s {
				delete(y.Copy, d)
				delete(y.copyFrom, s)
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
				delete(y.Rename, s)
				y.Delete[s] = struct{}{}
				continue
			} else if exists, known := y.existed(d); known && exists {
				err = &os.PathError{Op: "copy to", Path: d, Err: os.ErrExist}
				return
			}

			existed, existedknown := y.existed(d)
			exists, existsknown := y.exists(d)
			if (!existed && existedknown) && (exists && existsknown) {
				delete(y.Create, d)
				delete(x.Edit, d)
				y.Truncate[d] = struct{}{}
				continue
			}

			if x.Edit != nil {
				y.Edit[d] = y.Edit[s]
			}
			if x.Sel != nil {
				x.Sel[d] = x.Sel[s]
			}
		}

		for s, d := range x.Rename {
			if yd, yrenamed := y.Rename[d]; yrenamed && yd == s {
				delete(y.Rename, d)
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
				delete(y.Rename, s)
				continue
			} else if yrename {
				y.Copy[yd] = d
				delete(y.Rename, s)
				continue
			}
			if exists, known := zexists(s); known && !exists {
				err = &os.PathError{Op: "rename from", Path: s, Err: os.ErrNotExist}
				return
			}
			if ys, ycopied := y.Copy[d]; ycopied && ys == s {
				delete(y.Copy, d)
				continue
			} else if exists, known := y.exists(d); known && exists {
				err = &os.PathError{Op: "rename to", Path: d, Err: os.ErrExist}
				return
			}
			if _, ok := y.Edit[s]; ok {
				y.Edit[d] = y.Edit[s]
				delete(y.Edit, s)
			}
			if _, ok := y.Sel[s]; ok {
				y.Sel[d] = y.Sel[s]
				delete(y.Sel, s)
			}
		}

		for f := range x.Create {
			yexisted, yexistedknown := y.existed(f)
			if yexisted && yexistedknown {
				err = &os.PathError{Op: "create", Path: f, Err: os.ErrExist}
				return
			}
			if _, ydeleted := y.Delete[f]; ydeleted {
				return &os.PathError{Op: "create", Path: f, Err: errors.New("conflict: y delete")}
			}
			if _, ycreate := y.Create[f]; ycreate {
				delete(y.Create, f)
			} else {
				delete(y.Edit, f)
			}
			delete(y.Copy, f)
		}
		for f := range x.Delete {
			if exists, known := y.existed(f); known && !exists {
				err = &os.PathError{Op: "delete", Path: f, Err: os.ErrNotExist}
				return
			}
			if _, ycreated := y.Create[f]; ycreated {
				return &os.PathError{Op: "delete", Path: f, Err: errors.New("conflict: y create")}
			}
			delete(y.Delete, f)
			delete(y.Sel, f)
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
				if exists, known := zexists(f); known && !exists {
					err = &os.PathError{Op: "truncate", Path: f, Err: os.ErrNotExist}
					return
				}
			} else {
				delete(y.Truncate, f)
			}
			delete(y.Sel, f)
		}
		for f, edit := range x.Edit {
			if isFilePath(f) {
				b, _ := fileToBufferPath(f)
				if _, ysavedTo := y.Save[b]; ysavedTo {
					y.Edit[f], err = transformThenCompose(y.Edit[f], edit)
					if err != nil {
						return
					}
					delete(x.Edit, f)

					if s, copiedTo := y.Copy[b]; copiedTo && s == f {
						y.Edit[b], err = transformThenCompose(y.Edit[b], edit)
						if err != nil {
							return
						}
					}
				}
			}
		}
		for f, edit := range x.Edit {
			_ = edit
			if _, ysave := y.Save[f]; ysave {
				b := f
				f, _ := bufferToFilePath(b)
				y.Edit[f], err = transformEditOps(x.Edit[b], y.Edit[f])
				if err != nil {
					return
				}
				delete(x.Edit, b)
			}
		}

		for f, edit := range x.Edit {
			yf := f
			if _, ysave := y.Save[f]; ysave {
				yf, _ = bufferToFilePath(f)
			}
			if d, ycopy := y.Copy[yf]; ycopy {
				yf = d
			}
			if d, yrename := y.Rename[yf]; yrename {
				yf = d
			}
			if exists, known := y.exists(yf); known && !exists {
				err = &os.PathError{Op: "edit", Path: yf, Err: os.ErrNotExist}
				return
			}
			for _, d := range y.copyFrom[f] {
				y.Edit[d], err = transformEditOps(edit, y.Edit[d])
				if err != nil {
					return
				}
			}

			if yd, yrename := y.Rename[f]; yrename {
				delete(y.Edit, f)
				f = yd
			}
			if _, yedited := y.Edit[yf]; yedited {
				y.Edit[f], err = transformEditOps(edit, y.Edit[yf])
				if err != nil {
					return
				}
			}

			if y.Edit[f].Noop() {
				delete(y.Edit, f)
			}

			for u, sel := range y.Sel[f] {
				y.Sel[f][u] = AdjustSel(sel, edit)
			}
		}

		for f, usel := range x.Sel {
			f2, ok := y.Rename[f]
			if ok {
				f = f2
			}
			for u, sel := range usel {
				_ = sel
				ysel, yselected := y.Sel[f][u]
				if yselected && (sel.equal(ysel) || primary) {
					delete(y.Sel[f], u)
					if len(y.Sel[f]) == 0 {
						delete(y.Sel, f)
					}
				}
			}
		}

		return
	}

	var at1, bt1, at2, bt2 tmpWorkspaceOp
	if err = at1.from(a); err != nil {
		return
	}
	if err = bt1.from(b); err != nil {
		return
	}
	if err = at2.from(a); err != nil {
		return
	}
	if err = bt2.from(b); err != nil {
		return
	}
	// at1 == at2, bt1 == bt2

	if err = transform(&at1, &bt1, true); err != nil {
		return
	}
	if err = transform(&bt2, &at2, false); err != nil {
		return
	}

	a1 = at2.toWorkspaceOp()
	b1 = bt1.toWorkspaceOp()

	if a.GitHead != b.GitHead {
		if a.GitHead != "" && b.GitHead != "" {
			err = fmt.Errorf("transform encountered conflicting heads: %q != %q", a.GitHead, b.GitHead)
			return
		}
		a1.GitHead = a.GitHead
		b1.GitHead = b.GitHead
	} else {
		a1.GitHead = ""
		b1.GitHead = ""
	}

	a1 = NormalizeWorkspaceOp(a1)
	b1 = NormalizeWorkspaceOp(b1)
	return
}

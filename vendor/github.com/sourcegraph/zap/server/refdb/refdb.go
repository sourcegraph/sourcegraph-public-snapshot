package refdb

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

// RefDB is the interface to the Zap refdb. Ref values are immutable
// (although they may contain pointers to data structures that are not
// immutable).
type RefDB interface {
	// Exists queries the refdb and reports whether the named
	// reference exists.
	Exists(name string) bool

	// Lookup queries the refdb and returns the named reference (or
	// nil if it doesn't exist).
	Lookup(name string) *Ref

	// Resolve queries the refdb recursively to resolve the named
	// reference to a non-symbolic reference. If, at any step, the
	// reference's target does not refer to an existing reference, an
	// error of type *RefNotExistsError is returned.
	Resolve(name string) (*Ref, error)

	// List queries the refdb and returns all references that match
	// pattern. The pattern must match the entire reference name, and
	// it can contain '*' characters, which match any string of 0 or
	// more characters.
	//
	// Examples:
	//
	// - pattern "foo" matches "foo" but does not "foo/bar" or "bar/foo"
	// - pattern "foo*" matches "foo" and "foo/bar" but not "bar/foo"
	// - pattern "*foo*" matches "foo" and "bar/foo/baz" but not "qux"
	List(pattern string) []Ref

	// Write writes the given reference to the refdb.
	//
	// If force is false and a ref with the given name already exists,
	// an error of type *RefExistsError is returned. If force is true,
	// then it will overwrite any existing reference with the same
	// name.
	//
	// Write returns an error of type *WrongOldRefValueError if the
	// given old ref value differs from the actual value of the
	// existing ref (or nil if there is no existing ref).
	//
	// The log information is written to the reflog if this write
	// succeeds. It should describe the reason for the update.
	Write(ref Ref, force bool, old *Ref, log RefLogEntry) error

	// Rename renames the given reference in the refdb and returns the
	// reference with the new name.
	//
	// After calling Rename, no ref exists at the old name. The ref
	// that previously existed under the old name is now available at
	// the new name.
	//
	// As with Write, force determines whether an existing ref at the
	// new name will be overwritten. If force is false and a ref
	// exists at the new name, an error of type *RefExistsError is
	// returned.
	Rename(oldName, newName string, force bool, log RefLogEntry) (*Ref, error)

	// Delete deletes the named reference from the refdb. As with
	// Write, the old ref value must match the actual value of the
	// existing ref, or an error of type *WrongOldRefValueError is
	// returned.
	Delete(name string, old Ref, log RefLogEntry) error

	// TransitiveClosureRefs queries the refdb and returns the named
	// reference plus all symbolic references that refer to the named
	// reference, either directly or indirectly (via other symbolic
	// references).
	//
	// If no such named ref exists, it returns nil.
	TransitiveClosureRefs(name string) []Ref

	// RefLog returns the reflog used by this refdb. Subsequent calls
	// of this method always return the same value.
	RefLog() RefLog
}

// Ref represents a Zap ref. A ref is either an object reference (in
// which case the Object and Rev fields are consulted) or a symbolic
// reference (in which case Target is consulted).
//
// A Ref is immutable.
type Ref struct {
	Name string // ref name (e.g., "branch/foo")

	Target string // for symbolic refs, the target ref name

	Object interface{} // a pointer to a long-lived object
	Rev    uint        // the revision in Object's history that this reference refers to
}

// IsSymbolic reports whether this ref is a symbolic reference.
func (r Ref) IsSymbolic() bool {
	if r.Target != "" && (r.Object != nil || r.Rev != 0) {
		panic(fmt.Sprintf("invalid ref %q: exactly one of target or object/rev must be set", r.Name))
	}
	return r.Target != ""
}

// RefExistsError indicates that a reference unexpectedly exists at a
// given name.
//
// Note: The force parameter to (RefDB).Write and (RefDB).Rename
// causes an existing reference to be overwritten, suppressing any
// RefExistsError that might otherwise be returned..
type RefExistsError struct {
	Op   string // the operation that yielded this error ("write" or "rename")
	Name string // ref name that exists
}

func (e *RefExistsError) Error() string {
	return fmt.Sprintf("refdb %s: ref exists: %s", e.Op, e.Name)
}

// RefNotExistsError indicates that a reference unexpectedly did not
// exist at a given name.
type RefNotExistsError struct {
	Op   string // the operation that yielded this error ("rename" or "delete")
	Name string // ref name that does not exist
}

func (e *RefNotExistsError) Error() string {
	return fmt.Sprintf("refdb %s: ref does not exist: %s", e.Op, e.Name)
}

// checkOldValue returns whether expected is the correct old value for
// actual.
func checkOldValue(actual, expected *Ref) (reasons []string, ok bool) {
	if actual == expected {
		return nil, true // same pointers or both nil
	}
	if actual == nil && expected != nil {
		return []string{"actual == nil && expected != nil"}, false
	}
	if actual != nil && expected == nil {
		return []string{"actual != nil && expected == nil"}, false
	}
	if actual.IsSymbolic() && !expected.IsSymbolic() {
		reasons = append(reasons, fmt.Sprintf("type: symbolic ref (%s) != object ref", actual.Target))
	} else if !actual.IsSymbolic() && expected.IsSymbolic() {
		reasons = append(reasons, fmt.Sprintf("type: object ref != symbolic ref (%s)", expected.Target))
	} else {
		// Refs are of the same type.
		if actual.Target != expected.Target {
			reasons = append(reasons, fmt.Sprintf("target: %q != %q", actual.Target, expected.Target))
		}
		if actual.Name != expected.Name {
			reasons = append(reasons, fmt.Sprintf("name: %q != %q", actual.Name, expected.Name))
		}
		if actual.Rev != expected.Rev {
			reasons = append(reasons, fmt.Sprintf("rev: %d != %d", actual.Rev, expected.Rev))
		}
		if actual.Object != expected.Object {
			reasons = append(reasons, fmt.Sprintf("object: %T@%p != %T@%p", actual.Object, actual.Object, expected.Object, expected.Object))
		}
	}
	return reasons, len(reasons) == 0
}

// WrongOldRefValueError indicates that during a refdb operation, the
// actual value of an existing ref (or nil if there is no existing
// ref) differed from the given old value.
type WrongOldRefValueError struct {
	Op       string // the operation that yielded this error ("write" or "delete")
	Name     string // the name of the ref whose value was wrong
	Actual   *Ref   // the actual value of the existing ref (or nil if there is no existing ref)
	Expected *Ref   // the given old value
}

func (e *WrongOldRefValueError) Error() string {
	reasons, _ := checkOldValue(e.Actual, e.Expected)
	return fmt.Sprintf("refdb %s: wrong old value for ref %s: %s", e.Op, e.Name, strings.Join(reasons, ", "))
}

// CircularSymbolicReferenceError indicates that during a refdb
// operation, a symbolic reference cycle was encountered.
type CircularSymbolicReferenceError struct {
	Op   string // the operation that yielded this error ("resolve")
	Name string // the name of the ref that was encountered twice
}

func (e *CircularSymbolicReferenceError) Error() string {
	return fmt.Sprintf("refdb %s: circular symbolic reference at ref %s", e.Op, e.Name)
}

// RefLog is the interface to the Zap reflog, which records when
// references were updated (and why). This helps avoid data loss and
// simplifies debugging.
//
// It is not possible to write to the reflog directly. Only a RefDB
// can write to a reflog.
type RefLog interface {
	// Read returns the reflog for the named reference.
	Read(name string) ([]RefLogEntry, error)

	// Rename renames a reflog.
	Rename(oldName, newName string) error

	// Delete deletes a reflog.
	Delete(name string) error
}

// RefLogEntry is an entry in the reflog that records an update to a
// ref and why it occurred.
//
// A RefLogEntry is immutable.
type RefLogEntry struct {
	Who     string    // the name of the user or actor who caused the update
	Date    time.Time // when the update occurred
	Message string    // a message explaining the reason for the update
}

// sortRefs sorts refs in place.
func sortRefs(refs []Ref) {
	sort.Sort(sortableRefs(refs))
}

type sortableRefs []Ref

func (v sortableRefs) Len() int           { return len(v) }
func (v sortableRefs) Swap(i, j int)      { v[i], v[j] = v[j], v[i] }
func (v sortableRefs) Less(i, j int) bool { return v[i].Name < v[j].Name }

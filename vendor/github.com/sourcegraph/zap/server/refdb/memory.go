package refdb

import "sync"

// memoryRefDB is an in-memory refdb. It implements RefLog.
//
// It is not safe to access it from multiple goroutines
// concurrently. The use of a mutex internally in memoryRefDB and its
// methods could make concurrent access memory-safe, but that would
// likely mask application-level raciness in the caller.
type memoryRefDB struct {
	refs map[string]Ref

	reflog memoryRefLog
}

// NewMemoryRefDB creates a new in-memory refdb (with an in-memory
// reflog).
func NewMemoryRefDB() RefDB {
	return &memoryRefDB{
		refs: make(map[string]Ref, 10),
		reflog: memoryRefLog{
			refs: make(map[string][]RefLogEntry, 10),
		},
	}
}

// Exists implements RefDB.
func (db *memoryRefDB) Exists(name string) bool {
	return db.exists(name)
}

// exists assume the caller holds db.mu.
func (db *memoryRefDB) exists(name string) bool {
	_, ok := db.refs[name]
	return ok
}

// Lookup implements RefDB.
func (db *memoryRefDB) Lookup(name string) *Ref {
	return db.lookup(name)
}

// Resolve implements RefDB.
func (db *memoryRefDB) Resolve(name string) (*Ref, error) {
	ref := db.lookup(name)
	seen := map[string]struct{}{name: struct{}{}}
	for ref != nil && ref.IsSymbolic() {
		if _, seen := seen[ref.Target]; seen {
			return nil, &CircularSymbolicReferenceError{Op: "resolve", Name: ref.Target}
		}
		name = ref.Target
		ref = db.lookup(name)
	}
	if ref == nil {
		return nil, &RefNotExistsError{Op: "resolve", Name: name}
	}
	return ref, nil
}

// List implements RefDB.
func (db *memoryRefDB) List(pattern string) []Ref {
	var matches []Ref
	for _, ref := range db.refs {
		if MatchPattern(pattern, ref.Name) {
			matches = append(matches, ref)
		}
	}
	return matches
}

// lookup assumes the caller holds db.mu.
func (db *memoryRefDB) lookup(name string) *Ref {
	if ref, ok := db.refs[name]; ok {
		return &ref
	}
	return nil
}

// Write implements RefDB.
func (db *memoryRefDB) Write(ref Ref, force bool, old *Ref, log RefLogEntry) error {
	existing := db.lookup(ref.Name)
	if !force && existing != nil {
		return &RefExistsError{Op: "write", Name: ref.Name}
	}
	if _, ok := checkOldValue(existing, old); !ok {
		return &WrongOldRefValueError{Op: "write", Name: ref.Name, Actual: existing, Expected: old}
	}
	db.refs[ref.Name] = ref
	if err := db.reflog.write(ref.Name, old, &ref, log); err != nil {
		return err
	}
	return nil
}

// Rename implements RefDB.
func (db *memoryRefDB) Rename(oldName, newName string, force bool, log RefLogEntry) (*Ref, error) {
	oldRef, newNameExists := db.refs[newName]
	if !force && newNameExists {
		return nil, &RefExistsError{Op: "rename", Name: newName}
	}
	newRef, oldNameExists := db.refs[oldName]
	if !oldNameExists {
		return nil, &RefNotExistsError{Op: "rename", Name: oldName}
	}
	newRef.Name = newName
	delete(db.refs, oldName)
	db.refs[newName] = newRef

	var oldRefPtr *Ref
	if newNameExists {
		oldRefPtr = &oldRef
	}
	if err := db.reflog.renameAndWrite(oldName, newName, oldRefPtr, newRef, log); err != nil {
		return nil, err
	}
	return &newRef, nil
}

// Delete implements RefDB.
func (db *memoryRefDB) Delete(name string, old Ref, log RefLogEntry) error {
	existing := db.lookup(name)
	if existing == nil {
		return &RefNotExistsError{Op: "delete", Name: name}
	}
	if _, ok := checkOldValue(existing, &old); !ok {
		return &WrongOldRefValueError{Op: "delete", Name: name, Actual: existing, Expected: &old}
	}
	delete(db.refs, name)
	if err := db.reflog.Delete(name); err != nil {
		return err
	}
	return nil
}

// TransitiveClosureRefs implements RefDB.
func (db *memoryRefDB) TransitiveClosureRefs(name string) (refs []Ref) {
	for _, ref := range db.refs {
		if ref.Name == name {
			refs = append(refs, ref)
			continue
		}

		orig := ref
		if ref.IsSymbolic() {
			seen := map[string]struct{}{}
			for ref.IsSymbolic() {
				// Detect cycles.
				if _, seen := seen[ref.Name]; seen {
					break
				}
				seen[ref.Name] = struct{}{}

				next, exists := db.refs[ref.Target]
				if !exists {
					break
				}
				if next.Name == name {
					refs = append(refs, orig)
					break
				}

				ref = next
			}
		}
	}
	sortRefs(refs)
	return refs
}

// RefLog implements RefDB.
func (db *memoryRefDB) RefLog() RefLog { return &db.reflog }

// memoryRefLog is an in-memory reflog. It implements RefLog.
type memoryRefLog struct {
	mu   sync.Mutex
	refs map[string][]RefLogEntry
}

// write writes an entry to the reflog.
func (l *memoryRefLog) write(name string, old, new *Ref, log RefLogEntry) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.refs[name] = append(l.refs[name], log)
	return nil
}

// renameAndWrite renames the old reflog to the new reflog and then
// writes an entry to the new reflog.
func (l *memoryRefLog) renameAndWrite(oldName, newName string, old *Ref, new Ref, log RefLogEntry) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	entries := l.refs[oldName]
	entries = append(entries, log)
	l.refs[newName] = entries
	delete(l.refs, oldName)
	return nil
}

// Read implements RefLog.
func (l *memoryRefLog) Read(name string) ([]RefLogEntry, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.refs[name], nil
}

// Rename implements RefLog.
func (l *memoryRefLog) Rename(oldName, newName string) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.refs[newName] = l.refs[oldName]
	delete(l.refs, oldName)
	return nil
}

// Delete implements RefLog.
func (l *memoryRefLog) Delete(name string) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.refs, name)
	return nil
}

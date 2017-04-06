package refdb

import (
	"fmt"
	"sync"
)

// Sync returns the refdb wrapped with the SyncRefDB synchronization
// wrapper. See the SyncRefDB docstring for more information.
func Sync(refdb RefDB) *SyncRefDB {
	return &SyncRefDB{db: refdb}
}

// A OwnedRef is a Ref that holds an exclusive lock in a refdb on its
// ref name until its Unlock method is called.
type OwnedRef struct {
	Ref *Ref // a ref (or nil if the ref did not exist)

	name   string // the ref name that yielded this OwnedRef (constraint: Ref == nil || Ref.Name == name)
	unlock func() // unlocks the ref name lock in the containing refdb
}

// Unlock unlocks the exclusive lock on r's ref name in its
// refdb. After calling Unlock, r.Ref is no longer guaranteed to be
// current.
func (r *OwnedRef) Unlock() {
	if r.unlock == nil {
		panic("Unlock was already called on ref " + r.name)
	}
	r.unlock()
	r.unlock = nil
}

// A SyncError occurs when a SyncRefDB method is called and the caller
// is not holding the necessary lock.
type SyncError struct {
	Op     string // the method that was called
	Ref    string // the ref name that the operation was being performed on
	Locked string // the ref name whose lock is held (or "" if no lock is held)
}

func (e *SyncError) Error() string {
	if e.Locked != "" {
		return fmt.Sprintf("refdb sync %s: locked ref %q != %q", e.Op, e.Locked, e.Ref)
	}
	return fmt.Sprintf("refdb sync %s: ref %q not locked", e.Op, e.Ref)
}

// SyncRefDB wraps an underlying refdb, adding synchronization to make
// it safe for concurrent access by multiple goroutines.
//
// It provides for exclusive locks on ref names. The holder of a ref
// name lock must release the lock before any other operation can be
// performed on that ref name.
//
// The lock is on a ref name, not a ref. This means a call to
// Lookup("x") acquires a lock on "x" even if there is no existing ref
// "x". This allows the caller to then create the ref "x" without
// racing other concurrent goroutines that are trying to do the same
// thing.
type SyncRefDB struct {
	mu sync.RWMutex
	db RefDB

	nameMu sync.Mutex
	name   map[string]*sync.Mutex // ref name -> lock
}

// lock locks the provided ref name. The caller must not hold db.mu
// (or else there will likely be a deadlock).
func (db *SyncRefDB) lock(ref string) (unlock func()) {
	db.nameMu.Lock()
	if db.name == nil {
		db.name = map[string]*sync.Mutex{}
	}
	mu, ok := db.name[ref]
	if !ok {
		mu = new(sync.Mutex)
		db.name[ref] = mu
	}
	db.nameMu.Unlock()
	mu.Lock()
	return mu.Unlock
}

// Lookup acquires an exclusive lock on the ref name and calls
// (RefDB).Lookup on the underlying refdb.
//
// The caller of Lookup is responsible for unlocking the OwnedRef. For
// example:
//
//   ref := refdb.Lookup("x")
//   defer ref.Unlock()
//
// Note that even if the ref doesn't exist (ref.Ref == nil), the call
// to Lookup still acquires the *ref name* lock, which must be
// unlocked.
func (db *SyncRefDB) Lookup(name string) OwnedRef {
	unlock := db.lock(name)
	db.mu.RLock()
	defer db.mu.RUnlock()
	return OwnedRef{Ref: db.db.Lookup(name), name: name, unlock: unlock}
}

// LookupShared looks up the named ref without acquiring an exclusive
// lock. It is safe to call Lookup from concurrent goroutines, but the
// returned ref may become immediately out of date. The only safe
// reason to call LookupShared is if you already hold the lock on the
// ref name and want to see the new ref in the refdb after some other
// action you synchronously performed.
func (db *SyncRefDB) LookupShared(name string) *Ref {
	db.mu.Lock()
	defer db.mu.Unlock()
	return db.db.Lookup(name)
}

// List calls (RefDB).List on the underlying refdb. It is safe to call
// List from multiple concurrent goroutines. It does not attempt to
// lock each ref, because doing so would be prone to deadlocks;
// therefore the caller does not hold the ref name locks on the
// resulting refs.
func (db *SyncRefDB) List(pattern string) []Ref {
	db.mu.RLock()
	defer db.mu.RUnlock()
	return db.db.List(pattern)
}

// Resolve looks up the named ref. If name refers to a symbolic ref,
// its target is also looked up and returned.
//
// The lock for the named ref is always acquired.
//
// * If no ref exists with the provided name, target is nil and no
//   lock is acquired for the target ref name.
// * If the named ref is a symbolic ref, it also acquires the lock on
//   the target ref name. The returned target is non-nil (although
//   target.Ref == nil is possible if the symbolic ref's target points
//   to a nonexistent ref).
//
// Unlike (RefDB).Resolve, it returns an error if the symbolic
// reference's target is also a symbolic reference. This is to
// simplify the implementation (to not need to handle chains or detect
// cycles).
//
// The caller of Resolve is responsible for unlocking ref and (if
// non-nil) target. For example:
//
//   ref, target := refdb.Resolve("x")
//   defer ref.Unlock()
//   if target != nil {
//   	defer target.Unlock()
//   }
func (db *SyncRefDB) Resolve(name string) (ref OwnedRef, target *OwnedRef) {
	ref = db.Lookup(name)
	if ref.Ref != nil && ref.Ref.Target() != "" && ref.Ref.Target() != name {
		targetRef := db.Lookup(ref.Ref.Target())
		target = &targetRef
	}
	return ref, target
}

// Write first checks that ref is still holding its lock. If so, it
// calls (RefDB).Write on the underlying refdb to write ref.
//
// It panics if ref.Ref == nil.
func (db *SyncRefDB) Write(ref OwnedRef) error {
	if ref.Ref == nil {
		panic("(*SyncRefDB).Write: ref.Ref == nil")
	}
	if ref.unlock == nil {
		return &SyncError{Op: "Write", Ref: ref.Ref.Name}
	}
	if ref.Ref.Name != ref.name {
		return &SyncError{Op: "Write", Ref: ref.Ref.Name, Locked: ref.name}
	}
	db.mu.Lock()
	defer db.mu.Unlock()
	return db.db.Write(*ref.Ref, true, db.db.Lookup(ref.Ref.Name), RefLogEntry{})
}

// CompareAndWrite calls (RefDB.Write). It is safe to call from
// multiple concurrent goroutines.
func (db *SyncRefDB) CompareAndWrite(ref Ref, old *Ref) error {
	db.mu.Lock()
	defer db.mu.Unlock()
	return db.db.Write(ref, true, old, RefLogEntry{})
}

// Delete first checks that ref is still holding its lock. If so, it
// calls (RefDB).Delete on the underlying refdb to delete ref.
//
// It panics if ref.Ref == nil.
func (db *SyncRefDB) Delete(ref OwnedRef) error {
	if ref.Ref == nil {
		panic("(*SyncRefDB).Delete: ref.Ref == nil")
	}
	if ref.unlock == nil {
		return &SyncError{Op: "Delete", Ref: ref.Ref.Name}
	}
	if ref.Ref.Name != ref.name {
		return &SyncError{Op: "Delete", Ref: ref.Ref.Name, Locked: ref.name}
	}
	db.mu.Lock()
	defer db.mu.Unlock()
	return db.db.Delete(ref.Ref.Name, *ref.Ref, RefLogEntry{})
}

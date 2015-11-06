// +build !js

// Package bypass allows bypassing reflect restrictions on accessing unexported struct fields.
package bypass

import (
	"reflect"
	"unsafe"
)

// This code currently matches unexported code in https://github.com/davecgh/go-spew/blob/master/spew/common.go.

const (
	// ptrSize is the size of a pointer on the current arch.
	ptrSize = unsafe.Sizeof((*byte)(nil))
)

var (
	// offsetPtr, offsetScalar, and offsetFlag are the offsets for the
	// internal reflect.Value fields.  These values are valid before golang
	// commit ecccf07e7f9d which changed the format.  The are also valid
	// after commit 82f48826c6c7 which changed the format again to mirror
	// the original format.  Code in the init function updates these offsets
	// as necessary.
	offsetPtr    = uintptr(ptrSize)
	offsetScalar = uintptr(0)
	offsetFlag   = uintptr(ptrSize * 2)

	// flagKindWidth and flagKindShift indicate various bits that the
	// reflect package uses internally to track kind information.
	//
	// flagRO indicates whether or not the value field of a reflect.Value is
	// read-only.
	//
	// flagIndir indicates whether the value field of a reflect.Value is
	// the actual data or a pointer to the data.
	//
	// These values are valid before golang commit 90a7c3c86944 which
	// changed their positions.  Code in the init function updates these
	// flags as necessary.
	flagKindWidth = uintptr(5)
	flagKindShift = uintptr(flagKindWidth - 1)
	flagRO        = uintptr(1 << 0)
	flagIndir     = uintptr(1 << 1)
)

func init() {
	// Older versions of reflect.Value stored small integers directly in the
	// ptr field (which is named val in the older versions).  Versions
	// between commits ecccf07e7f9d and 82f48826c6c7 added a new field named
	// scalar for this purpose which unfortunately came before the flag
	// field, so the offset of the flag field is different for those
	// versions.
	//
	// This code constructs a new reflect.Value from a known small integer
	// and checks if the size of the reflect.Value struct indicates it has
	// the scalar field. When it does, the offsets are updated accordingly.
	vv := reflect.ValueOf(0xf00)
	if unsafe.Sizeof(vv) == (ptrSize * 4) {
		offsetScalar = ptrSize * 2
		offsetFlag = ptrSize * 3
	}

	// Commit 90a7c3c86944 changed the flag positions such that the low
	// order bits are the kind.  This code extracts the kind from the flags
	// field and ensures it's the correct type.  When it's not, the flag
	// order has been changed to the newer format, so the flags are updated
	// accordingly.
	upf := unsafe.Pointer(uintptr(unsafe.Pointer(&vv)) + offsetFlag)
	upfv := *(*uintptr)(upf)
	flagKindMask := uintptr((1<<flagKindWidth - 1) << flagKindShift)
	if (upfv&flagKindMask)>>flagKindShift != uintptr(reflect.Int) {
		flagKindShift = 0
		flagRO = 1 << 5
		flagIndir = 1 << 6
	}
}

// UnsafeReflectValue converts the passed reflect.Value into a one that bypasses
// the typical safety restrictions preventing access to unaddressable and
// unexported data.  It works by digging the raw pointer to the underlying
// value out of the protected value and generating a new unprotected (unsafe)
// reflect.Value to it.
//
// This allows us to check for implementations of the Stringer and error
// interfaces to be used for pretty printing ordinarily unaddressable and
// inaccessible values such as unexported struct fields.
func UnsafeReflectValue(v reflect.Value) (rv reflect.Value) {
	indirects := 1
	vt := v.Type()
	upv := unsafe.Pointer(uintptr(unsafe.Pointer(&v)) + offsetPtr)
	rvf := *(*uintptr)(unsafe.Pointer(uintptr(unsafe.Pointer(&v)) + offsetFlag))
	if rvf&flagIndir != 0 {
		vt = reflect.PtrTo(v.Type())
		indirects++
	} else if offsetScalar != 0 {
		// The value is in the scalar field when it's not one of the
		// reference types.
		switch vt.Kind() {
		case reflect.Uintptr:
		case reflect.Chan:
		case reflect.Func:
		case reflect.Map:
		case reflect.Ptr:
		case reflect.UnsafePointer:
		default:
			upv = unsafe.Pointer(uintptr(unsafe.Pointer(&v)) +
				offsetScalar)
		}
	}

	pv := reflect.NewAt(vt, upv)
	rv = pv
	for i := 0; i < indirects; i++ {
		rv = rv.Elem()
	}
	return rv
}

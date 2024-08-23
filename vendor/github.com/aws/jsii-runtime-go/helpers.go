package jsii

import (
	"fmt"
	"time"
)

type basicType interface {
	bool | string | float64 | time.Time
}

// Ptr returns a pointer to the provided value.
func Ptr[T basicType](v T) *T {
	return &v
}

// PtrSlice returns a pointer to a slice of pointers to all of the provided values.
func PtrSlice[T basicType](v ...T) *[]*T {
	slice := make([]*T, len(v))
	for i := 0; i < len(v); i++ {
		slice[i] = Ptr(v[i])
	}
	return &slice
}

// Bool returns a pointer to the provided bool.
func Bool(v bool) *bool { return Ptr(v) }

// Bools returns a pointer to a slice of pointers to all of the provided booleans.
func Bools(v ...bool) *[]*bool {
	return PtrSlice(v...)
}

type numberType interface {
	~float32 | ~float64 |
		~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr
}

// Number returns a pointer to the provided float64.
func Number[T numberType](v T) *float64 {
	return Ptr(float64(v))
}

// Numbers returns a pointer to a slice of pointers to all of the provided numbers.
func Numbers[T numberType](v ...T) *[]*float64 {
	slice := make([]*float64, len(v))
	for i := 0; i < len(v); i++ {
		slice[i] = Number(v[i])
	}
	return &slice
}

// String returns a pointer to the provided string.
func String(v string) *string { return Ptr(v) }

// Sprintf returns a pointer to a formatted string (semantics are the same as fmt.Sprintf).
func Sprintf(format string, a ...interface{}) *string {
	res := fmt.Sprintf(format, a...)
	return &res
}

// Strings returns a pointer to a slice of pointers to all of the provided strings.
func Strings(v ...string) *[]*string {
	return PtrSlice(v...)
}

// Time returns a pointer to the provided time.Time.
func Time(v time.Time) *time.Time { return Ptr(v) }

// Times returns a pointer to a slice of pointers to all of the provided time.Time values.
func Times(v ...time.Time) *[]*time.Time {
	return PtrSlice(v...)
}

package unsafeslice

import (
	"reflect"
	"unsafe"
)

const (
	Uint64Size = 8
	Uint32Size = 4
	Uint16Size = 2
	Uint8Size  = 1
)

func newRawSliceHeader(sh *reflect.SliceHeader, b []byte, stride int) *reflect.SliceHeader {
	sh.Len = len(b) / stride
	sh.Cap = len(b) / stride
	sh.Data = (uintptr)(unsafe.Pointer(&b[0]))
	return sh
}

func newSliceHeader(b []byte, stride int) unsafe.Pointer {
	sh := &reflect.SliceHeader{}
	return unsafe.Pointer(newRawSliceHeader(sh, b, stride))
}

func Uint64SliceFromByteSlice(b []byte) []uint64 {
	return *(*[]uint64)(newSliceHeader(b, Uint64Size))
}

func Int64SliceFromByteSlice(b []byte) []int64 {
	return *(*[]int64)(newSliceHeader(b, Uint64Size))
}

func Uint32SliceFromByteSlice(b []byte) []uint32 {
	return *(*[]uint32)(newSliceHeader(b, Uint32Size))
}

func Int32SliceFromByteSlice(b []byte) []int32 {
	return *(*[]int32)(newSliceHeader(b, Uint32Size))
}

func Uint16SliceFromByteSlice(b []byte) []uint16 {
	return *(*[]uint16)(newSliceHeader(b, Uint16Size))
}

func Int16SliceFromByteSlice(b []byte) []int16 {
	return *(*[]int16)(newSliceHeader(b, Uint16Size))
}

func Uint8SliceFromByteSlice(b []byte) []uint8 {
	return *(*[]uint8)(newSliceHeader(b, Uint8Size))
}

func Int8SliceFromByteSlice(b []byte) []int8 {
	return *(*[]int8)(newSliceHeader(b, Uint8Size))
}

// Create a slice of structs from a slice of bytes.
//
// 		var v []Struct
// 		StructSliceFromByteSlice(bytes, &v)
//
// Elements in the byte array must be padded correctly. See unsafe.AlignOf, et al.
func StructSliceFromByteSlice(b []byte, out interface{}) {
	ptr := reflect.ValueOf(out)
	if ptr.Kind() != reflect.Ptr {
		panic("expected pointer to a slice of structs (*[]X)")
	}
	slice := ptr.Elem()
	if slice.Kind() != reflect.Slice {
		panic("expected pointer to a slice of structs (*[]X)")
	}
	// TODO: More checks, such as ensuring that:
	// - elements are NOT pointers
	// - structs do not contain pointers, slices or maps
	stride := int(slice.Type().Elem().Size())
	if len(b)%stride != 0 {
		panic("size of byte buffer is not a multiple of struct size")
	}
	sh := (*reflect.SliceHeader)(unsafe.Pointer(slice.UnsafeAddr()))
	newRawSliceHeader(sh, b, stride)
}

func ByteSliceFromStructSlice(s interface{}) []byte {
	slice := reflect.ValueOf(s)
	if slice.Kind() != reflect.Slice {
		panic("expected a slice of structs (*[]X)")
	}
	var length int
	var data uintptr
	if slice.Len() != 0 {
		elem := slice.Index(0)
		length = int(elem.Type().Size()) * slice.Len()
		data = elem.UnsafeAddr()
	}
	out := &reflect.SliceHeader{
		Len:  length,
		Cap:  length,
		Data: data,
	}
	return *(*[]byte)(unsafe.Pointer(out))
}

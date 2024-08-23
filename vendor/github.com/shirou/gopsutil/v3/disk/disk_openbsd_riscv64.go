//go:build openbsd && riscv64
// +build openbsd,riscv64

// Code generated by cmd/cgo -godefs; DO NOT EDIT.
// cgo -godefs disk/types_openbsd.go

package disk

const (
	devstat_NO_DATA = 0x00
	devstat_READ    = 0x01
	devstat_WRITE   = 0x02
	devstat_FREE    = 0x03
)

const (
	sizeOfDiskstats = 0x70
)

type (
	Diskstats struct {
		Name       [16]int8
		Busy       int32
		Rxfer      uint64
		Wxfer      uint64
		Seek       uint64
		Rbytes     uint64
		Wbytes     uint64
		Attachtime Timeval
		Timestamp  Timeval
		Time       Timeval
	}
	Timeval struct {
		Sec  int64
		Usec int64
	}
)

type Diskstat struct{}
type bintime struct{}

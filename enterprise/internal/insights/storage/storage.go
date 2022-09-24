package storage

type DataFormat int

const (
	_ DataFormat = iota
	Uncompressed
	Gorilla
)

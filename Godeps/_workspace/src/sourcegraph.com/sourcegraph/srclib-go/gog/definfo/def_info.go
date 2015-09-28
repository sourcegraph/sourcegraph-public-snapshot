package definfo

type DefInfo struct {
	// Exported is whether this def is exported.
	Exported bool `json:",omitempty"`

	// PkgScope is whether this def is in Go package scope.
	PkgScope bool `json:",omitempty"`

	// PkgName is the name (not import path) of the package containing this
	// def.
	PkgName string

	// Receiver is the receiver of this def (or the empty string if this
	// def is not a method).
	Receiver string `json:",omitempty"`

	// FieldOfStruct is the struct that this def is a field of (or the empty string if this
	// def is not a struct field).
	FieldOfStruct string `json:",omitempty"`

	// TypeString is a string describing this def's Go type.
	TypeString string

	// UnderlyingTypeString is the function or method signature, if this is a function or method.
	UnderlyingTypeString string `json:",omitempty"`

	// Kind is the kind of Go thing this def is: struct, interface, func,
	// package, etc.
	Kind string `json:",omitempty"`
}

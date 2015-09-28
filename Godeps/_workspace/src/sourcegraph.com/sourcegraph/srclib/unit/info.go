package unit

type Info struct {
	// NameInRepository is the name to use when displaying the source unit in
	// the context of the repository in which it is defined. This name
	// typically needs less qualification than GlobalName.
	//
	// For example, a Go package's GlobalName is its repository URI basename
	// plus its directory path within the repository (e.g.,
	// "github.com/user/repo/x/y"'s NameInRepository is "repo/x/y"). Because npm
	// and pip packages are named globally, their name is probably appropriate
	// to use as both the unit's NameInRepository and GlobalName.
	NameInRepository string

	// GlobalName is the name to use when displaying the source unit *OUTSIDE OF*
	// the context of the repository in which it is defined.
	//
	// For example, a Go package's GlobalName is its full import path. Because
	// npm and pip packages are named globally, their name is probably
	// appropriate to use as both the unit's NameInRepository and GlobalName.
	GlobalName string

	// Description is a short (~1-sentence) description of the source unit.
	Description string

	// TypeName is the human-readable name of the type of source unit; e.g., "Go
	// package".
	TypeName string
}

// GetInfo returns a source unit's Info field if non-nil, or otherwise an Info
// struct filled with values derived from the source unit's name and type.
func GetInfo(u SourceUnit) Info {
	if u.Info != nil {
		return *u.Info
	}
	return Info{
		NameInRepository: u.Name,
		GlobalName:       u.Name,
		TypeName:         u.Type,
	}
}

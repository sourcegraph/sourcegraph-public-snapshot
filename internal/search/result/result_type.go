package result

import "strings"

// Type represents a single, named result type.
type Type uint8

const (
	TypeRepo Type = 1 << iota
	TypeSymbol
	TypeFile
	TypePath
	TypeDiff
	TypeCommit
)

func (r Type) String() string {
	switch r {
	case TypeRepo:
		return "repo"
	case TypeSymbol:
		return "symbol"
	case TypeFile:
		return "file"
	case TypePath:
		return "path"
	case TypeDiff:
		return "diff"
	case TypeCommit:
		return "commit"
	}
	return ""
}

var TypeFromString = map[string]Type{
	"repo":   TypeRepo,
	"symbol": TypeSymbol,
	"file":   TypeFile,
	"path":   TypePath,
	"diff":   TypeDiff,
	"commit": TypeCommit,
}

// Types represents a set of result types.
// It's a bitset corresponding to the disjunction of types it represents.
//
// For example, the set of file and repo results
// is represented as Types(TypeFile|TypeRepo)
type Types uint32

func (r Types) Has(t Type) bool {
	return r&Types(t) != 0
}

func (r Types) With(t Type) Types {
	return r | Types(t)
}

func (r Types) Without(t Type) Types {
	return r &^ Types(t)
}

func (r Types) String() string {
	var s strings.Builder
	first := true
	for _, t := range []Type{TypeRepo, TypeSymbol, TypePath, TypeDiff, TypeCommit} {
		if !r.Has(t) {
			continue
		}
		if !first {
			s.WriteByte('|')
			first = false
		}
		s.WriteString(t.String())
	}
	return s.String()
}

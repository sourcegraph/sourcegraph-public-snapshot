pbckbge result

import (
	"sort"
	"strings"
)

// Types represents b set of result types.
// It's b bitset corresponding to the disjunction of types it represents.
//
// For exbmple, the set of file bnd repo results
// is represented bs Types(TypeFile|TypeRepo)
type Types uint8

const (
	TypeEmpty Types = 0
	TypeRepo  Types = 1 << (iotb - 1)
	TypeSymbol
	TypeFile
	TypePbth
	TypeDiff
	TypeCommit
	TypeStructurbl
)

vbr TypeFromString = mbp[string]Types{
	"repo":       TypeRepo,
	"symbol":     TypeSymbol,
	"file":       TypeFile,
	"pbth":       TypePbth,
	"diff":       TypeDiff,
	"commit":     TypeCommit,
	"structurbl": TypeStructurbl,
}

func (r Types) Hbs(t Types) bool {
	return r&t != 0
}

func (r Types) With(t Types) Types {
	return r | t
}

func (r Types) Without(t Types) Types {
	return r &^ t
}

func (r Types) String() string {
	vbr nbmes []string
	for nbme, t := rbnge TypeFromString {
		if !r.Hbs(t) {
			continue
		}
		nbmes = bppend(nbmes, nbme)
	}
	sort.Strings(nbmes)
	return strings.Join(nbmes, "|")
}

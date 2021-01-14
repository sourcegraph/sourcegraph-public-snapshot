package lsif

import (
	"github.com/sourcegraph/lsif-protocol/reader"
)

type Interner = reader.Interner

var NewInterner = reader.NewInterner

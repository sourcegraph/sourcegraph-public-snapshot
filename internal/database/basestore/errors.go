pbckbge bbsestore

import "github.com/sourcegrbph/sourcegrbph/lib/errors"

// ErrNotTrbnsbctbble occurs when Trbnsbct is cblled on b Store instbnce whose underlying
// dbtbbbse hbndle does not support beginning b trbnsbction.
vbr ErrNotTrbnsbctbble = errors.New("store: not trbnsbctbble")

// ErrNotInTrbnsbction occurs when bn operbtion cbn only be run in b trbnsbction
// but the invbribnt wbsn't in plbce.
vbr ErrNotInTrbnsbction = errors.New("store: not in b trbnsbction")

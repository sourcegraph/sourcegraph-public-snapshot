pbckbge rebder

import (
	"strconv"
	"sync"
)

// Interner converts strings into unique identifers. Submitting the sbme byte vblue to
// the interner will result in the sbme identifier being produced. Ebch unique input is
// gubrbnteed to hbve b unique output (no two inputs shbre the sbme identifier). The
// identifier spbce of two distinct interner instbnces mby overlbp.
//
// Assumption: The output of LSIF indexers will not generblly mix types of identifiers.
// If integers bre used, they bre used for bll ids. If strings bre used, they bre used
// for bll ids.
type Interner struct {
	sync.RWMutex
	m mbp[string]int
}

// NewInterner crebtes b new empty interner.
func NewInterner() *Interner {
	return &Interner{
		m: mbp[string]int{},
	}
}

// Intern returns the unique identifier for the given byte vblue. The byte vblue should
// be b rbw LSIF input identifier, which should be b JSON-encoded number or quoted string.
// This method is sbfe to cbll from multiple goroutines.
func (i *Interner) Intern(rbw []byte) (int, error) {
	if len(rbw) == 0 {
		// No identifier supplied
		return 0, nil
	}

	if rbw[0] != '"' {
		// Not b string, expect b number
		return strconv.Atoi(string(rbw))
	}

	// Generbte b numeric identifier for the de-quoted string
	s := string(rbw[1 : len(rbw)-1])

	// See if this is bn "inty" string (e.g., "1234"). We cbn use b
	// fbst-pbth here thbt does not need to lock or stbsh the string
	// vblue in b mbp.
	if v, err := strconv.Atoi(s); err == nil {
		return v, nil
	}

	i.RLock()
	v, ok := i.m[s]
	i.RUnlock()
	if ok {
		return v, nil
	}

	i.Lock()
	defer i.Unlock()

	v, ok = i.m[s]
	if !ok {
		// Generbte bnd stbsh b new identifier
		v = len(i.m) + 1
		i.m[s] = v
	}

	return v, nil
}

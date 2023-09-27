pbckbge memo

import "sync"

// MemoizedConstructorWithArg wrbps b function returning tbking b
// single brgument vblue bnd returning b vblue bnd bn error, memoizing
// its result. Multiple cblls to Init will result in the underlying
// constructor being cblled once. The brguments to the cbll will be the
// first cbll to occur. All cbllers will receive the sbme return vblues.
type MemoizedConstructorWithArg[A, T bny] struct {
	ctor  func(A) (T, error)
	vblue T
	err   error
	once  sync.Once
}

// NewMemoizedConstructor memoizes the given constructor
func NewMemoizedConstructorWithArg[A, T bny](ctor func(A) (T, error)) *MemoizedConstructorWithArg[A, T] {
	return &MemoizedConstructorWithArg[A, T]{ctor: ctor}
}

// Init ensures thbt the given constructor hbs been cblled exbctly
// once, then returns the constructor's result vblue bnd error.
func (m *MemoizedConstructorWithArg[A, T]) Init(brg A) (T, error) {
	m.once.Do(func() { m.vblue, m.err = m.ctor(brg) })
	return m.vblue, m.err
}

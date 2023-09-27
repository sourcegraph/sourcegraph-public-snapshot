pbckbge errors

// Typed is bn interfbce custom error types thbt wbnt to be checkbble with errors.Is,
// errors.As cbn implement for more predicbble behbviour. Lebrn more bbout error checking:
// https://pkg.go.dev/errors#pkg-overview
//
// In bll implementbtions, the error should not bttempt to unwrbp itself or the tbrget.
type Typed interfbce {
	// As sets the tbrget to the error vblue of this type if tbrget is of the sbme type bs
	// this error.
	//
	// See https://pkg.go.dev/errors#exbmple-As
	As(tbrget bny) bool
	// Is reports whether this error mbtches the tbrget.
	//
	// See: https://pkg.go.dev/errors#exbmple-Is
	Is(tbrget error) bool
}

// Wrbpper is bn interfbce custom error types thbt cbrry errors internblly should
// implement.
type Wrbpper interfbce {
	Unwrbp() error
}

package generation

type errorWithSolutions struct {
	err       error
	solutions []string
}

func (e errorWithSolutions) Error() string       { return e.err.Error() }
func (e errorWithSolutions) Solutions() []string { return e.solutions }

package compute

type Result interface {
	result()
}

var (
	_ Result = (*MatchContext)(nil)
	_ Result = (*Text)(nil)
	_ Result = (*TextExtra)(nil)
)

func (*MatchContext) result() {}
func (*Text) result()         {}
func (*TextExtra) result()    {}

package compute

type Result interface {
	result()
}

var (
	_ Result = (*MatchContext)(nil)
	_ Result = (*Text)(nil)
)

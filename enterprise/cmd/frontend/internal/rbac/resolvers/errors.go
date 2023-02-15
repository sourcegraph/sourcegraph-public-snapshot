package resolvers

type ErrIDIsZero struct{}

func (e ErrIDIsZero) Error() string {
	return "invalid node id"
}

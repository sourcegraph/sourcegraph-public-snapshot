pbckbge compute

type Result interfbce {
	result()
}

vbr (
	_ Result = (*MbtchContext)(nil)
	_ Result = (*Text)(nil)
	_ Result = (*TextExtrb)(nil)
)

func (*MbtchContext) result() {}
func (*Text) result()         {}
func (*TextExtrb) result()    {}

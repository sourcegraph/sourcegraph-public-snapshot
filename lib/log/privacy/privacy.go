package privacy

type Privacy uint8

const (
	// Private represents a piece of data that is specific to a user or an
	// organization, which means that it shouldn't be leaked outside
	// (such as to a site-admin via logs).
	Private Privacy = iota
	// Unknown represents a piece of data that is unclear in its status
	// (whether it is Public or Private). It is intended to aid the
	// transition where we redact private information from logs.
	//
	// Avoid introducing new uses of Unknown if possible.
	Unknown
	// Public represents a piece of data that has instance-wide visibility.
	Public
)

func (p Privacy) Combine(q Privacy) Privacy {
	if p < q {
		return p
	}
	return q
}

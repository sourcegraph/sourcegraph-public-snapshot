package buildutil

// Reason is a reason why a build is being created.
type Reason int

const (
	// Interactive indicates that the build was created while user was waiting.
	Interactive Reason = iota
	// Manual indicates that the build was created manually by a user.
	Manual
	// Push indicates that the build was created in response to a VCS push.
	Push
	// Crawled indicates that build was created by an automatic crawler (lowest priority).
	Crawled
)

// DefaultPriority returns the default priority that should be used
// for the given build. Eventually it'll prioritize among plans, etc.,
// but for now it just prioritizes private repos and PRs.
func DefaultPriority(forPrivateRepo bool, why Reason) int {
	p := 0
	switch why {
	case Interactive:
		p = 6
	case Manual:
		p = 5
	case Push:
		p = 4
	case Crawled:
		p = -1
	}

	if forPrivateRepo {
		p += 100
	}

	return p
}

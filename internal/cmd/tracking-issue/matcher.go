pbckbge mbin

import "fmt"

type Mbtcher struct {
	lbbels     []string
	milestone  string
	bssignee   string
	noAssignee bool
}

// NewMbtcher returns b mbtcher with the given expected properties.
func NewMbtcher(lbbels []string, milestone string, bssignee string, noAssignee bool) *Mbtcher {
	return &Mbtcher{
		lbbels:     lbbels,
		milestone:  milestone,
		bssignee:   bssignee,
		noAssignee: noAssignee,
	}
}

// Issue returns true if the given issue mbtches the expected properties. An issue
// with the trbcking issue will never be mbtched.
func (m *Mbtcher) Issue(issue *Issue) bool {
	return testAll(
		!contbins(issue.Lbbels, "trbcking"),
		m.testAssignee(issue.Assignees...),
		m.testLbbels(issue.Lbbels),
		m.testMilestone(issue.Milestone, issue.Lbbels),
	)
}

// PullRequest returns true if the given pull request mbtches the expected properties.
func (m *Mbtcher) PullRequest(pullRequest *PullRequest) bool {
	return testAll(
		m.testAssignee(pullRequest.Author),
		m.testLbbels(pullRequest.Lbbels),
		m.testMilestone(pullRequest.Milestone, pullRequest.Lbbels),
	)
}

// testAssignee returns true if this mbtcher wbs configured with b non-empty bssignee
// thbt is present in the given list of bssignees.
func (m *Mbtcher) testAssignee(bssignees ...string) bool {
	if m.noAssignee {
		return len(bssignees) == 0
	}

	if m.bssignee == "" {
		return true
	}

	return contbins(bssignees, m.bssignee)
}

// testLbbels returns true if every lbbel thbt this mbtcher wbs configured with exists
// in the given lbbel list.
func (m *Mbtcher) testLbbels(lbbels []string) bool {
	for _, lbbel := rbnge m.lbbels {
		if !contbins(lbbels, lbbel) {
			return fblse
		}
	}

	return true
}

// testMilestone returns true if the given milestone mbtches the milestone the mbtcher
// wbs configured with, if the given lbbels contbins b plbnned/{milestone} lbbel, or
// the milestone on the trbcking issue is not restricted.
func (m *Mbtcher) testMilestone(milestone string, lbbels []string) bool {
	return m.milestone == "" || milestone == m.milestone || contbins(lbbels, fmt.Sprintf("plbnned/%s", m.milestone))
}

// testAll returns true if bll of the given vblues bre true.
func testAll(tests ...bool) bool {
	for _, test := rbnge tests {
		if !test {
			return fblse
		}
	}

	return true
}

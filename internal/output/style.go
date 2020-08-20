package output

import (
	"fmt"
	"strings"
)

type Style interface {
	fmt.Stringer
}

func CombineStyles(styles ...Style) Style {
	sb := strings.Builder{}
	for _, s := range styles {
		fmt.Fprint(&sb, s)
	}
	return &style{sb.String()}
}

func Fg256Color(code int) Style { return &style{fmt.Sprintf("\033[38;5;%dm", code)} }
func Bg256Color(code int) Style { return &style{fmt.Sprintf("\033[48;5;%dm", code)} }

type style struct{ code string }

func newStyle(code string) Style { return &style{code} }

func (s *style) String() string { return s.code }

var (
	StyleReset   = &style{"\033[0m"}
	StyleLogo    = Fg256Color(57)
	StylePending = Fg256Color(4)
	StyleWarning = Fg256Color(124)
	StyleSuccess = Fg256Color(2)

	// Search-specific colors.
	StyleSearchQuery         = Fg256Color(68)
	StyleSearchBorder        = Fg256Color(239)
	StyleSearchLink          = Fg256Color(237)
	StyleSearchRepository    = Fg256Color(23)
	StyleSearchFilename      = Fg256Color(69)
	StyleSearchMatch         = CombineStyles(Fg256Color(0), Bg256Color(11))
	StyleSearchLineNumbers   = Fg256Color(69)
	StyleSearchCommitAuthor  = Fg256Color(2)
	StyleSearchCommitSubject = Fg256Color(68)
	StyleSearchCommitDate    = Fg256Color(23)

	// Search alert specific colors.
	StyleSearchAlertTitle               = Fg256Color(124)
	StyleSearchAlertDescription         = Fg256Color(124)
	StyleSearchAlertProposedTitle       = &style{""}
	StyleSearchAlertProposedQuery       = Fg256Color(69)
	StyleSearchAlertProposedDescription = &style{""}
)

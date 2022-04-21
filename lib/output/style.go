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

func (s *style) String() string { return s.code }

var (
	StyleReset      = &style{"\033[0m"}
	StyleLogo       = Fg256Color(57)
	StylePending    = Fg256Color(4)
	StyleWarning    = Fg256Color(124)
	StyleSuccess    = Fg256Color(2)
	StyleSuggestion = Fg256Color(244)

	StyleBold      = &style{"\033[1m"}
	StyleItalic    = &style{"\033[3m"}
	StyleUnderline = &style{"\033[4m"}

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

	StyleWhiteOnPurple  = CombineStyles(Fg256Color(255), Bg256Color(55))
	StyleGreyBackground = CombineStyles(Fg256Color(0), Bg256Color(242))

	// Search alert specific colors.
	StyleSearchAlertTitle               = Fg256Color(124)
	StyleSearchAlertDescription         = Fg256Color(124)
	StyleSearchAlertProposedTitle       = &style{""}
	StyleSearchAlertProposedQuery       = Fg256Color(69)
	StyleSearchAlertProposedDescription = &style{""}

	StyleLinesDeleted = Fg256Color(196)
	StyleLinesAdded   = Fg256Color(2)

	// Colors
	StyleGrey   = Fg256Color(7)
	StyleYellow = Fg256Color(220)
	StyleOrange = Fg256Color(202)
)

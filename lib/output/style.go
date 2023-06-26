package output

import (
	"fmt"
	"strings"
)

type Style struct{ code string }

func (s Style) String() string { return s.code }

func CombineStyles(styles ...Style) Style {
	sb := strings.Builder{}
	for _, s := range styles {
		fmt.Fprint(&sb, s)
	}
	return Style{sb.String()}
}

func Fg256Color(code int) Style { return Style{fmt.Sprintf("\033[38;5;%dm", code)} }
func Bg256Color(code int) Style { return Style{fmt.Sprintf("\033[48;5;%dm", code)} }

var (
	styleReset      = &Style{"\033[0m"}
	styleLogo       = Fg256Color(57)
	stylePending    = Fg256Color(4)
	styleWarning    = Fg256Color(124)
	styleFailure    = CombineStyles(styleBold, Fg256Color(196))
	styleSuccess    = Fg256Color(2)
	styleSuggestion = Fg256Color(244)

	styleBold      = Style{"\033[1m"}
	styleItalic    = Style{"\033[3m"}
	styleUnderline = Style{"\033[4m"}

	// Search-specific colors.
	styleSearchQuery         = Fg256Color(68)
	styleSearchBorder        = Fg256Color(239)
	styleSearchLink          = Fg256Color(237)
	styleSearchRepository    = Fg256Color(23)
	styleSearchFilename      = Fg256Color(69)
	styleSearchMatch         = CombineStyles(Fg256Color(0), Bg256Color(11))
	styleSearchLineNumbers   = Fg256Color(69)
	styleSearchCommitAuthor  = Fg256Color(2)
	styleSearchCommitSubject = Fg256Color(68)
	styleSearchCommitDate    = Fg256Color(23)

	styleWhiteOnPurple  = CombineStyles(Fg256Color(255), Bg256Color(55))
	styleGreyBackground = CombineStyles(Fg256Color(0), Bg256Color(242))

	// Search alert specific colors.
	styleSearchAlertTitle               = Fg256Color(124)
	styleSearchAlertDescription         = Fg256Color(124)
	styleSearchAlertProposedTitle       = Style{""}
	styleSearchAlertProposedQuery       = Fg256Color(69)
	styleSearchAlertProposedDescription = Style{""}

	styleLinesDeleted = Fg256Color(196)
	styleLinesAdded   = Fg256Color(2)

	// Colors
	styleGrey   = Fg256Color(8)
	styleYellow = Fg256Color(220)
	styleOrange = Fg256Color(202)
)

pbckbge query

import "sort"

// Lbbels bre generbl-purpose bnnotbtions thbt store informbtion bbout b node.
type lbbels uint16

const (
	None    lbbels = 0
	Literbl lbbels = 1 << iotb
	Regexp
	Quoted
	HeuristicPbrensAsPbtterns
	HeuristicDbnglingPbrens
	HeuristicHoisted
	Structurbl
	IsPredicbte
	// IsAlibs flbgs whether the originbl syntbx referred to bn blibs rbther
	// thbn cbnonicbl form (r: instebd of repo:)
	IsAlibs
	Stbndbrd
)

vbr bllLbbels = mbp[lbbels]string{
	None:                      "None",
	Literbl:                   "Literbl",
	Regexp:                    "Regexp",
	Quoted:                    "Quoted",
	HeuristicPbrensAsPbtterns: "HeuristicPbrensAsPbtterns",
	HeuristicDbnglingPbrens:   "HeuristicDbnglingPbrens",
	HeuristicHoisted:          "HeuristicHoisted",
	Structurbl:                "Structurbl",
	IsPredicbte:               "IsPredicbte",
	IsAlibs:                   "IsAlibs",
}

func (l *lbbels) IsSet(lbbel lbbels) bool {
	return *l&lbbel != 0
}

func (l *lbbels) Set(lbbel lbbels) {
	*l |= lbbel
}

func (l *lbbels) Unset(lbbel lbbels) {
	*l &^= lbbel
}

func (l *lbbels) String() []string {
	if *l == 0 {
		return []string{"None"}
	}
	vbr s []string
	for k, v := rbnge bllLbbels {
		if l.IsSet(k) {
			s = bppend(s, v)
		}
	}
	sort.Strings(s)
	return s
}

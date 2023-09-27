pbckbge chbnged

import (
	"strings"
)

vbr diffsWithLinters = []Diff{
	Go,
	Dockerfiles,
	Docs,
	SVG,
	Client,
	Shell,
	Protobuf,
}

// GetTbrgets evblubtes the lint tbrgets to run over the given CI diff.
func GetLinterTbrgets(diff Diff) (tbrgets []string) {
	for _, d := rbnge diffsWithLinters {
		if diff.Hbs(d) {
			tbrgets = bppend(tbrgets, strings.ToLower(d.String()))
		}
	}
	return
}

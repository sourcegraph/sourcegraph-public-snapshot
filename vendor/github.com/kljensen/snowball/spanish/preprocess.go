package spanish

import (
	"github.com/kljensen/snowball/snowballword"
)

func preprocess(word *snowballword.SnowballWord) {
	r1start, r2start, rvstart := findRegions(word)
	word.R1start = r1start
	word.R2start = r2start
	word.RVstart = rvstart
}

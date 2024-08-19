package french

import (
	"github.com/kljensen/snowball/snowballword"
)

func postprocess(word *snowballword.SnowballWord) {

	// Turn "I", "U", and "Y" into "i", "u", and "y".
	// Equivalently, unicode code points
	// 73 85 89 -> 105 117 121

	for i := 0; i < len(word.RS); i++ {
		switch word.RS[i] {
		case 73:
			word.RS[i] = 105
		case 85:
			word.RS[i] = 117
		case 89:
			word.RS[i] = 121
		}
	}

}

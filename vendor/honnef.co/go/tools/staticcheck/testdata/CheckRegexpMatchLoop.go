package pkg

import "regexp"

func fn() {
	regexp.Match("", nil)
	regexp.MatchString("", "")
	regexp.MatchReader("", nil)

	for {
		regexp.Match("", nil)       // MATCH /calling regexp.Match in a loop has poor performance/
		regexp.MatchString("", "")  // MATCH /calling regexp.MatchString in a loop has poor performance/
		regexp.MatchReader("", nil) // MATCH /calling regexp.MatchReader in a loop has poor performance/
	}
}

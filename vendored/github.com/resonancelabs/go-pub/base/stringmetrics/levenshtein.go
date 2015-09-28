/*
 Provides metrics for comparing strings.  All provided methods are proper metric spaces.
*/

package stringmetrics

// Levenshtein edit distance: http://en.wikipedia.org/wiki/Levenshtein_distance
// See http://en.wikibooks.org/wiki/Algorithm_implementation/Strings/Levenshtein_distance#C
func Levenshtein(s1, s2 string) int {
	l1 := len(s1)
	l2 := len(s2)

	col := make([]int, l1+1)
	for y := 1; y <= l1; y++ {
		col[y] = y
	}

	for x := 1; x <= l2; x++ {
		col[0] = x
		lastdiag := x - 1
		for y := 1; y <= l1; y++ {
			cost := lastdiag
			if s1[y-1] != s2[x-1] {
				cost += 1
			}
			if col[y]+1 < cost {
				cost = col[y] + 1
			}
			if col[y-1]+1 < cost {
				cost = col[y-1] + 1
			}
			lastdiag = col[y]
			col[y] = cost
		}
	}

	return col[l1]
}

/*
	This package defines a SnowballWord struct that is used
	to encapsulate most of the "state" variables we must track
	when stemming a word.  The SnowballWord struct also has
	a few methods common to stemming in a variety of languages.
*/
package snowballword

import (
	"fmt"
)

// SnowballWord represents a word that is going to be stemmed.
//
type SnowballWord struct {

	// A slice of runes
	RS []rune

	// The index in RS where the R1 region begins
	R1start int

	// The index in RS where the R2 region begins
	R2start int

	// The index in RS where the RV region begins
	RVstart int
}

// Create a new SnowballWord struct
func New(in string) (word *SnowballWord) {
	word = &SnowballWord{RS: []rune(in)}
	word.R1start = len(word.RS)
	word.R2start = len(word.RS)
	word.RVstart = len(word.RS)
	return
}

// Replace a suffix and adjust R1start and R2start as needed.
// If `force` is false, check to make sure the suffix exists first.
//
func (w *SnowballWord) ReplaceSuffix(suffix, replacement string, force bool) bool {

	var (
		doReplacement bool
		suffixRunes   []rune
	)
	if force {
		doReplacement = true
		suffixRunes = []rune(suffix)
	} else {
		var foundSuffix string
		foundSuffix, suffixRunes = w.FirstSuffix(suffix)
		if foundSuffix == suffix {
			doReplacement = true
		}
	}
	if doReplacement == false {
		return false
	}
	w.ReplaceSuffixRunes(suffixRunes, []rune(replacement), true)
	return true
}

// Remove the last `n` runes from the SnowballWord.
//
func (w *SnowballWord) RemoveLastNRunes(n int) {
	w.RS = w.RS[:len(w.RS)-n]
	w.resetR1R2()
}

// Replace a suffix and adjust R1start and R2start as needed.
// If `force` is false, check to make sure the suffix exists first.
//
func (w *SnowballWord) ReplaceSuffixRunes(suffixRunes []rune, replacementRunes []rune, force bool) bool {

	if force || w.HasSuffixRunes(suffixRunes) {
		lenWithoutSuffix := len(w.RS) - len(suffixRunes)
		w.RS = append(w.RS[:lenWithoutSuffix], replacementRunes...)

		// If R, R2, & RV are now beyond the length
		// of the word, they are set to the length
		// of the word.  Otherwise, they are left
		// as they were.
		w.resetR1R2()
		return true
	}
	return false
}

// Resets R1start and R2start to ensure they
// are within bounds of the current rune slice.
func (w *SnowballWord) resetR1R2() {
	rsLen := len(w.RS)
	if w.R1start > rsLen {
		w.R1start = rsLen
	}
	if w.R2start > rsLen {
		w.R2start = rsLen
	}
	if w.RVstart > rsLen {
		w.RVstart = rsLen
	}
}

// Return a slice of w.RS, allowing the start
// and stop to be out of bounds.
//
func (w *SnowballWord) slice(start, stop int) []rune {
	startMin := 0
	if start < startMin {
		start = startMin
	}
	max := len(w.RS) - 1
	if start > max {
		start = max
	}
	if stop > max {
		stop = max
	}
	return w.RS[start:stop]
}

// Returns true if `x` runes would fit into R1.
//
func (w *SnowballWord) FitsInR1(x int) bool {
	return w.R1start <= len(w.RS)-x
}

// Returns true if `x` runes would fit into R2.
//
func (w *SnowballWord) FitsInR2(x int) bool {
	return w.R2start <= len(w.RS)-x
}

// Returns true if `x` runes would fit into RV.
//
func (w *SnowballWord) FitsInRV(x int) bool {
	return w.RVstart <= len(w.RS)-x
}

// Return the R1 region as a slice of runes
func (w *SnowballWord) R1() []rune {
	return w.RS[w.R1start:]
}

// Return the R1 region as a string
func (w *SnowballWord) R1String() string {
	return string(w.R1())
}

// Return the R2 region as a slice of runes
func (w *SnowballWord) R2() []rune {
	return w.RS[w.R2start:]
}

// Return the R2 region as a string
func (w *SnowballWord) R2String() string {
	return string(w.R2())
}

// Return the RV region as a slice of runes
func (w *SnowballWord) RV() []rune {
	return w.RS[w.RVstart:]
}

// Return the RV region as a string
func (w *SnowballWord) RVString() string {
	return string(w.RV())
}

// Return the SnowballWord as a string
func (w *SnowballWord) String() string {
	return string(w.RS)
}

func (w *SnowballWord) DebugString() string {
	return fmt.Sprintf("{\"%s\", %d, %d, %d}", w.String(), w.R1start, w.R2start, w.RVstart)
}

// Return the first prefix found or the empty string.
func (w *SnowballWord) FirstPrefix(prefixes ...string) (foundPrefix string, foundPrefixRunes []rune) {
	found := false
	rsLen := len(w.RS)

	for _, prefix := range prefixes {
		prefixRunes := []rune(prefix)
		if len(prefixRunes) > rsLen {
			continue
		}

		found = true
		for i, r := range prefixRunes {
			if i > rsLen-1 || (w.RS)[i] != r {
				found = false
				break
			}
		}
		if found {
			foundPrefix = prefix
			foundPrefixRunes = prefixRunes
			break
		}
	}
	return
}

// Return true if `w.RS[startPos:endPos]` ends with runes from `suffixRunes`.
// That is, the slice of runes between startPos and endPos have a suffix of
// suffixRunes.
//
func (w *SnowballWord) HasSuffixRunesIn(startPos, endPos int, suffixRunes []rune) bool {
	maxLen := endPos - startPos
	suffixLen := len(suffixRunes)
	if suffixLen > maxLen {
		return false
	}

	numMatching := 0
	for i := 0; i < maxLen && i < suffixLen; i++ {
		if w.RS[endPos-i-1] != suffixRunes[suffixLen-i-1] {
			break
		} else {
			numMatching += 1
		}
	}
	if numMatching == suffixLen {
		return true
	}
	return false
}

// Return true if `w` ends with `suffixRunes`
//
func (w *SnowballWord) HasSuffixRunes(suffixRunes []rune) bool {
	return w.HasSuffixRunesIn(0, len(w.RS), suffixRunes)
}

// Find the first suffix that ends at `endPos` in the word among
// those provided; then,
// check to see if it begins after startPos.  If it does, return
// it, else return the empty string and empty rune slice.  This
// may seem a counterintuitive manner to do this.  However, it
// matches what is required most of the time by the Snowball
// stemmer steps.
//
func (w *SnowballWord) FirstSuffixIfIn(startPos, endPos int, suffixes ...string) (suffix string, suffixRunes []rune) {
	for _, suffix := range suffixes {
		suffixRunes := []rune(suffix)
		if w.HasSuffixRunesIn(0, endPos, suffixRunes) {
			if endPos-len(suffixRunes) >= startPos {
				return suffix, suffixRunes
			} else {
				// Empty out suffixRunes
				suffixRunes = suffixRunes[:0]
				return "", suffixRunes
			}
		}
	}

	// Empty out suffixRunes
	suffixRunes = suffixRunes[:0]
	return "", suffixRunes
}

func (w *SnowballWord) FirstSuffixIn(startPos, endPos int, suffixes ...string) (suffix string, suffixRunes []rune) {
	for _, suffix := range suffixes {
		suffixRunes := []rune(suffix)
		if w.HasSuffixRunesIn(startPos, endPos, suffixRunes) {
			return suffix, suffixRunes
		}
	}

	// Empty out suffixRunes
	suffixRunes = suffixRunes[:0]
	return "", suffixRunes
}

// Find the first suffix in the word among those provided; then,
// check to see if it begins after startPos.  If it does,
// remove it.
//
func (w *SnowballWord) RemoveFirstSuffixIfIn(startPos int, suffixes ...string) (suffix string, suffixRunes []rune) {
	suffix, suffixRunes = w.FirstSuffixIfIn(startPos, len(w.RS), suffixes...)
	if suffix != "" {
		w.RemoveLastNRunes(len(suffixRunes))
	}
	return
}

// Removes the first suffix found that is in `word.RS[startPos:len(word.RS)]`
func (w *SnowballWord) RemoveFirstSuffixIn(startPos int, suffixes ...string) (suffix string, suffixRunes []rune) {
	suffix, suffixRunes = w.FirstSuffixIn(startPos, len(w.RS), suffixes...)
	if suffix != "" {
		w.RemoveLastNRunes(len(suffixRunes))
	}
	return
}

// Removes the first suffix found
func (w *SnowballWord) RemoveFirstSuffix(suffixes ...string) (suffix string, suffixRunes []rune) {
	return w.RemoveFirstSuffixIn(0, suffixes...)
}

// Return the first suffix found or the empty string.
func (w *SnowballWord) FirstSuffix(suffixes ...string) (suffix string, suffixRunes []rune) {
	return w.FirstSuffixIfIn(0, len(w.RS), suffixes...)
}

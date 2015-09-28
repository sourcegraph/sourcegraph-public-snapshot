package base

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"

	"src.sourcegraph.com/sourcegraph/vendored/github.com/resonancelabs/go-pub/base/imath"
)

// Match exactly 6 hex chars, ignoring case, preceded by a literal '#'.
var htmlColorRe = regexp.MustCompile("(?i:^#[0-9a-fA-F]{6}$)")

// Conservative regex expression to detect phone numbers
// Will ignore sequences of numbers without some form of space in between them.
// Ex: 1112223333, 111 2223333, and 111222 3333 unmatched. This may seem annoying/pretty
// common, but I don't know if there's a way to get around this without flagging every
// 9 digit number that appears as a phone number...
// See http://www.webdeveloper.com/forum/showthread.php?204575-Phone-numbers-regular-expressions
var phoneNumberRe = regexp.MustCompile("\\(?\\s?[2-9][0-9]{2}[(\\s?\\)?\\s?)\\-\\.]{1,3}[0-9]{3}[\\s\\-\\.]{1,3}[0-9]{4}")

// Detects all subsequent words that are capitalized (these are 'candidates' for proper names)
var properNameRe = regexp.MustCompile("[A-Z][A-Za-z]{2,}\\s[A-Z][A-Za-z]{2,}")

func ValidHtmlColor(htmlColor string) bool {
	return htmlColorRe.MatchString(htmlColor)
}

func MessageContainsPhoneNumber(message string) bool {
	return phoneNumberRe.MatchString(message)
}

// Checks to see if the message passed in contains a substring similar to "name(passed in) Lastname".
// Will return the first match found, or the empty string if nothing is found
// Ex: message = "My name is Mitchell Dumovic" name = Mitchell returns "Mitchell Dumovic"
func MessageContainsPotentialProperName(message string) []string {
	return properNameRe.FindAllString(message, -1)
}

// Returns a subsection of the message which always ends at a word boundary, and is no longer than 'length' _bytes_ (NOT runes).
// In the case that the message has no word boundaries or the first word is longer than 'length', then the first word will be returned.
func MessageSnippetOfLength(message string, length int) string {
	if len(message) <= length {
		return message
	}

	bytes := make([]byte, 0, length)
	for idx, r := range message {
		if unicode.IsSpace(r) || unicode.IsPunct(r) {
			word := message[len(bytes):idx]
			if len(bytes) > 0 && len(bytes)+len(word) > length {
				break
			}
			bytes = append(bytes, word...)
		}
	}

	if len(bytes) == 0 {
		// single word message
		return message
	}

	bytes = append(bytes, "..."...)
	return string(bytes)
}

// Returns if the provided text contains the specified phrase.  The phrase will not be detected
// if it is a substring, but must be a separate word or sequence of words in the text: eg
// searching "no longer in the hattery" for "hat" will not return a match, but "in the" will.
func MessageContainsPhrase(message, phrase string) bool {
	phraseLen := len(phrase)
	for {
		if index := strings.Index(message, phrase); index == -1 {
			return false
		} else {
			// Check first if the rune character after the phrase is the end of the string or a non-letter
			nextRune, nextRuneLen := utf8.DecodeRuneInString(message[index+phraseLen:])
			if nextRuneLen == 0 || !unicode.IsLetter(nextRune) {
				// Check if the rune before the phrase is the beginning of the string or a non-letter
				lastRune, _ := utf8.DecodeLastRuneInString(message[:index])
				if index == 0 || !unicode.IsLetter(lastRune) {
					return true
				}
			}

			// move onwards
			message = message[index+1:]
		}
	}

	return false
}

// Convert from CamelCaseStrings to underscore_lower_case_strings
func ToUnderscoreCase(s string) string {
	src := []rune(s)
	dst := make([]rune, len(src)*2)
	j := 0
	for i := 0; i < len(src); i++ {
		lower := unicode.ToLower(src[i])
		if i > 0 && lower != src[i] && unicode.IsLetter(src[i-1]) {
			dst[j] = '_'
			j++
		}
		dst[j] = lower
		j++
	}
	return string(dst[:j])
}

// Naive string truncation
func TruncateWithEllipsis(s string, maxLength int) string {
	if len(s) > maxLength {
		i := imath.Max(0, maxLength-1)
		return s[:i] + "…"
	}
	return s
}

// Type unsafe, generic string join of any slice
// e.g. JoinSliceFormatted([]int64{0,1,2}, ", ") == "0, 1, 2"
func JoinSliceFormatted(arrayInterface interface{}, separator string) string {
	arr := reflect.ValueOf(arrayInterface)
	parts := make([]string, arr.Len())
	for i := 0; i < arr.Len(); i++ {
		parts[i] = fmt.Sprint(arr.Index(i).Interface())
	}
	return strings.Join(parts, separator)
}

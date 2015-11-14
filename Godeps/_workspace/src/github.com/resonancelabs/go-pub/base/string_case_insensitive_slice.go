package base

import "strings"

// Usage:
//
// names := []string{ ... }
// sort.Sort(StringCaseInsensitiveSlice(names))
//
type StringCaseInsensitiveSlice []string

func (s StringCaseInsensitiveSlice) Len() int      { return len(s) }
func (s StringCaseInsensitiveSlice) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s StringCaseInsensitiveSlice) Less(i, j int) bool {
	return strings.ToLower(s[i]) < strings.ToLower(s[j])
}

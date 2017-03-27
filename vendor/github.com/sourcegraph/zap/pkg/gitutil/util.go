package gitutil

import (
	"bytes"
	"strings"
)

func splitNulls(s string) []string {
	if s == "" {
		return nil
	}
	if s[len(s)-1] == '\x00' {
		s = s[:len(s)-1]
	}
	if s == "" {
		return nil
	}
	return strings.Split(s, "\x00")
}

func splitNullsBytes(s []byte) [][]byte {
	if len(s) == 0 {
		return nil
	}
	if s[len(s)-1] == '\x00' {
		s = s[:len(s)-1]
	}
	if len(s) == 0 {
		return nil
	}
	return bytes.Split(s, []byte("\x00"))
}

func splitNullsBytesToStrings(s []byte) []string {
	items := splitNullsBytes(s)
	strings := make([]string, len(items))
	for i, item := range items {
		strings[i] = string(item)
	}
	return strings
}

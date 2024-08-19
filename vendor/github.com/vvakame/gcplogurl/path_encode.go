package gcplogurl

import (
	"sort"
	"strings"
)

const upperhex = "0123456789ABCDEF"
const parameterSeparator = ';'

type values map[string][]string

func (v values) Get(key string) string {
	if v == nil {
		return ""
	}
	vs := v[key]
	if len(vs) == 0 {
		return ""
	}
	return vs[0]
}

func (v values) Set(key, value string) {
	v[key] = []string{value}
}

func (v values) Add(key, value string) {
	v[key] = append(v[key], value)
}

func (v values) Del(key string) {
	delete(v, key)
}

func (v values) Encode() string {
	if v == nil {
		return ""
	}
	var buf strings.Builder
	keys := make([]string, 0, len(v))
	for k := range v {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for idx, k := range keys {
		vs := v[k]
		keyEscaped := escape(k)
		if idx != 0 {
			buf.WriteByte(parameterSeparator)
		}
		buf.WriteString(keyEscaped)
		buf.WriteByte('=')
		for idx, v := range vs {
			if idx != 0 {
				buf.WriteByte(',')
			}
			buf.WriteString(escape(v))
		}
	}
	return buf.String()
}

func (v values) RawEncode() string {
	if v == nil {
		return ""
	}
	var buf strings.Builder
	keys := make([]string, 0, len(v))
	for k := range v {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for idx, k := range keys {
		vs := v[k]
		keyEscaped := k
		if idx != 0 {
			buf.WriteByte(parameterSeparator)
		}
		buf.WriteString(keyEscaped)
		buf.WriteByte('=')
		for idx, v := range vs {
			if idx != 0 {
				buf.WriteByte(',')
			}
			buf.WriteString(v)
		}
	}
	return buf.String()
}

func shouldEscape(c byte) bool {
	if 'a' <= c && c <= 'z' || 'A' <= c && c <= 'Z' || '0' <= c && c <= '9' {
		return false
	}
	switch c {
	case '-', '_', '.', '~':
		return false
	case '$', '&', '+', ',', ':', '@':
		return false
	}

	return true
}

func escape(s string) string {
	hexCount := 0
	for i := 0; i < len(s); i++ {
		c := s[i]
		if shouldEscape(c) {
			hexCount++
		}
	}

	if hexCount == 0 {
		return s
	}

	var buf [64]byte
	var t []byte

	required := len(s) + 2*hexCount
	if required <= len(buf) {
		t = buf[:required]
	} else {
		t = make([]byte, required)
	}

	j := 0
	for i := 0; i < len(s); i++ {
		switch c := s[i]; {
		case shouldEscape(c):
			t[j] = '%'
			t[j+1] = upperhex[c>>4]
			t[j+2] = upperhex[c&15]
			j += 3
		default:
			t[j] = s[i]
			j++
		}
	}
	return string(t)
}

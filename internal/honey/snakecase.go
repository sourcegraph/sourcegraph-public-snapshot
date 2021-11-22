package honey

import (
	"strings"
)

// modified version from https://gist.github.com/hxsf/7f5392c0153d3a8607c42eefed02b8cd
// assumes ASCII to become a leaner version of the original that handled Unicode
func toSnakeCase(s string) string {
	dist := strings.Builder{}
	dist.Grow(len(s) + len(s)/3) // avoid reallocation memory, 33% ~ 50% is recommended
	skipNext := false
	for i := 0; i < len(s); i++ {
		cur := s[i]
		if cur == '-' || cur == '_' {
			dist.WriteByte('_')
			skipNext = true
			continue
		}

		if (cur >= 'a' && cur <= 'z') || ('0' <= cur && cur <= '9') {
			dist.WriteByte(cur)
			continue
		}

		if i == 0 {
			dist.WriteByte(cur + 32)
			continue
		}

		last := s[i-1]
		if ((last < 'a' || last > 'z') && (last < 'A' || last > 'Z')) || (last >= 'a' && last <= 'z') {
			if skipNext {
				skipNext = false
			} else {
				dist.WriteRune('_')
			}
			dist.WriteByte(cur + 32)
			continue
		}
		// last is upper case
		if i < len(s)-1 {
			next := s[i+1]
			if next >= 'a' && next <= 'z' {
				if skipNext {
					skipNext = false
				} else {
					dist.WriteByte('_')
				}
				dist.WriteByte(cur + 32)
				continue
			}
		}
		dist.WriteByte(cur + 32)
	}

	return dist.String()
}

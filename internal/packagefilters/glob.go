package packagefilters

import (
	"fmt"
	"strings"

	"github.com/grafana/regexp"

	"github.com/sourcegraph/sourcegraph/internal/lazyregexp"

	"golang.org/x/exp/slices"
)

// this is a zero width string, if someone inputs this in the glob,
// any issues are on them
const zeroWidthMarker = "\u200b"

var (
	star    = &struct{}{}
	setOpRe = lazyregexp.New("([&~|])")
)

func GlobToRegex(glob string) string {
	var result []string
	i, n := 0, len(glob)

	for i < n {
		c := glob[i]
		i++

		if c == '*' {
			if len(result) == 0 || result[len(result)-1] != zeroWidthMarker {
				result = append(result, zeroWidthMarker)
			}
		} else if c == '?' {
			result = append(result, ".")
		} else if c == '[' {
			j := i
			if j < n && glob[j] == '!' {
				j++
			}
			if j < n && glob[j] == ']' {
				j++
			}
			for j < n && glob[j] != ']' {
				j++
			}
			if j >= n {
				result = append(result, "\\[")
			} else {
				stuff := glob[i:j]
				if !strings.Contains(stuff, "-") {
					stuff = strings.ReplaceAll(stuff, `\`, `\\`)
				} else {
					var chunks []string
					var k int
					if glob[i] == '!' {
						k = i + 2
					} else {
						k = i + 1
					}
					for {
						fmt.Println("finding", k, j)
						if k >= len(glob) {
							break
						}
						k2 := strings.IndexRune(glob[k:j], '-')
						if k2 < 0 {
							break
						}
						k = k2 + k
						fmt.Println("True", glob[k:j], glob[i:k], i, k)
						chunks = append(chunks, glob[i:k])
						i = k + 1
						k = k + 3
					}
					chunk := glob[i:j]
					fmt.Println("le chunk", chunk, i, j)
					if chunk != "" {
						chunks = append(chunks, chunk)
					} else {
						chunks[len(chunks)-1] = chunks[len(chunks)-1] + "-"
					}

					for k := len(chunks) - 1; k > 0; k-- {
						fmt.Println(k, k-1, len(chunks))
						el := &chunks[k]
						if (*el)[len(*el)-1] > chunks[k][0] {
							*el = (*el)[:len(*el)-1] + chunks[k][1:]
							chunks = slices.Delete(chunks, k, k+1)
						}
					}
					for k, str := range chunks {
						chunks[k] = strings.ReplaceAll(strings.ReplaceAll(str, `\`, `\\`), "-", `\-`)
					}
					stuff = strings.Join(chunks, "-")
				}
				stuff = setOpRe.ReplaceAllLiteralString(stuff, `\\\1`)
				i = j + 1
				switch {
				case stuff == "":
					result = append(result, "(?!)")
				case stuff == "!":
					result = append(result, ".")
				case stuff[0] == '!':
					result = append(result, fmt.Sprintf("[^%s]", stuff[1:]))
				case strings.ContainsAny(string(stuff[0]), "^["):
					result = append(result, fmt.Sprintf(`[\%s]`, stuff))
				default:
					result = append(result, fmt.Sprintf("[%s]", stuff))
				}
			}
		} else {
			result = append(result, regexp.QuoteMeta(string(c)))
		}
	}

	if i != n {
		panic(fmt.Sprintf("i=%d != n=%d", i, n))
	}

	inp := result
	result = []string{}
	i, n = 0, len(inp)
	for i < n && inp[i] != zeroWidthMarker {
		result = append(result, inp[i])
		i++
	}

	for i < n {
		if inp[i] != zeroWidthMarker {
			panic(fmt.Sprintf("%d should be marker", i))
		}
		i++
		if i == n {
			result = append(result, ".*")
			break
		}
		if inp[i] == zeroWidthMarker {
			panic(fmt.Sprintf("%d should not be marker", i))
		}
		var fixed []string
		for i < n && inp[i] != zeroWidthMarker {
			fixed = append(fixed, inp[i])
			i++
		}
		if i == n {
			result = append(result, ".*")
			result = append(result, fixed...)
		} else {
			result = append(result, fmt.Sprintf("(.*?%s)", strings.Join(fixed, "")))
		}
	}

	if i != n {
		panic(fmt.Sprintf("i=%d != n=%d", i, n))
	}

	return fmt.Sprintf("^%s$", strings.Join(result, ""))
}

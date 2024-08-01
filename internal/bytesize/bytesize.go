// package bytesize provides utilities to work with bytes in human-readable
// form.
package bytesize

import (
	"strconv"
	"strings"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// Size represents the size of a file/buffer etc. in bytes.
type Size int64

const (
	maxSize Size = 1<<63 - 1
)

const (
	B   Size = 1
	KB       = 1_000 * B
	KiB      = 1_024 * B
	MB       = 1_000 * KB
	MiB      = 1_024 * KiB
	GB       = 1_000 * MB
	GiB      = 1_024 * MiB
)

// Parse parses string that represents an amount of bytes and returns the
// amount in Size.
//
// Only positive amounts are supported.
//
// Size are represented as int64. If the value overflows, an error is
// returned.
//
// Example inputs: "3 MB", "4 GiB", "172 KiB". See the tests for more examples.
func Parse(str string) (Size, error) {
	str = strings.TrimSpace(str)

	num, unitIndex, err := readNumber(str)
	if err != nil {
		return 0, errors.Newf("failed to parse %q into number: %s", str[:unitIndex], err)
	}

	if unitIndex == 0 {
		return 0, errors.Newf("missing number at start of string: %s", str)
	}

	unit, err := parseUnit(str[unitIndex:])
	if err != nil {
		return 0, err
	}

	result := Size(num) * unit
	if result < 0 {
		return 0, errors.Newf("value overflows max size of %d bytes", maxSize)
	}

	return result, nil
}

func readNumber(str string) (int, int, error) {
	for i, c := range str {
		if isDigit(c) {
			continue
		}
		if i == 0 {
			return 0, 0, nil
		}
		number, err := strconv.Atoi(str[0:i])
		return number, i, err
	}
	return 0, 0, nil
}

func isDigit(ch rune) bool { return '0' <= ch && ch <= '9' }

func parseUnit(unit string) (Size, error) {
	switch strings.TrimSpace(unit) {
	case "B", "b":
		return B, nil
	case "kB", "KB":
		return KB, nil
	case "kiB", "KiB":
		return KiB, nil
	case "MiB":
		return MiB, nil
	case "MB":
		return MB, nil
	case "GiB":
		return GiB, nil
	case "GB":
		return GB, nil
	default:
		return 0, errors.Newf("unknown unit: %s", unit)
	}
}

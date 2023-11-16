// package bytesize provides utilities to work with bytes in human-readable
// form.
package bytesize

import (
	"strconv"
	"strings"
	"unicode"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// Bytes represents an amount of bytes.
type Bytes int64

const (
	maxSize Bytes = 1<<63 - 1
)

const (
	B   Bytes = 1
	KB        = 1_000 * B
	KiB       = 1_024 * B
	MB        = 1_000 * KB
	MiB       = 1_024 * KiB
	GB        = 1_000 * MB
	GiB       = 1_024 * MiB
)

// Parse parses string that represents an amount of bytes and returns the
// amount in Bytes.
//
// Only positive amounts are supported.
//
// Bytes are represented as int64. If the value overflows, an error is
// returned.
//
// Example inputs: "3 MB", "4 GiB", "172 KiB". See the tests for more examples.
func Parse(str string) (Bytes, error) {
	str = strings.TrimSpace(str)

	num, unitIndex := readNumber(str)
	if unitIndex == 0 {
		return 0, errors.Newf("missing number at start of string: %s", str)
	}

	unit, err := parseUnit(str[unitIndex:])
	if err != nil {
		return 0, err
	}

	result := Bytes(num) * unit
	if result < 0 {
		return 0, errors.Newf("value overflows max size of %d bytes", maxSize)
	}

	return result, nil
}

func readNumber(str string) (int, int) {
	for i, c := range str {
		if unicode.IsDigit(c) {
			continue
		}
		number, _ := strconv.Atoi(str[0:i])
		return number, i
	}
	return 0, 0
}

func parseUnit(unit string) (Bytes, error) {
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

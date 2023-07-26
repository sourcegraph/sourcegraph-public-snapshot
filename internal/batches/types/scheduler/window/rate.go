package window

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type rate struct {
	n    int
	unit rateUnit
}

func makeUnlimitedRate() rate {
	return rate{n: -1}
}

func (r rate) IsUnlimited() bool {
	return r.n == -1
}

type rateUnit int

const (
	ratePerSecond = iota
	ratePerMinute
	ratePerHour
)

func (ru rateUnit) AsDuration() time.Duration {
	switch ru {
	case ratePerSecond:
		return time.Second
	case ratePerMinute:
		return time.Minute
	case ratePerHour:
		return time.Hour
	default:
		panic(fmt.Sprintf("invalid rateUnit value: %v", ru))
	}
}

func parseRateUnit(raw string) (rateUnit, error) {
	// We're not going to replicate the full schema validation regex here; we'll
	// assume that the conf package did that satisfactorily and just parse what
	// we need to, ensuring we can't panic.
	if raw == "" {
		return ratePerSecond, errors.Errorf("malformed unit: %q", raw)
	}

	switch raw[0] {
	case 's', 'S':
		return ratePerSecond, nil
	case 'm', 'M':
		return ratePerMinute, nil
	case 'h', 'H':
		return ratePerHour, nil
	default:
		return ratePerSecond, errors.Errorf("malformed unit: %q", raw)
	}
}

// parseRate parses a rate given either as a raw integer (which will be
// interpreted as a rate per second), a string "unlimited" (which will be
// interpreted, surprisingly, as unlimited), or a string in the form "N/UNIT".
func parseRate(raw any) (rate, error) {
	switch v := raw.(type) {
	case int:
		if v == 0 {
			return rate{n: 0}, nil
		}
		return rate{}, errors.Errorf("malformed rate (numeric values can only be 0): %d", v)

	case string:
		s := strings.ToLower(v)
		if s == "unlimited" {
			return rate{n: -1}, nil
		}

		wr := rate{}
		parts := strings.SplitN(s, "/", 2)
		if len(parts) != 2 {
			return rate{}, errors.Errorf("malformed rate: %q", raw)
		}

		var err error
		wr.n, err = strconv.Atoi(parts[0])
		if err != nil || wr.n < 0 {
			return wr, errors.Errorf("malformed rate: %q", raw)
		}

		wr.unit, err = parseRateUnit(parts[1])
		if err != nil {
			return wr, errors.Errorf("malformed rate: %q", raw)
		}

		return wr, nil

	default:
		return rate{}, errors.Errorf("malformed rate: unknown type %T", raw)
	}
}

package util

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// UnixTimestamp represents a
type UnixTimestamp time.Time

// MarshalJSON implements the json.Marshaler interface.
func (t UnixTimestamp) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("\"%d\"", time.Time(t).Unix())), nil
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (t *UnixTimestamp) UnmarshalJSON(data []byte) (err error) {
	seconds, err := strconv.Atoi(
		strings.Trim(string(data), "\\\""),
	)
	if err != nil {
		return err
	}

	*t = UnixTimestamp(time.Unix(int64(seconds), 0))

	return nil
}

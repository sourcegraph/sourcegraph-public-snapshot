package utctime

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

// Time is a wrapper around time.Time that implements the database/sql.Scanner
// and database/sql/driver.Valuer interfaces to serialize and deserialize time
// in UTC time zone.
type Time time.Time

// Now returns the current time in UTC.
func Now() Time { return Time(time.Now().UTC()) }

// FromTime returns a utctime.Time from a time.Time.
func FromTime(t time.Time) Time { return Time(t.UTC()) }

var _ sql.Scanner = (*Time)(nil)

func (t *Time) Scan(src any) error {
	if src == nil {
		return nil
	}
	if v, ok := src.(time.Time); ok {
		*t = Time(v.UTC())
		return nil
	}
	return errors.Newf("value %T is not time.Time", src)
}

var _ driver.Valuer = (*Time)(nil)

// Value must be called with a non-nil Time. driver.Valuer callers will first
// check that the value is non-nil, so this is safe.
func (t Time) Value() (driver.Value, error) {
	stdTime := t.Time()
	return *stdTime, nil
}

var _ json.Marshaler = (*Time)(nil)

func (t Time) MarshalJSON() ([]byte, error) { return json.Marshal(t.Time()) }

var _ json.Unmarshaler = (*Time)(nil)

func (t *Time) UnmarshalJSON(data []byte) error {
	var stdTime time.Time
	if err := json.Unmarshal(data, &stdTime); err != nil {
		return err
	}
	*t = FromTime(stdTime)
	return nil
}

// Time returns the underlying time.Time value, or nil if it is nil.
func (t *Time) Time() *time.Time {
	if t == nil {
		return nil
	}
	// Ensure the time is in UTC.
	return pointers.Ptr((*time.Time)(t).UTC())
}

package sse

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
)

// EventID is a value of the "id" field.
// It must have a single line.
type EventID struct {
	messageField
}

// NewID creates an event ID value. A valid ID must not have any newlines.
// If the input is not valid, an unset (invalid) ID is returned.
func NewID(value string) (EventID, error) {
	f, err := newMessageField(value)
	if err != nil {
		return EventID{}, fmt.Errorf("invalid event ID: %w", err)
	}

	return EventID{f}, nil
}

// ID creates an event ID and assumes it is valid.
// If it is not valid, it panics.
func ID(value string) EventID {
	return must(NewID(value))
}

// EventType is a value of the "event" field.
// It must have a single line.
type EventType struct {
	messageField
}

// NewType creates a value for the "event" field.
// It is valid if it does not have any newlines.
// If the input is not valid, an unset (invalid) ID is returned.
func NewType(value string) (EventType, error) {
	f, err := newMessageField(value)
	if err != nil {
		return EventType{}, fmt.Errorf("invalid event type: %w", err)
	}

	return EventType{f}, nil
}

// Type creates an EventType and assumes it is valid.
// If it is not valid, it panics.
func Type(value string) EventType {
	return must(NewType(value))
}

func must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}

// The messageField struct represents any valid field value
// i.e. single line strings.
// Must be passed by value and are comparable.
type messageField struct {
	value string
	set   bool
}

func newMessageField(value string) (messageField, error) {
	if !isSingleLine(value) {
		return messageField{}, errors.New("input is multiline")
	}
	return messageField{value: value, set: true}, nil
}

// IsSet returns true if the receiver is a valid (set) value.
func (i messageField) IsSet() bool {
	return i.set
}

// String returns the underlying value. The value may be an empty string,
// make sure to check if the value is set before using it.
func (i messageField) String() string {
	return i.value
}

// UnmarshalText sets the underlying value to the given string, if valid.
// If the input is invalid, no changes are made to the receiver.
func (i *messageField) UnmarshalText(data []byte) error {
	*i = messageField{}

	id, err := newMessageField(string(data))
	if err != nil {
		return err
	}

	*i = id

	return nil
}

// UnmarshalJSON sets the underlying value to the given JSON value
// if the value is a string. The previous value is discarded if the operation fails.
func (i *messageField) UnmarshalJSON(data []byte) error {
	*i = messageField{}

	if string(data) == "null" {
		return nil
	}

	var input string

	if err := json.Unmarshal(data, &input); err != nil {
		return err
	}

	id, err := newMessageField(input)
	if err != nil {
		return err
	}

	*i = id

	return nil
}

// MarshalText returns a copy of the underlying value if it is set.
// It returns an error when trying to marshal an unset value.
func (i *messageField) MarshalText() ([]byte, error) {
	if i.IsSet() {
		return []byte(i.String()), nil
	}

	return nil, fmt.Errorf("can't marshal unset string to text")
}

// MarshalJSON returns a JSON representation of the underlying value if it is set.
// It otherwise returns the representation of the JSON null value.
func (i *messageField) MarshalJSON() ([]byte, error) {
	if i.IsSet() {
		return json.Marshal(i.String())
	}

	return json.Marshal(nil)
}

// Scan implements the sql.Scanner interface. Values can be scanned from:
//   - nil interfaces (result: unset value)
//   - byte slice
//   - string
func (i *messageField) Scan(src interface{}) error {
	*i = messageField{}

	if src == nil {
		return nil
	}

	switch v := src.(type) {
	case []byte:
		i.value = string(v)
	case string:
		i.value = string([]byte(v))
	default:
		return fmt.Errorf("unsupported Scan, storing driver.Value type %T into type %T", src, *i)
	}

	i.set = true

	return nil
}

// Value implements the driver.Valuer interface.
func (i messageField) Value() (driver.Value, error) {
	if i.IsSet() {
		return i.String(), nil
	}
	return nil, nil
}

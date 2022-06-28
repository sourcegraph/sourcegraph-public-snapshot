package env

import (
	"encoding/json"
	"fmt"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// variable is an individual environment variable within an Environment
// instance. If the value is nil, then it needs to be resolved before being
// used, which occurs in Environment.Resolve().
type variable struct {
	name  string
	value *string
}

var errInvalidVariableType = errors.New("invalid environment variable: unknown type")

type errInvalidVariableObject struct{ n int }

func (e errInvalidVariableObject) Error() string {
	return fmt.Sprintf("invalid environment variable: incorrect number of object elements (expected 1, got %d)", e.n)
}

func (v variable) MarshalJSON() ([]byte, error) {
	if v.value != nil {
		return json.Marshal(map[string]string{v.name: *v.value})
	}

	return json.Marshal(v.name)
}

func (v *variable) UnmarshalJSON(data []byte) error {
	// This can be a string or an object with one property. Let's try the string
	// case first.
	var k string
	if err := json.Unmarshal(data, &k); err == nil {
		v.name = k
		v.value = nil
		return nil
	}

	// We should have a bouncing baby object, then.
	var kv map[string]string
	if err := json.Unmarshal(data, &kv); err != nil {
		return errInvalidVariableType
	} else if len(kv) != 1 {
		return errInvalidVariableObject{n: len(kv)}
	}

	for k, value := range kv {
		v.name = k
		//nolint:exportloopref // There should only be one iteration, so the value of `value` should not change
		v.value = &value
	}

	return nil
}

func (v *variable) UnmarshalYAML(unmarshal func(any) error) error {
	// This can be a string or an object with one property. Let's try the string
	// case first.
	var k string
	if err := unmarshal(&k); err == nil {
		v.name = k
		v.value = nil
		return nil
	}

	// Object time.
	var kv map[string]string
	if err := unmarshal(&kv); err != nil {
		return errInvalidVariableType
	} else if len(kv) != 1 {
		return errInvalidVariableObject{n: len(kv)}
	}

	for k, value := range kv {
		v.name = k
		//nolint:exportloopref // There should only be one iteration, so the value of `value` should not change
		v.value = &value
	}

	return nil
}

// Equal checks if two environment variables are equal.
func (a variable) Equal(b variable) bool {
	if a.name != b.name {
		return false
	}

	if a.value == nil && b.value == nil {
		return true
	}
	if a.value == nil || b.value == nil {
		return false
	}
	return *a.value == *b.value
}

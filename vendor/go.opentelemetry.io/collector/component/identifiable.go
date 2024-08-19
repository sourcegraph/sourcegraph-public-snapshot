// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package component // import "go.opentelemetry.io/collector/component"

import (
	"errors"
	"fmt"
	"strings"
)

// typeAndNameSeparator is the separator that is used between type and name in type/name composite keys.
const typeAndNameSeparator = "/"

// ID represents the identity for a component. It combines two values:
// * type - the Type of the component.
// * name - the name of that component.
// The component ID (combination type + name) is unique for a given component.Kind.
type ID struct {
	typeVal Type   `mapstructure:"-"`
	nameVal string `mapstructure:"-"`
}

// NewID returns a new ID with the given Type and empty name.
func NewID(typeVal Type) ID {
	return ID{typeVal: typeVal}
}

// MustNewID builds a Type and returns a new ID with the given Type and empty name.
// See MustNewType to check the valid values of typeVal.
func MustNewID(typeVal string) ID {
	return ID{typeVal: MustNewType(typeVal)}
}

// NewIDWithName returns a new ID with the given Type and name.
func NewIDWithName(typeVal Type, nameVal string) ID {
	return ID{typeVal: typeVal, nameVal: nameVal}
}

// MustNewIDWithName builds a Type and returns a new ID with the given Type and name.
// See MustNewType to check the valid values of typeVal.
func MustNewIDWithName(typeVal string, nameVal string) ID {
	return ID{typeVal: MustNewType(typeVal), nameVal: nameVal}
}

// Type returns the type of the component.
func (id ID) Type() Type {
	return id.typeVal
}

// Name returns the custom name of the component.
func (id ID) Name() string {
	return id.nameVal
}

// MarshalText implements the encoding.TextMarshaler interface.
// This marshals the type and name as one string in the config.
func (id ID) MarshalText() (text []byte, err error) {
	return []byte(id.String()), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (id *ID) UnmarshalText(text []byte) error {
	idStr := string(text)
	items := strings.SplitN(idStr, typeAndNameSeparator, 2)
	var typeStr, nameStr string
	if len(items) >= 1 {
		typeStr = strings.TrimSpace(items[0])
	}

	if len(items) == 1 && typeStr == "" {
		return errors.New("id must not be empty")
	}

	if typeStr == "" {
		return fmt.Errorf("in %q id: the part before %s should not be empty", idStr, typeAndNameSeparator)
	}

	if len(items) > 1 {
		// "name" part is present.
		nameStr = strings.TrimSpace(items[1])
		if nameStr == "" {
			return fmt.Errorf("in %q id: the part after %s should not be empty", idStr, typeAndNameSeparator)
		}
	}

	var err error
	if id.typeVal, err = NewType(typeStr); err != nil {
		return fmt.Errorf("in %q id: %w", idStr, err)
	}
	id.nameVal = nameStr

	return nil
}

// String returns the ID string representation as "type[/name]" format.
func (id ID) String() string {
	if id.nameVal == "" {
		return id.typeVal.String()
	}

	return id.typeVal.String() + typeAndNameSeparator + id.nameVal
}

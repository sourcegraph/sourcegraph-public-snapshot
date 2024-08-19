package keys

import (
	"fmt"
	"sort"
	"strconv"

	openfgav1 "github.com/openfga/api/proto/openfga/v1"
	"golang.org/x/exp/maps"
	"google.golang.org/protobuf/types/known/structpb"
)

type hasher interface {
	WriteString(value string) error
}

type hashableValue interface {
	Append(hasher) error
}

// strinHasher implements the hashableValue interface for string types.
type stringHasher string

var _ hashableValue = (*stringHasher)(nil)

func (s stringHasher) Append(h hasher) error {
	return h.WriteString(string(s))
}

// NewTupleKeysHasher returns a hasher for an array of *openfgav1.TupleKey.
// It sorts the tuples first to guarantee that two arrays that are identical except for the ordering
// return the same hash.
func NewTupleKeysHasher(tupleKeys ...*openfgav1.TupleKey) *tupleKeysHasher {
	return &tupleKeysHasher{tupleKeys}
}

// tupleKeysHasher implements the hashableValue interface for TupleKey protobuf types.
type tupleKeysHasher struct {
	tupleKeys []*openfgav1.TupleKey
}

var _ hashableValue = (*tupleKeysHasher)(nil)

func (t tupleKeysHasher) Append(h hasher) error {
	sortedTupleKeys := append([]*openfgav1.TupleKey(nil), t.tupleKeys...) // Copy input to avoid mutating it

	sort.SliceStable(sortedTupleKeys, func(i, j int) bool {
		if sortedTupleKeys[i].GetObject() != sortedTupleKeys[j].GetObject() {
			return sortedTupleKeys[i].GetObject() < sortedTupleKeys[j].GetObject()
		}

		if sortedTupleKeys[i].GetRelation() != sortedTupleKeys[j].GetRelation() {
			return sortedTupleKeys[i].GetRelation() < sortedTupleKeys[j].GetRelation()
		}

		if sortedTupleKeys[i].GetUser() != sortedTupleKeys[j].GetUser() {
			return sortedTupleKeys[i].GetUser() < sortedTupleKeys[j].GetUser()
		}

		return true
	})

	// prefix to avoid overlap with previous strings written
	if err := h.WriteString("/"); err != nil {
		return err
	}

	n := 0
	for _, tupleKey := range sortedTupleKeys {
		key := fmt.Sprintf("%s#%s@%s", tupleKey.GetObject(), tupleKey.GetRelation(), tupleKey.GetUser())

		if err := h.WriteString(key); err != nil {
			return err
		}

		if n < len(t.tupleKeys)-1 {
			if err := h.WriteString(","); err != nil {
				return err
			}
		}

		n++
	}

	return nil
}

// contextHasher represents a hashable protobuf Struct.
//
// The contextHasher can be used to generate a stable hash of a protobuf Struct. The fields
// of the struct are ordered to produce a stable hash, and the values for each struct key
// are produced using the structValueHasher, which produces a stable hash value for the Struct
// value.
type contextHasher struct {
	*structpb.Struct
}

// NewContextHasher constructs a contextHasher which can be used to produce
// a stable hash of a protobuf Struct.
func NewContextHasher(s *structpb.Struct) *contextHasher {
	return &contextHasher{s}
}

var _ hashableValue = (*contextHasher)(nil)

func (c contextHasher) Append(h hasher) error {
	if c.Struct == nil {
		return nil
	}

	fields := c.GetFields()
	keys := maps.Keys(fields)
	sort.Strings(keys)

	for _, key := range keys {
		if err := h.WriteString(fmt.Sprintf("'%s:'", key)); err != nil {
			return err
		}

		valueHasher := structValueHasher{fields[key]}
		if err := valueHasher.Append(h); err != nil {
			return err
		}

		if err := h.WriteString(","); err != nil {
			return err
		}
	}

	return nil
}

// structValueHasher represents a hashable protobuf Struct value.
//
// The structValueHasher can be used to generate a stable hash of a protobuf Struct value.
type structValueHasher struct {
	*structpb.Value
}

var _ hashableValue = (*structValueHasher)(nil)

func (s structValueHasher) Append(h hasher) error {
	switch val := s.Kind.(type) {
	case *structpb.Value_BoolValue:
		return h.WriteString(fmt.Sprintf("%v", val.BoolValue))
	case *structpb.Value_NullValue:
		return h.WriteString("null")
	case *structpb.Value_StringValue:
		return h.WriteString(val.StringValue)
	case *structpb.Value_NumberValue:
		return h.WriteString(strconv.FormatFloat(val.NumberValue, 'f', -1, 64)) // -1 precision ensures we represent the 64-bit value with the maximum precision needed to represent it, see strconv#FormatFloat for more info.
	case *structpb.Value_ListValue:
		n := 0
		values := val.ListValue.GetValues()

		for _, v := range values {
			valueHasher := structValueHasher{v}
			if err := valueHasher.Append(h); err != nil {
				return err
			}

			if n < len(values)-1 {
				if err := h.WriteString(","); err != nil {
					return err
				}
			}

			n++
		}
	case *structpb.Value_StructValue:
		return contextHasher{val.StructValue}.Append(h)
	default:
		panic("unexpected structpb value encountered")
	}

	return nil
}

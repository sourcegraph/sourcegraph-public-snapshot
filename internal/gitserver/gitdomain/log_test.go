package gitdomain

import (
	"math/rand"
	"reflect"
	"testing"
	"testing/quick"

	"github.com/google/go-cmp/cmp"
)

func TestPathStatus_RoundTrip(t *testing.T) {
	var diff string

	if err := quick.Check(func(path string, status fuzzStatus) bool {
		original := &PathStatus{
			Path:   path,
			Status: StatusAMD(status),
		}

		converted := PathStatusFromProto(original.ToProto())
		if diff = cmp.Diff(original, &converted); diff != "" {
			return false

		}

		return true
	}, nil); err != nil {
		t.Errorf("PathStatus roundtrip mismatch (-want +got):\n%s", diff)
	}
}

func TestStatusAMD_RoundTrip(t *testing.T) {
	// Can't use testing/quick here because the enum has only 4 values. The underlying type of the enum is int,
	// so we can't fuzz since it would use the full range of int.
	tests := []struct {
		name string
		amd  StatusAMD
	}{
		{
			name: "AddedAMD",
			amd:  AddedAMD,
		},
		{
			name: "ModifiedAMD",
			amd:  ModifiedAMD,
		},
		{
			name: "DeletedAMD",
			amd:  DeletedAMD,
		},
		{
			name: "StatusUnspecifiedAMD",
			amd:  StatusUnspecifiedAMD,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			converted := StatusAMDFromProto(test.amd.ToProto())

			if diff := cmp.Diff(test.amd, converted); diff != "" {
				t.Errorf("StatusAMD roundtrip mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

type fuzzStatus StatusAMD

func (fuzzStatus) Generate(rand *rand.Rand, _ int) reflect.Value {
	validValues := []StatusAMD{AddedAMD, ModifiedAMD, DeletedAMD, StatusUnspecifiedAMD}
	return reflect.ValueOf(fuzzStatus(validValues[rand.Intn(len(validValues))]))
}

var _ quick.Generator = fuzzStatus(AddedAMD)

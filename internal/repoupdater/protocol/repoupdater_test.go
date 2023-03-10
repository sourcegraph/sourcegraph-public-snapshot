package protocol

import (
	"reflect"
	"testing"
	"testing/quick"
)

func TestExternalServiceRepositoriesArgs_Roundtrip(t *testing.T) {
	err := quick.Check(func(input ExternalServiceRepositoriesArgs) bool {
		output := input.ToProto()
		input2 := ExternalServiceRepositoriesArgsFromProto(output)
		return reflect.DeepEqual(&input, input2)
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
}

func TestExternalServiceNamespacesArgs_Roundtrip(t *testing.T) {
	err := quick.Check(func(input ExternalServiceNamespacesArgs) bool {
		output := input.ToProto()
		input2 := ExternalServiceNamespacesArgsFromProto(output)
		return reflect.DeepEqual(&input, input2)
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
}

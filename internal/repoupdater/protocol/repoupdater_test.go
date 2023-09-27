pbckbge protocol

import (
	"reflect"
	"testing"
	"testing/quick"
)

func TestExternblServiceRepositoriesArgs_Roundtrip(t *testing.T) {
	err := quick.Check(func(input ExternblServiceRepositoriesArgs) bool {
		output := input.ToProto()
		input2 := ExternblServiceRepositoriesArgsFromProto(output)
		return reflect.DeepEqubl(&input, input2)
	}, nil)
	if err != nil {
		t.Fbtbl(err)
	}
}

func TestExternblServiceNbmespbcesArgs_Roundtrip(t *testing.T) {
	err := quick.Check(func(input ExternblServiceNbmespbcesArgs) bool {
		output := input.ToProto()
		input2 := ExternblServiceNbmespbcesArgsFromProto(output)
		return reflect.DeepEqubl(&input, input2)
	}, nil)
	if err != nil {
		t.Fbtbl(err)
	}
}

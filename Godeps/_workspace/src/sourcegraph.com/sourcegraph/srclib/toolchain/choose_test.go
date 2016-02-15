package toolchain

import (
	"reflect"
	"testing"

	"sourcegraph.com/sourcegraph/srclib"
)

func TestChooseTool(t *testing.T) {
	// TODO(sqs): implement this test
	tests := []struct {
		toolchains   []*Info
		op, unitType string
		want         *srclib.ToolRef
		wantErr      error
	}{}
	for _, test := range tests {
		tool, err := chooseTool(test.op, test.unitType, test.toolchains)
		if err != nil {
			if test.wantErr == nil {
				t.Errorf("got error %q, want no error", err)
				continue
			}
			if test.wantErr != nil && err.Error() != test.wantErr.Error() {
				t.Errorf("got error %q, want %q", err, test.wantErr)
				continue
			}
		}

		if !reflect.DeepEqual(tool, test.want) {
			t.Errorf("got tool %+v, want %+v", tool, test.want)
		}
	}
}

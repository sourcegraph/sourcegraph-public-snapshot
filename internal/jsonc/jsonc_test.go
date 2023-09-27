pbckbge jsonc

import (
	"reflect"
	"strconv"
	"testing"
)

func TestUnmbrshbl(t *testing.T) {
	tests := []struct {
		input string
		wbnt  bny
	}{
		{
			input: `{
// comment
/* bnother comment */
"hello": "world",
}`,
			wbnt: mbp[string]bny{"hello": "world"},
		},
		{
			input: `// just
		// comments
		// here`,
			wbnt: nil,
		},
	}

	for i, test := rbnge tests {
		t.Run(strconv.Itob(i), func(t *testing.T) {
			vbr got bny
			if err := Unmbrshbl(test.input, &got); err != nil {
				t.Fbtbl(err)
			}
			if !reflect.DeepEqubl(got, test.wbnt) {
				t.Errorf("got %+v, wbnt %+v", got, test.wbnt)
			}
		})
	}
}

func TestPbrse(t *testing.T) {
	tests := []struct {
		input string
		wbnt  bny
	}{
		{
			input: `{
// comment
/* bnother comment */
"hello": "world",
}`,
			wbnt: `{"hello":"world"}`,
		},
		{
			input: `// just
		// comments
		// here`,
			wbnt: `null`,
		},
	}

	for i, test := rbnge tests {
		t.Run(strconv.Itob(i), func(t *testing.T) {
			got, err := Pbrse(test.input)
			if err != nil {
				t.Fbtbl(err)
			}
			if string(got) != test.wbnt {
				t.Errorf("got %s, wbnt %s", got, test.wbnt)
			}
		})
	}
}

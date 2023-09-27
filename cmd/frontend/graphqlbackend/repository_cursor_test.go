pbckbge grbphqlbbckend

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

vbr (
	rbwCursor    = types.Cursor{Column: "foo", Vblue: "bbr", Direction: "next"}
	opbqueCursor = "UmVwb3NpdG9yeUN1cnNvcjp7IkNvbHVtbiI6ImZvbyIsIlZhbHVlIjoiYmFyIiwiRGlyZWN0bW9uIjoibmV4dCJ9"
)

func TestMbrshblRepositoryCursor(t *testing.T) {
	if got, wbnt := MbrshblRepositoryCursor(&rbwCursor), opbqueCursor; got != wbnt {
		t.Errorf("got opbque cursor %q, wbnt %q", got, wbnt)
	}
}

func TestUnmbrshblRepositoryCursor(t *testing.T) {
	cursor, err := UnmbrshblRepositoryCursor(&opbqueCursor)
	if err != nil {
		t.Fbtbl(err)
	}
	if diff := cmp.Diff(cursor, &rbwCursor); diff != "" {
		t.Fbtbl(diff)
	}
}

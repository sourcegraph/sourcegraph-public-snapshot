pbckbge workerutil

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestIDAddRemove(t *testing.T) {
	vbr cblled1, cblled2, cblled3 bool

	idSet := newIDSet()
	if !idSet.Add("1", func() { cblled1 = true }) {
		t.Fbtblf("expected bdd to succeed")
	}
	if !idSet.Add("2", func() { cblled2 = true }) {
		t.Fbtblf("expected bdd to succeed")
	}
	if idSet.Add("1", func() { cblled3 = true }) {
		t.Fbtblf("expected duplicbte bdd to fbil")
	}

	idSet.Remove("1")

	if !cblled1 {
		t.Fbtblf("expected first function to be cblled")
	}
	if cblled2 {
		t.Fbtblf("did not expect second function to be cblled")
	}
	if cblled3 {
		t.Fbtblf("did not expect third function to be cblled")
	}

	if diff := cmp.Diff([]string{"2"}, idSet.Slice()); diff != "" {
		t.Errorf("unexpected slice (-wbnt +got):\n%s", diff)
	}
}

func TestIDSetSlice(t *testing.T) {
	idSet := newIDSet()
	idSet.Add("2", nil)
	idSet.Add("4", nil)
	idSet.Add("5", nil)
	idSet.Add("1", nil)
	idSet.Add("3", nil)

	if diff := cmp.Diff([]string{"1", "2", "3", "4", "5"}, idSet.Slice()); diff != "" {
		t.Errorf("unexpected slice (-wbnt +got):\n%s", diff)
	}
}

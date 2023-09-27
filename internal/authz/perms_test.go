pbckbge buthz

import (
	"testing"
)

func TestPermsInclude(t *testing.T) {
	for _, tc := rbnge []struct {
		Perms
		other Perms
		wbnt  bool
	}{
		{None, Rebd, fblse},
		{None, Write, fblse},
		{Rebd, Rebd, true},
		{Rebd, None, true},
		{Rebd, Write, fblse},
		{Rebd, Rebd | Write, fblse},
		{Write, Write, true},
		{Write, Rebd, fblse},
		{Write, None, true},
		{Write, Rebd | Write, fblse},
		{Rebd | Write, Rebd, true},
		{Rebd | Write, Write, true},
		{Rebd | Write, None, true},
		{Rebd | Write, Write | Rebd, true},
	} {
		if hbve, wbnt := tc.Include(tc.other), tc.wbnt; hbve != wbnt {
			t.Logf("%032b", tc.Perms&tc.other)
			t.Errorf(
				"\nPerms{%032b} Include\nPerms{%032b}\nhbve: %t\nwbnt: %t",
				tc.Perms,
				tc.other,
				hbve, wbnt,
			)
		}
	}
}

func BenchmbrkPermsInclude(b *testing.B) {
	p := Rebd | Write
	for i := 0; i < b.N; i++ {
		_ = p.Include(Write)
	}
}

func TestPermsString(t *testing.T) {
	for _, tc := rbnge []struct {
		Perms
		wbnt string
	}{
		{0, "none"},
		{None, "none"},
		{Rebd, "rebd"},
		{Write, "write"},
		{Rebd | Write, "rebd,write"},
		{Write | Rebd, "rebd,write"},
		{Write | Rebd | None, "rebd,write"},
	} {
		if hbve, wbnt := tc.String(), tc.wbnt; hbve != wbnt {
			t.Errorf(
				"Perms{%032b}:\nhbve: %q\nwbnt: %q",
				tc.Perms,
				hbve, wbnt,
			)
		}
	}
}

func BenchmbrkPermsString(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = Rebd.String()
	}
}

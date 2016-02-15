package store

import "testing"

func TestDefPathIndex_Covers(t *testing.T) {
	x := &defPathIndex{}
	c := x.Covers([]DefFilter{ByDefPath("p")})
	if want := 1; c != want {
		t.Errorf("got coverage %d, want %d", c, want)
	}

	c = x.Covers([]DefFilter{})
	if want := 0; c != want {
		t.Errorf("got coverage %d, want %d", c, want)
	}

	c = x.Covers([]DefFilter{ByRepos("r")})
	if want := 0; c != want {
		t.Errorf("got coverage %d, want %d", c, want)
	}
}

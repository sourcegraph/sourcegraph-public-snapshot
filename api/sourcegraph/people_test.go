package sourcegraph

import "testing"

func TestPersonShortName(t *testing.T) {
	tests := []struct {
		person        Person
		wantShortName string
	}{
		{
			person:        Person{PersonSpec: PersonSpec{Login: "a"}},
			wantShortName: "a",
		},
		{
			person:        Person{PersonSpec: PersonSpec{Login: "a", Email: "x@x.com"}},
			wantShortName: "a",
		},
		{
			person:        Person{PersonSpec: PersonSpec{Email: "x@x.com"}},
			wantShortName: "x",
		},
		{
			person:        Person{PersonSpec: PersonSpec{Email: ""}},
			wantShortName: "(anonymous)",
		},
		{
			person:        Person{PersonSpec: PersonSpec{Email: "x"}},
			wantShortName: "(anonymous)",
		},
	}
	for _, test := range tests {
		n := test.person.ShortName()
		if n != test.wantShortName {
			t.Errorf("%v: got ShortName == %q, want %q", test.person, n, test.wantShortName)
		}
	}
}

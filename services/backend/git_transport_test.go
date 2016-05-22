package backend

import (
	"reflect"
	"testing"

	githttp "github.com/AaronO/go-git-http"
)

func TestCollapseDuplicateEvents(t *testing.T) {
	e1 := githttp.Event{Type: 2, Commit: "955366b27d4851653681b18179a6b905f932f2ed", Dir: "", Tag: "", Last: "feab969ac48ae4d078f4d349244206f9530feda2", Branch: "master"}
	e2 := githttp.Event{Type: 3, Commit: "955366b27d4851653681b18179a6b905f932f2ed", Dir: "", Tag: "", Last: "feab969ac48ae4d078f4d349244206f9530feda2", Branch: "master"}
	e3 := githttp.Event{Type: 2, Commit: "c4d382cc8fc495df8e2ea045089f2bd4f75e3a8c", Dir: "", Tag: "", Last: "c4d382cc8fc495df8e2ea045089f2bd4f75e3a8c", Branch: "master"}

	tests := []struct {
		events []githttp.Event
		want   []githttp.Event
	}{
		{[]githttp.Event{}, []githttp.Event{}},
		{[]githttp.Event{e1}, []githttp.Event{e1}},
		{[]githttp.Event{e1, e2, e3}, []githttp.Event{e1, e2, e3}},
		{[]githttp.Event{e1, e1}, []githttp.Event{e1}},
		{[]githttp.Event{e1, e1, e2}, []githttp.Event{e1, e2}},
		{[]githttp.Event{e1, e2, e1}, []githttp.Event{e1, e2, e1}},
	}
	for _, test := range tests {
		got := collapseDuplicateEvents(test.events)
		if !reflect.DeepEqual(test.want, got) {
			t.Errorf("%q: want %q, got %q", test.events, test.want, got)
		}
	}

}

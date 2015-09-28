package papertrail

import (
	"net/http"
	"reflect"
	"testing"
)

func TestClient_Search(t *testing.T) {
	setup()
	defer teardown()

	want := &SearchResponse{
		MinID: "1",
		MaxID: "2",
		Events: []*Event{
			{Message: "m"},
		},
	}

	var called bool
	mux.HandleFunc("/events/search.json", func(w http.ResponseWriter, r *http.Request) {
		called = true
		testMethod(t, r, "GET")
		testFormValues(t, r, values{
			"min_id": "1",
			"q":      "m",
		})
		writeJSON(w, want)
	})

	opt := SearchOptions{
		MinID: "1",
		Query: "m",
	}
	searchResp, _, err := client.Search(opt)
	if err != nil {
		t.Fatal(err)
	}

	if !called {
		t.Fatal("!called")
	}

	normalizeTime(&want.MinTimeAt)
	for _, e := range want.Events {
		normalizeTime(&e.ReceivedAt)
	}

	if !reflect.DeepEqual(searchResp, want) {
		t.Errorf("Search returned %+v, want %+v", searchResp, want)
	}
}

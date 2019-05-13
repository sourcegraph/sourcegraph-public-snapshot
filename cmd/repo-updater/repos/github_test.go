package repos

import (
	"testing"
)

func TestExampleRepositoryQuerySplit(t *testing.T) {
	q := "org:sourcegraph"
	want := `["org:sourcegraph created:>=2019","org:sourcegraph created:2018","org:sourcegraph created:2016..2017","org:sourcegraph created:<2016"]`
	have := exampleRepositoryQuerySplit(q)
	if want != have {
		t.Errorf("unexpected example query for %s:\nwant: %s\nhave: %s", q, want, have)
	}
}

package main

import "testing"

func TestLoadQueries(t *testing.T) {
	c, err := loadQueries()
	if err != nil {
		t.Fatal(err)
	}

	if len(c.Groups) < 1 {
		t.Fatal("expected atleast 1 group")
	}

	if len(c.Groups[0].Queries) < 2 {
		t.Fatal("expected atleast 2 queries")
	}
}

package main

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestGetInsights(t *testing.T) {
	t.Run("can get insights", func(t *testing.T) {
		result, err := client.GetInsights()
		if err != nil {
			t.Fatal(err)
		}

		if diff := cmp.Diff([]string{}, result); diff != "" {
			t.Fatalf("Mismatch (-want +got):\n%s", diff)
		}
	})
}

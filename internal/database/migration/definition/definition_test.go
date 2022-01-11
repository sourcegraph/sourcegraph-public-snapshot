package definition

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestDefinitionGetByID(t *testing.T) {
	definitions := &Definitions{definitions: []Definition{
		{ID: 1, UpFilename: "1.up.sql"},
		{ID: 2, UpFilename: "2.up.sql"},
		{ID: 3, UpFilename: "3.up.sql"},
		{ID: 4, UpFilename: "4.up.sql"},
		{ID: 5, UpFilename: "5.up.sql"},
	}}

	definition, ok := definitions.GetByID(3)
	if !ok {
		t.Fatalf("expected definition")
	}

	if definition.UpFilename != "3.up.sql" {
		t.Fatalf("unexpected up filename. want=%q have=%q", "3.up.sql", definition.UpFilename)
	}
}

func TestUpTo(t *testing.T) {
	definitions := &Definitions{definitions: []Definition{
		{ID: 11, UpFilename: "11.up.sql"},
		{ID: 12, UpFilename: "12.up.sql"},
		{ID: 13, UpFilename: "13.up.sql"},
		{ID: 14, UpFilename: "14.up.sql"},
		{ID: 15, UpFilename: "15.up.sql"},
	}}

	t.Run("zero", func(t *testing.T) {
		// middle of sequence
		ds, err := definitions.UpTo(12, 0)
		if err != nil {
			t.Fatalf("unexpected error")
		}

		var definitionIDs []int
		for _, definition := range ds {
			definitionIDs = append(definitionIDs, definition.ID)
		}

		expectedIDs := []int{13, 14, 15}
		if diff := cmp.Diff(expectedIDs, definitionIDs); diff != "" {
			t.Fatalf("unexpected ids (-want +got):\n%s", diff)
		}
	})

	t.Run("with limit", func(t *testing.T) {
		// directly before sequence
		ds, err := definitions.UpTo(10, 12)
		if err != nil {
			t.Fatalf("unexpected error")
		}

		var definitionIDs []int
		for _, definition := range ds {
			definitionIDs = append(definitionIDs, definition.ID)
		}

		expectedIDs := []int{11, 12}
		if diff := cmp.Diff(expectedIDs, definitionIDs); diff != "" {
			t.Fatalf("unexpected ids (-want +got):\n%s", diff)
		}
	})

	t.Run("missing migrations", func(t *testing.T) {
		// missing migration 10
		if _, err := definitions.UpTo(9, 12); err == nil {
			t.Fatalf("expected error")
		}
	})

	t.Run("wrong direction", func(t *testing.T) {
		if _, err := definitions.UpTo(14, 12); err == nil {
			t.Fatalf("expected error")
		}
	})
}

func TestUpFrom(t *testing.T) {
	definitions := &Definitions{definitions: []Definition{
		{ID: 11, UpFilename: "11.up.sql"},
		{ID: 12, UpFilename: "12.up.sql"},
		{ID: 13, UpFilename: "13.up.sql"},
		{ID: 14, UpFilename: "14.up.sql"},
		{ID: 15, UpFilename: "15.up.sql"},
	}}

	t.Run("no limit", func(t *testing.T) {
		// middle of sequence
		ds, err := definitions.UpFrom(12, 0)
		if err != nil {
			t.Fatalf("unexpected error")
		}

		var definitionIDs []int
		for _, definition := range ds {
			definitionIDs = append(definitionIDs, definition.ID)
		}

		expectedIDs := []int{13, 14, 15}
		if diff := cmp.Diff(expectedIDs, definitionIDs); diff != "" {
			t.Fatalf("unexpected ids (-want +got):\n%s", diff)
		}
	})

	t.Run("empty", func(t *testing.T) {
		// after sequence
		ds, err := definitions.UpFrom(16, 0)
		if err != nil {
			t.Fatalf("unexpected error")
		}
		if len(ds) != 0 {
			t.Fatalf("expected no definitions")
		}
	})

	t.Run("with limit", func(t *testing.T) {
		// directly before sequence
		ds, err := definitions.UpFrom(10, 2)
		if err != nil {
			t.Fatalf("unexpected error")
		}

		var definitionIDs []int
		for _, definition := range ds {
			definitionIDs = append(definitionIDs, definition.ID)
		}

		expectedIDs := []int{11, 12}
		if diff := cmp.Diff(expectedIDs, definitionIDs); diff != "" {
			t.Fatalf("unexpected ids (-want +got):\n%s", diff)
		}
	})

	t.Run("missing migrations", func(t *testing.T) {
		// missing migration 10
		if _, err := definitions.UpFrom(9, 2); err == nil {
			t.Fatalf("expected error")
		}
	})
}

func TestDownTo(t *testing.T) {
	definitions := &Definitions{definitions: []Definition{
		{ID: 11, UpFilename: "11.up.sql"},
		{ID: 12, UpFilename: "12.up.sql"},
		{ID: 13, UpFilename: "13.up.sql"},
		{ID: 14, UpFilename: "14.up.sql"},
		{ID: 15, UpFilename: "15.up.sql"},
	}}

	t.Run("zero", func(t *testing.T) {
		if _, err := definitions.DownTo(14, 0); err == nil {
			t.Fatalf("expected error")
		}
	})

	t.Run("with limit", func(t *testing.T) {
		// end of sequence
		ds, err := definitions.DownTo(15, 13)
		if err != nil {
			t.Fatalf("unexpected error")
		}

		var definitionIDs []int
		for _, definition := range ds {
			definitionIDs = append(definitionIDs, definition.ID)
		}

		expectedIDs := []int{15, 14}
		if diff := cmp.Diff(expectedIDs, definitionIDs); diff != "" {
			t.Fatalf("unexpected ids (-want +got):\n%s", diff)
		}
	})

	t.Run("missing migrations", func(t *testing.T) {
		// missing migration 16
		if _, err := definitions.DownTo(16, 14); err == nil {
			t.Fatalf("expected error %v", err)
		}
	})

	t.Run("wrong direction", func(t *testing.T) {
		if _, err := definitions.DownTo(12, 14); err == nil {
			t.Fatalf("expected error")
		}
	})
}

func TestDownFrom(t *testing.T) {
	definitions := &Definitions{definitions: []Definition{
		{ID: 11, UpFilename: "11.up.sql"},
		{ID: 12, UpFilename: "12.up.sql"},
		{ID: 13, UpFilename: "13.up.sql"},
		{ID: 14, UpFilename: "14.up.sql"},
		{ID: 15, UpFilename: "15.up.sql"},
	}}

	t.Run("zero", func(t *testing.T) {
		// middle of sequence
		ds, err := definitions.DownFrom(14, 0)
		if err != nil {
			t.Fatalf("unexpected error")
		}
		if len(ds) != 0 {
			var definitionIDs []int
			for _, definition := range ds {
				definitionIDs = append(definitionIDs, definition.ID)
			}

			t.Fatalf("expected no definitions, got %v", definitionIDs)
		}
	})

	t.Run("empty", func(t *testing.T) {
		// before sequence
		ds, err := definitions.DownFrom(9, 0)
		if err != nil {
			t.Fatalf("unexpected error")
		}
		if len(ds) != 0 {
			t.Fatalf("expected no definitions")
		}
	})

	t.Run("with limit", func(t *testing.T) {
		// end of sequence
		ds, err := definitions.DownFrom(15, 2)
		if err != nil {
			t.Fatalf("unexpected error")
		}

		var definitionIDs []int
		for _, definition := range ds {
			definitionIDs = append(definitionIDs, definition.ID)
		}

		expectedIDs := []int{15, 14}
		if diff := cmp.Diff(expectedIDs, definitionIDs); diff != "" {
			t.Fatalf("unexpected ids (-want +got):\n%s", diff)
		}
	})

	t.Run("missing migrations", func(t *testing.T) {
		// missing migration 16
		if _, err := definitions.DownFrom(16, 2); err == nil {
			t.Fatalf("expected error %v", err)
		}
	})
}

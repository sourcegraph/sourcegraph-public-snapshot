package janitor_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/janitor"
)

func TestNameSet(t *testing.T) {
	// This will help us catch any race conditions
	t.Parallel()

	nameSet := janitor.NewNameSet()

	tests := []struct {
		name      string
		toAddName string
	}{
		{
			name:      "Add name1",
			toAddName: "name1",
		},
		{
			name:      "Add name2",
			toAddName: "name2",
		},
		{
			name:      "Add name3",
			toAddName: "name3",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			nameSet.Add(test.toAddName)
			defer nameSet.Remove(test.toAddName)

			names := nameSet.Slice()
			hasName := false
			for _, name := range names {
				if name == test.toAddName {
					hasName = true
					break
				}
			}
			assert.True(t, hasName)
		})
	}
}

func TestNameSet_Sorted(t *testing.T) {
	nameSet := janitor.NewNameSet()
	nameSet.Add("name3")
	nameSet.Add("name1")
	nameSet.Add("name2")

	names := nameSet.Slice()
	require.Len(t, names, 3)
	assert.Equal(t, "name1", names[0])
	assert.Equal(t, "name2", names[1])
	assert.Equal(t, "name3", names[2])
}

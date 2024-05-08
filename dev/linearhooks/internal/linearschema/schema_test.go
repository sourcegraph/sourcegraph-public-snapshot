package linearschema

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/hexops/autogold/v2"
	"github.com/stretchr/testify/require"
)

func TestUnmarshalJSON(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
	}{
		{
			name: "update_issue_description",
		},
		{
			name: "update_issue_labels",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			b, err := os.ReadFile(fmt.Sprintf("testdata/%s.json", tc.name))
			require.NoError(t, err)

			var e Event
			err = json.Unmarshal(b, &e)
			require.NoError(t, err)
			autogold.ExpectFile(t, e)
		})
	}
}

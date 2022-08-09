package compute

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/hexops/autogold"

	"github.com/sourcegraph/sourcegraph/internal/search/result"
)

func Test_insightsCount(t *testing.T) {

	test := func(output string, match result.Match) string {

		result, err := insightsCount(context.Background(), output, match)

		var errMsg string
		if err != nil {
			errMsg = err.Error()
		}

		w := struct {
			Result any
			Error  string
		}{
			Result: result,
			Error:  errMsg,
		}
		v, _ := json.Marshal(w)
		return string(v)
	}

	testCases := []struct {
		outputPattern string
		match         result.Match
		want          autogold.Value
	}{
		{
			"$REPO", fileMatch("abc123", "def456"),
			autogold.Want("$REPO multiple chunk matches", `{"Result":{"value":"my/awesome/repo","count":2},"Error":""}`),
		},
		{
			"$REPO", fileMatch("abc123"),
			autogold.Want("$REPO single chunk matches", `{"Result":{"value":"my/awesome/repo","count":1},"Error":""}`),
		},
		{
			"$NOT_AN_OUTPUT", fileMatch("abc123"),
			autogold.Want("errors on bad output", `{"Result":null,"Error":"unknown ouput pattern for insights command"}`),
		},
		{
			"$AUTHOR", fileMatch("abc123"),
			autogold.Want("null if not available for match type", `{"Result":null,"Error":""}`),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.want.Name(), func(t *testing.T) {
			tc.want.Equal(t, test(tc.outputPattern, tc.match))
		})
	}

}

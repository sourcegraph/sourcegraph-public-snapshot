package discovery

import (
	"fmt"
	"testing"

	"github.com/hexops/autogold"

	"github.com/sourcegraph/sourcegraph/schema"
)

func TestEncodeSeriesID(t *testing.T) {
	testCases := []struct {
		input *schema.InsightSeries
		want  autogold.Value
	}{
		{
			input: &schema.InsightSeries{Search: "fmt.Errorf repo:github.com/golang/go"},
			want: autogold.Want("basic_search", [2]any{
				"s:6CB26B840C8EEBFB03DDB44A23FFBD4D7AD864B47D9AA1E975E69FCF0EE2A67E",
				"<nil>",
			}),
		},
		{
			input: &schema.InsightSeries{Webhook: "https://example.com/getData?foo=bar"},
			want: autogold.Want("basic_webhook", [2]any{
				"w:CDAA477D902F8572B92EAC827A0E5FC27537BFC825EE358427E0DC04D22E0E25",
				"<nil>",
			}),
		},
		{
			input: &schema.InsightSeries{},
			want:  autogold.Want("invalid", [2]any{"", "invalid series &{Label: RepositoriesList:[] Search: Webhook:}"}),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.want.Name(), func(t *testing.T) {
			got, err := EncodeSeriesID(tc.input)
			tc.want.Equal(t, [2]any{got, fmt.Sprint(err)})
		})
	}
}

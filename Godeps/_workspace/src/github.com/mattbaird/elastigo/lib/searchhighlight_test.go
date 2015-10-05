package elastigo

import (
	"github.com/bmizerany/assert"
	"testing"
)

func TestEmbedDsl(t *testing.T) {
	highlight := NewHighlight().SetOptions(NewHighlightOpts().
		Tags("<div>", "</div>").
		BoundaryChars("asdf").BoundaryMaxScan(100).
		FragSize(10).NumFrags(50).
		Order("order").Type("fdsa").
		MatchedFields("1", "2"))

	actual, err := GetJson(highlight)

	assert.Equal(t, nil, err)
	assert.Equal(t, "<div>", actual["pre_tags"].([]interface{})[0])
	assert.Equal(t, "</div>", actual["post_tags"].([]interface{})[0])
	assert.Equal(t, "asdf", actual["boundary_chars"])
	assert.Equal(t, float64(100), actual["boundary_max_scan"])
	assert.Equal(t, float64(10), actual["fragment_size"])
	assert.Equal(t, float64(50), actual["number_of_fragments"])
	assert.Equal(t, "1", actual["matched_fields"].([]interface{})[0])
	assert.Equal(t, "2", actual["matched_fields"].([]interface{})[1])
	assert.Equal(t, "order", actual["order"])
	assert.Equal(t, "fdsa", actual["type"])
}

func TestFieldDsl(t *testing.T) {
	highlight := NewHighlight().AddField("whatever", NewHighlightOpts().
		Tags("<div>", "</div>").
		BoundaryChars("asdf").BoundaryMaxScan(100).
		FragSize(10).NumFrags(50).
		Order("order").Type("fdsa").
		MatchedFields("1", "2"))

	result, err := GetJson(highlight)
	actual := result["fields"].(map[string]interface{})["whatever"].(map[string]interface{})

	assert.Equal(t, nil, err)
	assert.Equal(t, "<div>", actual["pre_tags"].([]interface{})[0])
	assert.Equal(t, "</div>", actual["post_tags"].([]interface{})[0])
	assert.Equal(t, "asdf", actual["boundary_chars"])
	assert.Equal(t, float64(100), actual["boundary_max_scan"])
	assert.Equal(t, float64(10), actual["fragment_size"])
	assert.Equal(t, float64(50), actual["number_of_fragments"])
	assert.Equal(t, "1", actual["matched_fields"].([]interface{})[0])
	assert.Equal(t, "2", actual["matched_fields"].([]interface{})[1])
	assert.Equal(t, "order", actual["order"])
	assert.Equal(t, "fdsa", actual["type"])
}

func TestEmbedAndFieldDsl(t *testing.T) {
	highlight := NewHighlight().
		SetOptions(NewHighlightOpts().Tags("<div>", "</div>")).
		AddField("afield", NewHighlightOpts().Type("something"))

	actual, err := GetJson(highlight)
	actualField := actual["fields"].(map[string]interface{})["afield"].(map[string]interface{})

	assert.Equal(t, nil, err)
	assert.Equal(t, "<div>", actual["pre_tags"].([]interface{})[0])
	assert.Equal(t, "</div>", actual["post_tags"].([]interface{})[0])
	assert.Equal(t, "something", actualField["type"])
}

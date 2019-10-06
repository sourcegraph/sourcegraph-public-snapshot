package comments

import (
	"html"

	"github.com/microcosm-cc/bluemonday"
	"github.com/sourcegraph/sourcegraph/internal/markdown"
)

func ToBodyText(body string) string {
	// TODO!(sqs): this doesnt remove markdown formatting like `*`, just HTML tags
	return html.UnescapeString(bluemonday.StrictPolicy().Sanitize(body))
}

func ToBodyHTML(body string) string {
	return markdown.Render(body)
}

package graphqlbackend

type markdownResolver struct {
	text string
	html *string
}

func (m *markdownResolver) Text() string {
	return m.text
}

func (m *markdownResolver) HTML() *string {
	return m.html
}

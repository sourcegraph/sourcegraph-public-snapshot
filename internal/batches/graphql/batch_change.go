package graphql

type BatchChange struct {
	ID          string
	Namespace   Namespace
	Name        string
	Description string
	URL         string
}

package gcplogurl

type Query string

func (q Query) marshalURL(vs values) {
	vs.Set("query", string(q))
}

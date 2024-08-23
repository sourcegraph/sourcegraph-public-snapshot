package transmission

type Logger interface {
	// Printf accepts the same msg, args style as fmt.Printf().
	Printf(msg string, args ...interface{})
}

type nullLogger struct{}

// Printf swallows messages
func (n *nullLogger) Printf(msg string, args ...interface{}) {
	// nothing to see here.
}

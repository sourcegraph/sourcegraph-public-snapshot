package ci

var AspectWorkflows = struct {
	TestStepKey string

	QueueDefault string
	QueueSmall   string
}{
	TestStepKey:  "__main__::test",
	QueueDefault: "aspect-default",
	QueueSmall:   "aspect-small",
}

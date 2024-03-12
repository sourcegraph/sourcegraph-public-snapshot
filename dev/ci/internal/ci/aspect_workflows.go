package ci

var AspectWorkflows = struct {
	TestStepKey            string
	IntegrationTestStepKey string

	QueueDefault string
	QueueSmall   string
}{
	TestStepKey:            "__main__::test",
	IntegrationTestStepKey: "__main__::test_2",
	QueueDefault:           "aspect-default",
	QueueSmall:             "aspect-small",
}

package ci

var AspectWorkflows = struct {
	// TestStepKey is the key of the primary test step
	TestStepKey string
	// IntegrationTestStepKey is the key of the secondary test step where Integration and E2E tests are done
	IntegrationTestStepKey string

	// QueueDefault is the name of the default queue and uses a "big" machine. Jobs requiring big builds or big tests should use this queue
	QueueDefault string
	// QueueSmall is the name of the small queue and uses a "small" machine. Jobs that typically do not use bazel or use prebuilt binaries should use this queue
	QueueSmall string

	// AgentHealthCheckScript is the script that gets executed to check that the current agent is healthy
	AgentHealthCheckScript string
}{
	TestStepKey:            "__main__::test",
	IntegrationTestStepKey: "__main__::test_2",
	QueueDefault:           "aspect-default",
	QueueSmall:             "aspect-small",
}

func aspectBazelRC() (string, string) {
	path := "/tmp/aspect-generated.bazelrc"
	bazelRCCmd := "rosetta bazelrc > " + path

	return bazelRCCmd, path
}

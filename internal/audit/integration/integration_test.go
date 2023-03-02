package integration

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/sourcegraph/run"
	"github.com/stretchr/testify/assert"
)

func TestIntegration(t *testing.T) {
	if os.Getenv("BAZEL_TEST") == "1" {
		t.Skip("Don't run test due to running commands in bazel sandbox.")
	}

	// audit logs are logged under INFO severity
	t.Setenv("SRC_LOG_LEVEL", "info")

	// start sampling after 5 messages
	t.Setenv("SRC_LOG_SAMPLING_INITIAL", "5")

	// create 10 log messages, none of them should be sampled
	output, _ := run.Cmd(context.Background(), "go", "run", "./cmd/", "10").Run().String()

	logMessages := filterAuditLogs(output)
	if len(logMessages) == 0 {
		t.Fatal("no log output captured")
	}

	// capture all 10 despite the sampling setting (5)
	assert.Equal(t, 10, len(logMessages))
}

func filterAuditLogs(output string) []string {
	lines := strings.Split(output, "\n")
	var filtered []string
	for _, line := range lines {
		if strings.Contains(line, "{\"audit\"") {
			filtered = append(filtered, line)
		}
	}
	return filtered
}

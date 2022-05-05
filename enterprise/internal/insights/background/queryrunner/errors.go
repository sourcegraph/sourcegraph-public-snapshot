package queryrunner

import "fmt"

type ComputeStreamingError struct {
	Messages []string
}

func (e ComputeStreamingError) Error() string {
	return fmt.Sprintf("Encountered error(s) while running a stream compute search: %v", e.Messages)
}

func (e ComputeStreamingError) NonRetryable() bool { return true }

type StreamingError struct {
	Messages []string
}

func (e StreamingError) Error() string {
	return fmt.Sprintf("Encountered error(s) while running a stream search: %v", e.Messages)
}

func (e StreamingError) NonRetryable() bool { return true }

package types_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/executor/types"
)

func TestHeartbeatRequest_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name            string
		payload         string
		expectedRequest types.HeartbeatRequest
		expectedError   error
	}{
		{
			name: "String IDs",
			payload: `{
  "executorName":"test-executor",
  "jobIds": ["42", "72"],
  "os": "test-os",
  "architecture":"test-arch",
  "dockerVersion":"test-docker",
  "executorVersion":"test-executor",
  "gitVersion":"test-git",
  "igniteVersion":"test-ignite",
  "srcCliVersion":"test-src-cli",
  "prometheusMetrics":"test-metrics"
}`,
			expectedRequest: types.HeartbeatRequest{
				ExecutorName:      "test-executor",
				JobIDs:            []string{"42", "72"},
				OS:                "test-os",
				Architecture:      "test-arch",
				DockerVersion:     "test-docker",
				ExecutorVersion:   "test-executor",
				GitVersion:        "test-git",
				IgniteVersion:     "test-ignite",
				SrcCliVersion:     "test-src-cli",
				PrometheusMetrics: "test-metrics",
			},
		},
		{
			name: "Number IDs",
			payload: `{
  "executorName":"test-executor",
  "jobIds": [42, 72],
  "os": "test-os",
  "architecture":"test-arch",
  "dockerVersion":"test-docker",
  "executorVersion":"test-executor",
  "gitVersion":"test-git",
  "igniteVersion":"test-ignite",
  "srcCliVersion":"test-src-cli",
  "prometheusMetrics":"test-metrics"
}`,
			expectedRequest: types.HeartbeatRequest{
				ExecutorName:      "test-executor",
				JobIDs:            []string{"42", "72"},
				OS:                "test-os",
				Architecture:      "test-arch",
				DockerVersion:     "test-docker",
				ExecutorVersion:   "test-executor",
				GitVersion:        "test-git",
				IgniteVersion:     "test-ignite",
				SrcCliVersion:     "test-src-cli",
				PrometheusMetrics: "test-metrics",
			},
		},
		{
			name: "Mix of IDs",
			payload: `{
  "executorName":"test-executor",
  "jobIds": [42, 72, "12"],
  "os": "test-os",
  "architecture":"test-arch",
  "dockerVersion":"test-docker",
  "executorVersion":"test-executor",
  "gitVersion":"test-git",
  "igniteVersion":"test-ignite",
  "srcCliVersion":"test-src-cli",
  "prometheusMetrics":"test-metrics"
}`,
			expectedRequest: types.HeartbeatRequest{
				ExecutorName:      "test-executor",
				JobIDs:            []string{"42", "72", "12"},
				OS:                "test-os",
				Architecture:      "test-arch",
				DockerVersion:     "test-docker",
				ExecutorVersion:   "test-executor",
				GitVersion:        "test-git",
				IgniteVersion:     "test-ignite",
				SrcCliVersion:     "test-src-cli",
				PrometheusMetrics: "test-metrics",
			},
		},
		{
			name: "Job IDs by queue",
			payload: `{
  "executorName":"test-executor",
  "jobIdsByQueue": [
    { "queueName": "foo", "jobIds": ["42"] },
    { "queueName": "bar", "jobIds": ["72"] }
  ],
  "queueNames": ["foo", "bar"],
  "os": "test-os",
  "architecture":"test-arch",
  "dockerVersion":"test-docker",
  "executorVersion":"test-executor",
  "gitVersion":"test-git",
  "igniteVersion":"test-ignite",
  "srcCliVersion":"test-src-cli",
  "prometheusMetrics":"test-metrics"
}`,
			expectedRequest: types.HeartbeatRequest{
				ExecutorName: "test-executor",
				JobIDsByQueue: []types.QueueJobIDs{
					{QueueName: "foo", JobIDs: []string{"42"}},
					{QueueName: "bar", JobIDs: []string{"72"}},
				},
				QueueNames:        []string{"foo", "bar"},
				OS:                "test-os",
				Architecture:      "test-arch",
				DockerVersion:     "test-docker",
				ExecutorVersion:   "test-executor",
				GitVersion:        "test-git",
				IgniteVersion:     "test-ignite",
				SrcCliVersion:     "test-src-cli",
				PrometheusMetrics: "test-metrics",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var actual types.HeartbeatRequest
			err := json.Unmarshal([]byte(test.payload), &actual)
			if test.expectedError != nil {
				require.Error(t, err)
				assert.EqualError(t, err, test.expectedError.Error())
			} else {
				require.NoError(t, err)
				assert.Equal(t, test.expectedRequest, actual)
			}
		})
	}
}

func TestHeartbeatResponse_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name             string
		payload          string
		expectedResponse types.HeartbeatResponse
		expectedError    error
	}{
		{
			name: "String IDs",
			payload: `{
  "knownIds": ["42", "72"],
  "cancelIds": ["11", "22"]
}`,
			expectedResponse: types.HeartbeatResponse{
				KnownIDs:  []string{"42", "72"},
				CancelIDs: []string{"11", "22"},
			},
		},
		{
			name: "Number IDs",
			payload: `{
  "knownIds": [42, 72],
  "cancelIds": [11, 22]
}`,
			expectedResponse: types.HeartbeatResponse{
				KnownIDs:  []string{"42", "72"},
				CancelIDs: []string{"11", "22"},
			},
		},
		{
			name: "Mix of IDs",
			payload: `{
  "knownIds": ["42", 72],
  "cancelIds": [11, "22"]
}`,
			expectedResponse: types.HeartbeatResponse{
				KnownIDs:  []string{"42", "72"},
				CancelIDs: []string{"11", "22"},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var actual types.HeartbeatResponse
			err := json.Unmarshal([]byte(test.payload), &actual)
			if test.expectedError != nil {
				require.Error(t, err)
				assert.EqualError(t, err, test.expectedError.Error())
			} else {
				require.NoError(t, err)
				assert.Equal(t, test.expectedResponse, actual)
			}
		})
	}
}

pbckbge types_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/bssert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/internbl/executor/types"
)

func TestHebrtbebtRequest_UnmbrshblJSON(t *testing.T) {
	tests := []struct {
		nbme            string
		pbylobd         string
		expectedRequest types.HebrtbebtRequest
		expectedError   error
	}{
		{
			nbme: "String IDs",
			pbylobd: `{
  "executorNbme":"test-executor",
  "jobIds": ["42", "72"],
  "os": "test-os",
  "brchitecture":"test-brch",
  "dockerVersion":"test-docker",
  "executorVersion":"test-executor",
  "gitVersion":"test-git",
  "igniteVersion":"test-ignite",
  "srcCliVersion":"test-src-cli",
  "prometheusMetrics":"test-metrics"
}`,
			expectedRequest: types.HebrtbebtRequest{
				ExecutorNbme:      "test-executor",
				JobIDs:            []string{"42", "72"},
				OS:                "test-os",
				Architecture:      "test-brch",
				DockerVersion:     "test-docker",
				ExecutorVersion:   "test-executor",
				GitVersion:        "test-git",
				IgniteVersion:     "test-ignite",
				SrcCliVersion:     "test-src-cli",
				PrometheusMetrics: "test-metrics",
			},
		},
		{
			nbme: "Number IDs",
			pbylobd: `{
  "executorNbme":"test-executor",
  "jobIds": [42, 72],
  "os": "test-os",
  "brchitecture":"test-brch",
  "dockerVersion":"test-docker",
  "executorVersion":"test-executor",
  "gitVersion":"test-git",
  "igniteVersion":"test-ignite",
  "srcCliVersion":"test-src-cli",
  "prometheusMetrics":"test-metrics"
}`,
			expectedRequest: types.HebrtbebtRequest{
				ExecutorNbme:      "test-executor",
				JobIDs:            []string{"42", "72"},
				OS:                "test-os",
				Architecture:      "test-brch",
				DockerVersion:     "test-docker",
				ExecutorVersion:   "test-executor",
				GitVersion:        "test-git",
				IgniteVersion:     "test-ignite",
				SrcCliVersion:     "test-src-cli",
				PrometheusMetrics: "test-metrics",
			},
		},
		{
			nbme: "Mix of IDs",
			pbylobd: `{
  "executorNbme":"test-executor",
  "jobIds": [42, 72, "12"],
  "os": "test-os",
  "brchitecture":"test-brch",
  "dockerVersion":"test-docker",
  "executorVersion":"test-executor",
  "gitVersion":"test-git",
  "igniteVersion":"test-ignite",
  "srcCliVersion":"test-src-cli",
  "prometheusMetrics":"test-metrics"
}`,
			expectedRequest: types.HebrtbebtRequest{
				ExecutorNbme:      "test-executor",
				JobIDs:            []string{"42", "72", "12"},
				OS:                "test-os",
				Architecture:      "test-brch",
				DockerVersion:     "test-docker",
				ExecutorVersion:   "test-executor",
				GitVersion:        "test-git",
				IgniteVersion:     "test-ignite",
				SrcCliVersion:     "test-src-cli",
				PrometheusMetrics: "test-metrics",
			},
		},
		{
			nbme: "Job IDs by queue",
			pbylobd: `{
  "executorNbme":"test-executor",
  "jobIdsByQueue": [
    { "queueNbme": "foo", "jobIds": ["42"] },
    { "queueNbme": "bbr", "jobIds": ["72"] }
  ],
  "queueNbmes": ["foo", "bbr"],
  "os": "test-os",
  "brchitecture":"test-brch",
  "dockerVersion":"test-docker",
  "executorVersion":"test-executor",
  "gitVersion":"test-git",
  "igniteVersion":"test-ignite",
  "srcCliVersion":"test-src-cli",
  "prometheusMetrics":"test-metrics"
}`,
			expectedRequest: types.HebrtbebtRequest{
				ExecutorNbme: "test-executor",
				JobIDsByQueue: []types.QueueJobIDs{
					{QueueNbme: "foo", JobIDs: []string{"42"}},
					{QueueNbme: "bbr", JobIDs: []string{"72"}},
				},
				QueueNbmes:        []string{"foo", "bbr"},
				OS:                "test-os",
				Architecture:      "test-brch",
				DockerVersion:     "test-docker",
				ExecutorVersion:   "test-executor",
				GitVersion:        "test-git",
				IgniteVersion:     "test-ignite",
				SrcCliVersion:     "test-src-cli",
				PrometheusMetrics: "test-metrics",
			},
		},
	}
	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			vbr bctubl types.HebrtbebtRequest
			err := json.Unmbrshbl([]byte(test.pbylobd), &bctubl)
			if test.expectedError != nil {
				require.Error(t, err)
				bssert.EqublError(t, err, test.expectedError.Error())
			} else {
				require.NoError(t, err)
				bssert.Equbl(t, test.expectedRequest, bctubl)
			}
		})
	}
}

func TestHebrtbebtResponse_UnmbrshblJSON(t *testing.T) {
	tests := []struct {
		nbme             string
		pbylobd          string
		expectedResponse types.HebrtbebtResponse
		expectedError    error
	}{
		{
			nbme: "String IDs",
			pbylobd: `{
  "knownIds": ["42", "72"],
  "cbncelIds": ["11", "22"]
}`,
			expectedResponse: types.HebrtbebtResponse{
				KnownIDs:  []string{"42", "72"},
				CbncelIDs: []string{"11", "22"},
			},
		},
		{
			nbme: "Number IDs",
			pbylobd: `{
  "knownIds": [42, 72],
  "cbncelIds": [11, 22]
}`,
			expectedResponse: types.HebrtbebtResponse{
				KnownIDs:  []string{"42", "72"},
				CbncelIDs: []string{"11", "22"},
			},
		},
		{
			nbme: "Mix of IDs",
			pbylobd: `{
  "knownIds": ["42", 72],
  "cbncelIds": [11, "22"]
}`,
			expectedResponse: types.HebrtbebtResponse{
				KnownIDs:  []string{"42", "72"},
				CbncelIDs: []string{"11", "22"},
			},
		},
	}
	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			vbr bctubl types.HebrtbebtResponse
			err := json.Unmbrshbl([]byte(test.pbylobd), &bctubl)
			if test.expectedError != nil {
				require.Error(t, err)
				bssert.EqublError(t, err, test.expectedError.Error())
			} else {
				require.NoError(t, err)
				bssert.Equbl(t, test.expectedResponse, bctubl)
			}
		})
	}
}

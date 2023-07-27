package messagesize

import (
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"

	newspb "github.com/sourcegraph/sourcegraph/internal/grpc/testprotos/news/v1"
)

func TestObserver(t *testing.T) {
	testCases := []struct {
		name     string
		messages []proto.Message
	}{
		{
			name: "single message",
			messages: []proto.Message{&newspb.BinaryAttachment{
				Name: "data1",
				Data: []byte("sample data"),
			}},
		},
		{
			name: "multiple messages",
			messages: []proto.Message{
				&newspb.BinaryAttachment{
					Name: "data1",
					Data: []byte("sample data"),
				},
				&newspb.KeyValueAttachment{
					Name: "data2",
					Data: map[string]string{
						"key1": "value1",
						"key2": "value2",
					},
				},
			}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var singleMessageSizes []uint64
			var totalSize uint64

			// Create a new observer with custom onSingleFunc and onFinishFunc
			obs := &messageSizeObserver{
				onSingleFunc: func(messageSizeBytes uint64) {
					singleMessageSizes = append(singleMessageSizes, messageSizeBytes)
				},
				onFinishFunc: func(totalSizeBytes uint64) {
					totalSize = totalSizeBytes
				},
			}

			// Call ObserveSingle for each message
			for _, msg := range tc.messages {
				obs.Observe(msg)
			}

			// Check that the singleMessageSizes are correct
			for i, msg := range tc.messages {
				expectedSize := uint64(proto.Size(msg))
				require.Equal(t, expectedSize, singleMessageSizes[i])
			}

			// Call FinishRPC
			obs.FinishRPC()

			// Check that the totalSize is correct
			expectedTotalSize := uint64(0)
			for _, size := range singleMessageSizes {
				expectedTotalSize += size
			}
			require.EqualValues(t, expectedTotalSize, totalSize)
		})
	}
}

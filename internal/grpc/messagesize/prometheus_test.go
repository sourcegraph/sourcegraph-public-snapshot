package messagesize

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	newspb "github.com/sourcegraph/sourcegraph/internal/grpc/testprotos/news/v1"
)

var (
	binaryMessage = &newspb.BinaryAttachment{
		Name: "data",
		Data: []byte(strings.Repeat("x", 1*1024*1024)),
	}

	keyValueMessage = &newspb.KeyValueAttachment{
		Name: "data",
		Data: map[string]string{
			"key1": strings.Repeat("x", 1*1024*1024),
			"key2": "value2",
		},
	}

	articleMessage = &newspb.Article{
		Author:  "author",
		Date:    &timestamppb.Timestamp{Seconds: 1234567890},
		Title:   "title",
		Content: "content",
		Status:  newspb.Article_STATUS_PUBLISHED,
		Attachments: []*newspb.Attachment{
			{Contents: &newspb.Attachment_KeyValueAttachment{KeyValueAttachment: keyValueMessage}},
			{Contents: &newspb.Attachment_KeyValueAttachment{KeyValueAttachment: keyValueMessage}},
			{Contents: &newspb.Attachment_BinaryAttachment{BinaryAttachment: binaryMessage}},
			{Contents: &newspb.Attachment_BinaryAttachment{BinaryAttachment: binaryMessage}},
		},
	}
)

func BenchmarkObserverBinary(b *testing.B) {
	o := messageSizeObserver{
		onSingleFunc: func(messageSizeBytes uint64) {},
		onFinishFunc: func(totalSizeBytes uint64) {},
	}

	benchmarkObserver(b, &o, binaryMessage)
}

func BenchmarkObserverKeyValue(b *testing.B) {
	o := messageSizeObserver{
		onSingleFunc: func(messageSizeBytes uint64) {},
		onFinishFunc: func(totalSizeBytes uint64) {},
	}

	benchmarkObserver(b, &o, keyValueMessage)
}

func BenchmarkObserverArticle(b *testing.B) {
	o := messageSizeObserver{
		onSingleFunc: func(messageSizeBytes uint64) {},
		onFinishFunc: func(totalSizeBytes uint64) {},
	}

	benchmarkObserver(b, &o, articleMessage)
}

func benchmarkObserver(b *testing.B, observer *messageSizeObserver, message proto.Message) {
	b.ReportAllocs()

	for n := 0; n < b.N; n++ {
		observer.Observe(message)
	}

	observer.FinishRPC()
}

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

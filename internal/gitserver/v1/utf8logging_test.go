package v1

import (
	"testing"

	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/proto"

	newspb "github.com/sourcegraph/sourcegraph/internal/grpc/testprotos/news"
)

func TestFindNonUTF8StringFields(t *testing.T) {
	// Create instances of the BinaryAttachment and KeyValueAttachment messages
	invalidBinaryAttachment := &newspb.BinaryAttachment{
		Name: "inval\x80id_binary",
		Data: []byte("sample data"),
	}

	invalidKeyValueAttachment := &newspb.KeyValueAttachment{
		Name: "inval\x80id_key_value",
		Data: map[string]string{
			"key1": "value1",
			"key2": "inval\x80id_value",
		},
	}

	// Create a sample Article message with invalid UTF-8 strings
	article := &newspb.Article{
		Author:  "inval\x80id_author",
		Date:    &timestamp.Timestamp{Seconds: 1234567890},
		Title:   "valid_title",
		Content: "valid_content",
		Status:  newspb.Article_PUBLISHED,
		Attachments: []*newspb.Attachment{
			{Contents: &newspb.Attachment_BinaryAttachment{BinaryAttachment: invalidBinaryAttachment}},
			{Contents: &newspb.Attachment_KeyValueAttachment{KeyValueAttachment: invalidKeyValueAttachment}},
		},
	}

	tests := []struct {
		name          string
		message       proto.Message
		expectedPaths []string
	}{
		{
			name:    "Article with invalid UTF-8 strings",
			message: article,
			expectedPaths: []string{
				"author",
				"attachments[0].binary_attachment.name",
				"attachments[1].key_value_attachment.name",
				`attachments[1].key_value_attachment.data["key2"]`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			invalidFields, err := findNonUTF8StringFields(tt.message)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if diff := cmp.Diff(tt.expectedPaths, invalidFields); diff != "" {
				t.Fatalf("unexpected invalid fields (-want +got):\n%s", diff)
			}
		})
	}
}

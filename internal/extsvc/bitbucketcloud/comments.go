package bitbucketcloud

import (
	"encoding/json"
	"time"
)

type Comment struct {
	ID        int64          `json:"id"`
	CreatedOn time.Time      `json:"created_on"`
	UpdatedOn time.Time      `json:"updated_on"`
	Content   RenderedMarkup `json:"content"`
	User      User           `json:"user"`
	Deleted   bool           `json:"deleted"`
	Parent    *Comment       `json:"parent,omitempty"`
	Inline    *CommentInline `json:"inline,omitempty"`
	Links     Links          `json:"links"`
}

type CommentInline struct {
	To   int64  `json:"to,omitempty"`
	From int64  `json:"from,omitempty"`
	Path string `json:"path"`
}

type CommentInput struct {
	Content string
}

var _ json.Marshaler = &CommentInput{}

func (ci *CommentInput) MarshalJSON() ([]byte, error) {
	type content struct {
		Raw string `json:"raw"`
	}
	type comment struct {
		Content content `json:"content"`
	}

	return json.Marshal(&comment{
		Content: content{
			Raw: ci.Content,
		},
	})
}

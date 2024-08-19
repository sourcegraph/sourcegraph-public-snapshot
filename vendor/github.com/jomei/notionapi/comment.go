package notionapi

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"
)

type CommentID string

func (cID CommentID) String() string {
	return string(cID)
}

type CommentService interface {
	Create(ctx context.Context, request *CommentCreateRequest) (*Comment, error)
	Get(context.Context, BlockID, *Pagination) (*CommentQueryResponse, error)
}

type CommentClient struct {
	apiClient *Client
}

// Creates a comment in a page or existing discussion thread.
//
// There are two locations you can add a new comment to:
// 1. A page
// 2. An existing discussion thread
//
// If the intention is to add a new comment to a page, a parent object must be
// provided in the body params. Alternatively, if a new comment is being added
// to an existing discussion thread, the discussion_id string must be provided
// in the body params. Exactly one of these parameters must be provided.
//
// See https://developers.notion.com/reference/create-a-comment
func (cc *CommentClient) Create(ctx context.Context, requestBody *CommentCreateRequest) (*Comment, error) {
	res, err := cc.apiClient.request(ctx, http.MethodPost, "comments", nil, requestBody)
	if err != nil {
		return nil, err
	}

	defer func() {
		if errClose := res.Body.Close(); errClose != nil {
			log.Println("failed to close body, should never happen")
		}
	}()

	var response Comment
	err = json.NewDecoder(res.Body).Decode(&response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

// CommentCreateRequest represents the request body for CommentClient.Create.
type CommentCreateRequest struct {
	Parent       Parent       `json:"parent,omitempty"`
	DiscussionID DiscussionID `json:"discussion_id,omitempty"`
	RichText     []RichText   `json:"rich_text"`
}

// Retrieves a list of un-resolved Comment objects from a page or block.
//
// See https://developers.notion.com/reference/retrieve-a-comment
func (cc *CommentClient) Get(ctx context.Context, id BlockID, pagination *Pagination) (*CommentQueryResponse, error) {
	queryParams := map[string]string{}
	if pagination != nil {
		queryParams = pagination.ToQuery()
	}

	queryParams["block_id"] = id.String()

	res, err := cc.apiClient.request(ctx, http.MethodGet, "comments", queryParams, nil)
	if err != nil {
		return nil, err
	}

	defer func() {
		if errClose := res.Body.Close(); errClose != nil {
			log.Println("failed to close body, should never happen")
		}
	}()

	var response CommentQueryResponse

	err = json.NewDecoder(res.Body).Decode(&response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

type DiscussionID string

func (dID DiscussionID) String() string {
	return string(dID)
}

type Comment struct {
	Object         ObjectType   `json:"object"`
	ID             ObjectID     `json:"id"`
	DiscussionID   DiscussionID `json:"discussion_id"`
	CreatedTime    time.Time    `json:"created_time"`
	LastEditedTime time.Time    `json:"last_edited_time"`
	CreatedBy      User         `json:"created_by,omitempty"`
	RichText       []RichText   `json:"rich_text"`
	Parent         Parent       `json:"parent"`
}

type CommentQueryResponse struct {
	Object     ObjectType `json:"object"`
	Results    []Comment  `json:"results"`
	HasMore    bool       `json:"has_more"`
	NextCursor Cursor     `json:"next_cursor"`
}

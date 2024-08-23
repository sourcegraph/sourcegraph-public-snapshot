package notionapi

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type SearchService interface {
	Do(context.Context, *SearchRequest) (*SearchResponse, error)
}

type SearchClient struct {
	apiClient *Client
}

// Searches all parent or child pages and databases that have been shared with
// an integration.
//
// Returns all pages or databases, excluding duplicated linked databases, that
// have titles that include the query param. If no query param is provided, then
// the response contains all pages or databases that have been shared with the
// integration. The results adhere to any limitations related to an integration’s
// capabilities.

// To limit the request to search only pages or to search only databases, use
// the filter param.
//
// See https://developers.notion.com/reference/post-search
func (sc *SearchClient) Do(ctx context.Context, request *SearchRequest) (*SearchResponse, error) {
	res, err := sc.apiClient.request(ctx, http.MethodPost, "search", nil, request)
	if err != nil {
		return nil, err
	}

	defer func() {
		if errClose := res.Body.Close(); errClose != nil {
			log.Println("failed to close body, should never happen")
		}
	}()

	var response SearchResponse
	err = json.NewDecoder(res.Body).Decode(&response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

type SearchRequest struct {
	// The text that the API compares page and database titles against.
	Query string `json:"query,omitempty"`
	// A set of criteria, direction and timestamp keys, that orders the results.
	// The only supported timestamp value is "last_edited_time". Supported
	// direction values are "ascending" and "descending". If sort is not provided,
	// then the most recently edited results are returned first.
	Sort *SortObject `json:"sort,omitempty"`
	// A set of criteria, value and property keys, that limits the results to
	// either only pages or only databases. Possible value values are "page" or
	// "database". The only supported property value is "object".
	Filter SearchFilter `json:"filter,omitempty"`
	// A cursor value returned in a previous response that If supplied, limits the
	// response to results starting after the cursor. If not supplied, then the
	// first page of results is returned. Refer to pagination for more details.
	StartCursor Cursor `json:"start_cursor,omitempty"`
	// The number of items from the full list to include in the response. Maximum: 100.
	PageSize int `json:"page_size,omitempty"`
}

type SearchResponse struct {
	Object     ObjectType `json:"object"`
	Results    []Object   `json:"results"`
	HasMore    bool       `json:"has_more"`
	NextCursor Cursor     `json:"next_cursor"`
}

func (sr *SearchResponse) UnmarshalJSON(data []byte) error {
	var tmp struct {
		Object     ObjectType    `json:"object"`
		Results    []interface{} `json:"results"`
		HasMore    bool          `json:"has_more"`
		NextCursor Cursor        `json:"next_cursor"`
	}

	err := json.Unmarshal(data, &tmp)
	if err != nil {
		return err
	}
	objects := make([]Object, len(tmp.Results))
	for i, rawObject := range tmp.Results {
		var o Object
		switch rawObject.(map[string]interface{})["object"].(string) {
		case ObjectTypeDatabase.String():
			o = &Database{}
		case ObjectTypePage.String():
			o = &Page{}
		default:
			return fmt.Errorf("unsupported object type %s", rawObject.(map[string]interface{})["object"].(string))
		}
		j, err := json.Marshal(rawObject)
		if err != nil {
			return err
		}

		err = json.Unmarshal(j, o)
		if err != nil {
			return err
		}
		objects[i] = o
	}

	*sr = SearchResponse{
		Object:     tmp.Object,
		Results:    objects,
		HasMore:    tmp.HasMore,
		NextCursor: tmp.NextCursor,
	}

	return nil
}

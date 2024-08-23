package notionapi

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

type PageID string

func (pID PageID) String() string {
	return string(pID)
}

type PageService interface {
	Create(context.Context, *PageCreateRequest) (*Page, error)
	Get(context.Context, PageID) (*Page, error)
	Update(context.Context, PageID, *PageUpdateRequest) (*Page, error)
}

type PageClient struct {
	apiClient *Client
}

// Creates a new page that is a child of an existing page or database.
//
// If the new page is a child of an existing page,title is the only valid
// property in the properties body param.
//
// If the new page is a child of an existing database, the keys of the
// properties object body param must match the parent database's properties.
//
// This endpoint can be used to create a new page with or without content using
// the children option. To add content to a page after creating it, use the
// Append block children endpoint.
//
// Returns a new page object.
//
// See https://developers.notion.com/reference/post-page
func (pc *PageClient) Create(ctx context.Context, requestBody *PageCreateRequest) (*Page, error) {
	res, err := pc.apiClient.request(ctx, http.MethodPost, "pages", nil, requestBody)
	if err != nil {
		return nil, err
	}

	defer func() {
		if errClose := res.Body.Close(); errClose != nil {
			log.Println("failed to close body, should never happen")
		}
	}()

	return handlePageResponse(res)
}

// PageCreateRequest represents the request body for PageClient.Create.
type PageCreateRequest struct {
	// The parent page or database where the new page is inserted, represented as
	// a JSON object with a page_id or database_id key, and the corresponding ID.
	Parent Parent `json:"parent"`
	// The values of the page’s properties. If the parent is a database, then the
	// schema must match the parent database’s properties. If the parent is a page,
	// then the only valid object key is title.
	Properties Properties `json:"properties"`
	// The content to be rendered on the new page, represented as an array of
	// block objects.
	Children []Block `json:"children,omitempty"`
	// The icon of the new page. Either an emoji object or an external file object.
	Icon *Icon `json:"icon,omitempty"`
	// The cover image of the new page, represented as a file object.
	Cover *Image `json:"cover,omitempty"`
}

// Retrieves a Page object using the ID specified.
//
// Responses contains page properties, not page content. To fetch page content,
// use the Retrieve block children endpoint.
//
// Page properties are limited to up to 25 references per page property. To
// retrieve data related to properties that have more than 25 references, use
// the Retrieve a page property endpoint.
//
// See https://developers.notion.com/reference/get-page
func (pc *PageClient) Get(ctx context.Context, id PageID) (*Page, error) {
	res, err := pc.apiClient.request(ctx, http.MethodGet, fmt.Sprintf("pages/%s", id.String()), nil, nil)
	if err != nil {
		return nil, err
	}

	defer func() {
		if errClose := res.Body.Close(); errClose != nil {
			log.Println("failed to close body, should never happen")
		}
	}()

	return handlePageResponse(res)
}

// Updates the properties of a page in a database. The properties body param of
// this endpoint can only be used to update the properties of a page that is a
// child of a database. The page’s properties schema must match the parent
// database’s properties.
//
// This endpoint can be used to update any page icon or cover, and can be used
// to archive or restore any page.
//
// To add page content instead of page properties, use the append block children
// endpoint. The page_id can be passed as the block_id when adding block
// children to the page.
//
// Returns the updated page object.
//
// See https://developers.notion.com/reference/patch-page
func (pc *PageClient) Update(ctx context.Context, id PageID, request *PageUpdateRequest) (*Page, error) {
	res, err := pc.apiClient.request(ctx, http.MethodPatch, fmt.Sprintf("pages/%s", id.String()), nil, request)
	if err != nil {
		return nil, err
	}

	defer func() {
		if errClose := res.Body.Close(); errClose != nil {
			log.Println("failed to close body, should never happen")
		}
	}()

	return handlePageResponse(res)
}

// PageUpdateRequest represents the request body for PageClient.Update.
type PageUpdateRequest struct {
	// The property values to update for the page. The keys are the names or IDs
	// of the property and the values are property values. If a page property ID
	// is not included, then it is not changed.
	Properties Properties `json:"properties"`
	// Whether the page is archived (deleted). Set to true to archive a page. Set
	// to false to un-archive (restore) a page.
	Archived bool `json:"archived"`
	// A page icon for the page. Supported types are external file object or emoji
	// object.
	Icon *Icon `json:"icon,omitempty"`
	// A cover image for the page. Only external file objects are supported.
	Cover *Image `json:"cover,omitempty"`
}

// The Page object contains the page property values of a single Notion page.
//
// See https://developers.notion.com/reference/page
type Page struct {
	Object         ObjectType `json:"object"`
	ID             ObjectID   `json:"id"`
	CreatedTime    time.Time  `json:"created_time"`
	LastEditedTime time.Time  `json:"last_edited_time"`
	CreatedBy      User       `json:"created_by,omitempty"`
	LastEditedBy   User       `json:"last_edited_by,omitempty"`
	Archived       bool       `json:"archived"`
	Properties     Properties `json:"properties"`
	Parent         Parent     `json:"parent"`
	URL            string     `json:"url"`
	PublicURL      string     `json:"public_url"`
	Icon           *Icon      `json:"icon,omitempty"`
	Cover          *Image     `json:"cover,omitempty"`
}

func (p *Page) GetObject() ObjectType {
	return p.Object
}

type ParentType string

// Pages, databases, and blocks are either located inside other pages,
// databases, and blocks, or are located at the top level of a workspace. This
// location is known as the "parent". Parent information is represented by a
// consistent parent object throughout the API.
//
// See https://developers.notion.com/reference/parent-object
type Parent struct {
	Type       ParentType `json:"type,omitempty"`
	PageID     PageID     `json:"page_id,omitempty"`
	DatabaseID DatabaseID `json:"database_id,omitempty"`
	BlockID    BlockID    `json:"block_id,omitempty"`
	Workspace  bool       `json:"workspace,omitempty"`
}

func handlePageResponse(res *http.Response) (*Page, error) {
	var response Page
	err := json.NewDecoder(res.Body).Decode(&response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

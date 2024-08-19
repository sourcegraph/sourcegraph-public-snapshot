package responses

import (
	"encoding/json"

	"github.com/uber/gonduit/entities"
	"github.com/uber/gonduit/util"
)

// TransactionSearchResponse contains fields that are in server
// response to transaction.search.
type TransactionSearchResponse struct {
	// Data contains search results.
	Data []*TransactionSearchResponseItem `json:"data"`

	// Curson contains paging data.
	Cursor entities.Cursor `json:"cursor,omitempty"`
}

// TransactionSearchResponseItem contains information about a
// particular search result.
type TransactionSearchResponseItem struct {
	ResponseObject
	Fields       TransactionSearchResponseItemFields    `json:"fields"`
	AuthorPHID   string                                 `json:"authorPHID"`
	ObjectPHID   string                                 `json:"objectPHID"`
	GroupID      string                                 `json:"groupID"`
	DateCreated  util.UnixTimestamp                     `json:"dateCreated"`
	DateModified util.UnixTimestamp                     `json:"dateModified"`
	Comments     []TransactionSearchResponseItemComment `json:"comments"`
}

// TransactionSearchResponseItemFields is a collection of object
// fields.
type TransactionSearchResponseItemFields struct {
	// make sure to update transactionSearchResponseItemFieldsDecoded and unmarshaller as well
	Old         string                                         `json:"old"`
	New         string                                         `json:"new"`
	Operations  []TransactionSearchResponseItemFieldsOperation `json:"operations"`
	CommitPHIDs []string                                       `json:"commitPHIDs"`
}

// special struct to fix issue when php return [] as empty "struct" which causes golang
// decoder to crash...
type transactionSearchResponseItemFieldsDecoded struct {
	Old         string                                         `json:"old"`
	New         string                                         `json:"new"`
	Operations  []TransactionSearchResponseItemFieldsOperation `json:"operations"`
	CommitPHIDs []string                                       `json:"commitPHIDs"`
}

// TransactionSearchResponseItemComment is transaction comment
type TransactionSearchResponseItemComment struct {
	ID           int                                  `json:"id"`
	PHID         string                               `json:"phid"`
	Version      int                                  `json:"version"`
	AuthorPHID   string                               `json:"authorPHID"`
	DateCreated  util.UnixTimestamp                   `json:"dateCreated"`
	DateModified util.UnixTimestamp                   `json:"dateModified"`
	Removed      bool                                 `json:"removed"`
	Content      TransactionSearchResponseItemContent `json:"content"`
}

// TransactionSearchResponseItemContent is transaction comment content
type TransactionSearchResponseItemContent struct {
	Raw string `json:"raw"`
}

// TransactionSearchResponseItemFielsOperation is collect of transaction field
// operations
type TransactionSearchResponseItemFieldsOperation struct {
	Operation  string `json:"operation"`
	PHID       string `json:"phid"`
	OldStatus  string `json:"oldStatus"`
	NewStatus  string `json:"newStatus"`
	IsBlocking bool   `json:"isBlocking"`
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (t *TransactionSearchResponseItemFields) UnmarshalJSON(data []byte) (err error) {
	if len(data) == 2 && string(data) == "[]" {
		return nil
	}
	res := transactionSearchResponseItemFieldsDecoded{}
	err = json.Unmarshal(data, &res)
	t.Old = res.Old
	t.New = res.New
	t.Operations = res.Operations
	t.CommitPHIDs = res.CommitPHIDs
	return err
}

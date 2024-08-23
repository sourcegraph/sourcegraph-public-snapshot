package entities

import "github.com/uber/gonduit/util"

// ManiphestTaskTranscation represents a single task's transcation on Maniphest.
type ManiphestTaskTranscation struct {
	TaskID          string             `json:"taskID"`
	TransactionID   string             `json:"transactionID"`
	TransactionPHID string             `json:"transactionPHID"`
	TransactionType string             `json:"transactionType"`
	OldValue        interface{}        `json:"oldValue"`
	NewValue        interface{}        `json:"newValue"`
	Comments        string             `json:"comments"`
	AuthorPHID      string             `json:"authorPHID"`
	DateCreated     util.UnixTimestamp `json:"dateCreated"`
}

package alert

import (
	"github.com/opsgenie/opsgenie-go-sdk-v2/client"
	"time"
)

type Alert struct {
	Seen           bool        `json:"seen,omitempty"`
	Id             string      `json:"id,omitempty"`
	TinyID         string      `json:"tinyId,omitempty"`
	Alias          string      `json:"alias,omitempty"`
	Message        string      `json:"message,omitempty"`
	Status         string      `json:"status,omitempty"`
	Acknowledged   bool        `json:"acknowledged,omitempty"`
	IsSeen         bool        `json:"isSeen,omitempty"`
	Tags           []string    `json:"tags,omitempty"`
	Snoozed        bool        `json:"snoozed,omitempty"`
	SnoozedUntil   time.Time   `json:"snoozedUntil,omitempty"`
	Count          int         `json:"count,omitempty"`
	LastOccurredAt time.Time   `json:"lastOccurredAt,omitempty"`
	CreatedAt      time.Time   `json:"createdAt,omitempty"`
	UpdatedAt      time.Time   `json:"updatedAt,omitempty"`
	Source         string      `json:"source,omitempty"`
	Owner          string      `json:"owner,omitempty"`
	Priority       Priority    `json:"priority,omitempty"`
	Responders     []Responder `json:"responders"`
	Integration    Integration `json:"integration,omitempty"`
	Report         Report      `json:"report,omitempty"`
	OwnerTeamID    string      `json:"ownerTeamId,omitempty"`
}

type Integration struct {
	Id   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
	Type string `json:"type,omitempty"`
}

type ListAlertResult struct {
	client.ResultMetadata
	Alerts []Alert `json:"data"`
}

type RequestStatusResult struct {
	client.ResultMetadata
	IsSuccess     bool      `json:"isSuccess,omitempty"`
	Action        string    `json:"action,omitempty"`
	ProcessedAt   time.Time `json:"processedAt,omitempty"`
	IntegrationId string    `json:"integrationId,omitempty"`
	Status        string    `json:"status,omitempty"`
	AlertID       string    `json:"alertId,omitempty"`
	Alias         string    `json:"alias,omitempty"`
}

type SavedSearchResult struct {
	client.ResultMetadata
	Id   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

type GetAlertResult struct {
	client.ResultMetadata
	Id             string            `json:"id,omitempty"`
	TinyId         string            `json:"tinyId,omitempty"`
	Alias          string            `json:"alias,omitempty"`
	Message        string            `json:"message,omitempty"`
	Status         string            `json:"status,omitempty"`
	Acknowledged   bool              `json:"acknowledged,omitempty"`
	IsSeen         bool              `json:"isSeen,omitempty"`
	Tags           []string          `json:"tags,omitempty"`
	Snoozed        bool              `json:"snoozed,omitempty"`
	SnoozedUntil   time.Time         `json:"snoozedUntil,omitempty"`
	Count          int               `json:"count,omitempty"`
	LastOccurredAt time.Time         `json:"lastOccurredAt,omitempty"`
	CreatedAt      time.Time         `json:"createdAt,omitempty"`
	UpdatedAt      time.Time         `json:"updatedAt,omitempty"`
	Source         string            `json:"source,omitempty"`
	Owner          string            `json:"owner,omitempty"`
	Priority       Priority          `json:"priority,omitempty"`
	Responders     []Responder       `json:"responders,omitempty"`
	Integration    Integration       `json:"integration,omitempty"`
	Report         Report            `json:"report,omitempty"`
	Actions        []string          `json:"actions,omitempty"`
	Entity         string            `json:"entity,omitempty"`
	Description    string            `json:"description,omitempty"`
	Details        map[string]string `json:"details,omitempty"`
}

type CountAlertResult struct {
	client.ResultMetadata
	Count int `json:"count,omitempty"`
}

type AlertRecipient struct {
	User      User      `json:"user,omitempty"`
	CreatedAt time.Time `json:"createdAt,omitempty"`
	UpdatedAt time.Time `json:"updatedAt,omitempty"`
	State     string    `json:"state,omitempty"`
	Method    string    `json:"method,omitempty"`
}

type ListAlertRecipientResult struct {
	client.ResultMetadata
	AlertRecipients []AlertRecipient `json:"data"`
}

type AlertLog struct {
	Log       string    `json:"log,omitempty"`
	Type      string    `json:"type,omitempty"`
	Owner     string    `json:"owner,omitempty"`
	CreatedAt time.Time `json:"createdAt,omitempty"`
	Offset    string    `json:"offset,omitempty"`
}

type ListAlertLogsResult struct {
	client.ResultMetadata
	AlertLog []AlertLog        `json:"data"`
	Paging   map[string]string `json:"paging,omitempty"`
}

type AlertNote struct {
	Note      string    `json:"note,omitempty"`
	Owner     string    `json:"owner,omitempty"`
	CreatedAt time.Time `json:"createdAt,omitempty"`
	Offset    string    `json:"offset,omitempty"`
}

type ListAlertNotesResult struct {
	client.ResultMetadata
	AlertLog []AlertNote       `json:"data"`
	Paging   map[string]string `json:"paging,omitempty"`
}

type GetSavedSearchResult struct {
	client.ResultMetadata
	Id          string    `json:"id,omitempty"`
	Name        string    `json:"name,omitempty"`
	CreatedAt   time.Time `json:"createdAt,omitempty"`
	UpdatedAt   time.Time `json:"updatedAt,omitempty"`
	Teams       []Team    `json:"teams,omitempty"`
	Description string    `json:"description,omitempty"`
	Query       string    `json:"query,omitempty"`
}

type CreateAlertAttachmentsResult struct {
	client.ResultMetadata
	Result     string            `json:"result,omitempty"`
	Attachment CreatedAttachment `json:"data,omitempty"`
}

type CreatedAttachment struct {
	Id string `json:"id,omitempty"`
}

type GetAttachmentResult struct {
	client.ResultMetadata
	Name string `json:"name,omitempty"`
	Url  string `json:"url,omitempty"`
}

type ListAttachmentsResult struct {
	client.ResultMetadata
	Attachment []ListedAttachment `json:"data"`
}

type ListedAttachment struct {
	Id   int64  `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

type DeleteAlertAttachmentResult struct {
	client.ResultMetadata
	Result string `json:"result,omitempty"`
}

type AsyncAlertResult struct {
	client.ResultMetadata
	Result          string `json:"result,omitempty"`
	asyncBaseResult *client.AsyncBaseResult
}

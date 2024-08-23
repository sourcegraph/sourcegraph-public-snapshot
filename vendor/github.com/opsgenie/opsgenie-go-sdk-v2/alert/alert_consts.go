package alert

import "github.com/pkg/errors"

type SortField string

const (
	CreatedAt       SortField = "createdAt"
	UpdatedAt       SortField = "updatedAt"
	TinyId          SortField = "tinyId"
	Alias           SortField = "alias"
	Message         SortField = "message"
	Status          SortField = "status"
	Acknowledged    SortField = "acknowledged"
	IsSeen          SortField = "isSeen"
	Snoozed         SortField = "snoozed"
	SnoozedUntil    SortField = "snoozedUntil"
	Count           SortField = "count"
	LastOccurredAt  SortField = "lastOccuredAt"
	Source          SortField = "source"
	Owner           SortField = "owner"
	IntegrationName SortField = "integration.name"
	IntegrationType SortField = "integration.type"
	AckTime         SortField = "report.ackTime"
	CloseTime       SortField = "report.closeTime"
	AcknowledgedBy  SortField = "report.acknowledgedBy"
	ClosedBy        SortField = "report.closedBy"
)

type Report struct {
	AckTime        int64  `json:"ackTime,omitempty"`
	CloseTime      int64  `json:"closeTime,omitempty"`
	AcknowledgedBy string `json:"acknowledgedBy,omitempty"`
	ClosedBy       string `json:"closedBy,omitempty"`
}

type Order string

const (
	Asc  Order = "asc"
	Desc Order = "desc"
)

type Priority string

const (
	P1 Priority = "P1"
	P2 Priority = "P2"
	P3 Priority = "P3"
	P4 Priority = "P4"
	P5 Priority = "P5"
)

type SearchIdentifierType string

const (
	ID   SearchIdentifierType = "id"
	NAME SearchIdentifierType = "name"
)

type AlertIdentifier uint32

const (
	ALERTID AlertIdentifier = iota
	ALIAS
	TINYID
)

func validateIdentifier(identifier string) error {
	if identifier == "" {
		return errors.New("Identifier can not be empty")
	}
	return nil
}

type RequestDirection string

const (
	NEXT RequestDirection = "next"
	PREV RequestDirection = "prev"
)

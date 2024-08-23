package entities

import "github.com/uber/gonduit/util"

// ManiphestTask represents a single task on Maniphest.
type ManiphestTask struct {
	ID                 string             `json:"id"`
	PHID               string             `json:"phid"`
	AuthorPHID         string             `json:"authorPHID"`
	OwnerPHID          string             `json:"ownerPHID"`
	CCPHIDs            []string           `json:"ccPHIDs"`
	Status             string             `json:"status"`
	StatusName         string             `json:"statusName"`
	IsClosed           bool               `json:"isClosed"`
	Priority           string             `json:"priority"`
	PriorityColor      string             `json:"priorityColor"`
	Title              string             `json:"title"`
	Description        string             `json:"description"`
	ProjectPHIDs       []string           `json:"projectPHIDs"`
	URI                string             `json:"uri"`
	Auxiliary          interface{}        `json:"auxiliary"`
	ObjectName         string             `json:"objectName"`
	DateCreated        util.UnixTimestamp `json:"dateCreated"`
	DateModified       util.UnixTimestamp `json:"dateModified"`
	DependsOnTaskPHIDs []string           `json:"dependsOnTaskPHIDs"`
}

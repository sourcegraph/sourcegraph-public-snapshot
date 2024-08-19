package gitlab

import (
	"fmt"
	"net/http"
	"time"
)

// AuditEvent represents an audit event for a group, a project or the instance.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/audit_events.html
type AuditEvent struct {
	ID         int               `json:"id"`
	AuthorID   int               `json:"author_id"`
	EntityID   int               `json:"entity_id"`
	EntityType string            `json:"entity_type"`
	Details    AuditEventDetails `json:"details"`
	CreatedAt  *time.Time        `json:"created_at"`
}

// AuditEventDetails represents the details portion of an audit event for
// a group, a project or the instance. The exact fields that are returned
// for an audit event depend on the action being recorded.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/audit_events.html
type AuditEventDetails struct {
	With          string      `json:"with"`
	Add           string      `json:"add"`
	As            string      `json:"as"`
	Change        string      `json:"change"`
	From          string      `json:"from"`
	To            string      `json:"to"`
	Remove        string      `json:"remove"`
	CustomMessage string      `json:"custom_message"`
	AuthorName    string      `json:"author_name"`
	TargetID      interface{} `json:"target_id"`
	TargetType    string      `json:"target_type"`
	TargetDetails string      `json:"target_details"`
	IPAddress     string      `json:"ip_address"`
	EntityPath    string      `json:"entity_path"`
	FailedLogin   string      `json:"failed_login"`
}

// AuditEventsService handles communication with the project/group/instance
// audit event related methods of the GitLab API.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/audit_events.html
type AuditEventsService struct {
	client *Client
}

// ListAuditEventsOptions represents the available ListProjectAuditEvents(),
// ListGroupAuditEvents() or ListInstanceAuditEvents() options.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/audit_events.html
type ListAuditEventsOptions struct {
	ListOptions
	CreatedAfter  *time.Time `url:"created_after,omitempty" json:"created_after,omitempty"`
	CreatedBefore *time.Time `url:"created_before,omitempty" json:"created_before,omitempty"`
}

// ListInstanceAuditEvents gets a list of audit events for instance.
// Authentication as Administrator is required.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/audit_events.html#retrieve-all-instance-audit-events
func (s *AuditEventsService) ListInstanceAuditEvents(opt *ListAuditEventsOptions, options ...RequestOptionFunc) ([]*AuditEvent, *Response, error) {
	req, err := s.client.NewRequest(http.MethodGet, "audit_events", opt, options)
	if err != nil {
		return nil, nil, err
	}

	var aes []*AuditEvent
	resp, err := s.client.Do(req, &aes)
	if err != nil {
		return nil, resp, err
	}

	return aes, resp, nil
}

// GetInstanceAuditEvent gets a specific instance audit event.
// Authentication as Administrator is required.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/audit_events.html#retrieve-single-instance-audit-event
func (s *AuditEventsService) GetInstanceAuditEvent(event int, options ...RequestOptionFunc) (*AuditEvent, *Response, error) {
	u := fmt.Sprintf("audit_events/%d", event)

	req, err := s.client.NewRequest(http.MethodGet, u, nil, options)
	if err != nil {
		return nil, nil, err
	}

	ae := new(AuditEvent)
	resp, err := s.client.Do(req, ae)
	if err != nil {
		return nil, resp, err
	}

	return ae, resp, nil
}

// ListGroupAuditEvents gets a list of audit events for the specified group
// viewable by the authenticated user.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/audit_events.html#retrieve-all-group-audit-events
func (s *AuditEventsService) ListGroupAuditEvents(gid interface{}, opt *ListAuditEventsOptions, options ...RequestOptionFunc) ([]*AuditEvent, *Response, error) {
	group, err := parseID(gid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("groups/%s/audit_events", PathEscape(group))

	req, err := s.client.NewRequest(http.MethodGet, u, opt, options)
	if err != nil {
		return nil, nil, err
	}

	var aes []*AuditEvent
	resp, err := s.client.Do(req, &aes)
	if err != nil {
		return nil, resp, err
	}

	return aes, resp, nil
}

// GetGroupAuditEvent gets a specific group audit event.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/audit_events.html#retrieve-a-specific-group-audit-event
func (s *AuditEventsService) GetGroupAuditEvent(gid interface{}, event int, options ...RequestOptionFunc) (*AuditEvent, *Response, error) {
	group, err := parseID(gid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("groups/%s/audit_events/%d", PathEscape(group), event)

	req, err := s.client.NewRequest(http.MethodGet, u, nil, options)
	if err != nil {
		return nil, nil, err
	}

	ae := new(AuditEvent)
	resp, err := s.client.Do(req, ae)
	if err != nil {
		return nil, resp, err
	}

	return ae, resp, nil
}

// ListProjectAuditEvents gets a list of audit events for the specified project
// viewable by the authenticated user.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/audit_events.html#retrieve-all-project-audit-events
func (s *AuditEventsService) ListProjectAuditEvents(pid interface{}, opt *ListAuditEventsOptions, options ...RequestOptionFunc) ([]*AuditEvent, *Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("projects/%s/audit_events", PathEscape(project))

	req, err := s.client.NewRequest(http.MethodGet, u, opt, options)
	if err != nil {
		return nil, nil, err
	}

	var aes []*AuditEvent
	resp, err := s.client.Do(req, &aes)
	if err != nil {
		return nil, resp, err
	}

	return aes, resp, nil
}

// GetProjectAuditEvent gets a specific project audit event.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/audit_events.html#retrieve-a-specific-project-audit-event
func (s *AuditEventsService) GetProjectAuditEvent(pid interface{}, event int, options ...RequestOptionFunc) (*AuditEvent, *Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("projects/%s/audit_events/%d", PathEscape(project), event)

	req, err := s.client.NewRequest(http.MethodGet, u, nil, options)
	if err != nil {
		return nil, nil, err
	}

	ae := new(AuditEvent)
	resp, err := s.client.Do(req, ae)
	if err != nil {
		return nil, resp, err
	}

	return ae, resp, nil
}

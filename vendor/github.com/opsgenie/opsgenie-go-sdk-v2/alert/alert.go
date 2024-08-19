package alert

import (
	"context"
	"github.com/opsgenie/opsgenie-go-sdk-v2/client"
)

type Client struct {
	client *client.OpsGenieClient
}

func NewClient(config *client.Config) (*Client, error) {

	opsgenieClient, err := client.NewOpsGenieClient(config)

	if err != nil {
		return nil, err
	}

	return &Client{client: opsgenieClient}, nil
}

func (c *Client) Create(ctx context.Context, req *CreateAlertRequest) (*AsyncAlertResult, error) {

	result := &AsyncAlertResult{}

	err := c.client.Exec(ctx, req, result)
	if err != nil {
		return nil, err
	}

	result.asyncBaseResult = &client.AsyncBaseResult{Client: c.client}

	return result, nil

}

func (c *Client) Delete(ctx context.Context, req *DeleteAlertRequest) (*AsyncAlertResult, error) {

	result := &AsyncAlertResult{}

	err := c.client.Exec(ctx, req, result)
	if err != nil {
		return nil, err
	}

	result.asyncBaseResult = &client.AsyncBaseResult{Client: c.client}

	return result, nil

}

func (c *Client) Get(ctx context.Context, req *GetAlertRequest) (*GetAlertResult, error) {

	result := &GetAlertResult{}

	err := c.client.Exec(ctx, req, result)
	if err != nil {
		return nil, err
	}

	return result, nil

}

func (c *Client) List(ctx context.Context, req *ListAlertRequest) (*ListAlertResult, error) {

	result := &ListAlertResult{}

	err := c.client.Exec(ctx, req, result)
	if err != nil {
		return nil, err
	}

	return result, nil

}

func (c *Client) CountAlerts(ctx context.Context, req *CountAlertsRequest) (*CountAlertResult, error) {

	result := &CountAlertResult{}

	err := c.client.Exec(ctx, req, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (c *Client) Acknowledge(ctx context.Context, req *AcknowledgeAlertRequest) (*AsyncAlertResult, error) {

	result := &AsyncAlertResult{}

	err := c.client.Exec(ctx, req, result)
	if err != nil {
		return nil, err
	}

	result.asyncBaseResult = &client.AsyncBaseResult{Client: c.client}

	return result, nil
}

func (c *Client) Close(ctx context.Context, req *CloseAlertRequest) (*AsyncAlertResult, error) {

	result := &AsyncAlertResult{}

	err := c.client.Exec(ctx, req, result)
	if err != nil {
		return nil, err
	}

	result.asyncBaseResult = &client.AsyncBaseResult{Client: c.client}

	return result, nil
}

func (c *Client) AddNote(ctx context.Context, req *AddNoteRequest) (*AsyncAlertResult, error) {

	result := &AsyncAlertResult{}

	err := c.client.Exec(ctx, req, result)
	if err != nil {
		return nil, err
	}

	result.asyncBaseResult = &client.AsyncBaseResult{Client: c.client}

	return result, nil
}

func (c *Client) ExecuteCustomAction(ctx context.Context, req *ExecuteCustomActionAlertRequest) (*AsyncAlertResult, error) {

	result := &AsyncAlertResult{}

	err := c.client.Exec(ctx, req, result)
	if err != nil {
		return nil, err
	}

	result.asyncBaseResult = &client.AsyncBaseResult{Client: c.client}

	return result, nil
}

func (c *Client) Unacknowledge(ctx context.Context, req *UnacknowledgeAlertRequest) (*AsyncAlertResult, error) {

	result := &AsyncAlertResult{}

	err := c.client.Exec(ctx, req, result)
	if err != nil {
		return nil, err
	}

	result.asyncBaseResult = &client.AsyncBaseResult{Client: c.client}

	return result, nil
}

func (c *Client) Snooze(ctx context.Context, req *SnoozeAlertRequest) (*AsyncAlertResult, error) {

	result := &AsyncAlertResult{}

	err := c.client.Exec(ctx, req, result)
	if err != nil {
		return nil, err
	}

	result.asyncBaseResult = &client.AsyncBaseResult{Client: c.client}

	return result, nil
}

func (c *Client) EscalateToNext(ctx context.Context, req *EscalateToNextRequest) (*AsyncAlertResult, error) {

	result := &AsyncAlertResult{}

	err := c.client.Exec(ctx, req, result)
	if err != nil {
		return nil, err
	}

	result.asyncBaseResult = &client.AsyncBaseResult{Client: c.client}

	return result, nil
}

func (c *Client) AssignAlert(ctx context.Context, req *AssignRequest) (*AsyncAlertResult, error) {

	result := &AsyncAlertResult{}

	err := c.client.Exec(ctx, req, result)
	if err != nil {
		return nil, err
	}

	result.asyncBaseResult = &client.AsyncBaseResult{Client: c.client}

	return result, nil
}

func (c *Client) AddTeam(ctx context.Context, req *AddTeamRequest) (*AsyncAlertResult, error) {

	result := &AsyncAlertResult{}

	err := c.client.Exec(ctx, req, result)
	if err != nil {
		return nil, err
	}

	result.asyncBaseResult = &client.AsyncBaseResult{Client: c.client}

	return result, nil
}

func (c *Client) AddResponder(ctx context.Context, req *AddResponderRequest) (*AsyncAlertResult, error) {

	result := &AsyncAlertResult{}

	err := c.client.Exec(ctx, req, result)
	if err != nil {
		return nil, err
	}

	result.asyncBaseResult = &client.AsyncBaseResult{Client: c.client}

	return result, nil
}

func (c *Client) AddTags(ctx context.Context, req *AddTagsRequest) (*AsyncAlertResult, error) {

	result := &AsyncAlertResult{}

	err := c.client.Exec(ctx, req, result)
	if err != nil {
		return nil, err
	}

	result.asyncBaseResult = &client.AsyncBaseResult{Client: c.client}

	return result, nil
}

func (c *Client) RemoveTags(ctx context.Context, req *RemoveTagsRequest) (*AsyncAlertResult, error) {

	result := &AsyncAlertResult{}

	err := c.client.Exec(ctx, req, result)
	if err != nil {
		return nil, err
	}

	result.asyncBaseResult = &client.AsyncBaseResult{Client: c.client}

	return result, nil
}

func (c *Client) AddDetails(ctx context.Context, req *AddDetailsRequest) (*AsyncAlertResult, error) {

	result := &AsyncAlertResult{}

	err := c.client.Exec(ctx, req, result)
	if err != nil {
		return nil, err
	}

	result.asyncBaseResult = &client.AsyncBaseResult{Client: c.client}

	return result, nil
}

func (c *Client) RemoveDetails(ctx context.Context, req *RemoveDetailsRequest) (*AsyncAlertResult, error) {

	result := &AsyncAlertResult{}

	err := c.client.Exec(ctx, req, result)
	if err != nil {
		return nil, err
	}

	result.asyncBaseResult = &client.AsyncBaseResult{Client: c.client}

	return result, nil
}

func (c *Client) UpdatePriority(ctx context.Context, req *UpdatePriorityRequest) (*AsyncAlertResult, error) {

	result := &AsyncAlertResult{}

	err := c.client.Exec(ctx, req, result)
	if err != nil {
		return nil, err
	}

	result.asyncBaseResult = &client.AsyncBaseResult{Client: c.client}

	return result, nil
}

func (c *Client) UpdateMessage(ctx context.Context, req *UpdateMessageRequest) (*AsyncAlertResult, error) {

	result := &AsyncAlertResult{}

	err := c.client.Exec(ctx, req, result)
	if err != nil {
		return nil, err
	}

	result.asyncBaseResult = &client.AsyncBaseResult{Client: c.client}

	return result, nil
}

func (c *Client) UpdateDescription(ctx context.Context, req *UpdateDescriptionRequest) (*AsyncAlertResult, error) {

	result := &AsyncAlertResult{}

	err := c.client.Exec(ctx, req, result)
	if err != nil {
		return nil, err
	}

	result.asyncBaseResult = &client.AsyncBaseResult{Client: c.client}

	return result, nil
}

func (c *Client) ListAlertRecipients(ctx context.Context, req *ListAlertRecipientRequest) (*ListAlertRecipientResult, error) {

	result := &ListAlertRecipientResult{}

	err := c.client.Exec(ctx, req, result)
	if err != nil {
		return nil, err
	}

	return result, nil

}

func (c *Client) ListAlertLogs(ctx context.Context, req *ListAlertLogsRequest) (*ListAlertLogsResult, error) {

	result := &ListAlertLogsResult{}

	err := c.client.Exec(ctx, req, result)
	if err != nil {
		return nil, err
	}

	return result, nil

}

func (c *Client) ListAlertNotes(ctx context.Context, req *ListAlertNotesRequest) (*ListAlertNotesResult, error) {

	result := &ListAlertNotesResult{}

	err := c.client.Exec(ctx, req, result)
	if err != nil {
		return nil, err
	}

	return result, nil

}

func (c *Client) CreateSavedSearch(ctx context.Context, req *CreateSavedSearchRequest) (*SavedSearchResult, error) {

	result := &SavedSearchResult{}

	err := c.client.Exec(ctx, req, result)
	if err != nil {
		return nil, err
	}

	return result, nil

}

func (c *Client) UpdateSavedSearch(ctx context.Context, req *UpdateSavedSearchRequest) (*SavedSearchResult, error) {

	result := &SavedSearchResult{}

	err := c.client.Exec(ctx, req, result)
	if err != nil {
		return nil, err
	}

	return result, nil

}

func (c *Client) GetSavedSearch(ctx context.Context, req *GetSavedSearchRequest) (*GetSavedSearchResult, error) {

	result := &GetSavedSearchResult{}

	err := c.client.Exec(ctx, req, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (c *Client) DeleteSavedSearch(ctx context.Context, req *DeleteSavedSearchRequest) (*AsyncAlertResult, error) {

	result := &AsyncAlertResult{}

	err := c.client.Exec(ctx, req, result)
	if err != nil {
		return nil, err
	}

	result.asyncBaseResult = &client.AsyncBaseResult{Client: c.client}

	return result, nil
}

func (c *Client) ListSavedSearches(ctx context.Context, req *ListSavedSearchRequest) (*SavedSearchResult, error) {

	result := &SavedSearchResult{}

	err := c.client.Exec(ctx, req, result)
	if err != nil {
		return nil, err
	}

	return result, nil

}

func (c *Client) GetRequestStatus(ctx context.Context, req *GetRequestStatusRequest) (*RequestStatusResult, error) {

	result := &RequestStatusResult{}

	err := c.client.Exec(ctx, req, result)
	if err != nil {
		return nil, err
	}

	return result, nil

}

func (c *Client) CreateAlertAttachments(ctx context.Context, req *CreateAlertAttachmentRequest) (*CreateAlertAttachmentsResult, error) {

	result := &CreateAlertAttachmentsResult{}

	err := c.client.Exec(ctx, req, result)
	if err != nil {
		return nil, err
	}

	return result, nil

}

func (c *Client) GetAlertAttachment(ctx context.Context, req *GetAttachmentRequest) (*GetAttachmentResult, error) {

	result := &GetAttachmentResult{}

	err := c.client.Exec(ctx, req, result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (c *Client) ListAlertsAttachments(ctx context.Context, req *ListAttachmentsRequest) (*ListAttachmentsResult, error) {

	result := &ListAttachmentsResult{}

	err := c.client.Exec(ctx, req, result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (c *Client) DeleteAlertAttachment(ctx context.Context, req *DeleteAttachmentRequest) (*DeleteAlertAttachmentResult, error) {

	result := &DeleteAlertAttachmentResult{}

	err := c.client.Exec(ctx, req, result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (ar *AsyncAlertResult) RetrieveStatus(ctx context.Context) (*RequestStatusResult, error) {

	req := &GetRequestStatusRequest{RequestId: ar.RequestId}
	result := &RequestStatusResult{}

	err := ar.asyncBaseResult.RetrieveStatus(ctx, req, result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

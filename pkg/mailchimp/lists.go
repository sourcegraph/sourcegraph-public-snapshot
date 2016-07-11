package mailchimp

import (
	"path"
	"strconv"
)

type PostListsMembersOptions struct {
	EmailType       string                 `json:"email_type,omitempty"`
	Status          string                 `json:"status"`
	MergeFields     map[string]interface{} `json:"merge_fields,omitempty"`
	Interests       map[string]interface{} `json:"interests,omitempty"` // TODO: better type safety
	Language        string                 `json:"language,omitempty"`
	VIP             bool                   `json:"location,omitempty"`
	Location        map[string]interface{} `json:"location,omitempty"` // TODO: better type safety
	IPSignup        string                 `json:"ip_signup,omitempty"`
	TimestampSignup string                 `json:"timestamp_signup,omitempty"`
	IPOpt           string                 `json:"ip_opt,omitempty"`
	TimestampOpt    string                 `json:"timestamp_opt,omitempty"`
	EmailAddress    string                 `json:"email_address"`
}

type PostListsMembersResponse struct {
	ID              string                 `json:"id,omitempty"`
	EmailAddress    string                 `json:"email_address,omitempty"`
	UniqueEmailID   string                 `json:"unique_email_id,omitempty"`
	EmailType       string                 `json:"email_type,omitempty"`
	Status          string                 `json:"status,omitempty"`
	StatusIfNew     string                 `json:"status_if_new,omitempty"`
	MergeFields     map[string]interface{} `json:"merge_fields,omitempty"`
	Interests       map[string]interface{} `json:"interests,omitempty"` // TODO: better type safety
	Stats           map[string]interface{} `json:"stats,omitempty"`     // TODO: better type safety
	IPSignup        string                 `json:"ip_signup,omitempty"`
	TimestampSignup string                 `json:"timestamp_signup,omitempty"`
	IPOpt           string                 `json:"ip_opt,omitempty"`
	TimestampOpt    string                 `json:"timestamp_opt,omitempty"`
	MemberRating    int                    `json:"member_rating,omitempty"`
	LastChanged     string                 `json:"last_changed,omitempty"`
	Language        string                 `json:"language,omitempty"`
	VIP             bool                   `json:"location,omitempty"`
	EmailClient     string                 `json:"email_client,omitempty"`
	Location        map[string]interface{} `json:"location,omitempty"` // TODO: better type safety
	ListID          string                 `json:"list_id,omitempty"`
	Links           []interface{}          `json:"_links,omitempty"` // TODO: better type safety
}

// PostListsMembers adds a new list member.
//
// http://developer.mailchimp.com/documentation/mailchimp/reference/lists/members/
func (c *Client) PostListsMembers(listID string, opts *PostListsMembersOptions) (*PostListsMembersResponse, error) {
	resp := &PostListsMembersResponse{}
	err := c.post("PostListsMembers", path.Join("lists", listID, "members"), opts, resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

type PutListsMembersOptions struct {
	EmailType       string                 `json:"email_type,omitempty"`
	Status          string                 `json:"status,omitempty"`
	MergeFields     map[string]interface{} `json:"merge_fields,omitempty"`
	Interests       map[string]interface{} `json:"interests,omitempty"` // TODO: better type safety
	Language        string                 `json:"language,omitempty"`
	VIP             bool                   `json:"location,omitempty"`
	Location        map[string]interface{} `json:"location,omitempty"` // TODO: better type safety
	IPSignup        string                 `json:"ip_signup,omitempty"`
	TimestampSignup string                 `json:"timestamp_signup,omitempty"`
	IPOpt           string                 `json:"ip_opt,omitempty"`
	TimestampOpt    string                 `json:"timestamp_opt,omitempty"`
	EmailAddress    string                 `json:"email_address"`
	StatusIfNew     string                 `json:"status_if_new"`
}

type PutListsMembersResponse struct {
	ID              string                 `json:"id,omitempty"`
	EmailAddress    string                 `json:"email_address,omitempty"`
	UniqueEmailID   string                 `json:"unique_email_id,omitempty"`
	EmailType       string                 `json:"email_type,omitempty"`
	Status          string                 `json:"status,omitempty"`
	MergeFields     map[string]interface{} `json:"merge_fields,omitempty"`
	Interests       map[string]interface{} `json:"interests,omitempty"` // TODO: better type safety
	Stats           map[string]interface{} `json:"stats,omitempty"`     // TODO: better type safety
	IPSignup        string                 `json:"ip_signup,omitempty"`
	TimestampSignup string                 `json:"timestamp_signup,omitempty"`
	IPOpt           string                 `json:"ip_opt,omitempty"`
	TimestampOpt    string                 `json:"timestamp_opt,omitempty"`
	MemberRating    int                    `json:"member_rating,omitempty"`
	LastChanged     string                 `json:"last_changed,omitempty"`
	Language        string                 `json:"language,omitempty"`
	VIP             bool                   `json:"location,omitempty"`
	EmailClient     string                 `json:"email_client,omitempty"`
	Location        map[string]interface{} `json:"location,omitempty"`  // TODO: better type safety
	LastNote        map[string]interface{} `json:"last_note,omitempty"` // TODO: better type safety
	ListID          string                 `json:"list_id,omitempty"`
	Links           []interface{}          `json:"_links,omitempty"` // TODO: better type safety
}

// PutListsMembers adds or updates a list member.
//
// http://developer.mailchimp.com/documentation/mailchimp/reference/lists/members/
func (c *Client) PutListsMembers(listID, subscriberHash string, opts *PutListsMembersOptions) (*PutListsMembersResponse, error) {
	resp := &PutListsMembersResponse{}
	err := c.put("PutListsMembers", path.Join("lists", listID, "members", subscriberHash), opts, resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

type ListsMergeFieldsOptions struct {
	Fields        []string `url:"fields,comma,omitempty"`
	ExcludeFields []string `url:"exclude_fields,comma,omitempty"`
	Count         int      `url:"count,omitempty"`
	Offset        int      `url:"offset,omitempty"`
	Type          string   `url:"type,omitempty"`
	Required      bool     `url:"required,omitempty"`
}

type ListsMergeFieldsResponse struct {
	MergeFields []map[string]interface{} `json:"merge_fields,omitempty"` // TODO: better type safety
	ListID      string                   `json:"list_id,omitempty"`
	TotalItems  int                      `json:"total_items,omitempty"`
	Links       []interface{}            `json:"_links,omitempty"` // TODO: better type safety
}

// ListsMergeFields gets all merge fields for a list.
//
// http://developer.mailchimp.com/documentation/mailchimp/reference/lists/merge-fields
func (c *Client) ListsMergeFields(listID string, opts *ListsMergeFieldsOptions) (*ListsMergeFieldsResponse, error) {
	resp := &ListsMergeFieldsResponse{}
	err := c.get("ListsMergeFields", path.Join("lists", listID, "merge-fields"), opts, resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

type PatchListsMergeFieldsOptions struct {
	Tag          string                 `json:"tag,omitempty"`
	Name         string                 `json:"name"`
	Type         string                 `json:"type"`
	Required     bool                   `json:"required,omitempty"`
	DefaultValue string                 `json:",default_value,omitempty"`
	Public       bool                   `json:"public,omitempty"`
	DisplayOrder int                    `json:"display_order,omitempty"`
	Options      map[string]interface{} `json:"options,omitempty"` // TODO: better type safety
	HelpText     string                 `json:"help_text,omitempty"`
}

type PatchListsMergeFieldsResponse struct {
	MergeID      int                    `json:"merge_id,omitempty"`
	Tag          string                 `json:"tag,omitempty"`
	Name         string                 `json:"name"`
	Type         string                 `json:"type"`
	Required     bool                   `json:"required,omitempty"`
	DefaultValue string                 `json:"default_value,omitempty"`
	Public       bool                   `json:"public,omitempty"`
	DisplayOrder int                    `json:"display_order,omitempty"`
	Options      map[string]interface{} `json:"options,omitempty"` // TODO: better type safety
	HelpText     string                 `json:"help_text,omitempty"`
	ListID       string                 `json:"list_id,omitempty"`
	Links        []interface{}          `json:"_links,omitempty"` // TODO: better type safety
}

func (c *Client) PatchListsMergeFields(listID string, mergeID int, opts *PatchListsMergeFieldsOptions) (*PatchListsMergeFieldsResponse, error) {
	resp := &PatchListsMergeFieldsResponse{}
	err := c.patch("PatchListsMergeFields", path.Join("lists", listID, "merge-fields", strconv.FormatInt(int64(mergeID), 10)), opts, resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

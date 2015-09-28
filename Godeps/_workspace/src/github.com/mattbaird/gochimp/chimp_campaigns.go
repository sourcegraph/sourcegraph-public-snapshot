// Copyright 2013 Matthew Baird
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//     http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package gochimp

import (
	"fmt"
	"strings"
)

const (
	get_content_endpoint     string = "/campaigns/content.%s"
	campaign_create_endpoint string = "/campaigns/create.json"
	campaign_send_endpoint   string = "/campaigns/send.json"
	campaign_list_endpoint   string = "/campaigns/list.json"
)

func (a *ChimpAPI) GetContentAsXML(cid string, options map[string]interface{}) (ContentResponse, error) {
	return a.GetContent(cid, options, "xml")
}

func (a *ChimpAPI) GetContentAsJson(cid string, options map[string]interface{}) (ContentResponse, error) {
	return a.GetContent(cid, options, "json")
}

func (a *ChimpAPI) GetContent(cid string, options map[string]interface{}, contentFormat string) (ContentResponse, error) {
	var response ContentResponse
	if !strings.EqualFold(strings.ToLower(contentFormat), "xml") && strings.EqualFold(strings.ToLower(contentFormat), "json") {
		return response, fmt.Errorf("contentFormat should be one of xml or json, you passed an unsupported value %s", contentFormat)
	}
	var params map[string]interface{} = make(map[string]interface{})
	params["apikey"] = a.Key
	params["cid"] = cid
	params["options"] = options
	err := parseChimpJson(a, fmt.Sprintf(get_content_endpoint, contentFormat), params, &response)
	return response, err
}

func (a *ChimpAPI) CampaignCreate(req CampaignCreate) (CampaignResponse, error) {
	req.ApiKey = a.Key
	var response CampaignResponse
	err := parseChimpJson(a, campaign_create_endpoint, req, &response)
	return response, err
}

func (a *ChimpAPI) CampaignSend(cid string) (CampaignSendResponse, error) {
	req := campaignSend{
		ApiKey:     a.Key,
		CampaignId: cid,
	}
	var response CampaignSendResponse
	err := parseChimpJson(a, campaign_send_endpoint, req, &response)
	return response, err
}

func (a *ChimpAPI) CampaignList(req CampaignList) (CampaignListResponse, error) {
	req.ApiKey = a.Key
	var response CampaignListResponse
	err := parseChimpJson(a, campaign_list_endpoint, req, &response)
	return response, err
}

type CampaignListResponse struct {
	Total     int                `json:"total"`
	Campaigns []CampaignResponse `json:"data"`
}

type CampaignList struct {
	// A valid API Key for your user account. Get by visiting your API dashboard
	ApiKey string `json:"apikey"`

	// Filters to apply to this query - all are optional:
	Filter CampaignListFilter `json:"filters,omitempty"`

	// Control paging of campaigns, start results at this campaign #,
	// defaults to 1st page of data (page 0)
	Start int `json:"start,omitempty"`

	// Control paging of campaigns, number of campaigns to return with each call, defaults to 25 (max=1000)
	Limit int `json:"limit,omitempty"`

	// One of "create_time", "send_time", "title", "subject". Invalid values
	// will fall back on "create_time" - case insensitive.
	SortField string `json:"sort_field,omitempty"`

	// "DESC" for descending (default), "ASC" for Ascending. Invalid values
	// will fall back on "DESC" - case insensitive.
	OrderOrder string `json:"sort_dir,omitempty"`
}

type CampaignListFilter struct {
	// Return the campaign using a know campaign_id. Accepts
	// multiples separated by commas when not using exact matching.
	CampaignID string `json:"campaign_id,omitempty"`

	// Return the child campaigns using a known parent campaign_id.
	// Accepts multiples separated by commas when not using exact matching.
	ParentID string `json:"parent_id,omitempty"`

	// The list to send this campaign to - Get lists using ListList.
	// Accepts multiples separated by commas when not using exact matching.
	ListID string `json:"list_id,omitempty"`

	// Only show campaigns from this folder id - get folders using FoldersList.
	// Accepts multiples separated by commas when not using exact matching.
	FolderID int `json:"folder_id,omitempty"`

	// Only show campaigns using this template id - get templates using TemplatesList.
	// Accepts multiples separated by commas when not using exact matching.
	TemplateID int `json:"template_id,omitempty"`

	// Return campaigns of a specific status - one of "sent", "save", "paused", "schedule", "sending".
	// Accepts multiples separated by commas when not using exact matching.
	Status string `json:"status,omitempty"`

	// Return campaigns of a specific type - one of "regular", "plaintext", "absplit", "rss", "auto".
	// Accepts multiples separated by commas when not using exact matching.
	Type string `json:"type,omitempty"`

	// Only show campaigns that have this "From Name"
	FromName string `json:"from_name,omitempty"`

	// Only show campaigns that have this "Reply-to Email"
	FromEmail string `json:"from_email,omitempty"`

	// Only show campaigns that have this title
	Title string `json:"title"`

	// Only show campaigns that have this subject
	Subject string `json:"subject"`

	// Only show campaigns that have been sent since this date/time (in GMT) - -
	// 24 hour format in GMT, eg "2013-12-30 20:30:00" - if this is invalid the whole call fails
	SendTimeStart string `json:"sendtime_start,omitempty"`

	// Only show campaigns that have been sent before this date/time (in GMT) - -
	// 24 hour format in GMT, eg "2013-12-30 20:30:00" - if this is invalid the whole call fails
	SendTimeEnd string `json:"sendtime_end,omitempty"`

	// Whether to return just campaigns with or without segments
	UsesSegment bool `json:"uses_segment,omitempty"`

	// Flag for whether to filter on exact values when filtering, or search within content for
	// filter values - defaults to true. Using this disables the use of any filters that accept multiples.
	Exact bool `json:"exact,omitempty"`
}

type campaignSend struct {
	ApiKey     string `json:"apikey"`
	CampaignId string `json:"cid"`
}

type CampaignSendResponse struct {
	Complete bool `json:"complete"`
}

type CampaignCreate struct {
	ApiKey  string                `json:"apikey"`
	Type    string                `json:"type"`
	Options CampaignCreateOptions `json:"options"`
	Content CampaignCreateContent `json:"content"`
}

type CampaignCreateOptions struct {
	// ListID is the list to send this campaign to
	ListID string `json:"list_id"`

	// TemplateID is the user-created template from which the HTML
	// content of the campaign should be created
	TemplateID string `json:"template_id"`

	// Subject is the subject line for your campaign message
	Subject string `json:"subject"`

	// FromEmail is the From: email address for your campaign message
	FromEmail string `json:"from_email"`

	// FromName is the From: name for your campaign message (not an email address)
	FromName string `json:"from_name"`

	// ToName is the To: name recipients will see (not email address)
	ToName string `json:"to_name"`
}

type CampaignCreateContent struct {
	// HTML is the raw/pasted HTML content for the campaign
	HTML string `json:"html"`

	// When using a template instead of raw HTML, each key
	// in the map should be the unique mc:edit area name from
	// the template.
	Sections map[string]string `json:"sections,omitempty"`

	// Text is the plain-text version of the body
	Text string `json:"text"`

	// MailChimp will pull in content from this URL. Note,
	// this will override any other content options - for lists
	// with Email Format options, you'll need to turn on
	// generate_text as well
	URL string `json:"url,omitempty"`

	// A Base64 encoded archive file for MailChimp to import all
	// media from. Note, this will override any other content
	// options - for lists with Email Format options, you'll
	// need to turn on generate_text as well
	Archive string `json:"archive,omitempty"`

	// ArchiveType only applies to the Archive field. Supported
	// formats are: zip, tar.gz, tar.bz2, tar, tgz, tbz.
	// If not included, we will default to zip
	ArchiveType string `json:"archive_options,omitempty"`
}

type CampaignResponse struct {
	Id                 string           `json:"id"`
	WebId              int              `json:"web_id"`
	ListId             string           `json:"list_id"`
	FolderId           int              `json:"folder_id"`
	TemplateId         int              `json:"template_id"`
	ContentType        string           `json:"content_type"`
	ContentEditedBy    string           `json:"content_edited_by"`
	Title              string           `json:"title"`
	Type               string           `json:"type"`
	CreateTime         string           `json:"create_time"`
	SendTime           string           `json:"send_time"`
	ContentUpdatedTime string           `json:"content_updated_time"`
	Status             string           `json:"status"`
	FromName           string           `json:"from_name"`
	FromEmail          string           `json:"from_email"`
	Subject            string           `json:"subject"`
	ToName             string           `json:"to_name"`
	ArchiveURL         string           `json:"archive_url"`
	ArchiveURLLong     string           `json:"archive_url_long"`
	EmailsSent         int              `json:"emails_sent"`
	Analytics          string           `json:"analytics"`
	AnalyticsTag       string           `json:"analytics_tag"`
	InlineCSS          bool             `json:"inline_css"`
	Authenticate       bool             `json:"authenticate"`
	Ecommm360          bool             `json:"ecomm360"`
	AutoTweet          bool             `json:"auto_tweet"`
	AutoFacebookPort   string           `json:"auto_fb_post"`
	AutoFooter         bool             `json:"auto_footer"`
	Timewarp           bool             `json:"timewarp"`
	TimewarpSchedule   string           `json:"timewarp_schedule,omitempty"`
	Tracking           CampaignTracking `json:"tracking"`
	ParentId           string           `json:"parent_id"`
	IsChild            bool             `json:"is_child"`
	TestsRemaining     int              `json:"tests_remain"`
	SegmentText        string           `json:"segment_text"`
}

type CampaignTracking struct {
	HTMLClicks bool `json:"html_clicks"`
	TextClicks bool `json:"text_clicks"`
	Opens      bool `json:"opens"`
}

type ContentResponse struct {
	Html string `json:"html"`
	Text string `json:"text"`
}

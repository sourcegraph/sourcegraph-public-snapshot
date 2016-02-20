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
	"encoding/json"
	"errors"
	"time"
)

// see https://mandrillapp.com/api/docs/messages.html
const messages_send_endpoint string = "/messages/send.json"                   // Send a new transactional message through Mandrill
const messages_send_template_endpoint string = "/messages/send-template.json" // Send a new transactional message through Mandrill using a template
const messages_search_endpoint string = "/messages/search.json"               // Search the content of recently sent messages and optionally narrow by date range, tags and senders
const messages_parse_endpoint string = "/messages/parse.json"                 // Parse the full MIME document for an email message, returning the content of the message broken into its constituent pieces
const messages_send_raw_endpoint string = "/messages/send-raw.json"           // Take a raw MIME document for a message, and send it exactly as if it were sent over the SMTP protocol

//todo: add send_at, ip_pool and key to messagesend
func (a *MandrillAPI) MessageSend(message Message, async bool) ([]SendResponse, error) {
	var response []SendResponse
	var params map[string]interface{} = make(map[string]interface{})
	params["message"] = message
	params["async"] = async
	err := parseMandrillJson(a, messages_send_endpoint, params, &response)
	return response, err
}

func (a *MandrillAPI) MessageSendTemplate(templateName string, templateContent []Var, message Message, async bool) ([]SendResponse, error) {
	var response []SendResponse
	if templateName == "" {
		return response, errors.New("templateName cannot be blank")
	}
	var params map[string]interface{} = make(map[string]interface{})
	params["message"] = message
	params["template_name"] = templateName
	params["async"] = async
	params["template_content"] = templateContent
	err := parseMandrillJson(a, messages_send_template_endpoint, params, &response)
	return response, err
}

func (a *MandrillAPI) MessageSearch(searchRequest SearchRequest) ([]SearchResponse, error) {
	var response []SearchResponse
	var params map[string]interface{} = make(map[string]interface{})
	//todo remove this hack
	params["query"] = searchRequest.Query
	params["date_from"] = searchRequest.DateFrom
	params["date_to"] = searchRequest.DateTo
	params["tags"] = searchRequest.Tags
	params["senders"] = searchRequest.Senders
	params["limit"] = searchRequest.Limit
	err := parseMandrillJson(a, messages_search_endpoint, params, &response)
	return response, err
}

func (a *MandrillAPI) MessageParse(rawMessage string, async bool) (Message, error) {
	var response Message
	if rawMessage == "" {
		return response, errors.New("rawMessage cannot be blank")
	}
	var params map[string]interface{} = make(map[string]interface{})
	params["raw_message"] = rawMessage
	err := parseMandrillJson(a, messages_parse_endpoint, params, &response)
	return response, err
}

// Can return oneof Invalid_Key, ValidationError or GeneralError
func (a *MandrillAPI) MessageSendRaw(rawMessage string, to []string, from Recipient, async bool) ([]SendResponse, error) {
	var response []SendResponse
	if rawMessage == "" {
		return response, errors.New("rawMessage cannot be blank")
	}
	if len(to) <= 0 {
		return response, errors.New("You need at least one recipient in the To array")
	}

	var params map[string]interface{} = make(map[string]interface{})
	params["raw_message"] = rawMessage
	params["from_email"] = from.Email
	params["from_name"] = from.Name
	params["to"] = to
	params["async"] = async
	err := parseMandrillJson(a, messages_send_raw_endpoint, params, &response)
	return response, err
}

type SearchResponse struct {
	Timestamp time.Duration       `json:"ts"`
	Id        string              `json:"_id"`
	Sender    string              `json:"sender"`
	Subject   string              `json:"subject"`
	Email     string              `json:"email"`
	Tags      []string            `json:"tags"`
	Opens     int                 `json:"opens"`
	Clicks    int                 `json:"clicks"`
	State     string              `json:"state"`
	Metadata  []map[string]string `json:"metadata"`
}

type SearchRequest struct {
	Query    string   `json:"query"`
	DateFrom APITime  `json:"date_from"`
	DateTo   APITime  `json:"date_to"`
	Tags     []string `json:"tags"`
	Senders  []string `json:"senders"`
	Limit    int      `json:"limit"`
}

type Message struct {
	Html                    string              `json:"html,omitempty"`
	Text                    string              `json:"text,omitempty"`
	Subject                 string              `json:"subject"`
	FromEmail               string              `json:"from_email"`
	FromName                string              `json:"from_name"`
	To                      []Recipient         `json:"to"`
	Headers                 map[string]string   `json:"headers,omitempty"`
	Important               bool                `json:"important,omitempty"`
	TrackOpens              bool                `json:"track_opens"`
	TrackClicks             bool                `json:"track_clicks"`
	ViewContentLink         bool                `json:"view_content_link,omitempty"`
	AutoText                bool                `json:"auto_text,omitempty"`
	AutoHtml                bool                `json:"auto_html,omitempty"`
	UrlStripQS              bool                `json:"url_strip_qs,omitempty"`
	InlineCss               bool                `json:"inline_css,omitempty"`
	PreserveRecipients      bool                `json:"preserve_recipients,omitempty"`
	BCCAddress              string              `json:"bcc_address,omitempty"`
	TrackingDomain          string              `json:"tracking_domain,omitempty"`
	SigningDomain           string              `json:"signing_domain,omitempty"`
	ReturnPathDomain        string              `json:"return_path_domain,omitempty"`
	Merge                   bool                `json:"merge,omitempty"`
	GlobalMergeVars         []Var               `json:"global_merge_vars,omitempty"`
	MergeVars               []MergeVars         `json:"merge_vars,omitempty"`
	Tags                    []string            `json:"tags,omitempty"`
	Subaccount              string              `json:"subaccount,omitempty"`
	GoogleAnalyticsDomains  []string            `json:"google_analytics_domains,omitempty"`
	GoogleAnalyticsCampaign []string            `json:"google_analytics_campaign,omitempty"`
	Metadata                map[string]string   `json:"metadata,omitempty"`
	RecipientMetadata       []RecipientMetaData `json:"recipient_metadata,omitempty"`
	Attachments             []Attachment        `json:"attachments,omitempty"`
	MergeLanguage           string              `json:"merge_language,omitempty"`
}

func (m *Message) String() string {
	b, _ := json.Marshal(m)
	return string(b)
}

func (m *Message) AddHeader(key, value string) {
	if m.Headers == nil {
		m.Headers = make(map[string]string)
	}
	m.Headers[key] = value
}

func (m *Message) AddRecipients(r ...Recipient) {
	m.To = append(m.To, r...)
}

func (m *Message) AddGlobalMergeVar(globalvars ...Var) {
	m.GlobalMergeVars = append(m.GlobalMergeVars, globalvars...)
}

func (m *Message) AddMergeVar(vars ...MergeVars) {
	m.MergeVars = append(m.MergeVars, vars...)
}

func (m *Message) AddTag(tags ...string) {
	m.Tags = append(m.Tags, tags...)
}

func (m *Message) AddGoogleAnalyticsDomains(domains ...string) {
	m.GoogleAnalyticsDomains = append(m.GoogleAnalyticsDomains, domains...)
}

func (m *Message) AddGoogleAnalyticsCampaign(campaigns ...string) {
	m.GoogleAnalyticsCampaign = append(m.GoogleAnalyticsCampaign, campaigns...)
}

func (m *Message) AddMetadata(key, value string) {
	if m.Metadata == nil {
		m.Metadata = make(map[string]string)
	}
	m.Metadata[key] = value
}

func (m *Message) AddRecipientMetadata(metadata ...RecipientMetaData) {
	m.RecipientMetadata = append(m.RecipientMetadata, metadata...)
}

func (m *Message) AddAttachments(attachement ...Attachment) {
	m.Attachments = append(m.Attachments, attachement...)
}

type Attachment struct {
	Type    string `json:"type"`
	Name    string `json:"name"`
	Content string `json:"content"`
}
type RecipientMetaData struct {
	Recipient string            `json:"rcpt"`
	Vars      map[string]string `json:"values"`
}

type MergeVars struct {
	Recipient string `json:"rcpt"`
	Vars      []Var  `json:"vars"`
}

type Var struct {
	Name    string      `json:"name"`
	Content interface{} `json:"content"`
}

func NewVar(name string, content interface{}) *Var {
	return &Var{Name: name, Content: content}
}

type Recipient struct {
	Email string `json:"email"`
	Name  string `json:"name"`
	Type  string `json:"type"`
}

type SendResponse struct {
	Email          string `json:"email"`
	Status         string `json:"status"`
	Id             string `json:"_id"`
	RejectedReason string `json:"reject_reason"`
}

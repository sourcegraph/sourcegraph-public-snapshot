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
	"errors"
	"time"
)

// see https://mandrillapp.com/api/docs/webhooks.html
const webhooks_list_endpoint string = "/webhooks/list.json"     //Get the list of all webhooks defined on the account
const webhooks_add_endpoint string = "/webhooks/add.json"       //Add a new webhook
const webhooks_info_endpoint string = "/webhooks/info.json"     //Given the ID of an existing webhook, return the data about it
const webhooks_update_endpoint string = "/webhooks/update.json" //Update an existing webhook
const webhooks_delete_endpoint string = "/webhooks/delete.json" //Delete an existing webhook

// can error with one of the following: Invalid_Key, ValidationError, GeneralError
func (a *MandrillAPI) WebhooksList() (Webhook, error) {
	var params map[string]interface{} = make(map[string]interface{})
	return getWebhook(a, params, webhooks_list_endpoint)
}

// can error with one of the following: Invalid_Key, ValidationError, GeneralError
func (a *MandrillAPI) WebhookAdd(url string, events []string) (Webhook, error) {
	if url == "" {
		return Webhook{}, errors.New("url cannot be blank")
	}
	var params map[string]interface{} = make(map[string]interface{})
	params["url"] = url
	params["events"] = events
	return getWebhook(a, params, webhooks_add_endpoint)
}

// can error with one of the following: Unknown_Webhook, Invalid_Key, ValidationError, GeneralError
func (a *MandrillAPI) WebhookInfo(id int) (Webhook, error) {
	if id <= 0 {
		return Webhook{}, errors.New("id must be >= 0")
	}
	var params map[string]interface{} = make(map[string]interface{})
	params["id"] = id
	return getWebhook(a, params, webhooks_info_endpoint)
}

// can error with one of the following: Unknown_Webhook, Invalid_Key, ValidationError, GeneralError
func (a *MandrillAPI) WebhookUpdate(url string, events []string) (Webhook, error) {
	if url == "" {
		return Webhook{}, errors.New("url cannot be blank")
	}
	var params map[string]interface{} = make(map[string]interface{})
	params["url"] = url
	params["events"] = events
	return getWebhook(a, params, webhooks_delete_endpoint)
}

// can error with one of the following: Unknown_Webhook, Invalid_Key, ValidationError, GeneralError
func (a *MandrillAPI) WebhookDelete(id int) (Webhook, error) {
	if id <= 0 {
		return Webhook{}, errors.New("id must be >= 0")
	}
	var params map[string]interface{} = make(map[string]interface{})
	params["id"] = id
	return getWebhook(a, params, webhooks_delete_endpoint)
}

func getWebhook(a *MandrillAPI, params map[string]interface{}, endpoint string) (Webhook, error) {
	var response Webhook
	err := parseMandrillJson(a, endpoint, params, &response)
	return response, err
}

type Webhook struct {
	Id          int       `json:"id"`
	Url         string    `json:"url"`
	Events      []string  `json:"events"`
	CreatedAt   time.Time `json:"created_at"`
	LastSentAt  time.Time `json:"last_sent_at"`
	BatchesSent int       `json:"batches_sent"`
	EventsSent  int       `json:"events_sent"`
	LastError   string    `json:"last_error"`
}

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

// see https://mandrillapp.com/api/docs/senders.html
const senders_list_endpoint string = "/senders/list.json"               //Return the senders that have tried to use this account.
const senders_domains_endpoint string = "/senders/domains.json"         //Returns the sender domains that have been added to this account.
const senders_info_endpoint string = "/senders/info.json"               //Return more detailed information about a single sender, including aggregates of recent stats
const senders_time_series_endpoint string = "/senders/time-series.json" //Return the recent history (hourly stats for the last 30 days) for a sender

// can error with one of the following: Invalid_Key, ValidationError, GeneralError
func (a *MandrillAPI) SenderList() ([]Sender, error) {
	var response []Sender
	var params map[string]interface{} = make(map[string]interface{})
	err := parseMandrillJson(a, senders_list_endpoint, params, &response)
	return response, err
}

// can error with one of the following: Invalid_Key, ValidationError, GeneralError
func (a *MandrillAPI) SenderDomains() ([]Domain, error) {
	var response []Domain
	var params map[string]interface{} = make(map[string]interface{})
	err := parseMandrillJson(a, senders_domains_endpoint, params, &response)
	return response, err
}

// can error with one of the following: Unknown_Sender, Invalid_Key, ValidationError, GeneralError
func (a *MandrillAPI) SenderInfo(address string) (SenderInfo, error) {
	var response SenderInfo
	if address == "" {
		return response, errors.New("address cannot be blank")
	}
	var params map[string]interface{} = make(map[string]interface{})
	params["address"] = address
	err := parseMandrillJson(a, senders_info_endpoint, params, &response)
	return response, err
}

// can error with one of the following: Unknown_Sender, Invalid_Key, ValidationError, GeneralError
func (a *MandrillAPI) SenderTimeSeries(address string) ([]TimeSeries, error) {
	var response []TimeSeries
	if address == "" {
		return response, errors.New("address cannot be blank")
	}
	var params map[string]interface{} = make(map[string]interface{})
	params["address"] = address
	err := parseMandrillJson(a, senders_time_series_endpoint, params, &response)
	return response, err
}

type SenderInfo struct {
	Address     string    `json:"address"`
	CreatedAt   time.Time `json:"created_at"`
	Sent        int32     `json:"sent"`
	HardBounces int32     `json:"hard_bounces"`
	SoftBounces int32     `json:"soft_bounces"`
	Rejects     int32     `json:"rejects"`
	Complaints  int32     `json:"complaints"`
	Unsubs      int32     `json:"unsubs"`
	Opens       int32     `json:"opens"`
	Clicks      int32     `json:"clicks"`
	Stats       []Stat    `json:"stats"`
}

type Domain struct {
	Domain    string    `json:"domain"`
	CreatedAt time.Time `json:"created_at"`
}

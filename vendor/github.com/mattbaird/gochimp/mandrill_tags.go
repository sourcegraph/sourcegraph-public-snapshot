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
)

// see https://mandrillapp.com/api/docs/tags.html
const tags_list_endpoint string = "/tags/list.json"                       // Return all of the user-defined tag information
const tags_info_endpoint string = "/tags/info.json"                       // Return more detailed information about a single tag, including aggregates of recent stats
const tags_time_series_endpoint string = "/tags/time-series.json"         // Return the recent history (hourly stats for the last 30 days) for a tag
const tags_all_time_series_endpoint string = "/tags/all-time-series.json" // Return the recent history (hourly stats for the last 30 days) for a tag

// can error with one of the following: Invalid_Key, ValidationError, GeneralError
func (a *MandrillAPI) TagList() ([]ListResponse, error) {
	var response []ListResponse
	var params map[string]interface{} = make(map[string]interface{})
	err := parseMandrillJson(a, tags_list_endpoint, params, &response)
	return response, err
}

// can error with one of the following: Invalid_Tag_Name, Invalid_Key, ValidationError, GeneralError
func (a *MandrillAPI) TagInfo(tag string) (TagInfo, error) {
	var response TagInfo
	if tag == "" {
		return response, errors.New("tag cannot be blank")
	}

	var params map[string]interface{} = make(map[string]interface{})
	params["tag"] = tag
	err := parseMandrillJson(a, tags_info_endpoint, params, &response)
	return response, err
}

// can error with one of the following: Invalid_Tag_Name, Invalid_Key, ValidationError, GeneralError
func (a *MandrillAPI) TagTimeSeries(tag string) ([]TimeSeries, error) {
	var response []TimeSeries
	if tag == "" {
		return response, errors.New("tag cannot be blank")
	}
	var params map[string]interface{} = make(map[string]interface{})
	params["tag"] = tag
	err := parseMandrillJson(a, tags_time_series_endpoint, params, &response)
	return response, err
}

// can error with one of the following: Invalid_Key, ValidationError, GeneralError
func (a *MandrillAPI) TagAllTimeSeries() ([]TimeSeries, error) {
	var response []TimeSeries
	var params map[string]interface{} = make(map[string]interface{})
	err := parseMandrillJson(a, tags_all_time_series_endpoint, params, &response)
	return response, err
}

type TimeSeries struct {
	Time         APITime `json:"time"`
	Sent         int32   `json:"sent"`
	HardBounces  int32   `json:"hard_bounces"`
	SoftBounces  int32   `json:"soft_bounces"`
	Rejects      int32   `json:"rejects"`
	Complaints   int32   `json:"complaints"`
	Unsubs       int32   `json:"unsubs"`
	Opens        int32   `json:"opens"`
	UniqueOpens  int32   `json:"unique_opens"`
	Clicks       int32   `json:"clicks"`
	UniqueClicks int32   `json:"unique_clicks"`
}

type TagInfo struct {
	Tag          string `json:"tag"`
	Reputation   int32  `json:"reputation,omitempty"`
	Sent         int32  `json:"sent"`
	HardBounces  int32  `json:"hard_bounces"`
	SoftBounces  int32  `json:"soft_bounces"`
	Rejects      int32  `json:"rejects"`
	Complaints   int32  `json:"complaints"`
	Unsubs       int32  `json:"unsubs"`
	Opens        int32  `json:"opens"`
	Clicks       int32  `json:"clicks"`
	UniqueOpens  int32  `json:"unique_opens"`
	UniqueClicks int32  `json:"unique_clicks"`
	Stats        []Stat `json:"stats,omitempty"`
}

type ListResponse struct {
	Tag         string `json:"tag"`
	Sent        int32  `json:"sent"`
	HardBounces int32  `json:"hard_bounces"`
	SoftBounces int32  `json:"soft_bounces"`
	Rejects     int32  `json:"rejects"`
	Complaints  int32  `json:"complaints"`
	Unsubs      int32  `json:"unsubs"`
	Opens       int32  `json:"opens"`
	Clicks      int32  `json:"clicks"`
}

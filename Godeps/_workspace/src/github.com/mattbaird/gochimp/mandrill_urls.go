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

// see https://mandrillapp.com/api/docs/urls.html
const urls_list_endpoint string = "/urls/list.json"               //Get the 100 most clicked URLs
const urls_search_endpoint string = "/urls/search.json"           //Return the 100 most clicked URLs that match the search query given
const urls_time_series_endpoint string = "/urls/time-series.json" //Return the recent history (hourly stats for the last 30 days) for a url

// can error with one of the following: Invalid_Key, ValidationError, GeneralError
func (a *MandrillAPI) UrlList() ([]UrlInfo, error) {
	var response []UrlInfo
	var params map[string]interface{} = make(map[string]interface{})
	err := parseMandrillJson(a, urls_list_endpoint, params, &response)
	return response, err
}

// can error with one of the following: Invalid_Key, ValidationError, GeneralError
func (a *MandrillAPI) UrlSearch(q string) ([]UrlInfo, error) {
	var response []UrlInfo
	if q == "" {
		return response, errors.New("query[q] cannot be blank")
	}
	var params map[string]interface{} = make(map[string]interface{})
	params["q"] = q
	err := parseMandrillJson(a, urls_search_endpoint, params, &response)
	return response, err
}

// can error with one of the following: Invalid_Key, ValidationError, GeneralError
func (a *MandrillAPI) UrlTimeSeries(url string) ([]UrlInfo, error) {
	var response []UrlInfo
	if url == "" {
		return response, errors.New("url cannot be blank")
	}
	var params map[string]interface{} = make(map[string]interface{})
	params["url"] = url
	err := parseMandrillJson(a, urls_time_series_endpoint, params, &response)
	return response, err
}

type UrlTimeSeriesInfo struct {
	Time         APITime `json:"time"`
	Sent         int     `json:"sent"`
	Clicks       int     `json:"clicks"`
	UniqueClicks int     `json:"unique_clicks"`
}

type UrlInfo struct {
	Url          string `json:"url"`
	Sent         int    `json:"sent"`
	Clicks       int    `json:"clicks"`
	UniqueClicks int    `json:"unique_clicks"`
}

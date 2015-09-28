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

const (
	reports_summary_endpoint string = "/reports/summary.json"
	reports_clicks_endpoint  string = "/reports/clicks.json"
)

func (a *ChimpAPI) GetSummary(req ReportsSummary) (ReportSummaryResponse, error) {
	req.ApiKey = a.Key
	var response ReportSummaryResponse
	err := parseChimpJson(a, reports_summary_endpoint, req, &response)
	return response, err
}

func (a *ChimpAPI) GetClicks(req ReportsClicks) (ReportClicksResponse, error) {
	req.ApiKey = a.Key
	var response ReportClicksResponse
	err := parseChimpJson(a, reports_clicks_endpoint, req, &response)
	return response, err
}

type ReportsClicks struct {
	ApiKey     string `json:"apikey"`
	CampaignId string `json:"cid"`
}

type ReportClicksResponse struct {
	Total []TrackedUrl `json:"total"`
}

type TrackedUrl struct {
	Url           string  `json:"url"`
	Clicks        int     `json:"clicks"`
	ClicksPercent float32 `json:"clicks_percent"`
	Unique        int     `json:"unique"`
	UniquePercent float32 `json:"unique_percent"`
}

type ReportsSummary struct {
	ApiKey     string `json:"apikey"`
	CampaignId string `json:"cid"`
}

type ReportSummaryResponse struct {
	HardBounce   int         `json:"hard_bounces"`
	SoftBounce   int         `json:"soft_bounces"`
	Unsubscribes int         `json:"unsubscribes"`
	AbuseReports int         `json:"abuse_reports"`
	Opens        int         `json:"opens"`
	UniqueOpens  int         `json:"unique_opens"`
	Clicks       int         `json:"clicks"`
	UniqueClicks int         `json:"unique_clicks"`
	EmailsSent   int         `json:"emails_sent"`
	TimeSeries   []TimeSerie `json:"timeseries"`
}

type TimeSerie struct {
	TimeStamp       string `json:"timestamp"`
	EmailsSent      int    `json:"emails_sent"`
	UniqueOpens     int    `json:"unique_opens"`
	RecipientsClick int    `json:"recipients_click"`
}

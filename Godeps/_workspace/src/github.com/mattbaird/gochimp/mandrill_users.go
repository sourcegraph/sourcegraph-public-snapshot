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
)

// see https://mandrillapp.com/api/docs/users.html
const users_info_endpoint string = "/users/info.json"       //Return the information about the API-connected user
const users_ping_endpoint string = "/users/ping.json"       // returns "PONG!"
const users_ping2_endpoint string = "/users/ping2.json"     // returns 'PING':'PONG!' for anal json parsers
const users_senders_endpoint string = "/users/senders.json" // Return the senders that have tried to use this account, both verified and unverified

func (a *MandrillAPI) Ping() (string, error) {
	return parseString(runMandrill(a, users_ping_endpoint, nil))
}

func (a *MandrillAPI) UserInfo() (Info, error) {
	var info Info
	err := parseMandrillJson(a, users_info_endpoint, nil, &info)
	return info, err
}

func (a *MandrillAPI) UserSenders() ([]Sender, error) {
	var senders []Sender
	err := parseMandrillJson(a, users_senders_endpoint, nil, &senders)
	return senders, err
}

type Info struct {
	Username    string          `json:"username"`
	CreatedAt   APITime         `json:"created_at"`
	PublicId    string          `json:"public_id"`
	Reputation  int             `json:"reputation"`
	HourlyQuota int             `json:"hourly_quota"`
	Backlog     int             `json:"backlog"`
	Stats       map[string]Stat `json:"stats"`
}

type Stat struct {
	Sent         int `json:"sent"`
	HardBounces  int `json:"hard_bounces"`
	SoftBounces  int `json:"soft_bounces"`
	Rejects      int `json:"rejects"`
	Complaints   int `json:"complaints"`
	Unsubs       int `json:"unsubs"`
	Opens        int `json:"opens"`
	UniqueOpens  int `json:"unique_opens"`
	Clicks       int `json:"clicks"`
	UniqueClicks int `json:"unique_clicks"`
}

type Sender struct {
	Sent         int     `json:"sent"`
	HardBounces  int     `json:"hard_bounces"`
	SoftBounces  int     `json:"soft_bounces"`
	Rejects      int     `json:"rejects"`
	Complaints   int     `json:"complaints"`
	Unsubs       int     `json:"unsubs"`
	Opens        int     `json:"opens"`
	Clicks       int     `json:"clicks"`
	UniqueOpens  int     `json:"unique_opens"`
	UniqueClicks int     `json:"unique_clicks"`
	Reputation   int     `json:"reputation"`
	Address      string  `json:"address"`
	CreatedAt    APITime `json:"created_at"`
}

func (s *Sender) String() string {
	b, _ := json.Marshal(s)
	return string(b)
}

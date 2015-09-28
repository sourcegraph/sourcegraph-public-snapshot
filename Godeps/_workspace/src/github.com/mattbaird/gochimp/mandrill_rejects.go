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
	"log"
)

// see https://mandrillapp.com/api/docs/rejects.html
//Retrieves your email rejection blacklist. You can provide an email address to limit the results.
// Returns up to 1000 results. By default, entries that have expired are excluded from the results;
// set include_expired to true to include them.
const rejects_list_endpoint string = "/rejects/list.json"

//Deletes an email rejection. There is no limit to how many rejections you can remove from your
// blacklist, but keep in mind that each deletion has an affect on your reputation.
const rejects_delete_endpoint string = "/rejects/delete.json"

// can error with one of the following: Invalid_Key, ValidationError, GeneralError
func (a *MandrillAPI) RejectsList(email string, includeExpired bool) ([]Reject, error) {
	var response []Reject
	if email == "" {
		return response, errors.New("email cannot be blank")
	}
	var params map[string]interface{} = make(map[string]interface{})
	params["email"] = email
	params["include_expired"] = includeExpired
	err := parseMandrillJson(a, rejects_list_endpoint, params, &response)
	return response, err
}

// can error with one of the following: Invalid_Reject, Invalid_Key, ValidationError, GeneralError
func (a *MandrillAPI) RejectsDelete(email string) (bool, error) {
	var response map[string]interface{}
	var retval bool = false
	if email == "" {
		return retval, errors.New("email cannot be blank")
	}
	var params map[string]interface{} = make(map[string]interface{})
	params["email"] = email
	err := parseMandrillJson(a, rejects_delete_endpoint, params, &response)
	var ok bool = false
	if err == nil {
		retval, ok = response["deleted"].(bool)
		if ok != true {
			log.Fatal("Received response with deleted parameter, however type was not bool, this should not happen")
		}
	}
	return retval, err
}

type Reject struct {
	Email       string  `json:"email"`
	Reason      string  `json:"reason"`
	Detail      string  `json:"detail"`
	CreatedAt   APITime `json:"created_at"`
	LastEventAt APITime `json:"last_event_at"`
	ExpiresAt   APITime `json:"expires_at"`
	Expired     bool    `json:"expired"`
	Sender      Sender  `json:"sender"`
}

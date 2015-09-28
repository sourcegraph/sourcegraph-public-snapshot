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
	"fmt"
)

/*
 four types of error
 Invalid_Key
 - The provided API key is not a valid Mandrill API key
 ValidationError
 - The parameters passed to the API call are invalid or not provided when required
 GeneralError
 - An unexpected error occurred processing the request. Mandrill developers will be notified.
 Unknown_Template
 - The requested template does not exist
*/
type MandrillError struct {
	Status  string `json:"status"`
	Code    int    `json:"code"`
	Name    string `json:"name"`
	Message string `json:"message"`
}

func (e MandrillError) Error() string {
	return fmt.Sprintf("%v: %v", e.Code, e.Message)
}

func mandrillErrorCheck(body []byte) error {
	var e MandrillError
	err := json.Unmarshal(body, &e)
	if err == nil {
		// it may have parsed successfully, however there if
		// there is no message or error code, it's not an error
		if e.Message != "" || e.Code > 0 {
			return e
		}
	}
	return nil
}

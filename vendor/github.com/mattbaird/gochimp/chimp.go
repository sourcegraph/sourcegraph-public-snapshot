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
	"time"
)

type APIError struct {
	Status string `json:"status"`
	Code   int    `json:"code"`
	Name   string `json:"name"`
	Err    string `json:"error"`
}

func (e APIError) Error() string {
	return fmt.Sprintf("%v: %v", e.Code, e.Err)
}

func chimpErrorCheck(body []byte) error {
	var e APIError
	json.Unmarshal(body, &e)
	if e.Err != "" || e.Code != 0 {
		return e
	}
	return nil
}

//Mailchimp does not conform to RFC3339 format, so we need custom UnmarshalJSON
type APITime struct {
	time.Time
}

func (t *APITime) UnmarshalJSON(data []byte) (err error) {
	s := string(data)
	l := len(s)
	switch {
	case l == 12:
		t.Time, err = time.Parse(`"2006-01-02"`, s)
	case l == 21:
		t.Time, err = time.Parse(`"2006-01-02 15:04:05"`, s)
	case l == 27:
		t.Time, err = time.Parse(`"2006-01-02 15:04:05.00000"`, s)
	case l == 9:
		t.Time, err = time.Parse(`"2006-01"`, s)
	}
	return
}

//format string for time.Format
const APITimeFormat = "2006-01-02 15:04:05"

func apiTime(t interface{}) interface{} {
	switch ti := t.(type) {
	case time.Time:
		return ti.Format(APITimeFormat)
	case string:
		return ti
	}
	return t
}

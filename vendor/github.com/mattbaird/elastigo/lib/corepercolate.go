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

package elastigo

import (
	"encoding/json"
	"fmt"
)

type PercolatorResult struct {
	SearchResult
	Matches []PercolatorMatch `json:"matches"`
}

type PercolatorMatch struct {
	Index string `json:"_index"`
	Id    string `json:"_id"`
}

// See http://www.elasticsearch.org/guide/reference/api/percolate.html
func (c *Conn) RegisterPercolate(index string, id string, data interface{}) (BaseResponse, error) {
	var url string
	var retval BaseResponse
	url = fmt.Sprintf("/%s/.percolator/%s", index, id)
	body, err := c.DoCommand("PUT", url, nil, data)
	if err != nil {
		return retval, err
	}
	if err == nil {
		// marshall into json
		jsonErr := json.Unmarshal(body, &retval)
		if jsonErr != nil {
			return retval, jsonErr
		}
	}
	return retval, err
}

func (c *Conn) Percolate(index string, _type string, name string, args map[string]interface{}, doc string) (PercolatorResult, error) {
	var url string
	var retval PercolatorResult
	url = fmt.Sprintf("/%s/%s/_percolate", index, _type)
	body, err := c.DoCommand("GET", url, args, doc)
	if err != nil {
		return retval, err
	}
	if err == nil {
		// marshall into json
		jsonErr := json.Unmarshal(body, &retval)
		if jsonErr != nil {
			return retval, jsonErr
		}
	}
	return retval, err
}

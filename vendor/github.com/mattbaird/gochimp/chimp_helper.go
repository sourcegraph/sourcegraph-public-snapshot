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
	helper_inline_css = "/helper/inline-css.json"
)

func (a *ChimpAPI) InlineCSS(req InlineCSSRequest) (InlineCSSResponse, error) {
	var response InlineCSSResponse
	req.ApiKey = a.Key
	err := parseChimpJson(a, helper_inline_css, req, &response)
	return response, err
}

type InlineCSSRequest struct {
	ApiKey   string `json:"apikey"`
	HTML     string `json:"html"`
	StripCSS bool   `json:"strip_html"`
}

type InlineCSSResponse struct {
	HTML string `json:"html"`
}

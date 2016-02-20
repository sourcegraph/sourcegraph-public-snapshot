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

/*
The API package provides the basic support for using HTTP to talk to the Mandrill and Mailchimp API's.
Each Struct contains a Key, Transport and endpoint property
*/
package gochimp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
)

const (
	mandrill_uri     = "mandrillapp.com"
	mandrill_version = "/api/1.0"
)

type MandrillAPI struct {
	Key       string
	Transport http.RoundTripper
	endpoint  string
}

type ChimpAPI struct {
	Key       string
	Transport http.RoundTripper
	endpoint  string
}

// see https://mandrillapp.com/api/docs/
// currently supporting json output formats
func NewMandrill(apiKey string) (*MandrillAPI, error) {
	u := url.URL{}
	u.Scheme = "https"
	u.Host = mandrill_uri
	u.Path = mandrill_version
	return &MandrillAPI{Key: apiKey, endpoint: u.String()}, nil
}

const mailchimp_uri string = "%s.api.mailchimp.com"
const mailchimp_version string = "/2.0"
const debug bool = false

var mailchimp_datacenter = regexp.MustCompile("[a-z]+[0-9]+$")

func NewChimp(apiKey string, https bool) *ChimpAPI {
	u := url.URL{}
	if https {
		u.Scheme = "https"
	} else {
		u.Scheme = "http"
	}
	u.Host = fmt.Sprintf("%s.api.mailchimp.com", mailchimp_datacenter.FindString(apiKey))
	u.Path = mailchimp_version
	return &ChimpAPI{Key: apiKey, endpoint: u.String()}
}

func runChimp(api *ChimpAPI, path string, parameters interface{}) ([]byte, error) {
	b, err := json.Marshal(parameters)
	if err != nil {
		return nil, err
	}
	requestUrl := fmt.Sprintf("%s%s", api.endpoint, path)
	if debug {
		log.Printf("Request URL:%s", requestUrl)
	}
	client := &http.Client{Transport: api.Transport}
	resp, err := client.Post(requestUrl, "application/json", bytes.NewBuffer(b))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if debug {
		log.Printf("Response Body:%s", string(body))
	}
	if err = chimpErrorCheck(body); err != nil {
		return nil, err
	}
	return body, nil
}

func runMandrill(api *MandrillAPI, path string, parameters map[string]interface{}) ([]byte, error) {
	if parameters == nil {
		parameters = make(map[string]interface{})
	}
	parameters["key"] = api.Key
	b, err := json.Marshal(parameters)
	if debug {
		log.Printf("Payload:%s", string(b))
	}
	if err != nil {
		return nil, err
	}
	requestUrl := fmt.Sprintf("%s%s", api.endpoint, path)
	if debug {
		log.Printf("Request URL:%s", requestUrl)
	}
	client := &http.Client{Transport: api.Transport}
	resp, err := client.Post(requestUrl, "application/json", bytes.NewBuffer(b))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if debug {
		log.Printf("Response Body:%s", string(body))
	}
	if err = mandrillErrorCheck(body); err != nil {
		return nil, err
	}
	return body, nil
}

func parseString(body []byte, err error) (string, error) {
	if err != nil {
		return "", err
	}
	return strconv.Unquote(string(body))
}

func parseMandrillJson(api *MandrillAPI, path string, parameters map[string]interface{}, retval interface{}) error {
	body, err := runMandrill(api, path, parameters)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(body, retval); err != nil {
		return err
	}

	return nil
}

func parseChimpJson(api *ChimpAPI, method string, parameters interface{}, retval interface{}) error {
	body, err := runChimp(api, method, parameters)

	if err != nil {
		return err
	}
	if retval != nil {
		return parseJson(body, retval)
	}
	return nil
}

type JsonAlterer interface {
	alterJson(b []byte) []byte
}

func parseJson(body []byte, retval interface{}) error {
	switch r := retval.(type) {
	case JsonAlterer:
		return json.Unmarshal(r.alterJson(body), retval)
	default:
		return json.Unmarshal(body, retval)
	}
}

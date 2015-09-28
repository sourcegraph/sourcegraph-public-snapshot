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

// see https://mandrillapp.com/api/docs/inbound.JSON.html
const inbound_domains_endpoint string = "/inbound/domains.json"     //List the domains that have been configured for inbound delivery
const add_domain_endpoint string = "/inbound/add-domain.json"       //Add an inbound domain to your account
const check_domain_endpoint string = "/inbound/check-domain.json"   //Check the MX settings for an inbound domain.
const delete_domain_endpoint string = "/inbound/delete-domain.json" //Delete an inbound domain from the account.
const routes_endopoint string = "/inbound/routes.json"              //List the mailbox routes defined for an inbound domain
const add_route_endpoint string = "/inbound/add-route.json"         //Add a new mailbox route to an inbound domain
const update_route_endpoint string = "/inbound/update-route.json"   //Update the pattern or webhook of an existing inbound mailbox route.
const delete_route_endpoint string = "/inbound/delete-route.json"   //Delete an existing inbound mailbox route
const send_raw_endpoint string = "/inbound/send-raw.json"           //Send a raw MIME message to an inbound hook as if it had been sent over SMTP.

// can error with one of the following: Invalid_Key, ValidationError, GeneralError
func (a *MandrillAPI) InboundDomainList() ([]InboundDomain, error) {
	var response []InboundDomain
	var params map[string]interface{} = make(map[string]interface{})
	err := parseMandrillJson(a, inbound_domains_endpoint, params, &response)
	return response, err
}

// can error with one of the following: Invalid_Key, ValidationError, GeneralError
func (a *MandrillAPI) InboundDomainAdd(domain string) (InboundDomain, error) {
	return getInboundDomain(a, domain, add_domain_endpoint)
}

// can error with one of the following: Invalid_Key, Unknown_InboundDomain, ValidationError, GeneralError
func (a *MandrillAPI) InboundDomainCheck(domain string) (InboundDomain, error) {
	return getInboundDomain(a, domain, check_domain_endpoint)
}

// can error with one of the following: Invalid_Key, Unknown_InboundDomain, ValidationError, GeneralError
func (a *MandrillAPI) InboundDomainDelete(domain string) (InboundDomain, error) {
	return getInboundDomain(a, domain, delete_domain_endpoint)
}

// can error with one of the following: Invalid_Key, Unknown_InboundDomain, ValidationError, GeneralError
func (a *MandrillAPI) RouteList(domain string) ([]Route, error) {
	var response []Route
	if domain == "" {
		return response, errors.New("domain cannot be blank")
	}
	var params map[string]interface{} = make(map[string]interface{})
	params["domain"] = domain
	err := parseMandrillJson(a, routes_endopoint, params, &response)
	return response, err
}

// can error with one of the following: Invalid_Key, Unknown_InboundDomain, ValidationError, GeneralError
func (a *MandrillAPI) RouteAdd(domain string, pattern string, url string) (Route, error) {
	var response Route
	if domain == "" {
		return response, errors.New("domain cannot be blank")
	}
	if pattern == "" {
		return response, errors.New("pattern cannot be blank")
	}
	if url == "" {
		return response, errors.New("url cannot be blank")
	}
	return getRoute(a, "", domain, pattern, url, add_route_endpoint)
}

// can error with one of the following: Invalid_Key, Unknown_InboundDomain, ValidationError, GeneralError
func (a *MandrillAPI) RouteUpdate(id string, domain string, pattern string, url string) (Route, error) {
	var response Route
	if id == "" {
		return response, errors.New("id cannot be blank")
	}
	return getRoute(a, id, domain, pattern, url, update_route_endpoint)
}

// can error with one of the following: Invalid_Key, Unknown_InboundDomain, ValidationError, GeneralError
func (a *MandrillAPI) RouteDelete(id string) (Route, error) {
	var response Route
	if id == "" {
		return response, errors.New("id cannot be blank")
	}
	return getRoute(a, id, "", "", "", delete_route_endpoint)
}

// can error with one of the following: Invalid_Key, ValidationError, GeneralError
func (a *MandrillAPI) SendRawMIME(raw_message string, to []string, mail_from string, helo string, client_address string) ([]InboundRecipient, error) {
	var response []InboundRecipient
	if raw_message == "" {
		return response, errors.New("raw_message cannot be blank")
	}
	var params map[string]interface{} = make(map[string]interface{})
	params["raw_message"] = raw_message
	if len(to) > 0 {
		params["to"] = to
	}
	if mail_from != "" {
		params["mail_from"] = mail_from
	}
	if helo != "" {
		params["helo"] = helo
	}
	if client_address != "" {
		params["client_address"] = client_address
	}

	err := parseMandrillJson(a, send_raw_endpoint, params, &response)
	return response, err
}

func getRoute(a *MandrillAPI, id string, domain string, pattern string, url string, endpoint string) (Route, error) {
	var params map[string]interface{} = make(map[string]interface{})
	var response Route
	if id != "" {
		params["id"] = id
	}
	if domain != "" {
		params["domain"] = domain
	}
	if pattern != "" {
		params["pattern"] = pattern
	}
	if url != "" {
		params["url"] = url
	}

	err := parseMandrillJson(a, endpoint, params, &response)
	return response, err
}

func getInboundDomain(a *MandrillAPI, domain string, endpoint string) (InboundDomain, error) {
	if domain == "" {
		return InboundDomain{}, errors.New("domain cannot be blank")
	}
	var params map[string]interface{} = make(map[string]interface{})
	params["domain"] = domain
	var response InboundDomain
	err := parseMandrillJson(a, endpoint, params, &response)
	return response, err
}

type InboundDomain struct {
	Domain    string  `json:"domain"`
	CreatedAt APITime `json:"created_at"`
	ValidMx   bool    `json:"valid_mx"`
}

type Route struct {
	Id      string `json:"id"`
	Pattern string `json:"pattern"`
	Url     string `json:"url"`
}

type InboundRecipient struct {
	Email   string `json:"email"`
	Pattern string `json:"pattern"`
	Url     string `json:"url"`
}

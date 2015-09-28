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

// see https://mandrillapp.com/api/docs/templates.html
const templates_add_endpoint string = "/templates/add.json"       //Add a new template
const templates_info_endpoint string = "/templates/info.json"     //Get the information for an existing template
const templates_update_endpoint string = "/templates/update.json" //Update the code for an existing template
// Publish the content for the template. Any new messages sent using this template will start
//using the content that was previously in draft.
const templates_publish_endpoint string = "/templates/publish.json"
const templates_delete_endpoint string = "/templates/delete.json"           //Delete a template
const templates_list_endpoint string = "/templates/list.json"               //Return a list of all the templates available to this user
const templates_time_series_endpoint string = "/templates/time-series.json" //Return the recent history (hourly stats for the last 30 days) for a template
const templates_render_endpoint string = "/templates/render.json"           //Inject content and optionally merge fields into a template, returning the HTML that results

// can error with one of the following: Unknown_Template, Invalid_Key, ValidationError, GeneralError
func (a *MandrillAPI) TemplateAdd(name string, code string, publish bool) (Template, error) {
	if name == "" {
		return Template{}, errors.New("name cannot be blank")
	}
	if code == "" {
		return Template{}, errors.New("code cannot be blank")
	}
	var params map[string]interface{} = make(map[string]interface{})
	params["name"] = name
	params["code"] = code
	params["publish"] = publish
	return execute(a, params, templates_add_endpoint)
}

// can error with one of the following: Unknown_Template, Invalid_Key, ValidationError, GeneralError
func (a *MandrillAPI) TemplateInfo(name string) (Template, error) {
	if name == "" {
		return Template{}, errors.New("name cannot be blank")
	}
	var params map[string]interface{} = make(map[string]interface{})
	params["name"] = name
	return execute(a, params, templates_info_endpoint)
}

// can error with one of the following: Unknown_Template, Invalid_Key, ValidationError, GeneralError
func (a *MandrillAPI) TemplateUpdate(name string, code string, publish bool) (Template, error) {
	if name == "" {
		return Template{}, errors.New("name cannot be blank")
	}
	if code == "" {
		return Template{}, errors.New("code cannot be blank")
	}
	var params map[string]interface{} = make(map[string]interface{})
	params["name"] = name
	params["code"] = code
	params["publish"] = publish
	return execute(a, params, templates_update_endpoint)
}

// can error with one of the following: Unknown_Template, Invalid_Key, ValidationError, GeneralError
func (a *MandrillAPI) TemplatePublish(name string) (Template, error) {
	if name == "" {
		return Template{}, errors.New("name cannot be blank")
	}
	var params map[string]interface{} = make(map[string]interface{})
	params["name"] = name
	return execute(a, params, templates_publish_endpoint)
}

// can error with one of the following: Unknown_Template, Invalid_Key, ValidationError, GeneralError
func (a *MandrillAPI) TemplateDelete(name string) (Template, error) {
	if name == "" {
		return Template{}, errors.New("name cannot be blank")
	}
	var params map[string]interface{} = make(map[string]interface{})
	params["name"] = name
	return execute(a, params, templates_delete_endpoint)
}

// can error with one of the following: Invalid_Key, ValidationError, GeneralError
func (a *MandrillAPI) TemplateList() ([]Template, error) {
	var response []Template
	var params map[string]interface{} = make(map[string]interface{})
	err := parseMandrillJson(a, templates_list_endpoint, params, &response)
	return response, err
}

// can error with one of the following: Unknown_Template, Invalid_Key, ValidationError, GeneralError
func (a *MandrillAPI) TemplateTimeSeries(name string) ([]Template, error) {
	if name == "" {
		return []Template{}, errors.New("name cannot be blank")
	}
	var response []Template
	var params map[string]interface{} = make(map[string]interface{})
	params["name"] = name
	err := parseMandrillJson(a, templates_time_series_endpoint, params, &response)
	return response, err
}

// can error with one of the following: Unknown_Template, Invalid_Key, ValidationError, GeneralError
func (a *MandrillAPI) TemplateRender(templateName string, templateContent []Var, mergeVars []Var) (string, error) {
	if templateName == "" {
		return "", errors.New("templateName cannot be blank")
	}
	var response map[string]interface{}
	var params map[string]interface{} = make(map[string]interface{})
	params["template_name"] = templateName
	params["template_content"] = templateContent
	params["merge_vars"] = mergeVars
	err := parseMandrillJson(a, templates_render_endpoint, params, &response)
	var retval string = ""
	var ok bool = false
	if err == nil {
		retval, ok = response["html"].(string)
		if ok != true {
			log.Fatal("Received response with html parameter, however type was not string, this should not happen")
		}
	}
	return retval, err
}

func execute(a *MandrillAPI, params map[string]interface{}, endpoint string) (Template, error) {
	var response Template
	err := parseMandrillJson(a, endpoint, params, &response)
	return response, err
}

type Template struct {
	Name        string  `json:"name"`
	Code        string  `json:"code"`
	PublishName string  `json:"publish_name"`
	PublishCode string  `json:"publish_code"`
	CreatedAt   APITime `json:"published_at"`
	UpdateAt    APITime `json:"updated_at"`
}

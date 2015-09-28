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
	"fmt"
	"io/ioutil"
	"os"
	"testing"
)

var mandrill, err = NewMandrill(os.Getenv("MANDRILL_KEY"))
var user string = os.Getenv("MANDRILL_USER")

func TestPing(t *testing.T) {
	response, err := mandrill.Ping()
	if response != "PONG!" {
		t.Error(fmt.Sprintf("failed to return PONG!, returned [%s]", response), err)
	}
}

func TestUserInfo(t *testing.T) {
	response, err := mandrill.UserInfo()
	if err != nil {
		t.Error("Error:", err)
	}
	if response.Username != user {
		t.Error("wrong user")
	}
}

func TestUserSenders(t *testing.T) {
	response, err := mandrill.UserSenders()
	if response == nil {
		t.Error("response was nil", err)
	}
	if err != nil {
		t.Error("Error:", err)
	}
}

func TestMessageSending(t *testing.T) {
	var message Message = Message{Html: "<b>hi there</b>", Text: "hello text", Subject: "Test Mail", FromEmail: user,
		FromName: user}
	message.AddRecipients(Recipient{Email: user, Name: user, Type: "to"})
	message.AddRecipients(Recipient{Email: user, Name: user, Type: "cc"})
	message.AddRecipients(Recipient{Email: user, Name: user, Type: "bcc"})
	response, err := mandrill.MessageSend(message, false)
	if err != nil {
		t.Error("Error:", err)
	}
	if response != nil && len(response) > 0 {
		if len(response) != 3 {
			t.Errorf("Did not send to all users. Expected 3, got %d", len(response))
		} else {
			if response[0].Email != user || response[1].Email != user || response[2].Email != user {
				t.Errorf(
					"Wrong email recipient, expecting %s,, got (%s, %s, %s)",
					user,
					response[0].Email,
					response[1].Email,
					response[2].Email,
				)
			}
		}
	} else {
		t.Errorf("No response, probably due to API KEY issues")
	}
}

const testTemplateName string = "test_transactional_template"

func TestTemplateAdd(t *testing.T) {
	// delete the test template if it exists already
	mandrill.TemplateDelete(testTemplateName)
	template, err := mandrill.TemplateAdd(testTemplateName, readTemplate("templates/transactional_basic.html"), true)
	if err != nil {
		t.Error("Error:", err)
	}
	if template.Name != "test_transactional_template" {
		t.Errorf("Wrong template name, expecting %s, got %s", testTemplateName, template.Name)
	}
	// try recreating, should error out
	template, err = mandrill.TemplateAdd(testTemplateName, readTemplate("templates/transactional_basic.html"), true)
	if err == nil {
		t.Error("Should have error'd on duplicate template")
	}
}

func TestTemplateList(t *testing.T) {
	_, err := mandrill.TemplateAdd("listTest", "testing 123", true)
	if err != nil {
		t.Error("Error:", err)
	}
	templates, err := mandrill.TemplateList()
	if err != nil {
		t.Error("Error:", err)
	}
	if len(templates) <= 0 {
		t.Errorf("Should have retrieved templates")
	}
	mandrill.TemplateDelete("listTest")
}

func TestTemplateInfo(t *testing.T) {
	template, err := mandrill.TemplateInfo(testTemplateName)
	if err != nil {
		t.Error("Error:", err)
	}
	if template.Name != "test_transactional_template" {
		t.Errorf("Wrong template name, expecting %s, got %s", testTemplateName, template.Name)
	}
}

func TestTemplateUpdate(t *testing.T) {
	// add a simple template
	template, err := mandrill.TemplateAdd("updateTest", "testing 123", true)
	template, err = mandrill.TemplateUpdate("updateTest", "testing 321", true)
	if err != nil {
		t.Error("Error:", err)
	}
	if template.Name != "updateTest" {
		t.Errorf("Wrong template name, expecting %s, got %s", "updateTest", template.Name)
	}
	if template.Code != "testing 321" {
		t.Errorf("Wrong template code, expecting %s, got %s", "testing 321", template.Code)
	}
	// be nice and tear down after test
	mandrill.TemplateDelete("updateTest")
}

func TestTemplatePublish(t *testing.T) {
	mandrill.TemplateDelete("publishTest")
	// add a simple template
	template, err := mandrill.TemplateAdd("publishTest", "testing 123", false)
	if err != nil {
		t.Error("Error:", err)
	}
	if template.Name != "publishTest" {
		t.Errorf("Wrong template name, expecting %s, got %v", testTemplateName, template.Name)
	}
	if template.PublishCode != "" {
		t.Errorf("Template should not have a publish code, got %v", template.PublishCode)
	}
	template, err = mandrill.TemplatePublish("publishTest")
	if err != nil {
		t.Error("Error:", err)
	}
	if template.Name != "publishTest" {
		t.Errorf("Wrong template name, expecting %s, got %v", testTemplateName, template.Name)
	}
	if template.PublishCode == "" {
		t.Errorf("Template should have a publish code, got %v", template.PublishCode)
	}
	mandrill.TemplateDelete("publishTest")
}

func TestTemplateRender(t *testing.T) {
	//make sure it's freshly added
	mandrill.TemplateDelete("renderTest")
	mandrill.TemplateAdd("renderTest", "*|MC:SUBJECT|*", true)
	//weak - should check results
	mergeVars := []Var{*NewVar("SUBJECT", "Hello, welcome")}
	result, err := mandrill.TemplateRender("renderTest", nil, mergeVars)
	if err != nil {
		t.Error("Error:", err)
	}
	if result != "Hello, welcome" {
		t.Errorf("Rendered Result incorrect, expecting %s, got %v", "Hello, welcome", result)
	}
}

func TestTemplateRender2(t *testing.T) {
	//make sure it's freshly added
	mandrill.TemplateDelete("renderTest")
	mandrill.TemplateAdd("renderTest", "<div mc:edit=\"std_content00\"></div>", true)
	//weak - should check results
	templateContent := []Var{*NewVar("std_content00", "Hello, welcome")}
	result, err := mandrill.TemplateRender("renderTest", templateContent, nil)
	if err != nil {
		t.Error("Error:", err)
	}
	if result != "<div>Hello, welcome</div>" {
		t.Errorf("Rendered Result incorrect, expecting %s, got %s", "<div>Hello, welcome</div>", result)
	}
}

func TestMessageTemplateSend(t *testing.T) {
	//make sure it's freshly added
	mandrill.TemplateDelete(testTemplateName)
	mandrill.TemplateAdd(testTemplateName, readTemplate("templates/transactional_basic.html"), true)
	//weak - should check results
	templateContent := []Var{*NewVar("std_content00", "Hello, welcome")}
	mergeVars := []Var{*NewVar("SUBJECT", "Hello, welcome")}
	var message Message = Message{Subject: "Test Template Mail", FromEmail: user,
		FromName: user, GlobalMergeVars: mergeVars}
	message.AddRecipients(Recipient{Email: user, Name: user})
	_, err := mandrill.MessageSendTemplate(testTemplateName, templateContent, message, true)
	if err != nil {
		t.Error("Error:", err)
	}
	//todo - how do we test this better?
}

func readTemplate(path string) string {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}
	return string(b)
}

// senders tests

func TestSendersList(t *testing.T) {
	//make sure it's freshly added
	results, err := mandrill.SenderList()
	if err != nil {
		t.Error("Error:", err)
	}
	var foundUser = false
	for i := range results {
		var info Sender = results[i]
		fmt.Printf("sender:%v %s", info, info.Address)
		if info.Address == user {
			foundUser = true
		}
	}
	if !foundUser {
		t.Errorf("should have found User %s in [%v] length array", user, len(results))
	}
}

// incoming tests

func TestInboundDomainListAddCheckDelete(t *testing.T) {
	domainName := "improbable.example.com"
	domains, err := mandrill.InboundDomainList()
	if err != nil {
		t.Error("Error:", err)
	}
	originalCount := len(domains)
	domain, err := mandrill.InboundDomainAdd(domainName)
	if err != nil {
		t.Error("Error:", err)
	}
	domains, err = mandrill.InboundDomainList()
	if err != nil {
		t.Error("Error:", err)
	}
	newCount := len(domains)
	if newCount != originalCount+1 {
		t.Errorf("Expected %v domains, found %v after adding %v.", originalCount+1, newCount, domainName)
	}
	newDomain, err := mandrill.InboundDomainCheck(domainName)
	if err != nil {
		t.Error("Error:", err)
	}
	if domain.CreatedAt != newDomain.CreatedAt {
		t.Errorf("Domain check of %v and %v do not match.", domain.CreatedAt, newDomain.CreatedAt)
	}
	if domain.Domain != newDomain.Domain {
		t.Errorf("Domain check of %v and %v do not match.", domain.Domain, newDomain.Domain)
	}
	if domain.ValidMx != newDomain.ValidMx {
		t.Errorf("Domain check of %v and %v do not match.", domain.ValidMx, newDomain.ValidMx)
	}
	_, err = mandrill.InboundDomainDelete(domainName)
	if err != nil {
		t.Error("Error:", err)
	}
	domains, err = mandrill.InboundDomainList()
	if err != nil {
		t.Error("Error:", err)
	}
	deletedCount := len(domains)
	if deletedCount != originalCount {
		t.Errorf("Expected %v domains, found %v after deleting %v.", originalCount, deletedCount, domainName)
	}
}

func TestInboundDomainRoutesAndRaw(t *testing.T) {
	domainName := "www.google.com"
	emailAddress := "test"
	webhookUrl := fmt.Sprintf("http://%v/", domainName)
	_, err := mandrill.InboundDomainAdd(domainName)
	if err != nil {
		t.Error("Error:", err)
	}
	routeList, err := mandrill.RouteList(domainName)
	if err != nil {
		t.Error("Error:", err)
	}
	count := len(routeList)
	if count != 0 {
		t.Errorf("Expected no routes at %v, found %v.", domainName, count)
	}
	route, err := mandrill.RouteAdd(domainName, emailAddress, webhookUrl)
	if err != nil {
		t.Error("Error:", err)
	}
	if route.Pattern != emailAddress {
		t.Errorf("Expected pattern %v, found %v.", emailAddress, route.Pattern)
	}
	if route.Url != webhookUrl {
		t.Errorf("Expected URL %v, found %v.", webhookUrl, route.Url)
	}
	newDomainName := "www.google.com"
	newEmailAddress := "test2"
	newWebhookUrl := fmt.Sprintf("http://%v/", newDomainName)
	_, err = mandrill.InboundDomainCheck(newDomainName)
	if err != nil {
		t.Error("Error:", err)
	}
	route, err = mandrill.RouteUpdate(route.Id, newDomainName, newEmailAddress, newWebhookUrl)
	if err != nil {
		t.Error("Error:", err)
	}
	if route.Pattern != newEmailAddress {
		t.Errorf("Expected pattern %v, found %v.", newEmailAddress, route.Pattern)
	}
	if route.Url != newWebhookUrl {
		t.Errorf("Expected URL %v, found %v.", newWebhookUrl, route.Pattern)
	}
	route, err = mandrill.RouteDelete(route.Id)
	if err != nil {
		t.Error("Error:", err)
	}
	routeList, err = mandrill.RouteList(domainName)
	if err != nil {
		t.Error("Error:", err)
	}
	newCount := len(routeList)
	if newCount != count {
		t.Errorf("Expected %v routes at %v, found %v.", count, domainName, newCount)
	}
	rawMessage := "From: sender@example.com\nTo: test2@www.google.com\nSubject: Some Subject\n\nSome content."
	_, err = mandrill.SendRawMIME(rawMessage, []string{"test2@www.google.com"}, "test@www.google.com", "", "127.0.0.1")
	if err != nil {
		t.Error("Error:", err)
	}
	_, err = mandrill.InboundDomainDelete(domainName)
	if err != nil {
		t.Error("Error:", err)
	}
}

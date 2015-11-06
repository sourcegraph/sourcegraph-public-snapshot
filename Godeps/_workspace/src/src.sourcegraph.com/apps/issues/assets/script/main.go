// +build js

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/gopherjs/gopherjs/js"
	"github.com/shurcooL/github_flavored_markdown"
	"honnef.co/go/js/dom"
	"src.sourcegraph.com/apps/issues/common"
	"src.sourcegraph.com/apps/issues/issues"
)

var document = dom.GetWindow().Document().(dom.HTMLDocument)

var state common.State

func main() {
	js.Global.Set("MarkdownPreview", MarkdownPreview)
	js.Global.Set("SwitchWriteTab", SwitchWriteTab)
	js.Global.Set("CreateNewIssue", CreateNewIssue)
	js.Global.Set("ToggleIssueState", ToggleIssueState)
	js.Global.Set("PostComment", PostComment)

	stateJSON := js.Global.Get("State").String()
	fmt.Println(stateJSON)
	err := json.Unmarshal([]byte(stateJSON), &state)
	if err != nil {
		panic(err)
	}

	document.AddEventListener("DOMContentLoaded", false, func(_ dom.Event) {
		setup()
	})
}

func setup() {
	if issueToggleButton := document.GetElementByID("issue-toggle-button"); issueToggleButton != nil {
		commentEditor := document.GetElementByID("comment-editor").(*dom.HTMLTextAreaElement)
		commentEditor.AddEventListener("input", false, func(_ dom.Event) {
			if strings.TrimSpace(commentEditor.Value) == "" {
				issueToggleButton.SetTextContent(issueToggleButton.GetAttribute("data-1-action"))
			} else {
				issueToggleButton.SetTextContent(issueToggleButton.GetAttribute("data-2-actions"))
			}
		})
	}

	if createIssueButton, ok := document.GetElementByID("create-issue-button").(dom.HTMLElement); ok {
		titleEditor := document.GetElementByID("title-editor").(*dom.HTMLInputElement)
		titleEditor.AddEventListener("input", false, func(_ dom.Event) {
			if strings.TrimSpace(titleEditor.Value) == "" {
				createIssueButton.SetAttribute("disabled", "disabled")
			} else {
				createIssueButton.RemoveAttribute("disabled")
			}
		})
	}
}

func CreateNewIssue() {
	titleEditor := document.GetElementByID("title-editor").(*dom.HTMLInputElement)
	commentEditor := document.GetElementByID("comment-editor").(*dom.HTMLTextAreaElement)

	title := titleEditor.Value
	body := commentEditor.Value
	if strings.TrimSpace(title) == "" {
		log.Println("cannot create issue with empty title")
		return
	}

	go func() {
		resp, err := http.PostForm("new", url.Values{"csrf_token": {state.CSRFToken}, "title": {title}, "body": {body}})
		if err != nil {
			log.Println(err)
			return
		}
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Println(err)
			return
		}

		fmt.Printf("got reply: %v\n%q\n", resp.Status, string(body))

		switch resp.StatusCode {
		case http.StatusOK:
			// Redirect.
			dom.GetWindow().Location().Href = string(body)
		}
	}()
}

func ToggleIssueState(issueState issues.State) {
	go func() {
		// Post comment first if there's text entered, and we're closing.
		if strings.TrimSpace(document.GetElementByID("comment-editor").(*dom.HTMLTextAreaElement).Value) != "" &&
			issueState == issues.ClosedState {
			err := postComment()
			if err != nil {
				log.Println(err)
				return
			}
		}

		ir := issues.IssueRequest{
			State: &issueState,
		}
		value, err := json.Marshal(ir)
		if err != nil {
			panic(err)
		}

		resp, err := http.PostForm(state.BaseURI+state.ReqPath+"/edit", url.Values{"csrf_token": {state.CSRFToken}, "value": {string(value)}})
		if err != nil {
			log.Println(err)
			return
		}
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Println(err)
			return
		}

		data, err := url.ParseQuery(string(body))
		if err != nil {
			log.Println(err)
			return
		}
		fmt.Printf("got reply: %v\n%q\n", resp.Status, data)

		switch resp.StatusCode {
		case http.StatusOK:
			issueStateBadge := document.GetElementByID("issue-state-badge")
			issueStateBadge.SetInnerHTML(data.Get("issue-state-badge"))

			issueToggleButton := document.GetElementByID("issue-toggle-button")
			issueToggleButton.Underlying().Set("outerHTML", data.Get("issue-toggle-button"))

			// Create event.
			newEvent := document.CreateElement("div").(*dom.HTMLDivElement)
			newItemMarker := document.GetElementByID("new-item-marker")
			newItemMarker.ParentNode().InsertBefore(newEvent, newItemMarker)
			newEvent.Underlying().Set("outerHTML", data.Get("new-event"))
		}

		// Post comment after if there's text entered, and we're reopening.
		if strings.TrimSpace(document.GetElementByID("comment-editor").(*dom.HTMLTextAreaElement).Value) != "" &&
			issueState == issues.OpenState {
			err := postComment()
			if err != nil {
				log.Println(err)
				return
			}
		}
	}()
}

func PostComment() {
	go func() {
		err := postComment()
		if err != nil {
			log.Println(err)
		}
	}()
}

// postComment posts the comment.
func postComment() error {
	commentEditor := document.GetElementByID("comment-editor").(*dom.HTMLTextAreaElement)

	value := commentEditor.Value
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("cannot post empty comment")
	}

	resp, err := http.PostForm(state.BaseURI+state.ReqPath+"/comment", url.Values{"csrf_token": {state.CSRFToken}, "value": {value}})
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	fmt.Printf("got reply: %v\n%q\n", resp.Status, string(body))

	switch resp.StatusCode {
	case http.StatusOK:
		// Create comment.
		newComment := document.CreateElement("div").(*dom.HTMLDivElement)

		newItemMarker := document.GetElementByID("new-item-marker")
		newItemMarker.ParentNode().InsertBefore(newComment, newItemMarker)

		newComment.Underlying().Set("outerHTML", string(body))

		// Reset new-comment component.
		commentEditor.Value = ""
		commentEditor.Underlying().Call("dispatchEvent", js.Global.Get("CustomEvent").New("input")) // Trigger "input" event listeners.
		SwitchWriteTab()

		return nil
	default:
		return fmt.Errorf("did not get acceptable stauts code")
	}
}

var previewActive = false

func MarkdownPreview() {
	if previewActive {
		return
	}

	commentEditor := document.GetElementByID("comment-editor").(*dom.HTMLTextAreaElement)
	commentPreview := document.GetElementByID("comment-preview").(*dom.HTMLDivElement)

	in := commentEditor.Value
	if in == "" {
		in = "Nothing to preview."
	}
	commentPreview.SetInnerHTML(string(github_flavored_markdown.Markdown([]byte(in))))

	document.GetElementByID("write-tab-link").(dom.Element).Class().Remove("active")
	document.GetElementByID("preview-tab-link").(dom.Element).Class().Add("active")
	commentEditor.Style().SetProperty("display", "none", "")
	commentPreview.Style().SetProperty("display", "block", "")
	previewActive = true
}

func SwitchWriteTab() {
	commentEditor := document.GetElementByID("comment-editor").(*dom.HTMLTextAreaElement)
	commentPreview := document.GetElementByID("comment-preview").(*dom.HTMLDivElement)

	if previewActive {
		document.GetElementByID("write-tab-link").(dom.Element).Class().Add("active")
		document.GetElementByID("preview-tab-link").(dom.Element).Class().Remove("active")
		commentEditor.Style().SetProperty("display", "block", "")
		commentPreview.Style().SetProperty("display", "none", "")
		previewActive = false
	}

	commentEditor.Focus()
}

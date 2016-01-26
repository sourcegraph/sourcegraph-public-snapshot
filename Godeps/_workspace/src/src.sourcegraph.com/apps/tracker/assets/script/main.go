// +build js

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/gopherjs/gopherjs/js"
	"github.com/shurcooL/github_flavored_markdown"
	"github.com/shurcooL/go/gopherjs_http/jsutil"
	"github.com/shurcooL/markdownfmt/markdown"
	"honnef.co/go/js/dom"
	"src.sourcegraph.com/apps/tracker/common"
	"src.sourcegraph.com/apps/tracker/issues"
)

var document = dom.GetWindow().Document().(dom.HTMLDocument)

var state common.State

func main() {
	js.Global.Set("MarkdownPreview", jsutil.Wrap(MarkdownPreview))
	js.Global.Set("SwitchWriteTab", jsutil.Wrap(SwitchWriteTab))
	js.Global.Set("PasteHandler", jsutil.Wrap(PasteHandler))
	js.Global.Set("CreateNewIssue", CreateNewIssue)
	js.Global.Set("ToggleIssueState", ToggleIssueState)
	js.Global.Set("PostComment", PostComment)
	js.Global.Set("EditComment", jsutil.Wrap(EditComment))
	js.Global.Set("OpenReactionMenu", jsutil.Wrap(OpenReactionMenu))
	js.Global.Set("RenderReactionList", jsutil.Wrap(RenderReactionList))

	stateJSON := js.Global.Get("State").String()
	err := json.Unmarshal([]byte(stateJSON), &state)
	if err != nil {
		panic(err)
	}

	document.AddEventListener("DOMContentLoaded", false, func(_ dom.Event) {
		setup()
	})
}

func setup() {
	setupIssueToggleButton()
	setupReactions()

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

func setupIssueToggleButton() {
	if issueToggleButton := document.GetElementByID("issue-toggle-button"); issueToggleButton != nil {
		commentEditor := document.QuerySelector("#new-comment-container .comment-editor").(*dom.HTMLTextAreaElement)
		commentEditor.AddEventListener("input", false, func(_ dom.Event) {
			if strings.TrimSpace(commentEditor.Value) == "" {
				issueToggleButton.SetTextContent(issueToggleButton.GetAttribute("data-1-action"))
			} else {
				issueToggleButton.SetTextContent(issueToggleButton.GetAttribute("data-2-actions"))
			}
		})
	}
}

func postJSON(url string, body []byte) (*http.Response, error) {
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Csrf-Token", state.CSRFToken)
	return http.DefaultClient.Do(req)
}

func CreateNewIssue() {
	titleEditor := document.GetElementByID("title-editor").(*dom.HTMLInputElement)
	commentEditor := document.QuerySelector(".comment-editor").(*dom.HTMLTextAreaElement)

	title := strings.TrimSpace(titleEditor.Value)
	if title == "" {
		log.Println("cannot create issue with empty title")
		return
	}
	fmted, _ := markdown.Process("", []byte(commentEditor.Value), nil)
	body := string(bytes.TrimSpace(fmted))

	value, err := json.Marshal(issues.Issue{
		Title: title,
		Comment: issues.Comment{
			Body: body,
		},
	})
	if err != nil {
		panic(err)
	}

	go func() {
		resp, err := postJSON("new", value)
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
		if strings.TrimSpace(document.QuerySelector("#new-comment-container .comment-editor").(*dom.HTMLTextAreaElement).Value) != "" &&
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

		switch resp.StatusCode {
		case http.StatusOK:
			issueStateBadge := document.GetElementByID("issue-state-badge")
			issueStateBadge.SetInnerHTML(data.Get("issue-state-badge"))

			issueToggleButton := document.GetElementByID("issue-toggle-button")
			issueToggleButton.Underlying().Set("outerHTML", data.Get("issue-toggle-button"))
			setupIssueToggleButton()

			for _, newEventData := range data["new-event"] {
				// Create event.
				newEvent := document.CreateElement("div").(*dom.HTMLDivElement)
				newItemMarker := document.GetElementByID("new-item-marker")
				newItemMarker.ParentNode().InsertBefore(newEvent, newItemMarker)
				newEvent.Underlying().Set("outerHTML", newEventData)
			}
		}

		// Post comment after if there's text entered, and we're reopening.
		if strings.TrimSpace(document.QuerySelector("#new-comment-container .comment-editor").(*dom.HTMLTextAreaElement).Value) != "" &&
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

// postComment posts the comment to the remote API.
func postComment() error {
	commentEditor := document.QuerySelector("#new-comment-container .comment-editor").(*dom.HTMLTextAreaElement)

	fmted, _ := markdown.Process("", []byte(commentEditor.Value), nil)
	if len(fmted) == 0 {
		return fmt.Errorf("cannot post empty comment")
	}
	value := string(bytes.TrimSpace(fmted))

	resp, err := http.PostForm(state.BaseURI+state.ReqPath+"/comment", url.Values{"csrf_token": {state.CSRFToken}, "value": {value}})
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	fmt.Printf("got reply: %v\n", resp.Status)

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
		switchWriteTab(document.GetElementByID("new-comment-container"), commentEditor)

		return nil
	default:
		return fmt.Errorf("did not get acceptable status code: %v", resp.Status)
	}
}

// editComment edits a comment to the remote API.
// Comment must be validated before calling this.
func editComment(id string, content string) error {
	resp, err := http.PostForm(state.BaseURI+state.ReqPath+"/comment/"+id, url.Values{"csrf_token": {state.CSRFToken}, "value": {content}})
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	fmt.Printf("got reply: %v\n", resp.Status)

	switch resp.StatusCode {
	case http.StatusOK:
		return nil
	default:
		return fmt.Errorf("did not get acceptable status code: %v", resp.Status)
	}
}

func toggleReaction(commentID string, reaction issues.EmojiID) error {
	resp, err := http.PostForm(state.BaseURI+state.ReqPath+"/comment/"+commentID+"/reaction", url.Values{"csrf_token": {state.CSRFToken}, "reaction": {string(reaction)}})
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	fmt.Printf("got reply: %v\n", resp.Status)

	switch resp.StatusCode {
	case http.StatusOK:
		return nil
	default:
		return fmt.Errorf("did not get acceptable status code: %v", resp.Status)
	}
}

func MarkdownPreview(this dom.HTMLElement) {
	container := getAncestorByClassName(this, "edit-container")

	if container.QuerySelector(".preview-tab-link").(dom.Element).Class().Contains("active") {
		return
	}

	commentEditor := container.QuerySelector(".comment-editor").(*dom.HTMLTextAreaElement)
	commentPreview := container.QuerySelector(".comment-preview").(*dom.HTMLDivElement)

	fmted, _ := markdown.Process("", []byte(commentEditor.Value), nil)
	value := bytes.TrimSpace(fmted)

	if len(value) != 0 {
		commentPreview.SetInnerHTML(string(github_flavored_markdown.Markdown(value)))
	} else {
		commentPreview.SetInnerHTML(`<i class="gray">Nothing to preview.</i>`)
	}

	container.QuerySelector(".write-tab-link").(dom.Element).Class().Remove("active")
	container.QuerySelector(".preview-tab-link").(dom.Element).Class().Add("active")
	commentEditor.Style().SetProperty("display", "none", "")
	commentPreview.Style().SetProperty("display", "block", "")
}

func SwitchWriteTab(this dom.HTMLElement) {
	container := getAncestorByClassName(this, "edit-container")
	commentEditor := container.QuerySelector(".comment-editor").(*dom.HTMLTextAreaElement)
	switchWriteTab(container, commentEditor)
}

func switchWriteTab(container dom.Element, commentEditor *dom.HTMLTextAreaElement) {
	if container.QuerySelector(".preview-tab-link").(dom.Element).Class().Contains("active") {
		commentPreview := container.QuerySelector(".comment-preview").(*dom.HTMLDivElement)

		container.QuerySelector(".write-tab-link").(dom.Element).Class().Add("active")
		container.QuerySelector(".preview-tab-link").(dom.Element).Class().Remove("active")
		commentEditor.Style().SetProperty("display", "block", "")
		commentPreview.Style().SetProperty("display", "none", "")
	}

	commentEditor.Focus()
}

//
// Copyright 2021, Sander van Harmelen
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package gitlab

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// EventType represents a Gitlab event type.
type EventType string

// List of available event types.
const (
	EventConfidentialIssue EventType = "Confidential Issue Hook"
	EventConfidentialNote  EventType = "Confidential Note Hook"
	EventTypeBuild         EventType = "Build Hook"
	EventTypeDeployment    EventType = "Deployment Hook"
	EventTypeFeatureFlag   EventType = "Feature Flag Hook"
	EventTypeIssue         EventType = "Issue Hook"
	EventTypeJob           EventType = "Job Hook"
	EventTypeMember        EventType = "Member Hook"
	EventTypeMergeRequest  EventType = "Merge Request Hook"
	EventTypeNote          EventType = "Note Hook"
	EventTypePipeline      EventType = "Pipeline Hook"
	EventTypePush          EventType = "Push Hook"
	EventTypeRelease       EventType = "Release Hook"
	EventTypeServiceHook   EventType = "Service Hook"
	EventTypeSubGroup      EventType = "Subgroup Hook"
	EventTypeSystemHook    EventType = "System Hook"
	EventTypeTagPush       EventType = "Tag Push Hook"
	EventTypeWikiPage      EventType = "Wiki Page Hook"
)

const (
	eventObjectKindPush         = "push"
	eventObjectKindTagPush      = "tag_push"
	eventObjectKindMergeRequest = "merge_request"
)

const (
	noteableTypeCommit       = "Commit"
	noteableTypeIssue        = "Issue"
	noteableTypeMergeRequest = "MergeRequest"
	noteableTypeSnippet      = "Snippet"
)

type noteEvent struct {
	ObjectKind       string `json:"object_kind"`
	ObjectAttributes struct {
		NoteableType string `json:"noteable_type"`
	} `json:"object_attributes"`
}

type serviceEvent struct {
	ObjectKind string `json:"object_kind"`
}

const eventTypeHeader = "X-Gitlab-Event"

// HookEventType returns the event type for the given request.
func HookEventType(r *http.Request) EventType {
	return EventType(r.Header.Get(eventTypeHeader))
}

// ParseHook tries to parse both web- and system hooks.
//
// Example usage:
//
//	func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
//	    payload, err := ioutil.ReadAll(r.Body)
//	    if err != nil { ... }
//	    event, err := gitlab.ParseHook(gitlab.HookEventType(r), payload)
//	    if err != nil { ... }
//	    switch event := event.(type) {
//	    case *gitlab.PushEvent:
//	        processPushEvent(event)
//	    case *gitlab.MergeEvent:
//	        processMergeEvent(event)
//	    ...
//	    }
//	}
func ParseHook(eventType EventType, payload []byte) (event interface{}, err error) {
	switch eventType {
	case EventTypeSystemHook:
		return ParseSystemhook(payload)
	default:
		return ParseWebhook(eventType, payload)
	}
}

// ParseSystemhook parses the event payload. For recognized event types, a
// value of the corresponding struct type will be returned. An error will be
// returned for unrecognized event types.
//
// Example usage:
//
//	func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
//	    payload, err := ioutil.ReadAll(r.Body)
//	    if err != nil { ... }
//	    event, err := gitlab.ParseSystemhook(payload)
//	    if err != nil { ... }
//	    switch event := event.(type) {
//	    case *gitlab.PushSystemEvent:
//	        processPushSystemEvent(event)
//	    case *gitlab.MergeSystemEvent:
//	        processMergeSystemEvent(event)
//	    ...
//	    }
//	}
func ParseSystemhook(payload []byte) (event interface{}, err error) {
	e := &systemHookEvent{}
	err = json.Unmarshal(payload, e)
	if err != nil {
		return nil, err
	}

	switch e.EventName {
	case eventObjectKindPush:
		event = &PushSystemEvent{}
	case eventObjectKindTagPush:
		event = &TagPushSystemEvent{}
	case "repository_update":
		event = &RepositoryUpdateSystemEvent{}
	case
		"project_create",
		"project_update",
		"project_destroy",
		"project_transfer",
		"project_rename":
		event = &ProjectSystemEvent{}
	case
		"group_create",
		"group_destroy",
		"group_rename":
		event = &GroupSystemEvent{}
	case "key_create", "key_destroy":
		event = &KeySystemEvent{}
	case
		"user_create",
		"user_destroy",
		"user_rename",
		"user_failed_login":
		event = &UserSystemEvent{}
	case
		"user_add_to_group",
		"user_remove_from_group",
		"user_update_for_group":
		event = &UserGroupSystemEvent{}
	case
		"user_add_to_team",
		"user_remove_from_team",
		"user_update_for_team":
		event = &UserTeamSystemEvent{}
	default:
		switch e.ObjectKind {
		case string(MergeRequestEventTargetType):
			event = &MergeEvent{}
		default:
			return nil, fmt.Errorf("unexpected system hook type %s", e.EventName)
		}
	}

	if err := json.Unmarshal(payload, event); err != nil {
		return nil, err
	}

	return event, nil
}

// WebhookEventType returns the event type for the given request.
func WebhookEventType(r *http.Request) EventType {
	return EventType(r.Header.Get(eventTypeHeader))
}

// ParseWebhook parses the event payload. For recognized event types, a
// value of the corresponding struct type will be returned. An error will
// be returned for unrecognized event types.
//
// Example usage:
//
//	func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
//	    payload, err := ioutil.ReadAll(r.Body)
//	    if err != nil { ... }
//	    event, err := gitlab.ParseWebhook(gitlab.HookEventType(r), payload)
//	    if err != nil { ... }
//	    switch event := event.(type) {
//	    case *gitlab.PushEvent:
//	        processPushEvent(event)
//	    case *gitlab.MergeEvent:
//	        processMergeEvent(event)
//	    ...
//	    }
//	}
func ParseWebhook(eventType EventType, payload []byte) (event interface{}, err error) {
	switch eventType {
	case EventTypeBuild:
		event = &BuildEvent{}
	case EventTypeDeployment:
		event = &DeploymentEvent{}
	case EventTypeFeatureFlag:
		event = &FeatureFlagEvent{}
	case EventTypeIssue, EventConfidentialIssue:
		event = &IssueEvent{}
	case EventTypeJob:
		event = &JobEvent{}
	case EventTypeMember:
		event = &MemberEvent{}
	case EventTypeMergeRequest:
		event = &MergeEvent{}
	case EventTypeNote, EventConfidentialNote:
		note := &noteEvent{}
		err := json.Unmarshal(payload, note)
		if err != nil {
			return nil, err
		}

		if note.ObjectKind != string(NoteEventTargetType) {
			return nil, fmt.Errorf("unexpected object kind %s", note.ObjectKind)
		}

		switch note.ObjectAttributes.NoteableType {
		case noteableTypeCommit:
			event = &CommitCommentEvent{}
		case noteableTypeMergeRequest:
			event = &MergeCommentEvent{}
		case noteableTypeIssue:
			event = &IssueCommentEvent{}
		case noteableTypeSnippet:
			event = &SnippetCommentEvent{}
		default:
			return nil, fmt.Errorf("unexpected noteable type %s", note.ObjectAttributes.NoteableType)
		}
	case EventTypePipeline:
		event = &PipelineEvent{}
	case EventTypePush:
		event = &PushEvent{}
	case EventTypeRelease:
		event = &ReleaseEvent{}
	case EventTypeServiceHook:
		service := &serviceEvent{}
		err := json.Unmarshal(payload, service)
		if err != nil {
			return nil, err
		}
		switch service.ObjectKind {
		case eventObjectKindPush:
			event = &PushEvent{}
		case eventObjectKindTagPush:
			event = &TagEvent{}
		case eventObjectKindMergeRequest:
			event = &MergeEvent{}
		default:
			return nil, fmt.Errorf("unexpected service type %s", service.ObjectKind)
		}
	case EventTypeSubGroup:
		event = &SubGroupEvent{}
	case EventTypeTagPush:
		event = &TagEvent{}
	case EventTypeWikiPage:
		event = &WikiPageEvent{}
	default:
		return nil, fmt.Errorf("unexpected event type: %s", eventType)
	}

	if err := json.Unmarshal(payload, event); err != nil {
		return nil, err
	}

	return event, nil
}

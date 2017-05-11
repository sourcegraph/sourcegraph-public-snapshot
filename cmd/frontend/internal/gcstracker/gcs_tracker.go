package gcstracker

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"cloud.google.com/go/storage"

	"github.com/google/uuid"
	"github.com/sourcegraph/go-github/github"
	"golang.org/x/net/context"

	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/jscontext"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/actor"
	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
)

var gcstrackerVersion = "v0.0.1"

var googleCloudProjectID = "telligentsourcegraph"
var googleCloudBucketName = "telligent-sourcegraph-backend-data"

var prodEnv = "production"

// Client represents a connection to GCS for data tracking
// with environment, user, and session context
type Client struct {
	env, appID, sessionID string
	userInfo              *UserInfo
	gcsClient             *storage.Client
	ctx                   context.Context
}

// New returns a new GCS Tracker client using the given API key
// based on a provided actor.Actor
func New(user *actor.Actor, webSessionID string) (*Client, error) {
	return NewFromUserInfo(generateUserInfo(user), webSessionID)
}

// NewFromUserInfo returns a new GCS Tracker client using the given API key
func NewFromUserInfo(info *UserInfo, webSessionID string) (*Client, error) {
	ctx := context.Background()

	gcsClient, err := storage.NewClient(ctx)
	if err != nil {
		return nil, err
	}

	return &Client{
		appID:     jscontext.TrackingAppID,
		env:       prodEnv,
		sessionID: webSessionID,
		userInfo:  info,
		gcsClient: gcsClient,
		ctx:       ctx,
	}, nil
}

// Write object data to the GCS bucket
func (c *Client) Write(body *TrackedObjects) error {
	data, err := json.Marshal(*body)
	if err != nil {
		return err
	}

	obj := c.gcsClient.Bucket(googleCloudBucketName).Object(c.getCurrentGCSBucketPath())
	w := obj.NewWriter(c.ctx)
	defer w.Close()

	if _, err := io.Copy(w, bytes.NewReader(data)); err != nil {
		return err
	}

	return nil
}

func (c *Client) getCurrentGCSBucketPath() string {
	t := time.Now().UTC()
	return strings.Join([]string{c.env,
		"/",
		gcstrackerVersion,
		"/",
		fmt.Sprintf("%04d", t.Year()),
		"/",
		fmt.Sprintf("%02d", t.Month()),
		"/",
		fmt.Sprintf("%02d", t.Day()),
		"/",
		fmt.Sprintf("%02d", t.Hour()),
		"/",
		fmt.Sprintf("%d", t.Unix()),
		".txt"}, "")
}

// NewTrackedObjects generates a new TrackedObjects struct in the context of the
// GCS tracker client
func (c *Client) NewTrackedObjects(event string) *TrackedObjects {
	return &TrackedObjects{
		DeviceInfo: &DeviceInfo{
			Platform:         "Web",
			TrackerNamespace: "sg",
		},
		Objects: []*TrackedObject{},
		Header: &Header{
			AppID:        c.appID,
			Env:          c.env,
			SessionID:    c.sessionID,
			ServerTstamp: time.Now().UTC().Unix(),
			Event:        event,
		},
		BatchID:  uuid.New().String(),
		UserInfo: c.userInfo,
	}
}

func generateUserInfo(user *actor.Actor) *UserInfo {
	isPrivateCodeUser := false
	for _, v := range user.GitHubScopes {
		if v == "repo" {
			isPrivateCodeUser = true
			break
		}
	}
	return &UserInfo{
		BusinessUserID:    user.Login,
		Email:             user.Email,
		IsPrivateCodeUser: isPrivateCodeUser,
	}
}

// AddTrackedObject appends a new object to a TrackedObjects struct
func (tos *TrackedObjects) AddTrackedObject(objectType string, oc interface{}) {
	tos.Objects = append(tos.Objects, &TrackedObject{
		ObjectID: uuid.New().String(),
		Type:     objectType,
		Ctx:      oc,
	})
}

// AddReposWithDetailsObjects adds a series of RepoDetails objects to a TrackedObjects struct
// based on a sourcegraph.ReposWithDetailsList
func (tos *TrackedObjects) AddReposWithDetailsObjects(rl *sourcegraph.GitHubReposWithDetailsList) {
	for _, repo := range rl.ReposWithDetails {
		newRepo := &RepoWithDetailsContext{
			URI:                  repo.URI,
			Owner:                repo.Owner,
			Name:                 repo.Name,
			IsFork:               repo.Fork,
			IsPrivate:            repo.Private,
			Languages:            make([]*RepoLanguage, len(repo.Languages)),
			CommitTimes:          make([]int64, len(repo.CommitTimes)),
			ErrorFetchingDetails: repo.ErrorFetchingDetails,
			Skipped:              repo.Skipped,
		}
		if repo.CreatedAt != nil {
			newRepo.CreatedAt = repo.CreatedAt.UTC().Unix()
		}
		if repo.PushedAt != nil {
			newRepo.PushedAt = repo.PushedAt.UTC().Unix()
		}
		for i, lang := range repo.Languages {
			newRepo.Languages[i] = &RepoLanguage{
				Language: lang.Language,
				Count:    lang.Count,
			}
		}
		for i, commitTime := range repo.CommitTimes {
			if commitTime != nil {
				newRepo.CommitTimes[i] = commitTime.UTC().Unix()
			}
		}
		tos.AddTrackedObject("RepoDetails", newRepo)
	}
}

// AddOrgsWithDetailsObjects adds a series of OrgDetails objects to a TrackedObjects struct
// based on a map from org name => a sourcegraph.OrgMembersList
func (tos *TrackedObjects) AddOrgsWithDetailsObjects(ml map[string]([]*github.User)) {
	for orgName, orgMembers := range ml {
		newOrg := &OrgWithDetailsContext{
			OrgName: orgName,
			Members: make([]*OrgMember, len(orgMembers)),
		}
		for i, member := range orgMembers {
			newOrg.Members[i] = &OrgMember{
				MemberUserID: *member.Login,
			}
		}
		tos.AddTrackedObject("OrgDetails", newOrg)
	}
}

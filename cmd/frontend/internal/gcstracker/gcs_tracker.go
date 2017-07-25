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
	return newFromUserInfo(generateUserInfo(user), webSessionID)
}

// NewFromUserProperties returns a new GCS Tracker client using
// the given API key based on provided user properties
func NewFromUserProperties(login string, email string, webSessionID string) (*Client, error) {
	return newFromUserInfo(&UserInfo{
		BusinessUserID: login,
		Email:          email,
	}, webSessionID)
}

// newFromUserInfo returns a new GCS Tracker client using the given API key
func newFromUserInfo(info *UserInfo, webSessionID string) (*Client, error) {
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

// AddGitHubInstallationEvent adds a GitHub installation object to a TrackedObjects struct
func (tos *TrackedObjects) AddGitHubInstallationEvent(ghi *github.InstallationEvent) {
	var senderEmail string
	if ghi.Sender.Email != nil {
		senderEmail = *ghi.Sender.Email
	}

	install := GitHubInstallationEvent{
		Action: *ghi.Action,
		Sender: &GitHubAccount{
			Login:     *ghi.Sender.Login,
			ID:        fmt.Sprintf("github|%d", *ghi.Sender.ID),
			AvatarURL: *ghi.Sender.AvatarURL,
			Email:     senderEmail,
			Type:      *ghi.Sender.Type,
		},
		Installation: &GitHubInstallation{
			ID: *ghi.Installation.ID,
			Account: &GitHubAccount{
				Login:     *ghi.Installation.Account.Login,
				ID:        fmt.Sprintf("github|%d", *ghi.Installation.Account.ID),
				AvatarURL: *ghi.Installation.Account.AvatarURL,
				Type:      *ghi.Installation.Account.Type,
			},
		},
	}

	tos.AddTrackedObject("GitHubAppInstallationEvent", install)
}

// AddGitHubRepositoriesEvent adds a GitHub installation object to a TrackedObjects struct
func (tos *TrackedObjects) AddGitHubRepositoriesEvent(ghr *github.InstallationRepositoriesEvent) {
	var senderEmail string
	if ghr.Sender.Email != nil {
		senderEmail = *ghr.Sender.Email
	}

	event := GitHubRepositoriesEvent{
		Action: *ghr.Action,
		Sender: &GitHubAccount{
			Login:     *ghr.Sender.Login,
			ID:        fmt.Sprintf("github|%d", *ghr.Sender.ID),
			AvatarURL: *ghr.Sender.AvatarURL,
			Email:     senderEmail,
			Type:      *ghr.Sender.Type,
		},
		Installation: &GitHubInstallation{
			ID: *ghr.Installation.ID,
			Account: &GitHubAccount{
				Login:     *ghr.Installation.Account.Login,
				ID:        fmt.Sprintf("github|%d", *ghr.Installation.Account.ID),
				AvatarURL: *ghr.Installation.Account.AvatarURL,
				Type:      *ghr.Installation.Account.Type,
			},
		},
		RepositorySelection: *ghr.RepositorySelection,
	}

	tos.AddTrackedObject("GitHubAppRepositoriesEvent", event)

	for _, ghRepo := range ghr.RepositoriesAdded {
		repo := GitHubInstalledRepository{
			Action:   "added",
			ID:       *ghRepo.ID,
			Name:     *ghRepo.Name,
			FullName: *ghRepo.FullName,
		}
		tos.AddTrackedObject("GitHubAppRepository", repo)
	}
	for _, ghRepo := range ghr.RepositoriesRemoved {
		repo := GitHubInstalledRepository{
			Action:   "removed",
			ID:       *ghRepo.ID,
			Name:     *ghRepo.Name,
			FullName: *ghRepo.FullName,
		}
		tos.AddTrackedObject("GitHubAppRepository", repo)
	}
}

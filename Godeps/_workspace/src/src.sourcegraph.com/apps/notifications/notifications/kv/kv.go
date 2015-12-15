// Package kv implements notifications.Service using the Sourcegraph platform storage API.
package kv

import (
	"encoding/json"
	"fmt"
	"html/template"
	"os"
	"strconv"

	"golang.org/x/net/context"
	"src.sourcegraph.com/apps/notifications/notifications"
	"src.sourcegraph.com/apps/tracker/issues"
	approuter "src.sourcegraph.com/sourcegraph/app/router"
	"src.sourcegraph.com/sourcegraph/conf"
	"src.sourcegraph.com/sourcegraph/platform/putil"
	"src.sourcegraph.com/sourcegraph/platform/storage"
)

// NewService creates a Sourcegraph platform storage-backed notifications.Service,
// using appCtx context and platformStorageAppName as the app name identifier.
func NewService(appCtx context.Context, platformStorageAppName string) notifications.Service {
	return service{
		appCtx:  appCtx,
		appName: platformStorageAppName,
	}
}

type service struct {
	// appCtx is the app context with high priveldge. It's used to access the Sourcegraph platform storage
	// (on behalf of users that may not have write access). This service implementation is responsible for doing
	// authorization checks.
	appCtx context.Context

	// appName is the app name used for Sourcegraph platform storage.
	appName string
}

func (s service) List(ctx context.Context, opt interface{}) (notifications.Notifications, error) {
	currentUser := putil.UserFromContext(ctx)
	if currentUser == nil {
		return nil, os.ErrPermission
	}

	userKV := storage.Namespace(s.appCtx, s.appName, "")

	keys, err := userKV.List(formatUint64(uint64(currentUser.UID)))
	if err != nil {
		return nil, err
	}

	var ns notifications.Notifications

	for _, key := range keys {
		var n notification
		err := storage.GetJSON(userKV, formatUint64(uint64(currentUser.UID)), key, &n)
		if err != nil {
			return nil, fmt.Errorf("error reading %s/%s: %v", formatUint64(uint64(currentUser.UID)), key, err)
		}
		ns = append(ns, notifications.Notification{
			RepoSpec:  n.RepoSpec.RepoSpec(),
			RepoURL:   template.URL(conf.AppURL(s.appCtx).ResolveReference(approuter.Rel.URLToRepo(n.RepoSpec.URI)).String()),
			Title:     n.Title,
			HTMLURL:   n.HTMLURL,
			UpdatedAt: n.UpdatedAt,
			Icon:      n.Icon.OcticonID(),
		})
	}

	return ns, nil
}

func (s service) Count(ctx context.Context, opt interface{}) (uint64, error) {
	currentUser := putil.UserFromContext(ctx)
	if currentUser == nil {
		return 0, os.ErrPermission
	}

	userKV := storage.Namespace(s.appCtx, s.appName, "")

	notifications, err := userKV.List(formatUint64(uint64(currentUser.UID)))
	if err != nil {
		return 0, err
	}
	return uint64(len(notifications)), nil
}

func (s service) Notify(ctx context.Context, appID string, repo issues.RepoSpec, threadID uint64, op notifications.Notification) error {
	currentUser := putil.UserFromContext(ctx)

	userKV := storage.Namespace(s.appCtx, s.appName, "")
	repoKV := storage.Namespace(s.appCtx, s.appName, repo.URI)

	subscribers, err := repoKV.List(subscribersBucket(appID, threadID))
	if err != nil {
		return err
	}

	for _, subscriber := range subscribers {
		// TODO: Do this comparison better (int32-int32 instead of string-string), if possible.
		if currentUser != nil && subscriber == formatUint64(uint64(currentUser.UID)) {
			// TODO: Remove this.
			//fmt.Println("DEBUG: not skipping own user, notifying them anyway (for testing)!")

			// Don't notify user of his own actions.
			continue
		}

		n := notification{
			RepoSpec:  fromRepoSpec(repo),
			Title:     op.Title,
			HTMLURL:   op.HTMLURL,
			UpdatedAt: op.UpdatedAt,
			Icon:      fromOcticonID(op.Icon),
		}
		data, err := json.Marshal(n)
		if err != nil {
			return err
		}
		err = userKV.Put(subscriber, notificationKey(repo, appID, threadID), data)
		// TODO: Maybe in future read previous value, and use it to preserve some fields, like earliest HTML URL.
		//       Maybe that shouldn't happen here though.
		if err != nil {
			return fmt.Errorf("error writing %s/%s: %v", subscriber, notificationKey(repo, appID, threadID), err)
		}
	}

	return nil
}

func (s service) Subscribe(ctx context.Context, appID string, repo issues.RepoSpec, threadID uint64, subscribers []issues.UserSpec) error {
	currentUser := putil.UserFromContext(ctx)
	if currentUser == nil {
		return os.ErrPermission
	}

	repoKV := storage.Namespace(s.appCtx, s.appName, repo.URI)

	for _, subscriber := range subscribers {
		err := repoKV.Put(subscribersBucket(appID, threadID), formatUint64(subscriber.ID), nil)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s service) MarkRead(ctx context.Context, appID string, repo issues.RepoSpec, threadID uint64) error {
	currentUser := putil.UserFromContext(ctx)
	if currentUser == nil {
		return os.ErrPermission
	}

	userKV := storage.Namespace(s.appCtx, s.appName, "")

	// TODO: Move notification instead of outright removing, maybe?
	err := userKV.Delete(formatUint64(uint64(currentUser.UID)), notificationKey(repo, appID, threadID))
	if err != nil {
		return err
	}

	return nil
}

func formatUint64(n uint64) string { return strconv.FormatUint(n, 10) }

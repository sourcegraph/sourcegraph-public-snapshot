package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"runtime"
	"strings"
	"sync"
	"time"

	log15 "gopkg.in/inconshreveable/log15.v2"

	"github.com/pkg/errors"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/query-runner/queryrunnerapi"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
)

var allSavedQueries = &allSavedQueriesCached{}

// allSavedQueriesCached allows us to get a list of all the saved queries
// configured for every user/org on the entire server, without the overhead of
// constantly querying, unmarshaling, and transferring over the network all of
// the saved query setting values. Instead, we ask for the list once on startup
// and frontend instances notify us of created/updated/deleted saved queries in
// user/org configurations.
type allSavedQueriesCached struct {
	mu              sync.Mutex
	allSavedQueries map[string]api.SavedQuerySpecAndConfig
}

func savedQueryIDSpecKey(s api.SavedQueryIDSpec) string {
	return s.Subject.String() + s.Key
}

// get returns a copy of sq.allSavedQueries to avoid retaining the lock and
// blocking other oparations that call savedQueryWas[Created|Updated|Deleted]
// which also need the lock.
func (sq *allSavedQueriesCached) get() map[string]api.SavedQuerySpecAndConfig {
	sq.mu.Lock()
	defer sq.mu.Unlock()

	cpy := make(map[string]api.SavedQuerySpecAndConfig, len(sq.allSavedQueries))
	for k, v := range sq.allSavedQueries {
		cpy[k] = v
	}
	return cpy
}

// fetchInitialListFromFrontend blocks until the initial list can be initialized.
func (sq *allSavedQueriesCached) fetchInitialListFromFrontend() {
	sq.mu.Lock()
	defer sq.mu.Unlock()

	attempts := 0
	for {
		allSavedQueries, err := api.InternalClient.SavedQueriesListAll(context.Background())
		if err != nil {
			if attempts > 3 {
				// Only print the error if we've retried a few times, otherwise
				// we would be needlessly verbose when the frontend just hasn't
				// started yet but will soon.
				log15.Error("executor: error fetching saved queries list (trying again in 5s)", "error", err)
			}
			time.Sleep(5 * time.Second)
			attempts++
			continue
		}
		sq.allSavedQueries = make(map[string]api.SavedQuerySpecAndConfig, len(allSavedQueries))
		for spec, config := range allSavedQueries {
			sq.allSavedQueries[savedQueryIDSpecKey(spec)] = api.SavedQuerySpecAndConfig{
				Spec:   spec,
				Config: config,
			}
		}
		log15.Info("existing saved queries detected", "total_saved_queries", len(sq.allSavedQueries))
		return
	}
}

func serveSavedQueryWasCreatedOrUpdated(w http.ResponseWriter, r *http.Request) {
	allSavedQueries.mu.Lock()
	defer allSavedQueries.mu.Unlock()

	var args *queryrunnerapi.SavedQueryWasCreatedOrUpdatedArgs
	if err := json.NewDecoder(r.Body).Decode(&args); err != nil {
		writeError(w, errors.Wrap(err, "decoding JSON arguments"))
		return
	}

	for _, query := range args.SubjectAndConfig.Config.SavedQueries {
		spec := api.SavedQueryIDSpec{Subject: args.SubjectAndConfig.Subject, Key: query.Key}
		key := savedQueryIDSpecKey(spec)
		newValue := api.SavedQuerySpecAndConfig{
			Spec:   spec,
			Config: query,
		}

		// Handle notifying users of saved query creation / deletion.
		oldValue, exists := allSavedQueries.allSavedQueries[key]
		notifySavedQueryWasCreatedOrUpdated(oldValue, newValue, exists, args.DisableSubscriptionNotifications)

		allSavedQueries.allSavedQueries[key] = newValue
	}
	log15.Info("saved query created or updated", "total_saved_queries", len(allSavedQueries.allSavedQueries))
	w.WriteHeader(http.StatusOK)
}

func serveSavedQueryWasDeleted(w http.ResponseWriter, r *http.Request) {
	allSavedQueries.mu.Lock()
	defer allSavedQueries.mu.Unlock()

	var args *queryrunnerapi.SavedQueryWasDeletedArgs
	if err := json.NewDecoder(r.Body).Decode(&args); err != nil {
		writeError(w, errors.Wrap(err, "decoding JSON arguments"))
		return
	}

	key := savedQueryIDSpecKey(args.Spec)
	query, ok := allSavedQueries.allSavedQueries[key]
	if !ok {
		return // query to delete already doesn't exist; do nothing
	}
	qq := strings.Join([]string{query.Config.ScopeQuery, query.Config.Query}, " ")
	delete(allSavedQueries.allSavedQueries, key)

	if !args.DisableSubscriptionNotifications {
		// Inform any subscribers that they have been unsubscribed.
		go func() {
			if r := recover(); r != nil {
				// Same as net/http
				const size = 64 << 10
				buf := make([]byte, size)
				buf = buf[:runtime.Stack(buf, false)]
				log.Printf("executor: failed due to internal panic: %v\n%s", r, buf)
			}
			usersToNotify, orgsToNotify := getUsersAndOrgsToNotify(context.Background(), query.Spec, query.Config)
			emailNotifySubscribeUnsubscribe(context.Background(), usersToNotify, query, notifyUnsubscribedTemplate)
			slackNotifyDeleted(context.Background(), orgsToNotify, query)
		}()
	}

	// Delete from database, but only if another saved query is not the same.
	anotherExists := false
	for _, query := range allSavedQueries.allSavedQueries {
		queryStr := strings.Join([]string{query.Config.ScopeQuery, query.Config.Query}, " ")
		if queryStr == qq {
			anotherExists = true
			break
		}
	}
	if !anotherExists {
		if err := api.InternalClient.SavedQueriesDeleteInfo(r.Context(), qq); err != nil {
			log15.Error("Failed to delete saved query from DB: SavedQueriesDeleteInfo", "error", err)
			return
		}
	}
	log15.Info("saved query deleted", "total_saved_queries", len(allSavedQueries.allSavedQueries))
}

func notifySavedQueryWasCreatedOrUpdated(oldValue, newValue api.SavedQuerySpecAndConfig, exists, disableSubscriptionNotifications bool) {
	if disableSubscriptionNotifications {
		return
	}
	go func() {
		if r := recover(); r != nil {
			// Same as net/http
			const size = 64 << 10
			buf := make([]byte, size)
			buf = buf[:runtime.Stack(buf, false)]
			log.Printf("executor: failed due to internal panic: %v\n%s", r, buf)
		}

		if !exists {
			// Saved query (newValue) was created.
			usersToNotify, orgsToNotify := getUsersAndOrgsToNotify(context.Background(), newValue.Spec, newValue.Config)
			emailNotifySubscribeUnsubscribe(context.Background(), usersToNotify, newValue, notifySubscribedTemplate)
			slackNotifyCreated(context.Background(), orgsToNotify, newValue)
			return
		}

		// Users may have been added or removed from the configuration. Notify them accordingly.
		oldUsersToNotify, oldOrgsToNotify := int32MapDual(getUsersAndOrgsToNotify(context.Background(), oldValue.Spec, oldValue.Config))
		newUsersToNotify, newOrgsToNotify := int32MapDual(getUsersAndOrgsToNotify(context.Background(), newValue.Spec, newValue.Config))
		subscribed, unsubscribed := diffMap(oldUsersToNotify, newUsersToNotify)
		if len(subscribed) > 0 {
			emailNotifySubscribeUnsubscribe(context.Background(), subscribed, newValue, notifySubscribedTemplate)
		}
		if len(unsubscribed) > 0 {
			emailNotifySubscribeUnsubscribe(context.Background(), unsubscribed, oldValue, notifyUnsubscribedTemplate)
		}

		subscribedOrgs, unsubscribedOrgs := diffMap(oldOrgsToNotify, newOrgsToNotify)
		if len(subscribedOrgs) > 0 {
			slackNotifySubscribed(context.Background(), subscribedOrgs, newValue)
		}
		if len(unsubscribedOrgs) > 0 {
			slackNotifyUnsubscribed(context.Background(), unsubscribedOrgs, oldValue)
		}
	}()
}

func int32Map(v []int32) map[int32]struct{} {
	m := make(map[int32]struct{}, len(v))
	for _, v := range v {
		m[v] = struct{}{}
	}
	return m
}

func int32MapDual(a, b []int32) (map[int32]struct{}, map[int32]struct{}) {
	return int32Map(a), int32Map(b)
}

func diffMap(oldIDs, newIDs map[int32]struct{}) (added, removed []int32) {
	for id := range newIDs {
		_, didExist := oldIDs[id]
		if didExist {
			continue
		}
		added = append(added, id)
	}
	for id := range oldIDs {
		_, stillExists := newIDs[id]
		if stillExists {
			continue
		}
		removed = append(removed, id)
	}
	return
}

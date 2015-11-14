package listeners

import (
	"encoding/base64"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"

	"golang.org/x/net/context"
	"gopkg.in/inconshreveable/log15.v2"

	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/events"
	"src.sourcegraph.com/sourcegraph/notif"
)

func init() {
	events.RegisterListener(&clientListener{})
}

type clientListener struct{}

func (g *clientListener) Scopes() []string {
	return []string{"app:clients"}
}

func (g *clientListener) Start(ctx context.Context) {
	notifyCallback := func(id events.EventID, p events.ClientPayload) {
		notifyClientEvent(ctx, id, p)
	}

	events.Subscribe(events.ClientRegisterEvent, notifyCallback)
	events.Subscribe(events.ClientUpdateEvent, notifyCallback)
	events.Subscribe(events.ClientGrantAccessEvent, notifyCallback)
}

func notifyClientEvent(ctx context.Context, id events.EventID, payload events.ClientPayload) {
	cl := sourcegraph.NewClientFromContext(ctx)

	if payload.ClientID == "" {
		log15.Warn("ClientHook: ignoring event", "event", id, "error", "client id not set in payload")
		return
	}

	if payload.Actor.UID == 0 {
		log15.Warn("ClientHook: ignoring event", "event", id, "error", "uid not set in payload")
		return
	}

	client, err := cl.RegisteredClients.Get(ctx, &sourcegraph.RegisteredClientSpec{ID: payload.ClientID})
	if err != nil {
		log15.Warn("ClientHook: could not fetch client info", "event", id, "payload", payload, "error", err)
		return
	}

	userLogin := getUserLogin(cl, ctx, &payload.Actor)
	var actionStr string

	switch id {
	case events.ClientRegisterEvent:
		actionStr = "registered a new Sourcegraph"
	case events.ClientUpdateEvent:
		actionStr = "updated their Sourcegraph"
	case events.ClientGrantAccessEvent:
		actionStr = "granted access to their Sourcegraph"
	default:
		log15.Warn("ClientHook: ignoring unknown event", "event", id)
		return
	}

	appURL := strings.TrimSuffix(client.RedirectURIs[0], "/login/oauth/receive")

	msg := fmt.Sprintf("*%s* (UID %v) %s: *%s* (%s)",
		userLogin,
		payload.Actor.UID,
		actionStr,
		client.ClientName,
		appURL,
	)
	escapedClientID := url.QueryEscape(client.ID)

	if id == events.ClientGrantAccessEvent {
		granteeLogin := getUserLogin(cl, ctx, &payload.Grantee)
		permsString := getPermsStr(payload.Perms)
		msg += fmt.Sprintf("\nGrantee: *%s* (UID %v) [%s]", granteeLogin, payload.Grantee.UID, permsString)
		if userUrl := getDecodedEnvVar("SG_KIBANA_USER_URL_BASE64"); userUrl != "" {
			userUrl = strings.Replace(userUrl, "{ClientID}", escapedClientID, 1)

			escapedUID := url.QueryEscape(strconv.Itoa(int(payload.Grantee.UID)))
			userUrl = strings.Replace(userUrl, "{UID}", escapedUID, 1)

			msg += fmt.Sprintf(" (<%s|View user activity>)", userUrl)
		}
	}

	if clientUrl := getDecodedEnvVar("SG_KIBANA_CLIENT_URL_BASE64"); clientUrl != "" {
		clientUrl = strings.Replace(clientUrl, "{ClientID}", escapedClientID, 1)
		msg += fmt.Sprintf("\n<%s|View client activity>", clientUrl)
	}

	notif.ActionSlackMessage(notif.ActionContext{SlackMsg: msg})
}

func getUserLogin(cl *sourcegraph.Client, ctx context.Context, userSpec *sourcegraph.UserSpec) string {
	userLogin := "anonymous"
	if userSpec.UID != 0 {
		user, err := cl.Users.Get(ctx, userSpec)
		if err != nil {
			log15.Warn("ClientHook: could not fetch user info", "uid", userSpec.UID, "error", err)
		} else {
			userLogin = user.Login
		}
	}
	return userLogin
}

func getPermsStr(perms *sourcegraph.UserPermissions) string {
	permsArr := []string{"-", "-", "-"}
	if perms.Read {
		permsArr[0] = "r"
	}
	if perms.Write {
		permsArr[1] = "w"
	}
	if perms.Admin {
		permsArr[2] = "a"
	}
	return strings.Join(permsArr, "")
}

func getDecodedEnvVar(varName string) string {
	if envVar := os.Getenv(varName); envVar != "" {
		decoded, err := base64.StdEncoding.DecodeString(envVar)
		if err != nil {
			log15.Warn("ClientHook: error decoding env var", "var", varName, "error", err)
			return ""
		}
		return string(decoded)
	}
	return ""
}

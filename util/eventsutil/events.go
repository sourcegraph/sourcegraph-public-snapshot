package eventsutil

import (
	"fmt"
	"net/url"
	"runtime"
	"strconv"
	"strings"

	"golang.org/x/net/context"

	"src.sourcegraph.com/sourcegraph/auth"
	"src.sourcegraph.com/sourcegraph/auth/idkey"
	"src.sourcegraph.com/sourcegraph/conf"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
)

// LogStartServer records a server startup event.
func LogStartServer(clientID string) {
	Log(&sourcegraph.Event{
		Type:     "StartServer",
		DeviceID: clientID,
		ClientID: clientID,
		EventProperties: map[string]string{
			"OS-Arch":  fmt.Sprintf("%s-%s", runtime.GOOS, runtime.GOARCH),
			"ClientID": clientID,
		},
	})
}

// LogRegisterServer records that this client registered with the mothership.
func LogRegisterServer(clientID, clientName string) {
	Log(&sourcegraph.Event{
		Type:     "RegisterServer",
		DeviceID: clientID,
		ClientID: clientID,
	})
}

// LogCreateAccount records that an account got created, possibly with
// an invite code.
func LogCreateAccount(ctx context.Context, newAcct *sourcegraph.NewAccount, admin, write, firstUser bool, inviteCode string) {
	clientID := idkey.FromContext(ctx).ID
	userID, deviceID := getUserOrDeviceID(clientID, newAcct.Login)

	userProperties := map[string]string{
		"UID":         strconv.Itoa(int(newAcct.UID)),
		"UserID":      userID,
		"Email":       newAcct.Email,
		"ClientID":    clientID,
		"AccessLevel": getAccessLevel(admin, write),
	}

	if strings.Contains(newAcct.Email, "@") {
		userProperties["Domain"] = strings.SplitN(newAcct.Email, "@", 2)[1]
	}

	appURL := conf.AppURL(ctx)
	if appURL != nil {
		userProperties["AppURL"] = appURL.String()
	}

	firstUserStr := "False"
	if firstUser {
		firstUserStr = "True"
	}
	eventProperties := map[string]string{
		"FirstUser":  firstUserStr,
		"InviteCode": inviteCode,
	}

	Log(&sourcegraph.Event{
		Type:            "CreateAccount",
		ClientID:        clientID,
		UserID:          userID,
		DeviceID:        deviceID,
		UserProperties:  userProperties,
		EventProperties: eventProperties,
	})
}

// LogSendInvite records that an invite link was created.
func LogSendInvite(ctx context.Context, user *sourcegraph.User, email, inviteCode string, admin, write bool) {
	clientID := idkey.FromContext(ctx).ID
	userID, deviceID := getUserOrDeviceID(clientID, user.Login)

	eventProperties := map[string]string{
		"Invitee":     email,
		"InviteCode":  inviteCode,
		"AccessLevel": getAccessLevel(admin, write),
	}

	Log(&sourcegraph.Event{
		Type:            "SendInvite",
		ClientID:        clientID,
		UserID:          userID,
		DeviceID:        deviceID,
		EventProperties: eventProperties,
	})
}

func LogAddRepo(ctx context.Context, cloneURL, language string, mirror, private bool) {
	clientID := idkey.FromContext(ctx).ID
	userID, deviceID := getUserOrDeviceID(clientID, auth.ActorFromContext(ctx).Login)

	source := "local"
	if mirror {
		if u, err := url.Parse(cloneURL); err != nil {
			source = "unknown"
		} else {
			source = u.Host
		}
	}

	visibility := "public"
	if private {
		visibility = "private"
	}

	Log(&sourcegraph.Event{
		Type:     "AddRepo",
		ClientID: clientID,
		UserID:   userID,
		DeviceID: deviceID,
		EventProperties: map[string]string{
			"Source":     source,
			"Visibility": visibility,
			"Language":   language,
		},
	})
}

func LogBuildRepo(ctx context.Context, result string) {
	clientID := idkey.FromContext(ctx).ID
	userID, deviceID := getUserOrDeviceID(clientID, auth.ActorFromContext(ctx).Login)

	Log(&sourcegraph.Event{
		Type:     "BuildRepo",
		ClientID: clientID,
		UserID:   userID,
		DeviceID: deviceID,
		EventProperties: map[string]string{
			"CodeIntelligence": result,
		},
	})
}

func LogHTTPGitPush(ctx context.Context) {
	clientID := idkey.FromContext(ctx).ID
	userID, deviceID := getUserOrDeviceID(clientID, auth.ActorFromContext(ctx).Login)

	Log(&sourcegraph.Event{
		Type:     "GitPush",
		ClientID: clientID,
		UserID:   userID,
		DeviceID: deviceID,
		EventProperties: map[string]string{
			"Protocol": "HTTP",
		},
	})
}

func LogSSHGitPush(clientID, login string) {
	userID, deviceID := getUserOrDeviceID(clientID, login)

	Log(&sourcegraph.Event{
		Type:     "GitPush",
		ClientID: clientID,
		UserID:   userID,
		DeviceID: deviceID,
		EventProperties: map[string]string{
			"Protocol": "SSH",
		},
	})
}

func LogSearchQuery(ctx context.Context, searchType string, numResults int32) {
	clientID := idkey.FromContext(ctx).ID
	userID, deviceID := getUserOrDeviceID(clientID, auth.ActorFromContext(ctx).Login)

	Log(&sourcegraph.Event{
		Type:     searchType,
		ClientID: clientID,
		UserID:   userID,
		DeviceID: deviceID,
		EventProperties: map[string]string{
			"NumResults": strconv.Itoa(int(numResults)),
		},
	})
}

// func LogGoToDefinition(ctx context.Context) {
// 	clientID := idkey.FromContext(ctx).ID
// 	log.Printf("clientID: %s", clientID)
// 	user := handlerutil.UserFromContext(ctx)
// 	userID, deviceID := getUserOrDeviceID(clientID, getUserLogin(user))

// 	Log(&sourcegraph.Event{
// 		Type:     "GoToDefinition",
// 		ClientID: clientID,
// 		UserID:   userID,
// 		DeviceID: deviceID,
// 	})
// }

// func LogViewDef(ctx context.Context) {
// 	clientID := idkey.FromContext(ctx).ID
// 	user := handlerutil.UserFromContext(ctx)
// 	userID, deviceID := getUserOrDeviceID(clientID, getUserLogin(user))

// 	Log(&sourcegraph.Event{
// 		Type:     "ViewDef",
// 		ClientID: clientID,
// 		UserID:   userID,
// 		DeviceID: deviceID,
// 	})
// }

func LogCreateChangeset(ctx context.Context) {
	clientID := idkey.FromContext(ctx).ID
	userID, deviceID := getUserOrDeviceID(clientID, auth.ActorFromContext(ctx).Login)

	Log(&sourcegraph.Event{
		Type:     "CreateChangeset",
		ClientID: clientID,
		UserID:   userID,
		DeviceID: deviceID,
	})
}

func LogPageView(ctx context.Context, user *sourcegraph.UserSpec, route string) {
	clientID := idkey.FromContext(ctx).ID
	userID, deviceID := getUserOrDeviceID(clientID, getUserLogin(user))

	eventProperties := map[string]string{
		"Route": route,
	}

	Log(&sourcegraph.Event{
		Type:            "PageView",
		ClientID:        clientID,
		UserID:          userID,
		DeviceID:        deviceID,
		EventProperties: eventProperties,
	})
}

func getAccessLevel(admin, write bool) string {
	if admin {
		return "Admin"
	} else if write {
		return "Write"
	}
	return "Read"
}

func getShortClientID(clientID string) string {
	shortLen := 6
	if len(clientID) < shortLen {
		shortLen = len(clientID)
	}
	return clientID[:shortLen]
}

func getUserOrDeviceID(clientID, login string) (string, string) {
	if login == "" {
		return "", clientID
	}
	shortClientID := getShortClientID(clientID)
	return fmt.Sprintf("%s@%s", login, shortClientID), ""
}

func getUserLogin(user *sourcegraph.UserSpec) string {
	if user != nil {
		return user.Login
	}
	return ""
}

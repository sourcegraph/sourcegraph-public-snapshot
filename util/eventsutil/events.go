package eventsutil

import (
	"fmt"
	"net/http"
	"net/url"
	"runtime"
	"strconv"
	"strings"
	"unicode"

	"github.com/sourcegraph/go-github/github"
	"github.com/sourcegraph/mux"
	"golang.org/x/net/context"
	"gopkg.in/inconshreveable/log15.v2"

	"sourcegraph.com/sourcegraph/sourcegraph/auth"
	"sourcegraph.com/sourcegraph/sourcegraph/conf"
	"sourcegraph.com/sourcegraph/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/util/handlerutil"
	"sourcegraph.com/sourcegraph/sourcegraph/util/httputil/httpctx"
)

// LogStartServer records a server startup event.
func LogStartServer() {
	clientID := sourcegraphClientID
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
func LogRegisterServer(clientName string) {
	clientID := sourcegraphClientID

	Log(&sourcegraph.Event{
		Type:     "RegisterServer",
		DeviceID: clientID,
		ClientID: clientID,
	})
}

// LogCreateAccount records that an account got created, possibly with
// an invite code.
func LogCreateAccount(ctx context.Context, newAcct *sourcegraph.NewAccount, admin, write, firstUser bool, inviteCode string) {
	clientID := sourcegraphClientID
	userID, deviceID := getUserAndDeviceID(ctx, newAcct.Login)

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

	// Truncate the invite code to avoid any possibility of leaking it, but
	// preserve the ability to link the sign up event with the invite event
	if len(inviteCode) > 5 {
		inviteCode = inviteCode[:5]
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
func LogSendInvite(ctx context.Context, email, inviteCode string, admin, write bool) {
	clientID := sourcegraphClientID
	userID, deviceID := getUserAndDeviceID(ctx, auth.ActorFromContext(ctx).Login)

	// Truncate the invite code to avoid any possibility of leaking it, but
	// preserve the ability to link the sign up event with the invite event
	if len(inviteCode) > 5 {
		inviteCode = inviteCode[:5]
	}

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
	clientID := sourcegraphClientID
	userID, deviceID := getUserAndDeviceID(ctx, auth.ActorFromContext(ctx).Login)

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

	eventProperties := map[string]string{
		"Source":     source,
		"Visibility": visibility,
		"Language":   language,
	}

	organization := returnOrganization(cloneURL)
	if organization != "" {
		eventProperties["Org"] = organization
	}

	Log(&sourcegraph.Event{
		Type:            "AddRepo",
		ClientID:        clientID,
		UserID:          userID,
		DeviceID:        deviceID,
		EventProperties: eventProperties,
	})
}

func LogBuildRepo(ctx context.Context, result string, build *sourcegraph.Build) {
	if result == "" {
		return
	}

	repoRevSpec := &sourcegraph.RepoRevSpec{
		RepoSpec: sourcegraph.RepoSpec{build.Repo},
		CommitID: build.CommitID,
	}
	cl, err := sourcegraph.NewClientFromContext(ctx)
	if err != nil {
		log15.Debug(err.Error())
		return
	}

	inventory, err := cl.Repos.GetInventory(ctx, repoRevSpec)
	if err != nil {
		log15.Debug(err.Error())
		return
	}

	var languages []string
	for _, v := range inventory.Languages {
		languages = append(languages, v.Name)
	}
	langs := strings.Join(languages, ",")

	clientID := sourcegraphClientID
	userID, deviceID := getUserAndDeviceID(ctx, auth.ActorFromContext(ctx).Login)

	eventProperties := map[string]string{
		"CodeIntelligence": result,
		"ProgramLanguages": langs,
	}

	organization := returnOrganization(build.Repo)
	if organization != "" {
		eventProperties["Org"] = organization
	}

	Log(&sourcegraph.Event{
		Type:            "BuildRepo",
		ClientID:        clientID,
		UserID:          userID,
		DeviceID:        deviceID,
		EventProperties: eventProperties,
	})
}

func LogFinishBuildTask(ctx context.Context, label string, success bool, failure bool) {
	var eventType string
	if strings.Contains(strings.ToLower(label), "(indexing)") {
		// Log srclib code intelligence task result.
		eventType = "FinishSrclibBuild"
	} else if strings.ToLower(label) == "build" {
		// Log CI (continuous integration) build task result.
		eventType = "FinishCIBuild"
	} else {
		// Don't log other task types.
		return
	}

	clientID := sourcegraphClientID
	userID, deviceID := getUserAndDeviceID(ctx, auth.ActorFromContext(ctx).Login)

	result := "N/A"
	if success {
		result = "success"
	} else if failure {
		result = "failed"
	}

	Log(&sourcegraph.Event{
		Type:     eventType,
		ClientID: clientID,
		UserID:   userID,
		DeviceID: deviceID,
		EventProperties: map[string]string{
			"Label":  label,
			"Result": result,
		},
	})

}

func LogBrowseCode(ctx context.Context, entryType string, tc *handlerutil.TreeEntryCommon, rc *handlerutil.RepoCommon) {
	clientID := sourcegraphClientID
	user := handlerutil.UserFromContext(ctx)
	userID, deviceID := getUserAndDeviceID(ctx, getUserLogin(user))
	userAgent := UserAgentFromContext(ctx)

	codeIntelligenceAvailable := "false"
	if tc != nil && tc.SrclibDataVersion != nil {
		codeIntelligenceAvailable = "true"
	}

	source := "local"
	if rc != nil && rc.Repo != nil && rc.Repo.Mirror {
		if u, err := url.Parse(rc.Repo.HTTPCloneURL); err != nil {
			source = "unknown"
		} else {
			source = u.Host
		}
	}

	eventProperties := map[string]string{
		"EntryType":        entryType,
		"CodeIntelligence": codeIntelligenceAvailable,
		"Source":           source,
	}

	if userAgent != "" {
		eventProperties["UserAgent"] = userAgent
	}

	Log(&sourcegraph.Event{
		Type:            "ViewRepoTree",
		ClientID:        clientID,
		UserID:          userID,
		DeviceID:        deviceID,
		EventProperties: eventProperties,
	})
}

func LogGitPush(ctx context.Context) {
	clientID := sourcegraphClientID
	userID, deviceID := getUserAndDeviceID(ctx, auth.ActorFromContext(ctx).Login)

	Log(&sourcegraph.Event{
		Type:     "GitPush",
		ClientID: clientID,
		UserID:   userID,
		DeviceID: deviceID,
	})
}

func LogSearchQuery(ctx context.Context, searchType string, numResults int32) {
	clientID := sourcegraphClientID
	userID, deviceID := getUserAndDeviceID(ctx, auth.ActorFromContext(ctx).Login)

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

func LogViewDef(ctx context.Context, eventType string) {
	clientID := sourcegraphClientID
	user := handlerutil.UserFromContext(ctx)
	userID, deviceID := getUserAndDeviceID(ctx, getUserLogin(user))

	Log(&sourcegraph.Event{
		Type:     eventType,
		ClientID: clientID,
		UserID:   userID,
		DeviceID: deviceID,
	})
}

func LogPageView(ctx context.Context, user *sourcegraph.UserSpec, req *http.Request) {
	route := httpctx.RouteName(req)
	eventType := getPageViewEventType(route)
	if eventType == "" {
		return
	}

	clientID := sourcegraphClientID
	userID, deviceID := getUserAndDeviceID(ctx, getUserLogin(user))
	repoSpec, err := sourcegraph.UnmarshalRepoSpec(mux.Vars(req))
	var organization string
	if err != nil {
		organization = ""
	} else {
		organization = returnOrganization(repoSpec.URI)
	}
	userAgent := UserAgentFromContext(ctx)

	var eventProperties map[string]string
	if organization != "" {
		eventProperties = make(map[string]string)
		eventProperties["Org"] = organization
	}

	if userAgent != "" {
		if organization == "" {
			eventProperties = make(map[string]string)
		}
		eventProperties["UserAgent"] = userAgent
	}

	Log(&sourcegraph.Event{
		Type:            eventType,
		ClientID:        clientID,
		UserID:          userID,
		DeviceID:        deviceID,
		EventProperties: eventProperties,
	})
}

func LogSignIn(ctx context.Context) {
	LogEvent(ctx, "UserSignIn")
}

func LogSignOut(ctx context.Context) {
	LogEvent(ctx, "UserSignOut")
}

func LogEvent(ctx context.Context, event string) {
	login := auth.ActorFromContext(ctx).Login
	clientID := sourcegraphClientID
	userID, deviceID := getUserAndDeviceID(ctx, login)
	userAgent := UserAgentFromContext(ctx)

	var eventProperties map[string]string
	if userAgent != "" {
		eventProperties = make(map[string]string)
		eventProperties["UserAgent"] = userAgent
	}

	Log(&sourcegraph.Event{
		Type:            event,
		ClientID:        clientID,
		UserID:          userID,
		DeviceID:        deviceID,
		EventProperties: eventProperties,
	})
}

func LogLinkGitHub(ctx context.Context, ghUser *github.User) {
	login := auth.ActorFromContext(ctx).Login
	clientID := sourcegraphClientID
	userID, deviceID := getUserAndDeviceID(ctx, login)
	userAgent := UserAgentFromContext(ctx)

	var ghlogin, name, company, location, email string
	if ghUser.Name != nil {
		name = *ghUser.Name
	}
	if ghUser.Login != nil {
		ghlogin = *ghUser.Login
	}
	if ghUser.Location != nil {
		location = *ghUser.Location
	}
	if ghUser.Company != nil {
		company = *ghUser.Company
	}
	if ghUser.Email != nil {
		email = *ghUser.Email
	}
	userProperties := map[string]string{
		"GHLogin":    ghlogin,
		"GHName":     name,
		"GHLocation": location,
		"GHCompany":  company,
		"GHEmail":    email,
	}

	var eventProperties map[string]string
	if userAgent != "" {
		eventProperties = make(map[string]string)
		eventProperties["UserAgent"] = userAgent
	}

	Log(&sourcegraph.Event{
		Type:            "LinkGitHub",
		ClientID:        clientID,
		UserID:          userID,
		DeviceID:        deviceID,
		EventProperties: eventProperties,
		UserProperties:  userProperties,
	})
}

func LogAddTeammates(ctx context.Context, numSuccess, numFail int32) {
	clientID := sourcegraphClientID
	userID, deviceID := getUserAndDeviceID(ctx, auth.ActorFromContext(ctx).Login)

	eventProperties := map[string]string{
		"NumSuccess": fmt.Sprintf("%d", numSuccess),
		"NumFail":    fmt.Sprintf("%d", numFail),
	}

	Log(&sourcegraph.Event{
		Type:            "AddTeammates",
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

func getUserAndDeviceID(ctx context.Context, login string) (string, string) {
	deviceId := DeviceIdFromContext(ctx)
	if deviceId == "" {
		deviceId = sourcegraphClientID
	}

	if login == "" {
		return "", deviceId
	}
	shortClientID := getShortClientID(sourcegraphClientID)
	return fmt.Sprintf("%s@%s", login, shortClientID), deviceId
}

func getUserLogin(user *sourcegraph.UserSpec) string {
	if user != nil {
		return user.Login
	}
	return ""
}

func getPageViewEventType(route string) string {
	if route == "" {
		return ""
	}

	// Filter out routes that have their own top-level event
	// to avoid double logging the same user event.
	switch route {
	case "repo.tree":
		return ""
	}

	eventType := "View"
	chunks := strings.Split(route, ".")
	for i := range chunks {
		if len(chunks[i]) > 0 {
			token := []rune(chunks[i])
			token[0] = unicode.ToUpper(token[0])
			eventType = eventType + string(token)
		}
	}

	return eventType
}

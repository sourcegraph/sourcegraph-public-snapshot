package app

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"src.sourcegraph.com/sourcegraph/ext/slack"
	"src.sourcegraph.com/sourcegraph/notif"
)

func serveBetaSignup(w http.ResponseWriter, r *http.Request) error {
	if err := r.ParseForm(); err != nil {
		return err
	}
	txt, err := extractFormData(r)
	if err != nil {
		return err
	}
	if err := notif.SendAdminEmail("Beta signup", txt); err != nil {
		txt += "\n*WARNING*: Failed to send-email"
	}
	slack.PostMessage(slack.PostOpts{
		Msg:       txt,
		Username:  "beta signup",
		IconEmoji: ":+1:",
		Channel:   "#notif-bot",
	})
	http.Redirect(w, r, "/signup-complete", http.StatusSeeOther)
	return nil
}

func extractFormData(r *http.Request) (string, error) {
	langs := strings.Join(r.PostForm["Lang"], ", ")
	if r.FormValue("OtherLangs") != "" {
		langs += ", " + r.FormValue("Otherlangs")
	}

	buf := fmt.Sprintf(`%s signed-up on _%s_.
Name: %s %s
Phone Number: %s
Company: %s
Title: %s
Team Size: %s

Languages: %s`, r.FormValue("Email"), time.Now().Format("January 2, 15:04 MST"), r.FormValue("FirstName"), r.FormValue("LastName"), r.FormValue("Phone"), r.FormValue("Company"), r.FormValue("Title"), r.FormValue("TeamSize"), langs)
	return buf, nil
}

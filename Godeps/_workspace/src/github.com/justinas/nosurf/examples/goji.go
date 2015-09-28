// Demonstrates how to tie together Goji (https://goji.io) and nosurf.
package main

import (
	"fmt"
	"html/template"
	"net/http"

	"github.com/zenazn/goji"
	"github.com/zenazn/goji/web"

	"github.com/justinas/nosurf"
)

var templateString = `
<!doctype html>
<html>
<body>
{{ if .name }}
<p>Your name: {{ .name }}</p>
{{ end }}
<form action="/signup/submit" method="POST">
<input type="text" name="name">

<!-- Try removing this or changing its value 
     and see what happens -->
<input type="hidden" name="csrf_token" value="{{ .csrf_token }}">
<input type="submit" value="Send">
</form>
</body>
</html>
`

var templ = template.Must(template.New("t1").Parse(templateString))

type M map[string]interface{}

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Our index page")
}

func ShowSignupForm(c web.C, w http.ResponseWriter, r *http.Request) {
	templ.Execute(w, M{
		"csrf_token": nosurf.Token(r), // Pass the CSRF token to the template
	})
}

func SubmitSignupForm(c web.C, w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "Successfully submitted, %s!", r.FormValue("name"))
}

func main() {
	goji.Get("/", IndexHandler) // Doesn't need CSRF protection (no POST/PUT/DELETE actions).

	signup := web.New()
	goji.Handle("/signup/*", signup)
	// But our signup forms do, so we add nosurf to their middleware stack (only).
	signup.Use(nosurf.NewPure)
	signup.Get("/signup/new", ShowSignupForm)
	signup.Post("/signup/submit", SubmitSignupForm)

	admin := web.New()
	// A more advanced example: we enforce secure cookies (HTTPS only),
	// set a domain and keep the expiry time low.
	a := nosurf.New(admin)
	a.SetBaseCookie(http.Cookie{
		Name:     "csrf_token",
		Domain:   "localhost",
		Path:     "/admin",
		MaxAge:   3600 * 4,
		HttpOnly: true,
		Secure:   true,
	})

	// Our /admin/* routes now have CSRF protection.
	goji.Handle("/admin/*", a)

	goji.Serve()
}

// Demonstrates advanced usage of nosurf in conjuction with net/http:
// * wrapping DefaultServeMux (http.Handle(), etc.)
// * exempting URLs
// * setting your own failure handler
package main

import (
	"fmt"
	"html/template"
	"net/http"

	"github.com/justinas/nosurf"
)

var templateString = `
<!doctype html>
<html>
<body>
{{ if .name }}
<p>Your name: {{ .name }}</p>
{{ end }}
<form action="" method="POST">
<input type="text" name="name">
<input type="hidden" name="csrf_token" value="{{ .token }}">
<input type="submit" value="Send">
</form>
</body>
</html>
`

var templ = template.Must(template.New("t1").Parse(templateString))

var esc = template.HTMLEscapeString

func Index(w http.ResponseWriter, r *http.Request) {
	context := map[string]string{
		"token": nosurf.Token(r),
	}
	if r.Method == "POST" {
		context["name"] = r.FormValue("name")
	}

	templ.Execute(w, context)
}

func Handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Name: %s", esc(r.FormValue("name")))
}

func failHand(w http.ResponseWriter, r *http.Request) {
	// will return the reason of the failure
	fmt.Fprintf(w, "%s\n", nosurf.Reason(r))
}

func main() {
	http.HandleFunc("/", Index)

	// when you route urls with .Handle[Func]() they end up on DefaultServeMux
	csrfHandler := nosurf.New(http.DefaultServeMux)

	// exempting by an exact path...
	// won't exempt /faq/question-1
	csrfHandler.ExemptPath("/faq")

	// exempting by a glob
	// will exempt /post, /post1, /post2, etc.
	// won't exempt /post1/comments, as * stops at a /
	csrfHandler.ExemptGlob("/post*")

	// exempting by a regexp
	// will exempt /static, /static/, /static/favicon.ico, /static/css/style.css, etc.
	csrfHandler.ExemptRegexp("/static(.*)")

	// setting the failureHandler. Will call this in case the CSRF check fails.
	csrfHandler.SetFailureHandler(http.HandlerFunc(failHand))

	http.ListenAndServe(":8000", csrfHandler)
}

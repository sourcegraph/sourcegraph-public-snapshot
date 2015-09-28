// Demonstrates usage of nosurf with the web.go package
package main

import (
	"html/template"
	"net/http"

	"github.com/hoisie/web"
	"github.com/justinas/nosurf"
)

var templateString = `
<!doctype html>
<html>
<body>
{{ if .name }}
<p>Your name: {{ .name }}</p>
{{ end }}
<form action="/" method="POST">
<input type="text" name="name">
<input type="hidden" name="csrf_token" value="{{ .token }}">
<input type="submit" value="Send">
</form>
</body>
</html>
`

var templ = template.Must(template.New("t1").Parse(templateString))

func myHandler(ctx *web.Context) {
	templateCtx := make(map[string]string)
	templateCtx["token"] = nosurf.Token(ctx.Request)

	if ctx.Request.Method == "POST" {
		templateCtx["name"] = ctx.Params["name"]
	}

	templ.Execute(ctx, templateCtx)
}

func main() {
	server := web.NewServer()
	server.Get("/", myHandler)
	server.Post("/", myHandler)

	http.ListenAndServe(":8000", nosurf.New(server))
}

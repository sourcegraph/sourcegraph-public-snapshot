// Demonstrates the simplest usage of nosurf
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
<form action="/" method="POST">
<input type="text" name="name">

<!-- Try removing this or changing its value 
     and see what happens -->
<input type="hidden" name="csrf_token" value="{{ .token }}">
<input type="submit" value="Send">
</form>
</body>
</html>
`

var templ = template.Must(template.New("t1").Parse(templateString))

func myFunc(w http.ResponseWriter, r *http.Request) {
	context := make(map[string]string)
	context["token"] = nosurf.Token(r)

	if r.Method == "POST" {
		context["name"] = r.FormValue("name")
	}

	templ.Execute(w, context)
}

func main() {
	myHandler := http.HandlerFunc(myFunc)
	fmt.Println("Listening on http://127.0.0.1:8000/")
	http.ListenAndServe(":8000", nosurf.New(myHandler))
}

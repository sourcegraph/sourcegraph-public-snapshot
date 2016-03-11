# nosurf

[![Build Status](https://travis-ci.org/justinas/nosurf.svg?branch=master)](https://travis-ci.org/justinas/nosurf)
[![GoDoc](http://godoc.org/github.com/justinas/nosurf?status.png)](http://godoc.org/github.com/justinas/nosurf)

`nosurf` is an HTTP package for Go
that helps you prevent Cross-Site Request Forgery attacks.
It acts like a middleware and therefore 
is compatible with basically any Go HTTP application.

### Why?
Even though CSRF is a prominent vulnerability,
Go's web-related package infrastructure mostly consists of
micro-frameworks that neither do implement CSRF checks,
nor should they.

`nosurf` solves this problem by providing a `CSRFHandler`
that wraps your `http.Handler` and checks for CSRF attacks
on every non-safe (non-GET/HEAD/OPTIONS/TRACE) method.

`nosurf` requires Go 1.1 or later.

### Features

* Supports any `http.Handler` (frameworks, your own handlers, etc.)
and acts like one itself.
* Allows exempting specific endpoints from CSRF checks by
an exact URL, a glob, or a regular expression.
* Allows specifying your own failure handler. 
Want to present the hacker with an ASCII middle finger
instead of the plain old `HTTP 400`? No problem.
* Uses masked tokens to mitigate the BREACH attack.
* Has no dependencies outside the Go standard library.

### Example
```go
package main

import (
	"fmt"
	"github.com/justinas/nosurf"
	"html/template"
	"net/http"
)

var templateString string = `
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
```

More examples can be found in the 
[examples/](https://github.com/justinas/nosurf/tree/master/examples/) directory.
Feel free to add one for your favorite framework 
or an unusual setup of the default HTTP tools.

### Manual token verification
In some cases the CSRF token may be send through a non standard way,
e.g. a body or request is a JSON encoded message with one of the fields
being a token.

In such case the handler(path) should be excluded from an automatic
verification by using one of the exemption methods:

```go
	func (h *CSRFHandler) ExemptFunc(fn func(r *http.Request) bool)
	func (h *CSRFHandler) ExemptGlob(pattern string)
	func (h *CSRFHandler) ExemptGlobs(patterns ...string)
	func (h *CSRFHandler) ExemptPath(path string)
	func (h *CSRFHandler) ExemptPaths(paths ...string)
	func (h *CSRFHandler) ExemptRegexp(re interface{})
	func (h *CSRFHandler) ExemptRegexps(res ...interface{})
```

Later on, the token **must** be verified by manually getting the token from the cookie
and providing the token sent in body through: `VerifyToken(tkn, tkn2 string) bool`.

Example:
```go
func HandleJson(w http.ResponseWriter, r *http.Request) {
	d := struct{
		X,Y int
		Tkn string
	}{}
	json.Unmarshal(ioutil.ReadAll(r.Body), &d)
	if !nosurf.VerifyToken(Token(r), d.Tkn) {
		http.Errorf(w, "CSRF token incorrect", http.StatusBadRequest)
		return
	}
	// do smth cool
}
```

### Contributing

0. Find an issue that bugs you / open a new one.
1. Discuss.
2. Branch off, commit, test.
3. Make a pull request / attach the commits to the issue.

package browser

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/headzoo/ut"
	"sourcegraph.com/sourcegraph/surf/jar"
)

func TestBrowserForm(t *testing.T) {
	ut.Run(t)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			fmt.Fprint(w, htmlForm)
		} else {
			r.ParseForm()
			fmt.Fprint(w, r.Form.Encode())
		}
	}))
	defer ts.Close()

	bow := &Browser{}
	bow.headers = make(http.Header, 10)
	bow.history = jar.NewMemoryHistory()

	err := bow.Open(ts.URL)
	ut.AssertNil(err)

	f, err := bow.Form("[name='default']")
	ut.AssertNil(err)

	f.Input("age", "55")
	f.Input("gender", "male")
	err = f.Click("submit2")
	ut.AssertNil(err)
	ut.AssertContains("age=55", bow.Body())
	ut.AssertContains("gender=male", bow.Body())
	ut.AssertContains("submit2=submitted2", bow.Body())
}

var htmlForm = `<!doctype html>
<html>
	<head>
		<title>Echo Form</title>
	</head>
	<body>
		<form method="post" action="/" name="default">
			<input type="text" name="age" value="" />
			<input type="radio" name="gender" value="male" />
			<input type="radio" name="gender" value="female" />
			<input type="submit" name="submit1" value="submitted1" />
			<input type="submit" name="submit2" value="submitted2" />
		</form>
	</body>
</html>
`

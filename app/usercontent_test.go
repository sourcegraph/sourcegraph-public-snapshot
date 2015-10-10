package app_test

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"golang.org/x/net/webdav"
	"src.sourcegraph.com/sourcegraph/app/internal/apptest"
	"src.sourcegraph.com/sourcegraph/app/router"
	"src.sourcegraph.com/sourcegraph/ui"
	ui_router "src.sourcegraph.com/sourcegraph/ui/router"
	"src.sourcegraph.com/sourcegraph/usercontent"
	"src.sourcegraph.com/sourcegraph/util/httptestutil"
)

func TestUserContent(t *testing.T) {
	origStore := usercontent.Store
	usercontent.Store = webdav.NewMemFS()
	defer func() { usercontent.Store = origStore }()

	const content = "ABC...z"
	var name string

	{
		uic, _ := httptestutil.NewTest(ui.NewHandler(nil, false))

		req, err := http.NewRequest("POST", ui_router.RelURLTo(ui_router.UserContentUpload).String(), strings.NewReader(content))
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Content-Type", "image/png")
		resp, err := uic.Do(req)
		if err != nil {
			t.Fatal(err)
		}

		var upload struct {
			Name  string
			Error string
		}
		err = json.NewDecoder(resp.Body).Decode(&upload)
		if err != nil {
			t.Fatal(err)
		}
		if upload.Error != "" {
			t.Errorf("%s %s response: %s", req.Method, ui_router.RelURLTo(ui_router.UserContentUpload).String(), upload.Error)
		}

		name = upload.Name
	}

	{
		c, _ := apptest.New()

		resp, err := c.Get(router.Rel.URLTo(router.UserContent, "Name", name).String())
		if err != nil {
			t.Fatal(err)
		}
		if err := checkHeader(resp, "Content-Type", "image/png"); err != nil {
			t.Error(err)
		}
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Fatal(err)
		}
		if string(body) != content {
			t.Errorf("body doesn't match expected content")
		}
	}
}

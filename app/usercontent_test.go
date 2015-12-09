package app_test

import (
	"bytes"
	"encoding/json"
	"image/png"
	"io/ioutil"
	"net/http"
	"testing"

	"image"

	"sourcegraph.com/sourcegraph/rwvfs"
	"src.sourcegraph.com/sourcegraph/app/internal/apptest"
	"src.sourcegraph.com/sourcegraph/app/router"
	"src.sourcegraph.com/sourcegraph/ui"
	ui_router "src.sourcegraph.com/sourcegraph/ui/router"
	"src.sourcegraph.com/sourcegraph/usercontent"
	"src.sourcegraph.com/sourcegraph/util/httptestutil"
)

func TestUserContent(t *testing.T) {
	origStore := usercontent.Store
	usercontent.Store = rwvfs.Map(make(map[string]string))
	defer func() { usercontent.Store = origStore }()

	image := image.NewRGBA(image.Rect(0, 0, 10, 10))
	content := new(bytes.Buffer)
	png.Encode(content, image)
	imageString := content.String()

	var name string

	{
		uic, _ := httptestutil.NewTest(ui.NewHandler(nil, false))

		req, err := http.NewRequest("POST", ui_router.Rel.URLTo(ui_router.UserContentUpload).String(), content)
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
			t.Errorf("%s %s response: %s", req.Method, ui_router.Rel.URLTo(ui_router.UserContentUpload).String(), upload.Error)
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

		if string(body) != imageString {
			t.Errorf("body doesn't match expected content\n")
		}
	}
}

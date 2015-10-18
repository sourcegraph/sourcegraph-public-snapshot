package issues

import (
	"bytes"
	"html/template"
	"io/ioutil"
	"log"

	netctx "golang.org/x/net/context"
	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/platform/apps/issues/assets"
	"src.sourcegraph.com/sourcegraph/platform/apps/issues/markdown"
	"src.sourcegraph.com/sourcegraph/platform/putil"
)

var fmap = map[string]interface{}{
	"render": render,
}

func render(ctx netctx.Context, e Event) template.HTML {
	t := parseReply("reply.html")
	wr := new(bytes.Buffer)

	sg := sourcegraph.NewClientFromContext(ctx)
	author, err := sg.Users.Get(ctx, &sourcegraph.UserSpec{UID: e.AuthorUID})
	if err != nil {
		return template.HTML("unknown author")
	}

	err = t.Execute(wr, struct {
		Author  string
		Body    template.HTML
		UID     int
		Creator bool
	}{
		Author:  author.Name,
		Body:    markdown.Parse(e.Body),
		UID:     e.UID,
		Creator: putil.UserFromContext(ctx).UID == e.AuthorUID,
	})
	if err != nil {
		log.Println(err)
		return template.HTML("")
	}
	b, err := ioutil.ReadAll(wr)
	if err != nil {
		log.Println(err)
		return template.HTML("")
	}
	return template.HTML(string(b))
}

func parse(tmpl string) *template.Template {
	f, err := assets.Assets.Open("/" + tmpl)
	if err != nil {
		panic(err)
	}
	b, err := ioutil.ReadAll(f)
	f.Close()
	if err != nil {
		panic(err)
	}
	t := template.New(tmpl)
	t = t.Funcs(fmap)
	t, err = t.Parse(string(b))
	if err != nil {
		panic(err)
	}
	return t
}

// Hack to avoid initialization cycles.
func parseReply(tmpl string) *template.Template {
	f, err := assets.Assets.Open("/" + tmpl)
	if err != nil {
		panic(err)
	}
	b, err := ioutil.ReadAll(f)
	f.Close()
	if err != nil {
		panic(err)
	}
	t := template.New(tmpl)
	t, err = t.Parse(string(b))
	if err != nil {
		panic(err)
	}
	return t
}

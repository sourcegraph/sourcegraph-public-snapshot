package surf

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/headzoo/ut"
	"sourcegraph.com/sourcegraph/surf/jar"
)

func TestGet(t *testing.T) {
	ut.Run(t)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path == "/page1" {
			fmt.Fprint(w, htmlPage1)
		} else if req.URL.Path == "/page2" {
			fmt.Fprint(w, htmlPage2)
		}
	}))
	defer ts.Close()

	bow := NewBrowser()

	err := bow.Open(ts.URL + "/page1")
	ut.AssertNil(err)
	ut.AssertEquals("Surf Page 1", bow.Title())
	ut.AssertContains("<p>Hello, Surf!</p>", bow.Body())

	err = bow.Open(ts.URL + "/page2")
	ut.AssertNil(err)
	ut.AssertEquals("Surf Page 2", bow.Title())

	ok := bow.Back()
	ut.AssertTrue(ok)
	ut.AssertEquals("Surf Page 1", bow.Title())

	ok = bow.Back()
	ut.AssertFalse(ok)
	ut.AssertEquals("Surf Page 1", bow.Title())
}

func TestPost(t *testing.T) {
	ut.Run(t)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Method == "POST" {
			w.WriteHeader(200)
			w.Write([]byte("OK"))
		} else {
			w.WriteHeader(500)
			w.Write([]byte("ERROR"))
		}
	}))
	defer ts.Close()

	bow := NewBrowser()
	bow.Post(ts.URL, "application/x-www-form-urlencoded", nil)
	ut.AssertEquals(200, bow.StatusCode())
}

func TestHead(t *testing.T) {
	ut.Run(t)
	var r *http.Request

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		r = req
	}))
	defer ts.Close()

	bow := NewBrowser()

	err := bow.Head(ts.URL + "/page1")
	ut.AssertNil(err)
	ut.AssertNotNil(r)
}

func TestDownload(t *testing.T) {
	ut.Run(t)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		fmt.Fprint(w, req.UserAgent())
	}))
	defer ts.Close()

	bow := NewBrowser()
	bow.Open(ts.URL)

	buff := &bytes.Buffer{}
	l, err := bow.Download(buff)
	ut.AssertNil(err)
	ut.AssertGreaterThan(0, int(l))
	ut.AssertEquals(int(l), buff.Len())
}

func TestDownloadContentType(t *testing.T) {
	ut.Run(t)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		b := bytes.NewBufferString("Hello")
		fmt.Fprint(w, b)
	}))
	defer ts.Close()

	bow := NewBrowser()
	bow.Open(ts.URL)

	buff := &bytes.Buffer{}
	bow.Download(buff)
	ut.AssertEquals("Hello", buff.String())
}

func TestUserAgent(t *testing.T) {
	ut.Run(t)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		fmt.Fprint(w, req.UserAgent())
	}))
	defer ts.Close()

	bow := NewBrowser()
	bow.SetUserAgent("Testing/1.0")
	err := bow.Open(ts.URL)
	ut.AssertNil(err)
	ut.AssertEquals("Testing/1.0", bow.Body())
}

func TestHeaders(t *testing.T) {
	ut.Run(t)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		fmt.Fprint(w, req.Header.Get("X-Testing-1"))
		fmt.Fprint(w, req.Header.Get("X-Testing-2"))
	}))
	defer ts.Close()

	bow := NewBrowser()
	bow.AddRequestHeader("X-Testing-1", "Testing-1")
	bow.AddRequestHeader("X-Testing-2", "Testing-2")
	err := bow.Open(ts.URL)
	ut.AssertNil(err)
	ut.AssertContains("Testing-1", bow.Body())
	ut.AssertContains("Testing-2", bow.Body())
}

// TestHeadersSet
// See: https://github.com/headzoo/surf/pull/19
func TestHeadersBug19(t *testing.T) {
	ut.Run(t)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		fmt.Fprint(w, req.Header.Get("X-Testing"))
	}))
	defer ts.Close()

	bow := NewBrowser()
	bow.AddRequestHeader("X-Testing", "Testing-1")
	bow.AddRequestHeader("X-Testing", "Testing-2")
	err := bow.Open(ts.URL)
	ut.AssertNil(err)
	ut.AssertContains("Testing-2", bow.Body())
}

func TestBookmarks(t *testing.T) {
	ut.Run(t)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, htmlPage1)
	}))
	defer ts.Close()

	bookmarks := jar.NewMemoryBookmarks()
	bow := NewBrowser()
	bow.SetBookmarksJar(bookmarks)

	bookmarks.Save("test1", ts.URL)
	bow.OpenBookmark("test1")
	ut.AssertEquals("Surf Page 1", bow.Title())
	ut.AssertContains("<p>Hello, Surf!</p>", bow.Body())

	err := bow.Bookmark("test2")
	ut.AssertNil(err)
	bow.OpenBookmark("test2")
	ut.AssertEquals("Surf Page 1", bow.Title())
}

func TestClick(t *testing.T) {
	ut.Run(t)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			fmt.Fprint(w, htmlPage1)
		} else if r.URL.Path == "/page2" {
			fmt.Fprint(w, htmlPage1)
		}
	}))
	defer ts.Close()

	bow := NewBrowser()
	err := bow.Open(ts.URL)
	ut.AssertNil(err)

	err = bow.Click("a:contains('click')")
	ut.AssertNil(err)
	ut.AssertContains("<p>Hello, Surf!</p>", bow.Body())
}

func TestLinks(t *testing.T) {
	ut.Run(t)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, htmlPage1)
	}))
	defer ts.Close()

	bow := NewBrowser()
	err := bow.Open(ts.URL)
	ut.AssertNil(err)

	links := bow.Links()
	ut.AssertEquals(2, len(links))
	ut.AssertEquals("", links[0].ID)
	ut.AssertEquals(ts.URL+"/page2", links[0].URL.String())
	ut.AssertEquals("click", links[0].Text)
	ut.AssertEquals("page3", links[1].ID)
	ut.AssertEquals(ts.URL+"/page3", links[1].URL.String())
	ut.AssertEquals("no clicking", links[1].Text)
}

func TestImages(t *testing.T) {
	ut.Run(t)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, htmlPage1)
	}))
	defer ts.Close()

	bow := NewBrowser()
	err := bow.Open(ts.URL)
	ut.AssertNil(err)

	images := bow.Images()
	ut.AssertEquals(2, len(images))
	ut.AssertEquals("imgur-image", images[0].ID)
	ut.AssertEquals("http://i.imgur.com/HW4bJtY.jpg", images[0].URL.String())
	ut.AssertEquals("", images[0].Alt)
	ut.AssertEquals("It's a...", images[0].Title)

	ut.AssertEquals("", images[1].ID)
	ut.AssertEquals(ts.URL+"/Cxagv.jpg", images[1].URL.String())
	ut.AssertEquals("A picture", images[1].Alt)
	ut.AssertEquals("", images[1].Title)

	buff := &bytes.Buffer{}
	l, err := images[0].Download(buff)
	ut.AssertNil(err)
	ut.AssertGreaterThan(0, buff.Len())
	ut.AssertEquals(int(l), buff.Len())
}

func TestStylesheets(t *testing.T) {
	ut.Run(t)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, htmlPage1)
	}))
	defer ts.Close()

	bow := NewBrowser()
	err := bow.Open(ts.URL)
	ut.AssertNil(err)

	stylesheets := bow.Stylesheets()
	ut.AssertEquals(2, len(stylesheets))
	ut.AssertEquals("http://godoc.org/-/site.css", stylesheets[0].URL.String())
	ut.AssertEquals("all", stylesheets[0].Media)
	ut.AssertEquals("text/css", stylesheets[0].Type)

	ut.AssertEquals(ts.URL+"/print.css", stylesheets[1].URL.String())
	ut.AssertEquals("print", stylesheets[1].Media)
	ut.AssertEquals("text/css", stylesheets[1].Type)

	buff := &bytes.Buffer{}
	l, err := stylesheets[0].Download(buff)
	ut.AssertNil(err)
	ut.AssertGreaterThan(0, buff.Len())
	ut.AssertEquals(int(l), buff.Len())
}

func TestScripts(t *testing.T) {
	ut.Run(t)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, htmlPage1)
	}))
	defer ts.Close()

	bow := NewBrowser()
	err := bow.Open(ts.URL)
	ut.AssertNil(err)

	scripts := bow.Scripts()
	ut.AssertEquals(2, len(scripts))
	ut.AssertEquals("http://godoc.org/-/site.js", scripts[0].URL.String())
	ut.AssertEquals("text/javascript", scripts[0].Type)

	ut.AssertEquals(ts.URL+"/jquery.min.js", scripts[1].URL.String())
	ut.AssertEquals("text/javascript", scripts[1].Type)

	buff := &bytes.Buffer{}
	l, err := scripts[0].Download(buff)
	ut.AssertNil(err)
	ut.AssertGreaterThan(0, buff.Len())
	ut.AssertEquals(int(l), buff.Len())
}

var htmlPage1 = `<!doctype html>
<html>
	<head>
		<title>Surf Page 1</title>
		<link href="/favicon.ico" rel="icon" type="image/x-icon">
		<link href="http://godoc.org/-/site.css" media="all" rel="stylesheet" type="text/css" />
		<link href="/print.css" rel="stylesheet" media="print" />
	</head>
	<body>
		<p>Hello, Surf!</p>
		<img src="http://i.imgur.com/HW4bJtY.jpg" id="imgur-image" title="It's a..." />
		<img src="/Cxagv.jpg" alt="A picture" />

		<p>Click the link below.</p>
		<a href="/page2">click</a>
		<a href="/page3" id="page3">no clicking</a>

		<script src="http://godoc.org/-/site.js" type="text/javascript"></script>
		<script src="/jquery.min.js" type="text/javascript"></script>
		<script type="text/javascript">
			var _gaq = _gaq || [];
		</script>
	</body>
</html>
`

var htmlPage2 = `<!doctype html>
<html>
	<head>
		<title>Surf Page 2</title>
	</head>
	<body>
		<p>Hello, Surf!</p>
	</body>
</html>
`

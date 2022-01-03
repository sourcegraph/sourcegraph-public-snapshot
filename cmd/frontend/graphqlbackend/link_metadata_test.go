package graphqlbackend

import (
	"fmt"
	"testing"
)

func TestLinkMetadata(t *testing.T) {
	title := "title"
	description := "description"
	imageUrl := "imageUrl"

	t.Run("can parse valid HTML", func(t *testing.T) {
		html := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en" data-color-mode="dark" data-light-theme="light" data-dark-theme="dark">
  <head>
    <meta charset="utf-8">
    <link rel="dns-prefetch" href="https://github.githubassets.com">
    <meta name="viewport" content="test">
    <title>sourcegraph/sourcegraph: Universal code search (self-hosted)</title>
    <meta name="description" content="test">
    <link rel="search" type="application/opensearchdescription+xml" href="/opensearch.xml" title="GitHub">
    <meta property="fb:app_id" content="test">
    <meta name="apple-itunes-app" content="test" />
    <meta name="twitter:image:src" content="test" />
    <meta name="twitter:site" content="test" />
    <meta name="twitter:title" content="test" />
    <meta name="twitter:description" content="test" />
    <meta property="og:image" content="%s" />
    <meta property="og:image:width" content="1200" />
    <meta property="og:title" content="%s" />
    <meta property="og:url" content="test" />
    <meta property="og:description" content="%s" />
    <link rel="assets" href="https://github.githubassets.com/">
    <!-- To prevent page flashing, the optimizely JS needs to be loaded in the
         <head> tag before the DOM renders -->
    <meta name="hostname" content="github.com">
    <script type="application/json" id="memex_keyboard_shortcuts_preference">"all"</script>
    <meta name="theme-color" content="#1e2327">
    <meta name="color-scheme" content="dark light" />
  </head>
  <body class="logged-in env-production page-responsive" style="word-wrap: break-word;">
    <div class="position-relative js-header-wrapper "></div>
  </body>
</html>`, imageUrl, title, description)

		linkMetadata := parseBody(html)

		if linkMetadata.title == nil {
			t.Errorf("title is nil when it should be a value")
		} else if *linkMetadata.title != title {
			t.Errorf("got %q want %q", *linkMetadata.title, title)
		}
		if linkMetadata.description == nil {
			t.Errorf("description is nil when it should be a value")
		} else if *linkMetadata.description != description {
			t.Errorf("got %q want %q", *linkMetadata.description, description)
		}
		if linkMetadata.imageUrl == nil {
			t.Errorf("imageUrl is nil when it should be a value")
		} else if *linkMetadata.imageUrl != imageUrl {
			t.Errorf("got %q want %q", *linkMetadata.imageUrl, imageUrl)
		}
	})

	t.Run("prefers og: then twitter: then naked", func(t *testing.T) {
		html := fmt.Sprintf(`<!DOCTYPE html>
<html>
  <head>
    <meta name="description" content="test">
    <meta name="title" content="test">
    <meta property="og:title" content="%s" />
    <meta property="image" content="%s" />
    <meta name="twitter:title" content="test" />
    <meta name="twitter:description" content="%s" />
  </head>
</html>`, title, imageUrl, description)

		linkMetadata := parseBody(html)

		if linkMetadata.title == nil {
			t.Errorf("title is nil when it should be a value")
		} else if *linkMetadata.title != title {
			t.Errorf("got %q want %q", *linkMetadata.title, title)
		}
		if linkMetadata.description == nil {
			t.Errorf("description is nil when it should be a value")
		} else if *linkMetadata.description != description {
			t.Errorf("got %q want %q", *linkMetadata.description, description)
		}
		if linkMetadata.imageUrl == nil {
			t.Errorf("imageUrl is nil when it should be a value")
		} else if *linkMetadata.imageUrl != imageUrl {
			t.Errorf("got %q want %q", *linkMetadata.imageUrl, imageUrl)
		}
	})

	t.Run("returns nil value if no metadata is found for a field", func(t *testing.T) {
		html := fmt.Sprintf(`<!DOCTYPE html>
<html>
  <head>
    <meta property="og:title" content="%s" />
  </head>
</html>`, title)

		linkMetadata := parseBody(html)

		if linkMetadata.title == nil {
			t.Errorf("title is nil when it should be a value")
		} else if *linkMetadata.title != title {
			t.Errorf("got %q want %q", *linkMetadata.title, title)
		}
		if linkMetadata.description != nil {
			t.Errorf("description should be nil")
		}
		if linkMetadata.imageUrl != nil {
			t.Errorf("imageUrl should be nil")
		}
	})

	t.Run("returns a struct of three nil references if no metadata is found for any fields", func(t *testing.T) {
		html := fmt.Sprintf(`<!DOCTYPE html>
<html>
  <head>
  </head>
</html>`)

		linkMetadata := parseBody(html)

		if linkMetadata.title != nil {
			t.Errorf("title should be nil")
		}
		if linkMetadata.description != nil {
			t.Errorf("description should be nil")
		}
		if linkMetadata.imageUrl != nil {
			t.Errorf("imageUrl should be nil")
		}
	})

	t.Run("returns nil values on invalid/weird HTML", func(t *testing.T) {
		html := `abc<!DOCTYPE html><html><meta charset="UTF-8" /><title>Invalid HTML Example</title></head><body>`

		linkMetadata := parseBody(html)

		if linkMetadata.title != nil || linkMetadata.description != nil || linkMetadata.imageUrl != nil {
			t.Errorf("all fields should be nil")
		}
	})
}

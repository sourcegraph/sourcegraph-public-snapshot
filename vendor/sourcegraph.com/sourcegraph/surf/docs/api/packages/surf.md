# surf
--
    import "github.com/headzoo/surf"

Package surf ensembles other packages into a usable browser.

## Usage

```go
var (
	// DefaultUserAgent is the global user agent value.
	DefaultUserAgent = agent.Create()

	// DefaultSendReferer is the global value for the AttributeSendReferer attribute.
	DefaultSendReferer = true

	// DefaultMetaRefreshHandling is the global value for the AttributeHandleRefresh attribute.
	DefaultMetaRefreshHandling = true

	// DefaultFollowRedirects is the global value for the AttributeFollowRedirects attribute.
	DefaultFollowRedirects = true
)
```

#### func  NewBrowser

```go
func NewBrowser() *browser.Browser
```
NewBrowser creates and returns a *browser.Browser type.

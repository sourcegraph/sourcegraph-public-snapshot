# browser
--
    import "github.com/headzoo/surf/browser"

Package browser contains the primary browser implementation.

## Usage

```go
var InitialAssetsSliceSize = 20
```
InitialAssetsArraySize is the initial size when allocating a slice of page
assets. Increasing this size may lead to a very small performance increase when
downloading assets from a page with a lot of assets.

#### func  DownloadAsset

```go
func DownloadAsset(asset Downloadable, out io.Writer) (int64, error)
```
DownloadAsset copies a remote file to the given writer.

#### func  DownloadAssetAsync

```go
func DownloadAssetAsync(asset Downloadable, out io.Writer, c AsyncDownloadChannel)
```
DownloadAssetAsync downloads an asset asynchronously and notifies the given
channel when the download is complete.

#### type Asset

```go
type Asset struct {
	// ID is the value of the id attribute if available.
	ID string

	// URL is the asset URL.
	URL *url.URL

	// Type describes the type of asset.
	Type AssetType
}
```

Asset implements Assetable.

#### func (*Asset) AssetType

```go
func (at *Asset) AssetType() AssetType
```
Type returns the asset type.

#### func (*Asset) Id

```go
func (at *Asset) Id() string
```
Id returns the asset ID or an empty string when not available.

#### func (*Asset) Url

```go
func (at *Asset) Url() *url.URL
```
Url returns the asset URL.

#### type AssetType

```go
type AssetType uint16
```

AssetType describes a type of page asset, such as an image or stylesheet.

```go
const (
	// LinkAsset describes a *Link asset.
	LinkAsset AssetType = iota

	// ImageAsset describes an *Image asset.
	ImageAsset

	// StylesheetAsset describes a *Stylesheet asset.
	StylesheetAsset

	// ScriptAsset describes a *Script asset.
	ScriptAsset
)
```

#### type Assetable

```go
type Assetable interface {
	// Url returns the asset URL.
	Url() *url.URL

	// Id returns the asset ID or an empty string when not available.
	Id() string

	// Type describes the type of asset.
	AssetType() AssetType
}
```

Assetable represents a page asset, such as an image or stylesheet.

#### type AsyncDownloadChannel

```go
type AsyncDownloadChannel chan *AsyncDownloadResult
```

AsyncDownloadChannel is a channel upon which the results of an async download
are passed.

#### type AsyncDownloadResult

```go
type AsyncDownloadResult struct {
	// Asset is a pointer to the Downloadable asset that was downloaded.
	Asset Downloadable

	// Writer where the asset data was written.
	Writer io.Writer

	// Size is the number of bytes written to the io.Writer.
	Size int64

	// Error contains any error that occurred during the download or nil.
	Error error
}
```

AsyncDownloadResult has the results of an asynchronous download.

#### type Attribute

```go
type Attribute int
```

Attribute represents a Browser capability.

```go
const (
	// SendRefererAttribute instructs a Browser to send the Referer header.
	SendReferer Attribute = iota

	// MetaRefreshHandlingAttribute instructs a Browser to handle the refresh meta tag.
	MetaRefreshHandling

	// FollowRedirectsAttribute instructs a Browser to follow Location headers.
	FollowRedirects
)
```

#### type AttributeMap

```go
type AttributeMap map[Attribute]bool
```

AttributeMap represents a map of Attribute values.

#### type Browsable

```go
type Browsable interface {
	// SetUserAgent sets the user agent.
	SetUserAgent(ua string)

	// SetAttribute sets a browser instruction attribute.
	SetAttribute(a Attribute, v bool)

	// SetAttributes is used to set all the browser attributes.
	SetAttributes(a AttributeMap)

	// SetState sets the init browser state.
	SetState(sj *jar.State)

	// SetBookmarksJar sets the bookmarks jar the browser uses.
	SetBookmarksJar(bj jar.BookmarksJar)

	// SetCookieJar is used to set the cookie jar the browser uses.
	SetCookieJar(cj http.CookieJar)

	// SetHistoryJar is used to set the history jar the browser uses.
	SetHistoryJar(hj jar.History)

	// SetHeadersJar sets the headers the browser sends with each request.
	SetHeadersJar(h http.Header)

	// AddRequestHeader adds a header the browser sends with each request.
	AddRequestHeader(name, value string)

	// Open requests the given URL using the GET method.
	Open(url string) error

	// OpenForm appends the data values to the given URL and sends a GET request.
	OpenForm(url string, data url.Values) error

	// OpenBookmark calls Get() with the URL for the bookmark with the given name.
	OpenBookmark(name string) error

	// Post requests the given URL using the POST method.
	Post(url string, contentType string, body io.Reader) error

	// PostForm requests the given URL using the POST method with the given data.
	PostForm(url string, data url.Values) error

	// PostMultipart requests the given URL using the POST method with the given data using multipart/form-data format.
	PostMultipart(u string, data url.Values) error

	// Back loads the previously requested page.
	Back() bool

	// Reload duplicates the last successful request.
	Reload() error

	// Bookmark saves the page URL in the bookmarks with the given name.
	Bookmark(name string) error

	// Click clicks on the page element matched by the given expression.
	Click(expr string) error

	// Form returns the form in the current page that matches the given expr.
	Form(expr string) (Submittable, error)

	// Forms returns an array of every form in the page.
	Forms() []Submittable

	// Links returns an array of every link found in the page.
	Links() []*Link

	// Images returns an array of every image found in the page.
	Images() []*Image

	// Stylesheets returns an array of every stylesheet linked to the document.
	Stylesheets() []*Stylesheet

	// Scripts returns an array of every script linked to the document.
	Scripts() []*Script

	// SiteCookies returns the cookies for the current site.
	SiteCookies() []*http.Cookie

	// ResolveUrl returns an absolute URL for a possibly relative URL.
	ResolveUrl(u *url.URL) *url.URL

	// ResolveStringUrl works just like ResolveUrl, but the argument and return value are strings.
	ResolveStringUrl(u string) (string, error)

	// Download writes the contents of the document to the given writer.
	Download(o io.Writer) (int64, error)

	// Url returns the page URL as a string.
	Url() *url.URL

	// StatusCode returns the response status code.
	StatusCode() int

	// Title returns the page title.
	Title() string

	// ResponseHeaders returns the page headers.
	ResponseHeaders() http.Header

	// Body returns the page body as a string of html.
	Body() string

	// Dom returns the inner *goquery.Selection.
	Dom() *goquery.Selection

	// Find returns the dom selections matching the given expression.
	Find(expr string) *goquery.Selection
}
```

Browsable represents an HTTP web browser.

#### type Browser

```go
type Browser struct {
}
```

Default is the default Browser implementation.

#### func (*Browser) AddRequestHeader

```go
func (bow *Browser) AddRequestHeader(name, value string)
```
AddRequestHeader sets a header the browser sends with each request.

#### func (*Browser) Back

```go
func (bow *Browser) Back() bool
```
Back loads the previously requested page.

Returns a boolean value indicating whether a previous page existed, and was
successfully loaded.

#### func (*Browser) Body

```go
func (bow *Browser) Body() string
```
Body returns the page body as a string of html.

#### func (*Browser) Bookmark

```go
func (bow *Browser) Bookmark(name string) error
```
Bookmark saves the page URL in the bookmarks with the given name.

#### func (*Browser) Click

```go
func (bow *Browser) Click(expr string) error
```
Click clicks on the page element matched by the given expression.

Currently this is only useful for click on links, which will cause the browser
to load the page pointed at by the link. Future versions of Surf may support
JavaScript and clicking on elements will fire the click event.

#### func (*Browser) DelRequestHeader

```go
func (bow *Browser) DelRequestHeader(name string)
```
DelRequestHeader deletes a header so the browser will not send it with future
requests.

#### func (*Browser) Dom

```go
func (bow *Browser) Dom() *goquery.Selection
```
Dom returns the inner *goquery.Selection.

#### func (*Browser) Download

```go
func (bow *Browser) Download(o io.Writer) (int64, error)
```
Download writes the contents of the document to the given writer.

#### func (*Browser) Find

```go
func (bow *Browser) Find(expr string) *goquery.Selection
```
Find returns the dom selections matching the given expression.

#### func (*Browser) Form

```go
func (bow *Browser) Form(expr string) (Submittable, error)
```
Form returns the form in the current page that matches the given expr.

#### func (*Browser) Forms

```go
func (bow *Browser) Forms() []Submittable
```
Forms returns an array of every form in the page.

#### func (*Browser) Images

```go
func (bow *Browser) Images() []*Image
```
Images returns an array of every image found in the page.

#### func (*Browser) Links

```go
func (bow *Browser) Links() []*Link
```
Links returns an array of every link found in the page.

#### func (*Browser) Open

```go
func (bow *Browser) Open(u string) error
```
Open requests the given URL using the GET method.

#### func (*Browser) OpenBookmark

```go
func (bow *Browser) OpenBookmark(name string) error
```
OpenBookmark calls Open() with the URL for the bookmark with the given name.

#### func (*Browser) OpenForm

```go
func (bow *Browser) OpenForm(u string, data url.Values) error
```
OpenForm appends the data values to the given URL and sends a GET request.

#### func (*Browser) Post

```go
func (bow *Browser) Post(u string, contentType string, body io.Reader) error
```
Post requests the given URL using the POST method.

#### func (*Browser) PostForm

```go
func (bow *Browser) PostForm(u string, data url.Values) error
```
PostForm requests the given URL using the POST method with the given data.

#### func (*Browser) PostMultipart

```go
func (bow *Browser) PostMultipart(u string, data url.Values) error
```
PostMultipart requests the given URL using the POST method with the given data
using multipart/form-data format.

#### func (*Browser) Reload

```go
func (bow *Browser) Reload() error
```
Reload duplicates the last successful request.

#### func (*Browser) ResolveStringUrl

```go
func (bow *Browser) ResolveStringUrl(u string) (string, error)
```
ResolveStringUrl works just like ResolveUrl, but the argument and return value
are strings.

#### func (*Browser) ResolveUrl

```go
func (bow *Browser) ResolveUrl(u *url.URL) *url.URL
```
ResolveUrl returns an absolute URL for a possibly relative URL.

#### func (*Browser) ResponseHeaders

```go
func (bow *Browser) ResponseHeaders() http.Header
```
ResponseHeaders returns the page headers.

#### func (*Browser) Scripts

```go
func (bow *Browser) Scripts() []*Script
```
Scripts returns an array of every script linked to the document.

#### func (*Browser) SetAttribute

```go
func (bow *Browser) SetAttribute(a Attribute, v bool)
```
SetAttribute sets a browser instruction attribute.

#### func (*Browser) SetAttributes

```go
func (bow *Browser) SetAttributes(a AttributeMap)
```
SetAttributes is used to set all the browser attributes.

#### func (*Browser) SetBookmarksJar

```go
func (bow *Browser) SetBookmarksJar(bj jar.BookmarksJar)
```
SetBookmarksJar sets the bookmarks jar the browser uses.

#### func (*Browser) SetCookieJar

```go
func (bow *Browser) SetCookieJar(cj http.CookieJar)
```
SetCookieJar is used to set the cookie jar the browser uses.

#### func (*Browser) SetHeadersJar

```go
func (bow *Browser) SetHeadersJar(h http.Header)
```
SetHeadersJar sets the headers the browser sends with each request.

#### func (*Browser) SetHistoryJar

```go
func (bow *Browser) SetHistoryJar(hj jar.History)
```
SetHistoryJar is used to set the history jar the browser uses.

#### func (*Browser) SetState

```go
func (bow *Browser) SetState(sj *jar.State)
```
SetState sets the browser state.

#### func (*Browser) SetUserAgent

```go
func (bow *Browser) SetUserAgent(userAgent string)
```
SetUserAgent sets the user agent.

#### func (*Browser) SiteCookies

```go
func (bow *Browser) SiteCookies() []*http.Cookie
```
SiteCookies returns the cookies for the current site.

#### func (*Browser) StatusCode

```go
func (bow *Browser) StatusCode() int
```
StatusCode returns the response status code.

#### func (*Browser) Stylesheets

```go
func (bow *Browser) Stylesheets() []*Stylesheet
```
Stylesheets returns an array of every stylesheet linked to the document.

#### func (*Browser) Title

```go
func (bow *Browser) Title() string
```
Title returns the page title.

#### func (*Browser) Url

```go
func (bow *Browser) Url() *url.URL
```
Url returns the page URL as a string.

#### type Downloadable

```go
type Downloadable interface {
	Assetable

	// Download writes the contents of the element to the given writer.
	//
	// Returns the number of bytes written.
	Download(out io.Writer) (int64, error)

	// DownloadAsync downloads the contents of the element asynchronously.
	//
	// An instance of AsyncDownloadResult will be sent down the given channel
	// when the download is complete.
	DownloadAsync(out io.Writer, ch AsyncDownloadChannel)
}
```

Downloadable represents an asset that may be downloaded.

#### type DownloadableAsset

```go
type DownloadableAsset struct {
	Asset
}
```

DownloadableAsset is an asset that may be downloaded.

#### func (*DownloadableAsset) Download

```go
func (at *DownloadableAsset) Download(out io.Writer) (int64, error)
```
Download writes the asset to the given io.Writer type.

#### func (*DownloadableAsset) DownloadAsync

```go
func (at *DownloadableAsset) DownloadAsync(out io.Writer, ch AsyncDownloadChannel)
```
DownloadAsync downloads the asset asynchronously.

#### type Form

```go
type Form struct {
}
```

Form is the default form element.

#### func  NewForm

```go
func NewForm(bow Browsable, s *goquery.Selection) *Form
```
NewForm creates and returns a *Form type.

#### func (*Form) Action

```go
func (f *Form) Action() string
```
Action returns the form action URL. The URL will always be absolute.

#### func (*Form) Click

```go
func (f *Form) Click(button string) error
```
Click submits the form by clicking the button with the given name.

#### func (*Form) Dom

```go
func (f *Form) Dom() *goquery.Selection
```
Dom returns the inner *goquery.Selection.

#### func (*Form) Input

```go
func (f *Form) Input(name, value string) error
```
Input sets the value of a form field.

#### func (*Form) Method

```go
func (f *Form) Method() string
```
Method returns the form method, eg "GET" or "POST".

#### func (*Form) Submit

```go
func (f *Form) Submit() error
```
Submit submits the form. Clicks the first button in the form, or submits the
form without using any button when the form does not contain any buttons.

#### type Image

```go
type Image struct {
	DownloadableAsset

	// Alt is the value of the image alt attribute if available.
	Alt string

	// Title is the value of the image title attribute if available.
	Title string
}
```

Image stores the properties of an image.

#### func  NewImageAsset

```go
func NewImageAsset(url *url.URL, id, alt, title string) *Image
```
NewImageAsset creates and returns a new *Image type.

#### type Link

```go
type Link struct {
	Asset

	// Text is the text appearing between the opening and closing anchor tag.
	Text string
}
```

Link stores the properties of a page link.

#### func  NewLinkAsset

```go
func NewLinkAsset(u *url.URL, id, text string) *Link
```
NewLinkAsset creates and returns a new *Link type.

#### type Script

```go
type Script struct {
	DownloadableAsset

	// Type is the value of the type attribute. Defaults to "text/javascript" when not specified.
	Type string
}
```

Script stores the properties of a linked script.

#### func  NewScriptAsset

```go
func NewScriptAsset(url *url.URL, id, typ string) *Script
```
NewScriptAsset creates and returns a new *Script type.

#### type Stylesheet

```go
type Stylesheet struct {
	DownloadableAsset

	// Media is the value of the media attribute. Defaults to "all" when not specified.
	Media string

	// Type is the value of the type attribute. Defaults to "text/css" when not specified.
	Type string
}
```

Stylesheet stores the properties of a linked stylesheet.

#### func  NewStylesheetAsset

```go
func NewStylesheetAsset(url *url.URL, id, media, typ string) *Stylesheet
```
NewStylesheetAsset creates and returns a new *Stylesheet type.

#### type Submittable

```go
type Submittable interface {
	Method() string
	Action() string
	Input(name, value string) error
	Click(button string) error
	Submit() error
	Dom() *goquery.Selection
}
```

Submittable represents an element that may be submitted, such as a form.

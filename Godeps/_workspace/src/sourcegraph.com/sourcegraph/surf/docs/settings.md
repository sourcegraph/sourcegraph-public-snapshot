# User Agent
Set the user agent this browser instance will send with each request.
```go
bow := surf.NewBrowser()
bow.SetUserAgent("SuperCrawler/1.0")
```

Or set the user agent globally so every new browser you create uses it.
```go
browser.DefaultUserAgent = "SuperCrawler/1.0"
```

# Attributes
Attributes control how the browser behaves. Use the SetAttribute() method
to set attributes one at a time.
```go
bow := surf.NewBrowser()
bow.SetAttribute(browser.SendReferer, false)
bow.SetAttribute(browser.MetaRefreshHandling, false)
bow.SetAttribute(browser.FollowRedirects, false)
```

Or set the attributes all at once using SetAttributes().
```go
bow := surf.NewBrowser()
bow.SetAttributes(browser.AttributeMap{
    browser.SendReferer:         surf.DefaultSendReferer,
    browser.MetaRefreshHandling: surf.DefaultMetaRefreshHandling,
    browser.FollowRedirects:     surf.DefaultFollowRedirects,
})
```

The attributes can also be set globally. Now every new browser you create
will be set with these defaults.
```go
surf.DefaultSendReferer = false
surf.DefaultMetaRefreshHandling = false
surf.DefaultFollowRedirects = false
```

# Storage Jars
Override the build in cookie jar. Surf uses cookiejar.Jar by default.
```go
bow := surf.NewBrowser()
bow.SetCookieJar(jar.NewMemoryCookies())
```

Override the build in bookmarks jar. Surf uses jar.MemoryBookmarks by default.
```go
bow := surf.NewBrowser()
bow.SetBookmarksJar(jar.NewMemoryBookmarks())
```

Use jar.FileBookmarks to read and write your bookmarks to a JSON file.
```go
bookmarks, err = jar.NewFileBookmarks("/home/joe/bookmarks.json")
if err != nil { panic(err) }
bow := surf.NewBrowser()
bow.SetBookmarksJar(bookmarks)
```

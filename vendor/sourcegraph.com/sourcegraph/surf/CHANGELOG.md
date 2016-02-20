#### v0.5.5 - 2014/05/24
* Added Browser.Head() method. [#24](https://github.com/headzoo/surf/pull/24)


#### v0.5.4 - 2015/04/29
* Added Browser.SetTransport() method. [#15](https://github.com/headzoo/surf/issues/15)


#### v0.5.3 - 2015/04/28
* Added SURF_DEBUG_HEADERS environment variable. [#20](https://github.com/headzoo/surf/pull/20)
* Fixed bug where request headers are being added instead of set. [#19](https://github.com/headzoo/surf/pull/19)
* Fix for redirects. [#18](https://github.com/headzoo/surf/pull/18)
* Added support for textareas. [#16](https://github.com/headzoo/surf/pull/16)


#### v0.5.2 - 2015/03/06
* Allow calling Post without first opening a page. [#14](https://github.com/headzoo/surf/issues/14)
* Browser.Download() writes the raw body instead of the parsed DOM. [#13](https://github.com/headzoo/surf/issues/13)


#### v0.5.1 - 2015/01/29
* Doc updates.
* Lint fixes.


#### v0.5.0.1 - 2015/01/29
* Refactoring of documentation.
* Added .travis.yml for continuous integration.


#### v0.5 - 2015/01/29
* Extended osName and osVersion to work on various systems. [#7](https://github.com/headzoo/surf/pull/7)
* Added Browser.DelRequestHeader() method. [#11](https://github.com/headzoo/surf/pull/11)
* Setting the Referrer header to the current URL in Browser.Post(). [#10](https://github.com/headzoo/surf/pull/10)


#### v0.4.9 - 2014/09/18
* Added Browser.PostMultipart() method.
* Internal changes when building a request to ensure Content-Length is set.


#### v0.4.8 - 2014/08/28
* surf.NewBrowser() no longer returns an error.
* Added jar.NewMemoryCookies() method.
* Added jar.NewMemoryHeaders() method.
* Moved default attributes into surf package.


#### v0.4.7 - 2014/08/28
* Created type Stylesheet.
* Created type Script.
* Normalized asset URLs (src, href) to always use URL.
* Added DownloadAsync() methods.


#### v0.4.6 - 2014/08/27
* Broke up the packages to be more organized.
* Created jar.FileBookmarks for saving bookmarks to a file.
* Removed unittest package. Now at github.com/headzoo/ut.


#### v0.4.5 - 2014/08/27
* Created type Downloadable.
* Renamed Browser.Write() to Browser.Download().
* Created type Image.
* Added Browser.Images() method.
* Removed generated docs. API docs are viewable from godoc.org.


#### v0.4.4 - 2014/08/25
* Renamed Browser.FollowLink() to Click().
* Renamed Browser.Get() to Open().
* Renamed Browser.GetForm() to OpenForm().
* Renamed Browser.GetBookmark() to OpenBookmark().
* Moved attributes to their own package.
* Renamed Query() methods to Dom().
* Renamed jars package to jar.
* Added the Browser.Write() method.

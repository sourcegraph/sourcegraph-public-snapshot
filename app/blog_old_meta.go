package app

import "encoding/json"

var blogOldMetaData = blogParseOldMetaData()

// blogParseOldMetaData parses the old JSON meta-data and returns a map of
// slug-to-meta. It panics on any error as the JSON input must always be valid
// input (and it's parsed upon initialization, see above).
func blogParseOldMetaData() map[string]*blogPostMeta {
	var data []*blogPostMeta
	err := json.Unmarshal(blogOldMetaDataBytes, &data)
	if err != nil {
		panic(err)
	}
	dataMap := make(map[string]*blogPostMeta, len(data))
	for _, m := range data {
		dataMap[m.URL] = m
	}
	return dataMap
}

// JSON meta-data from the old Markdown blog posts. Effectively we parse this
// meta-data and prefer it over the meta-data we get from Tumblr (e.g. the date
// for any Tumblr post will not be valid for these old ones all uploaded on
// April 21).
var blogOldMetaDataBytes = []byte(`[
  {
    "slug": "most-popular-django-model-field",
    "title": "What's the most popular kind of Django DB model field?",
    "author": {
      "login": "sqs",
      "name": "Quinn Slack"
    },
    "thumbnail": "https://s3-us-west-2.amazonaws.com/sourcegraph-assets/django128.png",
    "date": "2013-10-16T22:00:44Z"
  },{
    "slug": "chrome-extension-for-github-repo-most-used-functions",
    "title": "Sourcegraph Chrome extension: Function listings (with docs) for GitHub repositories",
    "author": {
      "login": "sqs",
      "name": "Quinn Slack"
    },
    "thumbnail": "https://s3-us-west-2.amazonaws.com/sourcegraph-assets/chromium.png",
    "relatedPostSlugs": ["chrome-extension-annotate"],
    "date": "2013-10-17T23:17:23Z"
  },{
    "slug": "2013-october-25-31-changelog",
    "title": "Sourcegraph CHANGELOG (Oct 25-31, 2013)",
    "author": {
      "login": "sqs",
      "name": "Quinn Slack"
    },
    "date": "2013-10-31T22:27:35Z"
  },{
    "slug": "mapping-the-python-universe-introduction",
    "title": "How Sourcegraph maps the Python universe: Introduction",
    "author": {
      "login": "beyang",
      "name": "Beyang Liu"
    },
    "thumbnail": "https://s3-us-west-2.amazonaws.com/sourcegraph-assets/python.png",
    "relatedPostSlugs": ["python-static-analysis"],
    "date": "2013-11-04T17:38:54Z"
  },{
    "slug": "go-interfaces-and-implementations",
    "title": "Finding all types that implement a Go interface, globally",
    "author": {
      "login": "sqs",
      "name": "Quinn Slack"
    },
    "thumbnail": "https://s3-us-west-2.amazonaws.com/sourcegraph-assets/golang.png",
    "date": "2013-11-07T06:19:10Z"
  },{
    "slug": "new-repository-page",
    "title": "Announcing the new repository page",
    "author": {
      "login": "sqs",
      "name": "Quinn Slack"
    },
    "date": "2013-12-09T09:05:03Z"
  },{
    "slug": "python-static-analysis",
    "title": "What makes Python static analysis hard and interesting",
    "author": {
      "login": "yin",
      "name": "Yin Wang"
    },
    "thumbnail": "https://s3-us-west-2.amazonaws.com/sourcegraph-assets/python.png",
    "relatedPostSlugs": ["mapping-the-python-universe-introduction"],
    "date": "2013-12-09T11:11:03Z"
  },{
    "slug": "example-filter-and-sidebar",
    "title": "Example filtering and better browsability",
    "author": {
      "login": "beyang",
      "name": "Beyang Liu"
    },
    "date": "2013-12-10T11:32:03Z"
  },{
    "slug": "open-source-fellowship",
    "title": "Announcing the Sourcegraph Open Source Fellowship: $1000 stipend and mentorship for students who get involved in open source",
    "author": {
      "login": "sqs",
      "name": "Quinn Slack"
    },
    "date": "2013-12-12T16:40:31Z"
  },{
    "slug": "sourcegraph-site-documentation",
    "title": "Sourcegraph, now with site documentation",
    "author": {
      "login": "sqs",
      "name": "Quinn Slack"
    },
    "date": "2013-12-13T03:19:47Z"
  },{
    "slug": "repository-and-user-search",
    "title": "More things to search for",
    "author": {
      "login": "beyang",
      "name": "Beyang Liu"
    },
    "date": "2013-12-19T20:19:47Z"
  },{
    "slug": "sosf-1-kickoff",
    "title": "Kicking off the 1st Sourcegraph Open Source Fellowship: projects and bios",
    "author": {
      "login": "sqs",
      "name": "Quinn Slack"
    },
    "date": "2014-01-28T18:31:51Z"
  },{
    "slug": "fosdem-2014-go-devroom",
    "title": "FOSDEM 2014 Go devroom project roundup",
    "author": {
      "login": "sqs",
      "name": "Quinn Slack"
    },
    "thumbnail": "https://s3-us-west-2.amazonaws.com/sourcegraph-assets/golang.png",
    "relatedPostSlugs": ["go-interfaces-and-implementations"],
    "date": "2014-02-02T22:00:44Z"
  },{
    "slug": "chrome-extension-annotate",
    "title": "Chrome extension update: GitHub code annotations",
    "author": {
      "login": "beyang",
      "name": "Beyang Liu"
    },
    "thumbnail": "https://s3-us-west-2.amazonaws.com/sourcegraph-assets/chromium.png",
    "relatedPostSlugs": ["chrome-extension-for-github-repo-most-used-functions"],
    "date": "2014-02-05T19:19:47Z"
  },{
    "slug": "switching-from-angularjs-to-server-side-html",
    "title": "5 surprisingly painful things about client-side JS",
    "author": {
      "login": "sqs",
      "name": "Quinn Slack"
    },
    "thumbnail": "https://s3-us-west-2.amazonaws.com/sourcegraph-assets/javascript.png",
    "date": "2014-02-17T20:43:31Z"
  },{
    "slug": "rubysonar",
    "title": "Ruby code search and browsing with RubySonar",
    "author": {
      "login": "yinwang0",
      "name": "Yin Wang"
    },
    "thumbnail": "https://s3-us-west-2.amazonaws.com/sourcegraph-assets/ruby.png",
    "date": "2014-01-29T18:31:51Z"
  },{
    "slug": "two-new-repository-badges",
    "title": "Two new repository badges: docs & examples and dependencies",
    "author": {
      "login": "sqs",
      "name": "Quinn Slack"
    },
    "date": "2014-02-24T08:40:25Z"
  },{
    "slug": "gophercon2014-liveblog",
    "title": "GopherCon 2014 liveblog",
    "author": {
      "login": "sqs",
      "name": "Quinn Slack"
    },
    "thumbnail": "https://s3-us-west-2.amazonaws.com/sourcegraph-assets/golang.png",
    "date": "2014-04-24T17:00:44Z"
  },{
    "slug": "codeinsider-interview",
    "title": "Code Insider interview with Sourcegraph",
    "author": {
      "login": "sqs",
      "name": "Quinn Slack"
    },
    "date": "2014-06-10T21:04:44Z"
  },{
    "slug": "new-sourcegraphers-charles-varun",
    "title": "Two new members of our team: Charles Vickery and Varun Ramesh",
    "author": {
      "login": "sqs",
      "name": "Quinn Slack"
    },
    "date": "2014-07-11T08:20:23Z"
  },{
    "slug": "google-io-2014-building-sourcegraph-a-large-scale-code-search-engine-in-go",
    "title": "Google I/O 2014: Building Sourcegraph, a large-scale code search engine in Go",
    "author": {
      "login": "sqs",
      "name": "Quinn Slack"
    },
    "date": "2014-07-14T23:30:32Z"
  },{
    "slug": "announcing-ruby-and-ruby-on-rails-beta",
    "title": "Announcing support for Ruby & Ruby on Rails (beta)",
    "author": {
      "login": "sqs",
      "name": "Quinn Slack"
    },
    "thumbnail": "https://s3-us-west-2.amazonaws.com/sourcegraph-assets/ruby.png",
    "date": "2014-07-17T21:30:25Z"
  },{
    "slug": "fireside-chat-with-jiahua-chen-creator-of-gogs",
    "title": "Fireside chat with Jiahua Chen, creator of Gogs (Go Git Service) and Macaron",
    "author": {
      "login": "sqs",
      "name": "Quinn Slack"
    },
    "date": "2014-07-21T23:09:21Z"
  },{
    "slug": "linking-to-functions-on-sourcegraph",
    "title": "A URL for every function in the world",
    "author": {
      "login": "sqs",
      "name": "Quinn Slack"
    },
    "date": "2014-07-17T16:15:21Z"
  },{
    "slug": "andrey_petrov_how_to_make_your_open_source_project_thrive",
    "title": "How to make your open-source project thrive, with Andrey Petrov",
    "author": {
      "login": "charlesvickery",
      "name": "Charles Vickery"
    },
    "date": "2015-03-11T16:00:31Z"
  },{
    "slug": "building-a-product-one-interview-at-a-time",
    "title": "Building a product, one interview at a time",
    "author": {
      "login": "charlesvickery",
      "name": "Charles Vickery"
    },
    "date": "2014-09-18T00:59:23Z"
  },{
    "slug": "new-and-improved-chrome-extension",
    "title": "An improved Chrome extension for browsing code on GitHub",
    "author": {
      "login": "beyang",
      "name": "Beyang Liu"
    },
    "thumbnail": "https://s3-us-west-2.amazonaws.com/sourcegraph-assets/chromium.png",
    "relatedPostSlugs": ["chrome-extension-annotate"],
    "date": "2014-07-09T08:17:23Z"
  },{
    "slug": "chrome-extension-tooltips-and-seamless",
    "title": "Browse GitHub code like you're in an IDE, with the Sourcegraph Chrome extension",
    "author": {
      "login": "sqs",
      "name": "Quinn Slack"
    },
    "thumbnail": "https://s3-us-west-2.amazonaws.com/sourcegraph-assets/chromium.png",
    "date": "2014-08-05T10:30:35Z"
  },{
    "slug": "welcome-james",
    "title": "Welcome James to our team!",
    "author": {
      "login": "charlesvickery",
      "name": "Charles Vickery"
    },
    "thumbnail": "https://s3-us-west-2.amazonaws.com/sourcegraph-assets/cuddy_thumb.jpg",
    "date": "2015-03-19T16:00:31Z"
  },{
    "slug": "code_review_processes",
    "title": "The Pain of Code Review: How Different Teams Manage, Scale, and Perform Code Reviews",
    "author": {
      "login": "charlesvickery",
      "name": "Charles Vickery"
    },
    "date": "2015-04-05T16:00:31Z"
  },{
    "slug": "fireside-chat-with-dmitri-shuralyov",
    "title": "Fireside chat with Dmitri Shuralyov, dev tools hacker",
    "author": {
      "login": "sqs",
      "name": "Quinn Slack"
    },
    "date": "2014-07-27T23:15:07Z"
  },{
    "slug": "file-browser",
    "title": "Announcing the file and directory browsing feature",
    "author": {
      "login": "yinwang0",
      "name": "Yin Wang"
    },
    "date": "2014-02-28T12:19:34Z",
    "thumbnail": "https://s3-us-west-2.amazonaws.com/sourcegraph-assets/file_browser.png"
  },{
    "slug": "welcome-gabriel",
    "title": "Welcome Gabriel to our team!",
    "author": {
      "login": "charlesvickery",
      "name": "Charles Vickery"
    },
    "thumbnail": "https://s3-us-west-2.amazonaws.com/sourcegraph-assets/gabriel_thumb.png",
    "date": "2015-02-17T16:00:31Z"
  },{
    "slug": "go-at-sourcegraph",
    "title": "Go at Sourcegraph: serving terabytes of git data, tracing app perf, and caching HTTP resources",
    "author": {
      "login": "sqs",
      "name": "Quinn Slack"
    },
    "thumbnail": "https://s3-us-west-2.amazonaws.com/sourcegraph-assets/golang.png",
    "date": "2014-12-03T22:01:05Z"
  },{
    "slug": "go-challenge-luke-champine",
    "title": "Go Challenge Winner: Luke Champine",
    "author": {
      "login": "charlesvickery",
      "name": "Charles Vickery"
    },
    "thumbnail": "https://s3-us-west-2.amazonaws.com/sourcegraph-assets/golang.png",
    "date": "2015-04-09T23:17:23Z"
  },{
    "slug": "programming-interview",
    "title": "The programming interview experiment",
    "author": {
      "login": "beyang",
      "name": "Beyang Liu"
    },
    "date": "2014-02-24T19:01:32Z",
    "thumbnail": "https://s3-us-west-2.amazonaws.com/sourcegraph-assets/blog/einstein_code.jpeg"
  },{
    "slug": "ipfs-the-permanent-web-by-juan-benet-talk",
    "title": "IPFS: The Permanent Web, by Juan Benet (Talks at Sourcegraph 003)",
    "author": {
      "login": "sqs",
      "name": "Quinn Slack"
    },
    "date": "2014-07-22T16:00:31Z"
  },{
    "slug": "go-challenge-jeremyjay",
    "title": "Go Challenge Winner: Jeremy Jay",
    "author": {
      "login": "charlesvickery",
      "name": "Charles Vickery"
    },
    "thumbnail": "https://s3-us-west-2.amazonaws.com/sourcegraph-assets/golang.png",
    "date": "2015-04-14T18:21:52Z"
  },{
    "slug": "multi-language-lexer-and-scanner-for-go",
    "title": "Multi-language lexer and scanner for Go",
    "author": {
      "login": "Southern",
      "name": "Colton Baker"
    },
    "thumbnail": "https://s3-us-west-2.amazonaws.com/sourcegraph-assets/golang.png",
    "date": "2014-07-29T20:27:53Z"
  },{
    "slug": "mandatory-vacation",
    "title": "Why vacation at tech companies should be mandatory: better code, happier people",
    "author": {
      "login": "sqs",
      "name": "Quinn Slack"
    },
    "thumbnail": "https://s3-us-west-2.amazonaws.com/sourcegraph-assets/beach-and-boat-sm.png",
    "relatedPostSlugs": ["programming-interview"],
    "date": "2014-03-03T16:33:31Z"
  },{
    "slug": "meteor-key-lessons-from-1.0",
    "title": "Key technical and community lessons from Meteor 1.0, by Emily Stark\n(Talks at Sourcegraph 006)",
    "author": {
      "login": "beyang",
      "name": "Beyang Liu"
    },
    "thumbnail": "https://s3-us-west-2.amazonaws.com/sourcegraph-assets/blog/meteor_logo.png",
    "date": "2014-12-11T16:00:31Z"
  },{
    "slug": "new-team-member",
    "title": "Welcome Samer to our team!",
    "author": {
      "login": "charlesvickery",
      "name": "Charles Vickery"
    },
    "date": "2014-10-06T17:30:00Z"
  },{
    "slug": "most-used-golang-functions",
    "title": "The top 150 most used Go functions",
    "author": {
      "login": "sqs",
      "name": "Quinn Slack"
    },
    "thumbnail": "https://s3-us-west-2.amazonaws.com/sourcegraph-assets/golang.png",
    "date": "2014-08-07T17:41:23Z"
  },{
    "slug": "open-doors-for-open-source",
    "title": "Open doors for open source",
    "author": {
      "login": "charlesvickery",
      "name": "Charles Vickery"
    },
    "date": "2014-11-25T16:00:31Z"
  },{
    "slug": "building-a-testable-webapp",
    "title": "Building a testable Go web app",
    "author": {
      "login": "beyang",
      "name": "Beyang Liu"
    },
    "thumbnail": "https://s3-us-west-2.amazonaws.com/sourcegraph-assets/golang.png",
    "date": "2014-09-22T14:00:31Z"
  },{
    "slug": "new-product-announcement",
    "title": "A faster, redesigned Sourcegraph",
    "author": {
      "login": "charlesvickery",
      "name": "Charles Vickery"
    },
    "date": "2014-07-09T08:20:23Z"
  },{
    "slug": "sandstorm-by-kenton-varda-talk",
    "title": "Sandstorm: a Personal Cloud Platform, by Kenton Varda\n(Talks at Sourcegraph 004)",
    "author": {
      "login": "beyang",
      "name": "Beyang Liu"
    },
    "thumbnail": "https://s3-us-west-2.amazonaws.com/sourcegraph-assets/blog/sandstorm.png",
    "date": "2014-08-08T16:00:31Z"
  },{
    "slug": "sourceboxes-a-better-way-to-embed-code-snippets",
    "title": "Sourceboxes: a better way to embed code snippets",
    "author": {
      "login": "sqs",
      "name": "Quinn Slack"
    },
    "date": "2014-08-05T10:20:18Z"
  },{
    "slug": "announcing-srclib",
    "title": "Announcing srclib, a polyglot code analysis library for building better developer tools",
    "author": {
      "login": "beyang",
      "name": "Beyang Liu"
    },
    "thumbnail": "https://s3-us-west-2.amazonaws.com/sourcegraph-assets/blog/srclib/srclib.png",
    "date": "2014-08-13T16:00:31Z"
  },{
    "slug": "three_challenges_dev_teams",
    "title": "The Top Three Challenges (and Solutions) of Development Teams @ Startups",
    "author": {
      "login": "charlesvickery",
      "name": "Charles Vickery"
    },
    "date": "2015-02-16T16:00:31Z"
  },{
    "slug": "throwingawesomemeetups",
    "title": "Throwing awesome meetups with a small team",
    "author": {
      "login": "charlesvickery",
      "name": "Charles Vickery"
    },
    "date": "2014-11-24T16:00:31Z"
  },{
    "slug": "gabriel-bianconi-django-google-prediction",
    "title": "Featured project: django-google-prediction by Gabriel Bianconi",
    "author": {
      "login": "sqs",
      "name": "Quinn Slack"
    },
    "thumbnail": "https://s3-us-west-2.amazonaws.com/sourcegraph-assets/django128.png",
    "date": "2014-02-27T22:06:52Z"
  }
]`)

// A map of "old" blog slugs to "new" blog slugs for redirection. These are
// mostly hacks to keep legacy URLs alive on social media sites etc.
var blogSlugRedirects = map[string]string{
	"meteor-1.0": "meteor-key-lessons-from-1.0",

	// redirect because I misspelled his name -@sqs
	"fireside-chat-with-dmitri-shuryalov": "fireside-chat-with-dmitri-shuralyov",

	"top-150-most-used-golang-functions": "most-used-golang-functions",
}

// Effectively just a map of old blog slugs to their Tumblr IDs -- these just
// keep the previous URLs for these posts identical (new blog posts should not
// be added here, as they are pulled from the Tumblr URL).
var blogSlugAliases = map[string]string{
	"meteor-key-lessons-from-1.0":                          "117140755404",
	"fireside-chat-with-dmitri-shuralyov":                  "117138112254",
	"most-used-golang-functions":                           "117138625009",
	"most-popular-django-model-field":                      "117133211274",
	"chrome-extension-for-github-repo-most-used-functions": "117133282004",
	"2013-october-25-31-changelog":                         "117133372184",
	"mapping-the-python-universe-introduction":             "117133869919",
	"go-interfaces-and-implementations":                    "117134013684",
	"new-repository-page":                                  "117134581539",
	"python-static-analysis":                               "117134721694",
	"example-filter-and-sidebar":                           "117134853654",
	"open-source-fellowship":                               "117135045304",
	"sourcegraph-site-documentation":                       "117135107394",
	"repository-and-user-search":                           "117135210689",
	"sosf-1-kickoff":                                       "117135271119",
	"fosdem-2014-go-devroom":                               "117135823754",
	"chrome-extension-annotate":                            "117135881869",
	"switching-from-angularjs-to-server-side-html":         "117135991499",
	"rubysonar":                                                                  "117135469344",
	"two-new-repository-badges":                                                  "117136056194",
	"gophercon2014-liveblog":                                                     "117136895834",
	"codeinsider-interview":                                                      "117136934579",
	"new-sourcegraphers-charles-varun":                                           "117137724524",
	"google-io-2014-building-sourcegraph-a-large-scale-code-search-engine-in-go": "117137797304",
	"announcing-ruby-and-ruby-on-rails-beta":                                     "117137912379",
	"fireside-chat-with-jiahua-chen-creator-of-gogs":                             "117137984539",
	"linking-to-functions-on-sourcegraph":                                        "117137861404",
	"andrey_petrov_how_to_make_your_open_source_project_thrive":                  "117141225759",
	"building-a-product-one-interview-at-a-time":                                 "117139178964",
	"new-and-improved-chrome-extension":                                          "117137298659",
	"chrome-extension-tooltips-and-seamless":                                     "117138423314",
	"welcome-james":                                                              "117141388099",
	"code_review_processes":                                                      "117141489014",
	"file-browser":                                                               "117136561499",
	"welcome-gabriel":                                                            "117141054854",
	"go-at-sourcegraph":                                                          "117140591794",
	"go-challenge-luke-champine":                                                 "117141723399",
	"programming-interview":                                                      "117136252984",
	"ipfs-the-permanent-web-by-juan-benet-talk":                                  "117138056989",
	"go-challenge-jeremyjay":                                                     "117141869404",
	"multi-language-lexer-and-scanner-for-go":                                    "117138261589",
	"mandatory-vacation":                                                         "117136759154",
	"new-team-member":                                                            "117139565859",
	"open-doors-for-open-source":                                                 "117140279854",
	"building-a-testable-webapp":                                                 "117139368029",
	"new-product-announcement":                                                   "117137617759",
	"sandstorm-by-kenton-varda-talk":                                             "117138868974",
	"sourceboxes-a-better-way-to-embed-code-snippets":                            "117138326949",
	"announcing-srclib":                                                          "117138961539",
	"three_challenges_dev_teams":                                                 "117140910334",
	"throwingawesomemeetups":                                                     "117140180019",
	"gabriel-bianconi-django-google-prediction":                                  "117136430659",
}

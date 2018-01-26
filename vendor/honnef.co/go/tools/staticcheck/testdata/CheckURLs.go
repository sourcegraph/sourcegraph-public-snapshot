package pkg

import "net/url"

func fn() {
	url.Parse("foobar")
	url.Parse(":") // MATCH /is not a valid URL/
	url.Parse("https://golang.org")
}

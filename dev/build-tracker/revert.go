package main

import "net/url"

type Revert struct {
	build      *Build
	approveUrl *url.URL
	rejectUrl  *url.URL
	prUrl      *url.URL
}

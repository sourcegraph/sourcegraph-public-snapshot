package app

import "strings"

var gosrcBookmarklet = prepBookmarklet(`
javascript:
(function() {
	var byHost = {
		"github.com": function() {
			var path = window.location.pathname.slice(1);
			var parts = path.split("/");
			if (parts.length < 2) {
				alert("This bookmarklet may only be used on repository pages.");
				return;
			}
			window.location = "http://gosrc.org/github.com/" + parts.slice(0, 2).join("/");
		},
		"sourcegraph.com": function() {
			var path = window.location.pathname.slice(1);
			var parts = path.split("/");
			if (parts.length < 3) {
				alert("This bookmarklet may only be used on repository pages.");
				return;
			}
			window.location = "http://gosrc.org/" + parts.slice(0, 3).join("/");
		},
		"godoc.org": function() {
			window.location = "http://gosrc.org" + window.location.pathname;
		},
	};
	if (byHost[window.location.hostname]) {
		byHost[window.location.hostname]();
	} else {
		alert("This bookmarklet may only be used on the following sites: " + Object.keys(byHost).join(", "));
	}
})();
`)

func prepBookmarklet(s string) string {
	return strings.Replace(strings.Replace(s, "\n", "", -1), "\t", "", -1)
}

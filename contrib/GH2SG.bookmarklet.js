javascript:
(function(){
    // This bookmarklet jumps from GitHub.com to Sourcegraph.com.
    if (window.location.hostname !== "github.com" && window.location.hostname !== "sourcegraph.com") {
        alert("This bookmarklet may only be used on GitHub.com or Sourcegraph.com, not " + window.location.hostname + ".");
        return;
    }
    var pats = [
      ["^/([^/]+)/([^/]+)/commits/([^/]+)$", "/github.com/$1/$2@$3/.commits", "^/github\.com/([^/]+)/([^/]+)@([^/]+)/.commits$", "/$1/$2/commits/$3"],
      ["^/([^/]+)/([^/]+)/branches$", "/github.com/$1/$2/.branches", "^/github\.com/([^/]+)/([^/]+)/.branches$", "/$1/$2/branches"],
      ["^/([^/]+)/([^/]+)/tree/([^/]+)$", "/github.com/$1/$2@$3", "^/github\.com/([^/]+)/([^/@]+)@([^/]+)$", "/$1/$2/tree/$3"],
      ["^/([^/]+)/([^/]+)/tree/([^/]+)/(.+)$", "/github.com/$1/$2@$3/.tree/$4", "^/github\.com/([^/]+)/([^/@]+)@([^/]+)/.tree/(.+)$", "/$1/$2/tree/$3/$4"],
      ["^/([^/]+)/([^/]+)/blob/([^/]+)/(.+)$", "/github.com/$1/$2@$3/.tree/$4", "", ""], // can't disambiguate between blob and tree on GitHub
      ["^/([^/]+)/([^/]+)/pull/(\\d+)", "/github.com/$1/$2/.pulls/$3", "^/github\.com/([^/]+)/([^/]+)/.pulls/(\\d+)", "/$1/$2/pull/$3"],
      ["^/([^/]+)/([^/]+)/pull/(\\d+)/(files|commits)", "/github.com/$1/$2/.pulls/$3/$4", "^/github\.com/([^/]+)/([^/]+)/.pulls/(\\d+)/(files|commits)", "/$1/$2/pull/$3/$4"],
      ["^/([^/]+)/([^/]+)/pulls", "/github.com/$1/$2/.pulls", "^/github\.com/([^/]+)/([^/]+)/.pulls", "/$1/$2/pulls"],
      ["^/([^/]+)/([^/]+)$", "/github.com/$1/$2", "^/github\.com/([^/]+)/([^/]+)$", "/$1/$2"],
      ["^/([^/]+)$", "/$1", "^/([^/]+)$", "/$1"],
    ];
    var pathname = window.location.pathname;
    if (window.location.hostname === 'sourcegraph.com') {
      if (pathname.indexOf('/sourcegraph.com/') === 0) { // normalize pathname
        pathname = pathname.replace('/sourcegraph.com/', '/github.com/');
      } else if (pathname.indexOf('/sourcegraph/') === 0) {
        pathname = '/github.com' + pathname;
      }
    }

    for (var i = 0; i < pats.length; i++) {
      var pat = pats[i];
      if (window.location.hostname === "github.com") {
        if (pat[0] === "") { continue; }

        var r = new RegExp(pat[0]);
        if (pathname.match(r)) {
          var pathname2 = pathname.replace(r, pat[1]);
          window.location = "https://sourcegraph.com" + pathname2;
          return;
        }
      } else {
        if (pat[2] === "") { continue; }

        var r = new RegExp(pat[2]);
        if (pathname.match(r)) {
          var pathname2 = pathname.replace(r, pat[3]);
          window.location = "https://github.com" + pathname2;
          return;
        }
      }
    }
    alert("Unable to jump to Sourcegraph (no matching URL pattern).");
})();

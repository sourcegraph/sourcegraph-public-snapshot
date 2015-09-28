package ui

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sourcegraph/srclib/graph"
	"src.sourcegraph.com/sourcegraph/app/router"
	"src.sourcegraph.com/sourcegraph/sourcecode"
	"src.sourcegraph.com/sourcegraph/ui/payloads"
	"src.sourcegraph.com/sourcegraph/util/handlerutil"
	"src.sourcegraph.com/sourcegraph/util/htmlutil"
	"src.sourcegraph.com/sourcegraph/util/httputil/httpctx"
)

var (
	defKeyRegexp    = regexp.MustCompile(`^/((?:(?:[^/.@][^/@]*/)+(?:[^/.@][^/@]*))|(?:R\$\d+))((?:@(?:[^/]*(?:/(?:[^/.]+[^/]*)+)*))?)/\.([^/]+)/(.*)\.def((?:(?:/(?:[^/.][^/]*/)*(?:[^/.][^/]*))|))[/]?$`)
	errNoParam      = fmt.Errorf("UI: missing '%s' URL parameter(s)", keysQueryParam)
	errInvalidParam = fmt.Errorf("UI: '%s' URL parameter(s) should contain DefKey based URLs", keysQueryParam)
)

const keysQueryParam = "key"

func serveDefList(w http.ResponseWriter, r *http.Request) error {
	var d []*payloads.DefCommon
	e := json.NewEncoder(w)
	urls := r.URL.Query()[keysQueryParam]
	if len(urls) == 0 {
		return errNoParam
	}

	opt := sourcegraph.DefListOptions{
		DefKeys:     make([]*graph.DefKey, 0, len(urls)),
		IncludeTest: true,
		Doc:         true,
	}
	dedupRepoRevs := make(map[string]struct{}) // Deduplicate repeated repo revs.
	for _, URL := range urls {
		u, err := url.QueryUnescape(URL)
		if err != nil {
			continue
		}
		p := defKeyRegexp.FindStringSubmatch(u)
		if p == nil || len(p) != 6 {
			return errInvalidParam
		}
		dedupRepoRevs[p[1]+p[2]] = struct{}{}
		opt.DefKeys = append(opt.DefKeys, &graph.DefKey{
			Repo:     p[1],
			CommitID: strings.TrimLeft(p[2], "@"),
			UnitType: p[3],
			Unit:     strings.TrimRight(p[4], "/"),
			Path:     strings.TrimLeft(p[5], "/"),
		})
	}
	for repoRev := range dedupRepoRevs {
		opt.RepoRevs = append(opt.RepoRevs, repoRev)
	}
	opt.Unit = opt.DefKeys[0].Unit
	opt.UnitType = opt.DefKeys[0].UnitType

	ctx := httpctx.FromRequest(r)
	cl := handlerutil.APIClient(r)
	defs, err := cl.Defs.List(ctx, &opt)
	if err != nil {
		return err
	}
	d = make([]*payloads.DefCommon, 0, len(defs.Defs))
	for _, def := range defs.Defs {
		qualifiedName := sourcecode.DefQualifiedNameAndType(def, "scope")
		qualifiedName = sourcecode.OverrideStyleViaRegexpFlags(qualifiedName)
		d = append(d, &payloads.DefCommon{
			Def:               def,
			URL:               router.Rel.URLToDefAtRev(def.DefKey, def.CommitID).String(),
			QualifiedName:     htmlutil.SanitizeForPB(string(qualifiedName)),
			ByteStartPosition: def.DefStart,
			ByteEndPosition:   def.DefEnd,
			Found:             true,
		})
	}

	return e.Encode(&struct{ Defs []*payloads.DefCommon }{d})
}

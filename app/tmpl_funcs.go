package app

import (
	"context"
	"encoding/json"
	"fmt"
	htmpl "html/template"
	"log"
	"strings"

	opentracing "github.com/opentracing/opentracing-go"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/app/appconf"
	"sourcegraph.com/sourcegraph/sourcegraph/app/assets"
	"sourcegraph.com/sourcegraph/sourcegraph/app/internal/tmpl"
	"sourcegraph.com/sourcegraph/sourcegraph/app/router"
	"sourcegraph.com/sourcegraph/sourcegraph/cli/buildvar"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/envutil"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/traceutil"
	"sourcegraph.com/sourcegraph/srclib/graph"
)

func init() {
	for name, fn := range tmplFuncs {
		if _, present := tmpl.FuncMap[name]; present {
			panic("template func already exists: " + name)
		}
		tmpl.FuncMap[name] = fn
	}
}

var tmplFuncs = htmpl.FuncMap{
	"appconf": func() interface{} { return &appconf.Flags },

	"json": func(v interface{}) (string, error) {
		b, err := json.Marshal(v)
		if err != nil {
			return "", err
		}
		return string(b), nil
	},

	"customFeedbackForm": func() htmpl.HTML { return appconf.Flags.CustomFeedbackForm },

	"maxLen": func(maxLen int, s string) string {
		if len(s) <= maxLen {
			return s
		}
		return s[:maxLen]
	},

	"assetURL":    assets.URL,
	"absAssetURL": assets.AbsURL,

	"googleAnalyticsTrackingID": func() string { return appconf.Flags.GoogleAnalyticsTrackingID },

	"deployedGitCommitID": func() string { return envutil.GitCommitID },
	"fileSearchDisabled":  func() bool { return appconf.Flags.DisableSearch },

	"publicRavenDSN": func() string { return conf.PublicRavenDSN },

	"urlToTrace": func(ctx context.Context) string {
		if span := opentracing.SpanFromContext(ctx); span != nil {
			return traceutil.SpanURL(span)
		}
		return ""
	},

	"buildvar": func() buildvar.Vars { return buildvar.All },

	"dangerouslySetHTML": func(s string) htmpl.HTML { return htmpl.HTML(s) },

	"renderSnippet": renderSnippet,

	"numberedNoun": func(count int32, word string) string {
		if strings.HasSuffix(word, "y") {
			if count > 1 || count == 0 {
				return word[:len(word)-1] + "ies"
			}
			return word
		}
		return word
	},

	"defToBlobURL": func(def graph.Def) string {
		return router.Rel.URLToBlob(def.Repo, def.CommitID, def.File, 0).String()
	},
}

type Snippet struct {
	StartByte   int64
	Code        string
	Annotations *sourcegraph.AnnotationList
}

func renderSnippet(s *Snippet) htmpl.HTML {
	var toks []string

	var clsAnns, urlAnns []*sourcegraph.Annotation
	for _, ann := range s.Annotations.Annotations {
		if ann.Class != "" {
			clsAnns = append(clsAnns, ann)
		} else if ann.URL != "" {
			urlAnns = append(urlAnns, ann)
		}
	}

	log.Printf("# rendering snippet: %+v", s)
	for _, ann := range clsAnns {
		log.Printf("# ann: %+v", ann)
	}

	var prevEnd int64 = 0
	for _, ann := range clsAnns {
		start, end := int64(ann.StartByte), int64(ann.EndByte)
		if start < 0 || end > int64(len(s.Code)) {
			continue
		}

		if start > prevEnd {
			toks = append(toks, htmpl.HTMLEscapeString(s.Code[prevEnd:start]))
		}
		toks = append(toks, fmt.Sprintf("<span class=%s>", ann.Class), htmpl.HTMLEscapeString(s.Code[start:end]), "</span>")
		prevEnd = int64(ann.EndByte)
	}
	toks = append(toks, htmpl.HTMLEscapeString(s.Code[prevEnd:]))

	// TODO: add in URL annotations

	//////////////////////////////////////////

	/*
		var prevAnn *sourcegraph.Annotation
		for _, ann := range s.Annotations.Annotations {
			var prevEnd int64 = 0
			if prevAnn != nil {
				prevEnd = int64(prevAnn.EndByte) - s.StartByte
			}
			start, end := int64(ann.StartByte)-s.StartByte, int64(ann.EndByte)-s.StartByte

			if start < 0 || end > int64(len(s.Code)) {
				continue
			}

			if start > prevEnd {
				toks = append(toks, htmpl.HTMLEscapeString(s.Code[prevEnd:start]))
			}
			// log.Printf("# %+v", ann)
			toks = append(toks, fmt.Sprintf("<span class=%s>", ann.Class), htmpl.HTMLEscapeString(s.Code[start:end]), "</span>")
			prevAnn = ann
		}
		if prevAnn != nil {
			toks = append(toks, htmpl.HTMLEscapeString(s.Code[int64(prevAnn.EndByte)-s.StartByte:]))
		}
	*/

	// log.Printf("# toks: %+v", toks)

	return htmpl.HTML(strings.Join(toks, ""))
}

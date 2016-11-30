package app

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/sourcegraph/sitemap"

	"sourcegraph.com/sourcegraph/sourcegraph/app/router"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/handlerutil"
)

func serveSitemapIndex(w http.ResponseWriter, r *http.Request) error {
	var si sitemap.Index

	// TODO: remove these static sitemaps once we have proper sitemap generation! These just cover
	// def info pages.
	{
		lastMod, err := time.Parse(time.UnixDate, "Thu Sep 01 11:59:00 PST 2016")
		if err != nil {
			panic(err)
		}
		indices := []int{0, 1, 2, 3, 4, 5, 6}
		for _, index := range indices {
			si.Sitemaps = append(si.Sitemaps, sitemap.Sitemap{
				Loc:     fmt.Sprintf("https://storage.googleapis.com/static-sitemaps/sitemap_%d.xml.gz", index),
				LastMod: &lastMod,
			})
		}

		// include sitemap for top 1k repo landing pages
		si.Sitemaps = append(si.Sitemaps, sitemap.Sitemap{
			Loc:     "https://storage.googleapis.com/static-sitemaps/sitemap_repo_top1k.xml.gz",
			LastMod: &lastMod,
		})

		// include sitemap for top 4k def landing pages
		si.Sitemaps = append(si.Sitemaps, sitemap.Sitemap{
			Loc:     "https://storage.googleapis.com/static-sitemaps/sitemap_def_top4k.xml.gz",
			LastMod: &lastMod,
		})
	}

	// Truncate to sitemaps limit.
	if len(si.Sitemaps) > sitemap.MaxSitemaps {
		si.Sitemaps = si.Sitemaps[:sitemap.MaxSitemaps]
	}

	siXML, err := sitemap.MarshalIndex(&si)
	if err != nil {
		return err
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Cache-Control", "private, max-age=900")
	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	w.Write(siXML)
	return nil
}

func serveRepoSitemap(w http.ResponseWriter, r *http.Request) error {
	rc, _, err := handlerutil.GetRepoAndRevCommon(r.Context(), mux.Vars(r))
	if err != nil {
		return err
	}

	var sm sitemap.URLSet

	const (
		repoPriority  = 0.9
		defHiPriority = 0.7
		defLoPriority = 0.3
	)

	// Default change freq is Weekly, but if the repo was rebuilt in
	// the last week, use Daily.
	chgFreq := sitemap.Weekly

	// Add repo main page.
	sm.URLs = append(sm.URLs, sitemap.URL{
		Loc:        conf.AppURL.ResolveReference(router.Rel.URLToRepo(rc.Repo.URI)).String(),
		ChangeFreq: chgFreq,
		Priority:   repoPriority,
	})

	// Truncate to sitemaps limit.
	if len(sm.URLs) > sitemap.MaxURLs {
		sm.URLs = sm.URLs[:sitemap.MaxURLs]
	}

	sitemapXML, err := sitemap.Marshal(&sm)
	if err != nil {
		return err
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Cache-Control", "private, max-age=900")
	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	w.Write(sitemapXML)
	return nil
}

package app

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/sourcegraph/sitemap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/services/backend"

	"sourcegraph.com/sourcegraph/sourcegraph/app/router"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/handlerutil"
	"sourcegraph.com/sourcegraph/srclib/graph"
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
	rc, vc, err := handlerutil.GetRepoAndRevCommon(r.Context(), mux.Vars(r))
	if err != nil {
		return err
	}

	var sm sitemap.URLSet

	// TODO(sqs): add back the last-modified date (not sure how best
	// to determine it).
	var lastMod *time.Time

	const (
		repoPriority  = 0.9
		defHiPriority = 0.7
		defLoPriority = 0.3
	)

	// Default change freq is Weekly, but if the repo was rebuilt in
	// the last week, use Daily.
	chgFreq := sitemap.Weekly
	if lastMod != nil && time.Since(*lastMod) < time.Hour*24*7 {
		chgFreq = sitemap.Daily
	}

	// Add repo main page.
	sm.URLs = append(sm.URLs, sitemap.URL{
		Loc:        conf.AppURL.ResolveReference(router.Rel.URLToRepo(rc.Repo.URI)).String(),
		ChangeFreq: chgFreq,
		Priority:   repoPriority,
	})

	// Add defs if there is a valid srclib version.
	dataVer, err := backend.Repos.GetSrclibDataVersionForPath(r.Context(), &sourcegraph.TreeEntrySpec{RepoRev: vc.RepoRevSpec})
	if err != nil && grpc.Code(err) != codes.NotFound {
		return err
	}
	if dataVer != nil {
		seenDefs := map[graph.DefKey]bool{}
		defs, err := backend.Defs.List(r.Context(), &sourcegraph.DefListOptions{
			RepoRevs:    []string{rc.Repo.URI + "@" + dataVer.CommitID},
			Exported:    true,
			IncludeTest: false,
			Doc:         true,
			ListOptions: sourcegraph.ListOptions{PerPage: sitemap.MaxURLs - 1},
		})
		if err != nil {
			return err
		}

		for _, def := range defs.Defs {
			// TODO(sqs): when can defs be repeated? probably only
			// when srclib store pagination is faulty.
			if seenDefs[def.DefKey] {
				continue
			}
			seenDefs[def.DefKey] = true

			if !def.Exported {
				continue
			}

			var pri float64
			if len(def.Docs) == 0 {
				pri = defLoPriority
			} else {
				pri = defHiPriority
			}

			url := conf.AppURL.ResolveReference(router.Rel.URLToDefKey(def.DefKey)).String()
			if len(url) > 1000 {
				// Google rejects long URLs >2000 chars, but let's limit
				// them to 1000 just to be safe/sane.
				continue
			}

			sm.URLs = append(sm.URLs, sitemap.URL{
				Loc:        url,
				LastMod:    lastMod,
				ChangeFreq: chgFreq,
				Priority:   pri,
			})
		}
	}

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

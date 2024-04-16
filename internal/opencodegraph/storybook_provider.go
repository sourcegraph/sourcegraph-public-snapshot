package opencodegraph

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/grafana/regexp"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
	"golang.org/x/net/context/ctxhttp"
)

func init() {
	RegisterProvider(storybookProvider{})
}

type storybookProvider struct{}

func (storybookProvider) Name() string { return "storybook" }

func (storybookProvider) Capabilities(ctx context.Context, params schema.CapabilitiesParams) (*schema.CapabilitiesResult, error) {
	return &schema.CapabilitiesResult{
		Selector: []*schema.Selector{
			{Path: "**/*.story.(t|j)s?(x)"},
			{Path: "**/*.(t|j)s(x)", ContentContains: "React"},
		},
	}, nil
}

func (storybookProvider) Annotations(ctx context.Context, params schema.AnnotationsParams) (*schema.AnnotationsResult, error) {
	var result schema.AnnotationsResult

	if strings.HasSuffix(params.File, ".story.tsx") {
		if component := getStoryTitle(params.Content); component != "" {
			stories, ranges := firstSubmatchNamesAndRanges(exportedStory, params.Content)
			for i, story := range stories {
				id := fmt.Sprintf("%s:%d", story, i)
				story = getStoryNameAlias(story, params.Content)
				storyURL := chromaticStoryURL(component, story)
				item := &schema.OpenCodeGraphItem{
					Id:         id,
					Title:      "🖼️ Storybook: " + component + "/" + story,
					Url:        storyURL,
					Preview:    true,
					PreviewUrl: chromaticIframeURL(component, story),
				}

				info, err := getEmbedInfoForChromaticStorybook(ctx, storyURL)
				if err != nil {
					return nil, err
				}
				item.Image = info.image

				result.Items = append(result.Items, item)
				result.Annotations = append(result.Annotations, &schema.OpenCodeGraphAnnotation{
					Item:  schema.OpenCodeGraphItemRef{Id: id},
					Range: ranges[i],
				})
			}
		}
	} else {
		names, ranges := firstSubmatchNamesAndRanges(exportedReactComponentName, params.Content)
		for i, name := range names {
			id := fmt.Sprintf("%s:%d", name, i)
			component := getStoryComponentTitleForReactComponent(params.File, name)
			if component == "" {
				continue
			}

			const story = "Default"
			storyURL := chromaticStoryURL(component, story)
			item := &schema.OpenCodeGraphItem{
				Id:         id,
				Title:      "🖼️ Storybook: " + component,
				Url:        storyURL,
				Preview:    true,
				PreviewUrl: chromaticIframeURL(component, story),
			}

			info, err := getEmbedInfoForChromaticStorybook(ctx, storyURL)
			if err != nil {
				return nil, err
			}
			item.Image = info.image

			result.Items = append(result.Items, item)
			result.Annotations = append(result.Annotations, &schema.OpenCodeGraphAnnotation{
				Item:  schema.OpenCodeGraphItemRef{Id: id},
				Range: ranges[i],
			})
		}
	}

	return &result, nil
}

var storyTitle = regexp.MustCompile(`\btitle: '([^']+)'`)

func getStoryTitle(content string) string {
	m := storyTitle.FindStringSubmatch(content)
	if m == nil {
		return ""
	}
	return string(m[1])
}

func getStoryNameAlias(story string, content string) string {
	// Look for `PlainRequest.storyName = 'plain request'` or similar.
	storyNameAlias := regexp.MustCompile(story + `\.storyName = '(\w+)'`)
	m := storyNameAlias.FindStringSubmatch(content)
	if m != nil {
		story = string(m[1])
	}
	return story
}

var (
	exportedStory              = regexp.MustCompile(`export const (\w+): Story`)
	exportedReactComponentName = regexp.MustCompile(`export const ([A-Z]\w+): React\.`)
)

func getStoryComponentTitleForReactComponent(path, reactComponentName string) string {
	_ = path
	_ = reactComponentName
	m := map[string]string{
		// TODO(sqs): un-hardcode for sourcegraph
		"SignInPage": "web/auth/SignInPage",
	}
	return m[reactComponentName]
}

func chromaticStorySlug(component, story string) string {
	return strings.ToLower(strings.Replace(component, "/", "-", -1)) + "--" + kebabCase(story)
}

func chromaticStoryURL(component, story string) string {
	return (&url.URL{
		Scheme: "https",
		// TODO(sqs): un-hardcode for sourcegraph
		Host:     "5f0f381c0e50750022dc6bf7-qjtkjsausw.chromatic.com",
		Path:     "/",
		RawQuery: (url.Values{"path": []string{"/story/" + chromaticStorySlug(component, story)}}).Encode(),
	}).String()
}

func chromaticIframeURL(component, story string) string {
	return (&url.URL{
		Scheme: "https",
		// TODO(sqs): un-hardcode for sourcegraph
		Host: "5f0f381c0e50750022dc6bf7-qjtkjsausw.chromatic.com",
		Path: "/iframe.html",
		RawQuery: (url.Values{
			"id":          []string{chromaticStorySlug(component, story)},
			"singleStory": []string{"true"},
			"controls":    []string{"false"},
			"embed":       []string{"true"},
			"viewMode":    []string{"story"},
		}).Encode(),
	}).String()
}

type chromaticEmbedInfo struct {
	image *schema.OpenCodeGraphImage
}

func getEmbedInfoForChromaticStorybook(ctx context.Context, chromaticStoryURL string) (*chromaticEmbedInfo, error) {
	oembedURL := &url.URL{
		Scheme: "https",
		Host:   "www.chromatic.com",
		Path:   "/oembed",
		RawQuery: (url.Values{
			"url":    []string{chromaticStoryURL},
			"format": []string{"json"},
		}).Encode(),
	}
	resp, err := ctxhttp.Get(ctx, http.DefaultClient, oembedURL.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("chromatic oembed endpoint %s returned HTTP %d", oembedURL, resp.StatusCode)
	}

	var oembedData struct {
		Title           string `json:"title"`
		ThumbnailURL    string `json:"thumbnail_url,omitempty"`
		ThumbnailWidth  int    `json:"thumbnail_width,omitempty"`
		ThumbnailHeight int    `json:"thumbnail_height,omitempty"`
		HTML            string `json:"html,omitempty"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&oembedData); err != nil {
		return nil, err
	}

	var info chromaticEmbedInfo
	if oembedData.ThumbnailURL != "" {
		info.image = &schema.OpenCodeGraphImage{
			Url:    oembedData.ThumbnailURL,
			Width:  float64(oembedData.ThumbnailWidth),
			Height: float64(oembedData.ThumbnailHeight),
			Alt:    oembedData.Title,
		}
	}
	return &info, nil
}

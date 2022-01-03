package graphqlbackend

import (
	"context"
	"encoding/json"
	"github.com/inconshreveable/log15"
	"github.com/sourcegraph/sourcegraph/internal/rcache"
	"golang.org/x/net/html"
	"io"
	"net/http"
	"strings"
	"sync"
)

type LinkMetadataResolver struct {
	once sync.Once
	URL  string

	// Cache results because they are used by multiple fields
	html string
}

type linkMetadataArgs struct {
	URL string
}

type LinkMetadata struct {
	title       *string
	description *string
	imageUrl    *string
}

// redisCache is an HTTP cache backed by Redis. The TTL of a month is a balance between caching values
// for a useful amount of time versus growing the cache too large or having them get stale.
var redisCache = rcache.NewWithTTL("link_metadata", 60*60*24*7)

func (r *schemaResolver) LinkMetadata(ctx context.Context, args *linkMetadataArgs) *LinkMetadataResolver {
	return &LinkMetadataResolver{sync.Once{}, args.URL, ""}
}

func (r *LinkMetadataResolver) getMetadataWithCaching(url string) LinkMetadata {
	if linkMetadata, ok := getMetadataFromCache(url); ok {
		return linkMetadata
	}
	htmlSource := r.retrieveHtml(url)
	linkMetadata := parseBody(htmlSource)
	if linkMetadataJsonBytes, err := linkMetadata.toJSON(); err == nil {
		log15.Debug("Saving to cache: ", "url", url)
		redisCache.Set(url, linkMetadataJsonBytes)
	} else {
		log15.Warn("Error marshalling link metadata.", "error", err, "linkMetadata", linkMetadata)
	}
	return linkMetadata
}

func getMetadataFromCache(url string) (LinkMetadata, bool) {
	var linkMetadata LinkMetadata
	if bytes, ok := redisCache.Get(url); ok {
		// Cache hit
		log15.Debug("Cache hit for", "url", url)
		if err := linkMetadata.fromJSON(bytes); err != nil {
			log15.Warn("Failed to unmarshal cached link metadata.", "url", url, "err", err)
			return LinkMetadata{}, false
		} else {
			return linkMetadata, true
		}
	} else {
		// Cache miss
		log15.Debug("Cache miss for", "url", url)
		return LinkMetadata{}, false
	}
}

func (m *LinkMetadata) toJSON() ([]byte, error) {
	temp := map[string]interface{}{
		"title":       m.title,
		"description": m.description,
		"imageUrl":    m.imageUrl,
	}
	return json.Marshal(temp)
}

func (m *LinkMetadata) fromJSON(jsonBytes []byte) error {
	temp := map[string]string{}
	if err := json.Unmarshal(jsonBytes, &temp); err != nil {
		return err
	} else {
		m.title = strptr(temp["title"])
		m.description = strptr(temp["description"])
		m.imageUrl = strptr(temp["imageUrl"])
		return nil
	}
}

func (r *LinkMetadataResolver) retrieveHtml(url string) string {
	log15.Debug("Getting HTML for", "url", url)
	r.once.Do(func() {
		resp, err := http.Get(url)
		if err != nil {
			return
		}
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				log15.Info("Could not load HTML content for unfurling.", "url", url)
			}
		}(resp.Body)
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return
		}

		r.html = string(body)
	})

	return r.html
}

func parseBody(body string) LinkMetadata {
	metaTags := extractAllMetaTags(body)

	title := extractField(metaTags, []string{"og:title", "twitter:title", "title"})
	description := extractField(metaTags, []string{"og:description", "twitter:description", "description"})
	imageUrl := extractField(metaTags, []string{"og:image", "twitter:image", "image"})
	return LinkMetadata{title, description, imageUrl}
}

func extractAllMetaTags(body string) map[string]string {
	tokenizer := html.NewTokenizer(strings.NewReader(body))
	metaTags := make(map[string]string)
	inHead := false
	for {
		tokenType := tokenizer.Next()
		if tokenType == html.ErrorToken {
			// Returning io.EOF indicates success.
			if tokenizer.Err() == io.EOF {
				break
			} else {
				// Swallows unexpected error
				break
			}
		} else if tokenType == html.StartTagToken || (inHead && tokenType == html.SelfClosingTagToken) {
			token := tokenizer.Token()
			if token.Data == "head" {
				inHead = true
			}
			if inHead && token.Data == "meta" {
				attributes := convertTokenToAttributeMap(token)
				if attributes["property"] != "" {
					metaTags[attributes["property"]] = attributes["content"]
				} else if attributes["name"] != "" {
					metaTags[attributes["name"]] = attributes["content"]
				}
			}
		} else if inHead && tokenType == html.EndTagToken {
			token := tokenizer.Token()
			if token.Data == "head" {
				return metaTags
			}
		}
	}
	return metaTags
}

func convertTokenToAttributeMap(token html.Token) map[string]string {
	attributeMap := map[string]string{}
	for _, attr := range token.Attr {
		attributeMap[attr.Key] = attr.Val
	}
	return attributeMap
}

func extractField(metaTags map[string]string, fieldNames []string) *string {
	for _, fieldName := range fieldNames {
		if metaTags[fieldName] != "" {
			return strptr(metaTags[fieldName])
		}
	}
	return nil
}

func (r *LinkMetadataResolver) Title() *string {
	return r.getMetadataWithCaching(r.URL).title
}

func (r *LinkMetadataResolver) Description() *string {
	return r.getMetadataWithCaching(r.URL).description
}

func (r *LinkMetadataResolver) ImageURL() *string {
	return r.getMetadataWithCaching(r.URL).imageUrl
}

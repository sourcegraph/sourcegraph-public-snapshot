package graphqlbackend

import (
	"context"
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

func (r *schemaResolver) LinkMetadata(ctx context.Context, args *linkMetadataArgs) *LinkMetadataResolver {
	return &LinkMetadataResolver{sync.Once{}, args.URL, ""}
}

func (r *LinkMetadataResolver) retrieveHtml(url string) string {
	r.once.Do(func() {
		resp, err := http.Get(url)
		if err != nil {
			return
		}
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				// TODO: Handle error
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
	htmlSource := r.retrieveHtml(r.URL)
	linkMetadata := parseBody(htmlSource)

	return linkMetadata.title
}

func (r *LinkMetadataResolver) Description() *string {
	htmlSource := r.retrieveHtml(r.URL)
	linkMetadata := parseBody(htmlSource)

	return linkMetadata.description
}

func (r *LinkMetadataResolver) ImageURL() *string {
	htmlSource := r.retrieveHtml(r.URL)
	linkMetadata := parseBody(htmlSource)

	return linkMetadata.imageUrl
}

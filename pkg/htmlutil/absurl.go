package htmlutil

import (
	"bytes"
	"io"
	"net/url"
	"strings"

	"golang.org/x/net/html"
)

func MakeURLsAbsolute(in string, absURL *url.URL) (string, error) {
	tokenizer := html.NewTokenizer(strings.NewReader(in))
	var buf bytes.Buffer
	for {
		if tokenizer.Next() == html.ErrorToken {
			err := tokenizer.Err()
			if err == io.EOF {
				break
			}
			return "", err
		}
		t := tokenizer.Token()
		for i := range t.Attr {
			switch t.Attr[i].Key {
			case "href", "src":
				if absVal, err := absURL.Parse(t.Attr[i].Val); err == nil {
					t.Attr[i].Val = absVal.String()
				}
			}
		}
		buf.WriteString(t.String())
	}
	return buf.String(), nil
}

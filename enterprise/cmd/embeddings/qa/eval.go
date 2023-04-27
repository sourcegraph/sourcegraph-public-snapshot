package qa

import (
	"bufio"
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/embeddings"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// We embed the test data because we might want to run the script independently
// of the repo on CI.
//
//go:embed context_data.tsv
var fs embed.FS

func Run(url string) error {
	if url == "" {
		return errors.New("url is empty")
	}
	c := newClient(url)

	count, recall := 0.0, 0.0

	file, err := fs.Open("context_data.tsv")
	if err != nil {
		return errors.Wrap(err, "failed to open file")
	}

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		line := scanner.Text()

		fields := strings.Split(line, "\t")
		query := fields[0]
		relevantFile := fields[1]

		args := embeddings.EmbeddingsSearchParameters{
			RepoName:         "github.com/sourcegraph/sourcegraph",
			Query:            query,
			CodeResultsCount: 20,
			TextResultsCount: 2,
			Debug:            true,
		}

		results, err := c.search(args)
		if err != nil {
			return errors.Wrap(err, "search failed")
		}

		merged := append(results.CodeResults, results.TextResults...)
		fmt.Println("Query:", query)
		fmt.Println("Results:")

		fileFound := false
		for i, result := range merged {
			if result.FileName == relevantFile {
				fmt.Printf(">> ")
				fileFound = true
			} else {
				fmt.Printf("   ")
			}
			fmt.Printf("%d. %s", i+1, result.FileName)
			if result.Debug != "" {
				fmt.Printf(" (%s)\n", result.Debug)
			} else {
				fmt.Print("\n")
			}
		}
		fmt.Println()
		if fileFound {
			recall++
		}
		count++
	}

	fmt.Println()
	fmt.Printf("Recall: %f\n", recall/count)

	return nil
}

type client struct {
	httpClient *http.Client
	url        string
}

func newClient(url string) *client {
	return &client{
		httpClient: http.DefaultClient,
		url:        url,
	}
}

func (c *client) search(args embeddings.EmbeddingsSearchParameters) (*embeddings.EmbeddingSearchResults, error) {
	b, err := json.Marshal(args)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", c.url, bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	response, err := c.httpClient.Do(req)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to send request %+v", req)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, errors.Newf("unexpected status code: %d", response.StatusCode)
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	res := embeddings.EmbeddingSearchResults{}
	err = json.Unmarshal(body, &res)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal response")
	}

	return &res, nil
}

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

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/embeddings"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// We embed the test data because we might want to run the script independently
// of the repo on CI.
//
//go:embed context_data.tsv
var fs embed.FS

type embeddingsSearcher interface {
	Search(args embeddings.EmbeddingsSearchParameters) (*embeddings.EmbeddingCombinedSearchResults, error)
}

// Run runs the evaluation and returns recall for the test data.
func Run(searcher embeddingsSearcher) (float64, error) {
	count, recall := 0.0, 0.0

	file, err := fs.Open("context_data.tsv")
	if err != nil {
		return -1, errors.Wrap(err, "failed to open file")
	}

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		line := scanner.Text()

		fields := strings.Split(line, "\t")
		query := fields[0]
		relevantFile := fields[1]

		args := embeddings.EmbeddingsSearchParameters{
			RepoNames:        []api.RepoName{"github.com/sourcegraph/sourcegraph"},
			RepoIDs:          []api.RepoID{0},
			Query:            query,
			CodeResultsCount: 20,
			TextResultsCount: 2,
		}

		results, err := searcher.Search(args)
		if err != nil {
			return -1, errors.Wrap(err, "search failed")
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
			fmt.Printf(" (%s)\n", result.ScoreDetails.String())
		}
		fmt.Println()
		if fileFound {
			recall++
		}
		count++
	}

	recall = recall / count

	fmt.Println()
	fmt.Printf("Recall: %f\n", recall)

	return recall, nil
}

type client struct {
	httpClient *http.Client
	url        string
}

func NewClient(url string) *client {
	return &client{
		httpClient: http.DefaultClient,
		url:        url,
	}
}

func (c *client) Search(args embeddings.EmbeddingsSearchParameters) (*embeddings.EmbeddingCombinedSearchResults, error) {
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

	res := embeddings.EmbeddingCombinedSearchResults{}
	err = json.Unmarshal(body, &res)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal response")
	}

	return &res, nil
}

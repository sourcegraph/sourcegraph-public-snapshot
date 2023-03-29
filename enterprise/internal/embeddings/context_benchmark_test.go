package embeddings

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// generate the token in the Sourcegraph UI: settings > access tokens
var sourcegraphToken = os.Getenv("SOURCEGRAPH_TOKEN")

func TestContextFetching(t *testing.T) {
	count, recall := 0.0, 0.0

	file, err := os.Open("testdata/context_data.tsv")
	if err != nil {
		t.Fatal(err)
	}

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		line := scanner.Text()

		fields := strings.Split(line, "\t")
		query := fields[0]
		relevantFile := fields[1]

		args := EmbeddingsSearchParameters{
			RepoName:         "github.com/sourcegraph/sourcegraph",
			Query:            query,
			CodeResultsCount: 20,
			TextResultsCount: 2,
			Debug:            true,
		}

		results, err := search(args)
		if err != nil {
			t.Fatal(err)
		}

		// List text results first so that they are always part of the results set.
		merged := append(results.TextResults, results.CodeResults...)
		fmt.Println("Query:", query)
		fmt.Println("Results:")

		for i, result := range merged {
			if result.FileName == relevantFile {
				recall++
				fmt.Print(">> ")
			} else {
				fmt.Print("   ")
			}
			fmt.Printf("%d. %s", i+1, result.FileName)
			if result.Debug != "" {
				fmt.Printf(" (%s)\n", result.Debug)
				//fmt.Printf("%s\n\n", result.Content)
			} else {
				fmt.Print("\n")
			}
		}
		fmt.Println()
		count++
	}

	fmt.Println()
	fmt.Printf("Recall: %f\n", recall/count)
}

const graphqlQuery = `
query ($repo:  ID!, $query: String!, $codeResultsCount: Int!, $textResultsCount: Int!, $debug: Boolean) {
			embeddingsSearch(repo: $repo, query: $query, codeResultsCount: $codeResultsCount, textResultsCount: $textResultsCount, debug: $debug){
							codeResults {
								fileName
								content
								debug
							}
							textResults {
								fileName
								content
								debug
							}
			}
}
	`

// maps repo names to GraphQL IDs
var toRepoID = map[string]string{
	// To get the GraphQL ID of a particular repo, run this query:
	//	query {
	//	  repository(cloneURL: "github.com/sourcegraph/sourcegraph") {
	//	    id
	//	  }
	//	}
	"github.com/sourcegraph/sourcegraph": "UmVwb3NpdG9yeTo3",
}

func search(args EmbeddingsSearchParameters) (EmbeddingSearchResults, error) {
	vars := map[string]interface{}{
		"repo":             toRepoID[string(args.RepoName)],
		"query":            args.Query,
		"codeResultsCount": args.CodeResultsCount,
		"textResultsCount": args.TextResultsCount,
		"debug":            args.Debug,
	}

	requestBody, err := json.Marshal(map[string]interface{}{
		"query":     graphqlQuery,
		"variables": vars,
	})
	if err != nil {
		return EmbeddingSearchResults{}, err
	}

	req, err := http.NewRequest("POST", "https://sourcegraph.test:3443/.api/graphql", bytes.NewReader(requestBody))
	if err != nil {
		return EmbeddingSearchResults{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("token %s", sourcegraphToken))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return EmbeddingSearchResults{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		b, _ := io.ReadAll(resp.Body)
		return EmbeddingSearchResults{}, errors.Wrap(err, string(b))
	}

	// Parse the GraphQL response and return the results
	payload := struct {
		Data struct {
			EmbeddingsSearch EmbeddingSearchResults
		}
	}{}
	err = json.NewDecoder(resp.Body).Decode(&payload)
	return payload.Data.EmbeddingsSearch, err
}

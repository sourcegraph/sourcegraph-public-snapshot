package backend

import (
	"context"
	"fmt"
	"strings"

	"github.com/sourcegraph/zoekt"

	"github.com/sourcegraph/sourcegraph/internal/completions/client"
	"github.com/sourcegraph/sourcegraph/internal/completions/types"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func rerank(pattern string, result *zoekt.SearchResult) (*zoekt.SearchResult, error) {
	completionsConfig := conf.GetCompletionsConfig(conf.Get().SiteConfig())
	if completionsConfig == nil {
		return result, errors.New("completions are not configured")
	}

	completionsClient, err := client.GetBasic(
		completionsConfig.Endpoint,
		completionsConfig.Provider,
		completionsConfig.AccessToken,
	)

	if err != nil {
		return result, err
	}

	prompt := fmt.Sprintf("Please rerank these code search results for the following search query: \"%s\"."+
		" Format the result as XML, like this: <list><item><filename>filename 1</filename>"+
		"<explanation>this is why I chose this item</explanation></item><item><filename>filename 2</filename>"+
		"<explanation>why I chose this item</explanation></item></list>"+
		"\n\n.", pattern)

	for f, file := range result.Files {
		prompt += file.FileName + "\n"

		for m, match := range file.ChunkMatches {
			prompt += string(match.Content) + "\n"
			if m > 10 {
				break
			}
		}
		prompt += "\n"

		if f > 20 {
			break
		}
	}

	params := types.CompletionRequestParameters{
		Model: completionsConfig.FastChatModel,
		Messages: []types.Message{{
			Speaker: "human",
			Text:    prompt,
		}, {
			Speaker: "assistant",
		}},
		MaxTokensToSample: 500,
		Temperature:       0.2,
		TopK:              -1,
		TopP:              -1,
	}

	resp, err := completionsClient.Complete(context.Background(), types.CompletionsFeatureChat, params)
	if err != nil {
		return result, err
	}

	fmt.Println(resp.Completion)

	rerankedResult := zoekt.SearchResult{
		Stats:         result.Stats,
		Progress:      result.Progress,
		RepoURLs:      result.RepoURLs,
		LineFragments: result.LineFragments,
	}
	fileMatches := make([]zoekt.FileMatch, 0)

	var ranked = map[string]bool{}
	next := resp.Completion
	for true {
		start := strings.Index(next, "<filename>")
		end := strings.Index(next, "</filename>")
		if start < 0 || end < 0 {
			break
		}

		filename := next[start+len("<filename>") : end]
		for _, match := range result.Files {
			if match.FileName == filename {
				fileMatches = append(fileMatches, match)
				ranked[filename] = true
			}
		}
		next = next[end+len("</filename>"):]
	}

	for _, match := range result.Files {
		if !ranked[match.FileName] {
			fileMatches = append(fileMatches, match)
		}
	}

	rerankedResult.Files = fileMatches
	return &rerankedResult, nil
}

package shared

import (
	"context"
	"strings"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/embeddings"
	"github.com/sourcegraph/sourcegraph/internal/lazyregexp"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type getContextDetectionEmbeddingIndexFn func(ctx context.Context) (*embeddings.ContextDetectionEmbeddingIndex, error)

const MIN_NO_CONTEXT_SIMILARITY_DIFF = float32(0.02)
const MIN_QUERY_WITH_CONTEXT_LENGTH = 16

var NO_CONTEXT_MESSAGES_REGEXPS = []*lazyregexp.Regexp{
	lazyregexp.New(`(previous|above)\s+(message|code|text)`),
	lazyregexp.New(
		`(translate|convert|change|for|make|refactor|rewrite|ignore|explain|fix|try|show)\s+(that|this|above|previous|it|again)`,
	),
	lazyregexp.New(
		`(this|that).*?\s+(is|seems|looks)\s+(wrong|incorrect|bad|good)`,
	),
	lazyregexp.New(`^(yes|no|correct|wrong|nope|yep|now|cool)(\s|.|,)`),
	// User provided their own code context in the form of a Markdown code block.
	lazyregexp.New("```"),
}

func isContextRequiredForChatQuery(
	ctx context.Context,
	getQueryEmbedding getQueryEmbeddingFn,
	getContextDetectionEmbeddingIndex getContextDetectionEmbeddingIndexFn,
	query string,
) (bool, error) {
	queryTrimmed := strings.TrimSpace(query)
	if len(queryTrimmed) < MIN_QUERY_WITH_CONTEXT_LENGTH {
		return false, nil
	}

	queryLower := strings.ToLower(queryTrimmed)
	for _, regexp := range NO_CONTEXT_MESSAGES_REGEXPS {
		if submatches := regexp.FindStringSubmatch(queryLower); len(submatches) > 0 {
			return false, nil
		}
	}

	isSimilarToNoContextMessages, err := isQuerySimilarToNoContextMessages(ctx, getQueryEmbedding, getContextDetectionEmbeddingIndex, queryTrimmed)
	if err != nil {
		return false, err
	}
	// If the query is similar to messages that require context, then we can assume context is required for the query.
	return !isSimilarToNoContextMessages, nil
}

func isQuerySimilarToNoContextMessages(
	ctx context.Context,
	getQueryEmbedding getQueryEmbeddingFn,
	getContextDetectionEmbeddingIndex getContextDetectionEmbeddingIndexFn,
	query string,
) (bool, error) {
	contextDetectionEmbeddingIndex, err := getContextDetectionEmbeddingIndex(ctx)
	if err != nil {
		return false, errors.Wrap(err, "getting context detection embedding index")
	}

	queryEmbedding, err := getQueryEmbedding(query)
	if err != nil {
		return false, errors.Wrap(err, "getting query embedding")
	}

	messagesWithContextSimilarity := embeddings.CosineSimilarity(contextDetectionEmbeddingIndex.MessagesWithAdditionalContextMeanEmbedding, queryEmbedding)
	messagesWithoutContextSimilarity := embeddings.CosineSimilarity(contextDetectionEmbeddingIndex.MessagesWithoutAdditionalContextMeanEmbedding, queryEmbedding)

	// We have to be really sure that the query is similar to no context messages, so we include the `MIN_NO_CONTEXT_SIMILARITY_DIFF` threshold.
	isSimilarToNoContextMessages := (messagesWithoutContextSimilarity - messagesWithContextSimilarity) >= MIN_NO_CONTEXT_SIMILARITY_DIFF
	return isSimilarToNoContextMessages, nil
}

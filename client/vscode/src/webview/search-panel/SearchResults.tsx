import React, { useCallback } from 'react'

import { StreamingSearchResultsList } from '@sourcegraph/branded/src/search/results/StreamingSearchResultsList'
import { fetchHighlightedFileLineRanges } from '@sourcegraph/shared/src/backend/file'
import { FetchFileParameters } from '@sourcegraph/shared/src/components/CodeExcerpt'
import {
    AggregateStreamingSearchResults,
    CommitMatch,
    ContentMatch,
    PathMatch,
    RepositoryMatch,
    SearchMatch,
    SymbolMatch,
} from '@sourcegraph/shared/src/search/stream'

import { CommitSearchResultFields, FileMatchFields, RepositoryFields, SearchResult } from '../../graphql-operations'
import { WebviewPageProps } from '../platform/context'

import { useQueryState } from '.'

interface SearchResultsProps extends Pick<WebviewPageProps, 'platformContext' | 'theme'> {}

// TODO(tj): just try to move the whole StreamingSearchResults file to shared and use THAT!
// Only difference is "show more" button, which we can add here
//  (refer to https://sourcegraph.com/github.com/sourcegraph/sourcegraph@v3.27.4/-/blob/client/web/src/search/results/SearchResultsList.tsx?L442)
// Also varies in that "location.search" is source of truth for search query in streaming search result.
// can change prop to 'queriedSearch'/'executedQuery', pass location.search in webapp, pass zustand value in vsce?
// also make location optional in streaming then.

// Stremaing result footer also makes no sense here, figure out a way to use the same NoResultsPage and pass it as a child to the footer
// optionally... maybe renderFooter: (children: JSX.Element) => JSX.Element

export const SearchResults = React.memo<SearchResultsProps>(({ platformContext, theme }) => {
    // TODO handler search results changing. maybe pass search results prop from parent and don't render this while loading?
    const executedQuery = useQueryState(({ state }) => state.queryToRun.query)
    const searchResults = useQueryState(({ state }) => state.searchResults)

    const fetchHighlightedFileLineRangesWithContext = useCallback(
        (parameters: FetchFileParameters) => fetchHighlightedFileLineRanges({ ...parameters, platformContext }),
        [platformContext]
    )

    // Convert GQL type to SearchMatch

    if (!searchResults) {
        // TODO loading state.. might not delegate to <StreamingSearchResultsList>
        return null
    }
    // todo memoize
    const matches = convertGQLSearchToSearchMatches(searchResults)

    // TODO error state
    const results: AggregateStreamingSearchResults = {
        state: 'complete',
        results: matches,
        filters: searchResults.search?.results.dynamicFilters ?? [],
        progress: {
            matchCount: searchResults.search?.results.matchCount ?? 0,
            durationMs: searchResults.search?.results.elapsedMilliseconds ?? 0,
            repositoriesCount: searchResults.search?.results.repositoriesCount ?? 0,
            skipped: [],
        },
    }

    return (
        <>
            <StreamingSearchResultsList
                fetchHighlightedFileLineRanges={fetchHighlightedFileLineRangesWithContext}
                isLightTheme={theme === 'theme-light'}
                executedQuery={executedQuery}
                // TODO use real settings (getSettings() on comlink extension API)
                settingsCascade={{ final: {}, subjects: [] }}
                // TODO use real telemetry service
                telemetryService={{
                    log: () => {},
                    logViewEvent: () => {},
                }}
                // Default to false until we implement <SearchResultsInfoBar>, which is where this value is set.
                allExpanded={false}
                isSourcegraphDotCom={false}
                searchContextsEnabled={true}
                showSearchContext={true}
                platformContext={platformContext}
                results={results}
            />
            {/* TODO show more button */}
        </>
    )
})

export function convertGQLSearchToSearchMatches(searchResult: SearchResult): SearchMatch[] {
    return (
        searchResult.search?.results.results.map(result => {
            switch (result.__typename) {
                case 'FileMatch': {
                    return convertFileMatch(result)
                }
                case 'CommitSearchResult':
                    return convertCommitSearchResult(result)
                case 'Repository':
                    return convertRepository(result)
            }
        }) ?? []
    )
}

function convertFileMatch(
    result: {
        __typename: 'FileMatch'
    } & FileMatchFields
): ContentMatch | SymbolMatch | PathMatch {
    if (result.symbols.length > 0) {
        const symbolMatch: SymbolMatch = {
            type: 'symbol',
            path: result.file.path,
            repository: result.repository.name,
            symbols: result.symbols.map(symbol => ({ ...symbol, containerName: symbol.containerName ?? '' })),
            repoStars: result.repository.stars,
        }
        return symbolMatch
    }
    if (result.lineMatches.length > 0) {
        const lines = result.file.content.split('\n')

        const lineMatchesWithLine = result.lineMatches.map(match => {
            const line = lines[match.lineNumber]

            return {
                line,
                ...match,
            }
        })

        const contentMatch: ContentMatch = {
            type: 'content',
            path: result.file.path, // TODO sourcegraph uri
            repository: result.repository.name,
            lineMatches: lineMatchesWithLine,
            repoStars: result.repository.stars,
        }

        return contentMatch
    }
    // TODO: Doesn't necessarily make this a patch match (just no line matches?)
    const pathMatch: PathMatch = {
        type: 'path',
        path: result.file.path,
        repository: result.repository.name,
        repoStars: result.repository.stars,
    }
    return pathMatch
}

function convertCommitSearchResult(
    result: {
        __typename: 'CommitSearchResult'
    } & CommitSearchResultFields
): CommitMatch {
    let content: string | undefined
    const ranges: number[][] = []
    for (const match of result.matches) {
        content = match.body.text
        for (const highlight of match.highlights) {
            ranges.push([highlight.line, highlight.character, highlight.length])
        }
    }

    const commitMatch: CommitMatch = {
        type: 'commit',
        label: result.label.text,
        url: result.commit.url,
        detail: result.detail.text,
        repository: result.commit.repository.name,
        repoStars: result.commit.repository.stars,
        content: content || result.commit.message,
        ranges,
    }

    return commitMatch
}

function convertRepository(
    result: {
        __typename: 'Repository'
        description: string
    } & RepositoryFields
): RepositoryMatch {
    const repositoryMatch: RepositoryMatch = {
        type: 'repo',
        repository: result.name,
        repoStars: result.stars,
        description: result.description,
    }

    return repositoryMatch
}

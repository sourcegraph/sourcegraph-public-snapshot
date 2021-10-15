import FileIcon from 'mdi-react/FileIcon'
import React, { useCallback } from 'react'

import { fetchHighlightedFileLineRanges } from '@sourcegraph/shared/src/backend/file'
import { FetchFileParameters } from '@sourcegraph/shared/src/components/CodeExcerpt'
import { FileMatch } from '@sourcegraph/shared/src/components/FileMatch'
import {
    CommitMatch,
    ContentMatch,
    PathMatch,
    RepositoryMatch,
    SearchMatch,
    SymbolMatch,
} from '@sourcegraph/shared/src/search/stream'
import { isDefined } from '@sourcegraph/shared/src/util/types'

import { CommitSearchResultFields, FileMatchFields, RepositoryFields, SearchResult } from '../../graphql-operations'
import { WebviewPageProps } from '../platform/context'

import styles from './SearchResults.module.scss'

import { useQueryState } from '.'

interface SearchResultsProps extends Pick<WebviewPageProps, 'platformContext'> {}

export const SearchResults = React.memo<SearchResultsProps>(({ platformContext }) => {
    const searchResults = useQueryState(({ state }) => state.searchResults)

    const fetchHighlightedFileLineRangesWithContext = useCallback(
        (parameters: FetchFileParameters) => fetchHighlightedFileLineRanges({ ...parameters, platformContext }),
        [platformContext]
    )

    if (!searchResults) {
        return null
    }

    // Convert GQL type to SearchMatch
    const matches = convertGQLSearchToSearchMatches(searchResults)

    const renderedMatches: JSX.Element[] | undefined = matches
        .map(match => {
            switch (match.type) {
                case 'content':
                case 'path':
                case 'symbol': {
                    const renderedFileMatch = (
                        <FileMatch
                            icon={FileIcon}
                            result={match}
                            expanded={true}
                            settingsCascade={{ final: {}, subjects: [] }}
                            showAllMatches={false}
                            telemetryService={{
                                log: () => {},
                                logViewEvent: () => {},
                            }}
                            fetchHighlightedFileLineRanges={fetchHighlightedFileLineRangesWithContext}
                            onSelect={() => console.log('on select!')}
                        />
                    )
                    return renderedFileMatch
                }
                // TODO render these
                case 'commit':
                    return null
                case 'repo':
                    return null
            }
        })
        .filter(isDefined)

    // We need <FileMatch> and <SearchResult>.
    // - <FileMatch> is already in shared
    // - move <SearchResult> to shared

    return <div className={styles.result}>{renderedMatches}</div>
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

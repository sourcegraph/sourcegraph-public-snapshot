import { useCallback } from 'react'

import { LazyQueryResult } from '@apollo/client'

import { gql, useLazyQuery } from '@sourcegraph/http-client'
import { SearchPatternTypeProps } from '@sourcegraph/search'
import { PlatformContext } from '@sourcegraph/shared/src/platform/context'
import { IQuery, SearchResult } from '@sourcegraph/shared/src/schema'

import { eventLogger } from '../../tracking/eventLogger'

interface ExportSearchResultsConfig extends SearchPatternTypeProps, Pick<PlatformContext, 'sourcegraphURL'> {
    query: string
}

type ExportSearchResultsQueryResult = Pick<IQuery, 'search'>

interface ExportSearchResultsQueryVariables extends SearchPatternTypeProps {
    query: string
}

type UseExportSearchResultsQuery = (
    config: ExportSearchResultsConfig
) => [() => void, LazyQueryResult<ExportSearchResultsQueryResult, ExportSearchResultsQueryVariables>]

const SEARCH_RESULTS_QUERY = gql`
    query SearchResults($query: String!, $patternType: SearchPatternType) {
        search(query: $query, patternType: $patternType) {
            results {
                results {
                    __typename
                    ... on CommitSearchResult {
                        url
                        commit {
                            subject
                            author {
                                date
                                person {
                                    displayName
                                }
                            }
                        }
                    }
                    ... on Repository {
                        name
                        externalURLs {
                            url
                        }
                    }
                    ... on FileMatch {
                        repository {
                            name
                            externalURLs {
                                url
                            }
                        }
                        file {
                            path
                            canonicalURL
                            externalURLs {
                                url
                            }
                        }
                        lineMatches {
                            preview
                            offsetAndLengths
                        }
                    }
                }
            }
        }
    }
`

const searchResultsToFileContent = (searchResults: SearchResult[], sourcegraphURL: string): string => {
    const headers =
        searchResults[0].__typename !== 'CommitSearchResult'
            ? [
                  'Match type',
                  'Repository',
                  'Repository external URL',
                  'File path',
                  'File URL',
                  'File external URL',
                  'Search matches',
              ]
            : ['Date', 'Author', 'Subject', 'Commit URL']
    const content = [
        headers,
        ...searchResults.map(result => {
            switch (result.__typename) {
                // on FileMatch
                case 'FileMatch': {
                    const searchMatches = result.lineMatches
                        .map(line =>
                            line.offsetAndLengths
                                .map(offset => line.preview?.slice(offset[0], offset[0] + offset[1]))
                                .join(' ')
                        )
                        .join(' ')
                    return [
                        result.__typename,
                        result.repository.name,
                        result.repository.externalURLs[0]?.url,
                        result.file.path,
                        new URL(result.file.canonicalURL, sourcegraphURL).toString(),
                        result.file.externalURLs[0]?.url,
                        searchMatches,
                    ].map(string_ => JSON.stringify(string_))
                }
                // on Repository
                case 'Repository':
                    return [result.__typename, result.name, result.externalURLs[0]?.url].map(string_ =>
                        JSON.stringify(string_)
                    )
                // on CommitSearchResult
                case 'CommitSearchResult':
                    return [
                        result.commit.author.date,
                        result.commit.author.person.displayName,
                        result.commit.subject,
                        result.url,
                    ].map(string_ => JSON.stringify(string_))
                // If no typename can be found
                default:
                    throw new Error('Please try another query.')
            }
        }),
    ]
        .map(row => row.join(','))
        .join('\n')

    return content
}

export const useExportSearchResultsQuery: UseExportSearchResultsQuery = ({
    query = '',
    patternType,
    sourcegraphURL,
}) => {
    const [getSearchResults, data] = useLazyQuery<ExportSearchResultsQueryResult, ExportSearchResultsQueryVariables>(
        SEARCH_RESULTS_QUERY,
        {
            onCompleted: data => {
                const results = data.search?.results.results
                if (!results?.length || !results[0]) {
                    eventLogger.log('SearchExportFailed')
                    throw new Error('No results to be exported.')
                }
                const content = searchResultsToFileContent(results, sourcegraphURL)
                const downloadFilename = `sourcegraph-search-export-${query.replace(/\W/g, '-')}.csv`
                const blob = new Blob([content], { type: 'text/csv' })
                const url = URL.createObjectURL(blob)

                const a = document.createElement('a')
                a.href = url
                a.style.display = 'none'
                a.download = downloadFilename
                a.click()
                eventLogger.log('SearchExportPerformed', { count: results.length }, { count: results.length })

                // cleanup
                a.remove()
                URL.revokeObjectURL(url)
            },
        }
    )

    const requestSearchResultsExport = useCallback(() => {
        // eslint-disable-next-line no-void
        void getSearchResults({ variables: { query, patternType } })
    }, [query, patternType, getSearchResults])

    return [requestSearchResultsExport, data]
}

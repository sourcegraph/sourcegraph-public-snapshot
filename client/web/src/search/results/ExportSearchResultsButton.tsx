import { useContext } from 'react'

import { mdiDownload } from '@mdi/js'

import { NotificationType } from '@sourcegraph/extension-api-classes'
import { gql, useLazyQuery } from '@sourcegraph/http-client'
import { SearchPatternTypeProps } from '@sourcegraph/search'
import { NotificationContext } from '@sourcegraph/shared/src/notifications/Notifications'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { IQuery } from '@sourcegraph/shared/src/schema'
import { SettingsCascadeOrError, SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { Button, Icon } from '@sourcegraph/wildcard'

import { getFromSettings } from '../../util/settings'

interface Props extends SearchPatternTypeProps, SettingsCascadeProps, PlatformContextProps<'sourcegraphURL'> {
    query?: string
}

type SearchResultsQueryData = Pick<IQuery, 'search'>

interface SearchResultsQueryVariables extends SearchPatternTypeProps {
    query: string
}

const DEFAULT_MAX_CONTENT_LENGTH = 200

/**
 * Truncate "Search matches" string to avoid hitting the maximum URL length in Chrome:
 * https://chromium.googlesource.com/chromium/src/+/refs/heads/main/docs/security/url_display_guidelines/url_display_guidelines.md#url-length)
 */
function truncateMatches(settingsCascade: SettingsCascadeOrError, searchMatches: string): string {
    // Determine at what length "Search matches" string should be truncated (to avoid hitting the maximum URL length in Chrome:
    // https://chromium.googlesource.com/chromium/src/+/refs/heads/main/docs/security/url_display_guidelines/url_display_guidelines.md#url-length)
    let maxMatchContentLength = getFromSettings<number>(settingsCascade, 'searchExport.maxMatchContentLength')

    if (typeof maxMatchContentLength !== 'number' || maxMatchContentLength < 0) {
        maxMatchContentLength = DEFAULT_MAX_CONTENT_LENGTH
    }

    return searchMatches.length > maxMatchContentLength
        ? `${searchMatches.slice(0, maxMatchContentLength)}...`
        : searchMatches
}

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

export const ExportSearchResultsButton: React.FC<Props> = ({
    query = '',
    patternType,
    platformContext,
    settingsCascade,
}) => {
    const { addNotification } = useContext(NotificationContext)
    const [fetchCSV] = useLazyQuery<SearchResultsQueryData, SearchResultsQueryVariables>(SEARCH_RESULTS_QUERY, {
        variables: { query, patternType },
        onCompleted: data => {
            const results = data.search?.results.results
            if (!results?.length || !results[0]) {
                throw new Error('No results to be exported.')
            }
            const headers =
                results[0].__typename !== 'CommitSearchResult'
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
            const csvData = [
                headers,
                ...results.map(result => {
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
                                new URL(result.file.canonicalURL, platformContext.sourcegraphURL).toString(),
                                result.file.externalURLs[0]?.url,
                                truncateMatches(settingsCascade, searchMatches),
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
            const encodedData = encodeURIComponent(csvData.map(row => row.join(',')).join('\n'))
            const downloadFilename = `sourcegraph-search-export-${query.replace(/\W/g, '-')}.csv`
            addNotification({
                type: NotificationType.Success,
                message: `Search results export is complete.\n\n<a href="data:text/csv;charset=utf-8,${encodedData}" download="${downloadFilename}"><strong>Download CSV</strong></a>`,
            })
        },
    })

    return (
        <Button
            className="btn btn-sm btn-outline-secondary text-decoration-none"
            variant="secondary"
            outline={true}
            onClick={() => fetchCSV()}
        >
            <Icon aria-hidden={true} className="mr-1" svgPath={mdiDownload} />
            Export Results
        </Button>
    )
}

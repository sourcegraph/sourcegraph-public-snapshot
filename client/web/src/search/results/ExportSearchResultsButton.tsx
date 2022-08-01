import { useContext, useState } from 'react'

import { mdiDownload } from '@mdi/js'

import { NotificationType } from '@sourcegraph/extension-api-classes'
import { gql, useLazyQuery } from '@sourcegraph/http-client'
import { SearchPatternTypeProps } from '@sourcegraph/search'
import { NotificationContext } from '@sourcegraph/shared/src/notifications/Notifications'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { Button, Icon } from '@sourcegraph/wildcard'

interface Props extends SearchPatternTypeProps, PlatformContextProps<'sourcegraphURL'> {
    query?: string
}

const searchResultsQuery = gql`
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

const DEFAULT_MAX_CONTENT_LENGTH = 200

function truncateMatches(searchMatches: string): string {
    // Determine at what length "Search matches" string should be truncated (to avoid hitting the maximum URL length in Chrome:
    // https://chromium.googlesource.com/chromium/src/+/refs/heads/main/docs/security/url_display_guidelines/url_display_guidelines.md#url-length)
    // let maxMatchContentLength = sourcegraph.configuration.get<Settings>().value['searchExport.maxMatchContentLength']

    // if (typeof maxMatchContentLength !== 'number' || maxMatchContentLength < 0) {
    //     maxMatchContentLength =  DEFAULT_MAX_CONTENT_LENGTH
    // }

    const maxMatchContentLength = DEFAULT_MAX_CONTENT_LENGTH

    return searchMatches.length > maxMatchContentLength
        ? `${searchMatches.slice(0, maxMatchContentLength)}...`
        : searchMatches
}

export const ExportSearchResultsButton: React.FC<Props> = ({ query, patternType, platformContext }) => {
    const { addNotification } = useContext(NotificationContext)
    const [csv, setCsv] = useState<string>('')
    const [fetchCSV] = useLazyQuery<any, any>(searchResultsQuery, {
        variables: { query, patternType },
        onCompleted: data => {
            const results = data.search.results.results
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
                ...results.map(r => {
                    switch (r.__typename) {
                        // on FileMatch
                        case 'FileMatch':
                            const searchMatches = r.lineMatches
                                .map(line =>
                                    line.offsetAndLengths
                                        .map(offset => line.preview?.substring(offset[0], offset[0] + offset[1]))
                                        .join(' ')
                                )
                                .join(' ')
                            return [
                                r.__typename,
                                r.repository.name,
                                r.repository.externalURLs[0]?.url,
                                r.file.path,
                                new URL(r.file.canonicalURL, platformContext.sourcegraphURL).toString(),
                                r.file.externalURLs[0]?.url,
                                truncateMatches(searchMatches),
                            ].map(s => JSON.stringify(s))
                        // on Repository
                        case 'Repository':
                            return [r.__typename, r.name, r.externalURLs[0]?.url].map(s => JSON.stringify(s))
                        // on CommitSearchResult
                        case 'CommitSearchResult':
                            return [
                                r.commit.author.date,
                                r.commit.author.person.displayName,
                                r.commit.subject,
                                r.url,
                            ].map(s => JSON.stringify(s))
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
            // sourcegraph.app.activeWindow?.showNotification(
            //     `Search results export is complete.\n\n<a href="data:text/csv;charset=utf-8,${encodedData}" download="${downloadFilename}"><strong>Download CSV</strong></a>`,
            //     sourcegraph.NotificationType.Success
            // )
            // setCsv(encodedData)
            // console.log(encodedData)

            // const element = document.createElement('a')
            // element.setAttribute('href', 'data:text/csv;charset=utf-8,' + encodedData)
            // element.setAttribute('download', 'search_results.csv')
            // element.style.display = 'none'
            // document.body.append(element)
            // element.click()
            // element.remove()
        },
    })

    return (
        <>
            <Button
                className="btn btn-sm btn-outline-secondary text-decoration-none"
                variant="secondary"
                outline={true}
                onClick={() => fetchCSV()}
            >
                <Icon aria-hidden={true} className="mr-1" svgPath={mdiDownload} />
                Export Results
            </Button>
        </>
    )
}

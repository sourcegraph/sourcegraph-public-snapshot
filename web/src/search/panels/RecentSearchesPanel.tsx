import classNames from 'classnames'
import React, { useEffect, useState } from 'react'
import { AuthenticatedUser } from '../../auth'
import { dataOrThrowErrors, gql, requestGraphQL } from '../../../../shared/src/graphql/graphql'
import { Link } from '../../../../shared/src/components/Link'
import { map } from 'rxjs/operators'
import { Maybe } from '../../../../shared/src/graphql-operations'
import { PanelContainer } from './PanelContainer'
import { RecentSearchesPanelDataResult, RecentSearchesPanelDataVariables } from '../../graphql-operations'
import { Timestamp } from '../../components/time/Timestamp'

interface EventLogResult {
    totalCount: number
    nodes: { argument: Maybe<string>; timestamp: string; url: string }[]
    pageInfo: { endCursor: Maybe<string>; hasNextPage: boolean }
}

interface RecentSearch {
    count: number
    searchText: string
    timestamp: string
    url: string
}

const getData = ({ userId, first }: RecentSearchesPanelDataVariables): Promise<EventLogResult> => {
    const result = requestGraphQL<RecentSearchesPanelDataResult, RecentSearchesPanelDataVariables>({
        request: gql`
            query RecentSearchesPanelData($userId: ID!, $first: Int) {
                node(id: $userId) {
                    ... on User {
                        recentSearches: eventLogs(first: $first, eventName: "SearchResultsQueried") {
                            nodes {
                                argument
                                timestamp
                                url
                            }
                            pageInfo {
                                endCursor
                                hasNextPage
                            }
                            totalCount
                        }
                    }
                }
            }
        `,
        variables: { userId, first: first ?? null },
    })

    return result
        .pipe(
            map(dataOrThrowErrors),
            map(
                (data: RecentSearchesPanelDataResult): EventLogResult => {
                    if (!data.node) {
                        throw new Error('User not found')
                    }
                    return data.node.recentSearches
                }
            )
        )
        .toPromise()
}

const processRecentSearches = (eventLogResult: EventLogResult): RecentSearch[] => {
    const recentSearches: RecentSearch[] = []

    for (const node of eventLogResult.nodes) {
        if (node.argument) {
            const parsedArguments = JSON.parse(node.argument)
            const searchText: string = parsedArguments?.code_search?.query_data?.combined

            if (recentSearches.length > 0 && recentSearches[recentSearches.length - 1].searchText === searchText) {
                recentSearches[recentSearches.length - 1].count += 1
            } else {
                const parsedUrl = new URL(node.url)
                recentSearches.push({
                    count: 1,
                    url: parsedUrl.pathname + parsedUrl.search, // Strip domain from URL so clicking on it doesn't reload page
                    searchText,
                    timestamp: node.timestamp,
                })
            }
        }
    }

    return recentSearches
}

export const RecentSearchesPanel: React.FunctionComponent<{
    className?: string
    authenticatedUser: AuthenticatedUser | null
}> = ({ className, authenticatedUser }) => {
    const [state, setState] = useState<'loading' | 'populated' | 'empty'>('loading')
    const [recentSearches, setRecentSearches] = useState<RecentSearch[]>([])

    useEffect(() => {
        const getDataAsync = async (): Promise<void> => {
            if (!authenticatedUser) {
                return
            }
            const data = await getData({ userId: authenticatedUser.id, first: 100 })
            setRecentSearches(processRecentSearches(data))
            if (!data.totalCount) {
                setState('empty')
            } else {
                setState('populated')
            }
        }

        // eslint-disable-next-line @typescript-eslint/no-floating-promises
        getDataAsync()
    }, [authenticatedUser])

    const loadingDisplay = <div>Loading</div>
    const contentDisplay = (
        <table className="recent-searches-panel__results-table">
            <thead className="recent-searches-panel__results-table-head">
                <tr>
                    <th>Count</th>
                    <th>Search</th>
                    <th>Date</th>
                </tr>
            </thead>
            <tbody className="recent-searches-panel__results-table-body">
                {recentSearches.map(recentSearch => (
                    <tr key={recentSearch.timestamp}>
                        <td className="recent-searches-panel__results-count-cell">
                            <span className="recent-searches-panel__results-count">{recentSearch.count}</span>
                        </td>
                        <td>
                            <Link to={recentSearch.url}>{recentSearch.searchText}</Link>
                        </td>
                        <td>
                            <Timestamp noAbout={true} date={recentSearch.timestamp} />
                        </td>
                    </tr>
                ))}
            </tbody>
        </table>
    )
    const emptyDisplay = <div>Empty</div>

    return (
        <PanelContainer
            className={classNames(className, 'recent-searches-panel')}
            title="Recent searches"
            state={state}
            loadingContent={loadingDisplay}
            populatedContent={contentDisplay}
            emptyContent={emptyDisplay}
        />
    )
}

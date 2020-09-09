import classNames from 'classnames'
import React, { useState, useMemo } from 'react'
import { AuthenticatedUser } from '../../auth'
import { Link } from '../../../../shared/src/components/Link'
import { PanelContainer } from './PanelContainer'
import { Timestamp } from '../../components/time/Timestamp'
import { EventLogResult } from '../backend'
import { Observable } from 'rxjs'
import { useObservable } from '../../../../shared/src/util/useObservable'
import { SearchPatternType } from '../../graphql-operations'
import { buildSearchURLQuery } from '../../../../shared/src/util/url'
import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'

interface RecentSearch {
    count: number
    searchText: string
    timestamp: string
    url: string
}

const processRecentSearches = (eventLogResult?: EventLogResult): RecentSearch[] | null => {
    if (!eventLogResult) {
        return null
    }

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
    fetchRecentSearches: (userId: string, first: number) => Observable<EventLogResult>
}> = ({ className, authenticatedUser, fetchRecentSearches }) => {
    const pageSize = 20

    const [itemsToLoad, setItmesToLoad] = useState(pageSize)
    const recentSearches = useObservable(
        useMemo(() => fetchRecentSearches(authenticatedUser?.id || '', itemsToLoad), [
            authenticatedUser?.id,
            fetchRecentSearches,
            itemsToLoad,
        ])
    )

    const processedResults = useMemo(() => (recentSearches ? processRecentSearches(recentSearches) : null), [
        recentSearches,
    ])

    const loadingDisplay = (
        <div className="d-flex justify-content-center align-items-center panel-container__empty-container">
            <div className="icon-inline">
                <LoadingSpinner />
            </div>
            Loading recent searches
        </div>
    )
    const emptyDisplay = (
        <div className="panel-container__empty-container">
            <small className="mb-2">
                Your recent searches will be displayed here. Here are a few searches to get you started:
            </small>

            <ul className="recent-searches-panel__examples-list">
                <li className="recent-searches-panel__examples-list-item">
                    <Link
                        to={
                            '/search?' +
                            buildSearchURLQuery(
                                'lang:c if(:[eval_match]) { :[statement_match] }',
                                SearchPatternType.structural,
                                false
                            )
                        }
                        className="text-monospace"
                    >
                        lang:c if(:[eval_match]) {'{'} :[statement_match] {'}'}
                    </Link>
                </li>
                <li className="recent-searches-panel__examples-list-item">
                    <Link
                        to={
                            '/search?' +
                            buildSearchURLQuery(
                                'lang:java type:diff after:"1 week ago"',
                                SearchPatternType.literal,
                                false
                            )
                        }
                        className="text-monospace"
                    >
                        lang:java type:diff after:"1 week ago"
                    </Link>
                </li>
                <li className="recent-searches-panel__examples-list-item">
                    <Link
                        to={'/search?' + buildSearchURLQuery('lang:java', SearchPatternType.literal, false)}
                        className="text-monospace"
                    >
                        lang:java
                    </Link>
                </li>
            </ul>
        </div>
    )

    const contentDisplay = (
        <>
            <table className="recent-searches-panel__results-table">
                <thead>
                    <tr className="recent-searches-panel__results-table-row">
                        <th>
                            <small>Count</small>
                        </th>
                        <th>
                            <small>Search</small>
                        </th>
                        <th>
                            <small>Date</small>
                        </th>
                    </tr>
                </thead>
                <tbody className="recent-searches-panel__results-table-body">
                    {processedResults?.map((recentSearch, index) => (
                        <tr key={index} className="recent-searches-panel__results-table-row">
                            <td className="recent-searches-panel__results-table-count-col">
                                <span className="recent-searches-panel__results-table-count btn btn-secondary d-inline-flex justify-content-center align-items-center">
                                    {recentSearch.count}
                                </span>
                            </td>
                            <td>
                                <Link to={recentSearch.url} className="text-monospace">
                                    {recentSearch.searchText}
                                </Link>
                            </td>
                            <td className="recent-searches-panel__results-table-date-col">
                                <Timestamp noAbout={true} date={recentSearch.timestamp} />
                            </td>
                        </tr>
                    ))}
                </tbody>
            </table>
            {recentSearches?.pageInfo.hasNextPage && (
                <div className="text-center">
                    <button
                        type="button"
                        className="btn btn-secondary"
                        onClick={() => setItmesToLoad(current => current + pageSize)}
                    >
                        Show more
                    </button>
                </div>
            )}
        </>
    )

    return (
        <PanelContainer
            className={classNames(className, 'recent-searches-panel')}
            title="Recent searches"
            state={processedResults ? (processedResults.length > 0 ? 'populated' : 'empty') : 'loading'}
            loadingContent={loadingDisplay}
            populatedContent={contentDisplay}
            emptyContent={emptyDisplay}
        />
    )
}

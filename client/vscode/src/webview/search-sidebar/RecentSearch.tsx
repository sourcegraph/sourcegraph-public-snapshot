import classNames from 'classnames'
import ChevronDownIcon from 'mdi-react/ChevronDownIcon'
import ChevronLeftIcon from 'mdi-react/ChevronLeftIcon'
import React, { useEffect, useState } from 'react'

import { SyntaxHighlightedSearchQuery } from '@sourcegraph/branded/src/components/SyntaxHighlightedSearchQuery'
import { AuthenticatedUser } from '@sourcegraph/shared/src/auth'
import { Link } from '@sourcegraph/shared/src/components/Link'
import { EventLogsDataResult, EventLogsDataVariables } from '@sourcegraph/shared/src/graphql-operations'
import { EventLogResult } from '@sourcegraph/shared/src/search/backend'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { LocalRecentSeachProps } from '../contract'
import { WebviewPageProps } from '../platform/context'
import { eventsQuery } from '../search-panel/queries'

import styles from './HistorySidebar.module.scss'

interface RecentSearch {
    count: number
    searchText: string
    timestamp: string
    url: string
}

interface SearchHistoryProps extends WebviewPageProps, TelemetryProps {
    localRecentSearches: LocalRecentSeachProps[] | undefined
    authenticatedUser: AuthenticatedUser | null
}

export const RecentSearch: React.FunctionComponent<SearchHistoryProps> = ({
    localRecentSearches,
    sourcegraphVSCodeExtensionAPI,
    authenticatedUser,
    telemetryService,
    platformContext,
}) => {
    const itemsToLoad = 15
    const [collapsed, setCollapsed] = useState(false)
    const [calledAPI, setCalledAPI] = useState(false)
    const [processedResults, setProcessedResults] = useState<RecentSearch[] | null>(null)

    useEffect(() => {
        if (authenticatedUser && !calledAPI) {
            ;(async () => {
                const eventVariables = {
                    userId: authenticatedUser.id,
                    first: itemsToLoad,
                    eventName: 'SearchResultsQueried',
                }
                const userSearchHistory = await platformContext
                    .requestGraphQL<EventLogsDataResult, EventLogsDataVariables>({
                        request: eventsQuery,
                        variables: eventVariables,
                        mightContainPrivateInfo: true,
                    })
                    .toPromise()
                if (userSearchHistory.data?.node?.__typename === 'User') {
                    setProcessedResults(processRecentSearches(userSearchHistory.data.node.eventLogs))
                }
            })().catch(error => console.error(error))
        }
        setCalledAPI(true)
    }, [authenticatedUser, calledAPI, itemsToLoad, localRecentSearches, platformContext])

    return (
        <div className={styles.sidebarSection}>
            <button
                type="button"
                className={classNames('btn btn-outline-secondary', styles.sidebarSectionCollapseButton)}
                onClick={() => setCollapsed(!collapsed)}
            >
                <h5 className="flex-grow-1">Recent Searches</h5>
                {collapsed ? (
                    <ChevronLeftIcon className="icon-inline mr-1" />
                ) : (
                    <ChevronDownIcon className="icon-inline mr-1" />
                )}
            </button>
            {/* Display results from cloud for registered users and results from local Storage for non registered users */}
            {authenticatedUser && processedResults && !collapsed && (
                <div className={classNames('p-1', styles.sidebarSectionList)}>
                    {processedResults?.map((search, index) => (
                        <div key={index}>
                            <small key={index} className={styles.sidebarSectionListItem}>
                                <Link
                                    to="/"
                                    onClick={() =>
                                        sourcegraphVSCodeExtensionAPI.setActiveWebviewQueryState({
                                            query: search.searchText,
                                        })
                                    }
                                >
                                    <SyntaxHighlightedSearchQuery query={search.searchText} />
                                </Link>
                            </small>
                        </div>
                    ))}
                </div>
            )}
            {!authenticatedUser && localRecentSearches && !collapsed && (
                <div className={classNames('p-1', styles.sidebarSectionList)}>
                    {localRecentSearches
                        ?.slice(0)
                        .reverse()
                        .filter((search, index) => index <= itemsToLoad - 1)
                        .map((search, index) => (
                            <div key={index}>
                                <small key={index} className={styles.sidebarSectionListItem}>
                                    <Link
                                        to="/"
                                        onClick={() =>
                                            sourcegraphVSCodeExtensionAPI.setActiveWebviewQueryState({
                                                query: search.lastQuery,
                                            })
                                        }
                                    >
                                        <SyntaxHighlightedSearchQuery query={search.lastQuery} />
                                    </Link>
                                </small>
                            </div>
                        ))}
                </div>
            )}
        </div>
    )
}

function processRecentSearches(eventLogResult?: EventLogResult): RecentSearch[] | null {
    if (!eventLogResult) {
        return null
    }

    const recentSearches: RecentSearch[] = []

    for (const node of eventLogResult.nodes) {
        if (node.argument && node.url) {
            const parsedArguments = JSON.parse(node.argument)
            const searchText: string | undefined = parsedArguments?.code_search?.query_data?.combined

            if (searchText) {
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
    }

    return recentSearches
}

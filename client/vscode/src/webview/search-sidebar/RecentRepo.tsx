import classNames from 'classnames'
import ChevronDownIcon from 'mdi-react/ChevronDownIcon'
import ChevronLeftIcon from 'mdi-react/ChevronLeftIcon'
import React, { useEffect, useState } from 'react'

import { SyntaxHighlightedSearchQuery } from '@sourcegraph/branded/src/components/SyntaxHighlightedSearchQuery'
import { AuthenticatedUser } from '@sourcegraph/shared/src/auth'
import { Link } from '@sourcegraph/shared/src/components/Link'
import { EventLogsDataResult, EventLogsDataVariables } from '@sourcegraph/shared/src/graphql-operations'
import { EventLogResult } from '@sourcegraph/shared/src/search/backend'
import { scanSearchQuery } from '@sourcegraph/shared/src/search/query/scanner'
import { isRepoFilter } from '@sourcegraph/shared/src/search/query/validate'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { LocalRecentSeachProps } from '../contract'
import { WebviewPageProps } from '../platform/context'
import { eventsQuery } from '../search-panel/queries'

import styles from './HistorySidebar.module.scss'

interface RecentRepo {
    repoName: string
    filePath: string
    timestamp: string
    url: string
}

interface RecentRepoProps extends WebviewPageProps, TelemetryProps {
    localRecentSearches: LocalRecentSeachProps[] | undefined
    authenticatedUser: AuthenticatedUser | null
}

export const RecentRepo: React.FunctionComponent<RecentRepoProps> = ({
    localRecentSearches,
    sourcegraphVSCodeExtensionAPI,
    authenticatedUser,
    telemetryService,
    platformContext,
}) => {
    const itemsToLoad = 15
    const [calledAPI, setCalledAPI] = useState(false)
    const [collapsed, setCollapsed] = useState(false)
    const [processedResults, setProcessedResults] = useState<string[] | null>(null)

    useEffect(() => {
        // only call API once
        if (authenticatedUser && !calledAPI) {
            ;(async () => {
                const eventVariables = {
                    userId: authenticatedUser.id,
                    first: 200,
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
                    const results = processRepositories(userSearchHistory.data.node.eventLogs)
                    setProcessedResults(results)
                } else {
                    setProcessedResults(null)
                }
                setCalledAPI(true)
            })().catch(error => console.error(error))
        }
        if (!authenticatedUser && localRecentSearches && !calledAPI) {
            setProcessedResults(processLocalRepositories(localRecentSearches))
            setCalledAPI(true)
        }
    }, [authenticatedUser, calledAPI, itemsToLoad, localRecentSearches, platformContext, processedResults])

    if (!processedResults) {
        return null
    }

    return (
        <div className={styles.sidebarSection}>
            <button
                type="button"
                className={classNames('btn btn-outline-secondary', styles.sidebarSectionCollapseButton)}
                onClick={() => setCollapsed(!collapsed)}
            >
                <h5 className="flex-grow-1">Recent Repositories</h5>
                {collapsed ? (
                    <ChevronLeftIcon className="icon-inline mr-1" />
                ) : (
                    <ChevronDownIcon className="icon-inline mr-1" />
                )}
            </button>
            {!collapsed && (
                <div className={classNames('p-1', styles.sidebarSectionList)}>
                    {processedResults
                        ?.filter((search, index) => index < itemsToLoad)
                        .map((repo, index) => (
                            <div key={index}>
                                <small key={index} className={styles.sidebarSectionListItem}>
                                    <Link
                                        data-testid="recent-files-item"
                                        to="/"
                                        onClick={() =>
                                            sourcegraphVSCodeExtensionAPI.setActiveWebviewQueryState({
                                                query: `repo:${repo}`,
                                            })
                                        }
                                    >
                                        <SyntaxHighlightedSearchQuery query={`r:${repo}`} />
                                    </Link>
                                </small>
                            </div>
                        ))}
                </div>
            )}
        </div>
    )
}

export function parseSearchURLQuery(query: string): string | undefined {
    const searchParameters = new URLSearchParams(query)
    return searchParameters.get('q') || undefined
}

function processRepositories(eventLogResult: EventLogResult): string[] | null {
    if (!eventLogResult) {
        return null
    }
    const recentlySearchedRepos: string[] = []
    for (const node of eventLogResult.nodes) {
        if (node.url) {
            const url = new URL(node.url)
            const queryFromURL = parseSearchURLQuery(url.search)
            const scannedQuery = scanSearchQuery(queryFromURL || '')
            if (scannedQuery.type === 'success') {
                for (const token of scannedQuery.term) {
                    if (isRepoFilter(token)) {
                        if (token.value && !recentlySearchedRepos.includes(token.value.value)) {
                            recentlySearchedRepos.push(token.value.value)
                        }
                    }
                }
            }
        }
    }
    return recentlySearchedRepos
}

function processLocalRepositories(localRecentSearches: LocalRecentSeachProps[]): string[] | null {
    const recentlySearchedRepoNames = new Set<string>()

    for (const search of localRecentSearches) {
        const repoNameRegex = /(?<=repo:)(\S+)/
        const repoName = search.lastQuery.match(repoNameRegex)
        if (repoName?.[0] && typeof repoName?.[0] === 'string') {
            recentlySearchedRepoNames.add(repoName?.[0])
        }
    }

    return recentlySearchedRepoNames ? [...recentlySearchedRepoNames].reverse() : null
}

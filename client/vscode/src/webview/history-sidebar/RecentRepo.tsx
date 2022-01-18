import classNames from 'classnames'
import PlusIcon from 'mdi-react/PlusIcon'
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
    const [showMore, setShowMore] = useState(false)
    const [itemsToLoad, setItemsToLoad] = useState(5)

    function loadMoreItems(): void {
        setItemsToLoad(current => current + 5)
        telemetryService.log('RecentSearchesPanelShowMoreClicked')
    }

    const [processedResults, setProcessedResults] = useState<string[] | null>(null)

    useEffect(() => {
        if (authenticatedUser && itemsToLoad) {
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
                    setShowMore(userSearchHistory.data.node.eventLogs.pageInfo.hasNextPage)
                    setProcessedResults(processRepositories(userSearchHistory.data.node.eventLogs))
                }
            })().catch(error => console.error(error))
        }
        if (!authenticatedUser) {
            if (localRecentSearches) {
                setProcessedResults(processLocalRepositories(localRecentSearches))
            }
        }
    }, [authenticatedUser, itemsToLoad, localRecentSearches, platformContext])

    return (
        <div className={styles.sidebarSection}>
            <button
                type="button"
                className={classNames('btn btn-outline-secondary', styles.sidebarSectionCollapseButton)}
                onClick={() => sourcegraphVSCodeExtensionAPI.openSearchPanel()}
            >
                <h5 className="flex-grow-1">Repositories</h5>
                <PlusIcon className="icon-inline mr-1" />
            </button>
            <div className={classNames('p-1', styles.sidebarSectionList)}>
                {processedResults?.map((repo, index) => (
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
                                <SyntaxHighlightedSearchQuery query={`repo:${repo}`} />
                            </Link>
                        </small>
                    </div>
                ))}
                {showMore && <ShowMoreButton onClick={loadMoreItems} className="my-0" />}
            </div>
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
        if (typeof repoName?.[0] === 'string') {
            recentlySearchedRepoNames.add(repoName?.[0])
        }
    }

    return recentlySearchedRepoNames ? [...recentlySearchedRepoNames].reverse() : null
}

const ShowMoreButton: React.FunctionComponent<{ onClick: () => void; className?: string }> = ({
    onClick,
    className,
}) => (
    <div className="text-center py-3">
        <button type="button" className={classNames('btn btn-link', className)} onClick={onClick}>
            Show more
        </button>
    </div>
)

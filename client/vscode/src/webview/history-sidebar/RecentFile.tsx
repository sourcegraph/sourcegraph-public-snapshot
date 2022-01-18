import classNames from 'classnames'
import PlusIcon from 'mdi-react/PlusIcon'
import React, { useEffect, useState } from 'react'

import { AuthenticatedUser } from '@sourcegraph/shared/src/auth'
import { Link } from '@sourcegraph/shared/src/components/Link'
import { EventLogsDataResult, EventLogsDataVariables } from '@sourcegraph/shared/src/graphql-operations'
import { EventLogResult } from '@sourcegraph/shared/src/search/backend'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { LocalRecentSeachProps } from '../contract'
import { WebviewPageProps } from '../platform/context'
import { eventsQuery } from '../search-panel/queries'

import styles from './HistorySidebar.module.scss'

interface RecentFile {
    repoName: string
    filePath: string
    timestamp: string
    url: string
}

interface RecentFileProps extends WebviewPageProps, TelemetryProps {
    localRecentSearches: LocalRecentSeachProps[] | undefined
    authenticatedUser: AuthenticatedUser | null
}

export const RecentFile: React.FunctionComponent<RecentFileProps> = ({
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

    const [processedResults, setProcessedResults] = useState<RecentFile[] | null>(null)

    useEffect(() => {
        if (authenticatedUser && itemsToLoad) {
            ;(async () => {
                const eventVariables = {
                    userId: authenticatedUser.id,
                    first: itemsToLoad,
                    eventName: 'ViewBlob',
                }
                const userSearchHistory = await platformContext
                    .requestGraphQL<EventLogsDataResult, EventLogsDataVariables>({
                        request: eventsQuery,
                        variables: eventVariables,
                        mightContainPrivateInfo: true,
                    })
                    .toPromise()
                console.log(userSearchHistory)
                if (userSearchHistory.data?.node?.__typename === 'User') {
                    setShowMore(userSearchHistory.data.node.eventLogs.pageInfo.hasNextPage)
                    setProcessedResults(processRecentFiles(userSearchHistory.data.node.eventLogs))
                }
            })().catch(error => console.error(error))
        }
    }, [authenticatedUser, itemsToLoad, platformContext])

    return (
        <div className={styles.sidebarSection}>
            <button
                type="button"
                className={classNames('btn btn-outline-secondary', styles.sidebarSectionCollapseButton)}
                onClick={() => sourcegraphVSCodeExtensionAPI.openSearchPanel()}
            >
                <h5 className="flex-grow-1">Recent Files</h5>
                <PlusIcon className="icon-inline mr-1" />
            </button>
            {/* Display results from cloud for registered users and results from local Storage for non registered users */}
            {authenticatedUser && processedResults ? (
                <div className={classNames('p-1', styles.sidebarSectionList)}>
                    {processedResults?.map((recentFile, index) => (
                        <div key={index}>
                            <small key={index} className={styles.sidebarSectionListItem}>
                                <Link
                                    data-testid="recent-files-item"
                                    to="/"
                                    onClick={() =>
                                        sourcegraphVSCodeExtensionAPI.setActiveWebviewQueryState({
                                            query: `repo:^${recentFile.repoName}$ file:^${recentFile.filePath}`,
                                        })
                                    }
                                >
                                    {recentFile.repoName} › {recentFile.filePath}
                                </Link>
                            </small>
                        </div>
                    ))}
                    {showMore && <ShowMoreButton onClick={loadMoreItems} className="my-0" />}
                </div>
            ) : (
                <div className={classNames('p-1', styles.sidebarSectionList)}>
                    <p>For registered users only</p>
                </div>
            )}
        </div>
    )
}

function processRecentFiles(eventLogResult?: EventLogResult): RecentFile[] | null {
    if (!eventLogResult) {
        return null
    }

    const recentFiles: RecentFile[] = []

    for (const node of eventLogResult.nodes) {
        if (node.argument && node.url) {
            const parsedArguments = JSON.parse(node.argument)
            let repoName = parsedArguments?.repoName as string
            let filePath = parsedArguments?.filePath as string

            if (!repoName || !filePath) {
                ;({ repoName, filePath } = extractFileInfoFromUrl(node.url))
            }

            if (
                filePath &&
                repoName &&
                !recentFiles.some(file => file.repoName === repoName && file.filePath === filePath) // Don't show the same file twice
            ) {
                const parsedUrl = new URL(node.url)
                recentFiles.push({
                    url: parsedUrl.pathname + parsedUrl.search, // Strip domain from URL so clicking on it doesn't reload page
                    repoName,
                    filePath,
                    timestamp: node.timestamp,
                })
            }
        }
    }

    return recentFiles
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

function extractFileInfoFromUrl(url: string): { repoName: string; filePath: string } {
    const parsedUrl = new URL(url)

    // Remove first character as it's a '/'
    const [repoName, filePath] = parsedUrl.pathname.slice(1).split('/-/blob/')
    if (!repoName || !filePath) {
        return { repoName: '', filePath: '' }
    }

    return { repoName, filePath }
}

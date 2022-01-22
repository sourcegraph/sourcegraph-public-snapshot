import classNames from 'classnames'
import ChevronDownIcon from 'mdi-react/ChevronDownIcon'
import ChevronLeftIcon from 'mdi-react/ChevronLeftIcon'
import React, { useEffect, useState } from 'react'

import { AuthenticatedUser } from '@sourcegraph/shared/src/auth'
import { Link } from '@sourcegraph/shared/src/components/Link'
import { EventLogsDataResult, EventLogsDataVariables } from '@sourcegraph/shared/src/graphql-operations'
import { EventLogResult } from '@sourcegraph/shared/src/search/backend'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'

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
    localFileHistory: string[]
    authenticatedUser: AuthenticatedUser | null
}

export const RecentFile: React.FunctionComponent<RecentFileProps> = ({
    localFileHistory,
    sourcegraphVSCodeExtensionAPI,
    authenticatedUser,
    telemetryService,
    platformContext,
}) => {
    const [showMore, setShowMore] = useState(false)
    const [itemsToLoad, setItemsToLoad] = useState(5)
    const [processedResults, setProcessedResults] = useState<RecentFile[] | null>(null)
    const [collapsed, setCollapsed] = useState(false)
    function loadMoreItems(): void {
        setItemsToLoad(current => current + 5)
        telemetryService.log('VSCERecentFilesPanelShowMoreClicked')
    }

    useEffect(() => {
        if (authenticatedUser && itemsToLoad < 21) {
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
                if (userSearchHistory.data?.node?.__typename === 'User') {
                    setShowMore(userSearchHistory.data.node.eventLogs.pageInfo.hasNextPage)
                    setProcessedResults(processRecentFiles(userSearchHistory.data.node.eventLogs))
                }
            })().catch(error => console.error(error))
        } else if (!authenticatedUser && localFileHistory) {
            if (processedResults === null) {
                setProcessedResults(processLocalRecentFiles(localFileHistory))
            } else {
                setShowMore(localFileHistory.length > itemsToLoad)
            }
        }
        if (showMore && itemsToLoad > 20) {
            setShowMore(false)
        }
    }, [authenticatedUser, itemsToLoad, localFileHistory, platformContext, processedResults, showMore])

    if (!authenticatedUser && localFileHistory.length === 0) {
        return null
    }

    return (
        <div className={styles.sidebarSection}>
            <button
                type="button"
                className={classNames('btn btn-outline-secondary', styles.sidebarSectionCollapseButton)}
                onClick={() => setCollapsed(!collapsed)}
            >
                <h5 className="flex-grow-1">Recent Files</h5>
                {collapsed ? (
                    <ChevronLeftIcon className="icon-inline mr-1" />
                ) : (
                    <ChevronDownIcon className="icon-inline mr-1" />
                )}
            </button>
            {/* Display results from cloud for registered users and results from local Storage for non registered users */}
            {processedResults && !collapsed && (
                <div className={classNames('p-1', styles.sidebarSectionList)}>
                    {processedResults
                        ?.filter((search, index) => index <= itemsToLoad - 1)
                        .map((recentFile, index) => (
                            <div key={index}>
                                <small key={index} className={styles.sidebarSectionListItem}>
                                    <Link
                                        data-testid="recent-files-item"
                                        to="/"
                                        onClick={() =>
                                            authenticatedUser
                                                ? sourcegraphVSCodeExtensionAPI.setActiveWebviewQueryState({
                                                      query: `repo:^${recentFile.repoName}$ file:^${recentFile.filePath}`,
                                                  })
                                                : sourcegraphVSCodeExtensionAPI.openFile(recentFile.url)
                                        }
                                    >
                                        {recentFile.repoName} › {recentFile.filePath}
                                    </Link>
                                </small>
                            </div>
                        ))}
                    {showMore && <ShowMoreButton onClick={loadMoreItems} />}
                </div>
            )}
        </div>
    )
}

const ShowMoreButton: React.FunctionComponent<{ onClick: () => void }> = ({ onClick }) => (
    <div className="text-center py-3">
        <button type="button" className={classNames('btn', styles.sidebarSectionButtonLink)} onClick={onClick}>
            Show more
        </button>
    </div>
)

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
                    url: parsedUrl.pathname.replace('https://', 'sourcegraph://') + parsedUrl.search, // Strip domain from URL so clicking on it doesn't reload page
                    repoName,
                    filePath,
                    timestamp: node.timestamp,
                })
            }
        }
    }

    return recentFiles
}

function processLocalRecentFiles(localFiles?: string[]): RecentFile[] | null {
    if (!localFiles) {
        return null
    }

    const recentFiles: RecentFile[] = []

    for (const fileUrl of localFiles) {
        const { repoName, filePath } = extractFileInfoFromUrl(fileUrl)

        if (
            filePath &&
            repoName &&
            !recentFiles.some(file => file.repoName === repoName && file.filePath === filePath) // Don't show the same file twice
        ) {
            recentFiles.push({
                url: fileUrl,
                repoName,
                filePath,
                timestamp: '',
            })
        }
    }

    return recentFiles
}

function extractFileInfoFromUrl(url: string): { repoName: string; filePath: string } {
    const parsedUrl = new URL(url)

    // Remove first character as it's a '/'
    const [repoName, filePath] = parsedUrl.pathname.slice(1).split('/-/blob/')
    if (!repoName || !filePath) {
        return { repoName: '', filePath: '' }
    }
    return { repoName, filePath }
}

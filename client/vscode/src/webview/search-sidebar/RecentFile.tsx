import classNames from 'classnames'
import ChevronDownIcon from 'mdi-react/ChevronDownIcon'
import ChevronLeftIcon from 'mdi-react/ChevronLeftIcon'
import React, { useEffect, useState } from 'react'

import { AuthenticatedUser } from '@sourcegraph/shared/src/auth'
import { Link } from '@sourcegraph/shared/src/components/Link'
import { EventLogResult } from '@sourcegraph/shared/src/search/backend'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { EventLogsDataResult } from '../../graphql-operations'
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
    platformContext,
}) => {
    const itemsToLoad = 15
    const [processedResults, setProcessedResults] = useState<RecentFile[] | null>(null)
    const [calledAPI, setCalledAPI] = useState(false)
    const [collapsed, setCollapsed] = useState(false)

    useEffect(() => {
        if (authenticatedUser && !calledAPI) {
            ;(async () => {
                const eventVariables = {
                    userId: authenticatedUser.id,
                    first: 15,
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
                    setProcessedResults(processRecentFiles(userSearchHistory.data.node.eventLogs))
                }
            })().catch(error => console.error(error))
        } else if (!authenticatedUser && localFileHistory) {
            if (processedResults === null) {
                setProcessedResults(processLocalRecentFiles(localFileHistory))
            }
        }
        setCalledAPI(true)
    }, [authenticatedUser, calledAPI, itemsToLoad, localFileHistory, platformContext, processedResults])

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
                                        onClick={() => sourcegraphVSCodeExtensionAPI.openFile(recentFile.url)}
                                    >
                                        {recentFile.repoName.split('@')[0]} › {recentFile.filePath}
                                    </Link>
                                </small>
                            </div>
                        ))}
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
                recentFiles.push({
                    url: node.url.replace('https://', 'sourcegraph://'), // So that clicking on link would open the file directly
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

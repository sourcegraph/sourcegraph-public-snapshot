import classNames from 'classnames'
import FileCodeIcon from 'mdi-react/FileCodeIcon'
import React, { useEffect, useMemo, useState } from 'react'
import { AuthenticatedUser } from '../../auth'
import { EventLogResult } from '../backend'
import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { Observable } from 'rxjs'
import { PanelContainer } from './PanelContainer'
import { useObservable } from '../../../../shared/src/util/useObservable'
import { Link } from '../../../../shared/src/components/Link'

export const RecentFilesPanel: React.FunctionComponent<{
    className?: string
    authenticatedUser: AuthenticatedUser | null
    fetchRecentFiles: (userId: string, first: number) => Observable<EventLogResult | null>
}> = ({ className, authenticatedUser, fetchRecentFiles }) => {
    const pageSize = 20

    const [itemsToLoad, setItmesToLoad] = useState(pageSize)
    const recentFiles = useObservable(
        useMemo(() => fetchRecentFiles(authenticatedUser?.id || '', itemsToLoad), [
            authenticatedUser?.id,
            fetchRecentFiles,
            itemsToLoad,
        ])
    )

    const [processedResults, setProcessedResults] = useState<RecentFile[] | null>(null)

    // Only update processed results when results are valid to prevent
    // flashing loading screen when "Show more" button is clicked
    useEffect(() => {
        if (recentFiles) {
            setProcessedResults(processRecentFiles(recentFiles))
        }
    }, [recentFiles])

    const loadingDisplay = (
        <div className="d-flex justify-content-center align-items-center panel-container__empty-container">
            <div className="icon-inline">
                <LoadingSpinner />
            </div>
            Loading recent files
        </div>
    )

    const emptyDisplay = (
        <div className="panel-container__empty-container align-items-center">
            <FileCodeIcon className="mb-2" size="2rem" />
            <small className="mb-2">This panel will display your most recently viewed files.</small>
        </div>
    )

    const contentDisplay = (
        <div>
            <small className="mb-1">File</small>
            <dl className="list-group-flush">
                {processedResults?.map((recentFile, index) => (
                    <dd key={index} className="text-monospace test-recent-files-item">
                        <Link to={recentFile.url}>
                            {recentFile.repoName} › {recentFile.filePath}
                        </Link>
                    </dd>
                ))}
            </dl>
            {recentFiles?.pageInfo.hasNextPage && (
                <div className="text-center">
                    <button
                        type="button"
                        className="btn btn-secondary test-recent-files-panel-show-more"
                        onClick={() => setItmesToLoad(current => current + pageSize)}
                    >
                        Show more
                    </button>
                </div>
            )}
        </div>
    )

    return (
        <PanelContainer
            className={classNames(className, 'recent-files-panel')}
            title="Recent files"
            state={processedResults ? (processedResults.length > 0 ? 'populated' : 'empty') : 'loading'}
            loadingContent={loadingDisplay}
            populatedContent={contentDisplay}
            emptyContent={emptyDisplay}
        />
    )
}

interface RecentFile {
    repoName: string
    filePath: string
    timestamp: string
    url: string
}

function processRecentFiles(eventLogResult?: EventLogResult): RecentFile[] | null {
    if (!eventLogResult) {
        return null
    }

    const recentFiles: RecentFile[] = []

    for (const node of eventLogResult.nodes) {
        if (node.argument) {
            const parsedArguments = JSON.parse(node.argument)
            const repoName = parsedArguments?.repoName as string
            const filePath = parsedArguments?.filePath as string

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

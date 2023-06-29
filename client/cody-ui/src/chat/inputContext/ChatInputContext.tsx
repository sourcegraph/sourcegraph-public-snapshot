import React, { useEffect, useState } from 'react'

import {
    mdiDatabaseCheckOutline,
    mdiDatabaseOffOutline,
    mdiDatabaseRemoveOutline,
    mdiDatabaseSyncOutline,
    mdiDatabaseClockOutline,
    mdiDatabaseAlertOutline,
} from '@mdi/js'
import classNames from 'classnames'

import { ChatContextStatus } from '@sourcegraph/cody-shared/src/chat/context'
import { RepoEmbeddingJobState } from '@sourcegraph/cody-shared/src/sourcegraph-api/graphql/client'
import { basename } from '@sourcegraph/common'

import { Icon } from '../../utils/Icon'

import styles from './ChatInputContext.module.css'

const formatFilePath = (filePath: string, selection: ChatContextStatus['selection']): string => {
    const fileName = basename(filePath)

    if (!selection) {
        return fileName
    }

    const startLine = selection.start.line + 1
    const endLine = selection.end.line + 1

    if (
        startLine === endLine ||
        (startLine + 1 === endLine && selection.end.character === 0) // A single line selected with the cursor at the start of the next line
    ) {
        return `${fileName}:${startLine}`
    }

    return `${fileName}:${startLine}-${endLine}`
}

export const ChatInputContext: React.FunctionComponent<{
    contextStatus: ChatContextStatus
    className?: string
}> = React.memo(function ChatInputContextContent({ contextStatus, className }) {
    const [completed, setCompleted] = useState(0)
    const stats = contextStatus.indexStatus?.stats
    // Only App has indexStatus output where non-App has contextStatus.connection to indicates if the repo is indexed or not
    const state =
        contextStatus.isApp && contextStatus.connection
            ? RepoEmbeddingJobState.COMPLETED
            : contextStatus.indexStatus?.state
    useEffect(() => {
        if (state === RepoEmbeddingJobState.COMPLETED) {
            return
        }
        const filesEmbedded = stats?.filesEmbedded || 0
        const filesSkipped = stats?.filesSkipped || 0
        const filesScheduled = stats?.filesScheduled || 0
        const precentage = Math.floor(((filesEmbedded + filesSkipped) / filesScheduled) * 100)
        setCompleted(precentage)
    }, [contextStatus.connection, state, stats?.filesEmbedded, stats?.filesScheduled, stats?.filesSkipped])

    const ProgressBarContainer: React.FunctionComponent<{
        filesEmbedded: number
        filesScheduled: number
        filesSkipped: number
    }> = ({ filesEmbedded, filesScheduled, filesSkipped }) => (
        <div className={styles.contextProgress}>
            <p className={styles.progressBar}>
                {/* eslint-disable-next-line react/forbid-dom-props */}
                <p className={styles.progressBarInner} style={{ width: `${completed}%` }} />
            </p>
            <p className={styles.progressInfo}>
                {filesEmbedded + filesSkipped}/{filesScheduled} files indexed
            </p>
        </div>
    )

    if (!contextStatus.codebase) {
        return <CodebaseState state="MISSING" codebase="INVALID" />
    }

    return (
        <div className={classNames(styles.container, className)}>
            {/* Index progress bar is only available to Cody App user */}
            {contextStatus.isApp &&
                state === RepoEmbeddingJobState.PROCESSING &&
                stats?.filesEmbedded &&
                stats.filesScheduled &&
                stats.filesSkipped && (
                    <ProgressBarContainer
                        filesEmbedded={stats.filesEmbedded}
                        filesScheduled={stats.filesScheduled}
                        filesSkipped={stats.filesSkipped}
                    />
                )}
            <div className={classNames(styles.footer, className)}>
                <CodebaseState
                    codebase={contextStatus.codebase}
                    state={contextStatus.connection ? 'COMPLETED' : contextStatus.indexStatus?.state || 'MISSING'}
                />
                {contextStatus.filePath && (
                    <p className={styles.file} title={contextStatus.filePath}>
                        {formatFilePath(contextStatus.filePath, contextStatus.selection)}
                    </p>
                )}
            </div>
        </div>
    )
})

const CodebaseState: React.FunctionComponent<{
    codebase: string
    state: string
    iconClassName?: string
}> = ({ state, codebase, iconClassName }) => {
    const embeddedState = ItemsByState[state as RepoEmbeddingJobState] || ItemsByState.MISSING
    const icon = embeddedState.icon
    const tooltip = embeddedState.tooltip
    const isDanger = embeddedState.type === 'danger'
    const isInfo = embeddedState.type === 'info'
    const isWarning = embeddedState.type === 'warning'
    return (
        <h3 title={tooltip} className={styles.codebase}>
            <Icon
                svgPath={icon}
                className={classNames(
                    styles.codebaseIcon,
                    iconClassName,
                    isDanger ? styles.dangerColor : isWarning ? styles.warningColor : isInfo ? styles.infoColor : null
                )}
            />
            {codebase && (
                <span className={styles.codebaseLabel}>
                    {basename(codebase.replace(/^(github|gitlab)\.com\//, ''))}
                </span>
            )}
        </h3>
    )
}

const ItemsByState = {
    CANCELED: { tooltip: 'Indexing failed', icon: mdiDatabaseRemoveOutline, type: 'danger' },
    COMPLETED: { tooltip: 'Repository is indexed and has embeddings', icon: mdiDatabaseCheckOutline, type: 'success' },
    ERRORED: { tooltip: 'Indexing failed', icon: mdiDatabaseAlertOutline, type: 'danger' },
    FAILED: { tooltip: 'Indexing failed', icon: mdiDatabaseOffOutline, type: 'danger' },
    PROCESSING: { tooltip: 'Repository is being indexed for embeddings', icon: mdiDatabaseSyncOutline, type: 'info' },
    QUEUED: { tooltip: 'Repository has been added to queue', icon: mdiDatabaseClockOutline, type: 'info' },
    MISSING: {
        tooltip: 'Repository is not indexed and has no embeddings',
        icon: mdiDatabaseOffOutline,
        type: 'warning',
    },
}

import SourcePullIcon from 'mdi-react/SourcePullIcon'
import React, { useState, useCallback } from 'react'
import { Markdown } from '../../../../../../../shared/src/components/Markdown'
import { displayRepoName } from '../../../../../../../shared/src/components/RepoFileLink'
import { renderMarkdown } from '../../../../../../../shared/src/util/markdown'
import { ThreadSettings } from '../../../settings'
import { Changeset, getChangesetExternalStatus } from '../../backend'
import { ThreadChangesetExternalStatusBadge } from './ThreadChangesetExternalStatusBadge'
import { ThreadChangesetFileDiffItem } from './ThreadChangesetFileDiffItem'
import { ChatIcon } from '../../../../../../../shared/src/components/icons'
import MessageOutlineIcon from 'mdi-react/MessageOutlineIcon'

interface Props {
    threadSettings: ThreadSettings
    changeset: Changeset
    className?: string
    headerClassName?: string
    headerStyle?: React.CSSProperties
}

/**
 * A changeset file in a thread (consisting of zero or more file diffs).
 */
export const ThreadChangesetItem: React.FunctionComponent<Props> = ({
    changeset,
    className = '',
    headerClassName = '',
    headerStyle,
}) => {
    const [isExpanded, setIsExpanded] = useState(false)
    const toggleIsExpanded = useCallback(() => setIsExpanded(!isExpanded), [isExpanded])

    const externalStatus = getChangesetExternalStatus(changeset)
    return (
        <div className={`card border ${className}`}>
            <header className={`card-header d-flex align-items-start ${headerClassName}`} style={headerStyle}>
                <div className="flex-1 d-flex align-items-center">
                    <h3 className="mb-0 h5 mr-2">
                        <span className="font-weight-bold">{displayRepoName(changeset.repo)}</span>{' '}
                        <span className="text-muted">{externalStatus.title}</span>
                    </h3>
                    <ThreadChangesetExternalStatusBadge status={externalStatus.status} className="px-2" />
                    <div className="flex-1" />
                    <div className="d-flex align-items-center text-muted small font-weight-bold">
                        <MessageOutlineIcon className="icon-inline mr-1" /> {externalStatus.commentsCount}
                    </div>
                </div>
            </header>
            <div className="card-body">
                <h4>{changeset.pullRequest.title}</h4>
                <Markdown dangerousInnerHTML={renderMarkdown(changeset.pullRequest.description)} />
            </div>
            <div className="card-body">
                <button type="button" className="btn btn-secondary btn-sm" onClick={toggleIsExpanded}>
                    {isExpanded ? 'Hide' : 'Show'} changed files
                </button>
            </div>
            {isExpanded &&
                changeset.fileDiffs.map((fileDiff, i) => <ThreadChangesetFileDiffItem key={i} fileDiff={fileDiff} />)}
        </div>
    )
}

import classNames from 'classnames'
import copy from 'copy-to-clipboard'
import ContentCopyIcon from 'mdi-react/ContentCopyIcon'
import DotsHorizontalIcon from 'mdi-react/DotsHorizontalIcon'
import FileDocumentIcon from 'mdi-react/FileDocumentIcon'
import React, { useState, useCallback } from 'react'

import { Tooltip } from '@sourcegraph/branded/src/components/tooltip/Tooltip'
import { Link } from '@sourcegraph/shared/src/components/Link'
import { pluralize } from '@sourcegraph/shared/src/util/strings'
import { useRedesignToggle } from '@sourcegraph/shared/src/util/useRedesignToggle'

import { Timestamp } from '../../components/time/Timestamp'
import { GitCommitFields } from '../../graphql-operations'
import { eventLogger } from '../../tracking/eventLogger'
import { DiffModeSelector } from '../commit/DiffModeSelector'
import { DiffMode } from '../commit/RepositoryCommitPage'

import { GitCommitNodeByline } from './GitCommitNodeByline'

export interface GitCommitNodeProps {
    node: GitCommitFields

    /** An optional additional CSS class name to apply to this element. */
    className?: string

    /** Display in a single line (more compactly). */
    compact?: boolean

    /** Expand the commit message body. */
    expandCommitMessageBody?: boolean

    /** Hide the button to expand the commit message body. */
    hideExpandCommitMessageBody?: boolean

    /** Show the full 40-character SHA and parents on their own row. */
    showSHAAndParentsRow?: boolean

    /** Fragment to show at the end to the right of the SHA. */
    afterElement?: React.ReactFragment

    /** Determine the git diff visualization UI */
    diffMode?: DiffMode

    /** Handler for change the diff mode */
    onHandleDiffMode?: (mode: DiffMode) => void
}

/** Displays a Git commit. */
export const GitCommitNode: React.FunctionComponent<GitCommitNodeProps> = ({
    node,
    afterElement,
    className,
    compact,
    expandCommitMessageBody,
    hideExpandCommitMessageBody,
    showSHAAndParentsRow,
    diffMode,
    onHandleDiffMode,
}) => {
    const [showCommitMessageBody, setShowCommitMessageBody] = useState<boolean>(false)
    const [flashCopiedToClipboardMessage, setFlashCopiedToClipboardMessage] = useState<boolean>(false)

    const toggleShowCommitMessageBody = useCallback((): void => {
        eventLogger.log('CommitBodyToggled')
        setShowCommitMessageBody(!showCommitMessageBody)
    }, [showCommitMessageBody])

    const copyToClipboard = useCallback((oid): void => {
        eventLogger.log('CommitSHACopiedToClipboard')
        copy(oid)
        setFlashCopiedToClipboardMessage(true)
        Tooltip.forceUpdate()
        setTimeout(() => {
            setFlashCopiedToClipboardMessage(false)
            Tooltip.forceUpdate()
        }, 1500)
    }, [])

    const [isRedesignEnabled] = useRedesignToggle()

    const messageElement = (
        <div
            className={classNames('git-commit-node__message flex-grow-1', compact && 'git-commit-node__message-small')}
        >
            <Link to={node.canonicalURL} className="git-commit-node__message-subject" title={node.message}>
                {node.subject}
            </Link>
            {node.body && !hideExpandCommitMessageBody && !expandCommitMessageBody && (
                <button
                    type="button"
                    className="btn btn-secondary btn-sm git-commit-node__message-toggle"
                    onClick={toggleShowCommitMessageBody}
                >
                    <DotsHorizontalIcon className="icon-inline" />
                </button>
            )}
            {compact && (
                <small className="text-muted git-commit-node__message-timestamp">
                    <Timestamp noAbout={true} date={node.committer ? node.committer.date : node.author.date} />
                </small>
            )}
        </div>
    )

    const commitMessageBody =
        expandCommitMessageBody || showCommitMessageBody ? (
            <div className="w-100">
                <pre className="git-commit-node__message-body">{node.body}</pre>
            </div>
        ) : undefined

    const bylineElement = (
        <GitCommitNodeByline
            className="d-flex text-muted git-commit-node__byline"
            author={node.author}
            committer={node.committer}
            // TODO compact needs to be always a boolean
            compact={Boolean(compact)}
            messageElement={messageElement}
            commitMessageBody={commitMessageBody}
        />
    )

    const shaDataElement = showSHAAndParentsRow && (
        <div className="w-100 git-commit-node__sha-and-parents">
            <div className={classNames('d-flex', isRedesignEnabled && 'mb-1')}>
                {isRedesignEnabled && <span className="git-commit-node__sha-and-parents-label">Commit:</span>}
                <code
                    className={classNames(
                        'git-commit-node__sha-and-parents-sha',
                        !isRedesignEnabled && 'git-ref-tag-2'
                    )}
                >
                    {node.oid}{' '}
                    <button
                        type="button"
                        className="btn btn-icon git-commit-node__sha-and-parents-copy"
                        onClick={() => copyToClipboard(node.oid)}
                        data-tooltip={flashCopiedToClipboardMessage ? 'Copied!' : 'Copy full SHA'}
                    >
                        <ContentCopyIcon className="icon-inline" />
                    </button>
                </code>
            </div>
            <div className="align-items-center d-flex">
                {node.parents.length > 0 ? (
                    <>
                        <span className="git-commit-node__sha-and-parents-label">
                            {node.parents.length === 1
                                ? 'Parent'
                                : `${node.parents.length} ${pluralize('parent', node.parents.length)}`}
                            :
                        </span>{' '}
                        {node.parents.map((parent, index) => (
                            <div className="d-flex" key={index}>
                                <Link
                                    className={classNames(
                                        'git-commit-node__sha-and-parents-parent',
                                        !isRedesignEnabled && 'git-ref-tag-2'
                                    )}
                                    to={parent.url}
                                >
                                    <code>{parent.oid}</code>
                                </Link>
                                {isRedesignEnabled && (
                                    <button
                                        type="button"
                                        className="btn btn-icon git-commit-node__sha-and-parents-copy"
                                        onClick={() => copyToClipboard(parent.oid)}
                                        data-tooltip={flashCopiedToClipboardMessage ? 'Copied!' : 'Copy full SHA'}
                                    >
                                        <ContentCopyIcon className="icon-inline" />
                                    </button>
                                )}
                            </div>
                        ))}
                    </>
                ) : (
                    '(root commit)'
                )}
            </div>
        </div>
    )

    const diffModeSelector = (): JSX.Element | null => {
        if (diffMode && onHandleDiffMode) {
            return <DiffModeSelector onHandleDiffMode={onHandleDiffMode} diffMode={diffMode} small={true} />
        }

        return null
    }

    const viewFilesCommitElement = node.tree && (
        <div className="d-flex justify-content-between">
            <Link
                className={classNames(
                    'btn btn-sm',
                    isRedesignEnabled ? 'align-center btn-outline-secondary d-inline-flex' : 'btn-secondary'
                )}
                to={node.tree.canonicalURL}
                data-tooltip="View files at this commit"
            >
                <FileDocumentIcon className={classNames('icon-inline', !isRedesignEnabled ? 'small' : 'mr-1')} />
                {isRedesignEnabled && 'View files in commit'}
            </Link>
            {isRedesignEnabled && diffModeSelector()}
        </div>
    )

    const oidElement = <code className="git-commit-node__oid">{node.abbreviatedOID}</code>
    return (
        <div
            key={node.id}
            className={`git-commit-node ${compact ? 'git-commit-node--compact' : ''} ${className || ''}`}
        >
            <>
                {!compact ? (
                    <>
                        <div
                            className={classNames(
                                'w-100 d-flex justify-content-between align-items-start',
                                !isRedesignEnabled && 'flex-wrap-reverse'
                            )}
                        >
                            <div className="git-commit-node__signature">
                                {!isRedesignEnabled && messageElement}
                                {bylineElement}
                            </div>
                            <div className="git-commit-node__actions">
                                {!showSHAAndParentsRow && (
                                    <div>
                                        <div className="btn-group btn-group-sm mr-2" role="group">
                                            <Link
                                                className="btn btn-secondary"
                                                to={node.canonicalURL}
                                                data-tooltip="View this commit"
                                            >
                                                <strong>{oidElement}</strong>
                                            </Link>
                                            <button
                                                type="button"
                                                className="btn btn-secondary"
                                                onClick={() => copyToClipboard(node.oid)}
                                                data-tooltip={
                                                    flashCopiedToClipboardMessage ? 'Copied!' : 'Copy full SHA'
                                                }
                                            >
                                                <ContentCopyIcon className="icon-inline small" />
                                            </button>
                                        </div>
                                        {node.tree && (
                                            <Link
                                                className="btn btn-sm btn-secondary"
                                                to={node.tree.canonicalURL}
                                                data-tooltip="View files at this commit"
                                            >
                                                <FileDocumentIcon className="icon-inline mr-1" />
                                            </Link>
                                        )}
                                    </div>
                                )}
                                {!isRedesignEnabled && showSHAAndParentsRow ? viewFilesCommitElement : shaDataElement}
                            </div>
                        </div>
                        <div>
                            {isRedesignEnabled && showSHAAndParentsRow ? (
                                viewFilesCommitElement
                            ) : (
                                <>
                                    {!isRedesignEnabled && commitMessageBody}
                                    {shaDataElement}
                                </>
                            )}
                        </div>
                    </>
                ) : (
                    <div>
                        <div className="w-100 d-flex justify-content-between align-items-center flex-wrap-reverse">
                            {bylineElement}
                            {messageElement}
                            <Link to={node.canonicalURL}>{oidElement}</Link>
                            {afterElement}
                        </div>
                        {commitMessageBody}
                    </div>
                )}
            </>
        </div>
    )
}

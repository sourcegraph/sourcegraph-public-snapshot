import classNames from 'classnames'
import copy from 'copy-to-clipboard'
import ContentCopyIcon from 'mdi-react/ContentCopyIcon'
import DotsHorizontalIcon from 'mdi-react/DotsHorizontalIcon'
import FileDocumentIcon from 'mdi-react/FileDocumentIcon'
import React, { useState, useCallback } from 'react'

import { Link } from '@sourcegraph/shared/src/components/Link'
import { pluralize } from '@sourcegraph/shared/src/util/strings'
import { TooltipController } from '@sourcegraph/wildcard'

import { Timestamp } from '../../components/time/Timestamp'
import { GitCommitFields } from '../../graphql-operations'
import { eventLogger } from '../../tracking/eventLogger'
import { DiffModeSelector } from '../commit/DiffModeSelector'
import { DiffMode } from '../commit/RepositoryCommitPage'

import styles from './GitCommitNode.module.scss'
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

    /** An optional additional css class name to apply this to commit node message subject */
    messageSubjectClassName?: string
}

/** Displays a Git commit. */
export const GitCommitNode: React.FunctionComponent<GitCommitNodeProps> = ({
    node,
    afterElement,
    className,
    compact,
    expandCommitMessageBody,
    hideExpandCommitMessageBody,
    messageSubjectClassName,
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
        TooltipController.forceUpdate()
        setTimeout(() => {
            setFlashCopiedToClipboardMessage(false)
            TooltipController.forceUpdate()
        }, 1500)
    }, [])

    const messageElement = (
        <div
            className={classNames('flex-grow-1', styles.message, compact && styles.messageSmall)}
            data-testid="git-commit-node-message"
        >
            <Link
                to={node.canonicalURL}
                className={classNames(messageSubjectClassName, styles.messageSubject)}
                title={node.message}
                data-testid="git-commit-node-message-subject"
            >
                {node.subject}
            </Link>
            {node.body && !hideExpandCommitMessageBody && !expandCommitMessageBody && (
                <button
                    type="button"
                    className={classNames('btn btn-secondary btn-sm', styles.messageToggle)}
                    onClick={toggleShowCommitMessageBody}
                >
                    <DotsHorizontalIcon className="icon-inline" />
                </button>
            )}
            {compact && (
                <small className={classNames('text-muted', styles.messageTimestamp)}>
                    <Timestamp noAbout={true} date={node.committer ? node.committer.date : node.author.date} />
                </small>
            )}
        </div>
    )

    const commitMessageBody =
        expandCommitMessageBody || showCommitMessageBody ? (
            <div className="w-100">
                <pre className={styles.messageBody}>{node.body}</pre>
            </div>
        ) : undefined

    const bylineElement = (
        <GitCommitNodeByline
            className={classNames('d-flex text-muted', styles.byline)}
            avatarClassName={compact ? undefined : styles.signatureUserAvatar}
            author={node.author}
            committer={node.committer}
            // TODO compact needs to be always a boolean
            compact={Boolean(compact)}
            messageElement={messageElement}
            commitMessageBody={commitMessageBody}
        />
    )

    const shaDataElement = showSHAAndParentsRow && (
        <div className={classNames('w-100', styles.shaAndParents)}>
            <div className="d-flex mb-1">
                <span className={styles.shaAndParentsLabel}>Commit:</span>
                <code className={styles.shaAndParentsSha}>
                    {node.oid}{' '}
                    <button
                        type="button"
                        className={classNames('btn btn-icon', styles.shaAndParentsCopy)}
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
                        <span className={styles.shaAndParentsLabel}>
                            {node.parents.length === 1
                                ? 'Parent'
                                : `${node.parents.length} ${pluralize('parent', node.parents.length)}`}
                            :
                        </span>{' '}
                        {node.parents.map((parent, index) => (
                            <div className="d-flex" key={index}>
                                <Link className={styles.shaAndParentsParent} to={parent.url}>
                                    <code>{parent.oid}</code>
                                </Link>
                                <button
                                    type="button"
                                    className={classNames('btn btn-icon', styles.shaAndParentsCopy)}
                                    onClick={() => copyToClipboard(parent.oid)}
                                    data-tooltip={flashCopiedToClipboardMessage ? 'Copied!' : 'Copy full SHA'}
                                >
                                    <ContentCopyIcon className="icon-inline" />
                                </button>
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
                className="btn btn-sm btn-outline-secondary align-center d-inline-flex"
                to={node.tree.canonicalURL}
                data-tooltip="Browse files in the repository at this point in history"
            >
                <FileDocumentIcon className="icon-inline mr-1" />
                Browse files at @{node.abbreviatedOID}
            </Link>
            {diffModeSelector()}
        </div>
    )

    const oidElement = (
        <code className={styles.oid} data-testid="git-commit-node-oid">
            {node.abbreviatedOID}
        </code>
    )
    return (
        <div
            key={node.id}
            className={classNames(styles.gitCommitNode, compact && styles.gitCommitNodeCompact, className)}
        >
            <>
                {!compact ? (
                    <>
                        <div className="w-100 d-flex justify-content-between align-items-start">
                            <div className={styles.signature}>{bylineElement}</div>
                            <div className={styles.actions}>
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
                                {shaDataElement}
                            </div>
                        </div>
                        <div>{showSHAAndParentsRow ? viewFilesCommitElement : shaDataElement}</div>
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

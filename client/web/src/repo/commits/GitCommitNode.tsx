import React, { useState, useCallback, useEffect } from 'react'

import classNames from 'classnames'
import copy from 'copy-to-clipboard'
import ContentCopyIcon from 'mdi-react/ContentCopyIcon'
import DotsHorizontalIcon from 'mdi-react/DotsHorizontalIcon'
import FileDocumentIcon from 'mdi-react/FileDocumentIcon'

import { pluralize } from '@sourcegraph/common'
import { Button, ButtonGroup, TooltipController, Link, Icon } from '@sourcegraph/wildcard'

import { Timestamp } from '../../components/time/Timestamp'
import { GitCommitFields } from '../../graphql-operations'
import { eventLogger } from '../../tracking/eventLogger'
import { DiffModeSelector } from '../commit/DiffModeSelector'
import { DiffMode } from '../commit/RepositoryCommitPage'

import { GitCommitNodeByline } from './GitCommitNodeByline'

import styles from './GitCommitNode.module.scss'

export interface GitCommitNodeProps {
    node: GitCommitFields

    /** An optional additional CSS class name to apply to this element. */
    className?: string

    /** Display in a single line (more compactly). */
    compact?: boolean

    /** Display in sidebar mode. */
    sidebar?: boolean

    /** Expand the commit message body. */
    expandCommitMessageBody?: boolean

    /** Hide the button to expand the commit message body. */
    hideExpandCommitMessageBody?: boolean

    /** Show the full 40-character SHA and parents on their own row. */
    showSHAAndParentsRow?: boolean

    /** Show the absolute timestamp and move relative time to tooltip. */
    preferAbsoluteTimestamps?: boolean

    /** Fragment to show at the end to the right of the SHA. */
    afterElement?: React.ReactFragment

    /** Determine the git diff visualization UI */
    diffMode?: DiffMode

    /** Handler for change the diff mode */
    onHandleDiffMode?: (mode: DiffMode) => void

    /** An optional additional css class name to apply this to commit node message subject */
    messageSubjectClassName?: string

    /**
     * Element that should wrap the commit data. Only use 'li' when rendering the component in a list
     *
     * Note: This is primarily required for support when using this component in `FilteredConnection`
     * Tracking issue to migrate away from this component: https://github.com/sourcegraph/sourcegraph/issues/23157
     * */
    wrapperElement?: 'div' | 'li'
}

/** Displays a Git commit. */
export const GitCommitNode: React.FunctionComponent<React.PropsWithChildren<GitCommitNodeProps>> = ({
    node,
    afterElement,
    className,
    compact,
    sidebar,
    expandCommitMessageBody,
    hideExpandCommitMessageBody,
    messageSubjectClassName,
    showSHAAndParentsRow,
    preferAbsoluteTimestamps,
    diffMode,
    onHandleDiffMode,
    wrapperElement: WrapperElement = 'div',
}) => {
    const [showCommitMessageBody, setShowCommitMessageBody] = useState<boolean>(false)
    const [flashCopiedToClipboardMessage, setFlashCopiedToClipboardMessage] = useState<boolean>(false)

    const toggleShowCommitMessageBody = useCallback((): void => {
        eventLogger.log('CommitBodyToggled')
        setShowCommitMessageBody(!showCommitMessageBody)
    }, [showCommitMessageBody])

    useEffect(() => {
        TooltipController.forceUpdate()
    }, [flashCopiedToClipboardMessage])

    const copyToClipboard = useCallback((oid): void => {
        eventLogger.log('CommitSHACopiedToClipboard')
        copy(oid)
        setFlashCopiedToClipboardMessage(true)

        setTimeout(() => {
            setFlashCopiedToClipboardMessage(false)
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
                <Button
                    className={styles.messageToggle}
                    onClick={toggleShowCommitMessageBody}
                    variant="secondary"
                    size="sm"
                    aria-label={showCommitMessageBody ? 'Hide commit message body' : 'Show commit message body'}
                >
                    <Icon as={DotsHorizontalIcon} />
                </Button>
            )}
            {compact && (
                <small className={classNames('text-muted', styles.messageTimestamp)}>
                    <Timestamp
                        noAbout={true}
                        preferAbsolute={preferAbsoluteTimestamps}
                        date={node.committer ? node.committer.date : node.author.date}
                    />
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
            className={classNames(styles.byline, sidebar ? 'd-flex text-muted w-50' : 'd-flex text-muted')}
            avatarClassName={compact ? undefined : styles.signatureUserAvatar}
            author={node.author}
            committer={node.committer}
            // TODO compact needs to be always a boolean
            compact={Boolean(compact)}
            messageElement={messageElement}
            commitMessageBody={commitMessageBody}
            preferAbsoluteTimestamps={preferAbsoluteTimestamps}
        />
    )

    const shaDataElement = showSHAAndParentsRow && (
        <div className={classNames('w-100', styles.shaAndParents)}>
            <div className="d-flex mb-1">
                <span className={styles.shaAndParentsLabel}>Commit:</span>
                <code className={styles.shaAndParentsSha}>
                    {node.oid}{' '}
                    <Button
                        variant="icon"
                        className={styles.shaAndParentsCopy}
                        onClick={() => copyToClipboard(node.oid)}
                        data-tooltip={flashCopiedToClipboardMessage ? 'Copied!' : 'Copy full SHA'}
                    >
                        <Icon as={ContentCopyIcon} />
                    </Button>
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
                                <Button
                                    variant="icon"
                                    className={styles.shaAndParentsCopy}
                                    onClick={() => copyToClipboard(parent.oid)}
                                    data-tooltip={flashCopiedToClipboardMessage ? 'Copied!' : 'Copy full SHA'}
                                >
                                    <Icon as={ContentCopyIcon} />
                                </Button>
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
            <Button
                className="align-center d-inline-flex"
                to={node.tree.canonicalURL}
                data-tooltip="Browse files in the repository at this point in history"
                variant="secondary"
                outline={true}
                size="sm"
                as={Link}
            >
                <Icon className="mr-1" as={FileDocumentIcon} />
                Browse files at @{node.abbreviatedOID}
            </Button>
            {diffModeSelector()}
        </div>
    )

    const oidElement = (
        <code className={styles.oid} data-testid="git-commit-node-oid">
            {node.abbreviatedOID}
        </code>
    )

    if (sidebar) {
        return (
            <WrapperElement
                key={node.id}
                className={classNames(styles.gitCommitNode, styles.gitCommitNodeCompact, className)}
            >
                <div className="w-100 d-flex justify-content-between align-items-center flex-wrap-reverse">
                    {bylineElement}
                    <small className={classNames('text-muted', styles.messageTimestamp)}>
                        <Timestamp
                            noAbout={true}
                            preferAbsolute={preferAbsoluteTimestamps}
                            date={node.committer ? node.committer.date : node.author.date}
                        />
                    </small>
                    <Link to={node.canonicalURL}>{oidElement}</Link>
                </div>
            </WrapperElement>
        )
    }

    return (
        <WrapperElement
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
                                        <ButtonGroup className="mr-2">
                                            <Button
                                                to={node.canonicalURL}
                                                data-tooltip="View this commit"
                                                variant="secondary"
                                                as={Link}
                                                size="sm"
                                                aria-label="View this commit"
                                            >
                                                <strong>{oidElement}</strong>
                                            </Button>
                                            <Button
                                                onClick={() => copyToClipboard(node.oid)}
                                                data-tooltip={
                                                    flashCopiedToClipboardMessage ? 'Copied!' : 'Copy full SHA'
                                                }
                                                variant="secondary"
                                                size="sm"
                                                aria-label="Copy full SHA"
                                            >
                                                <Icon className="small" as={ContentCopyIcon} />
                                            </Button>
                                        </ButtonGroup>
                                        {node.tree && (
                                            <Button
                                                to={node.tree.canonicalURL}
                                                data-tooltip="View files at this commit"
                                                variant="secondary"
                                                size="sm"
                                                as={Link}
                                                aria-label="View files at this commit"
                                            >
                                                <Icon className="mr-1" as={FileDocumentIcon} />
                                            </Button>
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
        </WrapperElement>
    )
}

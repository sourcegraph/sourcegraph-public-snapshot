import React, { useState, useCallback } from 'react'

import { mdiDotsHorizontal, mdiContentCopy, mdiFileDocument } from '@mdi/js'
import classNames from 'classnames'
import copy from 'copy-to-clipboard'

import { Timestamp } from '@sourcegraph/branded/src/components/Timestamp'
import { pluralize } from '@sourcegraph/common'
import { Button, ButtonGroup, Link, Icon, Code, screenReaderAnnounce, Tooltip } from '@sourcegraph/wildcard'

import { eventLogger } from '../../tracking/eventLogger'
import { CommitMessageWithLinks } from '../commit/CommitMessageWithLinks'
import { DiffModeSelector } from '../commit/DiffModeSelector'
import { GitCommitNodeProps } from '../commits/GitCommitNode'
import { Linkified } from '../linkifiy/Linkified'

import { PerforceDepotChangelistNodeByline } from './PerforceDepotChangelistNodeByline'

import styles from './PerforceDepotChangelist.module.scss'

/** Displays a Perforce changelist. */
export const PerforceDepotChangelistNode: React.FunctionComponent<React.PropsWithChildren<GitCommitNodeProps>> = ({
    node,
    afterElement,
    className,
    compact,
    extraCompact,
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

    const toggleShowChangelistMessageBody = useCallback((): void => {
        eventLogger.log('ChangelistBodyToggled')
        setShowCommitMessageBody(!showCommitMessageBody)
    }, [showCommitMessageBody])

    const copyToClipboard = useCallback((oid: string): void => {
        eventLogger.log('ChangelistIDCopiedToClipboard')
        copy(oid)
        setFlashCopiedToClipboardMessage(true)
        screenReaderAnnounce('Copied!')

        setTimeout(() => {
            setFlashCopiedToClipboardMessage(false)
        }, 1500)
    }, [])

    if (extraCompact) {
        // Implied by extraCompact
        compact = true
    }

    const messageElement = (
        <div
            className={classNames(
                !extraCompact && 'flex-grow-1',
                styles.message,
                compact && styles.messageSmall,
                extraCompact && styles.messageExtraSmall
            )}
            data-testid="git-commit-node-message"
        >
            <span className={classNames('mr-2', styles.messageSubject)}>
                <CommitMessageWithLinks
                    to={node.canonicalURL}
                    className={classNames(messageSubjectClassName, styles.messageLink)}
                    message={node.subject}
                    externalURLs={node.externalURLs}
                />
            </span>

            {node.body && !hideExpandCommitMessageBody && !expandCommitMessageBody && (
                <Button
                    className={styles.messageToggle}
                    onClick={toggleShowChangelistMessageBody}
                    variant="secondary"
                    size="sm"
                    aria-label={showCommitMessageBody ? 'Hide changelist message body' : 'Show changelist message body'}
                >
                    <Icon aria-hidden={true} svgPath={mdiDotsHorizontal} />
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
                <pre className={styles.messageBody}>
                    {node.body && <Linkified input={node.body} externalURLs={node.externalURLs} />}
                </pre>
            </div>
        ) : undefined

    const bylineElement = (
        <PerforceDepotChangelistNodeByline
            className={classNames(styles.byline, 'd-flex text-muted')}
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
                <Code className={styles.shaAndParentsSha}>
                    {node.oid}{' '}
                    <Tooltip content={flashCopiedToClipboardMessage ? 'Copied!' : 'Copy changelist ID'}>
                        <Button
                            variant="icon"
                            className={styles.shaAndParentsCopy}
                            onClick={() => copyToClipboard(node.oid)}
                            aria-label="Copy changelist ID"
                        >
                            <Icon aria-hidden={true} svgPath={mdiContentCopy} />
                        </Button>
                    </Tooltip>
                </Code>
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
                        {node.parents.map(parent => (
                            <div className="d-flex" key={parent.oid}>
                                <Link className={styles.shaAndParentsParent} to={parent.url}>
                                    <Code>{parent.oid}</Code>
                                </Link>
                                <Tooltip content={flashCopiedToClipboardMessage ? 'Copied!' : 'Copy changelist ID'}>
                                    <Button
                                        variant="icon"
                                        className={styles.shaAndParentsCopy}
                                        onClick={() => copyToClipboard(parent.oid)}
                                        aria-label="Copy full SHA"
                                    >
                                        <Icon aria-hidden={true} svgPath={mdiContentCopy} />
                                    </Button>
                                </Tooltip>
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
        <div className="d-flex justify-content-between align-items-start">
            <Tooltip content="Browse files in the repository at this point in history">
                <Button
                    className="align-center d-inline-flex"
                    to={node.tree.canonicalURL}
                    variant="secondary"
                    outline={true}
                    size="sm"
                    as={Link}
                >
                    <Icon className="mr-1" aria-hidden={true} svgPath={mdiFileDocument} />
                    Browse files at @{node.abbreviatedOID}
                </Button>
            </Tooltip>
            {diffModeSelector()}
        </div>
    )

    const oidElement = (
        <Code className={styles.oid} data-testid="git-commit-node-oid">
            {node.abbreviatedOID}
        </Code>
    )

    return (
        <WrapperElement
            key={node.id}
            className={classNames(
                styles.gitCommitNode,
                compact && styles.gitCommitNodeCompact,
                extraCompact && styles.gitCommitNodeExtraCompact,
                className
            )}
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
                                            <Tooltip content="View this changelist">
                                                <Button to={node.canonicalURL} variant="secondary" as={Link} size="sm">
                                                    <strong>{oidElement}</strong>
                                                </Button>
                                            </Tooltip>
                                            <Tooltip
                                                content={
                                                    flashCopiedToClipboardMessage ? 'Copied!' : 'Copy changelist ID'
                                                }
                                            >
                                                <Button
                                                    onClick={() => copyToClipboard(node.oid)}
                                                    variant="secondary"
                                                    size="sm"
                                                    aria-label="Copy changelist ID"
                                                >
                                                    <Icon
                                                        className="small"
                                                        aria-hidden={true}
                                                        svgPath={mdiContentCopy}
                                                    />
                                                </Button>
                                            </Tooltip>
                                        </ButtonGroup>
                                        {node.tree && (
                                            <Tooltip content="View files at this changelist">
                                                <Button
                                                    aria-label="View files"
                                                    to={node.tree.canonicalURL}
                                                    variant="secondary"
                                                    size="sm"
                                                    as={Link}
                                                >
                                                    <Icon
                                                        className="mr-1"
                                                        aria-hidden={true}
                                                        svgPath={mdiFileDocument}
                                                    />
                                                </Button>
                                            </Tooltip>
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
                        <div className={styles.innerWrapper}>
                            {bylineElement}
                            {messageElement}
                            {!extraCompact && <Link to={node.canonicalURL}>{oidElement}</Link>}
                            {afterElement}
                        </div>
                        {commitMessageBody}
                    </div>
                )}
            </>
        </WrapperElement>
    )
}

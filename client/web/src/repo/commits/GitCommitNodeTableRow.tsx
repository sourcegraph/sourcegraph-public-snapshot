import React, { useState, useCallback } from 'react'

import { mdiChevronUp, mdiChevronDown } from '@mdi/js'
import classNames from 'classnames'

import { Timestamp } from '@sourcegraph/branded/src/components/Timestamp'
import { EVENT_LOGGER } from '@sourcegraph/shared/src/telemetry/web/eventLogger'
import { Button, Link, Icon, Code } from '@sourcegraph/wildcard'

import { CommitMessageWithLinks } from '../commit/CommitMessageWithLinks'
import { Linkified } from '../linkifiy/Linkified'
import { isPerforceChangelistMappingEnabled } from '../utils'

import type { GitCommitNodeProps } from './GitCommitNode'
import { GitCommitNodeByline } from './GitCommitNodeByline'

import styles from './GitCommitNode.module.scss'

export const GitCommitNodeTableRow: React.FC<
    Omit<
        GitCommitNodeProps,
        | 'wrapperElement'
        | 'afterElement'
        | 'preferAbsoluteTimestamps'
        | 'showSHAAndParentsRow'
        | 'onHandleDiffMode'
        | 'diffMode'
    >
> = ({
    node,
    className,
    expandCommitMessageBody,
    hideExpandCommitMessageBody,
    messageSubjectClassName,
    telemetryRecorder,
}) => {
    const [showCommitMessageBody, setShowCommitMessageBody] = useState<boolean>(false)

    const toggleShowCommitMessageBody = useCallback((): void => {
        EVENT_LOGGER.log('CommitBodyToggled')
        telemetryRecorder.recordEvent('repo.commit.body', 'toggle')
        setShowCommitMessageBody(!showCommitMessageBody)
    }, [showCommitMessageBody, telemetryRecorder])

    const canonicalURL =
        isPerforceChangelistMappingEnabled() && node.perforceChangelist?.canonicalURL
            ? node.perforceChangelist.canonicalURL
            : node.canonicalURL

    const messageElement = (
        <div className={classNames(styles.message, styles.messageSmall)} data-testid="git-commit-node-message">
            <span className={classNames('mr-2', styles.messageSubject)}>
                <CommitMessageWithLinks
                    to={canonicalURL}
                    className={classNames(messageSubjectClassName, styles.messageLink)}
                    message={node.subject}
                    externalURLs={node.externalURLs}
                />
            </span>
            {node.body && !hideExpandCommitMessageBody && !expandCommitMessageBody && (
                <Button
                    className={styles.messageToggle}
                    onClick={toggleShowCommitMessageBody}
                    variant="secondary"
                    size="sm"
                    aria-label={showCommitMessageBody ? 'Hide commit message body' : 'Show commit message body'}
                >
                    <Icon aria-hidden={true} svgPath={showCommitMessageBody ? mdiChevronUp : mdiChevronDown} />
                </Button>
            )}

            <small className={classNames('text-muted', styles.messageTimestamp)}>
                <Timestamp noAbout={true} date={node.committer ? node.committer.date : node.author.date} />
            </small>
        </div>
    )

    const commitMessageBody =
        expandCommitMessageBody || showCommitMessageBody ? (
            <tr className={classNames(styles.commitMessage, className)}>
                <td className={classNames(styles.colByline)} />
                <td>
                    <div className={`${styles.messageBody} flex-1`}>
                        {node.body && <Linkified input={node.body} externalURLs={node.externalURLs} />}
                    </div>
                </td>
                <td className={classNames(styles.spacer)} />
            </tr>
        ) : undefined

    return (
        <>
            <tr
                className={classNames(styles.tableRow, 'px-1', className, {
                    [styles.tableRowOpen]: commitMessageBody !== undefined,
                })}
            >
                <GitCommitNodeByline
                    as="td"
                    className={classNames('d-flex', styles.colByline)}
                    avatarClassName={styles.fontWeightNormal}
                    author={node.author}
                    committer={node.committer}
                    compact={true}
                />
                <td className="flex-1 overflow-hidden">{messageElement}</td>
                <td className="text-right">
                    <Link to={canonicalURL}>
                        <Code data-testid="git-commit-node-oid">
                            {node.perforceChangelist?.cid ?? node.abbreviatedOID}
                        </Code>
                    </Link>
                </td>
            </tr>
            {commitMessageBody}
        </>
    )
}

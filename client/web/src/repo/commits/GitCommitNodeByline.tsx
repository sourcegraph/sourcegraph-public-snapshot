import React from 'react'

import classNames from 'classnames'

import { Timestamp } from '@sourcegraph/branded/src/components/Timestamp'
import { UserAvatar } from '@sourcegraph/shared/src/components/UserAvatar'
import { Tooltip } from '@sourcegraph/wildcard'

import type { SignatureFields } from '../../graphql-operations'
import { formatPersonName, PersonLink } from '../../person/PersonLink'

interface Props {
    author: SignatureFields
    committer: SignatureFields | null
    className?: string
    avatarClassName?: string
    compact?: boolean
    preferAbsoluteTimestamps?: boolean
    messageElement?: JSX.Element
    commitMessageBody?: JSX.Element
    isPerforceDepot?: boolean
    as?: 'div' | 'td'
}

/**
 * Displays a Git commit's author and committer (with avatars if available) and the dates.
 */
export const GitCommitNodeByline: React.FunctionComponent<React.PropsWithChildren<Props>> = ({
    author,
    committer,
    className = '',
    avatarClassName,
    compact,
    preferAbsoluteTimestamps,
    messageElement,
    commitMessageBody,
    isPerforceDepot,
    as = 'div',
}) => {
    const Wrapper = as

    const refActionType = isPerforceDepot ? 'submitted by' : 'committed by'

    // Omit GitHub as committer to reduce noise. (Edits and squash commits made in the GitHub UI
    // include GitHub as a committer.)

    if (committer && committer.person.name === 'GitHub' && committer.person.email === 'noreply@github.com') {
        committer = null
    }

    if (
        committer &&
        committer.person.email !== author.person.email &&
        ((!committer.person.name && !author.person.name) || committer.person.name !== author.person.name)
    ) {
        // The author and committer both exist and are different people.
        return (
            <Wrapper data-testid="git-commit-node-byline" className={className}>
                <div className="flex-shrink-0">
                    <Tooltip content={`${formatPersonName(author.person)} (author)`}>
                        <UserAvatar inline={true} className={avatarClassName} user={author.person} />
                    </Tooltip>{' '}
                    <Tooltip content={`${formatPersonName(committer.person)} (committer)`}>
                        <UserAvatar
                            inline={true}
                            className={classNames('mr-2', avatarClassName)}
                            user={committer.person}
                        />
                    </Tooltip>
                </div>
                <div className="overflow-hidden">
                    {!compact ? (
                        <>
                            {messageElement}
                            <PersonLink person={author.person} className="font-weight-bold" /> authored and{' '}
                            {refActionType} <PersonLink person={committer.person} className="font-weight-bold" />{' '}
                            <Timestamp date={committer.date} preferAbsolute={preferAbsoluteTimestamps} />
                            {commitMessageBody}
                        </>
                    ) : (
                        <>
                            <PersonLink person={author.person} /> and <PersonLink person={committer.person} />{' '}
                        </>
                    )}
                </div>
            </Wrapper>
        )
    }

    return (
        <Wrapper data-testid="git-commit-node-byline" className={className}>
            <div>
                <Tooltip content={formatPersonName(author.person)}>
                    <UserAvatar
                        inline={true}
                        className={classNames('mr-1 mr-2', avatarClassName)}
                        user={author.person}
                    />
                </Tooltip>
            </div>
            <div className="text-truncate">
                {!compact && (
                    <>
                        {messageElement}
                        {refActionType} <PersonLink person={author.person} className="font-weight-bold" />{' '}
                        <Timestamp date={author.date} preferAbsolute={preferAbsoluteTimestamps} />
                        {commitMessageBody}
                    </>
                )}
                {compact && <PersonLink person={author.person} />}
            </div>
        </Wrapper>
    )
}

import classNames from 'classnames'
import React from 'react'

import { useRedesignToggle } from '@sourcegraph/shared/src/util/useRedesignToggle'

import { Timestamp } from '../../components/time/Timestamp'
import { SignatureFields } from '../../graphql-operations'
import { formatPersonName, PersonLink } from '../../person/PersonLink'
import { UserAvatar } from '../../user/UserAvatar'

interface Props {
    author: SignatureFields
    committer: SignatureFields | null
    className?: string
    compact?: boolean
    messageElement?: JSX.Element
    commitMessageBody?: JSX.Element
}

/**
 * Displays a Git commit's author and committer (with avatars if available) and the dates.
 */
export const GitCommitNodeByline: React.FunctionComponent<Props> = ({
    author,
    committer,
    className = '',
    compact,
    messageElement,
    commitMessageBody,
}) => {
    const [isRedesignEnabled] = useRedesignToggle()

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
        if (!isRedesignEnabled) {
            // The author and committer both exist and are different people.
            return (
                <small data-testid="git-commit-node-byline" className={className}>
                    <UserAvatar
                        className="icon-inline-md"
                        user={author.person}
                        data-tooltip={`${formatPersonName(author.person)} (author)`}
                    />{' '}
                    <UserAvatar
                        className="icon-inline mr-1"
                        user={committer.person}
                        data-tooltip={`${formatPersonName(committer.person)} (committer)`}
                    />{' '}
                    <PersonLink person={author.person} className="font-weight-bold" /> {!compact && 'authored'} and{' '}
                    <PersonLink person={committer.person} className="font-weight-bold" />{' '}
                    {!compact && (
                        <>
                            committed <Timestamp date={committer.date} />
                        </>
                    )}
                </small>
            )
        }

        return (
            <div data-testid="git-commit-node-byline" className={className}>
                <div>
                    <UserAvatar
                        className="icon-inline"
                        user={author.person}
                        data-tooltip={`${formatPersonName(author.person)} (author)`}
                    />{' '}
                    <UserAvatar
                        className="icon-inline mr-2"
                        user={committer.person}
                        data-tooltip={`${formatPersonName(committer.person)} (committer)`}
                    />
                </div>
                <div>
                    {!compact ? (
                        <>
                            {messageElement}
                            <PersonLink person={author.person} className="font-weight-bold" /> authored and{' '}
                            <PersonLink person={committer.person} className="font-weight-bold" /> commited{' '}
                            <Timestamp date={author.date} />
                            {commitMessageBody}
                        </>
                    ) : (
                        <>
                            <PersonLink person={author.person} className="font-weight-bold" /> and{' '}
                            <PersonLink person={committer.person} className="font-weight-bold" />{' '}
                        </>
                    )}
                </div>
            </div>
        )
    }

    return (
        <div data-testid="git-commit-node-byline" className={className}>
            <div>
                <UserAvatar
                    className={classNames('icon-inline mr-1', isRedesignEnabled && 'mr-2')}
                    user={author.person}
                    data-tooltip={formatPersonName(author.person)}
                />
            </div>
            <div>
                {!compact && !isRedesignEnabled && (
                    <>
                        <PersonLink person={author.person} className="font-weight-bold" /> committed{' '}
                        <Timestamp date={author.date} />
                    </>
                )}
                {!compact && isRedesignEnabled && (
                    <>
                        {messageElement}
                        committed by <PersonLink person={author.person} className="font-weight-bold" />{' '}
                        <Timestamp date={author.date} />
                        {commitMessageBody}
                    </>
                )}
                {compact && (
                    <>
                        <PersonLink person={author.person} className="font-weight-bold" />{' '}
                    </>
                )}
            </div>
        </div>
    )
}

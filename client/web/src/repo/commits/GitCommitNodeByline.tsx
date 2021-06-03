import classNames from 'classnames'
import React from 'react'

import { useRedesignToggle } from '@sourcegraph/shared/src/util/useRedesignToggle'

import { Timestamp } from '../../components/time/Timestamp'
import { SignatureFields } from '../../graphql-operations'
import { formatPersonName, PersonLink } from '../../person/PersonLink'
import { UserAvatar } from '../../user/UserAvatar'

import styles from './GitCommitNodeByLine.module.scss'

interface Props {
    author: SignatureFields
    committer: SignatureFields | null
    className?: string
    compact?: boolean
}

/**
 * Displays a Git commit's author and committer (with avatars if available) and the dates.
 */
export const GitCommitNodeByline: React.FunctionComponent<Props> = ({ author, committer, className = '', compact }) => {
    const [isRedesignEnabled] = useRedesignToggle()

    const avatarMarginClass = isRedesignEnabled && compact ? 'mr-2' : 'mr-1'
    const avatarClassName = classNames('icon-inline mr-1', compact && styles.avatarCompact)

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
            <small data-testid="git-commit-node-byline" className={className}>
                <UserAvatar
                    className={avatarClassName}
                    user={author.person}
                    data-tooltip={`${formatPersonName(author.person)} (author)`}
                />
                <UserAvatar
                    className={classNames(avatarClassName, avatarMarginClass)}
                    user={committer.person}
                    data-tooltip={`${formatPersonName(committer.person)} (committer)`}
                />
                <span className={classNames(compact && styles.personCompact)}>
                    <PersonLink person={author.person} className="font-weight-bold" /> {!compact && 'authored'} and{' '}
                    <PersonLink person={committer.person} className="font-weight-bold" />{' '}
                    {!compact && (
                        <>
                            committed <Timestamp date={committer.date} />
                        </>
                    )}
                </span>
            </small>
        )
    }

    return (
        <small data-testid="git-commit-node-byline" className={className}>
            <UserAvatar
                className={classNames(avatarClassName, avatarMarginClass)}
                user={author.person}
                data-tooltip={formatPersonName(author.person)}
            />{' '}
            <span className={classNames(compact && styles.personCompact)}>
                <PersonLink person={author.person} className="font-weight-bold" />{' '}
                {!compact && (
                    <>
                        committed <Timestamp date={author.date} />
                    </>
                )}
            </span>
        </small>
    )
}

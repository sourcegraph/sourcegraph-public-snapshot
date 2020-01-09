import React from 'react'
import * as GQL from '../../../../shared/src/graphql/schema'
import { Timestamp } from '../../components/time/Timestamp'
import { UserAvatar } from '../../user/UserAvatar'
import { formatPersonName, PersonLink } from '../../person/PersonLink'

/**
 * The subset of {@link GQL.ISignature} information needed by {@link GitCommitNodeByline}. Using the
 * minimal subset makes testing easier.
 */
interface Signature extends Pick<GQL.ISignature, 'date'> {
    person: {
        user: Pick<GQL.IUser, 'username' | 'displayName' | 'url'> | null
    } & Pick<GQL.IPerson, 'email' | 'name' | 'displayName' | 'avatarURL'>
}

/**
 * Displays a Git commit's author and committer (with avatars if available) and the dates.
 */
export const GitCommitNodeByline: React.FunctionComponent<{
    author: GQL.ISignature | Signature
    committer: GQL.ISignature | Signature | null
    className?: string
    compact?: boolean
}> = ({ author, committer, className = '', compact }) => {
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
            <small className={`git-commit-node-byline git-commit-node-byline--has-committer ${className}`}>
                <UserAvatar
                    className="icon-inline"
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
        <small className={`git-commit-node-byline git-commit-node-byline--no-committer ${className}`}>
            <UserAvatar
                className="icon-inline mr-1"
                user={author.person}
                data-tooltip={formatPersonName(author.person)}
            />{' '}
            <PersonLink person={author.person} className="font-weight-bold" />{' '}
            {!compact && (
                <>
                    committed <Timestamp date={author.date} />
                </>
            )}
        </small>
    )
}

import React from 'react'
import * as GQL from '../../../../shared/src/graphql/schema'
import { Timestamp } from '../../components/time/Timestamp'
import { UserAvatar } from '../../user/UserAvatar'

/**
 * The subset of {@link GQL.ISignature} information needed by {@link GitCommitNodeByline}. Using the
 * minimal subset makes testing easier.
 */
export interface Signature extends Pick<GQL.ISignature, 'date'> {
    person: {
        user: Pick<GQL.IUser, 'username'> | null
    } & Pick<GQL.IPerson, 'email' | 'name' | 'displayName' | 'avatarURL'>
}

/**
 * Formats person names to: "username (Display Name)" or "Display Name"
 */
const formatPersonNames = ({ user, displayName }: Signature['person']): string =>
    user ? `${user.username} (${displayName})` : displayName

/**
 * Formats person names with {@link formatPersonNames}, and shows tooltip with user email.
 */
const PersonNames: React.FunctionComponent<{ person: Signature['person'] }> = ({ person }) => (
    <strong data-tooltip={person.email}>{formatPersonNames(person)}</strong>
)

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
                    data-tooltip={`${formatPersonNames(author.person)} (author)`}
                />{' '}
                <UserAvatar
                    className="icon-inline mr-1"
                    user={committer.person}
                    data-tooltip={`${formatPersonNames(committer.person)} (committer)`}
                />{' '}
                <PersonNames person={author.person} /> {!compact && 'authored'} and{' '}
                <PersonNames person={committer.person} />{' '}
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
                data-tooltip={formatPersonNames(author.person)}
            />{' '}
            <PersonNames person={author.person} />{' '}
            {!compact && (
                <>
                    committed <Timestamp date={author.date} />
                </>
            )}
        </small>
    )
}

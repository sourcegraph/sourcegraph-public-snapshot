import React from 'react'
import * as GQL from '../../../../shared/src/graphql/schema'
import { Timestamp } from '../../components/time/Timestamp'
import { UserAvatar } from '../../user/UserAvatar'

/**
 * Displays a Git commit's author and committer (with avatars if available) and the dates.
 */
export const GitCommitNodeByline: React.FunctionComponent<{
    author: GQL.ISignature
    committer: GQL.ISignature | null
    className: string
    compact: boolean
}> = ({ author, committer, className, compact }) => {
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
                    data-tooltip={`${author.person.displayName} (author)`}
                />{' '}
                <UserAvatar
                    className="icon-inline mr-1"
                    user={committer.person}
                    data-tooltip={`${committer.person.displayName} (committer)`}
                />{' '}
                <strong>{author.person.displayName}</strong> {!compact && 'authored'} and{' '}
                <strong>{committer.person.displayName}</strong>{' '}
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
            <UserAvatar className="icon-inline mr-1" user={author.person} data-tooltip={author.person.displayName} />{' '}
            <strong>{author.person.displayName}</strong>{' '}
            {!compact && (
                <>
                    committed <Timestamp date={author.date} />
                </>
            )}
        </small>
    )
}

import type { FC } from 'react'

import { Timestamp } from '@sourcegraph/branded/src/components/Timestamp'
import { Link } from '@sourcegraph/wildcard'

import type { Maybe, UserAreaUserFields } from '../../graphql-operations'

type UserData = Maybe<Pick<UserAreaUserFields, 'url' | 'username'>>

interface BylineProps {
    createdAt: string
    createdBy?: UserData
    updatedAt: Maybe<string>
    updatedBy?: UserData
    noAuthor?: boolean
}

/**
 * The created/updated byline containing information about creator, creation date, updater and update date.
 */
export const CreatedByAndUpdatedByInfoByline: FC<BylineProps> = ({
    createdAt,
    createdBy,
    updatedAt,
    updatedBy,
    noAuthor,
}) => {
    const createdByPart = noAuthor ?? (
        <> by {createdBy ? <Link to={createdBy.url}>{createdBy.username}</Link> : 'a deleted user'}</>
    )
    const updatedPart = (
        <>
            {updatedAt !== null && updatedAt !== createdAt && (
                <>
                    <span className="mx-2">|</span>
                    Updated <Timestamp date={updatedAt} />
                    {updatedBy?.username !== createdBy?.username && (
                        <> by {updatedBy ? <Link to={updatedBy.url}>{updatedBy.username}</Link> : 'a deleted user'}</>
                    )}
                </>
            )}
        </>
    )
    return (
        <>
            Created <Timestamp date={createdAt} /> {createdByPart} {updatedPart}
        </>
    )
}

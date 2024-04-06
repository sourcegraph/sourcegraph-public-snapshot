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
    type?: string
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
    type,
}) => {
    const createdByPart = noAuthor ? null : type === 'ExternalService' ? (
        createdBy ? (
            <>
                {' '}
                by <Link to={createdBy.url}>{createdBy.username}</Link>
            </>
        ) : null
    ) : (
        <> by {createdBy ? <Link to={createdBy.url}>{createdBy.username}</Link> : 'a deleted user'}</>
    )

    const updatedPart = (
        <>
            {updatedAt !== null && updatedAt !== createdAt && (
                <>
                    <span className="mx-2">|</span>
                    {type === 'ExternalService' ? (
                        <>
                            {updatedBy?.username && (
                                <>
                                    Updated by <Link to={updatedBy.url}>{updatedBy.username}</Link>
                                    <span className="mx-2">|</span>
                                </>
                            )}
                            <>
                                Last synced <Timestamp date={updatedAt} />
                            </>
                        </>
                    ) : (
                        <>
                            Updated <Timestamp date={updatedAt} />
                            {updatedBy?.username !== createdBy?.username && (
                                <>
                                    by{' '}
                                    {updatedBy ? (
                                        <Link to={updatedBy.url}>{updatedBy.username}</Link>
                                    ) : (
                                        'a deleted user'
                                    )}
                                </>
                            )}
                        </>
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

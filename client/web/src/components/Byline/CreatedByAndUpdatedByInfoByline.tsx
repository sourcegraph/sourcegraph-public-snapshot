import { FC } from 'react'

import { Link } from '@sourcegraph/wildcard'

import { Maybe, UserAreaUserFields } from '../../graphql-operations'
import { Timestamp } from '../time/Timestamp'

type UserData = Maybe<Pick<UserAreaUserFields, 'url' | 'username'>>

interface BylineProps {
    createdAt: string
    createdBy: UserData
    updatedAt: Maybe<string>
    updatedBy: UserData
}

/**
 * The created/updated byline containing information about creator, creation date, updater and update date.
 */
export const CreatedByAndUpdatedByInfoByline: FC<BylineProps> = ({ createdAt, createdBy, updatedAt, updatedBy }) => (
    <>
        Created <Timestamp date={createdAt} /> by{' '}
        {createdBy ? <Link to={createdBy.url}>{createdBy.username}</Link> : 'a deleted user'}
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

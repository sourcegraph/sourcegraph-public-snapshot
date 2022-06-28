import React from 'react'

import { Link } from '@sourcegraph/wildcard'

import { Timestamp } from '../../../components/time/Timestamp'
import { BatchChangeFields } from '../../../graphql-operations'

interface Props extends Pick<BatchChangeFields, 'createdAt' | 'creator' | 'lastAppliedAt' | 'lastApplier'> {}

/**
 * The created/updated byline in the batch change header.
 */
export const BatchChangeInfoByline: React.FunctionComponent<React.PropsWithChildren<Props>> = ({
    createdAt,
    creator,
    lastAppliedAt,
    lastApplier,
}) => (
    <>
        Created <Timestamp date={createdAt} /> by{' '}
        {creator ? <Link to={creator.url}>{creator.username}</Link> : 'a deleted user'}
        {lastAppliedAt !== null && lastAppliedAt !== createdAt && (
            <>
                <span className="mx-2">|</span>
                Updated <Timestamp date={lastAppliedAt} />
                {lastApplier?.username !== creator?.username && (
                    <> by {lastApplier ? <Link to={lastApplier.url}>{lastApplier.username}</Link> : 'a deleted user'}</>
                )}
            </>
        )}
    </>
)

import React from 'react'

import { RouterLink } from '@sourcegraph/wildcard'

import { Timestamp } from '../../../components/time/Timestamp'
import { BatchChangeFields } from '../../../graphql-operations'

interface Props extends Pick<BatchChangeFields, 'createdAt' | 'creator' | 'lastAppliedAt' | 'lastApplier'> {}

/**
 * The created/updated byline in the batch change header.
 */
export const BatchChangeInfoByline: React.FunctionComponent<Props> = ({
    createdAt,
    creator,
    lastAppliedAt,
    lastApplier,
}) => (
    <>
        Created <Timestamp date={createdAt} /> by{' '}
        {creator ? <RouterLink to={creator.url}>{creator.username}</RouterLink> : 'a deleted user'}
        {lastAppliedAt !== null && lastAppliedAt !== createdAt && (
            <>
                <span className="mx-2">|</span>
                Updated <Timestamp date={lastAppliedAt} />
                {lastApplier?.username !== creator?.username && (
                    <>
                        {' '}
                        by{' '}
                        {lastApplier ? (
                            <RouterLink to={lastApplier.url}>{lastApplier.username}</RouterLink>
                        ) : (
                            'a deleted user'
                        )}
                    </>
                )}
            </>
        )}
    </>
)

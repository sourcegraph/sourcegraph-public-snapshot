import React from 'react'

import { RouterLink } from '@sourcegraph/wildcard'

import { Timestamp } from '../../../components/time/Timestamp'
import { BatchSpecFields } from '../../../graphql-operations'

interface Props extends Pick<BatchSpecFields, 'createdAt' | 'creator'> {}

/**
 * The uploaded at byline to the batch change apply page header.
 */
export const BatchSpecInfoByline: React.FunctionComponent<Props> = ({ createdAt, creator }) => (
    <>
        Uploaded <Timestamp date={createdAt} /> by{' '}
        {creator && <RouterLink to={creator.url}>{creator.username}</RouterLink>}
        {!creator && <strong>deleted user</strong>}
    </>
)

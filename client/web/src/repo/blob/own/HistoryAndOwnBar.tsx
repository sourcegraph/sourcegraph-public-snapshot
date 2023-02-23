import React from 'react'

import { useQuery } from '@sourcegraph/http-client'
import { Button, LoadingSpinner } from '@sourcegraph/wildcard'

import { FetchOwnersAndHistoryResult, FetchOwnersAndHistoryVariables } from '../../../graphql-operations'
import { GitCommitNode } from '../../commits/GitCommitNode'

import { FETCH_OWNERS_AND_HISTORY } from './grapqlQueries'

import styles from './HistoryAndOwnBar.module.scss'

export const HistoryAndOwnBar: React.FunctionComponent<{
    repoID: string
    revision?: string
    filePath: string
}> = ({ repoID, revision, filePath }) => {
    const { data, loading } = useQuery<FetchOwnersAndHistoryResult, FetchOwnersAndHistoryVariables>(
        FETCH_OWNERS_AND_HISTORY,
        {
            variables: {
                repo: repoID,
                revision: revision ?? '',
                currentPath: filePath,
            },
        }
    )

    if (loading) {
        return <LoadingSpinner />
    }

    if (!(data?.node?.__typename === 'Repository' && data.node.commit)) {
        return <div>Error getting details about this file.</div>
    }

    const history = data?.node?.commit?.ancestors?.nodes?.[0]
    // const ownership = data.node.commit?.blob?.ownership

    return (
        <div>{history && <GitCommitNode node={history} extraCompact={true} hideExpandCommitMessageBody={true} />}</div>
    )
}

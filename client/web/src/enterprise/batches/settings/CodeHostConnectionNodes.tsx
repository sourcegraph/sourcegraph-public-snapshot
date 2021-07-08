import { ApolloError } from '@apollo/client'
import React, { useMemo } from 'react'

import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { Scalars } from '@sourcegraph/shared/src/graphql-operations'
import { Connection } from '@sourcegraph/web/src/components/FilteredConnection'

import {
    BatchChangesCodeHostFields,
    GlobalBatchChangesCodeHostsResult,
    GlobalBatchChangesCodeHostsVariables,
    UserBatchChangesCodeHostsResult,
    UserBatchChangesCodeHostsVariables,
} from '../../../graphql-operations'
import { usePaginatedConnection } from '../../../user/settings/accessTokens/usePaginatedConnection'

import { GLOBAL_BATCH_CHANGES_CODE_HOSTS, USER_BATCH_CHANGES_CODE_HOSTS } from './backend'
import { CodeHostConnectionNode } from './CodeHostConnectionNode'

interface CodeHostConnectionNodesProps {
    userID: Scalars['ID']
}

export const CodeHostConnectionNodes: React.FunctionComponent<CodeHostConnectionNodesProps> = ({ userID }) => {
    const response = usePaginatedConnection<
        UserBatchChangesCodeHostsResult,
        UserBatchChangesCodeHostsVariables,
        BatchChangesCodeHostFields
    >({
        query: USER_BATCH_CHANGES_CODE_HOSTS,
        variables: {
            user: userID,
            first: 15,
            after: null,
        },
        getConnection: data => {
            if (data.node === null) {
                throw new Error('User not found')
            }

            if (data.node.__typename !== 'User') {
                throw new Error(`Node is a ${data.node.__typename}, not a User`)
            }

            return data.node.batchChangesCodeHosts
        },
    })

    return <CodeHostConnectionNodesUI userID={userID} {...response} />
}

export const GlobalCodeHostConnectionNodes: React.FunctionComponent = () => {
    const response = usePaginatedConnection<
        GlobalBatchChangesCodeHostsResult,
        GlobalBatchChangesCodeHostsVariables,
        BatchChangesCodeHostFields
    >({
        query: GLOBAL_BATCH_CHANGES_CODE_HOSTS,
        variables: {
            first: 15,
            after: null,
        },
        getConnection: result => result.batchChangesCodeHosts,
    })

    return <CodeHostConnectionNodesUI userID={null} {...response} />
}

const CodeHostConnectionNodesUI: React.FunctionComponent<{
    connection?: Connection<BatchChangesCodeHostFields>
    error?: ApolloError
    loading: boolean
    userID: Scalars['ID'] | null
    fetchMore: () => void
}> = ({ connection, error, loading, userID, fetchMore }) => {
    if (error) {
        throw error
    }

    if (!connection) {
        return null
    }

    if (loading) {
        return <LoadingSpinner className="icon-inline" />
    }

    return (
        <>
            {connection.nodes.map((node, index) => (
                <CodeHostConnectionNode key={index} node={node} userID={userID} />
            ))}
            {connection?.pageInfo?.hasNextPage && (
                <button type="button" className="btn btn-sm btn-link" onClick={fetchMore}>
                    Show more
                </button>
            )}
        </>
    )
}

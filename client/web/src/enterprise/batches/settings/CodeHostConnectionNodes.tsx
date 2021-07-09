import { ApolloError } from '@apollo/client'
import { GraphQLError } from 'graphql'
import React from 'react'

import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { Scalars } from '@sourcegraph/shared/src/graphql-operations'
import { dataOrThrowErrors } from '@sourcegraph/shared/src/graphql/graphql'
import { createAggregateError } from '@sourcegraph/shared/src/util/errors'
import { Connection } from '@sourcegraph/web/src/components/FilteredConnection'
import {
    ConnectionContainer,
    ConnectionError,
    ConnectionList,
    ConnectionLoading,
    ConnectionSummary,
    ShowMoreButton,
    SummaryContainer,
} from '@sourcegraph/web/src/components/FilteredConnection/generic-ui'
import { hasNextPage } from '@sourcegraph/web/src/components/FilteredConnection/utils'

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
        getConnection: result => {
            const data = dataOrThrowErrors(result)
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
        getConnection: result => {
            const data = dataOrThrowErrors(result)
            return data.batchChangesCodeHosts
        },
    })

    return <CodeHostConnectionNodesUI userID={null} {...response} />
}

const CodeHostConnectionNodesUI: React.FunctionComponent<{
    connection?: Connection<BatchChangesCodeHostFields>
    errors: readonly GraphQLError[]
    loading: boolean
    hasNextPage: boolean
    userID: Scalars['ID'] | null
    fetchMore: () => void
}> = ({ connection, errors, loading, hasNextPage, userID, fetchMore }) => {
    if (!connection) {
        return null
    }

    return (
        <ConnectionContainer className="mb-3">
            {errors.length > 0 && <ConnectionError errors={errors} />}
            <ConnectionList className="list-group">
                {connection.nodes.map((node, index) => (
                    <CodeHostConnectionNode key={index} node={node} userID={userID} />
                ))}
            </ConnectionList>
            {loading && <ConnectionLoading />}
            {connection && (
                <SummaryContainer>
                    <ConnectionSummary
                        noSummaryIfAllNodesVisible={true}
                        connection={connection}
                        noun="code host"
                        pluralNoun="code hosts"
                        totalCount={connection.totalCount ?? null}
                        hasNextPage={hasNextPage}
                    />
                    {hasNextPage && <ShowMoreButton onClick={fetchMore} />}
                </SummaryContainer>
            )}
        </ConnectionContainer>
    )
}

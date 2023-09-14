import React from 'react'

import { Container, Link, H3, Text } from '@sourcegraph/wildcard'

import type { UseShowMorePaginationResult } from '../../../components/FilteredConnection/hooks/useShowMorePagination'
import {
    ConnectionContainer,
    ConnectionError,
    ConnectionList,
    ConnectionLoading,
    ConnectionSummary,
    ShowMoreButton,
    SummaryContainer,
} from '../../../components/FilteredConnection/ui'
import type {
    BatchChangesCodeHostFields,
    GlobalBatchChangesCodeHostsResult,
    Scalars,
    UserBatchChangesCodeHostsResult,
} from '../../../graphql-operations'

import { useGlobalBatchChangesCodeHostConnection, useUserBatchChangesCodeHostConnection } from './backend'
import { CodeHostConnectionNode } from './CodeHostConnectionNode'

export interface GlobalCodeHostConnectionsProps {
    headerLine: JSX.Element
}

export const GlobalCodeHostConnections: React.FunctionComponent<
    React.PropsWithChildren<GlobalCodeHostConnectionsProps>
> = props => (
    <CodeHostConnections userID={null} connectionResult={useGlobalBatchChangesCodeHostConnection()} {...props} />
)

export interface UserCodeHostConnectionsProps extends GlobalCodeHostConnectionsProps {
    userID: Scalars['ID']
}

export const UserCodeHostConnections: React.FunctionComponent<
    React.PropsWithChildren<UserCodeHostConnectionsProps>
> = props => <CodeHostConnections connectionResult={useUserBatchChangesCodeHostConnection(props.userID)} {...props} />

interface CodeHostConnectionsProps extends GlobalCodeHostConnectionsProps {
    userID: Scalars['ID'] | null
    connectionResult: UseShowMorePaginationResult<
        GlobalBatchChangesCodeHostsResult | UserBatchChangesCodeHostsResult,
        BatchChangesCodeHostFields
    >
}

const CodeHostConnections: React.FunctionComponent<React.PropsWithChildren<CodeHostConnectionsProps>> = ({
    userID,
    headerLine,
    connectionResult,
}) => {
    const { loading, hasNextPage, fetchMore, connection, error, refetchAll } = connectionResult
    return (
        <Container className="mb-3">
            <H3>Code host tokens</H3>
            {headerLine}
            <ConnectionContainer className="mb-3">
                {error && <ConnectionError errors={[error.message]} />}
                {loading && !connection && <ConnectionLoading />}
                <ConnectionList as="ul" className="list-group" aria-label="code host connections">
                    {connection?.nodes?.map(node => (
                        <CodeHostConnectionNode
                            key={node.externalServiceURL}
                            node={node}
                            refetchAll={refetchAll}
                            userID={userID}
                        />
                    ))}
                </ConnectionList>
                {connection && (
                    <SummaryContainer className="mt-2">
                        <ConnectionSummary
                            noSummaryIfAllNodesVisible={true}
                            first={15}
                            centered={true}
                            connection={connection}
                            noun="code host"
                            pluralNoun="code host connections"
                            hasNextPage={hasNextPage}
                        />
                        {hasNextPage && <ShowMoreButton centered={true} onClick={fetchMore} />}
                    </SummaryContainer>
                )}
            </ConnectionContainer>
            <Text className="mb-0">
                Code host not present? Site admins can add a code host in{' '}
                <Link to="/help/admin/external_service" target="_blank" rel="noopener noreferrer">
                    the manage repositories settings
                </Link>
                .
            </Text>
        </Container>
    )
}

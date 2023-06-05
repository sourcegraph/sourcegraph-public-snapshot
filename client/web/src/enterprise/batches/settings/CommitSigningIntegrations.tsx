import React from 'react'

import { Container, H3, Text } from '@sourcegraph/wildcard'

import { UseShowMorePaginationResult } from '../../../components/FilteredConnection/hooks/useShowMorePagination'
import {
    ConnectionContainer,
    ConnectionError,
    ConnectionList,
    ConnectionLoading,
    ConnectionSummary,
    ShowMoreButton,
    SummaryContainer,
} from '../../../components/FilteredConnection/ui'
import {
    BatchChangesCodeHostFields,
    GlobalBatchChangesCodeHostsResult,
    Scalars,
    UserBatchChangesCodeHostsResult,
} from '../../../graphql-operations'

import { useGlobalBatchChangesCodeHostConnection, useUserBatchChangesCodeHostConnection } from './backend'
import { CommitSigningIntegrationNode } from './CommitSigningIntegrationNode'

export const GlobalCommitSigningIntegrations: React.FunctionComponent<React.PropsWithChildren<{}>> = () => (
    <CommitSigningIntegrations connectionResult={useGlobalBatchChangesCodeHostConnection()} readOnly={false} />
)

interface UserCommitSigningIntegrationsProps {
    userID: Scalars['ID']
}

export const UserCommitSigningIntegrations: React.FunctionComponent<
    React.PropsWithChildren<UserCommitSigningIntegrationsProps>
> = ({ userID }) => (
    <CommitSigningIntegrations connectionResult={useUserBatchChangesCodeHostConnection(userID)} readOnly={true} />
)

interface CommitSigningIntegrationsProps {
    readOnly: boolean
    connectionResult: UseShowMorePaginationResult<
        GlobalBatchChangesCodeHostsResult | UserBatchChangesCodeHostsResult,
        BatchChangesCodeHostFields
    >
}

export const CommitSigningIntegrations: React.FunctionComponent<
    React.PropsWithChildren<CommitSigningIntegrationsProps>
> = ({ connectionResult, readOnly }) => {
    const { loading, hasNextPage, fetchMore, connection, error } = connectionResult
    return (
        <Container>
            <H3>Commit signing integrations</H3>
            <Text>
                Connect GitHub Apps to enable Batch Changes to sign commits for your changesets. Contact your site admin
                to manage connections.
            </Text>
            <ConnectionContainer className="mb-3">
                {error && <ConnectionError errors={[error.message]} />}
                {loading && !connection && <ConnectionLoading />}
                <ConnectionList as="ul" className="list-group" aria-label="commit signing integrations">
                    {connection?.nodes?.map(node =>
                        node.supportsCommitSigning ? (
                            <CommitSigningIntegrationNode
                                key={node.externalServiceURL}
                                node={node}
                                readOnly={readOnly}
                            />
                        ) : null
                    )}
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
        </Container>
    )
}

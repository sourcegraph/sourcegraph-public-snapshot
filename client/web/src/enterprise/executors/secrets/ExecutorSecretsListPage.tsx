import React from 'react'

import { Container, H3, PageHeader } from '@sourcegraph/wildcard'

import { UseConnectionResult } from '../../../components/FilteredConnection/hooks/useConnection'
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
    ExecutorSecretFields,
    ExecutorSecretScope,
    Scalars,
} from '../../../graphql-operations'

import { useGlobalExecutorSecretsConnection, useUserBatchChangesCodeHostConnection } from './backend'
import { ExecutorSecretNode } from './ExecutorSecretNode'

export interface GlobalExecutorSecretsListPageProps {
    headerLine: JSX.Element
}

export const GlobalExecutorSecretsListPage: React.FunctionComponent<
    React.PropsWithChildren<GlobalExecutorSecretsListPageProps>
> = props => (
    <ExecutorSecretsListPage
        userID={null}
        connectionResult={useGlobalExecutorSecretsConnection(ExecutorSecretScope.BATCHES)}
        {...props}
    />
)

export interface UserExecutorSecretsListPageProps extends GlobalExecutorSecretsListPageProps {
    userID: Scalars['ID']
}

export const UserExecutorSecretsListPage: React.FunctionComponent<
    React.PropsWithChildren<UserExecutorSecretsListPageProps>
> = props => (
    <ExecutorSecretsListPage connectionResult={useUserBatchChangesCodeHostConnection(props.userID)} {...props} />
)

interface ExecutorSecretsListPageProps extends GlobalExecutorSecretsListPageProps {
    userID: Scalars['ID'] | null
    connectionResult: UseConnectionResult<ExecutorSecretFields>
}

const ExecutorSecretsListPage: React.FunctionComponent<React.PropsWithChildren<ExecutorSecretsListPageProps>> = ({
    userID,
    headerLine,
    connectionResult,
}) => {
    const { loading, hasNextPage, fetchMore, connection, error, refetchAll } = connectionResult
    return (
        <>
            <PageHeader
                path={[{ text: 'Executor secrets' }]}
                headingElement="h2"
                description={headerLine}
                className="mb-3"
            />

            <Container>
                <ConnectionContainer className="mb-3">
                    {error && <ConnectionError errors={[error.message]} />}
                    {loading && !connection && <ConnectionLoading />}
                    <ConnectionList as="ul" className="list-group" aria-label="Code hosts">
                        {connection?.nodes?.map(node => (
                            <ExecutorSecretNode
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
                                pluralNoun="code hosts"
                                hasNextPage={hasNextPage}
                            />
                            {hasNextPage && <ShowMoreButton centered={true} onClick={fetchMore} />}
                        </SummaryContainer>
                    )}
                </ConnectionContainer>
            </Container>
        </>
    )
}

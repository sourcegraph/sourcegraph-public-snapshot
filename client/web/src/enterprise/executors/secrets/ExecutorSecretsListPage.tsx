import React, { useCallback, useState } from 'react'

import { Button, Container, PageHeader } from '@sourcegraph/wildcard'

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
import { ExecutorSecretFields, ExecutorSecretScope, Scalars } from '../../../graphql-operations'

import { AddSecretModal } from './AddSecretModal'
import { useExecutorSecretsConnection, useGlobalExecutorSecretsConnection } from './backend'
import { ExecutorSecretNode } from './ExecutorSecretNode'

export interface GlobalExecutorSecretsListPageProps {
    headerLine: JSX.Element
}

export const GlobalExecutorSecretsListPage: React.FunctionComponent<
    React.PropsWithChildren<GlobalExecutorSecretsListPageProps>
> = props => (
    <ExecutorSecretsListPage
        namespaceID={null}
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
    <ExecutorSecretsListPage
        namespaceID={props.userID}
        connectionResult={useExecutorSecretsConnection(props.userID, ExecutorSecretScope.BATCHES)}
        {...props}
    />
)

interface ExecutorSecretsListPageProps extends GlobalExecutorSecretsListPageProps {
    namespaceID: Scalars['ID'] | null
    connectionResult: UseConnectionResult<ExecutorSecretFields>
}

const ExecutorSecretsListPage: React.FunctionComponent<React.PropsWithChildren<ExecutorSecretsListPageProps>> = ({
    namespaceID,
    headerLine,
    connectionResult,
}) => {
    const { loading, hasNextPage, fetchMore, connection, error, refetchAll } = connectionResult

    const [showAddModal, setShowAddModal] = useState<boolean>(false)
    const onClickAdd = useCallback<React.MouseEventHandler>(event => {
        event.preventDefault()
        setShowAddModal(true)
    }, [])

    const closeModal = useCallback(() => {
        setShowAddModal(false)
    }, [])
    const afterAction = useCallback(() => {
        setShowAddModal(false)
        refetchAll()
    }, [refetchAll])

    return (
        <>
            <PageHeader
                path={[{ text: 'Executor secrets' }]}
                headingElement="h2"
                description={headerLine}
                actions={
                    <>
                        <Button onClick={onClickAdd} aria-label="Add new secret value" variant="primary">
                            Add secret
                        </Button>
                    </>
                }
                className="mb-3"
            />

            {showAddModal && (
                <AddSecretModal
                    onCancel={closeModal}
                    afterCreate={afterAction}
                    scope={ExecutorSecretScope.BATCHES}
                    namespaceID={namespaceID}
                />
            )}

            <Container>
                <ConnectionContainer className="mb-3">
                    {error && <ConnectionError errors={[error.message]} />}
                    {loading && !connection && <ConnectionLoading />}
                    <ConnectionList as="ul" className="list-group" aria-label="Code hosts">
                        {connection?.nodes?.map(node => (
                            <ExecutorSecretNode key={node.id} node={node} refetchAll={refetchAll} />
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

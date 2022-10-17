import React, { useCallback, useMemo, useState } from 'react'

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
import { ExecutorSecretScopeSelector } from './ExecutorSecretScopeSelector'

export interface GlobalExecutorSecretsListPageProps {
    headerLine: JSX.Element
}

export const GlobalExecutorSecretsListPage: React.FunctionComponent<
    React.PropsWithChildren<GlobalExecutorSecretsListPageProps>
> = props => {
    const connectionLoader = useCallback((scope: ExecutorSecretScope) => useGlobalExecutorSecretsConnection(scope), [])
    return <ExecutorSecretsListPage namespaceID={null} connectionLoader={connectionLoader} {...props} />
}

export interface UserExecutorSecretsListPageProps extends GlobalExecutorSecretsListPageProps {
    userID: Scalars['ID']
}

export const UserExecutorSecretsListPage: React.FunctionComponent<
    React.PropsWithChildren<UserExecutorSecretsListPageProps>
> = props => {
    const connectionLoader = useCallback(
        (scope: ExecutorSecretScope) => useExecutorSecretsConnection(props.userID, scope),
        []
    )
    return <ExecutorSecretsListPage namespaceID={props.userID} connectionLoader={connectionLoader} {...props} />
}

interface ExecutorSecretsListPageProps extends GlobalExecutorSecretsListPageProps {
    namespaceID: Scalars['ID'] | null
    connectionLoader: (scope: ExecutorSecretScope) => UseConnectionResult<ExecutorSecretFields>
}

const ExecutorSecretsListPage: React.FunctionComponent<React.PropsWithChildren<ExecutorSecretsListPageProps>> = ({
    namespaceID,
    headerLine,
    connectionLoader,
}) => {
    const [selectedScope, setSelectedScope] = useState<ExecutorSecretScope>(ExecutorSecretScope.BATCHES)
    const { loading, hasNextPage, fetchMore, connection, error, refetchAll } = connectionLoader(selectedScope)

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
                    scope={selectedScope}
                    namespaceID={namespaceID}
                />
            )}

            <div className="d-flex mb-3">
                {Object.values(ExecutorSecretScope).map(scope => (
                    <ExecutorSecretScopeSelector
                        key={scope}
                        scope={scope}
                        label={executorSecretScopeContext(scope).label}
                        onSelect={() => setSelectedScope(scope)}
                        selected={scope === selectedScope}
                        description={executorSecretScopeContext(scope).description}
                    />
                ))}
            </div>

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

function executorSecretScopeContext(scope: ExecutorSecretScope): { label: string; description: string } {
    switch (scope) {
        case ExecutorSecretScope.BATCHES:
            return { label: 'Batch changes', description: 'Batch change execution secrets' }
    }
}

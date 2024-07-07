import React, { useCallback, useEffect, useState, type FC } from 'react'

import { dataOrThrowErrors } from '@sourcegraph/http-client'
import type { TelemetryV2Props } from '@sourcegraph/shared/src/telemetry'
import { Button, Container, Link, PageHeader } from '@sourcegraph/wildcard'

import { useUrlSearchParamsForConnectionState } from '../../../components/FilteredConnection/hooks/connectionState'
import {
    useShowMorePagination,
    type UseShowMorePaginationResult,
} from '../../../components/FilteredConnection/hooks/useShowMorePagination'
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
    ExecutorSecretScope,
    type ExecutorSecretFields,
    type GlobalExecutorSecretsResult,
    type GlobalExecutorSecretsVariables,
    type OrgExecutorSecretsResult,
    type Scalars,
    type UserExecutorSecretsResult,
} from '../../../graphql-operations'

import { AddSecretModal } from './AddSecretModal'
import {
    GLOBAL_EXECUTOR_SECRETS,
    orgExecutorSecretsConnectionFactory,
    userExecutorSecretsConnectionFactory,
} from './backend'
import { ExecutorSecretNode } from './ExecutorSecretNode'
import { ExecutorSecretScopeSelector } from './ExecutorSecretScopeSelector'

export interface GlobalExecutorSecretsListPageProps extends TelemetryV2Props {}

export const GlobalExecutorSecretsListPage: FC<GlobalExecutorSecretsListPageProps> = props => {
    useEffect(
        () => props.telemetryRecorder.recordEvent('admin.executors.secretsList', 'view'),
        [props.telemetryRecorder]
    )

    const connectionState = useUrlSearchParamsForConnectionState()
    const connectionLoader = useCallback(
        (scope: ExecutorSecretScope) =>
            // Scope has to be injected dynamically.
            // eslint-disable-next-line react-hooks/rules-of-hooks
            useShowMorePagination<GlobalExecutorSecretsResult, GlobalExecutorSecretsVariables, ExecutorSecretFields>({
                query: GLOBAL_EXECUTOR_SECRETS,
                variables: {
                    scope,
                },
                options: {
                    fetchPolicy: 'network-only',
                },
                getConnection: result => {
                    const { executorSecrets } = dataOrThrowErrors(result)
                    return executorSecrets
                },
                state: connectionState,
            }),
        [connectionState]
    )
    return (
        <ExecutorSecretsListPage
            areaType="admin"
            namespaceID={null}
            headerLine={<>Configure executor secrets that will be available to everyone on the Sourcegraph instance.</>}
            connectionLoader={connectionLoader}
            {...props}
        />
    )
}

export interface UserExecutorSecretsListPageProps extends GlobalExecutorSecretsListPageProps {
    userID: Scalars['ID']
}

export const UserExecutorSecretsListPage: FC<UserExecutorSecretsListPageProps> = props => {
    useEffect(
        () => props.telemetryRecorder.recordEvent('settings.executors.secretsList', 'view'),
        [props.telemetryRecorder]
    )
    const connectionLoader = useCallback(
        (scope: ExecutorSecretScope) => userExecutorSecretsConnectionFactory(props.userID, scope),
        [props.userID]
    )
    return (
        <ExecutorSecretsListPage
            areaType="settings"
            namespaceID={props.userID}
            headerLine={
                <>
                    Configure executor secrets that will only be available to your executions.
                    <br />
                    Global secrets are available to executions in this namespace. Secrets in this namespace with the
                    same name as a global secret will overwrite the global secret. Site admins can configure global
                    secrets <Link to="/admin/executors/secrets">in site admin settings</Link>.
                </>
            }
            connectionLoader={connectionLoader}
            {...props}
        />
    )
}

export interface OrgExecutorSecretsListPageProps extends GlobalExecutorSecretsListPageProps {
    orgID: Scalars['ID']
}

export const OrgExecutorSecretsListPage: FC<OrgExecutorSecretsListPageProps> = props => {
    useEffect(() => props.telemetryRecorder.recordEvent('org.executors.secretsList', 'view'), [props.telemetryRecorder])
    const connectionLoader = useCallback(
        (scope: ExecutorSecretScope) => orgExecutorSecretsConnectionFactory(props.orgID, scope),
        [props.orgID]
    )
    return (
        <ExecutorSecretsListPage
            areaType="org"
            namespaceID={props.orgID}
            headerLine={
                <>
                    Configure executor secrets that will only be available to executions in this organization.
                    <br />
                    Global secrets are available to executions in this namespace. Secrets in this namespace with the
                    same name as a global secret will overwrite the global secret. Site admins can configure global
                    secrets <Link to="/admin/executors/secrets">in site admin settings</Link>.
                </>
            }
            connectionLoader={connectionLoader}
            {...props}
        />
    )
}

type executorSecretsAreaType = 'admin' | 'org' | 'settings'

export interface ExecutorSecretsListPageProps extends GlobalExecutorSecretsListPageProps {
    // Used as a prefix for telemetry event naming
    areaType: executorSecretsAreaType
    namespaceID: Scalars['ID'] | null
    headerLine: JSX.Element
    connectionLoader: (
        scope: ExecutorSecretScope
    ) => UseShowMorePaginationResult<
        OrgExecutorSecretsResult | UserExecutorSecretsResult | GlobalExecutorSecretsResult,
        ExecutorSecretFields
    >
}

const ExecutorSecretsListPage: FC<ExecutorSecretsListPageProps> = ({
    areaType,
    namespaceID,
    headerLine,
    connectionLoader,
    telemetryRecorder,
}) => {
    const [selectedScope, setSelectedScope] = useState<ExecutorSecretScope>(ExecutorSecretScope.BATCHES)
    const { loading, hasNextPage, fetchMore, connection, error, refetchAll } = connectionLoader(selectedScope)

    const [showAddModal, setShowAddModal] = useState<boolean>(false)
    const onClickAdd = useCallback<React.MouseEventHandler>(
        event => {
            event.preventDefault()
            telemetryRecorder.recordEvent(`${areaType}.executors.addSecret`, 'click')
            setShowAddModal(true)
        },
        [areaType, telemetryRecorder]
    )

    const closeModal = useCallback(() => {
        telemetryRecorder.recordEvent(`${areaType}.executors.addSecret`, 'cancel')
        setShowAddModal(false)
    }, [areaType, telemetryRecorder])
    const afterAction = useCallback(() => {
        telemetryRecorder.recordEvent(`${areaType}.executors.addSecret`, 'submit')
        setShowAddModal(false)
        refetchAll()
    }, [areaType, refetchAll, telemetryRecorder])

    return (
        <>
            <PageHeader
                path={[{ text: 'Executor secrets' }]}
                headingElement="h2"
                description={headerLine}
                actions={
                    <>
                        <Button onClick={onClickAdd} variant="primary">
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
                {(namespaceID === null ? Object.values(ExecutorSecretScope) : [ExecutorSecretScope.BATCHES]).map(
                    scope => (
                        <ExecutorSecretScopeSelector
                            key={scope}
                            scope={scope}
                            label={executorSecretScopeContext(scope).label}
                            onSelect={() => setSelectedScope(scope)}
                            selected={scope === selectedScope}
                            description={executorSecretScopeContext(scope).description}
                        />
                    )
                )}
            </div>

            <Container>
                <ConnectionContainer className="mb-3">
                    {error && <ConnectionError errors={[error.message]} />}
                    {loading && !connection && <ConnectionLoading />}
                    <ConnectionList as="ul" className="list-group" aria-label="Executor secrets">
                        {connection?.nodes?.map(node => (
                            <ExecutorSecretNode
                                key={node.id}
                                node={node}
                                namespaceID={namespaceID}
                                refetchAll={refetchAll}
                            />
                        ))}
                    </ConnectionList>
                    {connection && (
                        <SummaryContainer className="mt-2">
                            <ConnectionSummary
                                noSummaryIfAllNodesVisible={true}
                                centered={true}
                                connection={connection}
                                noun="executor secret"
                                pluralNoun="executor secrets"
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

// When you add a new secret scope, this function will generate a compile error,
// so that this is not forgotten. Add a new label/description tuple for the new
// scope added here and TS will be happy.
function executorSecretScopeContext(scope: ExecutorSecretScope): { label: string; description: string } {
    switch (scope) {
        case ExecutorSecretScope.BATCHES: {
            return { label: 'Batch changes', description: 'Batch change execution secrets' }
        }
        case ExecutorSecretScope.CODEINTEL: {
            return { label: 'Code graph', description: 'Code graph execution secrets' }
        }
    }
}

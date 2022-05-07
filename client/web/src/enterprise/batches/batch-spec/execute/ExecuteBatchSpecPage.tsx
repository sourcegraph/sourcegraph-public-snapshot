import React, { useMemo } from 'react'

import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import MapSearchIcon from 'mdi-react/MapSearchIcon'
import { Redirect, Route, RouteComponentProps, Switch } from 'react-router'

import { useQuery } from '@sourcegraph/http-client'
import { AuthenticatedUser } from '@sourcegraph/shared/src/auth'
import { LinkOrSpan } from '@sourcegraph/shared/src/components/LinkOrSpan'
import { Settings, SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { Icon, LoadingSpinner } from '@sourcegraph/wildcard'

import { withAuthenticatedUser } from '../../../../auth/withAuthenticatedUser'
import { HeroPage } from '../../../../components/HeroPage'
import { Timestamp } from '../../../../components/time/Timestamp'
import {
    BatchSpecExecutionByIDResult,
    BatchSpecExecutionByIDVariables,
    BatchSpecExecutionFields,
    GetBatchChangeToEditResult,
    GetBatchChangeToEditVariables,
    Scalars,
} from '../../../../graphql-operations'
// TODO: Move some of these to batch-spec/edit
// TODO: Make sure fields on GraphQL queries are all still needed
import { GET_BATCH_CHANGE_TO_EDIT } from '../../create/backend'
import { ConfigurationForm } from '../../create/ConfigurationForm'
import { NewBatchChangePreviewPage } from '../../preview/BatchChangePreviewPage'
import { BatchSpecContextProvider, useBatchSpecContext } from '../BatchSpecContext'
import { BatchChangeHeader } from '../header/BatchChangeHeader'
import { TabBar, TabsConfig } from '../TabBar'

import { FETCH_BATCH_SPEC_EXECUTION } from './backend'
import { ExecutionWorkspaces } from './ExecutionWorkspaces'
import { ReadOnlyBatchSpecForm } from './ReadOnlyBatchSpecForm'

import layoutStyles from '../Layout.module.scss'

export interface AuthenticatedExecuteBatchSpecPageProps
    extends SettingsCascadeProps<Settings>,
        ThemeProps,
        TelemetryProps,
        RouteComponentProps<{}> {
    batchChange: { name: string; namespace: Scalars['ID'] }
    batchSpecID: Scalars['ID']
    authenticatedUser: AuthenticatedUser
}

export const AuthenticatedExecuteBatchSpecPage: React.FunctionComponent<AuthenticatedExecuteBatchSpecPageProps> = ({
    batchChange,
    batchSpecID,
    ...props
}) => {
    const { data: batchChangeData, error: batchChangeError, loading: batchChangeLoading } = useQuery<
        GetBatchChangeToEditResult,
        GetBatchChangeToEditVariables
    >(GET_BATCH_CHANGE_TO_EDIT, {
        variables: batchChange,
        // Cache this data but always re-request it in the background when we revisit
        // this page to pick up newer changes.
        fetchPolicy: 'cache-and-network',
    })

    const { data, error, loading } = useQuery<BatchSpecExecutionByIDResult, BatchSpecExecutionByIDVariables>(
        FETCH_BATCH_SPEC_EXECUTION,
        {
            variables: { id: batchSpecID },
            fetchPolicy: 'cache-and-network',
            pollInterval: 2500,
        }
    )

    if ((loading || batchChangeLoading) && (!data || !batchChangeData)) {
        return (
            <div className="w-100 text-center">
                <Icon className="m-2" as={LoadingSpinner} />
            </div>
        )
    }

    if (!data?.node || data.node.__typename !== 'BatchSpec' || !batchChangeData?.batchChange) {
        if (error || batchChangeError) {
            return <HeroPage icon={AlertCircleIcon} title={String(error || batchChangeError)} />
        }
        return <HeroPage icon={AlertCircleIcon} title="Batch change not found" />
    }

    return (
        <BatchSpecContextProvider batchChange={batchChangeData.batchChange} batchSpec={data.node}>
            <ExecuteBatchSpecPageContent {...props} />
        </BatchSpecContextProvider>
    )
}

interface ExecuteBatchSpecPageContentProps
    extends SettingsCascadeProps<Settings>,
        ThemeProps,
        TelemetryProps,
        RouteComponentProps<{}> {
    authenticatedUser: AuthenticatedUser
}

const ExecuteBatchSpecPageContent: React.FunctionComponent<
    React.PropsWithChildren<ExecuteBatchSpecPageContentProps>
> = ({ isLightTheme, match, settingsCascade, telemetryService, authenticatedUser }) => {
    const { batchChange, batchSpec } = useBatchSpecContext<BatchSpecExecutionFields>()
    const { executionURL } = batchSpec

    const tabsConfig = useMemo<TabsConfig[]>(
        () => [
            { key: 'configuration', isEnabled: true, handler: { type: 'link' } },
            { key: 'spec', isEnabled: true, handler: { type: 'link' } },
            { key: 'execution', isEnabled: true, handler: { type: 'link' } },
            { key: 'preview', isEnabled: batchSpec.applyURL !== null, handler: { type: 'link' } },
        ],
        [batchSpec.applyURL]
    )

    return (
        <div className={layoutStyles.pageContainer}>
            <div className={layoutStyles.headerContainer}>
                <BatchChangeHeader
                    namespace={{
                        to: `${batchChange.namespace.url}/batch-changes`,
                        text: batchChange.namespace.namespaceName,
                    }}
                    title={{ to: batchChange.url, text: batchChange.name }}
                    description={
                        <>
                            Created <Timestamp date={batchSpec.createdAt} /> by{' '}
                            <LinkOrSpan to={batchSpec.creator?.url}>
                                {batchSpec.creator?.displayName || batchSpec.creator?.username || 'a deleted user'}
                            </LinkOrSpan>
                        </>
                    }
                />
            </div>

            <Switch>
                <Route render={() => <Redirect to={`${match.url}/execution`} />} path={match.url} exact={true} />
                <Route
                    path={`${match.url}/configuration`}
                    render={() => (
                        <>
                            <TabBar activeTabKey="configuration" tabsConfig={tabsConfig} matchURL={executionURL} />
                            <ConfigurationForm
                                isReadOnly={true}
                                batchChange={batchChange}
                                settingsCascade={settingsCascade}
                            />
                        </>
                    )}
                    exact={true}
                />
                <Route
                    path={`${match.url}/spec`}
                    render={() => (
                        <>
                            <TabBar activeTabKey="spec" tabsConfig={tabsConfig} matchURL={executionURL} />
                            <ReadOnlyBatchSpecForm
                                batchChange={batchChange}
                                originalInput={batchSpec.originalInput}
                                executionState={batchSpec.state}
                                isLightTheme={isLightTheme}
                            />
                        </>
                    )}
                    exact={true}
                />
                <Route
                    path={`${match.url}/execution/workspaces/:workspaceID`}
                    render={({ match }: RouteComponentProps<{ workspaceID: string }>) => (
                        <>
                            <TabBar activeTabKey="execution" tabsConfig={tabsConfig} matchURL={executionURL} />
                            <ExecutionWorkspaces
                                selectedWorkspaceID={match.params.workspaceID}
                                isLightTheme={isLightTheme}
                            />
                        </>
                    )}
                />
                <Route
                    path={`${match.url}/execution`}
                    render={() => (
                        <>
                            <TabBar activeTabKey="execution" tabsConfig={tabsConfig} matchURL={executionURL} />
                            <ExecutionWorkspaces isLightTheme={isLightTheme} />
                        </>
                    )}
                />
                {batchSpec.applyURL ? (
                    <Route
                        path={`${match.url}/preview`}
                        render={() => (
                            <>
                                <TabBar
                                    activeTabKey="preview"
                                    tabsConfig={tabsConfig}
                                    matchURL={executionURL}
                                    className="mb-4"
                                />
                                <NewBatchChangePreviewPage
                                    authenticatedUser={authenticatedUser}
                                    telemetryService={telemetryService}
                                    isLightTheme={isLightTheme}
                                    batchSpecID={batchSpec.id}
                                />
                            </>
                        )}
                        exact={true}
                    />
                ) : null}
                <Route component={() => <HeroPage icon={MapSearchIcon} title="404: Not Found" />} key="hardcoded-key" />
            </Switch>
        </div>
    )
}

export const ExecuteBatchSpecPage = withAuthenticatedUser(AuthenticatedExecuteBatchSpecPage)

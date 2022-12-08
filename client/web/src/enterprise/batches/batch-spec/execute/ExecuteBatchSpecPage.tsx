import React, { useMemo } from 'react'

import { mdiProgressClock } from '@mdi/js'
import { VisuallyHidden } from '@reach/visually-hidden'
import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import MapSearchIcon from 'mdi-react/MapSearchIcon'
import { Redirect, Route, RouteComponentProps, Switch } from 'react-router'

import { ErrorMessage } from '@sourcegraph/branded/src/components/alerts'
import { useQuery } from '@sourcegraph/http-client'
import { AuthenticatedUser } from '@sourcegraph/shared/src/auth'
import { LinkOrSpan } from '@sourcegraph/shared/src/components/LinkOrSpan'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { Badge, Icon, LoadingSpinner } from '@sourcegraph/wildcard'

import { withAuthenticatedUser } from '../../../../auth/withAuthenticatedUser'
import { HeroPage } from '../../../../components/HeroPage'
import { Duration } from '../../../../components/time/Duration'
import { Timestamp } from '../../../../components/time/Timestamp'
import {
    BatchSpecExecutionByIDResult,
    BatchSpecExecutionByIDVariables,
    BatchSpecExecutionFields,
    BatchSpecSource,
    GetBatchChangeToEditResult,
    GetBatchChangeToEditVariables,
    Scalars,
} from '../../../../graphql-operations'
import { GET_BATCH_CHANGE_TO_EDIT } from '../../create/backend'
import { ConfigurationForm } from '../../create/ConfigurationForm'
import { NewBatchChangePreviewPage } from '../../preview/BatchChangePreviewPage'
import { BatchSpecContextProvider, BatchSpecContextState, useBatchSpecContext } from '../BatchSpecContext'
import { ActionButtons } from '../header/ActionButtons'
import { BatchChangeHeader } from '../header/BatchChangeHeader'
import { TabBar, TabsConfig } from '../TabBar'

import { ActionsMenu } from './ActionsMenu'
import { FETCH_BATCH_SPEC_EXECUTION, queryWorkspacesList as _queryWorkspacesList } from './backend'
import { BatchSpecStateBadge } from './BatchSpecStateBadge'
import { ExecutionStat, ExecutionStatsBar } from './ExecutionStatsBar'
import { ReadOnlyBatchSpecForm } from './ReadOnlyBatchSpecForm'
import { ExecutionWorkspaces } from './workspaces/ExecutionWorkspaces'

import layoutStyles from '../Layout.module.scss'
import styles from './ExecuteBatchSpecPage.module.scss'

export interface AuthenticatedExecuteBatchSpecPageProps extends ThemeProps, TelemetryProps, RouteComponentProps<{}> {
    batchChange: { name: string; namespace: Scalars['ID'] }
    batchSpecID: Scalars['ID']
    authenticatedUser: AuthenticatedUser
    /** FOR TESTING ONLY */
    testContextState?: Partial<BatchSpecContextState<BatchSpecExecutionFields>>
    queryWorkspacesList?: typeof _queryWorkspacesList
}

export const AuthenticatedExecuteBatchSpecPage: React.FunctionComponent<
    React.PropsWithChildren<AuthenticatedExecuteBatchSpecPageProps>
> = ({ batchChange, batchSpecID, testContextState, ...props }) => {
    const {
        data: batchChangeData,
        error: batchChangeError,
        loading: batchChangeLoading,
    } = useQuery<GetBatchChangeToEditResult, GetBatchChangeToEditVariables>(GET_BATCH_CHANGE_TO_EDIT, {
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
                <Icon aria-label="Loading" className="m-2" as={LoadingSpinner} />
            </div>
        )
    }

    if (data?.node === null) {
        return <HeroPage icon={MapSearchIcon} title="404: Not Found" />
    }

    if (!data?.node || data.node.__typename !== 'BatchSpec' || !batchChangeData?.batchChange) {
        if (error || batchChangeError) {
            return <HeroPage icon={AlertCircleIcon} title={String(error || batchChangeError)} />
        }
        return <HeroPage icon={AlertCircleIcon} title="Batch change not found" />
    }

    return (
        <BatchSpecContextProvider
            batchChange={batchChangeData.batchChange}
            batchSpec={data.node}
            testState={testContextState}
        >
            <ExecuteBatchSpecPageContent {...props} />
        </BatchSpecContextProvider>
    )
}

interface ExecuteBatchSpecPageContentProps extends ThemeProps, TelemetryProps, RouteComponentProps<{}> {
    authenticatedUser: AuthenticatedUser
    queryWorkspacesList?: typeof _queryWorkspacesList
}

const ExecuteBatchSpecPageContent: React.FunctionComponent<
    React.PropsWithChildren<ExecuteBatchSpecPageContentProps>
> = props => {
    const { batchChange, batchSpec, errors } = useBatchSpecContext<BatchSpecExecutionFields>()

    return (
        <MemoizedExecuteBatchSpecContent {...props} batchChange={batchChange} batchSpec={batchSpec} errors={errors} />
    )
}

type MemoizedExecuteBatchSpecContentProps = ExecuteBatchSpecPageContentProps &
    Pick<BatchSpecContextState, 'batchChange' | 'batchSpec' | 'errors'>

const MemoizedExecuteBatchSpecContent: React.FunctionComponent<
    React.PropsWithChildren<MemoizedExecuteBatchSpecContentProps>
> = React.memo(function MemoizedExecuteBatchSpecContent({
    isLightTheme,
    match,
    telemetryService,
    authenticatedUser,
    batchChange,
    batchSpec,
    errors,
    queryWorkspacesList,
}) {
    const { executionURL, workspaceResolution, applyURL } = batchSpec

    const tabsConfig = useMemo<TabsConfig[]>(
        () => [
            { key: 'configuration', isEnabled: true, handler: { type: 'link' } },
            { key: 'spec', isEnabled: true, handler: { type: 'link' } },
            { key: 'execution', isEnabled: true, handler: { type: 'link' } },
            { key: 'preview', isEnabled: applyURL !== null, handler: { type: 'link' } },
        ],
        [applyURL]
    )

    return (
        <div className={layoutStyles.pageContainer}>
            <div className={layoutStyles.headerContainer}>
                <BatchChangeHeader
                    className={styles.header}
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
                <div className={styles.statsBar}>
                    <div className={styles.stateBadge}>
                        {batchSpec.source === BatchSpecSource.REMOTE ? (
                            <BatchSpecStateBadge state={batchSpec.state} />
                        ) : (
                            <>
                                <VisuallyHidden>This batch spec was executed with src-cli.</VisuallyHidden>
                                <Badge
                                    variant="secondary"
                                    tooltip="This batch spec was executed with src-cli."
                                    aria-hidden={true}
                                >
                                    LOCAL
                                </Badge>
                            </>
                        )}
                    </div>
                    {batchSpec.startedAt && (
                        <ExecutionStat>
                            <Icon aria-hidden={true} className={styles.durationIcon} svgPath={mdiProgressClock} />
                            <Duration
                                start={batchSpec.startedAt}
                                end={batchSpec.finishedAt ?? undefined}
                                labelPrefix={`The batch spec ${
                                    batchSpec.finishedAt ? 'finished executing in' : 'has been executing for'
                                }`}
                            />
                        </ExecutionStat>
                    )}
                    {workspaceResolution && <ExecutionStatsBar {...workspaceResolution.workspaces.stats} />}
                </div>

                <ActionButtons className="flex-shrink-0">
                    <ActionsMenu />
                </ActionButtons>
            </div>

            {errors.actions && <ErrorMessage error={errors.actions} key={String(errors.actions)} />}

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
                                authenticatedUser={authenticatedUser}
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
                            <ReadOnlyBatchSpecForm isLightTheme={isLightTheme} />
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
                                queryWorkspacesList={queryWorkspacesList}
                            />
                        </>
                    )}
                />
                <Route
                    path={`${match.url}/execution`}
                    render={() => (
                        <>
                            <TabBar activeTabKey="execution" tabsConfig={tabsConfig} matchURL={executionURL} />
                            <ExecutionWorkspaces
                                isLightTheme={isLightTheme}
                                queryWorkspacesList={queryWorkspacesList}
                            />
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
                ) : (
                    // If the batch spec is not ready to be previewed, redirect to the spec instead.
                    <Redirect to={`${match.url}/spec`} />
                )}
                <Route component={() => <HeroPage icon={MapSearchIcon} title="404: Not Found" />} key="hardcoded-key" />
            </Switch>
        </div>
    )
})

export const ExecuteBatchSpecPage = withAuthenticatedUser(AuthenticatedExecuteBatchSpecPage)

import React, { useContext } from 'react'

import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import MapSearchIcon from 'mdi-react/MapSearchIcon'
import { Redirect, Route, RouteComponentProps, Switch } from 'react-router'

import { useQuery } from '@sourcegraph/http-client'
import { Settings, SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { Alert, Button, Icon, LoadingSpinner } from '@sourcegraph/wildcard'

import { HeroPage } from '../../../../components/HeroPage'
import {
    BatchSpecExecutionByIDResult,
    BatchSpecExecutionByIDVariables,
    BatchSpecExecutionFields,
    GetBatchChangeToEditResult,
    GetBatchChangeToEditVariables,
    Scalars,
} from '../../../../graphql-operations'
// TODO: Move some of these to batch-spec/edit
import { BatchSpec } from '../../BatchSpec'
import { GET_BATCH_CHANGE_TO_EDIT } from '../../create/backend'
import { ConfigurationForm } from '../../create/ConfigurationForm'
import { FETCH_BATCH_SPEC_EXECUTION } from '../../execution/backend'
import { BatchSpecContext, BatchSpecContextProvider } from '../BatchSpecContext'
import { LibraryPane } from '../edit/library/LibraryPane'
import { WorkspacesPreviewPanel } from '../edit/workspaces-preview/WorkspacesPreviewPanel'
import { BatchChangeHeader } from '../header/BatchChangeHeader'
import { TabBar, TabsConfig } from '../TabBar'

import editorStyles from '../edit/EditBatchSpecPage.module.scss'
import layoutStyles from '../Layout.module.scss'

export interface ExecuteBatchSpecPageProps extends SettingsCascadeProps<Settings>, ThemeProps, RouteComponentProps<{}> {
    batchChange: { name: string; namespace: Scalars['ID'] }
    batchSpecID: Scalars['ID']
}

export const ExecuteBatchSpecPage: React.FunctionComponent<ExecuteBatchSpecPageProps> = ({
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
            nextFetchPolicy: 'network-only',
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
        <BatchSpecContextProvider batchChange={batchChangeData.batchChange}>
            <ExecuteBatchSpecPageContent {...props} batchSpec={data.node} />
        </BatchSpecContextProvider>
    )
}

interface ExecuteBatchSpecPageContentProps extends SettingsCascadeProps<Settings>, ThemeProps, RouteComponentProps<{}> {
    batchSpec: BatchSpecExecutionFields
}

const TABS_CONFIG: TabsConfig[] = [
    { key: 'configuration', isEnabled: true, handler: { type: 'link' } },
    { key: 'spec', isEnabled: true, handler: { type: 'link' } },
    { key: 'execution', isEnabled: true, handler: { type: 'link' } },
]

const ExecuteBatchSpecPageContent: React.FunctionComponent<
    React.PropsWithChildren<ExecuteBatchSpecPageContentProps>
> = ({ isLightTheme, batchSpec, match, settingsCascade }) => {
    const { batchChange } = useContext(BatchSpecContext)

    return (
        <div className={layoutStyles.pageContainer}>
            <div className={layoutStyles.headerContainer}>
                <BatchChangeHeader
                    namespace={{
                        to: `${batchChange.namespace.url}/batch-changes`,
                        text: batchChange.namespace.namespaceName,
                    }}
                    title={{ to: batchChange.url, text: batchChange.name }}
                    description={batchChange.description ?? undefined}
                />
            </div>

            <Switch>
                <Route render={() => <Redirect to={`${match.url}/execution`} />} path={match.url} exact={true} />
                <Route
                    path={`${match.url}/configuration`}
                    render={() => (
                        <>
                            <TabBar activeTabKey="configuration" tabsConfig={TABS_CONFIG} matchURL={match.url} />
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
                            <TabBar activeTabKey="spec" tabsConfig={TABS_CONFIG} matchURL={match.url} />
                            <ReadOnlyBatchSpecForm
                                batchChangeName={batchSpec.description.name}
                                originalInput={batchSpec.originalInput}
                                isLightTheme={isLightTheme}
                            />
                        </>
                    )}
                    exact={true}
                />
                <Route
                    path={`${match.url}/execution`}
                    render={() => (
                        <>
                            <TabBar activeTabKey="execution" tabsConfig={TABS_CONFIG} matchURL={match.url} />
                            <h1>EXECUTION</h1>
                        </>
                    )}
                />
                <Route
                    path={`${match.url}/preview`}
                    render={() => (
                        <>
                            <TabBar activeTabKey="preview" tabsConfig={TABS_CONFIG} matchURL={match.url} />
                            <h1>PREVIEW</h1>
                        </>
                    )}
                    exact={true}
                />
                <Route component={() => <HeroPage icon={MapSearchIcon} title="404: Not Found" />} key="hardcoded-key" />
            </Switch>
        </div>
    )
}

interface ReadOnlyBatchSpecFormProps extends ThemeProps {
    batchChangeName: string
    originalInput: string
}

const ReadOnlyBatchSpecForm: React.FunctionComponent<React.PropsWithChildren<ReadOnlyBatchSpecFormProps>> = ({
    batchChangeName,
    originalInput,
    isLightTheme,
}) => (
    <div className={editorStyles.form}>
        <LibraryPane name={batchChangeName} isReadOnly={true} />
        <div className={editorStyles.editorContainer}>
            <h4 className={editorStyles.header}>Batch spec</h4>
            {/* TODO: Banner should be different if execution finished... */}
            <Alert variant="warning" className="d-flex align-items-center pr-3">
                <div className="flex-grow-1">
                    <h4>The execution is still running</h4>
                    You are unable to edit the spec when an execution is running.
                </div>
                {/* TODO: Handle button */}
                <Button variant="danger">Cancel execution</Button>
            </Alert>
            <BatchSpec
                name={batchChangeName}
                className={editorStyles.editor}
                isLightTheme={isLightTheme}
                originalInput={originalInput}
            />
        </div>
        <WorkspacesPreviewPanel isReadOnly={true} />
    </div>
)

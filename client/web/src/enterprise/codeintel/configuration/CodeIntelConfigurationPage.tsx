import classNames from 'classnames'
import * as H from 'history'
import PencilIcon from 'mdi-react/PencilIcon'
import TrashIcon from 'mdi-react/TrashIcon'
import { editor } from 'monaco-editor'
import React, { FunctionComponent, useCallback, useEffect, useMemo, useState } from 'react'
import { RouteComponentProps } from 'react-router'

import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { GitObjectType } from '@sourcegraph/shared/src/graphql/schema'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { ErrorAlert } from '@sourcegraph/web/src/components/alerts'
import { PageTitle } from '@sourcegraph/web/src/components/PageTitle'
import { Button, Container, PageHeader } from '@sourcegraph/wildcard'

import { SaveToolbar, SaveToolbarProps, SaveToolbarPropsGenerator } from '../../../components/SaveToolbar'
import { CodeIntelligenceConfigurationPolicyFields } from '../../../graphql-operations'
import { DynamicallyImportedMonacoSettingsEditor } from '../../../settings/DynamicallyImportedMonacoSettingsEditor'

import {
    deletePolicyById as defaultDeletePolicyById,
    getConfigurationForRepository as defaultGetConfigurationForRepository,
    getInferredConfigurationForRepository as defaultGetInferredConfigurationForRepository,
    getPolicies as defaultGetPolicies,
    updateConfigurationForRepository as defaultUpdateConfigurationForRepository,
} from './backend'
import allConfigSchema from './schema.json'
import { formatDurationValue } from './shared'

export enum State {
    Idle,
    Deleting,
}

type SelectedTab = 'globalPolicies' | 'repositoryPolicies' | 'indexConfiguration'

export interface CodeIntelConfigurationPageProps extends RouteComponentProps<{}>, ThemeProps, TelemetryProps {
    repo?: { id: string }
    indexingEnabled?: boolean
    getPolicies?: typeof defaultGetPolicies
    updateConfigurationForRepository?: typeof defaultUpdateConfigurationForRepository
    deletePolicyById?: typeof defaultDeletePolicyById
    getConfigurationForRepository?: typeof defaultGetConfigurationForRepository
    getInferredConfigurationForRepository?: typeof defaultGetInferredConfigurationForRepository
    history: H.History

    /** For testing only. */
    openTab?: SelectedTab
}

export const CodeIntelConfigurationPage: FunctionComponent<CodeIntelConfigurationPageProps> = ({
    repo,
    indexingEnabled = window.context?.codeIntelAutoIndexingEnabled,
    getPolicies = defaultGetPolicies,
    updateConfigurationForRepository = defaultUpdateConfigurationForRepository,
    deletePolicyById = defaultDeletePolicyById,
    getConfigurationForRepository = defaultGetConfigurationForRepository,
    getInferredConfigurationForRepository = defaultGetInferredConfigurationForRepository,
    isLightTheme,
    telemetryService,
    history,
    openTab,
}) => {
    useEffect(() => telemetryService.logViewEvent('CodeIntelConfigurationPage'), [telemetryService])

    const [selectedTab, setSelectedTab] = useState<SelectedTab>(
        openTab ?? repo ? 'repositoryPolicies' : 'globalPolicies'
    )
    const [policies, setPolicies] = useState<CodeIntelligenceConfigurationPolicyFields[]>()
    const [globalPolicies, setGlobalPolicies] = useState<CodeIntelligenceConfigurationPolicyFields[]>()
    const [fetchError, setFetchError] = useState<Error>()

    useEffect(() => {
        const subscription = getPolicies().subscribe(policies => {
            setGlobalPolicies(policies)
        }, setFetchError)

        return () => subscription.unsubscribe()
    }, [getPolicies])

    useEffect(() => {
        if (!repo) {
            return
        }

        const subscription = getPolicies(repo.id).subscribe(policies => {
            setPolicies(policies)
        }, setFetchError)

        return () => subscription.unsubscribe()
    }, [repo, getPolicies])

    const [deleteError, setDeleteError] = useState<Error>()
    const [state, setState] = useState(() => State.Idle)

    const globalDeletePolicy = useCallback(
        async (id: string, name: string) => {
            if (!globalPolicies || !window.confirm(`Delete global policy ${name}?`)) {
                return
            }

            setState(State.Deleting)
            setDeleteError(undefined)

            try {
                await deletePolicyById(id).toPromise()
                setGlobalPolicies((globalPolicies || []).filter(policy => policy.id !== id))
            } catch (error) {
                setDeleteError(error)
            } finally {
                setState(State.Idle)
            }
        },
        [globalPolicies, deletePolicyById]
    )

    const deletePolicy = useCallback(
        async (id: string, name: string) => {
            if (!policies || !window.confirm(`Delete policy ${name}?`)) {
                return
            }

            setState(State.Deleting)
            setDeleteError(undefined)

            try {
                await deletePolicyById(id).toPromise()
                setPolicies((policies || []).filter(policy => policy.id !== id))
            } catch (error) {
                setDeleteError(error)
            } finally {
                setState(State.Idle)
            }
        },
        [policies, deletePolicyById]
    )

    const policyListButtonFragment = (
        <>
            <Button
                className="mt-2"
                variant="primary"
                onClick={() => history.push('./configuration/new')}
                disabled={state !== State.Idle}
            >
                Create new policy
            </Button>

            {state === State.Deleting && (
                <span className="ml-2 mt-2">
                    <LoadingSpinner className="icon-inline" /> Deleting...
                </span>
            )}
        </>
    )

    return fetchError ? (
        <ErrorAlert prefix="Error fetching configuration" error={fetchError} />
    ) : (
        <>
            <PageTitle title="Precise code intelligence configuration" />
            <PageHeader
                headingElement="h2"
                path={[
                    {
                        text: <>Precise code intelligence configuration</>,
                    },
                ]}
                description="Rules that define configuration for precise code intelligence indexes in this repository."
                className="mb-3"
            />

            {repo && (
                <CodeIntelligenceConfigurationTabHeader selectedTab={selectedTab} setSelectedTab={setSelectedTab} />
            )}

            <Container>
                {selectedTab === 'globalPolicies' ? (
                    <>
                        <h3>Global policies</h3>

                        {repo === undefined && deleteError && (
                            <ErrorAlert prefix="Error deleting configuration policy" error={deleteError} />
                        )}

                        <PoliciesList
                            policies={globalPolicies}
                            deletePolicy={repo ? undefined : globalDeletePolicy}
                            disabled={state !== State.Idle}
                            indexingEnabled={indexingEnabled}
                            buttonFragment={repo === undefined ? policyListButtonFragment : undefined}
                            history={history}
                        />
                    </>
                ) : selectedTab === 'repositoryPolicies' ? (
                    <>
                        <h3>Repository-specific policies</h3>

                        {deleteError && <ErrorAlert prefix="Error deleting configuration policy" error={deleteError} />}

                        <PoliciesList
                            policies={policies}
                            deletePolicy={deletePolicy}
                            disabled={state !== State.Idle}
                            indexingEnabled={indexingEnabled}
                            buttonFragment={policyListButtonFragment}
                            history={history}
                        />
                    </>
                ) : (
                    selectedTab === 'indexConfiguration' &&
                    repo &&
                    indexingEnabled && (
                        <>
                            <h3>Auto-indexing configuration</h3>

                            <ConfigurationEditor
                                repoId={repo.id}
                                updateConfigurationForRepository={updateConfigurationForRepository}
                                getConfigurationForRepository={getConfigurationForRepository}
                                getInferredConfigurationForRepository={getInferredConfigurationForRepository}
                                isLightTheme={isLightTheme}
                                telemetryService={telemetryService}
                                history={history}
                            />
                        </>
                    )
                )}
            </Container>
        </>
    )
}

const CodeIntelligenceConfigurationTabHeader: React.FunctionComponent<{
    selectedTab: SelectedTab
    setSelectedTab: (selectedTab: SelectedTab) => void
}> = ({ selectedTab, setSelectedTab }) => {
    const onGlobalPolicies = useCallback<React.MouseEventHandler>(
        event => {
            event.preventDefault()
            setSelectedTab('globalPolicies')
        },
        [setSelectedTab]
    )
    const onRepositoryPolicies = useCallback<React.MouseEventHandler>(
        event => {
            event.preventDefault()
            setSelectedTab('repositoryPolicies')
        },
        [setSelectedTab]
    )
    const onIndexConfiguratino = useCallback<React.MouseEventHandler>(
        event => {
            event.preventDefault()
            setSelectedTab('indexConfiguration')
        },
        [setSelectedTab]
    )

    return (
        <div className="overflow-auto mb-2">
            <ul className="nav nav-tabs d-inline-flex d-sm-flex flex-nowrap text-nowrap">
                <li className="nav-item">
                    {/* eslint-disable-next-line jsx-a11y/anchor-is-valid */}
                    <a
                        href=""
                        onClick={onGlobalPolicies}
                        className={classNames('nav-link', selectedTab === 'globalPolicies' && 'active')}
                        role="button"
                    >
                        <span className="text-content" data-tab-content="Global policies">
                            Global policies
                        </span>
                    </a>
                </li>
                <li className="nav-item">
                    {/* eslint-disable-next-line jsx-a11y/anchor-is-valid */}
                    <a
                        href=""
                        onClick={onRepositoryPolicies}
                        className={classNames('nav-link', selectedTab === 'repositoryPolicies' && 'active')}
                        role="button"
                    >
                        <span className="text-content" data-tab-content="Repository-specific policies">
                            Repository-specific policies
                        </span>
                    </a>
                </li>
                <li className="nav-item">
                    {/* eslint-disable-next-line jsx-a11y/anchor-is-valid */}
                    <a
                        href=""
                        onClick={onIndexConfiguratino}
                        className={classNames('nav-link', selectedTab === 'indexConfiguration' && 'active')}
                        role="button"
                    >
                        <span className="text-content" data-tab-content="Index configuration">
                            Index configuration
                        </span>
                    </a>
                </li>
            </ul>
        </div>
    )
}

interface PoliciesListProps {
    policies?: CodeIntelligenceConfigurationPolicyFields[]
    deletePolicy?: (id: string, name: string) => Promise<void>
    disabled: boolean
    indexingEnabled: boolean
    buttonFragment?: JSX.Element
    history: H.History
}

const PoliciesList: FunctionComponent<PoliciesListProps> = ({ policies, buttonFragment, ...props }) =>
    policies === undefined ? (
        <LoadingSpinner className="icon-inline" />
    ) : (
        <>
            {policies.length === 0 ? (
                <div>No policies have been defined.</div>
            ) : (
                <CodeIntelligencePolicyTable {...props} policies={policies} />
            )}
            {buttonFragment}
        </>
    )

const DescribeRetentionPolicy: FunctionComponent<{ policy: CodeIntelligenceConfigurationPolicyFields }> = ({
    policy,
}) =>
    policy.retentionEnabled ? (
        <p>
            <strong>Retention policy:</strong>{' '}
            <span>
                Retain uploads used to resolve code intelligence queries for{' '}
                {policy.type === GitObjectType.GIT_COMMIT
                    ? 'the matching commit'
                    : policy.type === GitObjectType.GIT_TAG
                    ? 'the matching tags'
                    : policy.type === GitObjectType.GIT_TREE
                    ? !policy.retainIntermediateCommits
                        ? 'the tip of the matching branches'
                        : 'any commit on the matching branches'
                    : ''}
                {policy.retentionDurationHours !== 0 &&
                    `for at least ${formatDurationValue(policy.retentionDurationHours)} after upload`}
                .
            </span>
        </p>
    ) : (
        <p className="text-muted">Data retention disabled.</p>
    )

const DescribeIndexingPolicy: FunctionComponent<{ policy: CodeIntelligenceConfigurationPolicyFields }> = ({ policy }) =>
    policy.indexingEnabled ? (
        <p>
            <strong>Indexing policy:</strong> Auto-index{' '}
            {policy.type === GitObjectType.GIT_COMMIT
                ? 'the matching commit'
                : policy.type === GitObjectType.GIT_TAG
                ? 'the matching tags'
                : policy.type === GitObjectType.GIT_TREE
                ? !policy.retainIntermediateCommits
                    ? 'the tip of the matching branches'
                    : 'any commit on the matching branches'
                : ''}
            {policy.indexCommitMaxAgeHours !== 0 &&
                ` if the target commit is no older than ${formatDurationValue(policy.indexCommitMaxAgeHours)}`}
            .
        </p>
    ) : (
        <p className="text-muted">Auto-indexing disabled.</p>
    )

interface CodeIntelligencePolicyTableProps {
    indexingEnabled: boolean
    disabled: boolean
    policies: CodeIntelligenceConfigurationPolicyFields[]
    deletePolicy?: (id: string, name: string) => Promise<void>
    history: H.History
}

const CodeIntelligencePolicyTable: FunctionComponent<CodeIntelligencePolicyTableProps> = ({
    indexingEnabled,
    disabled,
    policies,
    deletePolicy,
    history,
}) => (
    <div className="codeintel-configuration-policies__grid mb-3">
        {policies.map(policy => (
            <>
                <span className="codeintel-configuration-policy-node__separator" />

                <div className="d-flex flex-column codeintel-configuration-policy-node__name">
                    <div className="m-0">
                        <h3 className="m-0 d-block d-md-inline">{policy.name}</h3>
                    </div>

                    <div>
                        <div className="mr-2 d-block d-mdinline-block">
                            Applied to{' '}
                            {policy.type === GitObjectType.GIT_COMMIT
                                ? 'commits'
                                : policy.type === GitObjectType.GIT_TAG
                                ? 'tags'
                                : policy.type === GitObjectType.GIT_TREE
                                ? 'branches'
                                : ''}{' '}
                            matching <span className="text-monospace">{policy.pattern}</span>
                        </div>

                        <div>
                            {indexingEnabled && !policy.retentionEnabled && !policy.indexingEnabled ? (
                                <p className="text-muted mt-2">Data retention and auto-indexing disabled.</p>
                            ) : (
                                <>
                                    <p className="mt-2">
                                        <DescribeRetentionPolicy policy={policy} />
                                    </p>
                                    {indexingEnabled && (
                                        <p className="mt-2">
                                            <DescribeIndexingPolicy policy={policy} />
                                        </p>
                                    )}
                                </>
                            )}
                        </div>
                    </div>
                </div>

                <span className="d-none d-md-inline codeintel-configuration-policy-node__button">
                    {deletePolicy && (
                        <Button
                            onClick={() => history.push(`./configuration/${policy.id}`)}
                            className="p-0"
                            disabled={disabled}
                        >
                            <PencilIcon className="icon-inline" />
                        </Button>
                    )}
                </span>
                <span className="d-none d-md-inline codeintel-configuration-policy-node__button">
                    {deletePolicy && (
                        <Button
                            onClick={() => deletePolicy(policy.id, policy.name)}
                            className="ml-2 p-0"
                            disabled={disabled}
                        >
                            <TrashIcon className="icon-inline text-danger" />
                        </Button>
                    )}
                </span>
            </>
        ))}
    </div>
)

export interface ConfigurationEditorProps extends ThemeProps, TelemetryProps {
    repoId: string
    history: H.History
    getConfigurationForRepository: typeof defaultGetConfigurationForRepository
    getInferredConfigurationForRepository: typeof defaultGetInferredConfigurationForRepository
    updateConfigurationForRepository: typeof defaultUpdateConfigurationForRepository
}

enum EditorState {
    Idle,
    Saving,
}

export const ConfigurationEditor: FunctionComponent<ConfigurationEditorProps> = ({
    repoId,
    isLightTheme,
    telemetryService,
    history,
    getConfigurationForRepository,
    getInferredConfigurationForRepository,
    updateConfigurationForRepository,
}) => {
    const [configuration, setConfiguration] = useState<string>()
    const [inferredConfiguration, setInferredConfiguration] = useState<string>()
    const [fetchError, setFetchError] = useState<Error>()

    useEffect(() => {
        const subscription = getConfigurationForRepository(repoId).subscribe(config => {
            setConfiguration(config?.indexConfiguration?.configuration || '')
        }, setFetchError)

        return () => subscription.unsubscribe()
    }, [repoId, getConfigurationForRepository])

    useEffect(() => {
        const subscription = getInferredConfigurationForRepository(repoId).subscribe(config => {
            setInferredConfiguration(config?.indexConfiguration?.inferredConfiguration || '')
        }, setFetchError)

        return () => subscription.unsubscribe()
    }, [repoId, getInferredConfigurationForRepository])

    const [saveError, setSaveError] = useState<Error>()
    const [state, setState] = useState(() => EditorState.Idle)

    const save = useCallback(
        async (content: string) => {
            setState(EditorState.Saving)
            setSaveError(undefined)

            try {
                await updateConfigurationForRepository(repoId, content).toPromise()
                setDirty(false)
                setConfiguration(content)
            } catch (error) {
                setSaveError(error)
            } finally {
                setState(EditorState.Idle)
            }
        },
        [repoId, updateConfigurationForRepository]
    )

    const [dirty, setDirty] = useState<boolean>()
    const [editor, setEditor] = useState<editor.ICodeEditor>()
    const infer = useCallback(() => editor?.setValue(inferredConfiguration || ''), [editor, inferredConfiguration])

    const customToolbar = useMemo<{
        saveToolbar: React.FunctionComponent<SaveToolbarProps & AutoIndexProps>
        propsGenerator: SaveToolbarPropsGenerator<AutoIndexProps>
    }>(
        () => ({
            saveToolbar: CodeIntelAutoIndexSaveToolbar,
            propsGenerator: props => {
                const mergedProps = {
                    ...props,
                    onInfer: infer,
                    loading: inferredConfiguration === undefined,
                    inferEnabled: !!inferredConfiguration && configuration !== inferredConfiguration,
                }
                mergedProps.willShowError = () => !mergedProps.saving
                mergedProps.saveDiscardDisabled = () => mergedProps.saving || !dirty

                return mergedProps
            },
        }),
        [dirty, configuration, inferredConfiguration, infer]
    )

    return fetchError ? (
        <ErrorAlert prefix="Error fetching index configuration" error={fetchError} />
    ) : (
        <>
            {saveError && <ErrorAlert prefix="Error saving index configuration" error={saveError} />}

            {configuration === undefined ? (
                <LoadingSpinner className="icon-inline" />
            ) : (
                <DynamicallyImportedMonacoSettingsEditor
                    value={configuration}
                    jsonSchema={allConfigSchema}
                    canEdit={true}
                    onSave={save}
                    saving={state === EditorState.Saving}
                    height={600}
                    isLightTheme={isLightTheme}
                    history={history}
                    telemetryService={telemetryService}
                    customSaveToolbar={customToolbar}
                    onDirtyChange={setDirty}
                    onEditor={setEditor}
                />
            )}
        </>
    )
}

interface AutoIndexProps {
    loading: boolean
    inferEnabled: boolean
    onInfer?: () => void
}

const CodeIntelAutoIndexSaveToolbar: React.FunctionComponent<SaveToolbarProps & AutoIndexProps> = ({
    dirty,
    loading,
    saving,
    error,
    onSave,
    onDiscard,
    inferEnabled,
    onInfer,
    saveDiscardDisabled,
}) => (
    <SaveToolbar
        dirty={dirty}
        saving={saving}
        onSave={onSave}
        error={error}
        saveDiscardDisabled={saveDiscardDisabled}
        onDiscard={onDiscard}
    >
        {loading ? (
            <LoadingSpinner className="icon-inline mt-2 ml-2" />
        ) : (
            inferEnabled && (
                <Button type="button" title="Infer index configuration from HEAD" variant="link" onClick={onInfer}>
                    Infer index configuration from HEAD
                </Button>
            )
        )}
    </SaveToolbar>
)

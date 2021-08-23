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

export interface CodeIntelConfigurationPageProps extends RouteComponentProps<{}>, ThemeProps, TelemetryProps {
    repo?: { id: string }
    indexingEnabled?: boolean
    getPolicies?: typeof defaultGetPolicies
    updateConfigurationForRepository?: typeof defaultUpdateConfigurationForRepository
    deletePolicyById?: typeof defaultDeletePolicyById
    getConfigurationForRepository?: typeof defaultGetConfigurationForRepository
    getInferredConfigurationForRepository?: typeof defaultGetInferredConfigurationForRepository
    history: H.History
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
}) => {
    useEffect(() => telemetryService.logViewEvent('CodeIntelConfigurationPage'), [telemetryService])

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
                className="btn btn-primary mt-2"
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

            <Container>
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
            </Container>

            {repo && (
                <RepoConfiguration
                    repoId={repo.id}
                    policies={policies}
                    deletePolicy={deletePolicy}
                    deleteError={deleteError}
                    disabled={state !== State.Idle}
                    indexingEnabled={indexingEnabled}
                    policyListButtonFragment={policyListButtonFragment}
                    getConfigurationForRepository={getConfigurationForRepository}
                    getInferredConfigurationForRepository={getInferredConfigurationForRepository}
                    updateConfigurationForRepository={updateConfigurationForRepository}
                    isLightTheme={isLightTheme}
                    telemetryService={telemetryService}
                    history={history}
                />
            )}
        </>
    )
}

interface RepoConfigurationProps extends ThemeProps, TelemetryProps {
    repoId: string
    policies?: CodeIntelligenceConfigurationPolicyFields[]
    deletePolicy: (id: string, name: string) => Promise<void>
    deleteError?: Error
    disabled: boolean
    indexingEnabled: boolean
    policyListButtonFragment: JSX.Element
    getConfigurationForRepository: typeof defaultGetConfigurationForRepository
    getInferredConfigurationForRepository: typeof defaultGetInferredConfigurationForRepository
    updateConfigurationForRepository: typeof defaultUpdateConfigurationForRepository
    history: H.History
}

const RepoConfiguration: FunctionComponent<RepoConfigurationProps> = ({
    policies,
    deletePolicy,
    deleteError,
    disabled,
    indexingEnabled,
    policyListButtonFragment,
    ...props
}) => (
    <>
        <Container className="mt-2">
            <h3>Repository-specific policies</h3>

            {deleteError && <ErrorAlert prefix="Error deleting configuration policy" error={deleteError} />}

            <PoliciesList
                policies={policies}
                deletePolicy={deletePolicy}
                disabled={disabled}
                indexingEnabled={indexingEnabled}
                buttonFragment={policyListButtonFragment}
                history={props.history}
            />
        </Container>

        <Container className="mt-2 code-intel-index-configuration">
            <h3>Auto-indexing configuration</h3>

            <ConfigurationEditor {...props} />
        </Container>
    </>
)

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
            <CodeIntelligencePolicyTable {...props} policies={policies} />

            {buttonFragment}
        </>
    )

interface CodeIntelligencePolicyTableProps {
    indexingEnabled: boolean
    disabled: boolean
    policies: CodeIntelligenceConfigurationPolicyFields[]
    deletePolicy?: (id: string, name: string) => Promise<void>
    history: H.History
}

const describeRetentionPolicy: (policy: CodeIntelligenceConfigurationPolicyFields) => JSX.Element = policy =>
    policy.retentionEnabled ? (
        <p>
            <strong>Retention policy:</strong>{' '}
            <span>
                Retain uploads used to resolve code intelligence queries for{' '}
                {policy.type === GitObjectType.GIT_TREE
                    ? !policy.retainIntermediateCommits
                        ? 'the tip of this branch'
                        : 'any commit on this branch'
                    : `this ${policy.type}`}{' '}
                {policy.retentionDurationHours === 0 ? (
                    <></>
                ) : (
                    <>for at least {formatDurationValue(policy.retentionDurationHours)} after upload</>
                )}
                .
            </span>
        </p>
    ) : (
        <p className="text-muted">Data retention disabled.</p>
    )

const describeIndexingPolicy: (policy: CodeIntelligenceConfigurationPolicyFields) => JSX.Element = policy =>
    policy.indexingEnabled ? (
        <p>
            <strong>Indexing policy:</strong> Auto-index{' '}
            {policy.type === GitObjectType.GIT_TREE
                ? !policy.indexIntermediateCommits
                    ? 'the tip of this branch'
                    : 'all commits on this branch'
                : `this ${policy.type}`}{' '}
            {policy.indexCommitMaxAgeHours === 0 ? (
                <></>
            ) : (
                <>if the target commit is no older than {formatDurationValue(policy.indexCommitMaxAgeHours)} </>
            )}
            .
        </p>
    ) : (
        <p className="text-muted">Auto-indexing disabled.</p>
    )

const CodeIntelligencePolicyTable: FunctionComponent<CodeIntelligencePolicyTableProps> = ({
    indexingEnabled,
    disabled,
    policies,
    deletePolicy,
    history,
}) => (
    <table className="table table-striped table-borderless">
        <thead>
            <tr>
                <th>Rule name</th>
                <th>Type</th>
                <th>Pattern</th>
                <th>Policy</th>
                {deletePolicy && <th>Actions</th>}
            </tr>
        </thead>
        <tbody>
            {policies.map(policy => (
                <tr key={policy.id}>
                    <td>{policy.name}</td>
                    <td>
                        {policy.type === GitObjectType.GIT_COMMIT
                            ? 'commit'
                            : policy.type === GitObjectType.GIT_TAG
                            ? 'tag'
                            : policy.type === GitObjectType.GIT_TREE
                            ? 'branch'
                            : ''}
                    </td>
                    <td className="text-monospace">{policy.pattern}</td>
                    <td>
                        {indexingEnabled && !policy.retentionEnabled && !policy.indexingEnabled ? (
                            <p className="text-muted">Data retention and auto-indexing disabled.</p>
                        ) : (
                            <>
                                {describeRetentionPolicy(policy)}
                                {indexingEnabled && describeIndexingPolicy(policy)}
                            </>
                        )}
                    </td>
                    {deletePolicy && (
                        <td>
                            <Button
                                onClick={() => history.push(`./configuration/${policy.id}`)}
                                className="p-0"
                                disabled={disabled}
                            >
                                <PencilIcon className="icon-inline" />
                            </Button>
                            <Button
                                onClick={() => deletePolicy(policy.id, policy.name)}
                                className="ml-2 p-0"
                                disabled={disabled}
                            >
                                <TrashIcon className="icon-inline text-danger" />
                            </Button>
                        </td>
                    )}
                </tr>
            ))}
        </tbody>
    </table>
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
                <button
                    type="button"
                    title="Infer index configuration from HEAD"
                    className="btn btn-link"
                    onClick={onInfer}
                >
                    Infer index configuration from HEAD
                </button>
            )
        )}
    </SaveToolbar>
)

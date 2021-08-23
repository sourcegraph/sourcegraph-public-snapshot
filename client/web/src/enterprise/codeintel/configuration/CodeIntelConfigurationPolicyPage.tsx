import * as H from 'history'
import { debounce } from 'lodash'
import React, { FunctionComponent, useCallback, useEffect, useState } from 'react'
import { RouteComponentProps } from 'react-router'
import { of } from 'rxjs'

import { Toggle } from '@sourcegraph/branded/src/components/Toggle'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { ErrorAlert } from '@sourcegraph/web/src/components/alerts'
import { PageTitle } from '@sourcegraph/web/src/components/PageTitle'
import { Container, LoadingSpinner, PageHeader } from '@sourcegraph/wildcard'

import { CodeIntelligenceConfigurationPolicyFields, GitObjectType } from '../../../graphql-operations'

import {
    getPolicyById as defaultGetPolicyById,
    repoName as defaultRepoName,
    searchGitBranches as defaultSearchGitBranches,
    searchGitTags as defaultSearchGitTags,
    updatePolicy as defaultUpdatePolicy,
} from './backend'
import { defaultDurationValues } from './shared'

export interface CodeIntelConfigurationPolicyPageProps
    extends RouteComponentProps<{ id: string }>,
        ThemeProps,
        TelemetryProps {
    repo?: { id: string }
    indexingEnabled?: boolean
    getPolicyById?: typeof defaultGetPolicyById
    repoName?: typeof defaultRepoName
    searchGitBranches?: typeof defaultSearchGitBranches
    searchGitTags?: typeof defaultSearchGitTags
    updatePolicy?: typeof defaultUpdatePolicy
    history: H.History
}

enum State {
    Idle,
    Saving,
}

const emptyPolicy: CodeIntelligenceConfigurationPolicyFields = {
    __typename: 'CodeIntelligenceConfigurationPolicy',
    id: '',
    name: '',
    type: GitObjectType.GIT_COMMIT,
    pattern: '',
    retentionEnabled: false,
    retentionDurationHours: 0,
    retainIntermediateCommits: false,
    indexingEnabled: false,
    indexCommitMaxAgeHours: 0,
    indexIntermediateCommits: false,
}

export const CodeIntelConfigurationPolicyPage: FunctionComponent<CodeIntelConfigurationPolicyPageProps> = ({
    match: {
        params: { id },
    },
    repo,
    indexingEnabled = window.context?.codeIntelAutoIndexingEnabled,
    getPolicyById = defaultGetPolicyById,
    repoName = defaultRepoName,
    searchGitBranches = defaultSearchGitBranches,
    searchGitTags = defaultSearchGitTags,
    updatePolicy = defaultUpdatePolicy,
    history,
    telemetryService,
}) => {
    useEffect(() => telemetryService.logViewEvent('CodeIntelConfigurationPolicyPageProps'), [telemetryService])

    const [saved, setSaved] = useState<CodeIntelligenceConfigurationPolicyFields>()
    const [policy, setPolicy] = useState<CodeIntelligenceConfigurationPolicyFields>()
    const [fetchError, setFetchError] = useState<Error>()

    useEffect(() => {
        const subscription = (id === 'new' ? of(emptyPolicy) : getPolicyById(id)).subscribe(policy => {
            setSaved(policy)
            setPolicy(policy)
        }, setFetchError)

        return () => subscription.unsubscribe()
    }, [id, getPolicyById])

    const [saveError, setSaveError] = useState<Error>()
    const [state, setState] = useState(() => State.Idle)

    const save = useCallback(async () => {
        if (!policy) {
            return
        }

        let navigatingAway = false
        setState(State.Saving)
        setSaveError(undefined)

        try {
            await updatePolicy(policy, repo?.id).toPromise()
            history.push('./')
            navigatingAway = true
        } catch (error) {
            setSaveError(error)
        } finally {
            if (!navigatingAway) {
                setState(State.Idle)
            }
        }
    }, [repo?.id, policy, updatePolicy, history])

    return fetchError ? (
        <ErrorAlert prefix="Error fetching configuration policy" error={fetchError} />
    ) : (
        <>
            <PageTitle title="Precise code intelligence configuration policy" />
            <PageHeader
                headingElement="h2"
                path={[
                    {
                        text: <>{policy?.id === '' ? 'Create configuration policy' : 'Update configuration policy'}</>,
                    },
                ]}
                className="mb-3"
            />

            {policy === undefined ? (
                <LoadingSpinner className="icon-inline" />
            ) : (
                <>
                    <Container className="container form">
                        {saveError && <ErrorAlert prefix="Error saving configuration policy" error={saveError} />}
                        <BranchTargetSettings
                            repoId={repo?.id}
                            policy={policy}
                            setPolicy={setPolicy}
                            repoName={repoName}
                            searchGitBranches={searchGitBranches}
                            searchGitTags={searchGitTags}
                        />
                    </Container>

                    <RetentionSettings policy={policy} setPolicy={setPolicy} />
                    {indexingEnabled && <IndexingSettings policy={policy} setPolicy={setPolicy} />}

                    <Container className="mt-2">
                        <button
                            type="submit"
                            className="btn btn-primary"
                            onClick={save}
                            disabled={state !== State.Idle || comparePolicies(policy, saved)}
                        >
                            {policy.id === '' ? 'Create' : 'Update'} policy
                        </button>

                        <button
                            type="button"
                            className="btn btn-secondary ml-3"
                            onClick={() => history.push('./')}
                            disabled={state !== State.Idle}
                        >
                            Cancel
                        </button>

                        {state === State.Saving && (
                            <span className="ml-2">
                                <LoadingSpinner className="icon-inline" /> Saving...
                            </span>
                        )}
                    </Container>
                </>
            )}
        </>
    )
}

interface BranchTargetSettingsProps {
    repoId?: string
    policy: CodeIntelligenceConfigurationPolicyFields
    setPolicy: (p: CodeIntelligenceConfigurationPolicyFields) => void
    repoName: typeof defaultRepoName
    searchGitBranches: typeof defaultSearchGitBranches
    searchGitTags: typeof defaultSearchGitTags
}

const GIT_OBJECT_PREVIEW_DEBOUNCE_TIMEOUT = 300

const BranchTargetSettings: FunctionComponent<BranchTargetSettingsProps> = ({
    repoId,
    policy,
    setPolicy,
    repoName,
    searchGitBranches,
    searchGitTags,
}) => {
    const [debouncedPattern, setDebouncedPattern] = useState(policy.pattern)
    const setPattern = debounce(value => setDebouncedPattern(value), GIT_OBJECT_PREVIEW_DEBOUNCE_TIMEOUT)

    return (
        <>
            <div className="form-group">
                <label htmlFor="name">Name</label>
                <input
                    id="name"
                    type="text"
                    className="form-control"
                    value={policy.name}
                    onChange={event => setPolicy({ ...policy, name: event.target.value })}
                />
            </div>

            <div className="form-group">
                <label htmlFor="type">Type</label>
                <select
                    id="type"
                    className="form-control"
                    value={policy.type}
                    onChange={event =>
                        setPolicy({
                            ...policy,
                            type: event.target.value as GitObjectType,
                            ...(event.target.value !== GitObjectType.GIT_TREE
                                ? {
                                      retainIntermediateCommits: false,
                                      indexIntermediateCommits: false,
                                  }
                                : {}),
                        })
                    }
                >
                    <option value="">Select Git object type</option>
                    {repoId && <option value={GitObjectType.GIT_COMMIT}>Commit</option>}
                    <option value={GitObjectType.GIT_TAG}>Tag</option>
                    <option value={GitObjectType.GIT_TREE}>Branch</option>
                </select>
            </div>
            <div className="form-group">
                <label htmlFor="pattern">Pattern</label>
                <input
                    id="pattern"
                    type="text"
                    className="form-control text-monospace"
                    value={policy.pattern}
                    onChange={event => {
                        setPolicy({ ...policy, pattern: event.target.value })
                        setPattern(event.target.value)
                    }}
                />
            </div>

            {repoId && (
                <GitObjectPreview
                    pattern={debouncedPattern}
                    repoId={repoId}
                    type={policy.type}
                    repoName={repoName}
                    searchGitTags={searchGitTags}
                    searchGitBranches={searchGitBranches}
                />
            )}
        </>
    )
}

interface RetentionSettingsProps {
    policy: CodeIntelligenceConfigurationPolicyFields
    setPolicy: (p: CodeIntelligenceConfigurationPolicyFields) => void
}

const RetentionSettings: FunctionComponent<RetentionSettingsProps> = ({ policy, setPolicy }) => (
    <Container className="mt-2">
        <h3>Retention</h3>

        <div className="form-group">
            <Toggle
                id="retention-enabled"
                title="Enabled"
                value={policy.retentionEnabled}
                onToggle={value => setPolicy({ ...policy, retentionEnabled: value })}
            />
            <label htmlFor="retention-enabled" className="ml-2">
                Enabled / disabled
            </label>
        </div>

        <div className="form-group">
            <label htmlFor="retention-duration">Duration</label>

            <DurationSelection
                id="retention-duration"
                value={`${policy.retentionDurationHours}`}
                onChange={value => setPolicy({ ...policy, retentionDurationHours: value })}
                disabled={!policy.retentionEnabled}
            />
        </div>

        {policy.type === GitObjectType.GIT_TREE && (
            <div className="form-group">
                <Toggle
                    id="retain-intermediate-commits"
                    title="Enabled"
                    value={policy.retainIntermediateCommits}
                    onToggle={value => setPolicy({ ...policy, retainIntermediateCommits: value })}
                    disabled={!policy.retentionEnabled}
                />
                <label htmlFor="retain-intermediate-commits" className="ml-2">
                    Retain intermediate commits
                </label>
            </div>
        )}
    </Container>
)

interface IndexingSettingsProps {
    policy: CodeIntelligenceConfigurationPolicyFields
    setPolicy: (p: CodeIntelligenceConfigurationPolicyFields) => void
}

const IndexingSettings: FunctionComponent<IndexingSettingsProps> = ({ policy, setPolicy }) => (
    <Container className="mt-2">
        <h3>Auto-indexing</h3>

        <div className="form-group">
            <Toggle
                id="indexing-enabled"
                title="Enabled"
                value={policy.indexingEnabled}
                onToggle={value => setPolicy({ ...policy, indexingEnabled: value })}
            />
            <label htmlFor="indexing-enabled" className="ml-2">
                Enabled / disabled
            </label>
        </div>

        <div className="form-group">
            <label htmlFor="index-commit-max-age">Commit max age</label>
            <DurationSelection
                id="index-commit-max-age"
                value={`${policy.indexCommitMaxAgeHours}`}
                disabled={!policy.indexingEnabled}
                onChange={value => setPolicy({ ...policy, indexCommitMaxAgeHours: value })}
            />
        </div>

        {policy.type === GitObjectType.GIT_TREE && (
            <div className="form-group">
                <Toggle
                    id="index-intermediate-commits"
                    title="Enabled"
                    value={policy.indexIntermediateCommits}
                    onToggle={value => setPolicy({ ...policy, indexIntermediateCommits: value })}
                    disabled={!policy.indexingEnabled}
                />
                <label htmlFor="index-intermediate-commits" className="ml-2">
                    Index intermediate commits
                </label>
            </div>
        )}
    </Container>
)

interface GitObjectPreviewProps {
    repoId: string
    type: GitObjectType
    pattern: string
    repoName: typeof defaultRepoName
    searchGitTags: typeof defaultSearchGitTags
    searchGitBranches: typeof defaultSearchGitBranches
}

enum PreviewState {
    Idle,
    LoadingTags,
}

interface GitObjectPreviewResult {
    preview: { name: string; revlike: string }[]
    totalCount: number
}

const resultFromCommit = async (
    repoId: string,
    pattern: string,
    repoName: typeof defaultRepoName
): Promise<GitObjectPreviewResult> => {
    const result = await repoName(repoId).toPromise()
    if (!result) {
        return { preview: [], totalCount: 0 }
    }

    return { preview: [{ name: result.name, revlike: pattern }], totalCount: 1 }
}

const resultFromTag = async (
    repoId: string,
    pattern: string,
    searchGitTags: typeof defaultSearchGitTags
): Promise<GitObjectPreviewResult> => {
    const result = await searchGitTags(repoId, pattern).toPromise()
    if (!result) {
        return { preview: [], totalCount: 0 }
    }

    const { nodes, totalCount } = result.tags

    return {
        preview: nodes.map(node => ({ name: result.name, revlike: node.displayName })),
        totalCount,
    }
}

const resultFromBranch = async (
    repoId: string,
    pattern: string,
    searchGitBranches: typeof defaultSearchGitBranches
): Promise<GitObjectPreviewResult> => {
    const result = await searchGitBranches(repoId, pattern).toPromise()
    if (!result) {
        return { preview: [], totalCount: 0 }
    }

    const { nodes, totalCount } = result.branches

    return {
        preview: nodes.map(node => ({ name: result.name, revlike: node.displayName })),
        totalCount,
    }
}

const GitObjectPreview: FunctionComponent<GitObjectPreviewProps> = ({
    repoId,
    type,
    pattern,
    repoName,
    searchGitTags,
    searchGitBranches,
}) => {
    const [state, setState] = useState(() => PreviewState.Idle)
    const [commitPreview, setCommitPreview] = useState<GitObjectPreviewResult>()
    const [commitPreviewFetchError, setCommitPreviewFetchError] = useState<Error>()

    useEffect(() => {
        async function inner(): Promise<void> {
            setState(PreviewState.LoadingTags)
            setCommitPreviewFetchError(undefined)

            const resultFactories = [
                { type: GitObjectType.GIT_COMMIT, factory: () => resultFromCommit(repoId, pattern, repoName) },
                { type: GitObjectType.GIT_TAG, factory: () => resultFromTag(repoId, pattern, searchGitTags) },
                { type: GitObjectType.GIT_TREE, factory: () => resultFromBranch(repoId, pattern, searchGitBranches) },
            ]

            try {
                for (const { type: match, factory } of resultFactories) {
                    if (type === match) {
                        setCommitPreview(await factory())
                        break
                    }
                }
            } catch (error) {
                setCommitPreviewFetchError(error)
            } finally {
                setState(PreviewState.Idle)
            }
        }

        inner().catch(console.error)
    }, [repoId, type, pattern, repoName, searchGitTags, searchGitBranches])

    return (
        <>
            <h3>Preview of Git object filter</h3>

            {type ? (
                <>
                    <small>
                        {commitPreview?.preview.length === 0 ? (
                            <>Configuration policy does not match any known commits.</>
                        ) : (
                            <>
                                Configuration policy will be applied to the following
                                {type === GitObjectType.GIT_COMMIT
                                    ? ' commit'
                                    : type === GitObjectType.GIT_TAG
                                    ? ' tags'
                                    : type === GitObjectType.GIT_TREE
                                    ? ' branches'
                                    : ''}
                                .
                            </>
                        )}
                    </small>

                    {commitPreviewFetchError ? (
                        <ErrorAlert
                            prefix="Error fetching matching repository objects"
                            error={commitPreviewFetchError}
                        />
                    ) : (
                        <>
                            {commitPreview !== undefined && commitPreview.preview.length !== 0 && (
                                <div className="mt-2 p-2">
                                    <div className="bg-dark text-light">
                                        {commitPreview.preview.map(tag => (
                                            <p key={tag.revlike} className="text-monospace p-0 m-0">
                                                <span className="search-filter-keyword">repo:</span>
                                                <span>{tag.name}</span>
                                                <span className="search-filter-keyword">@</span>
                                                <span>{tag.revlike}</span>
                                            </p>
                                        ))}
                                    </div>

                                    {commitPreview.preview.length < commitPreview.totalCount && (
                                        <p className="pt-2">
                                            ...and {commitPreview.totalCount - commitPreview.preview.length} other
                                            matches
                                        </p>
                                    )}
                                </div>
                            )}
                            {state === PreviewState.LoadingTags && <LoadingSpinner />}
                        </>
                    )}
                </>
            ) : (
                <small>Select a Git object type to preview matching commits.</small>
            )}
        </>
    )
}

interface DurationSelectionProps {
    id: string
    value: string
    disabled: boolean
    onChange?: (value: number) => void
    durationValues?: { value: number; displayText: string }[]
}

const DurationSelection: FunctionComponent<DurationSelectionProps> = ({
    id,
    value,
    disabled,
    onChange,
    durationValues = defaultDurationValues,
}) => (
    <select
        id={id}
        className="form-control"
        value={value}
        disabled={disabled}
        onChange={event => onChange?.(Math.floor(parseInt(event.target.value, 10)))}
    >
        <option value="">Select duration</option>

        {durationValues.map(({ value, displayText }) => (
            <option key={value} value={value}>
                {displayText}
            </option>
        ))}
    </select>
)

function comparePolicies(
    a: CodeIntelligenceConfigurationPolicyFields,
    b?: CodeIntelligenceConfigurationPolicyFields
): boolean {
    return (
        b !== undefined &&
        a.id === b.id &&
        a.name === b.name &&
        a.type === b.type &&
        a.pattern === b.pattern &&
        a.retentionEnabled === b.retentionEnabled &&
        a.retentionDurationHours === b.retentionDurationHours &&
        a.retainIntermediateCommits === b.retainIntermediateCommits &&
        a.indexingEnabled === b.indexingEnabled &&
        a.indexCommitMaxAgeHours === b.indexCommitMaxAgeHours &&
        a.indexIntermediateCommits === b.indexIntermediateCommits
    )
}

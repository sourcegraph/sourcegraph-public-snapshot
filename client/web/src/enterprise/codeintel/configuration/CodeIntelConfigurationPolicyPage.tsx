import * as H from 'history'
import React, { FunctionComponent, useCallback, useEffect, useState } from 'react'
import { RouteComponentProps } from 'react-router'
import { of } from 'rxjs'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { ErrorAlert } from '@sourcegraph/web/src/components/alerts'
import { PageTitle } from '@sourcegraph/web/src/components/PageTitle'
import { Button, Container, LoadingSpinner, PageHeader } from '@sourcegraph/wildcard'

import { CodeIntelligenceConfigurationPolicyFields, GitObjectType } from '../../../graphql-operations'

import {
    getPolicyById as defaultGetPolicyById,
    repoName as defaultRepoName,
    searchGitBranches as defaultSearchGitBranches,
    searchGitTags as defaultSearchGitTags,
    updatePolicy as defaultUpdatePolicy,
} from './backend'
import { BranchTargetSettings } from './BranchTargetSettings'
import { IndexingSettings } from './IndexSettings'
import { RetentionSettings } from './RetentionSettings'

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
                        text: <>{policy?.id === '' ? 'Create' : 'Update'} configuration policy</>,
                    },
                ]}
                description={`${policy?.id === '' ? 'Create' : 'Update'} a new configuration policy that applies to ${
                    repo ? 'this repository' : 'all repositories'
                }.`}
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
                        <Button
                            type="submit"
                            variant="primary"
                            onClick={save}
                            disabled={state !== State.Idle || comparePolicies(policy, saved)}
                        >
                            {policy.id === '' ? 'Create' : 'Update'} policy
                        </Button>

                        <Button
                            type="button"
                            className="ml-3"
                            variant="secondary"
                            onClick={() => history.push('./')}
                            disabled={state !== State.Idle}
                        >
                            Cancel
                        </Button>

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

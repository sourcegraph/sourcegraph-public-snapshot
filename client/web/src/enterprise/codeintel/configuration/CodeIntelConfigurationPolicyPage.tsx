import { ApolloError } from '@apollo/client'
import * as H from 'history'
import React, { FunctionComponent, useCallback, useEffect, useState } from 'react'
import { RouteComponentProps } from 'react-router'

import { GitObjectType } from '@sourcegraph/shared/src/graphql-operations'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { ErrorAlert } from '@sourcegraph/web/src/components/alerts'
import { PageTitle } from '@sourcegraph/web/src/components/PageTitle'
import { Button, Container, LoadingSpinner, PageHeader } from '@sourcegraph/wildcard'

import { CodeIntelligenceConfigurationPolicyFields } from '../../../graphql-operations'

import { BranchTargetSettings } from './BranchTargetSettings'
import { IndexingSettings } from './IndexSettings'
import { RetentionSettings } from './RetentionSettings'
import { usePolicyConfigurationByID, useSavePolicyConfiguration } from './usePoliciesConfigurations'

export interface CodeIntelConfigurationPolicyPageProps
    extends RouteComponentProps<{ id: string }>,
        ThemeProps,
        TelemetryProps {
    repo?: { id: string }
    indexingEnabled?: boolean
    history: H.History
}

export const CodeIntelConfigurationPolicyPage: FunctionComponent<CodeIntelConfigurationPolicyPageProps> = ({
    match: {
        params: { id },
    },
    repo,
    indexingEnabled = window.context?.codeIntelAutoIndexingEnabled,
    history,
    telemetryService,
}) => {
    useEffect(() => telemetryService.logViewEvent('CodeIntelConfigurationPolicyPageProps'), [telemetryService])

    const { policyConfig, loadingPolicyConfig, policyConfigError } = usePolicyConfigurationByID(id)
    const [saved, setSaved] = useState<CodeIntelligenceConfigurationPolicyFields>()
    const [policy, setPolicy] = useState<CodeIntelligenceConfigurationPolicyFields | undefined>()

    const { savePolicyConfiguration, isSaving, savingError } = useSavePolicyConfiguration(policy?.id === '')

    useEffect(() => {
        setPolicy(policyConfig)
        setSaved(policyConfig)
    }, [policyConfig])

    const savePolicyConfig = useCallback(async () => {
        if (!policy) {
            return
        }

        const variables = repo?.id ? { ...policy, repositoryId: repo.id ?? null } : { ...policy }
        variables.pattern = variables.type === GitObjectType.GIT_COMMIT ? 'HEAD' : variables.pattern

        return savePolicyConfiguration({ variables })
            .then(() =>
                history.push({
                    pathname: './',
                    state: { modal: 'SUCCESS', message: `Configuration for policy ${policy.name} has been saved.` },
                })
            )
            .catch((error: ApolloError) =>
                history.push({
                    pathname: './',
                    state: {
                        modal: 'ERROR',
                        message: `There was an error while saving policy: ${policy.name}. See error: ${error.message}`,
                    },
                })
            )
    }, [policy, repo, savePolicyConfiguration, history])

    if (loadingPolicyConfig) {
        return <LoadingSpinner className="icon-inline" />
    }

    if (policyConfigError || policy === undefined) {
        return <ErrorAlert prefix="Error fetching configuration policy" error={policyConfigError} />
    }

    return (
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

            {policy.protected && (
                <div className="alert alert-info">
                    This configuration policy is protected. Protected configuration policies may not be deleted and only
                    the retention duration and indexing options are editable.
                </div>
            )}

            <Container className="container form mb-3">
                {savingError && <ErrorAlert prefix="Error saving configuration policy" error={savingError} />}
                <BranchTargetSettings
                    repoId={repo?.id}
                    policy={policy}
                    setPolicy={setPolicy}
                    disabled={policy.protected}
                />

                <RetentionSettings policy={policy} setPolicy={setPolicy} />

                {indexingEnabled && <IndexingSettings repo={repo} policy={policy} setPolicy={setPolicy} />}
            </Container>

            <div className="mb-3">
                <Button
                    type="submit"
                    variant="primary"
                    onClick={savePolicyConfig}
                    disabled={isSaving || !validatePolicy(policy) || comparePolicies(policy, saved)}
                >
                    {!isSaving && <>{policy.id === '' ? 'Create' : 'Update'} policy</>}
                    {isSaving && (
                        <>
                            <LoadingSpinner className="icon-inline" /> Saving...
                        </>
                    )}
                </Button>

                <Button
                    type="button"
                    className="ml-2"
                    variant="secondary"
                    onClick={() => history.push('./')}
                    disabled={isSaving}
                >
                    Cancel
                </Button>
            </div>
        </>
    )
}

function validatePolicy(policy: CodeIntelligenceConfigurationPolicyFields): boolean {
    const invalidConditions = [
        // Name is required
        policy.name === '',

        // Pattern is required if policy type is GIT_COMMIT
        policy.type !== GitObjectType.GIT_COMMIT && policy.pattern === '',

        // If repository patterns are supplied they must be non-empty
        policy.repositoryPatterns?.some(pattern => pattern === ''),

        // Policy type must be GIT_{COMMIT,TAG,TREE}
        ![GitObjectType.GIT_COMMIT, GitObjectType.GIT_TAG, GitObjectType.GIT_TREE].includes(policy.type),

        // If numeric values are supplied they must be non-negative
        policy.retentionDurationHours && policy.retentionDurationHours <= 0,
        policy.indexCommitMaxAgeHours && policy.indexCommitMaxAgeHours <= 0,
    ]

    return invalidConditions.every(isInvalid => !isInvalid)
}

function comparePolicies(
    a: CodeIntelligenceConfigurationPolicyFields,
    b?: CodeIntelligenceConfigurationPolicyFields
): boolean {
    if (b === undefined) {
        return false
    }

    const equalityConditions = [
        a.id === b.id,
        a.name === b.name,
        a.type === b.type,
        a.pattern === b.pattern,
        a.retentionEnabled === b.retentionEnabled,
        a.retentionDurationHours === b.retentionDurationHours,
        a.retainIntermediateCommits === b.retainIntermediateCommits,
        a.indexingEnabled === b.indexingEnabled,
        a.indexCommitMaxAgeHours === b.indexCommitMaxAgeHours,
        a.indexIntermediateCommits === b.indexIntermediateCommits,
        comparePatterns(a.repositoryPatterns, b.repositoryPatterns),
    ]

    return equalityConditions.every(isEqual => isEqual)
}

function comparePatterns(a: string[] | null, b: string[] | null): boolean {
    if (a === null && b === null) {
        // Neither supplied
        return true
    }

    if (!a || !b) {
        // Only one supplied
        return false
    }

    // Both supplied and their contents match
    return a.length === b.length && a.every((pattern, index) => b[index] === pattern)
}

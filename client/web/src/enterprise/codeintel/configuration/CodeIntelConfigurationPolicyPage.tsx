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

                {indexingEnabled && <IndexingSettings policy={policy} setPolicy={setPolicy} />}
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
    const invalid =
        false ||
        // Required values
        policy.name === '' ||
        (policy.pattern === '' && policy.type !== GitObjectType.GIT_COMMIT) ||
        // Required select values
        ![GitObjectType.GIT_COMMIT, GitObjectType.GIT_TAG, GitObjectType.GIT_TREE].includes(policy.type) ||
        // Numeric validation (optional)
        (policy.retentionDurationHours && policy.retentionDurationHours <= 0) ||
        (policy.indexCommitMaxAgeHours && policy.indexCommitMaxAgeHours <= 0)

    return !invalid
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

import * as H from 'history'
import React, { FunctionComponent, useCallback, useEffect, useState } from 'react'
import { RouteComponentProps } from 'react-router'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { ErrorAlert } from '@sourcegraph/web/src/components/alerts'
import { PageTitle } from '@sourcegraph/web/src/components/PageTitle'
import { Button, Container, LoadingSpinner, PageHeader } from '@sourcegraph/wildcard'

import { CodeIntelligenceConfigurationPolicyFields } from '../../../graphql-operations'

import {
    repoName as defaultRepoName,
    searchGitBranches as defaultSearchGitBranches,
    searchGitTags as defaultSearchGitTags,
} from './backend'
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
    repoName?: typeof defaultRepoName
    searchGitBranches?: typeof defaultSearchGitBranches
    searchGitTags?: typeof defaultSearchGitTags
    history: H.History
}

export const CodeIntelConfigurationPolicyPage: FunctionComponent<CodeIntelConfigurationPolicyPageProps> = ({
    match: {
        params: { id },
    },
    repo,
    indexingEnabled = window.context?.codeIntelAutoIndexingEnabled,
    repoName = defaultRepoName,
    searchGitBranches = defaultSearchGitBranches,
    searchGitTags = defaultSearchGitTags,
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
        return savePolicyConfiguration({ variables }).then(() => history.push('./'))
    }, [policy, repo, savePolicyConfiguration, history])

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
            {loadingPolicyConfig ? (
                <LoadingSpinner className="icon-inline" />
            ) : (
                <>
                    {policy.protected && (
                        <div className="alert alert-info">
                            This configuration policy is protected. Protected configuration policies may not be deleted
                            and only the retention duration and indexing options are editable.
                        </div>
                    )}

                    <Container className="container form">
                        {savingError && <ErrorAlert prefix="Error saving configuration policy" error={savingError} />}
                        <BranchTargetSettings
                            repoId={repo?.id}
                            policy={policy}
                            setPolicy={setPolicy}
                            repoName={repoName}
                            searchGitBranches={searchGitBranches}
                            searchGitTags={searchGitTags}
                            disabled={policy.protected}
                        />
                    </Container>

                    <RetentionSettings policy={policy} setPolicy={setPolicy} />
                    {indexingEnabled && <IndexingSettings policy={policy} setPolicy={setPolicy} />}

                    <Container className="mt-2">
                        <Button
                            type="submit"
                            variant="primary"
                            onClick={savePolicyConfig}
                            disabled={isSaving || comparePolicies(policy, saved)}
                        >
                            {policy.id === '' ? 'Create' : 'Update'} policy
                        </Button>

                        <Button
                            type="button"
                            className="ml-3"
                            variant="secondary"
                            onClick={() => history.push('./')}
                            disabled={isSaving}
                        >
                            Cancel
                        </Button>

                        {isSaving && (
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

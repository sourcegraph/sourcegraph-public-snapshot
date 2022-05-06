import { FunctionComponent, useCallback, useEffect, useState } from 'react'

import { ApolloError } from '@apollo/client'
import * as H from 'history'
import DeleteIcon from 'mdi-react/DeleteIcon'
import { RouteComponentProps } from 'react-router'

import { ErrorAlert } from '@sourcegraph/branded/src/components/alerts'
import { GitObjectType } from '@sourcegraph/shared/src/graphql-operations'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { Button, Container, LoadingSpinner, PageHeader, Alert, Icon } from '@sourcegraph/wildcard'

import { PageTitle } from '../../../../components/PageTitle'
import { CodeIntelligenceConfigurationPolicyFields } from '../../../../graphql-operations'
import { BranchTargetSettings } from '../components/BranchTargetSettings'
import { FlashMessage } from '../components/FlashMessage'
import { IndexingSettings } from '../components/IndexSettings'
import { RetentionSettings } from '../components/RetentionSettings'
import { useDeletePolicies } from '../hooks/useDeletePolicies'
import { usePolicyConfigurationByID } from '../hooks/usePolicyConfigurationById'
import { useSavePolicyConfiguration } from '../hooks/useSavePolicyConfiguration'

export interface CodeIntelConfigurationPolicyPageProps
    extends RouteComponentProps<{ id: string }>,
        ThemeProps,
        TelemetryProps {
    repo?: { id: string }
    indexingEnabled?: boolean
    history: H.History
}

export const CodeIntelConfigurationPolicyPage: FunctionComponent<
    React.PropsWithChildren<CodeIntelConfigurationPolicyPageProps>
> = ({
    match: {
        params: { id },
    },
    repo,
    indexingEnabled = window.context?.codeIntelAutoIndexingEnabled,
    history,
    telemetryService,
}) => {
    useEffect(() => telemetryService.logViewEvent('CodeIntelConfigurationPolicy'), [telemetryService])

    const { policyConfig, loadingPolicyConfig, policyConfigError } = usePolicyConfigurationByID(id)
    const [saved, setSaved] = useState<CodeIntelligenceConfigurationPolicyFields>()
    const [policy, setPolicy] = useState<CodeIntelligenceConfigurationPolicyFields | undefined>()

    const { savePolicyConfiguration, isSaving, savingError } = useSavePolicyConfiguration(policy?.id === '')
    const { handleDeleteConfig, isDeleting, deleteError } = useDeletePolicies()

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
                    state: {
                        modal: 'ERROR',
                        message: `There was an error while saving policy: ${policy.name}. See error: ${error.message}`,
                    },
                })
            )
    }, [policy, repo, savePolicyConfiguration, history])

    const handleDelete = useCallback(
        async (id: string, name: string) => {
            if (!policy || !window.confirm(`Delete policy ${name}?`)) {
                return
            }

            return handleDeleteConfig({
                variables: { id },
                update: cache => cache.modify({ fields: { node: () => {} } }),
            }).then(() =>
                history.push({
                    pathname: './',
                    state: { modal: 'SUCCESS', message: `Configuration policy ${name} has been deleted.` },
                })
            )
        },
        [policy, handleDeleteConfig, history]
    )

    if (loadingPolicyConfig) {
        return <LoadingSpinner />
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

            {savingError && <ErrorAlert prefix="Error saving configuration policy" error={savingError} />}
            {deleteError && <ErrorAlert prefix="Error deleting configuration policy" error={deleteError} />}

            {history.location.state && (
                <FlashMessage state={history.location.state.modal} message={history.location.state.message} />
            )}

            {policy.protected ? (
                <Alert variant="info">
                    This configuration policy is protected. Protected configuration policies may not be deleted and only
                    the retention duration and indexing options are editable.
                </Alert>
            ) : (
                policy.id !== '' && (
                    <Container className="mb-3">
                        <Button
                            type="button"
                            variant="danger"
                            disabled={isSaving || isDeleting}
                            onClick={() => handleDelete(policy.id, policy.name)}
                            data-tooltip={`Deleting this policy may immediate affect data retention${
                                indexingEnabled ? ' and auto-indexing' : ''
                            }.`}
                        >
                            {!isDeleting && (
                                <>
                                    <Icon as={DeleteIcon} /> Delete policy
                                </>
                            )}
                            {isDeleting && (
                                <>
                                    <LoadingSpinner /> Deleting...
                                </>
                            )}
                        </Button>
                    </Container>
                )
            )}

            <Container className="container form mb-3">
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
                    disabled={isSaving || isDeleting || !validatePolicy(policy) || comparePolicies(policy, saved)}
                >
                    {!isSaving && <>{policy.id === '' ? 'Create' : 'Update'} policy</>}
                    {isSaving && (
                        <>
                            <LoadingSpinner /> Saving...
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

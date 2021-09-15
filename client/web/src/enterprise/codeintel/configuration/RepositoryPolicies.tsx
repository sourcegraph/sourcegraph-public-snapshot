import * as H from 'history'
import React, { FunctionComponent, useCallback } from 'react'

import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { ErrorAlert } from '@sourcegraph/web/src/components/alerts'
import { Container } from '@sourcegraph/wildcard'

import { PoliciesList } from './PoliciesList'
import { PolicyListActions } from './PolicyListActions'
import { ConfigTypes, CONFIG_TEXT } from './RepositoryPolicies.config'
import { usePoliciesConfig, useDeletePolicies, updateDeletePolicyCache } from './usePoliciesConfigurations'

interface RepositoryPoliciesProps {
    repo?: { id: string }
    indexingEnabled: boolean
    history: H.History
    isGlobal: boolean
}

export const RepositoryPolicies: FunctionComponent<RepositoryPoliciesProps> = ({
    repo = { id: null },
    indexingEnabled,
    history,
    isGlobal,
}) => {
    const { policies, loadingPolicies, policiesError } = usePoliciesConfig(isGlobal ? null : repo.id)
    const { handleDeleteConfig, isDeleting, deleteError } = useDeletePolicies()
    const configType = isGlobal ? ConfigTypes.Global : ConfigTypes.Local
    const policyActions =
        !isGlobal || repo.id === null ? (
            <PolicyListActions disabled={loadingPolicies} deleting={isDeleting} history={history} />
        ) : undefined

    const handleDelete = useCallback(
        async (id: string, name: string) => {
            if (!policies || !window.confirm(`${CONFIG_TEXT[configType].deleteConfirm} ${name}?`)) {
                return
            }

            return handleDeleteConfig({
                variables: { id },
                update: cache => updateDeletePolicyCache(cache, id),
            })
        },
        [policies, handleDeleteConfig, configType]
    )

    if (policiesError) {
        return <ErrorAlert prefix="Error fetching configuration" error={policiesError} />
    }

    return (
        <Container>
            <h3>{CONFIG_TEXT[configType].title}</h3>

            {deleteError && <ErrorAlert prefix="Error deleting configuration policy" error={deleteError} />}

            {loadingPolicies ? (
                <LoadingSpinner className="icon-inline" />
            ) : (
                <PoliciesList
                    policies={policies}
                    onDeletePolicy={repo.id === null || !isGlobal ? handleDelete : undefined}
                    disabled={loadingPolicies}
                    indexingEnabled={indexingEnabled}
                    buttonFragment={policyActions}
                    history={history}
                />
            )}
        </Container>
    )
}

import * as H from 'history'
import React, { FunctionComponent, useCallback, useEffect } from 'react'

import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { ErrorAlert } from '@sourcegraph/web/src/components/alerts'
import { Container } from '@sourcegraph/wildcard'

import { AuthenticatedUser } from '../../../auth'

import { PoliciesList } from './PoliciesList'
import { ConfigTypes, CONFIG_TEXT } from './RepositoryPolicies.config'
import { usePoliciesConfig, useDeletePolicies, updateDeletePolicyCache } from './usePoliciesConfigurations'

interface RepositoryPoliciesProps {
    repo?: { id: string }
    indexingEnabled: boolean
    history: H.History
    isGlobal: boolean
    authenticatedUser: AuthenticatedUser | null
    onHandleDisplayAction: React.Dispatch<React.SetStateAction<boolean>>
    onHandleIsDeleting: React.Dispatch<React.SetStateAction<boolean>>
    onHandleIsLoading: React.Dispatch<React.SetStateAction<boolean>>
}

export const RepositoryPolicies: FunctionComponent<RepositoryPoliciesProps> = ({
    repo = { id: null },
    indexingEnabled,
    history,
    isGlobal,
    authenticatedUser,
    onHandleDisplayAction,
    onHandleIsDeleting,
    onHandleIsLoading,
}) => {
    const { policies, loadingPolicies, policiesError } = usePoliciesConfig(isGlobal ? null : repo.id)
    const { handleDeleteConfig, isDeleting, deleteError } = useDeletePolicies()
    const configType = isGlobal ? ConfigTypes.Global : ConfigTypes.Local

    const handleDelete = useCallback(
        async (id: string, name: string) => {
            if (!policies || !window.confirm(`${CONFIG_TEXT[configType].deleteConfirm} ${name}?`)) {
                return
            }

            return handleDeleteConfig({
                variables: { id },
                update: cache => updateDeletePolicyCache(cache, id),
            }).then(() => {
                onHandleIsDeleting(false)
                history.push({
                    state: { modal: 'SUCCESS', message: `Configuration for policy ${name} has been deleted.` },
                })
            })
        },
        [policies, handleDeleteConfig, configType, onHandleIsDeleting, history]
    )

    useEffect(() => {
        if (!isGlobal || repo.id === null) {
            onHandleDisplayAction(true)
        }
        onHandleIsDeleting(isDeleting)
        onHandleIsLoading(loadingPolicies)
    }, [onHandleDisplayAction, isGlobal, repo.id, onHandleIsDeleting, isDeleting, onHandleIsLoading, loadingPolicies])

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
                    onDeletePolicy={
                        (repo.id === null || !isGlobal) && authenticatedUser?.siteAdmin ? handleDelete : undefined
                    }
                    disabled={loadingPolicies}
                    indexingEnabled={indexingEnabled}
                    history={history}
                />
            )}
        </Container>
    )
}

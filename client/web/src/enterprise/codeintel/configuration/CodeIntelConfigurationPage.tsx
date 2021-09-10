import * as H from 'history'
import React, { FunctionComponent, useCallback, useEffect, useState } from 'react'
import { RouteComponentProps } from 'react-router'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { ErrorAlert } from '@sourcegraph/web/src/components/alerts'
import { PageTitle } from '@sourcegraph/web/src/components/PageTitle'
import { PageHeader } from '@sourcegraph/wildcard'

import { CodeIntelligenceConfigurationPolicyFields } from '../../../graphql-operations'

import {
    deletePolicyById as defaultDeletePolicyById,
    getConfigurationForRepository as defaultGetConfigurationForRepository,
    getInferredConfigurationForRepository as defaultGetInferredConfigurationForRepository,
    getPolicies as defaultGetPolicies,
    updateConfigurationForRepository as defaultUpdateConfigurationForRepository,
} from './backend'
import { GlobalPolicies } from './GlobalPolicies'
import { RepositoryConfiguration } from './RepositoryConfiguration'

enum State {
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

    const deleteGlobalPolicy = useCallback(
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
                description={`Rules that define configuration for precise code intelligence ${
                    repo ? 'in this repository' : 'over all repositories'
                }.`}
                className="mb-3"
            />

            {repo ? (
                <RepositoryConfiguration
                    repo={repo}
                    disabled={state !== State.Idle}
                    deleting={state === State.Deleting}
                    policies={policies}
                    deletePolicy={deletePolicy}
                    globalPolicies={globalPolicies}
                    deleteGlobalPolicy={deleteGlobalPolicy}
                    deleteError={deleteError}
                    updateConfigurationForRepository={updateConfigurationForRepository}
                    getConfigurationForRepository={getConfigurationForRepository}
                    getInferredConfigurationForRepository={getInferredConfigurationForRepository}
                    indexingEnabled={indexingEnabled}
                    isLightTheme={isLightTheme}
                    telemetryService={telemetryService}
                    history={history}
                />
            ) : (
                <GlobalPolicies
                    repo={repo}
                    disabled={state !== State.Idle}
                    deleting={state === State.Deleting}
                    globalPolicies={globalPolicies}
                    deleteGlobalPolicy={deleteGlobalPolicy}
                    deleteError={deleteError}
                    indexingEnabled={indexingEnabled}
                    history={history}
                />
            )}
        </>
    )
}

import * as H from 'history'
import React, { FunctionComponent, useCallback, useEffect, useState } from 'react'
import { RouteComponentProps } from 'react-router'

import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { ErrorAlert } from '@sourcegraph/web/src/components/alerts'
import { PageTitle } from '@sourcegraph/web/src/components/PageTitle'
import { Button, Container, PageHeader } from '@sourcegraph/wildcard'

import { CodeIntelligenceConfigurationPolicyFields } from '../../../graphql-operations'

import {
    deletePolicyById as defaultDeletePolicyById,
    getConfigurationForRepository as defaultGetConfigurationForRepository,
    getInferredConfigurationForRepository as defaultGetInferredConfigurationForRepository,
    getPolicies as defaultGetPolicies,
    updateConfigurationForRepository as defaultUpdateConfigurationForRepository,
} from './backend'
import { CodeIntelligenceConfigurationTabHeader, SelectedTab } from './CodeIntelligenceConfigurationTabHeader'
import { ConfigurationEditor } from './ConfigurationEditor'
import { PoliciesList } from './PoliciesList'

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

    /** For testing only. */
    openTab?: SelectedTab
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
    openTab,
}) => {
    useEffect(() => telemetryService.logViewEvent('CodeIntelConfigurationPage'), [telemetryService])

    const [selectedTab, setSelectedTab] = useState<SelectedTab>(
        openTab ?? repo ? 'repositoryPolicies' : 'globalPolicies'
    )
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
                className="mt-2"
                variant="primary"
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
                description={`Rules that define configuration for precise code intelligence ${
                    repo ? 'in this repository' : 'over all repositories'
                }.`}
                className="mb-3"
            />

            {repo && (
                <CodeIntelligenceConfigurationTabHeader
                    selectedTab={selectedTab}
                    setSelectedTab={setSelectedTab}
                    indexingEnabled={indexingEnabled}
                />
            )}

            <Container>
                {selectedTab === 'globalPolicies' ? (
                    <>
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
                    </>
                ) : selectedTab === 'repositoryPolicies' ? (
                    <>
                        <h3>Repository-specific policies</h3>

                        {deleteError && <ErrorAlert prefix="Error deleting configuration policy" error={deleteError} />}

                        <PoliciesList
                            policies={policies}
                            deletePolicy={deletePolicy}
                            disabled={state !== State.Idle}
                            indexingEnabled={indexingEnabled}
                            buttonFragment={policyListButtonFragment}
                            history={history}
                        />
                    </>
                ) : (
                    selectedTab === 'indexConfiguration' &&
                    repo &&
                    indexingEnabled && (
                        <>
                            <h3>Auto-indexing configuration</h3>

                            <ConfigurationEditor
                                repoId={repo.id}
                                updateConfigurationForRepository={updateConfigurationForRepository}
                                getConfigurationForRepository={getConfigurationForRepository}
                                getInferredConfigurationForRepository={getInferredConfigurationForRepository}
                                isLightTheme={isLightTheme}
                                telemetryService={telemetryService}
                                history={history}
                            />
                        </>
                    )
                )}
            </Container>
        </>
    )
}

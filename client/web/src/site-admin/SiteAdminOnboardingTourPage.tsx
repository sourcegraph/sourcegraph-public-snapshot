import { type FC, type PropsWithChildren, useState, useEffect } from 'react'

import {
    OnboardingTourConfigMutationResult,
    OnboardingTourConfigMutationVariables,
    OnboardingTourConfigResult,
    OnboardingTourConfigVariables,
} from 'src/graphql-operations'

import { gql, useMutation, useQuery } from '@sourcegraph/http-client'
import type { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { useIsLightTheme } from '@sourcegraph/shared/src/theme'
import { PageHeader, Text, Container, BeforeUnloadPrompt, LoadingSpinner, Alert } from '@sourcegraph/wildcard'

import onboardingSchemaJSON from '../../../../schema/onboardingtour.schema.json'
import { PageTitle } from '../components/PageTitle'
import { SaveToolbar } from '../components/SaveToolbar'
import { MonacoSettingsEditor } from '../settings/MonacoSettingsEditor'

interface Props extends TelemetryProps {}

const ONBOARDING_TOUR_QUERY = gql`
    query OnboardingTourConfig {
        onboardingTourContent {
            current {
                id
                value
            }
        }
    }
`

const ONBOARDING_TOUR_MUTATION = gql`
    mutation OnboardingTourConfigMutation($json: String!) {
        updateOnboardingTourContent(input: $json) {
            alwaysNil
        }
    }
`

const DEFAULT_VALUE = JSON.stringify(
    {
        tasks: [],
    },
    null,
    4
)

export const SiteAdminOnboardingTourPage: FC<PropsWithChildren<Props>> = () => {
    const isLightTheme = useIsLightTheme()
    const [value, setValue] = useState<string | null>(null)
    const { data, loading, error, previousData } = useQuery<OnboardingTourConfigResult, OnboardingTourConfigVariables>(
        ONBOARDING_TOUR_QUERY,
        {}
    )
    const existingConfiguration = data?.onboardingTourContent.current?.value
    const initialLoad = loading && !previousData
    const dirty = !loading && value !== null && existingConfiguration !== value
    const config = loading
        ? value ?? ''
        : value !== null
        ? value
        : existingConfiguration
        ? existingConfiguration
        : DEFAULT_VALUE

    const discard = (): void => {
        if (dirty && window.confirm('Discard onboarding tour edits?')) {
            setValue(null)
        }
    }

    // Placeholder values
    const [updateOnboardinTourConfig, { loading: saving, error: mutationError }] = useMutation<
        OnboardingTourConfigMutationResult,
        OnboardingTourConfigMutationVariables
    >(ONBOARDING_TOUR_MUTATION, { refetchQueries: ['OnboardingTourConfig'] })

    async function save() {
        if (value !== null) {
            updateOnboardinTourConfig({ variables: { json: value } })
        }
    }
    // End placeholder values

    return (
        <>
            <PageTitle title="Onboarding tour" />
            <PageHeader className="mb-3">
                <PageHeader.Heading as="h3" styleAs="h2">
                    <PageHeader.Breadcrumb>Onboarding tour</PageHeader.Breadcrumb>
                </PageHeader.Heading>
            </PageHeader>
            <Text>Configure the onboarding tour steps</Text>
            <Container>
                {initialLoad && <LoadingSpinner title="Loading onboarding configuration" />}
                {error && <Alert>{error.message}</Alert>}
                {mutationError && <Alert>{mutationError.message}</Alert>}
                {!initialLoad && (
                    <>
                        <BeforeUnloadPrompt when={saving || dirty} message="Discard settings changes?" />
                        <MonacoSettingsEditor
                            isLightTheme={isLightTheme}
                            language="json"
                            jsonSchema={onboardingSchemaJSON}
                            value={config}
                            onChange={setValue}
                        />
                        <SaveToolbar dirty={dirty} error={error} saving={saving} onSave={save} onDiscard={discard} />
                    </>
                )}
            </Container>
        </>
    )
}

import { type FC, type PropsWithChildren, useState } from 'react'

import type {
    OnboardingTourConfigMutationResult,
    OnboardingTourConfigMutationVariables,
    OnboardingTourConfigResult,
    OnboardingTourConfigVariables,
} from 'src/graphql-operations'

import { gql, useMutation, useQuery } from '@sourcegraph/http-client'
import type { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { useIsLightTheme } from '@sourcegraph/shared/src/theme'
import { PageHeader, Text, Container, BeforeUnloadPrompt, LoadingSpinner, H3 } from '@sourcegraph/wildcard'

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
    const config = loading ? value ?? '' : value !== null ? value : existingConfiguration || DEFAULT_VALUE

    const discard = (): void => {
        if (dirty && window.confirm('Discard onboarding tour changes?')) {
            setValue(null)
        }
    }

    const [updateOnboardinTourConfig, { loading: saving, error: mutationError }] = useMutation<
        OnboardingTourConfigMutationResult,
        OnboardingTourConfigMutationVariables
    >(ONBOARDING_TOUR_MUTATION, { refetchQueries: ['OnboardingTourConfig'] })

    function save(): void {
        if (value !== null) {
            // Conflicts with @typescript-eslint/no-floating-promises
            // eslint-disable-next-line no-void
            void updateOnboardinTourConfig({ variables: { json: value } })
        }
    }

    return (
        <>
            <PageTitle title="End user onboarding" />
            <PageHeader className="mb-3">
                <PageHeader.Heading as="h3" styleAs="h2">
                    <PageHeader.Breadcrumb>End user onboarding</PageHeader.Breadcrumb>
                </PageHeader.Heading>
            </PageHeader>
            <Text>This settings controls the onboarding task list that is displayed to all users by default.</Text>
            <Container>
                {initialLoad && <LoadingSpinner title="Loading onboarding configuration" />}
                {!initialLoad && (
                    <>
                        <BeforeUnloadPrompt when={saving || dirty} message="Discard settings changes?" />
                        <MonacoSettingsEditor
                            isLightTheme={isLightTheme}
                            language="json"
                            jsonSchema={onboardingSchemaJSON}
                            value={config}
                            onChange={setValue}
                            height={450}
                        />
                        <SaveToolbar
                            dirty={dirty}
                            error={error || mutationError}
                            saving={saving}
                            onSave={save}
                            onDiscard={discard}
                        />
                    </>
                )}
            </Container>
            <H3 as="h4" className="mt-3">
                Most common parameters reference
            </H3>
            <img
                src="https://storage.googleapis.com/sourcegraph-assets/onboarding/onboarding-config-reference.svg"
                alt="onboarding tour reference"
                className="percy-hide w-100"
            />
        </>
    )
}

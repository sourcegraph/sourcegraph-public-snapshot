import { type FC, type PropsWithChildren, useState, useEffect } from 'react'

import type { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { useIsLightTheme } from '@sourcegraph/shared/src/theme'
import { PageHeader, Text, Container, BeforeUnloadPrompt, LoadingSpinner } from '@sourcegraph/wildcard'

import onboardingSchemaJSON from '../../../../schema/onboardingtour.schema.json'
import { PageTitle } from '../components/PageTitle'
import { SaveToolbar } from '../components/SaveToolbar'
import { MonacoSettingsEditor } from '../settings/MonacoSettingsEditor'

interface Props extends TelemetryProps {}

function useLoadOnboardingConfig(): { data: string; loading: boolean; error?: Error } {
    const data = JSON.stringify(
        {
            tasks: [
                {
                    title: 'Code search use cases',
                    steps: [
                        {
                            id: 'SymbolsSearch',
                            label: 'Search multiple repos',
                            action: {
                                type: 'link',
                                value: {
                                    C: '/search?q=context:global+repo:torvalds/.*+lang:c+-file:.*/testing+magic&patternType=literal',
                                },
                            },
                            info: 'some info',
                        },
                        {
                            id: 'InstallOrSignUp',
                            label: 'Get free trial',
                            action: {
                                type: 'new-tab-link',
                                value: 'https://about.sourcegraph.com',
                            },
                            // This is done to mimic user creating an account, and signed in there is a different tour
                            completeAfterEvents: ['non-existing-event'],
                        },
                    ],
                },
            ],
        },
        undefined,
        4
    )

    const [loading, setLoading] = useState(true)

    useEffect(() => {
        const timer = window.setTimeout(() => {
            setLoading(false)
        }, 2000)
        return () => window.clearTimeout(timer)
    }, [])

    return { data, loading }
}

export const SiteAdminOnboardingTourPage: FC<PropsWithChildren<Props>> = () => {
    const isLightTheme = useIsLightTheme()
    const [value, setValue] = useState<string | null>(null)
    const { data, loading, error } = useLoadOnboardingConfig()
    const dirty = !loading && value !== null && data !== value
    const config = loading ? '' : value === null ? data : value

    const discard = (): void => {
        // TODO: Prompt user whether they really want to discard
        setValue(null)
    }

    // Placeholder values
    const saving = false
    const save = (): void => {}
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
                {loading && <LoadingSpinner title="Loading onboarding configuration" />}
                {!loading && error && 'Error'}
                {!loading && (
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

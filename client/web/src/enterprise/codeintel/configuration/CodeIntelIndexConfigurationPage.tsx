import * as H from 'history'
import React, { FunctionComponent, useCallback, useEffect, useState } from 'react'
import { RouteComponentProps } from 'react-router'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { Container, PageHeader } from '@sourcegraph/wildcard'

import { ErrorAlert } from '../../../components/alerts'
import { CodeIntelAutoIndexSaveToolbar, AutoIndexProps } from '../../../components/CodeIntelAutoIndexSaveToolbar'
import { PageTitle } from '../../../components/PageTitle'
import { SaveToolbarPropsGenerator, SaveToolbarProps } from '../../../components/SaveToolbar'
import { DynamicallyImportedMonacoSettingsEditor } from '../../../settings/DynamicallyImportedMonacoSettingsEditor'

import { getConfiguration as defaultGetConfiguration, updateConfiguration } from './backend'
import allConfigSchema from './schema.json'

export interface CodeIntelIndexConfigurationPageProps extends RouteComponentProps<{}>, ThemeProps, TelemetryProps {
    repo: { id: string }
    history: H.History
    getConfiguration?: typeof defaultGetConfiguration
}

enum State {
    Idle,
    Saving,
    Queueing,
}

export const CodeIntelIndexConfigurationPage: FunctionComponent<CodeIntelIndexConfigurationPageProps> = ({
    repo,
    isLightTheme,
    telemetryService,
    history,
    getConfiguration = defaultGetConfiguration,
}) => {
    useEffect(() => telemetryService.logViewEvent('CodeIntelIndexConfigurationPage'), [telemetryService])

    const [fetchError, setFetchError] = useState<Error>()
    const [saveError, setSaveError] = useState<Error>()
    const [state, setState] = useState(() => State.Idle)
    const [configuration, setConfiguration] = useState('')
    const [inferredConfiguration, setInferredConfiguration] = useState('')
    const [dirty, setDirty] = useState<boolean>()

    useEffect(() => {
        const subscription = getConfiguration({ id: repo.id }).subscribe(configuration => {
            setConfiguration(configuration?.indexConfiguration?.configuration || '')
            setInferredConfiguration(configuration?.indexConfiguration?.inferredConfiguration || '')
        }, setFetchError)

        return () => subscription.unsubscribe()
    }, [repo, getConfiguration])

    const save = useCallback(
        async (content: string) => {
            setState(State.Saving)
            setSaveError(undefined)

            try {
                await updateConfiguration({ id: repo.id, content }).toPromise()
                setConfiguration(content)
            } catch (error) {
                setSaveError(error)
            } finally {
                setState(State.Idle)
            }
        },
        [repo]
    )

    const infer = useCallback(() => {
        setConfiguration(inferredConfiguration)
    }, [inferredConfiguration])

    const onDirtyChange = useCallback((dirty: boolean) => {
        setDirty(dirty)
    }, [])

    const saving = state === State.Saving
    const queueing = state === State.Queueing

    const customToolbar: {
        propsGenerator: SaveToolbarPropsGenerator<AutoIndexProps>
        saveToolbar: React.FunctionComponent<SaveToolbarProps & AutoIndexProps>
    } = {
        propsGenerator: (props: Readonly<SaveToolbarProps> & Readonly<{}>): SaveToolbarProps & AutoIndexProps => {
            const autoIndexProps: AutoIndexProps = {
                onInfer: infer,
            }

            const mergedProps = { ...props, ...autoIndexProps }
            mergedProps.willShowError = (): boolean => !queueing && !mergedProps.saving
            mergedProps.saveDiscardDisabled = (): boolean => saving || !dirty || queueing
            return mergedProps
        },
        saveToolbar: CodeIntelAutoIndexSaveToolbar,
    }

    return fetchError ? (
        <ErrorAlert prefix="Error fetching index configuration" error={fetchError} />
    ) : (
        <div className="code-intel-index-configuration">
            <PageTitle title="Code intelligence index configuration" />

            <PageHeader
                headingElement="h2"
                path={[
                    {
                        text: <>Code intelligence index configuration</>,
                    },
                ]}
                description={
                    <>
                        Override the inferred configuration when automatically indexing repositories on{' '}
                        <a href="https://sourcegraph.com" target="_blank" rel="noreferrer noopener">
                            Sourcegraph.com
                        </a>
                        .
                    </>
                }
                className="mb-3"
            />

            <Container>
                {saveError && <ErrorAlert prefix="Error saving index configuration" error={saveError} />}

                <DynamicallyImportedMonacoSettingsEditor
                    value={configuration || ''}
                    jsonSchema={allConfigSchema}
                    canEdit={true}
                    onSave={save}
                    saving={state === State.Saving}
                    height={600}
                    isLightTheme={isLightTheme}
                    history={history}
                    telemetryService={telemetryService}
                    customSaveToolbar={customToolbar}
                    onDirtyChange={onDirtyChange}
                />
            </Container>
        </div>
    )
}

import * as H from 'history'
import React, { FunctionComponent, useCallback, useEffect, useMemo, useState } from 'react'
import { RouteComponentProps } from 'react-router'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { Container, PageHeader } from '@sourcegraph/wildcard'

import { ErrorAlert } from '../../../components/alerts'
import { PageTitle } from '../../../components/PageTitle'
import { SaveToolbarPropsGenerator, SaveToolbarProps } from '../../../components/SaveToolbar'
import { DynamicallyImportedMonacoSettingsEditor } from '../../../settings/DynamicallyImportedMonacoSettingsEditor'

import { getConfiguration as defaultGetConfiguration, updateConfiguration } from './backend'
import { CodeIntelAutoIndexSaveToolbar, AutoIndexProps } from './CodeIntelAutoIndexSaveToolbar'
import allConfigSchema from './schema.json'
import { editor } from 'monaco-editor'

export interface CodeIntelIndexConfigurationPageProps extends RouteComponentProps<{}>, ThemeProps, TelemetryProps {
    repo: { id: string }
    history: H.History
    getConfiguration?: typeof defaultGetConfiguration
}

enum State {
    Idle,
    Saving,
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
    const [editor, setEditor] = useState<editor.ICodeEditor>()

    useEffect(() => {
        const subscription = getConfiguration({ id: repo.id }).subscribe(config => {
            setConfiguration(config?.indexConfiguration?.configuration || '')
            setInferredConfiguration(config?.indexConfiguration?.inferredConfiguration || '')
        }, setFetchError)

        return () => subscription.unsubscribe()
    }, [repo, getConfiguration])

    const save = useCallback(
        async (content: string) => {
            setState(State.Saving)
            setSaveError(undefined)

            try {
                await updateConfiguration({ id: repo.id, content }).toPromise()
            } catch (error) {
                setSaveError(error)
            } finally {
                setState(State.Idle)
            }
        },
        [repo]
    )

    const onInfer = useCallback(() => editor?.setValue(inferredConfiguration), [editor, inferredConfiguration])

    const customToolbar: {
        propsGenerator: SaveToolbarPropsGenerator<AutoIndexProps>
        saveToolbar: React.FunctionComponent<SaveToolbarProps & AutoIndexProps>
    } = useMemo(
        () => ({
            propsGenerator: (props: Readonly<SaveToolbarProps> & Readonly<{}>): SaveToolbarProps & AutoIndexProps => {
                const mergedProps = {
                    ...props,
                    inferEnabled: inferredConfiguration !== '' && configuration !== inferredConfiguration,
                    onInfer,
                }
                mergedProps.willShowError = (): boolean => !mergedProps.saving
                mergedProps.saveDiscardDisabled = (): boolean => state === State.Saving || !dirty
                return mergedProps
            },
            saveToolbar: CodeIntelAutoIndexSaveToolbar,
        }),
        [editor, dirty, inferredConfiguration, onInfer, state]
    )

    return fetchError ? (
        <ErrorAlert prefix="Error fetching index configuration" error={fetchError} />
    ) : (
        <div className="code-intel-index-configuration">
            <PageTitle title="Auto-indexing configuration" />

            <PageHeader
                headingElement="h2"
                path={[
                    {
                        text: <>Auto-indexing configuration</>,
                    },
                ]}
                description="TODO"
                className="mb-3"
            />

            <Container>
                {saveError && <ErrorAlert prefix="Error saving index configuration" error={saveError} />}

                <DynamicallyImportedMonacoSettingsEditor
                    value={configuration}
                    jsonSchema={allConfigSchema}
                    canEdit={true}
                    onSave={save}
                    saving={state === State.Saving}
                    height={600}
                    isLightTheme={isLightTheme}
                    history={history}
                    telemetryService={telemetryService}
                    customSaveToolbar={customToolbar}
                    onDirtyChange={setDirty}
                    onEditor={setEditor}
                />
            </Container>
        </div>
    )
}

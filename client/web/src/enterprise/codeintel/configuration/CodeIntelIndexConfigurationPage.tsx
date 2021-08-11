import * as H from 'history'
import { editor } from 'monaco-editor'
import React, { FunctionComponent, useCallback, useEffect, useMemo, useState } from 'react'
import { RouteComponentProps } from 'react-router'

import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { Container, PageHeader } from '@sourcegraph/wildcard'

import { ErrorAlert } from '../../../components/alerts'
import { PageTitle } from '../../../components/PageTitle'
import { SaveToolbar, SaveToolbarProps, SaveToolbarPropsGenerator } from '../../../components/SaveToolbar'
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
}

export const CodeIntelIndexConfigurationPage: FunctionComponent<CodeIntelIndexConfigurationPageProps> = ({
    repo,
    isLightTheme,
    telemetryService,
    history,
    getConfiguration = defaultGetConfiguration,
}) => {
    useEffect(() => telemetryService.logViewEvent('CodeIntelIndexConfigurationPage'), [telemetryService])

    const [configuration, setConfiguration] = useState<string>()
    const [inferredConfiguration, setInferredConfiguration] = useState<string>()
    const [fetchError, setFetchError] = useState<Error>()

    useEffect(() => {
        const subscription = getConfiguration({ id: repo.id }).subscribe(config => {
            setConfiguration(config?.indexConfiguration?.configuration || '')
            setInferredConfiguration(config?.indexConfiguration?.inferredConfiguration || '')
        }, setFetchError)

        return () => subscription.unsubscribe()
    }, [repo, getConfiguration])

    const [saveError, setSaveError] = useState<Error>()
    const [state, setState] = useState(() => State.Idle)

    const save = useCallback(
        async (content: string) => {
            setState(State.Saving)
            setSaveError(undefined)

            try {
                await updateConfiguration({ id: repo.id, content }).toPromise()
                setDirty(false)
                setConfiguration(content)
            } catch (error) {
                setSaveError(error)
            } finally {
                setState(State.Idle)
            }
        },
        [repo]
    )

    const [dirty, setDirty] = useState<boolean>()
    const [editor, setEditor] = useState<editor.ICodeEditor>()
    const infer = useCallback(() => editor?.setValue(inferredConfiguration || ''), [editor, inferredConfiguration])

    const customToolbar = useMemo<{
        saveToolbar: React.FunctionComponent<SaveToolbarProps & AutoIndexProps>
        propsGenerator: SaveToolbarPropsGenerator<AutoIndexProps>
    }>(
        () => ({
            saveToolbar: CodeIntelAutoIndexSaveToolbar,
            propsGenerator: props => {
                const mergedProps = {
                    ...props,
                    onInfer: infer,
                    inferEnabled: !!inferredConfiguration && configuration !== inferredConfiguration,
                }
                mergedProps.willShowError = () => !mergedProps.saving
                mergedProps.saveDiscardDisabled = () => mergedProps.saving || !dirty

                return mergedProps
            },
        }),
        [dirty, configuration, inferredConfiguration, infer]
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
                className="mb-3"
            />

            <Container>
                {saveError && <ErrorAlert prefix="Error saving index configuration" error={saveError} />}

                {configuration === undefined ? (
                    <LoadingSpinner className="icon-inline" />
                ) : (
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
                )}
            </Container>
        </div>
    )
}

interface AutoIndexProps {
    inferEnabled: boolean
    onInfer?: () => void
}

const CodeIntelAutoIndexSaveToolbar: React.FunctionComponent<SaveToolbarProps & AutoIndexProps> = ({
    dirty,
    saving,
    error,
    onSave,
    onDiscard,
    inferEnabled,
    onInfer,
    saveDiscardDisabled,
}) => (
    <SaveToolbar
        dirty={dirty}
        saving={saving}
        onSave={onSave}
        error={error}
        saveDiscardDisabled={saveDiscardDisabled}
        onDiscard={onDiscard}
    >
        {inferEnabled && (
            <button
                type="button"
                title="Infer index configuration from HEAD"
                className="btn btn-link"
                onClick={onInfer}
            >
                Infer index configuration from HEAD
            </button>
        )}
    </SaveToolbar>
)

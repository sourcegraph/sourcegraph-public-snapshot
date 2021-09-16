import * as H from 'history'
import { editor } from 'monaco-editor'
import React, { FunctionComponent, useCallback, useEffect, useMemo, useState } from 'react'

import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { ErrorAlert } from '@sourcegraph/web/src/components/alerts'

import { SaveToolbarProps, SaveToolbarPropsGenerator } from '../../../components/SaveToolbar'
import { DynamicallyImportedMonacoSettingsEditor } from '../../../settings/DynamicallyImportedMonacoSettingsEditor'

import {
    getConfigurationForRepository as defaultGetConfigurationForRepository,
    getInferredConfigurationForRepository as defaultGetInferredConfigurationForRepository,
    updateConfigurationForRepository as defaultUpdateConfigurationForRepository,
} from './backend'
import { IndexConfigurationSaveToolbar, IndexConfigurationSaveToolbarProps } from './IndexConfigurationSaveToolbar'
import allConfigSchema from './schema.json'

export interface ConfigurationEditorProps extends ThemeProps, TelemetryProps {
    repoId: string
    history: H.History
    getConfigurationForRepository: typeof defaultGetConfigurationForRepository
    getInferredConfigurationForRepository: typeof defaultGetInferredConfigurationForRepository
    updateConfigurationForRepository: typeof defaultUpdateConfigurationForRepository
}

enum EditorState {
    Idle,
    Saving,
}

export const ConfigurationEditor: FunctionComponent<ConfigurationEditorProps> = ({
    repoId,
    isLightTheme,
    telemetryService,
    history,
    getConfigurationForRepository,
    getInferredConfigurationForRepository,
    updateConfigurationForRepository,
}) => {
    const [configuration, setConfiguration] = useState<string>()
    const [inferredConfiguration, setInferredConfiguration] = useState<string>()
    const [fetchError, setFetchError] = useState<Error>()

    useEffect(() => {
        const subscription = getConfigurationForRepository(repoId).subscribe(config => {
            setConfiguration(config?.indexConfiguration?.configuration || '')
        }, setFetchError)

        return () => subscription.unsubscribe()
    }, [repoId, getConfigurationForRepository])

    useEffect(() => {
        const subscription = getInferredConfigurationForRepository(repoId).subscribe(config => {
            setInferredConfiguration(config?.indexConfiguration?.inferredConfiguration || '')
        }, setFetchError)

        return () => subscription.unsubscribe()
    }, [repoId, getInferredConfigurationForRepository])

    const [saveError, setSaveError] = useState<Error>()
    const [state, setState] = useState(() => EditorState.Idle)

    const save = useCallback(
        async (content: string) => {
            setState(EditorState.Saving)
            setSaveError(undefined)

            try {
                await updateConfigurationForRepository(repoId, content).toPromise()
                setDirty(false)
                setConfiguration(content)
            } catch (error) {
                setSaveError(error)
            } finally {
                setState(EditorState.Idle)
            }
        },
        [repoId, updateConfigurationForRepository]
    )

    const [dirty, setDirty] = useState<boolean>()
    const [editor, setEditor] = useState<editor.ICodeEditor>()
    const infer = useCallback(() => editor?.setValue(inferredConfiguration || ''), [editor, inferredConfiguration])

    const customToolbar = useMemo<{
        saveToolbar: React.FunctionComponent<SaveToolbarProps & IndexConfigurationSaveToolbarProps>
        propsGenerator: SaveToolbarPropsGenerator<IndexConfigurationSaveToolbarProps>
    }>(
        () => ({
            saveToolbar: IndexConfigurationSaveToolbar,
            propsGenerator: props => {
                const mergedProps = {
                    ...props,
                    onInfer: infer,
                    loading: inferredConfiguration === undefined,
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
        <>
            {saveError && <ErrorAlert prefix="Error saving index configuration" error={saveError} />}

            {configuration === undefined ? (
                <LoadingSpinner className="icon-inline" />
            ) : (
                <DynamicallyImportedMonacoSettingsEditor
                    value={configuration}
                    jsonSchema={allConfigSchema}
                    canEdit={true}
                    onSave={save}
                    saving={state === EditorState.Saving}
                    height={600}
                    isLightTheme={isLightTheme}
                    history={history}
                    telemetryService={telemetryService}
                    customSaveToolbar={customToolbar}
                    onDirtyChange={setDirty}
                    onEditor={setEditor}
                />
            )}
        </>
    )
}

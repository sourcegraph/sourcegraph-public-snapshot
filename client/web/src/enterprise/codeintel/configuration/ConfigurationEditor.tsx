import * as H from 'history'
import { editor } from 'monaco-editor'
import React, { FunctionComponent, useCallback, useMemo, useState } from 'react'

import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { ErrorAlert } from '@sourcegraph/web/src/components/alerts'

import { SaveToolbarProps, SaveToolbarPropsGenerator } from '../../../components/SaveToolbar'
import { DynamicallyImportedMonacoSettingsEditor } from '../../../settings/DynamicallyImportedMonacoSettingsEditor'

import { IndexConfigurationSaveToolbar, IndexConfigurationSaveToolbarProps } from './IndexConfigurationSaveToolbar'
import allConfigSchema from './schema.json'
import {
    useInferredConfig,
    useRepositoryConfig,
    useUpdateConfigurationForRepository,
} from './usePoliciesConfigurations'

export interface ConfigurationEditorProps extends ThemeProps, TelemetryProps {
    repoId: string
    history: H.History
}

export const ConfigurationEditor: FunctionComponent<ConfigurationEditorProps> = ({
    repoId,
    isLightTheme,
    telemetryService,
    history,
}) => {
    const { inferredConfiguration, loadingInferred, inferredError } = useInferredConfig(repoId)
    const { configuration, loadingRepository, repositoryError } = useRepositoryConfig(repoId)
    const { updateConfigForRepository, isUpdating, updatingError } = useUpdateConfigurationForRepository()

    const save = useCallback(
        async (content: string) =>
            updateConfigForRepository({
                variables: { id: repoId, content },
                update: cache => cache.modify({ fields: { node: () => {} } }),
            }).then(() => setDirty(false)),
        [updateConfigForRepository, repoId]
    )

    const [dirty, setDirty] = useState<boolean>()
    const [editor, setEditor] = useState<editor.ICodeEditor>()
    const infer = useCallback(() => editor?.setValue(inferredConfiguration), [editor, inferredConfiguration])

    const customToolbar = useMemo<{
        saveToolbar: FunctionComponent<SaveToolbarProps & IndexConfigurationSaveToolbarProps>
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

    if (inferredError || repositoryError) {
        return <ErrorAlert prefix="Error fetching index configuration" error={inferredError || repositoryError} />
    }

    return (
        <>
            {updatingError && <ErrorAlert prefix="Error saving index configuration" error={updatingError} />}

            {loadingInferred || loadingRepository ? (
                <LoadingSpinner className="icon-inline" />
            ) : (
                <DynamicallyImportedMonacoSettingsEditor
                    value={configuration}
                    jsonSchema={allConfigSchema}
                    canEdit={true}
                    onSave={save}
                    saving={isUpdating}
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

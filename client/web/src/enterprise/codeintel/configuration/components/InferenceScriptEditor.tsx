import { FunctionComponent, useCallback, useMemo, useState } from 'react'

import * as H from 'history'

import { ErrorAlert } from '@sourcegraph/branded/src/components/alerts'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { LoadingSpinner, screenReaderAnnounce } from '@sourcegraph/wildcard'

import { AuthenticatedUser } from '../../../../auth'
import { SaveToolbar, SaveToolbarProps, SaveToolbarPropsGenerator } from '../../../../components/SaveToolbar'
import { DynamicallyImportedMonacoSettingsEditor } from '../../../../settings/DynamicallyImportedMonacoSettingsEditor'
import { useInferenceScript } from '../hooks/useInferenceScript'
import { useUpdateInferenceScript } from '../hooks/useUpdateInferenceScript'

export interface InferenceScriptEditorProps extends ThemeProps, TelemetryProps {
    authenticatedUser: AuthenticatedUser | null
    history: H.History
}

export const InferenceScriptEditor: FunctionComponent<React.PropsWithChildren<InferenceScriptEditorProps>> = ({
    authenticatedUser,
    isLightTheme,
    telemetryService,
    history,
}) => {
    const { inferenceScript, loadingScript, fetchError } = useInferenceScript()
    const { updateInferenceScript, isUpdating, updatingError } = useUpdateInferenceScript()

    const save = useCallback(
        async (content: string) =>
            updateInferenceScript({
                variables: { script: content },
                update: cache => cache.modify({ fields: { node: () => {} } }),
            }).then(() => {
                screenReaderAnnounce('Saved successfully')
                setDirty(false)
            }),
        [updateInferenceScript]
    )

    const [dirty, setDirty] = useState<boolean>()
    // const [_editor, setEditor] = useState<editor.ICodeEditor>()

    const customToolbar = useMemo<{
        saveToolbar: FunctionComponent<SaveToolbarProps>
        propsGenerator: SaveToolbarPropsGenerator<{}>
    }>(
        () => ({
            saveToolbar: SaveToolbar,
            propsGenerator: props => {
                console.log(JSON.stringify(props))
                const mergedProps = {
                    ...props,
                    loading: isUpdating,
                }
                mergedProps.willShowError = () => !mergedProps.saving
                mergedProps.saveDiscardDisabled = () => mergedProps.saving || !dirty

                return mergedProps
            },
        }),
        [dirty, isUpdating]
    )

    if (fetchError) {
        return <ErrorAlert prefix="Error fetching inference script" error={fetchError} />
    }

    return (
        <>
            {updatingError && <ErrorAlert prefix="Error saving index configuration" error={updatingError} />}

            {loadingScript ? (
                <LoadingSpinner />
            ) : (
                <DynamicallyImportedMonacoSettingsEditor
                    value={inferenceScript}
                    language="lua"
                    // jsonSchema={allConfigSchema}
                    canEdit={authenticatedUser?.siteAdmin}
                    readOnly={!authenticatedUser?.siteAdmin}
                    onSave={save}
                    saving={isUpdating}
                    height={600}
                    isLightTheme={isLightTheme}
                    history={history}
                    telemetryService={telemetryService}
                    customSaveToolbar={authenticatedUser?.siteAdmin ? customToolbar : undefined}
                    onDirtyChange={setDirty}
                    // onEditor={setEditor}
                />
            )}
        </>
    )
}

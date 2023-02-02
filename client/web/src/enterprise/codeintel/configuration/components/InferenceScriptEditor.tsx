import { FunctionComponent, useCallback, useMemo, useState } from 'react'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { LoadingSpinner, PageHeader, screenReaderAnnounce, ErrorAlert } from '@sourcegraph/wildcard'

import { AuthenticatedUser } from '../../../../auth'
import { PageTitle } from '../../../../components/PageTitle'
import { SaveToolbar, SaveToolbarProps, SaveToolbarPropsGenerator } from '../../../../components/SaveToolbar'
import { DynamicallyImportedMonacoSettingsEditor } from '../../../../settings/DynamicallyImportedMonacoSettingsEditor'
import { ThemePreference, useTheme } from '../../../../theme'
import { INFERENCE_SCRIPT, useInferenceScript } from '../hooks/useInferenceScript'
import { useUpdateInferenceScript } from '../hooks/useUpdateInferenceScript'

import styles from './CodeIntelConfigurationPageHeader.module.scss'

export interface InferenceScriptEditorProps extends TelemetryProps {
    authenticatedUser: AuthenticatedUser | null
}

export const InferenceScriptEditor: FunctionComponent<InferenceScriptEditorProps> = ({
    authenticatedUser,
    telemetryService,
}) => {
    const { inferenceScript, loadingScript, fetchError } = useInferenceScript()
    const { updateInferenceScript, isUpdating, updatingError } = useUpdateInferenceScript()

    const save = useCallback(
        async (script: string) =>
            updateInferenceScript({
                variables: { script },
                refetchQueries: [INFERENCE_SCRIPT],
            }).then(() => {
                screenReaderAnnounce('Saved successfully')
                setDirty(false)
            }),
        [updateInferenceScript]
    )

    const [dirty, setDirty] = useState<boolean>()
    const isLightTheme = useTheme().enhancedThemePreference === ThemePreference.Light

    const customToolbar = useMemo<{
        saveToolbar: FunctionComponent<SaveToolbarProps>
        propsGenerator: SaveToolbarPropsGenerator<{}>
    }>(
        () => ({
            saveToolbar: SaveToolbar,
            propsGenerator: props => {
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

    const title = (
        <>
            <PageTitle title="Code graph inference script" />
            <div className={styles.grid}>
                <PageHeader
                    headingElement="h2"
                    path={[
                        {
                            text: <>Code graph inference script</>,
                        },
                    ]}
                    description={`Lua script that emits complete and/or partial auto-indexing
                job specifications. `}
                    className="mb-3"
                />
            </div>
        </>
    )

    if (fetchError) {
        return (
            <>
                {title}
                <ErrorAlert prefix="Error fetching inference script" error={fetchError} />
            </>
        )
    }

    return (
        <>
            {title}
            {updatingError && <ErrorAlert prefix="Error saving index configuration" error={updatingError} />}

            {loadingScript ? (
                <LoadingSpinner />
            ) : (
                <DynamicallyImportedMonacoSettingsEditor
                    value={inferenceScript}
                    language="lua"
                    canEdit={authenticatedUser?.siteAdmin}
                    readOnly={!authenticatedUser?.siteAdmin}
                    onSave={save}
                    saving={isUpdating}
                    height={600}
                    isLightTheme={isLightTheme}
                    telemetryService={telemetryService}
                    customSaveToolbar={authenticatedUser?.siteAdmin ? customToolbar : undefined}
                    onDirtyChange={setDirty}
                />
            )}
        </>
    )
}

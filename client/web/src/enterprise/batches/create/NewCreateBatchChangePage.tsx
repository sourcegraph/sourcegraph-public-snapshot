import classNames from 'classnames'
import React, { useCallback, useMemo, useState } from 'react'

import { BatchSpecWorkspaceResolutionState } from '@sourcegraph/shared/src/graphql/schema'
import {
    SettingsCascadeProps,
    SettingsOrgSubject,
    SettingsUserSubject,
} from '@sourcegraph/shared/src/settings/settings'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { ErrorAlert } from '@sourcegraph/web/src/components/alerts'
import { ButtonTooltip } from '@sourcegraph/web/src/components/ButtonTooltip'
import { PageHeader } from '@sourcegraph/wildcard'

import { BatchChangesIcon } from '../../../batches/icons'
import { FeedbackBadge } from '../../../components/FeedbackBadge'
import { Settings } from '../../../schema/settings.schema'
import { BatchSpecDownloadLink } from '../BatchSpec'

import { MonacoBatchSpecEditor } from './editor/MonacoBatchSpecEditor'
import helloWorldSample from './library/hello-world.batch.yaml'
import { LibraryPane } from './library/LibraryPane'
import { NamespaceSelector } from './NamespaceSelector'
import styles from './NewCreateBatchChangePage.module.scss'
import { useBatchSpecCode } from './useBatchSpecCode'
import { usePreviewBatchSpec } from './useBatchSpecPreview'
import { useExecuteBatchSpec } from './useExecuteBatchSpec'
import { useNamespaces } from './useNamespaces'
import { useBatchSpecWorkspaceResolution, WorkspacesPreview } from './workspaces-preview/WorkspacesPreview'

const getNamespaceDisplayName = (namespace: SettingsUserSubject | SettingsOrgSubject): string => {
    switch (namespace.__typename) {
        case 'User':
            return namespace.displayName ?? namespace.username
        case 'Org':
            return namespace.displayName ?? namespace.name
    }
}

/** TODO: This duplicates the URL field from the org/user resolvers on the backend, but we
 * don't have access to that from the settings cascade presently. Can we get it included
 * in the cascade instead somehow? */
const getNamespaceBatchChangesURL = (namespace: SettingsUserSubject | SettingsOrgSubject): string => {
    switch (namespace.__typename) {
        case 'User':
            return '/users/' + namespace.username + '/batch-changes'
        case 'Org':
            return '/organizations/' + namespace.name + '/batch-changes'
    }
}

interface CreateBatchChangePageProps extends ThemeProps, SettingsCascadeProps<Settings> {}

export const NewCreateBatchChangePage: React.FunctionComponent<CreateBatchChangePageProps> = ({
    isLightTheme,
    settingsCascade,
}) => {
    const { namespaces, defaultSelectedNamespace } = useNamespaces(settingsCascade)

    // The namespace selected for creating the new batch spec under.
    const [selectedNamespace, setSelectedNamespace] = useState<SettingsUserSubject | SettingsOrgSubject>(
        defaultSelectedNamespace
    )

    const [noCache, setNoCache] = useState<boolean>(false)

    const onChangeNoCache = useCallback<React.ChangeEventHandler<HTMLInputElement>>(
        event => {
            setNoCache(event.target.checked)
            // Mark that the batch spec code on the backend is now stale.
            setBatchSpecStale(true)
        },
        [setNoCache]
    )

    // Manage the batch spec input YAML code that's being edited.
    const { code, debouncedCode, isValid, handleCodeChange, excludeRepo, errors: codeErrors } = useBatchSpecCode(
        helloWorldSample
    )

    // Track whenever the batch spec code that is presently in the editor gets ahead of
    // the batch spec that was last submitted to the backend.
    const [batchSpecStale, setBatchSpecStale] = useState(false)
    const markUnstale = useCallback(() => setBatchSpecStale(false), [])

    // Manage the batch spec that was last submitted to the backend for the workspaces preview.
    const {
        previewBatchSpec,
        batchSpecID,
        currentPreviewRequestTime,
        isLoading: isLoadingPreview,
        error: previewError,
        clearError: clearPreviewError,
    } = usePreviewBatchSpec(selectedNamespace, noCache, markUnstale)

    const clearErrorsAndHandleCodeChange = useCallback(
        (newCode: string) => {
            clearPreviewError()
            // Mark that the batch spec code on the backend is now stale.
            setBatchSpecStale(true)
            handleCodeChange(newCode)
        },
        [handleCodeChange, clearPreviewError]
    )

    // Disable the preview button if the batch spec code is invalid or the on: statement
    // is missing, or if we're already processing a preview.
    const previewDisabled = useMemo(() => isValid !== true || isLoadingPreview, [isValid, isLoadingPreview])

    const { resolution: workspacesPreviewResolution } = useBatchSpecWorkspaceResolution(
        batchSpecID,
        currentPreviewRequestTime,
        {
            fetchPolicy: 'cache-first',
        }
    )

    // Manage submitting a batch spec for execution.
    const { executeBatchSpec, isLoading: isExecuting, error: executeError } = useExecuteBatchSpec(batchSpecID)

    // Disable the execute button if any of the following are true:
    // * the batch spec code is invalid
    // * there was an error with the preview
    // * we're already in the middle of previewing or executing
    // * we haven't submitted the batch spec to the backend yet for the preview
    // * the batch spec on the backend is stale
    // * the current workspaces evaluation is not complete
    const [disableExecution, executionTooltip] = useMemo(() => {
        const disableExecution = Boolean(
            isValid !== true ||
                previewError ||
                isLoadingPreview ||
                isExecuting ||
                !batchSpecID ||
                batchSpecStale ||
                workspacesPreviewResolution?.state !== BatchSpecWorkspaceResolutionState.COMPLETED
        )
        // The execution tooltip only shows if the execute button is disabled, and explains why.
        const executionTooltip =
            isValid === false || previewError
                ? "There's a problem with your batch spec."
                : !batchSpecID
                ? 'Preview workspaces first before you run.'
                : batchSpecStale
                ? 'Update your workspaces preview before you run.'
                : isLoadingPreview || workspacesPreviewResolution?.state !== BatchSpecWorkspaceResolutionState.COMPLETED
                ? 'Wait for the preview to finish first.'
                : undefined

        return [disableExecution, executionTooltip]
    }, [
        batchSpecID,
        isValid,
        previewError,
        isLoadingPreview,
        isExecuting,
        batchSpecStale,
        workspacesPreviewResolution?.state,
    ])

    const errors =
        codeErrors.update || codeErrors.validation || previewError || executeError ? (
            <div className="w-100">
                {codeErrors.update && <ErrorAlert error={codeErrors.update} />}
                {codeErrors.validation && <ErrorAlert error={codeErrors.validation} />}
                {previewError && <ErrorAlert error={previewError} />}
                {executeError && <ErrorAlert error={executeError} />}
            </div>
        ) : null

    return (
        <div className="d-flex flex-column p-4 w-100 h-100">
            <div className="d-flex flex-0 justify-content-between">
                <div className="flex-1">
                    <PageHeader
                        path={[
                            { icon: BatchChangesIcon },
                            {
                                to: getNamespaceBatchChangesURL(selectedNamespace),
                                text: getNamespaceDisplayName(selectedNamespace),
                            },
                            { text: 'Create batch change' },
                        ]}
                        className="flex-1 pb-2"
                        annotation={
                            <FeedbackBadge status="experimental" feedback={{ mailto: 'support@sourcegraph.com' }} />
                        }
                        description="Run custom code over hundreds of repositories and manage the resulting changesets."
                    />

                    <NamespaceSelector
                        namespaces={namespaces}
                        selectedNamespace={selectedNamespace.id}
                        onSelect={setSelectedNamespace}
                    />
                </div>
                <div className="d-flex flex-column flex-0 align-items-center justify-content-center">
                    <ButtonTooltip
                        type="button"
                        className="btn btn-primary mb-2"
                        onClick={executeBatchSpec}
                        disabled={disableExecution}
                        tooltip={executionTooltip}
                    >
                        Run batch spec
                    </ButtonTooltip>
                    <BatchSpecDownloadLink name="new-batch-spec" originalInput={code} isLightTheme={isLightTheme}>
                        or download for src-cli
                    </BatchSpecDownloadLink>
                    <div className="form-group">
                        <label>
                            <input type="checkbox" className="mr-2" checked={noCache} onChange={onChangeNoCache} />
                            Disable cache
                        </label>
                    </div>
                </div>
            </div>
            <div className={classNames(styles.editorLayoutContainer, 'd-flex flex-1')}>
                <LibraryPane onReplaceItem={clearErrorsAndHandleCodeChange} />
                <div className={styles.editorContainer}>
                    <MonacoBatchSpecEditor
                        isLightTheme={isLightTheme}
                        value={code}
                        onChange={clearErrorsAndHandleCodeChange}
                    />
                </div>
                <div
                    className={classNames(
                        styles.workspacesPreviewContainer,
                        'd-flex flex-column align-items-center pl-4'
                    )}
                >
                    {errors}
                    <WorkspacesPreview
                        batchSpecID={batchSpecID}
                        currentPreviewRequestTime={currentPreviewRequestTime}
                        previewDisabled={previewDisabled}
                        preview={() => previewBatchSpec(debouncedCode)}
                        batchSpecStale={batchSpecStale}
                        excludeRepo={excludeRepo}
                    />
                </div>
            </div>
        </div>
    )
}

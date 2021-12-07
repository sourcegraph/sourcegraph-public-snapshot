import classNames from 'classnames'
import { noop } from 'lodash'
import React, { useCallback, useMemo, useState } from 'react'
import { useHistory } from 'react-router'

import { useMutation } from '@sourcegraph/shared/src/graphql/apollo'
import { BatchSpecWorkspaceResolutionState } from '@sourcegraph/shared/src/graphql/schema'
import {
    SettingsCascadeProps,
    SettingsOrgSubject,
    SettingsUserSubject,
} from '@sourcegraph/shared/src/settings/settings'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { ErrorAlert } from '@sourcegraph/web/src/components/alerts'
import { ButtonTooltip } from '@sourcegraph/web/src/components/ButtonTooltip'
import { Button, Container, Input, PageHeader } from '@sourcegraph/wildcard'

import { BatchChangesIcon } from '../../../batches/icons'
import { CreateEmptyBatchChangeResult, CreateEmptyBatchChangeVariables } from '../../../graphql-operations'
import { Settings } from '../../../schema/settings.schema'
import { BatchSpecDownloadLink } from '../BatchSpec'

import { CREATE_EMPTY_BATCH_CHANGE } from './backend'
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
    const [isSettingsFormOpen, setIsSettingsFormOpen] = useState(true)
    const [nameInput, setNameInput] = useState('')

    const [
        createEmptyBatchChange,
        { data: createEmptyBatchChangeData, loading: createEmptyBatchChangeLoading },
    ] = useMutation<CreateEmptyBatchChangeResult, CreateEmptyBatchChangeVariables>(CREATE_EMPTY_BATCH_CHANGE)

    const history = useHistory()
    const handleCancel = (): void => history.goBack()
    const handleCreate = (): void => {
        createEmptyBatchChange({
            variables: { namespace: selectedNamespace.id, name: nameInput },
        })
            .then(() => setIsSettingsFormOpen(false))
            .catch(noop)
    }

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
            <div className="d-flex flex-0 justify-content-between align-items-start">
                <PageHeader
                    path={[
                        { icon: BatchChangesIcon },
                        {
                            to: getNamespaceBatchChangesURL(selectedNamespace),
                            text: getNamespaceDisplayName(selectedNamespace),
                        },
                        { text: createEmptyBatchChangeData?.createEmptyBatchChange.name || 'Create batch change' },
                    ]}
                    className="flex-1 pb-2"
                    description="Run custom code over hundreds of repositories and manage the resulting changesets."
                />
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
                    <BatchSpecDownloadLink name="new-batch-spec" originalInput={code}>
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

            {isSettingsFormOpen ? (
                <div className={styles.settingsContainer}>
                    <h4>Batch specification settings</h4>
                    <Container>
                        <NamespaceSelector
                            namespaces={namespaces}
                            selectedNamespace={selectedNamespace.id}
                            onSelect={setSelectedNamespace}
                        />
                        <Input
                            className={styles.nameInput}
                            label="Batch change name"
                            value={nameInput}
                            onChange={event => setNameInput(event.target.value)}
                        />
                    </Container>
                    <div className="mt-3 align-self-end">
                        <Button variant="secondary" outline={true} className="mr-2" onClick={handleCancel}>
                            Cancel
                        </Button>
                        <Button variant="primary" onClick={handleCreate} disabled={createEmptyBatchChangeLoading}>
                            Create
                        </Button>
                    </div>
                </div>
            ) : (
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
            )}
        </div>
    )
}

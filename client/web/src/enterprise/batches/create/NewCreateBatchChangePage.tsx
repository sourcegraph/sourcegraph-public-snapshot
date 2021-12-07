// TODO: Rename me to editor page
import classNames from 'classnames'
import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import React, { useCallback, useMemo, useState } from 'react'

import { Scalars } from '@sourcegraph/shared/src/graphql-operations'
import { useMutation, useQuery } from '@sourcegraph/shared/src/graphql/apollo'
import { BatchSpecWorkspaceResolutionState } from '@sourcegraph/shared/src/graphql/schema'
import {
    SettingsCascadeProps,
    SettingsOrgSubject,
    SettingsUserSubject,
} from '@sourcegraph/shared/src/settings/settings'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { ErrorAlert } from '@sourcegraph/web/src/components/alerts'
import { ButtonTooltip } from '@sourcegraph/web/src/components/ButtonTooltip'
import { HeroPage } from '@sourcegraph/web/src/components/HeroPage'
import { LoadingSpinner, PageHeader } from '@sourcegraph/wildcard'

import { BatchChangesIcon } from '../../../batches/icons'
import {
    BatchChangeFields,
    EditBatchChangeFields,
    GetBatchChangeResult,
    GetBatchChangeVariables,
    OrgAreaOrganizationFields,
    UserAreaUserFields,
    CreateEmptyBatchChangeVariables,
    CreateEmptyBatchChangeResult,
} from '../../../graphql-operations'
import { Settings } from '../../../schema/settings.schema'
import { BatchSpecDownloadLink } from '../BatchSpec'

import { GET_BATCH_CHANGE, CREATE_EMPTY_BATCH_CHANGE } from './backend'
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

export interface NewCreateBatchChangePageProps extends ThemeProps, SettingsCascadeProps<Settings> {
    /** The namespace the batch change should be created in, or that it already belongs to. */
    namespace?: UserAreaUserFields | OrgAreaOrganizationFields
    /** The batch change name, if it already exists. */
    batchChangeName?: BatchChangeFields['name']
}

/**
 * TODO: Rename me once create/update settings form is ready
 * NewCreateBatchChangePage is the new SSBC-oriented page for creating a new batch change
 * or editing and re-executing a new batch spec for an existing one.
 */
export const NewCreateBatchChangePage: React.FunctionComponent<NewCreateBatchChangePageProps> = ({
    namespace,
    batchChangeName,
    ...props
}) => {
    const { data, loading } = useQuery<GetBatchChangeResult, GetBatchChangeVariables>(GET_BATCH_CHANGE, {
        // If we don't have a batch change name, the user hasn't created a batch change
        // yet, so skip the request.
        skip: !namespace || !batchChangeName,
        variables: {
            namespace: namespace?.id as Scalars['ID'],
            name: batchChangeName as BatchChangeFields['name'],
        },
        // Cache this data but always re-request it in the background when we revisit
        // this page to pick up newer changes.
        fetchPolicy: 'cache-and-network',
    })

    if (!batchChangeName) {
        return <EditPage namespace={namespace} {...props} />
    }

    if (loading) {
        return (
            <div className="text-center">
                <LoadingSpinner className="icon-inline mx-auto my-4" />
            </div>
        )
    }

    if (!data?.batchChange) {
        return <HeroPage icon={AlertCircleIcon} title="Batch change not found" />
    }

    return <EditPage namespace={namespace} batchChange={data.batchChange} {...props} />
}

interface EditPageProps extends ThemeProps, SettingsCascadeProps<Settings> {
    /**
     * The namespace the batch change should be created in, or that it already belongs to.
     * If none is provided, it will default to the user's own namespace.
     */
    namespace?: UserAreaUserFields | OrgAreaOrganizationFields
    /** The batch change, if it already exists */
    // TODO: Make this a required prop and otherwise show the "Create" form (exists on other branch)
    batchChange?: EditBatchChangeFields
}

const EditPage: React.FunctionComponent<EditPageProps> = ({
    namespace,
    batchChange,
    isLightTheme,
    settingsCascade,
}) => {
    // TODO: This will always be available once the "Create" form from the other
    // branch is ready.
    const batchSpecID = batchChange?.currentSpec.id

    const [
        createEmptyBatchChange,
        { data: createEmptyBatchChangeData, loading: createEmptyBatchChangeLoading },
    ] = useMutation<CreateEmptyBatchChangeResult, CreateEmptyBatchChangeVariables>(CREATE_EMPTY_BATCH_CHANGE)

    const { namespaces, defaultSelectedNamespace } = useNamespaces(settingsCascade, namespace)

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
        batchChange?.currentSpec.originalInput || helloWorldSample
    )

    // Track whenever the batch spec code that is presently in the editor gets ahead of
    // the batch spec that was last submitted to the backend.
    const [batchSpecStale, setBatchSpecStale] = useState(false)
    const markUnstale = useCallback(() => setBatchSpecStale(false), [])

    // Manage the batch spec that was last submitted to the backend for the workspaces preview.
    const {
        previewBatchSpec,
        currentPreviewRequestTime,
        isLoading: isLoadingPreview,
        error: previewError,
        clearError: clearPreviewError,
    } = usePreviewBatchSpec(batchSpecID || '', noCache, markUnstale)

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
    // * a batch spec has already been applied (the batch change is not a draft)
    // * the batch spec code is invalid
    // * there was an error with the preview
    // * we're already in the middle of previewing or executing
    // * we haven't submitted the batch spec to the backend yet for the preview
    // * the batch spec on the backend is stale
    // * the current workspaces evaluation is not complete
    const [disableExecution, executionTooltip] = useMemo(() => {
        const disableExecution = Boolean(
            batchChange === undefined ||
                batchChange.lastApplier !== null ||
                isValid !== true ||
                previewError ||
                isLoadingPreview ||
                isExecuting ||
                !currentPreviewRequestTime ||
                batchSpecStale ||
                workspacesPreviewResolution?.state !== BatchSpecWorkspaceResolutionState.COMPLETED
        )
        // The execution tooltip only shows if the execute button is disabled, and explains why.
        const executionTooltip =
            batchChange === undefined
                ? "There's nothing to run yet."
                : batchChange.lastApplier !== null
                ? 'This batch change has already had a spec applied.'
                : isValid === false || previewError
                ? "There's a problem with your batch spec."
                : !currentPreviewRequestTime
                ? 'Preview workspaces first before you run.'
                : batchSpecStale
                ? 'Update your workspaces preview before you run.'
                : isLoadingPreview || workspacesPreviewResolution?.state !== BatchSpecWorkspaceResolutionState.COMPLETED
                ? 'Wait for the preview to finish first.'
                : undefined

        return [disableExecution, executionTooltip]
    }, [
        batchChange,
        currentPreviewRequestTime,
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
                            { text: batchChange?.name || 'Create batch change' },
                        ]}
                        className="flex-1 pb-2"
                        description="Run custom code over hundreds of repositories and manage the resulting changesets."
                    />

                    {/* TODO: We'll be able to edit the namespace for an existing batch change from the new create/update form flow */}
                    {batchChange ? null : (
                        <NamespaceSelector
                            namespaces={namespaces}
                            selectedNamespace={selectedNamespace.id}
                            onSelect={setSelectedNamespace}
                        />
                    )}
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

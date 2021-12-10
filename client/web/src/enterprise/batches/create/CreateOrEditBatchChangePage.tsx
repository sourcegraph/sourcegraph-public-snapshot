import classNames from 'classnames'
import { noop } from 'lodash'
import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import React, { useCallback, useMemo, useState } from 'react'
import { useHistory } from 'react-router'

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
import { Button, Container, Input, LoadingSpinner } from '@sourcegraph/wildcard'

import {
    BatchChangeFields,
    EditBatchChangeFields,
    GetBatchChangeResult,
    GetBatchChangeVariables,
    CreateEmptyBatchChangeVariables,
    CreateEmptyBatchChangeResult,
} from '../../../graphql-operations'
import { Settings } from '../../../schema/settings.schema'
import { BatchChangePage } from '../BatchChangePage'
import { BatchSpecDownloadLink } from '../BatchSpec'

import { GET_BATCH_CHANGE, CREATE_EMPTY_BATCH_CHANGE } from './backend'
import styles from './CreateOrEditBatchChangePage.module.scss'
import { MonacoBatchSpecEditor } from './editor/MonacoBatchSpecEditor'
import helloWorldSample from './library/hello-world.batch.yaml'
import { LibraryPane } from './library/LibraryPane'
import { NamespaceSelector } from './NamespaceSelector'
import { useBatchSpecCode } from './useBatchSpecCode'
import { usePreviewBatchSpec } from './useBatchSpecPreview'
import { useExecuteBatchSpec } from './useExecuteBatchSpec'
import { useNamespaces } from './useNamespaces'
import { useBatchSpecWorkspaceResolution, WorkspacesPreview } from './workspaces-preview/WorkspacesPreview'
import { insertNameIntoLibraryItem, isMinimalBatchSpec } from './yaml-util'

export interface CreateOrEditBatchChangePageProps extends ThemeProps, SettingsCascadeProps<Settings> {
    /**
     * The id for the namespace that the batch change should be created in, or that it
     * already belongs to, if it already exists.
     */
    initialNamespaceID?: Scalars['ID']
    /** The batch change name, if it already exists. */
    batchChangeName?: BatchChangeFields['name']
}

/**
 * CreateOrEditBatchChangePage is the new SSBC-oriented page for creating a new batch change
 * or editing and re-executing a new batch spec for an existing one.
 */
export const CreateOrEditBatchChangePage: React.FunctionComponent<CreateOrEditBatchChangePageProps> = ({
    initialNamespaceID,
    batchChangeName,
    ...props
}) => {
    const { data, loading } = useQuery<GetBatchChangeResult, GetBatchChangeVariables>(GET_BATCH_CHANGE, {
        // If we don't have the batch change name or namespace, the user hasn't created a
        // batch change yet, so skip the request.
        skip: !initialNamespaceID || !batchChangeName,
        variables: {
            namespace: initialNamespaceID as Scalars['ID'],
            name: batchChangeName as BatchChangeFields['name'],
        },
        // Cache this data but always re-request it in the background when we revisit
        // this page to pick up newer changes.
        fetchPolicy: 'cache-and-network',
    })

    if (!batchChangeName) {
        return <CreatePage namespaceID={initialNamespaceID} {...props} />
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

    return <EditPage batchChange={data.batchChange} {...props} />
}

interface CreatePageProps extends SettingsCascadeProps<Settings> {
    /**
     * The namespace the batch change should be created in. If none is provided, it will
     * default to the user's own namespace.
     */
    namespaceID?: Scalars['ID']
}

const CreatePage: React.FunctionComponent<CreatePageProps> = ({ namespaceID, settingsCascade }) => {
    const [createEmptyBatchChange, { loading, error }] = useMutation<
        CreateEmptyBatchChangeResult,
        CreateEmptyBatchChangeVariables
    >(CREATE_EMPTY_BATCH_CHANGE)

    const { namespaces, defaultSelectedNamespace } = useNamespaces(settingsCascade, namespaceID)

    // The namespace selected for creating the new batch change under.
    const [selectedNamespace, setSelectedNamespace] = useState<SettingsUserSubject | SettingsOrgSubject>(
        defaultSelectedNamespace
    )

    const [nameInput, setNameInput] = useState('')

    const history = useHistory()
    const handleCancel = (): void => history.goBack()
    const handleCreate = (): void => {
        createEmptyBatchChange({
            variables: { namespace: selectedNamespace.id, name: nameInput },
        })
            .then(({ data }) => (data ? history.push(`${data.createEmptyBatchChange.url}/edit`) : noop()))
            // We destructure and surface the error from `useMutation` instead.
            .catch(noop)
    }

    return (
        <BatchChangePage namespace={selectedNamespace} title="Create batch change">
            <div className={styles.settingsContainer}>
                <h4>Batch specification settings</h4>
                <Container>
                    {error && <ErrorAlert error={error} />}
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
                        onKeyPress={event => {
                            if (event.key === 'Enter') {
                                handleCreate()
                            }
                        }}
                    />
                </Container>
                <div className="mt-3 align-self-end">
                    <Button variant="secondary" outline={true} className="mr-2" onClick={handleCancel}>
                        Cancel
                    </Button>
                    <Button variant="primary" onClick={handleCreate} disabled={loading}>
                        Create
                    </Button>
                </div>
            </div>
        </BatchChangePage>
    )
}

interface EditPageProps extends ThemeProps, SettingsCascadeProps<Settings> {
    /** The batch change, if it already exists */
    batchChange: EditBatchChangeFields
}

const EditPage: React.FunctionComponent<EditPageProps> = ({ batchChange, isLightTheme, settingsCascade }) => {
    const batchSpecID = batchChange.currentSpec.id

    const [
        createEmptyBatchChange,
        { data: createEmptyBatchChangeData, loading: createEmptyBatchChangeLoading },
    ] = useMutation<CreateEmptyBatchChangeResult, CreateEmptyBatchChangeVariables>(CREATE_EMPTY_BATCH_CHANGE)

    const { namespaces: _namespaces, defaultSelectedNamespace } = useNamespaces(
        settingsCascade,
        batchChange.namespace.id
    )

    // The namespace selected for creating the new batch spec under.
    const [selectedNamespace, _setSelectedNamespace] = useState<SettingsUserSubject | SettingsOrgSubject>(
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

    // Show the hello world sample code initially in the Monaco editor if the user hasn't
    // written any batch spec code yet, otherwise show the latest spec for the batch
    // change.
    const initialBatchSpecCode = useMemo(
        () =>
            isMinimalBatchSpec(batchChange.currentSpec.originalInput)
                ? insertNameIntoLibraryItem(helloWorldSample, batchChange.name)
                : batchChange.currentSpec.originalInput,
        [batchChange.currentSpec.originalInput, batchChange.name]
    )

    // Manage the batch spec input YAML code that's being edited.
    const { code, debouncedCode, isValid, handleCodeChange, excludeRepo, errors: codeErrors } = useBatchSpecCode(
        initialBatchSpecCode
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
    } = usePreviewBatchSpec(batchSpecID, noCache, markUnstale)

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
            batchChange.lastApplier !== null
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

    const buttons = (
        <>
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
        </>
    )

    return (
        <BatchChangePage
            namespace={selectedNamespace}
            title={batchChange.name}
            description={batchChange.description}
            actionButtons={buttons}
        >
            <div className={classNames(styles.editorLayoutContainer, 'd-flex flex-1')}>
                <LibraryPane name={batchChange.name} onReplaceItem={clearErrorsAndHandleCodeChange} />
                <div className={styles.editorContainer}>
                    <h4>Batch specification</h4>
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
        </BatchChangePage>
    )
}

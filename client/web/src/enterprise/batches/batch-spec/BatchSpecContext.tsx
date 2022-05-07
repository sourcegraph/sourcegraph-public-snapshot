import React, { useCallback, useMemo, useState } from 'react'

import { BatchSpecState, BatchSpecWorkspaceResolutionState, EditBatchChangeFields } from '../../../graphql-operations'

import { useExecuteBatchSpec } from './edit/useExecuteBatchSpec'
import { WorkspacePreviewFilters } from './edit/workspaces-preview/useWorkspaces'
import { useWorkspacesPreview, UseWorkspacesPreviewResult } from './edit/workspaces-preview/useWorkspacesPreview'
import { useBatchSpecCode, UseBatchSpecCodeResult } from './useBatchSpecCode'

export interface BatchSpecContextErrors {
    // Errors from trying to automatically apply updates to the batch spec code.
    codeUpdate?: string | Error
    // Errors from validating the batch spec code.
    codeValidation?: string | Error
    // Errors from trying to preview the workspaces that would be affected by the batch spec.
    preview?: string | Error
    // Errors from trying to execute the batch spec.
    execute?: string | Error
}

type NewBatchSpecState = EditBatchChangeFields['currentSpec'] & {
    // Whether or not the batch spec has already been applied.
    isApplied: boolean
    // Execution URL for this batch spec.
    executionURL: string
}

type EditorState = UseBatchSpecCodeResult & {
    execute: () => void
    // Whether or not the run batch spec button should be disabled, for example due to
    // there being a problem with the input batch spec YAML, or an execution already being
    // in progress. An optional tooltip string to display may be provided in place of
    // `true`.
    isExecutionDisabled: string | boolean
    // Options to apply to the workspaces preview and execution requests.
    executionOptions: ExecutionOptions
    // Callback to update options applied to workspaces preview and execution requests.
    setExecutionOptions: (options: ExecutionOptions) => void
}

type WorkspacesPreviewState = UseWorkspacesPreviewResult & {
    // Any filters to apply to the workspaces preview connection.
    filters?: WorkspacePreviewFilters
    // Callback to update the filters applied to the connection.
    setFilters: (filters: WorkspacePreviewFilters) => void
    // Whether or not the preview button should be disabled, for example due to there
    // being a problem with the input batch spec YAML, or a preview request already being
    // in flight. An optional tooltip string to display may be provided in place of
    // `true`.
    isPreviewDisabled: string | boolean
}

// Options to apply to the execution of a batch spec.
// NOTE: `runWithoutCache` is actually sent as options to the workspaces preview request,
// because it determines whether or not to use cached results for the workspaces.
export interface ExecutionOptions {
    runWithoutCache: boolean
}

const DEFAULT_EXECUTION_OPTIONS: ExecutionOptions = {
    runWithoutCache: false,
}

export interface BatchSpecContextState {
    readonly batchChange: EditBatchChangeFields
    readonly batchSpec: NewBatchSpecState

    // API for state managing the batch spec input YAML code in the Monaco editor.
    readonly editor: EditorState

    // API for state managing the workspaces resolution preview for the batch spec.
    readonly workspacesPreview: WorkspacesPreviewState

    readonly errors: BatchSpecContextErrors
}

export const defaultState = (): BatchSpecContextState => ({
    // eslint-disable-next-line @typescript-eslint/consistent-type-assertions
    batchChange: {} as EditBatchChangeFields,
    // eslint-disable-next-line @typescript-eslint/consistent-type-assertions
    batchSpec: {} as NewBatchSpecState,
    // eslint-disable-next-line @typescript-eslint/consistent-type-assertions
    editor: {} as EditorState,
    // eslint-disable-next-line @typescript-eslint/consistent-type-assertions
    workspacesPreview: {} as WorkspacesPreviewState,
    errors: {},
})

/**
 * TODO:
 *
 * @see BatchSpecContextProvider
 */
export const BatchSpecContext = React.createContext<BatchSpecContextState>(defaultState())

interface BatchSpecContextProviderProps {
    batchChange: EditBatchChangeFields
    refetchBatchChange?: () => Promise<unknown>
    batchSpec?: BatchSpecExecutionFields
}

export const BatchSpecContextProvider: React.FunctionComponent<BatchSpecContextProviderProps> = ({
    children,
    batchChange,
    refetchBatchChange,
    batchSpec: fullBatchSpec,
}) => {
    const {
        currentSpec,
        batchSpecs: { nodes },
    } = batchChange

    // The first node from the batch specs is the latest batch spec for a batch change. If
    // it's different from the `currentSpec` on the batch change, that means the latest
    // batch spec has not yet been applied.
    const batchSpec = fullBatchSpec || nodes[0] || currentSpec
    // TODO: This should probably just be a field on GraphQL.
    const isBatchSpecApplied = useMemo(() => currentSpec.id === batchSpec.id, [currentSpec.id, batchSpec.id])

    const editor = useBatchSpecCode(batchSpec.originalInput, batchChange.name)
    const { handleCodeChange, isValid, isServerStale } = editor

    const [filters, setFilters] = useState<WorkspacePreviewFilters>()
    const [executionOptions, setExecutionOptions] = useState<ExecutionOptions>(DEFAULT_EXECUTION_OPTIONS)

    // Manage the batch spec that was last submitted to the backend for the workspaces preview.
    const workspacesPreview = useWorkspacesPreview(batchSpec.id, {
        isBatchSpecApplied,
        namespaceID: batchChange.namespace.id,
        noCache: executionOptions.runWithoutCache,
        onComplete: refetchBatchChange,
        filters,
    })
    const { isInProgress: isWorkspacesPreviewInProgress, resolutionState } = workspacesPreview

    // Disable triggering a new preview if the batch spec code is invalid or if we're
    // already processing a preview.
    const isPreviewDisabled = useMemo<boolean | string>(
        () => (isValid !== true ? "There's a problem with your batch spec." : isWorkspacesPreviewInProgress),
        [isValid, isWorkspacesPreviewInProgress]
    )

    const { error: previewError, clearError: clearPreviewError, hasPreviewed } = workspacesPreview
    // Clear preview error when the batch spec code changes.
    const clearPreviewErrorsAndHandleCodeChange = useCallback(
        (newCode: string) => {
            clearPreviewError()
            handleCodeChange(newCode)
        },
        [handleCodeChange, clearPreviewError]
    )

    const isExecuting = batchSpec.state === BatchSpecState.QUEUED || batchSpec.state === BatchSpecState.PROCESSING
    const alreadyExecuted =
        batchSpec.applyURL !== null ||
        batchSpec.state === BatchSpecState.COMPLETED ||
        batchSpec.state === BatchSpecState.FAILED ||
        batchSpec.state === BatchSpecState.CANCELED ||
        batchSpec.state === BatchSpecState.CANCELING

    // Manage submitting a batch spec for execution.
    const { executeBatchSpec, isLoading: isExecutionRequestInProgress, error: executeError } = useExecuteBatchSpec(
        batchSpec.id
    )

    // Disable triggering a new execution if any of the following are true:
    // - The batch spec code is invalid.
    // - There was an error with the workspaces preview.
    // - We're in the middle of previewing or executing the batch spec.
    // - We haven't sent the latest spec to the backend for a workspaces preview yet.
    // - The batch spec on the backend is stale.
    // - The current workspaces resolution job is not complete.
    // - The batch spec is already executing, or has already been executed.
    const isExecutionDisabled = useMemo<boolean | string>(
        () =>
            isValid === false || previewError
                ? "There's a problem with your batch spec."
                : !hasPreviewed
                ? 'Preview workspaces first before you run.'
                : isServerStale
                ? 'Update your workspaces preview before you run.'
                : isWorkspacesPreviewInProgress || resolutionState !== BatchSpecWorkspaceResolutionState.COMPLETED
                ? 'Wait for the preview to finish first.'
                : isExecuting
                ? 'Batch spec is already executing.'
                : isExecutionRequestInProgress,
        [
            isValid,
            previewError,
            hasPreviewed,
            isWorkspacesPreviewInProgress,
            isExecutionRequestInProgress,
            isServerStale,
            resolutionState,
            isExecuting,
        ]
    )

    return (
        <BatchSpecContext.Provider
            value={{
                batchChange,
                batchSpec: {
                    ...batchSpec,
                    isApplied: isBatchSpecApplied,
                    executionURL: `${batchChange.url}/executions/${batchSpec.id}`,
                },
                editor: {
                    ...editor,
                    handleCodeChange: clearPreviewErrorsAndHandleCodeChange,
                    execute: executeBatchSpec,
                    isExecutionDisabled,
                    executionOptions,
                    setExecutionOptions,
                },
                workspacesPreview: { ...workspacesPreview, filters, setFilters, isPreviewDisabled },
                errors: {
                    codeUpdate: editor.errors.update,
                    codeValidation: editor.errors.validation,
                    preview: workspacesPreview.error,
                    execute: executeError || fullBatchSpec?.failureMessage || undefined,
                },
            }}
        >
            {children}
        </BatchSpecContext.Provider>
    )
}

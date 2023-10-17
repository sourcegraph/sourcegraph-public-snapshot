import React, { useCallback, useMemo, useState } from 'react'

import {
    type BatchSpecExecutionFields,
    BatchSpecState,
    BatchSpecWorkspaceResolutionState,
    type EditBatchChangeFields,
} from '../../../graphql-operations'

import { useExecuteBatchSpec } from './edit/useExecuteBatchSpec'
import type { WorkspacePreviewFilters } from './edit/workspaces-preview/useWorkspaces'
import { useWorkspacesPreview, type UseWorkspacesPreviewResult } from './edit/workspaces-preview/useWorkspacesPreview'
import { useBatchSpecCode, type UseBatchSpecCodeResult } from './useBatchSpecCode'

export interface BatchSpecContextErrors {
    // Errors from trying to automatically apply updates to the batch spec code.
    codeUpdate?: string | Error
    // Errors from validating the batch spec code.
    codeValidation?: string | Error
    // Errors from trying to preview the workspaces that would be affected by the batch spec.
    preview?: string | Error
    // Errors from trying to execute the batch spec.
    execute?: string | Error
    // Errors from trying to perform an action on a batch spec.
    actions?: string | Error
}

type MinimalBatchSpecFields = EditBatchChangeFields['currentSpec'] & Partial<BatchSpecExecutionFields>

type NewBatchSpecState<BatchSpecFields extends MinimalBatchSpecFields> = BatchSpecFields & {
    // Whether or not the batch spec has already been applied.
    isApplied: boolean
    // Execution URL for this batch spec.
    executionURL: string
    // Whether or not the batch spec is actively being executed.
    isExecuting: boolean
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

export interface BatchSpecContextState<BatchSpecFields extends MinimalBatchSpecFields = MinimalBatchSpecFields> {
    readonly batchChange: EditBatchChangeFields
    readonly batchSpec: NewBatchSpecState<BatchSpecFields>

    // API for state managing the batch spec input YAML code in the Monaco editor.
    readonly editor: EditorState

    // API for state managing the workspaces resolution preview for the batch spec.
    readonly workspacesPreview: WorkspacesPreviewState

    readonly errors: BatchSpecContextErrors
    readonly setActionsError: (error: string | Error | undefined) => void
}

export const defaultState = (): BatchSpecContextState<MinimalBatchSpecFields> => ({
    // eslint-disable-next-line @typescript-eslint/consistent-type-assertions
    batchChange: {} as EditBatchChangeFields,
    // eslint-disable-next-line @typescript-eslint/consistent-type-assertions
    batchSpec: {} as NewBatchSpecState<MinimalBatchSpecFields>,
    // eslint-disable-next-line @typescript-eslint/consistent-type-assertions
    editor: {} as EditorState,
    // eslint-disable-next-line @typescript-eslint/consistent-type-assertions
    workspacesPreview: {} as WorkspacesPreviewState,
    errors: {},
    setActionsError: () => {},
})

const BatchSpecContext = React.createContext<BatchSpecContextState<MinimalBatchSpecFields>>(defaultState())

interface BatchSpecContextProviderProps<BatchSpecFields extends MinimalBatchSpecFields> {
    batchChange: EditBatchChangeFields
    refetchBatchChange?: () => Promise<unknown>
    batchSpec: BatchSpecFields
    /** FOR TESTING ONLY */
    testState?: Partial<BatchSpecContextState<BatchSpecFields>>
}

export const BatchSpecContextProvider = <BatchSpecFields extends MinimalBatchSpecFields>({
    children,
    batchChange,
    refetchBatchChange,
    batchSpec,
    testState,
}: React.PropsWithChildren<BatchSpecContextProviderProps<BatchSpecFields>>): JSX.Element => {
    const { currentSpec } = batchChange

    // TODO: This should probably just be a field on GraphQL.
    const isBatchSpecApplied = useMemo(() => currentSpec.id === batchSpec.id, [currentSpec.id, batchSpec.id])

    const editor = useBatchSpecCode(batchSpec.originalInput, batchChange.name)
    const { handleCodeChange, isValid, isServerStale: isServerBatchSpecYAMLStale } = editor

    const [filters, setFilters] = useState<WorkspacePreviewFilters>()
    const [executionOptions, setExecutionOptions] = useState<ExecutionOptions>(DEFAULT_EXECUTION_OPTIONS)

    const isServerStale = isServerBatchSpecYAMLStale

    // Manage the batch spec that was last submitted to the backend for the workspaces preview.
    const workspacesPreview = useWorkspacesPreview(batchSpec.id, {
        isBatchSpecApplied,
        namespaceID: batchChange.namespace.id,
        noCache: executionOptions.runWithoutCache,
        onComplete: refetchBatchChange,
        filters,
        batchChange: batchChange.id,
    })
    const {
        isInProgress: isWorkspacesPreviewInProgress,
        resolutionState,
        error: previewError,
        clearError: clearPreviewError,
        hasPreviewed,
        preview,
    } = workspacesPreview

    // Disable triggering a new preview if the batch spec code is invalid or if we're
    // already processing a preview.
    const isPreviewDisabled = useMemo<boolean | string>(
        () => (isValid !== true ? "There's a problem with your batch spec." : isWorkspacesPreviewInProgress),
        [isValid, isWorkspacesPreviewInProgress]
    )

    // Clear preview error when the batch spec code changes.
    const clearPreviewErrorsAndHandleCodeChange = useCallback(
        (newCode: string) => {
            clearPreviewError()
            handleCodeChange(newCode)
        },
        [handleCodeChange, clearPreviewError]
    )

    const [actionsError, setActionsError] = useState<string | Error | undefined>()

    // TODO: This should probably just be a field on GraphQL.
    const isExecuting = batchSpec.state === BatchSpecState.QUEUED || batchSpec.state === BatchSpecState.PROCESSING

    // Manage submitting a batch spec for execution.
    const {
        executeBatchSpec,
        isLoading: isExecutionRequestInProgress,
        error: executeError,
    } = useExecuteBatchSpec(batchSpec.id, executionOptions.runWithoutCache)

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
                    isExecuting,
                    ...testState?.batchSpec,
                },
                editor: {
                    ...editor,
                    handleCodeChange: clearPreviewErrorsAndHandleCodeChange,
                    execute: executeBatchSpec,
                    isExecutionDisabled,
                    executionOptions,
                    isServerStale,
                    setExecutionOptions,
                    ...testState?.editor,
                },
                workspacesPreview: {
                    ...workspacesPreview,
                    preview,
                    filters,
                    setFilters,
                    isPreviewDisabled,
                    ...testState?.workspacesPreview,
                },
                errors: {
                    codeUpdate: editor.errors.update,
                    codeValidation: editor.errors.validation,
                    preview: workspacesPreview.error,
                    execute: executeError || batchSpec.failureMessage || undefined,
                    actions: actionsError,
                    ...testState?.errors,
                },
                setActionsError,
                ...testState,
            }}
        >
            {children}
        </BatchSpecContext.Provider>
    )
}

export const useBatchSpecContext = <
    BatchSpecFields extends MinimalBatchSpecFields = MinimalBatchSpecFields
>(): BatchSpecContextState<BatchSpecFields> => {
    const context = React.useContext<BatchSpecContextState<BatchSpecFields>>(
        BatchSpecContext as unknown as React.Context<BatchSpecContextState<BatchSpecFields>>
    )
    if (!context) {
        throw new Error('useBatchSpecContext must be used under BatchSpecContextProvider')
    }
    return context
}

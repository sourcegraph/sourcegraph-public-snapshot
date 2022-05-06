import React, { useCallback, useMemo, useState } from 'react'

import { noop } from 'lodash'

import { EditBatchChangeFields } from '../../../graphql-operations'

import { WorkspacePreviewFilters } from './edit/workspaces-preview/useWorkspaces'
import { useWorkspacesPreview, UseWorkspacesPreviewResult } from './edit/workspaces-preview/useWorkspacesPreview'
import { useBatchSpecCode, UseBatchSpecCodeResult } from './useBatchSpecCode'

// TODO: This is probably just edit context LOL
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

type BatchSpecState = EditBatchChangeFields['currentSpec'] & {
    // Whether or not the batch spec has already been applied.
    isApplied: boolean
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

export interface BatchSpecContextState {
    readonly batchChange: EditBatchChangeFields
    readonly batchSpec: BatchSpecState

    readonly editor: UseBatchSpecCodeResult

    readonly workspacesPreview: WorkspacesPreviewState

    readonly errors: BatchSpecContextErrors
    setError: (type: keyof BatchSpecContextErrors, error: string | Error | undefined) => void
}

export const defaultState = (): BatchSpecContextState => ({
    // eslint-disable-next-line @typescript-eslint/consistent-type-assertions
    batchChange: {} as EditBatchChangeFields,
    // eslint-disable-next-line @typescript-eslint/consistent-type-assertions
    batchSpec: {} as BatchSpecState,
    // eslint-disable-next-line @typescript-eslint/consistent-type-assertions
    editor: {} as UseBatchSpecCodeResult,
    // eslint-disable-next-line @typescript-eslint/consistent-type-assertions
    workspacesPreview: {} as WorkspacesPreviewState,
    errors: {},
    setError: noop,
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
}

export const BatchSpecContextProvider: React.FunctionComponent<BatchSpecContextProviderProps> = ({
    children,
    batchChange,
    refetchBatchChange,
}) => {
    const {
        currentSpec,
        batchSpecs: { nodes },
    } = batchChange

    // The first node from the batch specs is the latest batch spec for a batch change. If
    // it's different from the `currentSpec` on the batch change, that means the latest
    // batch spec has not yet been applied.
    const latestSpec = nodes[0] || currentSpec
    // TODO: This should probably just be a field on GraphQL.
    const isLatestSpecApplied = useMemo(() => currentSpec.id === latestSpec.id, [currentSpec.id, latestSpec.id])

    const editor = useBatchSpecCode(latestSpec.originalInput, batchChange.name)
    const { handleCodeChange, isValid } = editor

    const [filters, setFilters] = useState<WorkspacePreviewFilters>()

    // Manage the batch spec that was last submitted to the backend for the workspaces preview.
    const workspacesPreview = useWorkspacesPreview(latestSpec.id, {
        isBatchSpecApplied: isLatestSpecApplied,
        namespaceID: batchChange.namespace.id,
        // TODO:
        // noCache: executionOptions.runWithoutCache,
        noCache: false,
        onComplete: refetchBatchChange,
        filters,
    })
    const { isInProgress: isWorkspacesPreviewInProgress } = workspacesPreview

    // Disable triggering a new preview if the batch spec code is invalid or if we're
    // already processing a preview.
    const isPreviewDisabled = useMemo(
        () => (isValid !== true ? "There's a problem with your batch spec." : isWorkspacesPreviewInProgress),
        [isValid, isWorkspacesPreviewInProgress]
    )

    // TODO: This will probably go away
    const [errors, setErrors] = useState<Pick<BatchSpecContextErrors, 'execute'>>({})
    const setError = useCallback((type: keyof BatchSpecContextErrors, error: string | Error | undefined) => {
        setErrors(errors => ({ ...errors, [type]: error }))
    }, [])

    const { clearError: clearPreviewError } = workspacesPreview
    // Clear preview error when the batch spec code changes.
    const clearPreviewErrorsAndHandleCodeChange = useCallback(
        (newCode: string) => {
            clearPreviewError()
            handleCodeChange(newCode)
        },
        [handleCodeChange, clearPreviewError]
    )

    return (
        <BatchSpecContext.Provider
            value={{
                batchChange,
                batchSpec: { ...latestSpec, isApplied: isLatestSpecApplied },
                editor: { ...editor, handleCodeChange: clearPreviewErrorsAndHandleCodeChange },
                errors: {
                    ...errors,
                    codeUpdate: editor.errors.update,
                    codeValidation: editor.errors.validation,
                    preview: workspacesPreview.error,
                },
                workspacesPreview: { ...workspacesPreview, filters, setFilters, isPreviewDisabled },
                setError,
            }}
        >
            {children}
        </BatchSpecContext.Provider>
    )
}

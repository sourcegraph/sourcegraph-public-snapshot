import { useCallback, useState } from 'react'

import { Scalars } from '@sourcegraph/shared/src/graphql-operations'
import { useMutation } from '@sourcegraph/shared/src/graphql/graphql'

import { ReplaceBatchSpecInputResult, ReplaceBatchSpecInputVariables } from '../../../graphql-operations'

import { REPLACE_BATCH_SPEC_INPUT } from './backend'

interface UsePreviewBatchSpecResult {
    /**
     * Method to invoke the appropriate GraphQL mutation to submit the batch spec input
     * YAML to the backend and request a preview of the workspaces it would affect.
     */
    previewBatchSpec: (code: string) => Promise<void>
    /**
     * A reference to the time that we last requested a preview for the workspaces
     * evaluation as a form of unique ID so that interested components can know when the
     * workspaces list is stale and check for updates.
     */
    currentPreviewRequestTime?: string
    /** Whether or not a preview request is currently in flight. */
    isLoading: boolean
    /** Any error from `previewBatchSpec`. */
    error?: Error
    /** Callback to clear `error` when it is no longer relevant. */
    clearError: () => void
}

/**
 * Custom hook for "Create" page which packages up business logic and exposes an API for
 * powering the "preview" aspect of the workflow, i.e. submitting the batch spec input
 * YAML code to the backend to enqueue a batch spec resolution job to evaluate the
 * workspaces that a batch spec would run over.
 *
 * @param batchSpecID The id of the existing batch spec that would be replaced on a new preview.
 * @param noCache Whether or not the batch spec should be executed with the cache disabled.
 * @param onComplete An optional (stable) callback to invoke when the mutation is complete.
 */
export const usePreviewBatchSpec = (
    batchSpecID: Scalars['ID'],
    noCache: boolean,
    onComplete?: () => void
): UsePreviewBatchSpecResult => {
    // Mutation to replace the existing batch spec input YAML and re-evaluate the workspaces.
    const [replaceBatchSpecInput, { loading: isLoading }] = useMutation<
        ReplaceBatchSpecInputResult,
        ReplaceBatchSpecInputVariables
    >(REPLACE_BATCH_SPEC_INPUT)

    const [error, setError] = useState<Error>()
    // We keep a reference to the time that we last requested a preview for the workspaces
    // evaluation as a form of unique ID so that interested components can know when the
    // workspaces list is stale and check for updates.
    const [currentPreviewRequestTime, setCurrentPreviewRequestTime] = useState<string>()

    const previewBatchSpec = useCallback(
        (code: string) => {
            setError(undefined)

            const preview = (): Promise<unknown> =>
                replaceBatchSpecInput({ variables: { spec: code, previousSpec: batchSpecID, noCache } })

            return preview()
                .then(() => {
                    setCurrentPreviewRequestTime(new Date().toISOString())
                    onComplete?.()
                })
                .catch(setError)
        },
        [batchSpecID, noCache, replaceBatchSpecInput, onComplete]
    )

    return {
        previewBatchSpec,
        currentPreviewRequestTime,
        isLoading,
        error,
        clearError: () => setError(undefined),
    }
}

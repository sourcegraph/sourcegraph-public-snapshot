import { useCallback, useMemo, useState } from 'react'

import { Scalars } from '@sourcegraph/shared/src/graphql-operations'
import { useMutation } from '@sourcegraph/shared/src/graphql/graphql'

import {
    CreateBatchSpecFromRawResult,
    CreateBatchSpecFromRawVariables,
    ReplaceBatchSpecInputResult,
    ReplaceBatchSpecInputVariables,
} from '../../../graphql-operations'

import { CREATE_BATCH_SPEC_FROM_RAW, REPLACE_BATCH_SPEC_INPUT } from './backend'

interface UsePreviewBatchSpecResult {
    /**
     * Method to invoke the appropriate GraphQL mutation to submit the batch spec input
     * YAML to the backend and request a preview of the workspaces it would affect.
     */
    previewBatchSpec: (code: string) => Promise<void>
    /** Whether or not a preview request is currently in flight. */
    isLoading: boolean
    /** Any error from `previewBatchSpec`. */
    error?: Error
    /** Callback to clear `error` when it is no longer relevant. */
    clearError: () => void
    /**
     * Whether or not the user has previewed their batch spec at least once since arriving
     * on the page.
     */
    hasPreviewed: boolean
}

/**
 * Custom hook for "CreateOrEdit" page which packages up business logic and exposes an API
 * for powering the "preview" aspect of the workflow, i.e. submitting the batch spec input
 * YAML code to the backend to enqueue a batch spec resolution job to evaluate the
 * workspaces that a batch spec would run over. It will smartly determine whether or not
 * to create a new batch spec from raw or replace an existing one depending on whether or
 * not the most recent batch spec has already been applied.
 *
 * @param batchSpecID The ID of the most recent, existing batch spec.
 * @param isBatchSpecApplied Whether or not the existing batch spec was already applied to
 * the batch change to which it belongs.
 * @param namespaceID The ID of the namespace to which the batch change and new batch spec
 * should belong.
 * @param noCache Whether or not the batch spec should be executed with the cache
 * disabled.
 * @param onComplete An optional (stable) callback to invoke when the mutation is
 * complete.
 */
export const usePreviewBatchSpec = (
    batchSpecID: Scalars['ID'],
    isBatchSpecApplied: boolean,
    namespaceID: Scalars['ID'],
    noCache: boolean,
    onComplete?: () => void
): UsePreviewBatchSpecResult => {
    // Track whether the user has previewed the batch spec workspaces at least once.
    const [hasPreviewed, setHasPreviewed] = useState(false)

    // Mutation to create a new batch spec from the raw input YAML code.
    const [createBatchSpecFromRaw, { loading: createBatchSpecFromRawLoading }] = useMutation<
        CreateBatchSpecFromRawResult,
        CreateBatchSpecFromRawVariables
    >(CREATE_BATCH_SPEC_FROM_RAW)

    // Mutation to replace the existing batch spec input YAML and re-evaluate the workspaces.
    const [replaceBatchSpecInput, { loading: replaceBatchSpecInputLoading }] = useMutation<
        ReplaceBatchSpecInputResult,
        ReplaceBatchSpecInputVariables
    >(REPLACE_BATCH_SPEC_INPUT)

    const isLoading = useMemo(() => createBatchSpecFromRawLoading || replaceBatchSpecInputLoading, [
        createBatchSpecFromRawLoading,
        replaceBatchSpecInputLoading,
    ])

    const [error, setError] = useState<Error>()

    const previewBatchSpec = useCallback(
        (code: string) => {
            setError(undefined)

            const preview = (): Promise<unknown> =>
                isBatchSpecApplied
                    ? createBatchSpecFromRaw({
                          variables: { spec: code, namespace: namespaceID, noCache },
                      })
                    : replaceBatchSpecInput({ variables: { spec: code, previousSpec: batchSpecID, noCache } })

            return preview()
                .then(() => {
                    setHasPreviewed(true)
                    onComplete?.()
                })
                .catch(setError)
        },
        [
            batchSpecID,
            namespaceID,
            isBatchSpecApplied,
            noCache,
            createBatchSpecFromRaw,
            replaceBatchSpecInput,
            onComplete,
        ]
    )

    return {
        previewBatchSpec,
        isLoading,
        error,
        clearError: () => setError(undefined),
        hasPreviewed,
    }
}

import { useCallback, useMemo, useState } from 'react'

import { Scalars } from '@sourcegraph/shared/src/graphql-operations'
import { useMutation } from '@sourcegraph/shared/src/graphql/graphql'
import { SettingsOrgSubject, SettingsUserSubject } from '@sourcegraph/shared/src/settings/settings'

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
    /**
     * The id of the batch spec we are editing and evaluating workspaces for, if one has
     * been created.
     */
    batchSpecID?: Scalars['ID']
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
 * @param namespace The user or organization `SettingsSubject` which determines where the
 * resultant batch change would live.
 * @param onComplete An optional (stable) callback to invoke when the mutation is complete.
 */
export const usePreviewBatchSpec = (
    namespace: SettingsUserSubject | SettingsOrgSubject,
    noCache: boolean,
    onComplete?: () => void
): UsePreviewBatchSpecResult => {
    // Mutation to create a new batch spec from the raw input YAML code.
    const [
        createBatchSpecFromRaw,
        { data: createBatchSpecFromRawData, loading: createBatchSpecFromRawLoading },
    ] = useMutation<CreateBatchSpecFromRawResult, CreateBatchSpecFromRawVariables>(CREATE_BATCH_SPEC_FROM_RAW)

    // Mutation to replace the original batch spec input YAML and re-evaluate the workspaces.
    const [
        replaceBatchSpecInput,
        { data: replaceBatchSpecInputData, loading: replaceBatchSpecInputLoading },
    ] = useMutation<ReplaceBatchSpecInputResult, ReplaceBatchSpecInputVariables>(REPLACE_BATCH_SPEC_INPUT)

    const batchSpecID = useMemo(
        () =>
            // The id shouldn't change, but if we have data from replacing the batch spec
            // input already, the initial batch spec data has been superseded, so prefer
            // the replaced batch spec id.
            replaceBatchSpecInputData?.replaceBatchSpecInput.id ||
            createBatchSpecFromRawData?.createBatchSpecFromRaw.id,
        [replaceBatchSpecInputData, createBatchSpecFromRawData]
    )

    const isLoading = useMemo(() => createBatchSpecFromRawLoading || replaceBatchSpecInputLoading, [
        createBatchSpecFromRawLoading,
        replaceBatchSpecInputLoading,
    ])

    const [error, setError] = useState<Error>()
    // We keep a reference to the time that we last requested a preview for the workspaces
    // evaluation as a form of unique ID so that interested components can know when the
    // workspaces list is stale and check for updates.
    const [currentPreviewRequestTime, setCurrentPreviewRequestTime] = useState<string>()

    const previewBatchSpec = useCallback(
        (code: string) => {
            setError(undefined)

            // If we have a batch spec ID already, we're replacing the existing batch spec
            // input YAML with a new one.
            const preview = (): Promise<unknown> =>
                batchSpecID
                    ? replaceBatchSpecInput({ variables: { spec: code, previousSpec: batchSpecID, noCache } })
                    : // Otherwise, we're creating a new batch spec from the raw spec input YAML.
                      createBatchSpecFromRaw({
                          variables: { spec: code, namespace: namespace.id, noCache },
                      })

            return preview()
                .then(() => {
                    setCurrentPreviewRequestTime(new Date().toISOString())
                    onComplete?.()
                })
                .catch(setError)
        },
        [batchSpecID, namespace, noCache, createBatchSpecFromRaw, replaceBatchSpecInput, onComplete]
    )

    return {
        previewBatchSpec,
        batchSpecID,
        currentPreviewRequestTime,
        isLoading,
        error,
        clearError: () => setError(undefined),
    }
}

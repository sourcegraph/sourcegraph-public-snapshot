import { useCallback, useState, useEffect } from 'react'

import { useLazyQuery, useMutation, useQuery } from '@sourcegraph/http-client'
import { screenReaderAnnounce } from '@sourcegraph/wildcard'

import {
    type CreateBatchSpecFromRawResult,
    type CreateBatchSpecFromRawVariables,
    type ReplaceBatchSpecInputResult,
    type ReplaceBatchSpecInputVariables,
    type Scalars,
    BatchSpecWorkspaceResolutionState,
    type WorkspaceResolutionStatusVariables,
    type WorkspaceResolutionStatusResult,
    type BatchSpecWorkspacesPreviewResult,
    type BatchSpecWorkspacesPreviewVariables,
    type BatchSpecImportingChangesetsVariables,
    type BatchSpecImportingChangesetsResult,
} from '../../../../../graphql-operations'
import {
    CREATE_BATCH_SPEC_FROM_RAW,
    IMPORTING_CHANGESETS,
    REPLACE_BATCH_SPEC_INPUT,
    WORKSPACES,
    WORKSPACE_RESOLUTION_STATUS,
} from '../../../create/backend'

import { CHANGESETS_PER_PAGE_COUNT } from './useImportingChangesets'
import { WORKSPACES_PER_PAGE_COUNT, type WorkspacePreviewFilters } from './useWorkspaces'

export type ResolutionState = BatchSpecWorkspaceResolutionState | 'UNSTARTED' | 'CANCELED'

export interface UseWorkspacesPreviewResult {
    /**
     * Method to invoke the appropriate GraphQL mutation to submit the batch spec input
     * YAML to the backend and request a preview of the workspaces it would affect.
     */
    preview: (code: string) => Promise<void>
    /** Method to invoke to cancel any current workspaces resolution job. */
    cancel: () => void
    /**
     * Whether or not a preview request is currently in flight or the current workspaces
     * resolution job is in progress.
     */
    isInProgress: boolean
    /** The status of the current workspaces resolution job. */
    resolutionState: ResolutionState
    /** Any error from `previewBatchSpec` or the workspaces resolution job. */
    error?: string
    /** Callback to clear `error` when it is no longer relevant. */
    clearError: () => void
    /**
     * Whether or not the user has previewed their batch spec at least once since arriving
     * on the page.
     */
    hasPreviewed: boolean
    /** Whether or not the batch spec should be executed with the cache disabled. */
    noCache: boolean
}

interface UseWorkspacesPreviewOptions {
    /**
     * Whether or not the existing batch spec was already applied to the batch change to
     * which it belongs.
     */
    isBatchSpecApplied: boolean
    /** The ID of the namespace to which the batch change and new batch spec should belong. */
    namespaceID: Scalars['ID']
    /** Whether or not the batch spec should be executed with the cache disabled. */
    noCache: boolean
    /** An optional (stable) callback to invoke when the workspaces resolution job completes. */
    onComplete?: () => void
    /** Any filters currently applied to the workspaces connection preview. */
    filters?: WorkspacePreviewFilters
    /** The ID of the batch change associated with the batch spec. */
    batchChange: Scalars['ID']
}

export const POLLING_INTERVAL = 1000

export type WorkspaceResolution = (WorkspaceResolutionStatusResult['node'] & {
    __typename: 'BatchSpec'
})['workspaceResolution']

const getResolution = (queryResult?: WorkspaceResolutionStatusResult | null): WorkspaceResolution =>
    queryResult?.node?.__typename === 'BatchSpec' ? queryResult.node.workspaceResolution : null

/**
 * Custom hook to power the "preview" aspect of the batch spec creation workflow, i.e.
 * submitting batch spec input YAML code, enqueing a resolution job to evaluate the
 * workspaces that that batch spec would run over, and polling until the resolution job is
 * complete. It will smartly determine whether or not to create a new batch spec from raw
 * or replace an existing one depending on whether or not the most recent batch spec has
 * already been applied. Returns an API object in order to respond trigger a new preview,
 * monitor the resolution job progress, and handle any errors.
 *
 * @param batchSpecID The ID of the most recent, existing batch spec that we are replacing
 * @param options Aspects of the batch spec and properties to configure with the preview
 */
export const useWorkspacesPreview = (
    batchSpecID: Scalars['ID'],
    { isBatchSpecApplied, namespaceID, noCache, onComplete, filters, batchChange }: UseWorkspacesPreviewOptions
): UseWorkspacesPreviewResult => {
    const [currentBatchSpecID, setCurrentBatchSpecID] = useState(batchSpecID)
    // Track whether the user has previewed the batch spec workspaces at least once.
    const [hasRequestedPreview, setHasRequestedPreview] = useState(false)
    const [hasPreviewed, setHasPreviewed] = useState(false)

    // Mutation to create a new batch spec from the raw input YAML code.
    const [createBatchSpecFromRaw] = useMutation<CreateBatchSpecFromRawResult, CreateBatchSpecFromRawVariables>(
        CREATE_BATCH_SPEC_FROM_RAW
    )

    // Mutation to replace the existing batch spec input YAML and re-evaluate the workspaces.
    const [replaceBatchSpecInput] = useMutation<ReplaceBatchSpecInputResult, ReplaceBatchSpecInputVariables>(
        REPLACE_BATCH_SPEC_INPUT
    )

    const [isInProgress, setIsInProgress] = useState(false)
    // A computed state based on the state of any active workspaces resolution job as well
    // as any actions the user has taken on this page so far.
    const [uiState, setUIState] = useState<ResolutionState>('UNSTARTED')
    const [error, setError] = useState<string>()

    // Once we submit a batch spec to be previewed, we will poll for the resolution status
    // until it completes. We also request this upfront in case a workspace resolution is
    // already in progress.
    const { data, startPolling, stopPolling } = useQuery<
        WorkspaceResolutionStatusResult,
        WorkspaceResolutionStatusVariables
    >(WORKSPACE_RESOLUTION_STATUS, {
        variables: { batchSpec: currentBatchSpecID },
        fetchPolicy: 'network-only',
        onError: error => setError(error.message),
    })

    const stop = useCallback(() => {
        stopPolling()
        setIsInProgress(false)
    }, [stopPolling])

    const cancel = useCallback(() => {
        setError(undefined)
        stop()
        setUIState('CANCELED')
    }, [stop])

    const previewBatchSpec = useCallback(
        (code: string) => {
            // Update state
            setError(undefined)
            setIsInProgress(true)

            // Determine which mutation to use, depending on if the latest batch spec we
            // have was already applied or not.
            const preview = (): Promise<
                | CreateBatchSpecFromRawResult['createBatchSpecFromRaw']
                | ReplaceBatchSpecInputResult['replaceBatchSpecInput']
                | null
                | undefined
            > =>
                isBatchSpecApplied
                    ? createBatchSpecFromRaw({
                          variables: { spec: code, namespace: namespaceID, batchChange },
                      }).then(result => result.data?.createBatchSpecFromRaw)
                    : replaceBatchSpecInput({
                          variables: { spec: code, previousSpec: currentBatchSpecID },
                      }).then(result => result.data?.replaceBatchSpecInput)

            return preview()
                .then(result => {
                    const batchSpec = result?.__typename === 'BatchSpec' ? result : null
                    if (!batchSpec) {
                        return
                    }
                    setCurrentBatchSpecID(batchSpec.id)
                    const resolution = batchSpec.workspaceResolution
                    if (resolution?.state) {
                        // Set to the current workspace resolution state.
                        setUIState(resolution.state)
                    }
                    if (resolution?.failureMessage) {
                        setError(resolution.failureMessage)
                    }

                    setHasRequestedPreview(true)
                })
                .catch((error: Error) => {
                    setError(error.message)
                    setIsInProgress(false)
                })
        },
        [
            currentBatchSpecID,
            namespaceID,
            isBatchSpecApplied,
            createBatchSpecFromRaw,
            replaceBatchSpecInput,
            batchChange,
        ]
    )

    const [fetchWorkspaces] = useLazyQuery<BatchSpecWorkspacesPreviewResult, BatchSpecWorkspacesPreviewVariables>(
        WORKSPACES,
        {
            variables: {
                batchSpec: currentBatchSpecID,
                after: null,
                first: WORKSPACES_PER_PAGE_COUNT,
                search: filters?.search ?? null,
            },
            fetchPolicy: 'cache-and-network',
        }
    )

    const [fetchImportingChangesets] = useLazyQuery<
        BatchSpecImportingChangesetsResult,
        BatchSpecImportingChangesetsVariables
    >(IMPORTING_CHANGESETS, {
        variables: {
            batchSpec: currentBatchSpecID,
            after: null,
            first: CHANGESETS_PER_PAGE_COUNT,
        },
        fetchPolicy: 'cache-and-network',
    })

    // This effect triggers on workspaces resolution job status changes from the backend
    // and updates user-facing state.
    useEffect(() => {
        const resolution = getResolution(data)
        if (resolution?.state) {
            setUIState(resolution.state)
        }
        if (resolution?.failureMessage) {
            setError(resolution.failureMessage)
        }
    }, [data])

    // This effect triggers on computed `uiState` changes and controls the polling process.
    useEffect(() => {
        if (
            uiState === BatchSpecWorkspaceResolutionState.QUEUED ||
            uiState === BatchSpecWorkspaceResolutionState.PROCESSING
        ) {
            setError(undefined)
            // If the workspace resolution is still queued or processing, start polling.
            setIsInProgress(true)
            startPolling(POLLING_INTERVAL)
        } else if (
            uiState === BatchSpecWorkspaceResolutionState.ERRORED ||
            uiState === BatchSpecWorkspaceResolutionState.FAILED
        ) {
            screenReaderAnnounce('Workspaces preview failed.')
            // We can stop polling if the workspace resolution fails.
            stop()
        } else if (uiState === BatchSpecWorkspaceResolutionState.COMPLETED) {
            setError(undefined)
            setHasPreviewed(true)
            // We can stop polling once the workspace resolution completes.
            stop()
            // Fetch the results of the workspace preview resolution.
            // eslint-disable-next-line @typescript-eslint/no-floating-promises
            fetchWorkspaces()
            // eslint-disable-next-line @typescript-eslint/no-floating-promises
            fetchImportingChangesets()
            // Call the optional `onComplete` handler.
            onComplete?.()
        }
    }, [uiState, startPolling, stop, onComplete, fetchWorkspaces, fetchImportingChangesets])

    return {
        preview: previewBatchSpec,
        cancel,
        isInProgress,
        resolutionState: uiState,
        error,
        clearError: () => setError(undefined),
        hasPreviewed: hasRequestedPreview && hasPreviewed,
        noCache,
    }
}

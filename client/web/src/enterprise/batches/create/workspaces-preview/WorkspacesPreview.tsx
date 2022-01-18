import { ApolloError, WatchQueryFetchPolicy } from '@apollo/client'
import React, { useCallback, useEffect, useMemo, useState } from 'react'

import { useQuery } from '@sourcegraph/http-client'
import { ErrorAlert } from '@sourcegraph/web/src/components/alerts'

import {
    BatchSpecWorkspaceResolutionState,
    WorkspaceResolutionStatusVariables,
    WorkspaceResolutionStatusResult,
    EditBatchChangeFields,
} from '../../../../graphql-operations'
import { WORKSPACE_RESOLUTION_STATUS } from '../backend'

import { ImportingChangesetsPreviewList } from './ImportingChangesetsPreviewList'
import { PreviewLoadingSpinner } from './PreviewLoadingSpinner'
import { PreviewPrompt, PreviewPromptForm } from './PreviewPrompt'
import styles from './WorkspacesPreview.module.scss'
import { WorkspacesPreviewList } from './WorkspacesPreviewList'

interface WorkspacesPreviewProps {
    /** The existing, most recent batch spec for the batch change. */
    batchSpec: EditBatchChangeFields['currentSpec']
    /** Whether or not the user has previewed their batch spec at least once. */
    hasPreviewed: boolean
    /**
     * Whether or not the preview button should be disabled due to their being a problem
     * with the input batch spec YAML, or a preview request is already happening.
     */
    previewDisabled: boolean
    /**
     * Function to submit the current input batch spec YAML to trigger a workspaces
     * preview request.
     */
    preview: () => void
    /**
     * Whether or not the batch spec YAML on the server which was used to preview
     * workspaces is up-to-date with that which is presently in the editor.
     */
    batchSpecStale: boolean
    /**
     * Function to automatically update repo query of input batch spec YAML to exclude the
     * provided repo + branch.
     */
    excludeRepo: (repo: string, branch: string) => void
}

export const WorkspacesPreview: React.FunctionComponent<WorkspacesPreviewProps> = ({
    batchSpec,
    hasPreviewed,
    previewDisabled,
    preview,
    batchSpecStale,
    excludeRepo,
}) => {
    const [resolutionError, setResolutionError] = useState<string>()
    const [isResolvingPreview, setIsResolvingPreview] = useState(false)

    // We show a prompt for the user to trigger a new workspaces preview request (and
    // update the batch spec input YAML) if a preview isn't currently being resolved and
    // any of the following are true:
    // - They haven't previewed their batch spec workspaces at least once
    // - The preview workspaces resolution failed
    // - The batch spec YAML on the server is out of date with the one in the editor.
    const [showPreviewPrompt, previewPromptForm] = useMemo(() => {
        const showPreviewPrompt = !isResolvingPreview && (!hasPreviewed || resolutionError || batchSpecStale)
        const previewPromptForm: PreviewPromptForm = !hasPreviewed ? 'Initial' : resolutionError ? 'Error' : 'Update'

        return [showPreviewPrompt, previewPromptForm]
    }, [isResolvingPreview, hasPreviewed, batchSpecStale, resolutionError])

    const clearErrorAndPreview = useCallback(() => {
        setIsResolvingPreview(true)
        setResolutionError(undefined)
        preview()
    }, [preview])

    const onFinished = useCallback(() => setIsResolvingPreview(false), [])
    // Capture state changes when workspace resolution status changes.
    useBatchSpecWorkspaceResolution(batchSpec, { onError: setResolutionError, onFinished })

    return (
        <div className="d-flex flex-column align-items-center w-100 h-100">
            <h3 className={styles.header}>Workspaces preview</h3>
            {resolutionError && <ErrorAlert error={resolutionError} className="w-100 mb-3" />}
            {isResolvingPreview ? <PreviewLoadingSpinner className="mt-4" /> : null}
            {showPreviewPrompt && (
                <PreviewPrompt disabled={previewDisabled} preview={clearErrorAndPreview} form={previewPromptForm} />
            )}
            {hasPreviewed && !isResolvingPreview && (
                <div className="d-flex flex-column align-items-center overflow-auto w-100">
                    <WorkspacesPreviewList
                        batchSpecID={batchSpec.id}
                        isStale={batchSpecStale}
                        excludeRepo={excludeRepo}
                    />
                    <ImportingChangesetsPreviewList batchSpecID={batchSpec.id} isStale={batchSpecStale} />
                </div>
            )}
        </div>
    )
}

const POLLING_INTERVAL = 1000

type WorkspaceResolution = (WorkspaceResolutionStatusResult['node'] & {
    __typename: 'BatchSpec'
})['workspaceResolution']

const getResolution = (queryResult?: WorkspaceResolutionStatusResult): WorkspaceResolution =>
    queryResult?.node?.__typename === 'BatchSpec' ? queryResult.node.workspaceResolution : null

interface UseBatchSpecWorkspaceResolutionOptions {
    onError?: (error: string) => void
    onFinished?: () => void
    fetchPolicy?: WatchQueryFetchPolicy
}

export const useBatchSpecWorkspaceResolution = (
    batchSpec: EditBatchChangeFields['currentSpec'],
    { onError, onFinished, fetchPolicy = 'network-only' }: UseBatchSpecWorkspaceResolutionOptions = {}
): WorkspaceResolution => {
    const [isPolling, setIsPolling] = useState(false)

    const { data, refetch, startPolling, stopPolling } = useQuery<
        WorkspaceResolutionStatusResult,
        WorkspaceResolutionStatusVariables
    >(WORKSPACE_RESOLUTION_STATUS, {
        variables: { batchSpec: batchSpec.id },
        fetchPolicy,
        onError: error => onError?.(error.message),
    })

    // Re-query the workspace resolution status when an updated batch spec is created
    // (identified by `createdAt` changing).
    useEffect(() => {
        refetch().catch((error: ApolloError) => onError?.(error.message))
    }, [batchSpec.createdAt, refetch, onError])

    const resolution = useMemo(() => getResolution(data), [data])

    useEffect(() => {
        if (
            resolution?.state === BatchSpecWorkspaceResolutionState.QUEUED ||
            resolution?.state === BatchSpecWorkspaceResolutionState.PROCESSING
        ) {
            // If the workspace resolution is still queued or processing, start polling.
            startPolling(POLLING_INTERVAL)
            setIsPolling(true)
        } else if (
            resolution?.state === BatchSpecWorkspaceResolutionState.ERRORED ||
            resolution?.state === BatchSpecWorkspaceResolutionState.FAILED
        ) {
            // Report new workspace resolution worker errors back to the parent.
            onError?.(resolution.failureMessage || 'An unknown workspace resolution error occurred.')
            // We can stop polling if the workspace resolution fails.
            if (isPolling) {
                stopPolling()
                onFinished?.()
                setIsPolling(false)
            }
        } else if (resolution?.state === BatchSpecWorkspaceResolutionState.COMPLETED) {
            // We can stop polling once the workspace resolution is complete.
            if (isPolling) {
                stopPolling()
                onFinished?.()
                setIsPolling(false)
            }
        }
    }, [resolution, startPolling, stopPolling, isPolling, onError, onFinished])

    return resolution
}

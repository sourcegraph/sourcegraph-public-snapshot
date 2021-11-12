import { ApolloError } from '@apollo/client'
import React, { useCallback, useEffect, useMemo, useState } from 'react'

import { useQuery } from '@sourcegraph/shared/src/graphql/apollo'
import { ErrorAlert } from '@sourcegraph/web/src/components/alerts'
import { LoadingSpinner } from '@sourcegraph/wildcard'

import {
    BatchSpecWorkspaceResolutionState,
    Scalars,
    WorkspaceResolutionStatusVariables,
    WorkspaceResolutionStatusResult,
} from '../../../../graphql-operations'
import { WORKSPACE_RESOLUTION_STATUS } from '../backend'

import { PreviewPrompt, PreviewPromptForm } from './PreviewPrompt'
import styles from './WorkspacesPreview.module.scss'
import { WorkspacesPreviewList } from './WorkspacesPreviewList'

interface WorkspacesPreviewProps {
    batchSpecID?: Scalars['ID']
    currentPreviewRequestTime?: string
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
    batchSpecID,
    currentPreviewRequestTime,
    previewDisabled,
    preview,
    batchSpecStale,
    excludeRepo,
}) => {
    const [resolutionError, setResolutionError] = useState<string>()

    // We show a prompt for the user to trigger a new workspaces preview request (and
    // update the batch spec input YAML) if they haven't yet done so, if the preview
    // workspaces resolution failed, or if the batch spec YAML on the server is out of
    // date with the one in the editor.
    const [showPreviewPrompt, previewPromptForm] = useMemo(() => {
        const showPreviewPrompt = !batchSpecID || resolutionError || batchSpecStale
        const previewPromptForm: PreviewPromptForm = !batchSpecID ? 'Initial' : resolutionError ? 'Error' : 'Update'

        return [showPreviewPrompt, previewPromptForm]
    }, [batchSpecID, batchSpecStale, resolutionError])

    const clearErrorAndPreview = useCallback(() => {
        setResolutionError(undefined)
        preview()
    }, [preview])

    return (
        <>
            <h3 className={styles.header}>Workspaces preview</h3>
            {resolutionError && <ErrorAlert error={resolutionError} className="mb-3" />}
            {showPreviewPrompt && (
                <PreviewPrompt disabled={previewDisabled} preview={clearErrorAndPreview} form={previewPromptForm} />
            )}
            {batchSpecID && currentPreviewRequestTime && (
                <WithBatchSpec
                    batchSpecID={batchSpecID}
                    setResolutionError={setResolutionError}
                    excludeRepo={excludeRepo}
                    currentPreviewRequestTime={currentPreviewRequestTime}
                />
            )}
        </>
    )
}

const POLLING_INTERVAL = 1000

type WorkspaceResolutionStatus = (WorkspaceResolutionStatusResult['node'] & {
    __typename: 'BatchSpec'
})['workspaceResolution']

const getResolution = (queryResult?: WorkspaceResolutionStatusResult): WorkspaceResolutionStatus =>
    queryResult?.node?.__typename === 'BatchSpec' ? queryResult.node.workspaceResolution : null

interface WithBatchSpecProps
    extends Required<Pick<WorkspacesPreviewProps, 'batchSpecID' | 'excludeRepo' | 'currentPreviewRequestTime'>> {
    setResolutionError: (error: string) => void
}

const WithBatchSpec: React.FunctionComponent<WithBatchSpecProps> = ({
    batchSpecID,
    currentPreviewRequestTime,
    setResolutionError,
    excludeRepo,
    /**
     * Whether or not the workspaces previewed in the list are up-to-date with the batch
     * spec YAML that was last submitted for a preview.
     */
}) => {
    const { data, refetch, loading, startPolling, stopPolling } = useQuery<
        WorkspaceResolutionStatusResult,
        WorkspaceResolutionStatusVariables
    >(WORKSPACE_RESOLUTION_STATUS, {
        variables: { batchSpec: batchSpecID },
        // This data is intentionally transient, so there's no need to cache it.
        fetchPolicy: 'no-cache',
        // Report Apollo client errors back to the parent.
        onError: error => setResolutionError(error.message),
    })

    // Re-query the workspace resolution status when there's a new job requested.
    useEffect(() => {
        refetch().catch((error: ApolloError) => setResolutionError(error.message))
    }, [currentPreviewRequestTime, refetch, setResolutionError])

    useEffect(() => {
        const resolution = getResolution(data)
        if (
            resolution?.state === BatchSpecWorkspaceResolutionState.QUEUED ||
            resolution?.state === BatchSpecWorkspaceResolutionState.PROCESSING
        ) {
            // If the workspace resolution is still queued or processing, start polling.
            startPolling(POLLING_INTERVAL)
        } else if (
            resolution?.state === BatchSpecWorkspaceResolutionState.ERRORED ||
            resolution?.state === BatchSpecWorkspaceResolutionState.FAILED
        ) {
            // Report new workspace resolution worker errors back to the parent.
            setResolutionError(resolution.failureMessage || 'An unknown workspace resolution error occurred.')
        } else if (resolution?.state === BatchSpecWorkspaceResolutionState.COMPLETED) {
            // We can stop polling once the workspace resolution is complete.
            stopPolling()
        }
    }, [data, startPolling, stopPolling, setResolutionError])

    const resolution = getResolution(data)

    return (
        <>
            {loading || resolution?.state === 'QUEUED' || resolution?.state === 'PROCESSING' ? (
                // TODO: Show cooler loading indicator
                <LoadingSpinner className="my-4" />
            ) : null}
            {/* TODO: Keep stale workspaces list visible while we wait for the resolution. */}
            {resolution?.state === 'COMPLETED' ? (
                <div className="d-flex flex-column align-items-center overflow-auto w-100">
                    <WorkspacesPreviewList
                        batchSpecID={batchSpecID}
                        setResolutionError={setResolutionError}
                        excludeRepo={excludeRepo}
                    />
                </div>
            ) : null}
        </>
    )
}

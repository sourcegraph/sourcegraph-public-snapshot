import { ApolloError } from '@apollo/client'
import classNames from 'classnames'
import SearchIcon from 'mdi-react/SearchIcon'
import React, { useCallback, useEffect, useMemo, useState } from 'react'

import { CodeSnippet } from '@sourcegraph/branded/src/components/CodeSnippet'
import { useQuery } from '@sourcegraph/shared/src/graphql/apollo'
import { ErrorAlert } from '@sourcegraph/web/src/components/alerts'
import { Button, LoadingSpinner } from '@sourcegraph/wildcard'

import {
    BatchSpecWorkspaceResolutionState,
    Scalars,
    WorkspaceResolutionStatusVariables,
    WorkspaceResolutionStatusResult,
} from '../../../graphql-operations'

import { WORKSPACE_RESOLUTION_STATUS } from './backend'
import styles from './WorkspacesPreview.module.scss'
import { WorkspacesPreviewList } from './WorkspacesPreviewList'

interface WorkspacesPreviewProps {
    batchSpecID?: Scalars['ID']
    currentJobTime?: string
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
     * Whether or not the batch spec YAML on the server is up-to-date with that which is
     * presently in the editor.
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
    currentJobTime,
    previewDisabled,
    preview,
    batchSpecStale,
    excludeRepo,
}) => {
    const [resolutionError, setResolutionError] = useState<string>()

    const [showPreviewPrompt, previewPromptForm] = useMemo(() => {
        const showPreviewPrompt = !batchSpecID || batchSpecStale || resolutionError
        const previewPromptForm: PreviewPromptForm = !batchSpecID ? 'Initial' : resolutionError ? 'Error' : 'Update'

        return [showPreviewPrompt, previewPromptForm]
    }, [batchSpecID, batchSpecStale, resolutionError])

    const clearErrorAndPreview = useCallback(() => {
        setResolutionError(undefined)
        preview()
    }, [preview])

    return (
        <div className="h-100 d-flex flex-column align-items-center">
            <h3 className={styles.header}>Workspaces preview</h3>
            {resolutionError && <ErrorAlert error={resolutionError} className="mb-3" />}
            {showPreviewPrompt && (
                <PreviewPrompt disabled={previewDisabled} preview={clearErrorAndPreview} form={previewPromptForm} />
            )}
            {batchSpecID && currentJobTime && (
                <WithBatchSpec
                    batchSpecID={batchSpecID}
                    setResolutionError={setResolutionError}
                    excludeRepo={excludeRepo}
                    currentJobTime={currentJobTime}
                />
            )}
        </div>
    )
}

const ON_STATEMENT = `on:
  - repositoriesMatchingQuery: repo:my-org/.*
`

type PreviewPromptForm = 'Initial' | 'Error' | 'Update'

interface PreviewPromptProps {
    preview: () => void
    disabled: boolean
    form: PreviewPromptForm
}

const PreviewPrompt: React.FunctionComponent<PreviewPromptProps> = ({ preview, disabled, form }) => {
    const previewButton = (
        <Button variant="success" disabled={disabled} onClick={preview}>
            <SearchIcon className="icon-inline mr-1" />
            Preview workspaces
        </Button>
    )

    switch (form) {
        case 'Initial':
            return (
                <>
                    <div className={classNames(styles.previewPromptIcon, 'mt-4')} />
                    <h4 className={styles.previewPromptHeader}>
                        Use an <span className="text-monospace">on:</span> statement to preview repositories.
                    </h4>
                    {previewButton}
                    <div className={styles.previewPromptOnExample}>
                        <h3 className="align-self-start pt-4 text-muted">Example:</h3>
                        <CodeSnippet className="w-100" code={ON_STATEMENT} language="yaml" />
                    </div>
                </>
            )
        case 'Error':
            return previewButton
        case 'Update':
            return (
                <>
                    <h4 className={styles.previewPromptHeader}>
                        Finish editing your batch spec, then manually preview repositories.
                    </h4>
                    {previewButton}
                    <div className="mb-4" />
                </>
            )
    }
}

const POLLING_INTERVAL = 1000

type WorkspaceResolutionStatus = (WorkspaceResolutionStatusResult['node'] & {
    __typename: 'BatchSpec'
})['workspaceResolution']

const getResolution = (queryResult?: WorkspaceResolutionStatusResult): WorkspaceResolutionStatus =>
    queryResult?.node?.__typename === 'BatchSpec' ? queryResult.node.workspaceResolution : null

interface WithBatchSpecProps
    extends Required<Pick<WorkspacesPreviewProps, 'batchSpecID' | 'excludeRepo' | 'currentJobTime'>> {
    setResolutionError: (error: string) => void
}

const WithBatchSpec: React.FunctionComponent<WithBatchSpecProps> = ({
    batchSpecID,
    currentJobTime,
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

    // Requery the workspace resolution status when there's a new job requested.
    useEffect(() => {
        refetch().catch((error: ApolloError) => setResolutionError(error.message))
    }, [currentJobTime, refetch, setResolutionError])

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
            {resolution?.state === 'COMPLETED' ? (
                <WorkspacesPreviewList
                    batchSpecID={batchSpecID}
                    setResolutionError={setResolutionError}
                    excludeRepo={excludeRepo}
                />
            ) : null}
        </>
    )
}

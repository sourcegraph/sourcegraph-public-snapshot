import classNames from 'classnames'
import CloseIcon from 'mdi-react/CloseIcon'
import SearchIcon from 'mdi-react/SearchIcon'
import WarningIcon from 'mdi-react/WarningIcon'
import React, { useCallback, useEffect, useMemo, useState } from 'react'
import { useHistory, useLocation } from 'react-router'

import { CodeSnippet } from '@sourcegraph/branded/src/components/CodeSnippet'
import { LinkOrSpan } from '@sourcegraph/shared/src/components/LinkOrSpan'
import { BatchSpecWorkspaceResolutionState, Scalars } from '@sourcegraph/shared/src/graphql-operations'
import { useQuery } from '@sourcegraph/shared/src/graphql/apollo'
import { ErrorAlert } from '@sourcegraph/web/src/components/alerts'
import { Button, LoadingSpinner } from '@sourcegraph/wildcard'

import {
    BatchSpecWithWorkspacesFields,
    WorkspaceResolutionStatusResult,
    WorkspaceResolutionStatusVariables,
} from '../../../graphql-operations'

import { fetchBatchSpec, WORKSPACE_RESOLUTION_STATUS } from './backend'
import styles from './WorkspacesPreview.module.scss'

interface WorkspacesPreviewProps {
    batchSpecID?: Scalars['ID']
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
    // excludeRepo: (repo: string, branch: string) => void
    /** Whether or not the workspaces preview list is stale. */
    previewStale: boolean
}

export const WorkspacesPreview: React.FunctionComponent<WorkspacesPreviewProps> = ({
    batchSpecID,
    previewDisabled,
    preview,
    previewStale,
}) => {
    const [resolutionError, setResolutionError] = useState<string>()

    const [showPreviewPrompt, previewPromptForm] = useMemo(() => {
        const showPreviewPrompt = !batchSpecID || previewStale || resolutionError
        const previewPromptForm: PreviewPromptForm = !batchSpecID ? 'Initial' : resolutionError ? 'Error' : 'Update'

        return [showPreviewPrompt, previewPromptForm]
    }, [batchSpecID, previewStale, resolutionError])

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
            {batchSpecID && <WithBatchSpec batchSpecID={batchSpecID} setResolutionError={setResolutionError} />}
            {/*
            <ul className="list-group p-1 mb-0">
                {preview.workspaceResolution.workspaces.nodes.map(item => (
                    <li
                        className="d-flex border-bottom mb-3"
                        key={`${item.repository.id}_${item.branch.target.oid}_${item.path || '/'}`}
                    >
                        <button
                            className="btn align-self-start p-0 m-0 mr-3"
                            data-tooltip="Omit this repository from batch spec file"
                            type="button"
                            // TODO: Alert that for monorepos, we will exclude all paths
                            onClick={() => excludeRepo(item.repository.name, item.branch.displayName)}
                        >
                            <CloseIcon className="icon-inline" />
                        </button>
                        <div className="mb-2 flex-1">
                            <p>
                                {item.repository.name}:{item.branch.abbrevName} Path: {item.path || '/'}
                            </p>
                            <p>
                                {item.searchResultPaths.length} {pluralize('result', item.searchResultPaths.length)}
                            </p>
                        </div>
                    </li>
                ))}
            </ul>
            {preview.workspaceResolution.workspaces.nodes.length === 0 && (
                <span className="text-muted">No workspaces found</span>
            )}
            {preview.importingChangesets && preview.importingChangesets.totalCount > 0 && (
                <>
                    <h3>Importing changesets</h3>
                    <ul>
                        {preview.importingChangesets?.nodes.map(node => (
                            <li key={node.id}>
                                <LinkOrSpan
                                    to={
                                        node.__typename === 'VisibleChangesetSpec' &&
                                        node.description.__typename === 'ExistingChangesetReference'
                                            ? node.description.baseRepository.url
                                            : undefined
                                    }
                                >
                                    {node.__typename === 'VisibleChangesetSpec' &&
                                        node.description.__typename === 'ExistingChangesetReference' &&
                                        node.description.baseRepository.name}
                                </LinkOrSpan>{' '}
                                #
                                {node.__typename === 'VisibleChangesetSpec' &&
                                    node.description.__typename === 'ExistingChangesetReference' &&
                                    node.description.externalID}
                            </li>
                        ))}
                    </ul>
                </>
            )} */}
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
                </>
            )
    }
}

type WorkspaceResolutionStatus = (WorkspaceResolutionStatusResult['node'] & {
    __typename: 'BatchSpec'
})['workspaceResolution']

const getResolution = (queryResult?: WorkspaceResolutionStatusResult): WorkspaceResolutionStatus =>
    queryResult?.node?.__typename === 'BatchSpec' ? queryResult.node.workspaceResolution : null

const WithBatchSpec: React.FunctionComponent<{
    batchSpecID: Scalars['ID']
    setResolutionError: (error: string) => void
}> = ({ batchSpecID, setResolutionError }) => {
    const { data, loading, startPolling, stopPolling } = useQuery<
        WorkspaceResolutionStatusResult,
        WorkspaceResolutionStatusVariables
    >(WORKSPACE_RESOLUTION_STATUS, {
        variables: { batchSpec: batchSpecID },
        // This data is intentionally transient, so there's no need to cache it.
        fetchPolicy: 'no-cache',
        // Report new errors back to the parent.
        onCompleted: data => {
            const resolution = getResolution(data)
            if (
                resolution?.state === BatchSpecWorkspaceResolutionState.ERRORED ||
                resolution?.state === BatchSpecWorkspaceResolutionState.FAILED
            ) {
                setResolutionError(resolution.failureMessage || 'An unknown workspace resolution error occurred.')
            }
        },
        onError: error => setResolutionError(error.message),
    })

    const resolution = getResolution(data)

    return (
        <>
            {loading || resolution?.state === 'QUEUED' || resolution?.state === 'PROCESSING' ? (
                // TODO: Show cooler loading indicator
                <LoadingSpinner className="my-4" />
            ) : null}
            <Button onClick={() => startPolling(500)}>Start polling</Button>
            <Button onClick={() => stopPolling()}>Stop polling</Button>
        </>
    )
}

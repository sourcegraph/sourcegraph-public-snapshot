import classNames from 'classnames'
import CloseIcon from 'mdi-react/CloseIcon'
import SearchIcon from 'mdi-react/SearchIcon'
import WarningIcon from 'mdi-react/WarningIcon'
import React, { useCallback, useEffect, useMemo, useState } from 'react'
import { useHistory, useLocation } from 'react-router'

import { CodeSnippet } from '@sourcegraph/branded/src/components/CodeSnippet'
import { LinkOrSpan } from '@sourcegraph/shared/src/components/LinkOrSpan'
import { BatchSpecWorkspaceResolutionState, Scalars } from '@sourcegraph/shared/src/graphql-operations'
import { Button, LoadingSpinner } from '@sourcegraph/wildcard'

import { BatchSpecWithWorkspacesFields } from '../../../graphql-operations'

import { fetchBatchSpec } from './backend'
import styles from './WorkspacesPreview.module.scss'
import { hasOnStatement } from './yaml-util'

interface WorkspacesPreviewProps {
    batchSpecInput: string
    disabled: boolean
    preview: () => void
    // excludeRepo: (repo: string, branch: string) => void
    // preview: BatchSpecWithWorkspacesFields | Error | undefined
    // previewStale: boolean
}

export const WorkspacesPreview: React.FunctionComponent<WorkspacesPreviewProps> = ({
    batchSpecInput,
    disabled,
    preview,
}) => {
    const previewDisabled = useMemo(() => disabled || !hasOnStatement(batchSpecInput), [batchSpecInput, disabled])

    return (
        // if (!preview || previewStale) {
        //     return <LoadingSpinner />
        // }
        // if (isErrorLike(preview)) {
        //     return <ErrorAlert error={preview} className="mb-0" />
        // }
        // if (!preview.workspaceResolution) {
        //     throw new Error('Expected workspace resolution to exist.')
        // }
        <div className="h-100">
            <h3 className={styles.header}>Workspaces preview</h3>
            <Empty previewDisabled={previewDisabled} preview={preview} />
            {/* {preview.workspaceResolution.failureMessage !== null && (
                <ErrorAlert error={preview.workspaceResolution.failureMessage} />
            )}
            {preview.workspaceResolution.state === BatchSpecWorkspaceResolutionState.QUEUED && (
                <LoadingSpinner className="icon-inline" />
            )}
            {preview.workspaceResolution.state === BatchSpecWorkspaceResolutionState.PROCESSING && (
                <LoadingSpinner className="icon-inline" />
            )}
            {preview.workspaceResolution.state === BatchSpecWorkspaceResolutionState.ERRORED && (
                <WarningIcon className="text-danger icon-inline" />
            )}
            {preview.workspaceResolution.state === BatchSpecWorkspaceResolutionState.FAILED && (
                <WarningIcon className="text-danger icon-inline" />
            )}
            <p className="text-monospace">
                allowUnsupported: {JSON.stringify(preview.allowUnsupported)}
                <br />
                allowIgnored: {JSON.stringify(preview.allowIgnored)}
            </p>
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

const Empty: React.FunctionComponent<{ preview: () => void; previewDisabled: boolean }> = ({
    preview,
    previewDisabled,
}) => (
    <div className="d-flex flex-column align-items-center pt-4">
        <div className={styles.emptyIcon} />
        <h4 className={styles.emptyHeaderPrompt}>
            Use an <span className="text-monospace">on:</span> statement to preview repositories.
        </h4>
        <Button variant="success" disabled={previewDisabled} onClick={preview}>
            <SearchIcon className="icon-inline mr-1" />
            Preview workspaces
        </Button>
        <div className={styles.emptyOnExample}>
            <h3 className="align-self-start pt-4 text-muted">Example:</h3>
            <CodeSnippet className="w-100" code={ON_STATEMENT} language="yaml" />
        </div>
    </div>
)

import classNames from 'classnames'
import React, { useMemo } from 'react'

import { renderMarkdown } from '@sourcegraph/common'
import { Markdown } from '@sourcegraph/shared/src/components/Markdown'
import { Badge, Link } from '@sourcegraph/wildcard'

import { Timestamp } from '../../../components/time/Timestamp'
import {
    BatchChangeState,
    BatchSpecState,
    ListBatchChange,
    ListBatchChangeLatestSpecFields,
} from '../../../graphql-operations'
import {
    ChangesetStatusOpen,
    ChangesetStatusClosed,
    ChangesetStatusMerged,
} from '../detail/changesets/ChangesetStatusCell'

import styles from './BatchChangeNode.module.scss'
import { BatchChangeStatePill } from './BatchChangeStatePill'

export interface BatchChangeNodeProps {
    node: ListBatchChange
    isExecutionEnabled: boolean
    /** Used for testing purposes. Sets the current date */
    now?: () => Date
    displayNamespace: boolean
}

// This is the original, pre-SSBC version of the state badge. It has been superseded by
// `BatchChangeStatePill` and should be removed once SSBC is not longer behind a feature
// flag.
const StateBadge: React.FunctionComponent<{ state: BatchChangeState }> = ({ state }) => {
    switch (state) {
        case BatchChangeState.OPEN:
            return (
                <Badge variant="success" className={classNames(styles.batchChangeNodeBadge, 'text-uppercase')}>
                    Open
                </Badge>
            )
        case BatchChangeState.CLOSED:
            return (
                <Badge variant="danger" className={classNames(styles.batchChangeNodeBadge, 'text-uppercase')}>
                    Closed
                </Badge>
            )
        case BatchChangeState.DRAFT:
        default:
            return (
                <Badge variant="secondary" className={classNames(styles.batchChangeNodeBadge, 'text-uppercase')}>
                    Draft
                </Badge>
            )
    }
}

/**
 * An item in the list of batch changes.
 */
export const BatchChangeNode: React.FunctionComponent<BatchChangeNodeProps> = ({
    node,
    isExecutionEnabled,
    now = () => new Date(),
    displayNamespace,
}) => {
    const latestExecution: ListBatchChangeLatestSpecFields | undefined = useMemo(() => node.batchSpecs.nodes?.[0], [
        node.batchSpecs.nodes,
    ])

    // The URL to follow when a batch change is clicked on depends on the current state
    // and execution state.
    const nodeLink = useMemo(() => {
        // Before SSBC, all batch changes took you to the same place, the node detail
        // page. Closed batch changes also take you to this page.
        if (!isExecutionEnabled || node.state === BatchChangeState.CLOSED) {
            return node.url
        }

        const latestExecutionState = latestExecution?.state

        switch (latestExecutionState) {
            // If the latest spec hasn't been executed yet...
            case BatchSpecState.PENDING:
                // If it's a draft (no spec has been applied yet), we take you to the
                // editor page to continue working on it. Otherwise, we just take you to
                // the details page.
                return node.state === BatchChangeState.DRAFT ? `${node.url}/edit` : node.url
            // If the latest spec is in the middle of execution, or failed, we take you to
            // the execution details page.
            case BatchSpecState.QUEUED:
            case BatchSpecState.PROCESSING:
            case BatchSpecState.FAILED:
                return `${node.url}/executions/${latestExecution.id}`
            // If the latest spec finished execution successfully...
            case BatchSpecState.COMPLETED:
                // If it hasn't been applied, we take you to the preview page. Otherwise,
                // we just take you to the details page.
                return node.currentSpec.id === latestExecution.id
                    ? node.url
                    : `${node.url}/executions/${latestExecution.id}/preview`
            default:
                return node.url
        }
    }, [isExecutionEnabled, node.url, node.state, node.currentSpec, latestExecution])

    return (
        <>
            <span className={styles.batchChangeNodeSeparator} />
            {isExecutionEnabled ? (
                <BatchChangeStatePill
                    state={node.state}
                    latestExecutionState={node.batchSpecs.nodes[0]?.state}
                    currentSpecID={node.currentSpec.id}
                    latestSpecID={latestExecution?.id}
                    className={styles.batchChangeNodePill}
                />
            ) : (
                <StateBadge state={node.state} />
            )}
            <div className={styles.batchChangeNodeContent}>
                <div className="m-0 d-md-flex d-block align-items-baseline">
                    <h3 className={classNames(styles.batchChangeNodeTitle, 'm-0 d-md-inline-block d-block')}>
                        {displayNamespace && (
                            <div className="d-md-inline-block d-block">
                                <Link
                                    className="text-muted test-batches-namespace-link"
                                    to={`${node.namespace.url}/batch-changes`}
                                >
                                    {node.namespace.namespaceName}
                                </Link>
                                <span className="text-muted d-inline-block mx-1">/</span>
                            </div>
                        )}
                        <Link className="test-batches-link mr-2" to={nodeLink}>
                            {node.name}
                        </Link>
                    </h3>
                    <small className="text-muted d-sm-block">
                        created <Timestamp date={node.createdAt} now={now} />
                    </small>
                </div>
                <Markdown
                    className={classNames(
                        'text-truncate text-muted d-none d-md-block',
                        !node.description && 'font-italic'
                    )}
                    dangerousInnerHTML={
                        node.description ? renderMarkdown(node.description, { plainText: true }) : 'No description'
                    }
                />
            </div>
            {node.state !== BatchChangeState.DRAFT && (
                <>
                    <ChangesetStatusOpen
                        className="d-block d-sm-flex"
                        label={<span className="text-muted">{node.changesetsStats.open} open</span>}
                    />
                    <ChangesetStatusClosed
                        className="d-block d-sm-flex text-center"
                        label={<span className="text-muted">{node.changesetsStats.closed} closed</span>}
                    />
                    <ChangesetStatusMerged
                        className="d-block d-sm-flex"
                        label={<span className="text-muted">{node.changesetsStats.merged} merged</span>}
                    />
                </>
            )}
        </>
    )
}

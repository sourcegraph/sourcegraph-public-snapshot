import classNames from 'classnames'
import { upperFirst } from 'lodash'
import CancelIcon from 'mdi-react/CancelIcon'
import CheckCircleIcon from 'mdi-react/CheckCircleIcon'
import ChevronDownIcon from 'mdi-react/ChevronDownIcon'
import ChevronRightIcon from 'mdi-react/ChevronRightIcon'
import ErrorIcon from 'mdi-react/ErrorIcon'
import TimerSandIcon from 'mdi-react/TimerSandIcon'
import React, { useCallback, useMemo, useState } from 'react'

import { CodeSnippet } from '@sourcegraph/branded/src/components/CodeSnippet'
import { Link } from '@sourcegraph/shared/src/components/Link'
import { Timestamp } from '@sourcegraph/web/src/components/time/Timestamp'

import { BatchSpecExecutionsFields, BatchSpecExecutionState } from '../../../graphql-operations'

import { queryBatchSpecExecutions as _queryBatchSpecExecutions } from './backend'
import styles from './BatchSpecExecutionNode.module.scss'

export interface BatchSpecExecutionNodeProps {
    node: BatchSpecExecutionsFields
    /** Used for testing purposes. Sets the current date */
    now?: () => Date
}

export const BatchSpecExecutionNode: React.FunctionComponent<BatchSpecExecutionNodeProps> = ({
    node,
    now = () => new Date(),
}) => {
    const [isExpanded, setIsExpanded] = useState(false)
    const toggleIsExpanded = useCallback<React.MouseEventHandler<HTMLButtonElement>>(
        event => {
            event.preventDefault()
            setIsExpanded(!isExpanded)
        },
        [isExpanded]
    )

    const executionDuration = useMemo(() => {
        const endTime = node.finishedAt ? new Date(node.finishedAt).getTime() : now().getTime()
        return endTime - new Date(node.createdAt).getTime()
    }, [node.finishedAt, node.createdAt, now])

    return (
        <>
            <span className={styles.nodeSeparator} />
            <button
                type="button"
                className="btn btn-icon test-batches-expand-changeset d-none d-sm-block pb-1"
                aria-label={isExpanded ? 'Collapse section' : 'Expand section'}
                onClick={toggleIsExpanded}
            >
                {isExpanded ? (
                    <ChevronDownIcon className="icon-inline" aria-label="Close section" />
                ) : (
                    <ChevronRightIcon className="icon-inline" aria-label="Expand section" />
                )}
            </button>
            <div className="d-flex flex-column justify-content-center align-items-center px-2 pb-1">
                <ExecutionStateIcon state={node.state} />
                <span className="text-muted">{upperFirst(node.state.toLowerCase())}</span>
            </div>
            <div className="px-2 pb-1">
                <h3 className="pr-2">
                    <Link className="text-muted" to={`${node.namespace.url}/batch-changes`}>
                        {node.namespace.namespaceName}
                    </Link>
                    <span className="text-muted d-inline-block mx-1">/</span>
                    <Link to={`${node.namespace.url}/batch-changes/executions/${node.id}`}>{node.name || '-'}</Link>
                </h3>
                <small className="text-muted d-block">
                    Executed by <strong>{node.initiator.username}</strong> <Timestamp date={node.createdAt} now={now} />
                </small>
            </div>
            <div className="text-center pb-1">{(executionDuration / 1000).toFixed(0)}s</div>
            {isExpanded && (
                <div className={styles.nodeExpandedSection}>
                    <h4>Input spec</h4>
                    <CodeSnippet code={node.inputSpec} language="yaml" className="mb-0" />
                </div>
            )}
        </>
    )
}

const ExecutionStateIcon: React.FunctionComponent<{ state: BatchSpecExecutionState }> = ({ state }) => {
    switch (state) {
        case BatchSpecExecutionState.COMPLETED:
            return <CheckCircleIcon className={classNames(styles.nodeStateIcon, 'icon-inline text-success mb-1')} />

        case BatchSpecExecutionState.PROCESSING:
        case BatchSpecExecutionState.QUEUED:
            return <TimerSandIcon className={classNames(styles.nodeStateIcon, 'icon-inline text-muted mb-1')} />

        case BatchSpecExecutionState.CANCELED:
        case BatchSpecExecutionState.CANCELING:
            return <CancelIcon className={classNames(styles.nodeStateIcon, 'icon-inline text-muted mb-1')} />

        case BatchSpecExecutionState.ERRORED:
        case BatchSpecExecutionState.FAILED:
        default:
            return <ErrorIcon className={classNames(styles.nodeStateIcon, 'icon-inline text-danger mb-1')} />
    }
}

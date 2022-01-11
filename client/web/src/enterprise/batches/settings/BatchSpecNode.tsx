import classNames from 'classnames'
import { parseISO } from 'date-fns'
import { upperFirst } from 'lodash'
import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import CancelIcon from 'mdi-react/CancelIcon'
import CheckCircleIcon from 'mdi-react/CheckCircleIcon'
import ChevronDownIcon from 'mdi-react/ChevronDownIcon'
import ChevronRightIcon from 'mdi-react/ChevronRightIcon'
import TimerSandIcon from 'mdi-react/TimerSandIcon'
import React, { useCallback, useState } from 'react'

import { CodeSnippet } from '@sourcegraph/branded/src/components/CodeSnippet'
import { Link } from '@sourcegraph/shared/src/components/Link'
import { BatchSpecState } from '@sourcegraph/shared/src/graphql-operations'
import { Timestamp } from '@sourcegraph/web/src/components/time/Timestamp'
import { Button } from '@sourcegraph/wildcard'

import { BatchSpecListFields } from '../../../graphql-operations'

import styles from './BatchSpecNode.module.scss'

export interface BatchSpecNodeProps {
    node: BatchSpecListFields
    /** Used for testing purposes. Sets the current date */
    now?: () => Date
}

export const BatchSpecNode: React.FunctionComponent<BatchSpecNodeProps> = ({ node, now = () => new Date() }) => {
    const [isExpanded, setIsExpanded] = useState(false)
    const toggleIsExpanded = useCallback<React.MouseEventHandler<HTMLButtonElement>>(() => {
        setIsExpanded(!isExpanded)
    }, [isExpanded])

    return (
        <>
            <span className={styles.nodeSeparator} />
            <Button
                className="btn-icon"
                aria-label={isExpanded ? 'Collapse section' : 'Expand section'}
                onClick={toggleIsExpanded}
            >
                {isExpanded ? (
                    <ChevronDownIcon className="icon-inline" aria-label="Close section" />
                ) : (
                    <ChevronRightIcon className="icon-inline" aria-label="Expand section" />
                )}
            </Button>
            <div className="d-flex flex-column justify-content-center align-items-center px-2 pb-1">
                <StateIcon state={node.state} />
                <span className="text-muted">{upperFirst(node.state.toLowerCase())}</span>
            </div>
            <div className="px-2 pb-1">
                <h3 className="pr-2">
                    <Link className="text-muted" to={`${node.namespace.url}/batch-changes`}>
                        {node.namespace.namespaceName}
                    </Link>
                    <span className="text-muted d-inline-block mx-1">/</span>
                    <Link to={`/batch-changes/executions/${node.id}`}>{node.description.name || '-'}</Link>
                </h3>
                <small className="text-muted d-block">
                    Executed by <strong>{node.creator?.username}</strong> <Timestamp date={node.createdAt} now={now} />
                </small>
            </div>
            <div className="text-center pb-1">
                <Duration start={parseISO(node.createdAt)} end={node.finishedAt ? new Date(node.finishedAt) : now()} />
            </div>
            {isExpanded && (
                <div className={styles.nodeExpandedSection}>
                    <h4>Input spec</h4>
                    <CodeSnippet code={node.originalInput} language="yaml" className="mb-0" />
                </div>
            )}
        </>
    )
}

const StateIcon: React.FunctionComponent<{ state: BatchSpecState }> = ({ state }) => {
    switch (state) {
        case BatchSpecState.COMPLETED:
            return <CheckCircleIcon className={classNames(styles.nodeStateIcon, 'icon-inline text-success mb-1')} />

        case BatchSpecState.PROCESSING:
        case BatchSpecState.QUEUED:
            return <TimerSandIcon className={classNames(styles.nodeStateIcon, 'icon-inline text-muted mb-1')} />

        case BatchSpecState.CANCELED:
        case BatchSpecState.CANCELING:
            return <CancelIcon className={classNames(styles.nodeStateIcon, 'icon-inline text-muted mb-1')} />

        case BatchSpecState.FAILED:
        default:
            return <AlertCircleIcon className={classNames(styles.nodeStateIcon, 'icon-inline text-danger mb-1')} />
    }
}

const Duration: React.FunctionComponent<{ start: Date; end: Date }> = ({ start, end }) => {
    // The duration in seconds.
    let duration = (end.getTime() - start.getTime()) / 1000
    const hours = Math.floor(duration / (60 * 60))
    duration -= hours * 60 * 60
    const minutes = Math.floor(duration / 60)
    duration -= minutes * 60
    const seconds = Math.round(duration)
    return (
        <>
            {ensureTwoDigits(hours)}:{ensureTwoDigits(minutes)}:{ensureTwoDigits(seconds)}
        </>
    )
}

function ensureTwoDigits(value: number): string {
    return value < 10 ? `0${value}` : `${value}`
}

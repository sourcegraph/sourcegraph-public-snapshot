import React, { useCallback, useState } from 'react'

import classNames from 'classnames'
import { parseISO } from 'date-fns'
import { upperFirst } from 'lodash'
import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import CancelIcon from 'mdi-react/CancelIcon'
import CheckCircleIcon from 'mdi-react/CheckCircleIcon'
import ChevronDownIcon from 'mdi-react/ChevronDownIcon'
import ChevronRightIcon from 'mdi-react/ChevronRightIcon'
import StarIcon from 'mdi-react/StarIcon'
import TimerSandIcon from 'mdi-react/TimerSandIcon'

import { BatchSpecState } from '@sourcegraph/shared/src/graphql-operations'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { Button, Link, Icon, Typography } from '@sourcegraph/wildcard'

import { Timestamp } from '../../components/time/Timestamp'
import { BatchSpecListFields, Scalars } from '../../graphql-operations'

import { BatchSpec } from './BatchSpec'

import styles from './BatchSpecNode.module.scss'

export interface BatchSpecNodeProps extends ThemeProps {
    node: BatchSpecListFields
    currentSpecID?: Scalars['ID']
    /** Used for testing purposes. Sets the current date */
    now?: () => Date
}

export const BatchSpecNode: React.FunctionComponent<React.PropsWithChildren<BatchSpecNodeProps>> = ({
    node,
    currentSpecID,
    isLightTheme,
    now = () => new Date(),
}) => {
    const [isExpanded, setIsExpanded] = useState(currentSpecID === node.id)
    const toggleIsExpanded = useCallback<React.MouseEventHandler<HTMLButtonElement>>(() => {
        setIsExpanded(!isExpanded)
    }, [isExpanded])

    return (
        <>
            <span className={styles.nodeSeparator} />
            <Button
                variant="icon"
                aria-label={isExpanded ? 'Collapse section' : 'Expand section'}
                onClick={toggleIsExpanded}
            >
                {isExpanded ? (
                    <Icon aria-label="Close section" as={ChevronDownIcon} />
                ) : (
                    <Icon aria-label="Expand section" as={ChevronRightIcon} />
                )}
            </Button>
            <div className="d-flex flex-column justify-content-center align-items-center px-2 pb-1">
                <StateIcon state={node.state} />
                <span className="text-muted">{upperFirst(node.state.toLowerCase())}</span>
            </div>
            <div className="px-2 pb-1">
                <Typography.H3 className="pr-2">
                    {currentSpecID === node.id && (
                        <>
                            <Icon className="text-warning" data-tooltip="Currently applied spec" as={StarIcon} />{' '}
                        </>
                    )}
                    {currentSpecID && (
                        <Link to={`${node.namespace.url}/batch-changes/${node.description.name}/executions/${node.id}`}>
                            Executed by <strong>{node.creator?.username}</strong>{' '}
                            <Timestamp date={node.createdAt} now={now} />
                        </Link>
                    )}
                    {!currentSpecID && (
                        <>
                            <Link className="text-muted" to={`${node.namespace.url}/batch-changes`}>
                                {node.namespace.namespaceName}
                            </Link>
                            <span className="text-muted d-inline-block mx-1">/</span>
                            <Link
                                to={`${node.namespace.url}/batch-changes/${node.description.name}/executions/${node.id}`}
                            >
                                {node.description.name || '-'}
                            </Link>
                        </>
                    )}
                </Typography.H3>
                {!currentSpecID && (
                    <small className="text-muted d-block">
                        Executed by <strong>{node.creator?.username}</strong>{' '}
                        <Timestamp date={node.createdAt} now={now} />
                    </small>
                )}
            </div>
            <div className="text-center pb-1">
                <Duration start={parseISO(node.createdAt)} end={node.finishedAt ? new Date(node.finishedAt) : now()} />
            </div>
            {isExpanded && (
                <div className={styles.nodeExpandedSection}>
                    <Typography.H4>Input spec</Typography.H4>
                    <BatchSpec
                        isLightTheme={isLightTheme}
                        name={node.description.name}
                        originalInput={node.originalInput}
                        className={classNames(styles.batchSpec, 'mb-0')}
                    />
                </div>
            )}
        </>
    )
}

const StateIcon: React.FunctionComponent<React.PropsWithChildren<{ state: BatchSpecState }>> = ({ state }) => {
    switch (state) {
        case BatchSpecState.COMPLETED:
            return <Icon className={classNames(styles.nodeStateIcon, 'text-success mb-1')} as={CheckCircleIcon} />

        case BatchSpecState.PROCESSING:
        case BatchSpecState.QUEUED:
            return <Icon className={classNames(styles.nodeStateIcon, 'text-muted mb-1')} as={TimerSandIcon} />

        case BatchSpecState.CANCELED:
        case BatchSpecState.CANCELING:
            return <Icon className={classNames(styles.nodeStateIcon, 'text-muted mb-1')} as={CancelIcon} />

        case BatchSpecState.FAILED:
        default:
            return <Icon className={classNames(styles.nodeStateIcon, 'text-danger mb-1')} as={AlertCircleIcon} />
    }
}

const Duration: React.FunctionComponent<React.PropsWithChildren<{ start: Date; end: Date }>> = ({ start, end }) => {
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

import React, { useCallback, useState } from 'react'

import {
    mdiCheckCircle,
    mdiTimerSand,
    mdiCancel,
    mdiAlertCircle,
    mdiChevronDown,
    mdiChevronRight,
    mdiStar,
    mdiPencil,
} from '@mdi/js'
import classNames from 'classnames'
import { upperFirst } from 'lodash'

import { BatchSpecSource, BatchSpecState } from '@sourcegraph/shared/src/graphql-operations'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { Button, Link, Icon, H3, H4, Tooltip } from '@sourcegraph/wildcard'

import { Duration } from '../../components/time/Duration'
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
                <Icon aria-hidden={true} svgPath={isExpanded ? mdiChevronDown : mdiChevronRight} />
            </Button>
            <div className="d-flex flex-column justify-content-center align-items-center px-2 pb-1">
                <StateIcon source={node.source} state={node.state} />
                <span className="text-muted">
                    {node.source === BatchSpecSource.LOCAL && 'Uploaded'}
                    {node.source !== BatchSpecSource.LOCAL && upperFirst(node.state.toLowerCase())}
                </span>
            </div>
            <div className="px-2 pb-1">
                <H3 className="pr-2">
                    {currentSpecID && (
                        <Link to={`${node.namespace.url}/batch-changes/${node.description.name}/executions/${node.id}`}>
                            {currentSpecID === node.id && (
                                <>
                                    <Tooltip content="Currently applied spec">
                                        <Icon
                                            aria-label="Currently applied spec"
                                            className="text-warning"
                                            svgPath={mdiStar}
                                        />
                                    </Tooltip>{' '}
                                </>
                            )}
                            Created by <strong>{node.creator?.username}</strong>{' '}
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
                </H3>
                {!currentSpecID && (
                    <small className="text-muted d-block">
                        Created by <strong>{node.creator?.username}</strong>{' '}
                        <Timestamp date={node.createdAt} now={now} />
                    </small>
                )}
            </div>
            <div className="text-center pb-1">
                {node.startedAt && <Duration start={node.startedAt} end={node.finishedAt ?? undefined} />}
            </div>
            {isExpanded && (
                <div className={styles.nodeExpandedSection}>
                    <H4>Input spec</H4>
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

const StateIcon: React.FunctionComponent<
    React.PropsWithChildren<{ state: BatchSpecState; source: BatchSpecSource }>
> = ({ state, source }) => {
    if (source === BatchSpecSource.LOCAL) {
        return (
            <Icon
                aria-hidden={true}
                className={classNames(styles.nodeStateIcon, 'text-success mb-1')}
                svgPath={mdiCheckCircle}
            />
        )
    }
    switch (state) {
        case BatchSpecState.COMPLETED:
            return (
                <Icon
                    aria-hidden={true}
                    className={classNames(styles.nodeStateIcon, 'text-success mb-1')}
                    svgPath={mdiCheckCircle}
                />
            )

        case BatchSpecState.PROCESSING:
        case BatchSpecState.QUEUED:
            return (
                <Icon
                    aria-hidden={true}
                    className={classNames(styles.nodeStateIcon, 'text-muted mb-1')}
                    svgPath={mdiTimerSand}
                />
            )

        case BatchSpecState.CANCELED:
        case BatchSpecState.CANCELING:
            return (
                <Icon
                    aria-hidden={true}
                    className={classNames(styles.nodeStateIcon, 'text-muted mb-1')}
                    svgPath={mdiCancel}
                />
            )

        case BatchSpecState.FAILED:
            return (
                <Icon
                    aria-hidden={true}
                    className={classNames(styles.nodeStateIcon, 'text-danger mb-1')}
                    svgPath={mdiAlertCircle}
                />
            )
        case BatchSpecState.PENDING:
            return (
                <Icon
                    aria-hidden={true}
                    className={classNames(styles.nodeStateIcon, 'text-muted mb-1')}
                    svgPath={mdiPencil}
                />
            )
    }
}

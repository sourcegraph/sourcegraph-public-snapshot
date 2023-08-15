import React, { useCallback, useMemo, useState } from 'react'

import { mdiChevronDown, mdiChevronUp } from '@mdi/js'
import type { Subject } from 'rxjs'

import { Timestamp, TimestampFormat } from '@sourcegraph/branded/src/components/Timestamp'
import { Badge, type BadgeProps, Button, ErrorAlert, Icon } from '@sourcegraph/wildcard'

import { type ExternalServiceSyncJobListFields, ExternalServiceSyncJobState } from '../../graphql-operations'
import { ValueLegendList, type ValueLegendListProps } from '../../site-admin/analytics/components/ValueLegendList'
import { LoaderButton } from '../LoaderButton'
import { Duration } from '../time/Duration'

import { useCancelExternalServiceSync } from './backend'
import { EXTERNAL_SERVICE_SYNC_RUNNING_STATUSES } from './externalServices'

import styles from './ExternalServiceSyncJobNode.module.scss'

export interface ExternalServiceSyncJobNodeProps {
    node: ExternalServiceSyncJobListFields
    onUpdate: Subject<void>
}

const syncStateToBadgeVariant: Record<ExternalServiceSyncJobState, BadgeProps['variant']> = {
    [ExternalServiceSyncJobState.FAILED]: 'danger',
    [ExternalServiceSyncJobState.ERRORED]: 'warning',
    [ExternalServiceSyncJobState.COMPLETED]: 'success',
    [ExternalServiceSyncJobState.PROCESSING]: undefined,
    [ExternalServiceSyncJobState.QUEUED]: 'outlineSecondary',
    [ExternalServiceSyncJobState.CANCELED]: 'info',
    [ExternalServiceSyncJobState.CANCELING]: 'secondary',
}

export const ExternalServiceSyncJobNode: React.FunctionComponent<ExternalServiceSyncJobNodeProps> = ({
    node,
    onUpdate,
}) => {
    const [cancelExternalServiceSync, { error: cancelSyncJobError, loading: cancelSyncJobLoading }] =
        useCancelExternalServiceSync()

    const cancelJob = useCallback(
        () =>
            cancelExternalServiceSync({ variables: { id: node.id } }).then(() => {
                onUpdate.next()
                // Optimistically set state.
                node.state = ExternalServiceSyncJobState.CANCELING
            }),
        [cancelExternalServiceSync, node, onUpdate]
    )

    const legends = useMemo((): ValueLegendListProps['items'] | undefined => {
        if (!node) {
            return undefined
        }
        return [
            {
                value: node.reposAdded,
                description: 'Added',
                tooltip: 'The number of new repos discovered during this sync job.',
            },
            {
                value: node.reposDeleted,
                description: 'Deleted',
                tooltip: 'The number of repos deleted as a result of this sync job.',
            },
            {
                value: node.reposModified,
                description: 'Modified',
                tooltip: 'The number of existing repos whose metadata has changed during this sync job.',
            },
            {
                value: node.reposUnmodified,
                description: 'Unmodified',
                tooltip: 'The number of existing repos whose metadata did not change during this sync job.',
            },
            {
                value: node.reposSynced,
                description: 'Synced',
                color: 'var(--green)',
                tooltip: 'The number of repos synced during this sync job.',
                position: 'right',
            },
            {
                value: node.repoSyncErrors,
                description: 'Errors',
                color: 'var(--red)',
                tooltip: 'The number of times an error occurred syncing a repo during this sync job.',
                position: 'right',
            },
        ]
    }, [node])

    const [isExpanded, setIsExpanded] = useState(EXTERNAL_SERVICE_SYNC_RUNNING_STATUSES.has(node.state))
    const toggleIsExpanded = useCallback<React.MouseEventHandler<HTMLButtonElement>>(() => {
        setIsExpanded(!isExpanded)
    }, [isExpanded])

    return (
        <li className="list-group-item py-3">
            <div className="d-flex justify-content-left align-items-center">
                <div className="d-flex mr-2 justify-content-left">
                    <Button
                        variant="icon"
                        aria-label={isExpanded ? 'Collapse section' : 'Expand section'}
                        onClick={toggleIsExpanded}
                    >
                        <Icon aria-hidden={true} svgPath={isExpanded ? mdiChevronUp : mdiChevronDown} />
                    </Button>
                </div>
                <div className="d-flex mr-2 justify-content-left">
                    <Badge variant={syncStateToBadgeVariant[node.state]}>{node.state}</Badge>
                </div>
                <div className="flex-shrink-1 flex-grow-0 mr-1">
                    {node.startedAt === null && 'Not started yet.'}
                    {node.startedAt !== null && (
                        <>
                            Started at{' '}
                            <Timestamp
                                date={node.startedAt}
                                preferAbsolute={true}
                                timestampFormat={TimestampFormat.FULL_TIME}
                            />
                            .
                        </>
                    )}
                </div>
                <div className="flex-shrink-1 flex-grow-0 mr-1">
                    {node.finishedAt === null && 'Not finished yet.'}
                    {node.finishedAt !== null && (
                        <>
                            Finished at{' '}
                            <Timestamp
                                date={node.finishedAt}
                                preferAbsolute={true}
                                timestampFormat={TimestampFormat.FULL_TIME}
                            />
                            .
                        </>
                    )}
                </div>
                <div className="flex-shrink-0 flex-grow-1 mr-1">
                    {node.startedAt && (
                        <>
                            {node.finishedAt === null && <>Running for </>}
                            {node.finishedAt !== null && <>Ran for </>}
                            <Duration
                                start={node.startedAt}
                                end={node.finishedAt ?? undefined}
                                stableWidth={false}
                                className="d-inline"
                            />
                            {cancelSyncJobError && <ErrorAlert error={cancelSyncJobError} />}
                        </>
                    )}
                </div>
                {EXTERNAL_SERVICE_SYNC_RUNNING_STATUSES.has(node.state) && (
                    <LoaderButton
                        label="Cancel"
                        alwaysShowLabel={true}
                        variant="danger"
                        outline={true}
                        size="sm"
                        onClick={cancelJob}
                        loading={cancelSyncJobLoading || node.state === ExternalServiceSyncJobState.CANCELING}
                        disabled={cancelSyncJobLoading || node.state === ExternalServiceSyncJobState.CANCELING}
                        className={styles.cancelButton}
                    />
                )}
            </div>
            {isExpanded && legends && <ValueLegendList className="mb-0" items={legends} />}
            {isExpanded && node.failureMessage && (
                <ErrorAlert
                    error={`${node.failureMessage}\n\nWarning: Repositories will not be deleted if errors are present in the sync job.`}
                    className="mt-2 mb-0"
                />
            )}
        </li>
    )
}

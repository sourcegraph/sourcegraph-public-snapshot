import type { FunctionComponent } from 'react'

import { mdiAlert, mdiCheck, mdiTimerSand } from '@mdi/js'
import classNames from 'classnames'
import { formatDistance, parseISO } from 'date-fns'

import { Badge, Icon, Link, LoadingSpinner, Tooltip, useIsTruncated } from '@sourcegraph/wildcard'

import { type PreciseIndexFields, PreciseIndexState } from '../../../../graphql-operations'
import { INDEX_TERMINAL_STATES } from '../constants'

import { getIndexerKey } from './tree/util'

import styles from './IndexStateBadge.module.scss'

export interface IndexStateBadgeProps {
    indexes: PreciseIndexFields[]
    className?: string
}

const getIndexTerminalDate = (index: PreciseIndexFields): number => {
    // If we have the expected finished at date, use that.
    if (index.processingFinishedAt) {
        return new Date(index.uploadedAt ?? '').getDate()
    }

    // If we hit an error before processing, use the indexing finished at date.
    if (index.state === PreciseIndexState.INDEXING_ERRORED && index.indexingFinishedAt) {
        return new Date(index.indexingFinishedAt).getDate()
    }

    // Otherwise return old date to ensure index does not always take priority (this should never happen)
    return new Date(0).getDate()
}

export const IndexStateBadge: FunctionComponent<IndexStateBadgeProps> = ({ indexes, className }) => {
    const [ref, truncated, checkTruncation] = useIsTruncated<HTMLAnchorElement>()

    const mostRecentNonTerminalIndex = indexes
        .filter(index => !INDEX_TERMINAL_STATES.has(index.state))
        .sort((a, b) => new Date(a.uploadedAt ?? '').getDate() - new Date(b.uploadedAt ?? '').getDate())[0]

    const mostRecentTerminalIndex = indexes
        .filter(index => INDEX_TERMINAL_STATES.has(index.state))
        .sort((a, b) => getIndexTerminalDate(b) - getIndexTerminalDate(a))[0]

    // Prefer linking out to the most recent terminal index, if one exists.
    const preferredIndex = mostRecentTerminalIndex || mostRecentNonTerminalIndex

    if (!preferredIndex) {
        return null
    }

    const indexerKey = getIndexerKey(preferredIndex)

    return (
        <Tooltip content={truncated ? indexerKey : null}>
            <Badge
                as={Link}
                to={`../indexes/${preferredIndex.id}`}
                variant="outlineSecondary"
                className={className}
                ref={ref}
                onFocus={checkTruncation}
                onMouseEnter={checkTruncation}
            >
                {mostRecentTerminalIndex && (
                    <IndexStateBadgeIcon index={mostRecentTerminalIndex} className={styles.icon} />
                )}
                {mostRecentNonTerminalIndex && (
                    <IndexStateBadgeIcon index={mostRecentNonTerminalIndex} className={styles.icon} />
                )}
                {indexerKey}
            </Badge>
        </Tooltip>
    )
}

const formatDate = (date: string | null): string => {
    if (!date) {
        return ''
    }

    return ` ${formatDistance(parseISO(date), Date.now(), { addSuffix: true })}`
}

const getIndexStateTooltip = (index: PreciseIndexFields): string => {
    switch (index.state) {
        case PreciseIndexState.COMPLETED: {
            return `Indexing completed${formatDate(index.processingFinishedAt)}`
        }
        case PreciseIndexState.DELETED: {
            return 'Index deleted'
        }
        case PreciseIndexState.DELETING: {
            return 'Deleting index'
        }
        case PreciseIndexState.INDEXING: {
            return `Started indexing${formatDate(index.indexingStartedAt)}`
        }
        case PreciseIndexState.INDEXING_COMPLETED: {
            return `Finished indexing${formatDate(index.indexingFinishedAt)}`
        }
        case PreciseIndexState.INDEXING_ERRORED: {
            return `Indexing failed${formatDate(index.indexingFinishedAt)}`
        }
        case PreciseIndexState.PROCESSING: {
            return `Started processing index${formatDate(index.processingStartedAt)}`
        }
        case PreciseIndexState.PROCESSING_ERRORED: {
            return `Processing failed${formatDate(index.processingFinishedAt)}`
        }
        case PreciseIndexState.QUEUED_FOR_INDEXING: {
            return `Queued for indexing${formatDate(index.queuedAt)}`
        }
        case PreciseIndexState.QUEUED_FOR_PROCESSING: {
            return 'Queued for processing'
        }
        case PreciseIndexState.UPLOADING_INDEX: {
            return 'Uploading index'
        }
    }
}

interface IndexStateBadgeIconProps {
    index: PreciseIndexFields
    className?: string
}

const IndexStateBadgeIcon: FunctionComponent<IndexStateBadgeIconProps> = ({ index, className }) => {
    const label = getIndexStateTooltip(index)
    const ariaProps = label ? { 'aria-label': label } : ({ 'aria-hidden': true } as const)

    return (
        <Tooltip content={label}>
            {index.state === PreciseIndexState.COMPLETED ? (
                <Icon {...ariaProps} svgPath={mdiCheck} className={classNames('text-success', className)} />
            ) : index.state === PreciseIndexState.INDEXING ||
              index.state === PreciseIndexState.DELETING ||
              index.state === PreciseIndexState.PROCESSING ||
              index.state === PreciseIndexState.UPLOADING_INDEX ? (
                <LoadingSpinner {...ariaProps} className={className} />
            ) : index.state === PreciseIndexState.QUEUED_FOR_INDEXING ||
              index.state === PreciseIndexState.QUEUED_FOR_PROCESSING ? (
                <Icon {...ariaProps} svgPath={mdiTimerSand} className={className} />
            ) : index.state === PreciseIndexState.INDEXING_ERRORED ||
              index.state === PreciseIndexState.PROCESSING_ERRORED ? (
                <Icon {...ariaProps} svgPath={mdiAlert} className={classNames('text-danger', className)} />
            ) : (
                <></>
            )}
        </Tooltip>
    )
}

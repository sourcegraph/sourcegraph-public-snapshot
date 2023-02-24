import { FunctionComponent } from 'react'

import { mdiAlert, mdiCheck, mdiClose, mdiTimerSand } from '@mdi/js'
import classNames from 'classnames'

import { Badge, Icon, Link, LoadingSpinner, Tooltip } from '@sourcegraph/wildcard'

import { PreciseIndexFields, PreciseIndexState } from '../../../../graphql-operations'
import { INDEX_TERMINAL_STATES } from '../constants'

import { getIndexerKey } from './tree/util'

export interface IndexStateBadgeProps {
    indexes: PreciseIndexFields[]
    className?: string
}

export const IndexStateBadge: FunctionComponent<IndexStateBadgeProps> = ({ indexes, className }) => {
    const mostRecentNonTerminalIndex = indexes
        .filter(index => !INDEX_TERMINAL_STATES.has(index.state))
        // sort by descending uploaded at
        .sort((a, b) => new Date(a.uploadedAt ?? '').getDate() - new Date(b.uploadedAt ?? '').getDate())[0]

    const mostRecentTerminalIndex = indexes.find(index => INDEX_TERMINAL_STATES.has(index.state))

    // Prefer linking out to the most recent non-terminal index, if one exists.
    const preferredIndex = mostRecentTerminalIndex || mostRecentNonTerminalIndex

    if (!preferredIndex) {
        return null
    }

    return (
        <Badge as={Link} to={`../indexes/${preferredIndex.id}`} variant="outlineSecondary" className={className}>
            {mostRecentTerminalIndex && <IndexStateBadgeIcon state={mostRecentTerminalIndex.state} className="mr-1" />}
            {mostRecentNonTerminalIndex && (
                <IndexStateBadgeIcon state={mostRecentNonTerminalIndex.state} className="mr-1" />
            )}
            {getIndexerKey(preferredIndex)}
        </Badge>
    )
}

const preciseIndexStateTooltips: Partial<Record<PreciseIndexState, string>> = {
    [PreciseIndexState.COMPLETED]: 'Indexing completed successfully',

    [PreciseIndexState.INDEXING]: 'Currently indexing',
    [PreciseIndexState.PROCESSING]: 'Processing index',
    [PreciseIndexState.UPLOADING_INDEX]: 'Uploading index',

    [PreciseIndexState.QUEUED_FOR_INDEXING]: 'Queued for indexing',
    [PreciseIndexState.QUEUED_FOR_PROCESSING]: 'Queued for processing',

    [PreciseIndexState.INDEXING_ERRORED]: 'Indexing failed',
    [PreciseIndexState.PROCESSING_ERRORED]: 'Processing failed',
}

interface IndexStateBadgeIconProps {
    state: PreciseIndexState
    className?: string
}

const IndexStateBadgeIcon: FunctionComponent<IndexStateBadgeIconProps> = ({ state, className }) => {
    const label = preciseIndexStateTooltips[state]
    const ariaProps = label ? { 'aria-label': label } : ({ 'aria-hidden': true } as const)

    return (
        <Tooltip content={label}>
            {state === PreciseIndexState.COMPLETED ? (
                <Icon {...ariaProps} svgPath={mdiCheck} className={classNames('text-success', className)} />
            ) : state === PreciseIndexState.INDEXING ||
              state === PreciseIndexState.PROCESSING ||
              state === PreciseIndexState.UPLOADING_INDEX ? (
                <LoadingSpinner {...ariaProps} className={className} />
            ) : state === PreciseIndexState.QUEUED_FOR_INDEXING || state === PreciseIndexState.QUEUED_FOR_PROCESSING ? (
                <Icon {...ariaProps} svgPath={mdiTimerSand} className={className} />
            ) : state === PreciseIndexState.INDEXING_ERRORED || state === PreciseIndexState.PROCESSING_ERRORED ? (
                <Icon {...ariaProps} svgPath={mdiAlert} className={classNames('text-danger', className)} />
            ) : (
                <Icon {...ariaProps} svgPath={mdiClose} className={classNames('text-muted', className)} />
            )}
        </Tooltip>
    )
}

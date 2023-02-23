import { FunctionComponent } from 'react'

import { mdiCheck, mdiClose, mdiTimerSand } from '@mdi/js'
import classNames from 'classnames'

import { Badge, Icon, Link, LoadingSpinner, Tooltip } from '@sourcegraph/wildcard'

import { PreciseIndexFields, PreciseIndexState } from '../../../../graphql-operations'

import { getIndexerKey } from './tree/util'

import styles from './IndexStateBadge.module.scss'

const terminalStates = new Set<string>([
    PreciseIndexState.COMPLETED,
    PreciseIndexState.DELETED,
    PreciseIndexState.DELETING,
    PreciseIndexState.INDEXING_ERRORED,
    PreciseIndexState.PROCESSING_ERRORED,
])

export interface IndexStateBadgeProps {
    indexes: PreciseIndexFields[]
    className?: string
}

export const IndexStateBadge: FunctionComponent<IndexStateBadgeProps> = ({ indexes, className }) => {
    const mostRecentNonTerminalIndex = indexes
        .filter(index => !terminalStates.has(index.state))
        // sort by descending uploaded at
        .sort((a, b) => new Date(a.uploadedAt ?? '').getDate() - new Date(b.uploadedAt ?? '').getDate())[0]

    const mostRecentTerminalIndex = indexes.find(index => terminalStates.has(index.state))

    // Prefer linking out to the most recent non-terminal index, if one exists.
    const preferredIndex = mostRecentTerminalIndex || mostRecentNonTerminalIndex

    if (!preferredIndex) {
        return null
    }

    return (
        <Badge as={Link} to={`../indexes/${preferredIndex.id}`} variant="outlineSecondary" className={className}>
            {mostRecentTerminalIndex && <IndexStateBadgeIcon index={mostRecentTerminalIndex} className="mr-1" />}
            {mostRecentNonTerminalIndex && <IndexStateBadgeIcon index={mostRecentNonTerminalIndex} className="mr-1" />}
            {getIndexerKey(preferredIndex)}
        </Badge>
    )
}

const preciseIndexStateTooltips: Partial<Record<PreciseIndexState, string>> = {
    [PreciseIndexState.COMPLETED]: 'Indexing completed successfully',

    [PreciseIndexState.INDEXING]: 'Indexing completed successfully',
    [PreciseIndexState.PROCESSING]: 'Indexing completed successfully',
    [PreciseIndexState.UPLOADING_INDEX]: 'Indexing completed successfully',

    [PreciseIndexState.QUEUED_FOR_INDEXING]: 'Indexing completed successfully',
    [PreciseIndexState.QUEUED_FOR_PROCESSING]: 'Indexing completed successfully',

    [PreciseIndexState.INDEXING_ERRORED]: 'Indexing completed successfully',
    [PreciseIndexState.PROCESSING_ERRORED]: 'Indexing completed successfully',
}

interface IndexStateBadgeIconProps {
    index: PreciseIndexFields
    className?: string
}

const IndexStateBadgeIcon: FunctionComponent<IndexStateBadgeIconProps> = ({ index, className }) => {
    const { state } = index
    const label = preciseIndexStateTooltips[state]

    return (
        <Tooltip content={label}>
            <IndexStateIcon index={index} label={label} className={className} />
        </Tooltip>
    )
}

interface IndexStateIconProps {
    index: PreciseIndexFields
    label: string
    className?: string
}

const IndexStateIcon: FunctionComponent<IndexStateIconProps> = ({ index, label, className }) =>
    index.state === PreciseIndexState.COMPLETED ? (
        <Icon aria-label={label} svgPath={mdiCheck} className={classNames('text-success', className)} />
    ) : index.state === PreciseIndexState.INDEXING ||
      index.state === PreciseIndexState.PROCESSING ||
      index.state === PreciseIndexState.UPLOADING_INDEX ? (
        <LoadingSpinner aria-label={label} className={className} />
    ) : index.state === PreciseIndexState.QUEUED_FOR_INDEXING ||
      index.state === PreciseIndexState.QUEUED_FOR_PROCESSING ? (
        <Icon aria-label={label} svgPath={mdiTimerSand} className={className} />
    ) : index.state === PreciseIndexState.INDEXING_ERRORED || index.state === PreciseIndexState.PROCESSING_ERRORED ? (
        <Icon aria-label={label} svgPath={mdiClose} className={classNames('text-danger', className)} />
    ) : (
        <Icon aria-label={label} svgPath={mdiClose} className={classNames('text-muted', className)} />
    )

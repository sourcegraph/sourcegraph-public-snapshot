import { FunctionComponent } from 'react'

import { mdiCheck, mdiClose, mdiTimerSand } from '@mdi/js'
import classNames from 'classnames'

import { Icon, Link, LoadingSpinner } from '@sourcegraph/wildcard'

import { PreciseIndexFields, PreciseIndexState } from '../../../../graphql-operations'

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
}

export const IndexStateBadge: FunctionComponent<IndexStateBadgeProps> = ({ indexes }) => {
    const terminalIndexes = indexes.filter(index => terminalStates.has(index.state))
    const firstNonTerminalIndexes = indexes
        .filter(index => !terminalStates.has(index.state))
        // sort by descending uploaded at
        .sort((a, b) => new Date(a.uploadedAt ?? '').getDate() - new Date(b.uploadedAt ?? '').getDate())
        .slice(0, 1)

    // Only show one relevant terminal (assumed) index and one relevant non-terminal (explicit) index
    const collapsedIndexes = [...terminalIndexes, ...firstNonTerminalIndexes]

    return collapsedIndexes.length > 0 ? (
        <Link to={`./indexes/${collapsedIndexes[0].id}`}>
            <small className={classNames('float-right', 'ml-2', styles.hint)}>
                {collapsedIndexes.map(index => (
                    <IndexStateBadgeIcon index={index} key={index.id} />
                ))}
                {collapsedIndexes[0].indexer ? collapsedIndexes[0].indexer.key : collapsedIndexes[0].inputIndexer}
            </small>
        </Link>
    ) : (
        <></>
    )
}

interface IndexStateBadgeIconProps {
    index: PreciseIndexFields
}

const IndexStateBadgeIcon: FunctionComponent<IndexStateBadgeIconProps> = ({ index }) =>
    index.state === PreciseIndexState.COMPLETED ? (
        <Icon aria-hidden={true} svgPath={mdiCheck} className="text-success" />
    ) : index.state === PreciseIndexState.INDEXING ||
      index.state === PreciseIndexState.PROCESSING ||
      index.state === PreciseIndexState.UPLOADING_INDEX ? (
        <LoadingSpinner />
    ) : index.state === PreciseIndexState.QUEUED_FOR_INDEXING ||
      index.state === PreciseIndexState.QUEUED_FOR_PROCESSING ? (
        <Icon aria-hidden={true} svgPath={mdiTimerSand} />
    ) : index.state === PreciseIndexState.INDEXING_ERRORED || index.state === PreciseIndexState.PROCESSING_ERRORED ? (
        <Icon aria-hidden={true} svgPath={mdiClose} className="text-danger" />
    ) : (
        <Icon aria-hidden={true} svgPath={mdiClose} className="text-muted" />
    )

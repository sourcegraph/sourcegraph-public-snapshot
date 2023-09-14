import type { FunctionComponent } from 'react'

import classNames from 'classnames'

import { PreciseIndexState } from '../../../../graphql-operations'

export interface CodeIntelStateLabelProps {
    state: PreciseIndexState
    autoIndexed: boolean
    placeInQueue?: number | null
    className?: string
}

const labelClassName = 'text-muted text-center'

export const CodeIntelStateLabel: FunctionComponent<CodeIntelStateLabelProps> = ({
    state,
    autoIndexed,
    placeInQueue,
    className,
}) =>
    state === PreciseIndexState.QUEUED_FOR_PROCESSING || state === PreciseIndexState.QUEUED_FOR_INDEXING ? (
        <span className={classNames(labelClassName, className)}>
            Queued {placeInQueue ? <span className="d-block">(#{placeInQueue})</span> : <></>}
        </span>
    ) : state === PreciseIndexState.PROCESSING ? (
        <span className={classNames(labelClassName, className)}>Processing...</span>
    ) : state === PreciseIndexState.PROCESSING_ERRORED ? (
        <span className={classNames(labelClassName, className)}>Errored</span>
    ) : state === PreciseIndexState.COMPLETED ? (
        <span className={classNames(labelClassName, className)}>Completed</span>
    ) : state === PreciseIndexState.DELETED ? (
        <span className={classNames(labelClassName, className)}>Deleted</span>
    ) : state === PreciseIndexState.DELETING ? (
        <span className={classNames(labelClassName, className)}>Deleting</span>
    ) : state === PreciseIndexState.UPLOADING_INDEX ? (
        <span className={classNames(labelClassName, className)}>Uploading...</span>
    ) : state === PreciseIndexState.INDEXING ? (
        <span className={classNames(labelClassName, className)}>Indexing...</span>
    ) : state === PreciseIndexState.INDEXING_ERRORED ? (
        <span className={classNames(labelClassName, className)}>Errored</span>
    ) : state === PreciseIndexState.INDEXING_COMPLETED ? (
        <span className={classNames(labelClassName, className)}>completed</span>
    ) : autoIndexed ? (
        <span className={classNames(labelClassName, className)}>Completed</span>
    ) : (
        <></>
    )

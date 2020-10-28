import classNames from 'classnames'
import React, { FunctionComponent } from 'react'
import { LSIFIndexState, LSIFUploadState } from '../../../graphql-operations'

export interface CodeIntelStateLabelProps {
    state: LSIFUploadState | LSIFIndexState
    placeInQueue?: number | null
    className?: string
}

const labelClassNames = 'codeintel-state__label text-muted'

export const CodeIntelStateLabel: FunctionComponent<CodeIntelStateLabelProps> = ({ state, placeInQueue, className }) =>
    state === LSIFUploadState.UPLOADING ? (
        <span className={classNames(labelClassNames, className)}>Uploading</span>
    ) : state === LSIFUploadState.QUEUED || state === LSIFIndexState.QUEUED ? (
        <span className={classNames(labelClassNames, className)}>
            Queued <CodeIntelStateLabelPlaceInQueue placeInQueue={placeInQueue} />
        </span>
    ) : state === LSIFUploadState.PROCESSING || state === LSIFIndexState.PROCESSING ? (
        <span className={classNames(labelClassNames, className)}>Processing</span>
    ) : state === LSIFUploadState.COMPLETED || state === LSIFIndexState.COMPLETED ? (
        <span className={classNames(labelClassNames, className)}>Completed</span>
    ) : state === LSIFUploadState.ERRORED || state === LSIFIndexState.ERRORED ? (
        <span className={classNames(labelClassNames, className)}>Failed</span>
    ) : (
        <></>
    )

export interface CodeIntelStateLabelPlaceInQueueProps {
    placeInQueue?: number | null
}

const CodeIntelStateLabelPlaceInQueue: FunctionComponent<CodeIntelStateLabelPlaceInQueueProps> = ({ placeInQueue }) =>
    placeInQueue ? <span className="codeintel-state__label--block">(#{placeInQueue} in line)</span> : <></>

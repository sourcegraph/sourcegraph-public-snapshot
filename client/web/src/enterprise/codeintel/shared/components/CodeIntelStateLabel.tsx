import { FunctionComponent } from 'react'

import classNames from 'classnames'

import { LSIFIndexState, LSIFUploadState } from '../../../../graphql-operations'

import styles from './CodeIntelStateLabel.module.scss'

export interface CodeIntelStateLabelProps {
    state: LSIFUploadState | LSIFIndexState
    placeInQueue?: number | null
    className?: string
}

const labelClassNames = classNames(styles.label, 'text-muted')

export const CodeIntelStateLabel: FunctionComponent<React.PropsWithChildren<CodeIntelStateLabelProps>> = ({
    state,
    placeInQueue,
    className,
}) =>
    state === LSIFUploadState.UPLOADING ? (
        <span className={classNames(labelClassNames, className)}>Uploading</span>
    ) : state === LSIFUploadState.DELETING ? (
        <span className={classNames(labelClassNames, className)}>Deleting</span>
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

const CodeIntelStateLabelPlaceInQueue: FunctionComponent<
    React.PropsWithChildren<CodeIntelStateLabelPlaceInQueueProps>
> = ({ placeInQueue }) => (placeInQueue ? <span className={styles.block}>(#{placeInQueue} in line)</span> : <></>)

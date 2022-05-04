import { FunctionComponent } from 'react'

import classNames from 'classnames'
import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import CheckCircleIcon from 'mdi-react/CheckCircleIcon'
import FileUploadIcon from 'mdi-react/FileUploadIcon'
import TimerSandIcon from 'mdi-react/TimerSandIcon'

import { LoadingSpinner } from '@sourcegraph/wildcard'

import { LSIFIndexState, LSIFUploadState } from '../../../../graphql-operations'

export interface CodeIntelStateIconProps {
    state: LSIFUploadState | LSIFIndexState
    className?: string
}

export const CodeIntelStateIcon: FunctionComponent<React.PropsWithChildren<CodeIntelStateIconProps>> = ({
    state,
    className,
}) =>
    state === LSIFUploadState.UPLOADING ? (
        <FileUploadIcon className={className} />
    ) : state === LSIFUploadState.DELETING ? (
        <CheckCircleIcon className={classNames('text-muted', className)} />
    ) : state === LSIFUploadState.QUEUED || state === LSIFIndexState.QUEUED ? (
        <TimerSandIcon className={className} />
    ) : state === LSIFUploadState.PROCESSING || state === LSIFIndexState.PROCESSING ? (
        <LoadingSpinner inline={false} className={className} />
    ) : state === LSIFUploadState.COMPLETED || state === LSIFIndexState.COMPLETED ? (
        <CheckCircleIcon className={classNames('text-success', className)} />
    ) : state === LSIFUploadState.ERRORED || state === LSIFIndexState.ERRORED ? (
        <AlertCircleIcon className={classNames('text-danger', className)} />
    ) : (
        <></>
    )

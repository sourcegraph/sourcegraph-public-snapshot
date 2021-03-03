import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import classNames from 'classnames'
import CheckCircleIcon from 'mdi-react/CheckCircleIcon'
import ErrorIcon from 'mdi-react/ErrorIcon'
import FileUploadIcon from 'mdi-react/FileUploadIcon'
import TimerSandIcon from 'mdi-react/TimerSandIcon'
import React, { FunctionComponent } from 'react'
import { LSIFIndexState, LSIFUploadState } from '../../../graphql-operations'

export interface CodeIntelStateIconProps {
    state: LSIFUploadState | LSIFIndexState
    className?: string
}

export const CodeIntelStateIcon: FunctionComponent<CodeIntelStateIconProps> = ({ state, className }) =>
    state === LSIFUploadState.UPLOADING ? (
        <FileUploadIcon className={className} />
    ) : state === LSIFUploadState.QUEUED || state === LSIFIndexState.QUEUED ? (
        <TimerSandIcon className={className} />
    ) : state === LSIFUploadState.PROCESSING || state === LSIFIndexState.PROCESSING ? (
        <LoadingSpinner className={className} />
    ) : state === LSIFUploadState.COMPLETED || state === LSIFIndexState.COMPLETED ? (
        <CheckCircleIcon className={classNames('text-success', className)} />
    ) : state === LSIFUploadState.ERRORED || state === LSIFIndexState.ERRORED ? (
        <ErrorIcon className={classNames('text-danger', className)} />
    ) : (
        <></>
    )

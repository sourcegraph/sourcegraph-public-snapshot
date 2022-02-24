import classNames from 'classnames'
import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import CheckCircleIcon from 'mdi-react/CheckCircleIcon'
import FileUploadIcon from 'mdi-react/FileUploadIcon'
import TimerSandIcon from 'mdi-react/TimerSandIcon'
import React, { FunctionComponent } from 'react'

import { Icon, LoadingSpinner } from '@sourcegraph/wildcard'

import { LSIFIndexState, LSIFUploadState } from '../../../../graphql-operations'

export interface CodeIntelStateIconProps {
    state: LSIFUploadState | LSIFIndexState
    className?: string
}

export const CodeIntelStateIcon: FunctionComponent<CodeIntelStateIconProps> = ({ state, className }) =>
    state === LSIFUploadState.UPLOADING ? (
        <Icon as={FileUploadIcon} inline={false} className={className} />
    ) : state === LSIFUploadState.DELETING ? (
        <Icon as={CheckCircleIcon} inline={false} className={classNames('text-muted', className)} />
    ) : state === LSIFUploadState.QUEUED || state === LSIFIndexState.QUEUED ? (
        <Icon as={TimerSandIcon} inline={false} className={className} />
    ) : state === LSIFUploadState.PROCESSING || state === LSIFIndexState.PROCESSING ? (
        <Icon as={LoadingSpinner} inline={false} className={className} />
    ) : state === LSIFUploadState.COMPLETED || state === LSIFIndexState.COMPLETED ? (
        <Icon inline={false} as={CheckCircleIcon} className={classNames('text-success', className)} />
    ) : state === LSIFUploadState.ERRORED || state === LSIFIndexState.ERRORED ? (
        <Icon as={AlertCircleIcon} inline={false} className={classNames('text-danger', className)} />
    ) : (
        <></>
    )

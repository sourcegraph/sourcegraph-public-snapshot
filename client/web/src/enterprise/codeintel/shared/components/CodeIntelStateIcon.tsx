import { FunctionComponent } from 'react'

import classNames from 'classnames'

import { LoadingSpinner, Icon } from '@sourcegraph/wildcard'

import { LSIFIndexState, LSIFUploadState } from '../../../../graphql-operations'
import { mdiFileUpload, mdiCheckCircle, mdiTimerSand, mdiAlertCircle } from "@mdi/js";

export interface CodeIntelStateIconProps {
    state: LSIFUploadState | LSIFIndexState
    className?: string
}

export const CodeIntelStateIcon: FunctionComponent<React.PropsWithChildren<CodeIntelStateIconProps>> = ({
    state,
    className,
}) =>
    state === LSIFUploadState.UPLOADING ? (
        <Icon className={className} svgPath={mdiFileUpload} inline={false} aria-hidden={true} />
    ) : state === LSIFUploadState.DELETING ? (
        <Icon className={classNames('text-muted', className)} svgPath={mdiCheckCircle} inline={false} aria-hidden={true} />
    ) : state === LSIFUploadState.QUEUED || state === LSIFIndexState.QUEUED ? (
        <Icon className={className} svgPath={mdiTimerSand} inline={false} aria-hidden={true} />
    ) : state === LSIFUploadState.PROCESSING || state === LSIFIndexState.PROCESSING ? (
        <LoadingSpinner inline={false} className={className} />
    ) : state === LSIFUploadState.COMPLETED || state === LSIFIndexState.COMPLETED ? (
        <Icon className={classNames('text-success', className)} svgPath={mdiCheckCircle} inline={false} aria-hidden={true} />
    ) : state === LSIFUploadState.ERRORED || state === LSIFIndexState.ERRORED ? (
        <Icon className={classNames('text-danger', className)} svgPath={mdiAlertCircle} inline={false} aria-hidden={true} />
    ) : (
        <></>
    )

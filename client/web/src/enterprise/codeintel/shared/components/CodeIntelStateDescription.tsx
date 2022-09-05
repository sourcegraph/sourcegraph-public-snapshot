import { FunctionComponent } from 'react'

import { upperFirst } from 'lodash'

import { ErrorMessage } from '@sourcegraph/branded/src/components/alerts'
import { pluralize } from '@sourcegraph/common'

import { LSIFIndexState, LSIFUploadState } from '../../../../graphql-operations'

export interface CodeIntelStateDescriptionProps {
    typeName: string
    pluralTypeName?: string
    state: LSIFUploadState | LSIFIndexState
    placeInQueue?: number | null
    failure?: string | null
    className?: string
}

export const CodeIntelStateDescription: FunctionComponent<React.PropsWithChildren<CodeIntelStateDescriptionProps>> = ({
    typeName,
    pluralTypeName,
    state,
    placeInQueue,
    failure,
    className,
}) =>
    state === LSIFUploadState.UPLOADING ? (
        <span className={className}>Still uploading...</span>
    ) : state === LSIFUploadState.DELETING ? (
        <span className={className}>Upload is queued for deletion.</span>
    ) : state === LSIFUploadState.QUEUED || state === LSIFIndexState.QUEUED ? (
        <span className={className}>
            {upperFirst(typeName)} is queued.{' '}
            <CodeIntelStateDescriptionPlaceInQueue
                placeInQueue={placeInQueue}
                typeName={typeName}
                pluralTypeName={pluralTypeName}
            />
        </span>
    ) : state === LSIFUploadState.PROCESSING || state === LSIFIndexState.PROCESSING ? (
        <span className={className}>{upperFirst(typeName)} is currently being processed...</span>
    ) : state === LSIFUploadState.COMPLETED || state === LSIFIndexState.COMPLETED ? (
        <span className={className}>{upperFirst(typeName)} processed successfully.</span>
    ) : state === LSIFUploadState.ERRORED || state === LSIFIndexState.ERRORED ? (
        <span className={className}>
            {upperFirst(typeName)} failed to complete: <ErrorMessage error={failure} />
        </span>
    ) : (
        <></>
    )

export interface CodeIntelStateDescriptionPlaceInQueueProps {
    placeInQueue?: number | null
    typeName: string
    pluralTypeName?: string
}

const CodeIntelStateDescriptionPlaceInQueue: FunctionComponent<
    React.PropsWithChildren<CodeIntelStateDescriptionPlaceInQueueProps>
> = ({ placeInQueue, typeName, pluralTypeName }) => {
    if (placeInQueue === 1) {
        return <>This {typeName} is up next for processing.</>
    }
    return (
        <>
            {placeInQueue
                ? `There are ${placeInQueue - 1} ${
                      pluralTypeName !== undefined ? pluralTypeName : pluralize(typeName, placeInQueue - 1)
                  } ahead of this one.`
                : ''}
        </>
    )
}

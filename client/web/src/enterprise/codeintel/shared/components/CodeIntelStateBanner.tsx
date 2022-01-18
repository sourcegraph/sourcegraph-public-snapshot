import React, { FunctionComponent } from 'react'

import { Alert } from '@sourcegraph/wildcard'

import { LSIFIndexState, LSIFUploadState } from '../../../../graphql-operations'

import { CodeIntelStateDescription } from './CodeIntelStateDescription'

export interface CodeIntelStateBannerProps {
    typeName: string
    pluralTypeName: string
    state: LSIFUploadState | LSIFIndexState
    placeInQueue?: number | null
    failure?: string | null
    className?: string
}
// TODO: Need to fix to use variants instread of className
export const CodeIntelStateBanner: FunctionComponent<CodeIntelStateBannerProps> = ({
    typeName,
    pluralTypeName,
    state,
    placeInQueue,
    failure,
    className = 'alert-primary',
}) => (
    <Alert className={className}>
        <span>
            <CodeIntelStateDescription
                state={state}
                placeInQueue={placeInQueue}
                failure={failure}
                typeName={typeName}
                pluralTypeName={pluralTypeName}
            />
        </span>
    </Alert>
)

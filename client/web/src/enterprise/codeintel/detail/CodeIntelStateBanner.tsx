import * as H from 'history'
import React, { FunctionComponent } from 'react'
import { LSIFIndexState, LSIFUploadState } from '../../../graphql-operations'
import { CodeIntelStateDescription } from '../shared/CodeIntelStateDescription'
import { CodeIntelStateIcon } from '../shared/CodeIntelStateIcon'

export interface CodeIntelStateBannerProps {
    typeName: string
    pluralTypeName: string
    state: LSIFUploadState | LSIFIndexState
    placeInQueue?: number | null
    failure?: string | null
    className?: string
    history: H.History
}

export const CodeIntelStateBanner: FunctionComponent<CodeIntelStateBannerProps> = ({
    typeName,
    pluralTypeName,
    state,
    placeInQueue,
    failure,
    className = 'alert-primary',
    history,
}) => (
    <div className={`alert mb-4 mt-3 ${className}`}>
        <span className="icon-inline">
            <CodeIntelStateIcon state={state} />
        </span>
        <span className="ml-2">
            <CodeIntelStateDescription
                state={state}
                placeInQueue={placeInQueue}
                failure={failure}
                typeName={typeName}
                pluralTypeName={pluralTypeName}
                history={history}
            />
        </span>
    </div>
)

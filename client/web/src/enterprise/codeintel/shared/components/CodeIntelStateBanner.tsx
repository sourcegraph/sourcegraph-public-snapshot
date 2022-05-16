import { FunctionComponent } from 'react'

import { Alert, AlertProps } from '@sourcegraph/wildcard'

import { LSIFIndexState, LSIFUploadState } from '../../../../graphql-operations'

import { CodeIntelStateDescription } from './CodeIntelStateDescription'

export interface CodeIntelStateBannerProps {
    typeName: string
    pluralTypeName: string
    state: LSIFUploadState | LSIFIndexState
    placeInQueue?: number | null
    failure?: string | null
    variant?: AlertProps['variant']
}

export const CodeIntelStateBanner: FunctionComponent<React.PropsWithChildren<CodeIntelStateBannerProps>> = ({
    typeName,
    pluralTypeName,
    state,
    placeInQueue,
    failure,
    variant = 'primary',
}) => (
    <Alert variant={variant}>
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

import React from 'react'

import { Typography } from '@sourcegraph/wildcard'

import { defaultExternalServices } from '../../../components/externalServices/externalServices'
import { ExternalServiceKind } from '../../../graphql-operations'

export interface ModalHeaderProps {
    id: string
    externalServiceKind: ExternalServiceKind
    externalServiceURL: string
}

export const ModalHeader: React.FunctionComponent<React.PropsWithChildren<ModalHeaderProps>> = ({
    id,
    externalServiceKind,
    externalServiceURL,
}) => (
    <>
        <Typography.H3 id={id}>
            Batch Changes credentials: {defaultExternalServices[externalServiceKind].defaultDisplayName}
        </Typography.H3>
        <p className="mb-4">{externalServiceURL}</p>
    </>
)

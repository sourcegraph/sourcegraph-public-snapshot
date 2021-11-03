import React from 'react'

import { defaultExternalServices } from '@sourcegraph/web/src/components/externalServices/externalServices'

import { ExternalServiceKind } from '../../graphql-operations'

export interface Props {
    externalServiceURL: string
    externalServiceKind: ExternalServiceKind
}

export const CodeHost: React.FunctionComponent<Props> = ({ externalServiceURL, externalServiceKind }) => {
    const Icon = defaultExternalServices[externalServiceKind].icon
    return (
        <li>
            <Icon className="icon-inline mr-2" />
            {externalServiceURL}
        </li>
    )
}

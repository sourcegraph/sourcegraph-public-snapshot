import React from 'react'

import { defaultExternalServices } from '@sourcegraph/web/src/components/externalServices/externalServices'
import { Icon } from '@sourcegraph/wildcard'

import { ExternalServiceKind } from '../../graphql-operations'

export interface Props {
    externalServiceURL: string
    externalServiceKind: ExternalServiceKind
}

export const CodeHost: React.FunctionComponent<Props> = ({ externalServiceURL, externalServiceKind }) => {
    const ExternalServiceIcon = defaultExternalServices[externalServiceKind].icon
    return (
        <li>
            <Icon className="mr-2" as={ExternalServiceIcon} />
            {externalServiceURL}
        </li>
    )
}

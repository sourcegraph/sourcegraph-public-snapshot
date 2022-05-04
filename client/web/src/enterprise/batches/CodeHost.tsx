import React from 'react'

import { Icon } from '@sourcegraph/wildcard'

import { defaultExternalServices } from '../../components/externalServices/externalServices'
import { ExternalServiceKind } from '../../graphql-operations'

export interface Props {
    externalServiceURL: string
    externalServiceKind: ExternalServiceKind
}

export const CodeHost: React.FunctionComponent<React.PropsWithChildren<Props>> = ({
    externalServiceURL,
    externalServiceKind,
}) => {
    const ExternalServiceIcon = defaultExternalServices[externalServiceKind].icon
    return (
        <li>
            <Icon className="mr-2" as={ExternalServiceIcon} />
            {externalServiceURL}
        </li>
    )
}

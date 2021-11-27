import React from 'react'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { CatalogComponentDetailFields } from '../../../../../graphql-operations'

interface Props extends TelemetryProps {
    catalogComponent: CatalogComponentDetailFields
}

export const ComponentDetailContent: React.FunctionComponent<Props> = ({ catalogComponent }) => (
    <div>
        <h1>{catalogComponent.name}</h1>
    </div>
)

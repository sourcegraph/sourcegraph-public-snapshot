import React from 'react'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { CatalogComponentDetailFields } from '../../../../../graphql-operations'
import { CatalogComponentIcon } from '../../../components/CatalogComponentIcon'

interface Props extends TelemetryProps {
    catalogComponent: CatalogComponentDetailFields
}

export const ComponentDetailContent: React.FunctionComponent<Props> = ({ catalogComponent }) => (
    <div>
        <header>
            <h1>
                <CatalogComponentIcon catalogComponent={catalogComponent} className="icon-inline mr-1" />{' '}
                {catalogComponent.name}
            </h1>
            <ul className="list-unstyled">
                <li>
                    <strong>Owner</strong> alice
                </li>
                <li>
                    <strong>Lifecycle</strong> production
                </li>
            </ul>
        </header>
    </div>
)

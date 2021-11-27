import React, { useContext, useMemo } from 'react'

import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { useObservable } from '@sourcegraph/shared/src/util/useObservable'

import { CatalogBackendContext } from '../../../../core/backend/context'
import { CatalogComponentFiltersProps } from '../../../../core/component-filters'

export interface OverviewContentProps extends TelemetryProps, Pick<CatalogComponentFiltersProps, 'filters'> {
    // TODO(sqs): what scope of catalog (eg repo) or global
}

export const OverviewContent: React.FunctionComponent<OverviewContentProps> = ({ filters }) => {
    const { listComponents } = useContext(CatalogBackendContext)
    const components = useObservable(
        useMemo(() => listComponents({ query: filters.query }), [filters.query, listComponents])
    )

    if (components === undefined) {
        return <LoadingSpinner />
    }

    return (
        <div>
            <section className="d-flex flex-wrap align-items-center">
                <ul>
                    {components.nodes.map(node => (
                        <li key={node.id}>{node.name}</li>
                    ))}
                </ul>
            </section>
        </div>
    )
}

import React from 'react'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { CatalogComponentFiltersProps } from '../../../../core/component-filters'

export interface OverviewContentProps extends TelemetryProps, CatalogComponentFiltersProps {
    // TODO(sqs): what scope of catalog (eg repo) or global
}

export const OverviewContent: React.FunctionComponent<OverviewContentProps> = ({ filters, onFiltersChange }) => (
    /* const { listComponents } = useContext(CatalogBackendContext)
    const components = useObservable(
        useMemo(() => listComponents({ query: filters.query }), [filters.query, listComponents])
    )

    if (components === undefined) {
        return <LoadingSpinner />
    }
    */

    <>asdf{/* <ComponentList filters={filters} onFiltersChange={onFiltersChange} className="flex-1" /> */}</>
)

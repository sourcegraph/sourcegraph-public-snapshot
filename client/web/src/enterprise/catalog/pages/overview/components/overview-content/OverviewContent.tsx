import React from 'react'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { CatalogComponentFiltersProps } from '../../../../core/component-filters'
import { ComponentList } from '../component-list/ComponentList'

export interface OverviewContentProps extends TelemetryProps, Pick<CatalogComponentFiltersProps, 'filters'> {
    // TODO(sqs): what scope of catalog (eg repo) or global
}

export const OverviewContent: React.FunctionComponent<OverviewContentProps> = ({ filters }) => (
    /* const { listComponents } = useContext(CatalogBackendContext)
    const components = useObservable(
        useMemo(() => listComponents({ query: filters.query }), [filters.query, listComponents])
    )

    if (components === undefined) {
        return <LoadingSpinner />
    }
    */

    <ComponentList filters={filters} className="flex-1" />
)

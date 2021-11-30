import React, { useEffect } from 'react'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Page } from '@sourcegraph/web/src/components/Page'
import { Container, PageHeader } from '@sourcegraph/wildcard'

import { CatalogIcon } from '../../../../../catalog'
import { CatalogEntityFiltersProps } from '../../../core/entity-filters'
import { OverviewContent } from '../components/overview-content/OverviewContent'

interface Props extends CatalogEntityFiltersProps, TelemetryProps {}

/**
 * The catalog overview page.
 */
export const OverviewPage: React.FunctionComponent<Props> = ({ filters, onFiltersChange, telemetryService }) => {
    useEffect(() => {
        telemetryService.logViewEvent('CatalogOverview')
    }, [telemetryService])

    return (
        <Page>
            <PageHeader
                path={[{ icon: CatalogIcon, text: 'Catalog' }]}
                className="mb-4"
                description="Explore software components, services, libraries, APIs, and more."
            />

            <OverviewContent filters={filters} onFiltersChange={onFiltersChange} telemetryService={telemetryService} />
        </Page>
    )
}

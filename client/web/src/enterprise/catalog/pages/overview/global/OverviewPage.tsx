import React, { useEffect } from 'react'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Page } from '@sourcegraph/web/src/components/Page'
import { Container, PageHeader } from '@sourcegraph/wildcard'

import { CatalogIcon } from '../../../../../catalog'
import { CatalogComponentFiltersProps } from '../../../core/component-filters'
import { OverviewContent } from '../components/overview-content/OverviewContent'

interface Props extends CatalogComponentFiltersProps, TelemetryProps {}

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
                className="mb-3"
                description="Explore software components, services, libraries, APIs, and more."
            />
            <Container className="mb-4 p-0">
                <OverviewContent
                    filters={filters}
                    onFiltersChange={onFiltersChange}
                    telemetryService={telemetryService}
                />
            </Container>
        </Page>
    )
}

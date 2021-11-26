import React, { useEffect } from 'react'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Page } from '@sourcegraph/web/src/components/Page'
import { PageHeader } from '@sourcegraph/wildcard'

import { CatalogIcon } from '../../../../../catalog'
import { FeedbackBadge } from '../../../../../components/FeedbackBadge'
import { CatalogExplorer } from '../components/catalog-explorer/CatalogExplorer'

interface Props extends TelemetryProps {}

/**
 * The catalog overview page.
 */
export const ExplorePage: React.FunctionComponent<Props> = ({ telemetryService }) => {
    useEffect(() => {
        telemetryService.logViewEvent('CatalogExplore')
    }, [telemetryService])

    return (
        <Page>
            <PageHeader
                path={[{ icon: CatalogIcon, text: 'Catalog' }]}
                className="mb-4"
                description="Explore software components, services, libraries, APIs, and more."
                actions={<FeedbackBadge status="prototype" feedback={{ mailto: 'support@sourcegraph.com' }} />}
            />
            <CatalogExplorer />
        </Page>
    )
}

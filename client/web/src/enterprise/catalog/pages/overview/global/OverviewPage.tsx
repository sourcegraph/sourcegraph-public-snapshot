import React, { useEffect } from 'react'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { AuthenticatedUser } from '../../../../../auth'
import { useCatalogComponentFilters } from '../../../core/component-filters'
import { OverviewContent } from '../components/overview-content/OverviewContent'
import { Sidebar } from '../components/sidebar/Sidebar'

import styles from './OverviewPage.module.scss'

// TODO(sqs): extract the Insights components used above to the shared components area

export interface OverviewPageProps extends TelemetryProps {
    authenticatedUser: AuthenticatedUser
}

/**
 * The catalog overview page.
 */
export const OverviewPage: React.FunctionComponent<OverviewPageProps> = ({ telemetryService }) => {
    useEffect(() => {
        telemetryService.logViewEvent('CatalogOverview')
    }, [telemetryService])

    const { filters, onFiltersChange } = useCatalogComponentFilters()

    return (
        <div className={styles.container}>
            <Sidebar filters={filters} onFiltersChange={onFiltersChange} />
            <OverviewContent filters={filters} telemetryService={telemetryService} />
        </div>
    )
}

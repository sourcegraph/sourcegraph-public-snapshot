import React, { useEffect } from 'react'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { PageHeader } from '@sourcegraph/wildcard'

import { AuthenticatedUser } from '../../../../../auth'
import { CatalogComponentFiltersProps } from '../../../core/component-filters'

import { Sidebar } from '../../overview/components/sidebar/Sidebar'
import { ComponentDetailContent } from './ComponentDetailContent'

export interface OverviewPageProps extends CatalogComponentFiltersProps, TelemetryProps {
    authenticatedUser: AuthenticatedUser
}

/**
 * The catalog component detail page.
 */
export const ComponentDetailPage: React.FunctionComponent<OverviewPageProps> = ({
    filters,
    onFiltersChange,
    telemetryService,
}) => {
    useEffect(() => {
        telemetryService.logViewEvent('CatalogComponentDetail')
    }, [telemetryService])

    return (
        <>
            <Sidebar filters={filters} onFiltersChange={onFiltersChange} />
            <ComponentDetailContent telemetryService={telemetryService} />
        </>
    )
}

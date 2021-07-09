import React, { ChangeEvent, useState } from 'react';

import { isOrganizationDashboard, RealInsightDashboard } from '../../../../core/types';
import { isBuiltInDashboard, isSettingsBasedInsightsDashboard } from '../../../../core/types/dashboard/real-dashboard';
import { createDashboardsMap } from '../../utils';

interface InsightOrganizationVisibilityProps {
    /**
     * All accessible dashboards from settings cascade
     */
    dashboards: RealInsightDashboard[]

    /**
     * Set of dashboards that have particular insight in their insightIds
     */
    value: Record<string, RealInsightDashboard>

    /**
     * Change handlers fires whenever the user changes visibility settings.
     */
    onChange: (dashboards: Record<string, RealInsightDashboard>) => void

    className?: string
}

/**
 * Displays an insight organization visibility settings.
 */
export const InsightOrganizationVisibility: React.FunctionComponent<InsightOrganizationVisibilityProps> = props => {
    const { dashboards, value, onChange, className } = props

    // Get all built in org based dashboards that actually in terms of dashboards
    // means an organization
    const organizations = dashboards.filter(
        dashboard => isOrganizationDashboard(dashboard) && isBuiltInDashboard(dashboard)
    )

    // Get all custom organization dashboards
    const organizationsDashboards = dashboards.filter(
        dashboard => isOrganizationDashboard(dashboard) && !isBuiltInDashboard(dashboard)
    )

    // We use the fact that insight can not be in more than one organization
    // we infer active organization by simple search a first built org-based dashboard
    const [activeOrganization, setActiveOrganization] = useState(
        Object.values(value)
            .find(dashboard => isOrganizationDashboard(dashboard) && isBuiltInDashboard(dashboard))
        ?? organizations[0]
    )

    const isActive = (dashboardId: string | undefined): boolean => {
        if (dashboardId) {
            return !!value[dashboardId]
        }

        return false
    }

    const isOrganizationModeActive = (): boolean => {
        const dashboardIds = new Set(Object.keys(value))
        const pickedOrganizationDashboards = organizations
            .filter(orgDashboard => dashboardIds.has(orgDashboard.id))

        return !!pickedOrganizationDashboards.length
    }

    const handleOrganizationScopeChange = (event: ChangeEvent<HTMLInputElement>): void => {
        const isActive = event.target.checked
        const currentDashboard = dashboards.find(dashboard => dashboard.id === activeOrganization?.id) ?? organizations[0]

        if (isActive) {
            onChange({ ...value, [currentDashboard.id]: currentDashboard})
        } else {

            const nextValue = createDashboardsMap(
                ...Object.values(value).filter(dashboard => !isOrganizationDashboard(dashboard))
            )

            onChange(nextValue)
        }
    }

    const handleOrganizationChange = (event: ChangeEvent<HTMLSelectElement>): void => {
        const organizationDashboard =
            dashboards.find(dashboard => dashboard.id === event.target.value)

        if (!organizationDashboard) {
            return
        }

        // Update local state of organizations select component
        setActiveOrganization(organizationDashboard)

        const nextValue = createDashboardsMap(
            organizationDashboard,
            ...Object.values(value)
                .filter(dashboard => dashboard.owner.id !== organizationDashboard.id)
        )

        onChange(nextValue)
    }

    const handleDashboardChange = (event: ChangeEvent<HTMLInputElement>): void => {
        const dashboardId = event.target.value
        const isChecked = event.target.checked
        const currentDashboard = dashboards.find(dashboard => dashboard.id === dashboardId)

        if (!currentDashboard) {
            return
        }

        if (isChecked) {
            onChange({...value, [currentDashboard.id]: currentDashboard })

            return
        }

        const nextValue = createDashboardsMap(
            ...Object.values(value).filter(dashboard => dashboard.id !== currentDashboard.id)
        )

        onChange(nextValue)
    }

    return (
        <div className={className}>
            {
                organizations.length &&
                <>
                    <div className='d-flex align-items-center'>
                        <label className='d-flex align-items-center mr-2 mb-0'>
                            <input
                                type="checkbox"
                                value='organization'
                                checked={isOrganizationModeActive()}
                                onChange={handleOrganizationScopeChange}
                                className='mr-2'
                            />

                            <span className='font-weight-medium'>Organization</span>
                        </label>

                        <select
                            disabled={!isOrganizationModeActive()}
                            name="organizations"
                            value={activeOrganization?.id}
                            onChange={handleOrganizationChange}
                            className='custom-select w-auto'>
                            {
                                organizations
                                    .map(org => <option key={org.id} value={org.id}>{ org.title }</option>)

                            }
                        </select>
                    </div>

                    <div className='pl-3'>
                        {
                            organizationsDashboards
                                .filter(isSettingsBasedInsightsDashboard)
                                .filter(dashboard => dashboard.owner.id === activeOrganization?.id)
                                .map(dashboard =>
                                    <label key={dashboard.id} className='d-flex align-items-center mt-2'>

                                        <input
                                            disabled={!isOrganizationModeActive()}
                                            type="checkbox"
                                            value={dashboard.id}
                                            checked={isActive(dashboard.id)}
                                            onChange={handleDashboardChange}
                                        />

                                        <span className='pl-2'>{ dashboard.title }</span>
                                    </label>
                                )
                        }
                    </div>
                </>
            }
            {
                !organizations.length && <span className='text-muted'>You are not included in any organization</span>
            }
        </div>
    )
}

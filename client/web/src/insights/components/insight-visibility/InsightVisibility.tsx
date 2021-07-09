import React, { ChangeEvent } from 'react';

import { isPersonalDashboard, RealInsightDashboard } from '../../core/types';
import { isSettingsBasedInsightsDashboard } from '../../core/types/dashboard/real-dashboard';

import { InsightOrganizationVisibility } from './components/insight-organization-visibility/InsightOrganizationVisibility';
import { createDashboardsMap } from './utils';

interface InsightVisibilityProps {
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
}

/**
 * Renders visibility scope set of fields (personal and organization dashboards)
 */
export const InsightVisibility: React.FunctionComponent<InsightVisibilityProps> = props => {
    const { value, dashboards, onChange } = props

    const personalCustomDashboards = dashboards
        .filter(isPersonalDashboard)
        .filter(isSettingsBasedInsightsDashboard);

    const isActive = (dashboardId: string | undefined): boolean => {
        if (dashboardId) {
            return !!value[dashboardId]
        }

        return false
    }

    const handleChange = (event: ChangeEvent<HTMLInputElement>): void => {
        const dashboardId = event.target.value
        const isActive = event.target.checked
        const currentDashboard = dashboards.find(dashboard => dashboard.id === dashboardId)

        if (!currentDashboard) {
            return
        }

        if (isActive) {
            onChange({...value, [currentDashboard.id]: currentDashboard })

            return
        }

        const nextValue = createDashboardsMap(
            ...Object.values(value)
                .filter(dashboard => dashboard.id !== currentDashboard.id)
        )

        onChange(nextValue)
    }

    return (
        <fieldset>

            <span className='font-weight-medium'>Personal</span>

            <div className='pl-3'>
                {
                    personalCustomDashboards.map(dashboard =>
                        <label key={dashboard.id} className='d-flex align-items-center mt-2'>
                            <input
                                type="checkbox"
                                value={dashboard.id}
                                checked={isActive(dashboard.id)}
                                onChange={handleChange}
                            />
                            <span className='ml-2'>{ dashboard.title }</span>
                        </label>
                    )
                }
            </div>

            <InsightOrganizationVisibility
                value={value}
                dashboards={dashboards}
                onChange={onChange}
                className='mt-3'/>
        </fieldset>
    )
}


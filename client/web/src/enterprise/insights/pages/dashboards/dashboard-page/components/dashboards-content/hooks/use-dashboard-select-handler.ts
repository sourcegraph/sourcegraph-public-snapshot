import { useHistory } from 'react-router-dom'

import { InsightDashboard, isVirtualDashboard } from '../../../../../../core/types'

type SelectHandler = (dashboard: InsightDashboard) => void

/**
 * Hook for managing URL of the dashboard page whenever the user picks
 * another dashboard via dashboard select
 */
export function useDashboardSelectHandler(): SelectHandler {
    const history = useHistory()

    return (dashboard: InsightDashboard): void => {
        if (isVirtualDashboard(dashboard)) {
            history.push(`/insights/dashboards/${dashboard.id}`)

            return
        }

        history.push(`/insights/dashboards/${dashboard.id}`)
    }
}

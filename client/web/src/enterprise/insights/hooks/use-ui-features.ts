import { useContext, useMemo } from 'react'

import { CodeInsightsBackendContext } from '../core/backend/code-insights-backend-context'
import { InsightDashboard } from '../core/types'
import { getTooltipMessage } from '../pages/dashboards/dashboard-page/utils/get-dashboard-permissions'

interface DashboardMenuItem {
    disabled?: boolean
    tooltip?: string
    display: boolean
}

type DashboardMenuItemKey = 'configure' | 'copy' | 'delete'

export interface UseUiFeatures {
    licensed: boolean
    dashboards: {
        addRemoveInsightsButton: {
            disabled: boolean
            tooltip: string | undefined
        }
        menu: Record<DashboardMenuItemKey, DashboardMenuItem>
    }
}

export interface UseUiFeaturesProps {
    currentDashboard?: InsightDashboard
}

export function useUiFeatures({ currentDashboard }: UseUiFeaturesProps): UseUiFeatures {
    const { getUiFeatures } = useContext(CodeInsightsBackendContext)

    const { licensed, permissions } = useMemo(() => getUiFeatures(currentDashboard), [getUiFeatures, currentDashboard])

    return {
        licensed,
        dashboards: {
            addRemoveInsightsButton: {
                disabled: !permissions.isConfigurable,
                tooltip: getTooltipMessage(currentDashboard, permissions),
            },
            // Available menu items
            menu: {
                configure: {
                    display: !licensed,
                    disabled: !permissions.isConfigurable,
                    tooltip: getTooltipMessage(currentDashboard, permissions),
                },
                copy: {
                    display: true,
                },
                delete: {
                    display: !licensed,
                    disabled: !permissions.isConfigurable,
                    tooltip: getTooltipMessage(currentDashboard, permissions),
                },
            },
        },
    }
}

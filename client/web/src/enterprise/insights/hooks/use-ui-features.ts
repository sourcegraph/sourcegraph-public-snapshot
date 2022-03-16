import { useContext, useMemo } from 'react'

import { Observable, of } from 'rxjs'
import { map } from 'rxjs/operators'

import { CodeInsightsBackendContext } from '../core/backend/code-insights-backend-context'
import { Insight, InsightDashboard, isSearchBasedInsight } from '../core/types'
import {
    getDashboardPermissions,
    getTooltipMessage,
} from '../pages/dashboards/dashboard-page/utils/get-dashboard-permissions'

interface DashboardMenuItem {
    disabled?: boolean
    tooltip?: string
    display: boolean
}

type DashboardMenuItemKey = 'configure' | 'copy' | 'delete'

export interface UseUiFeatures {
    licensed: boolean
    dashboard: {
        createPermissions: {
            submit: {
                disabled: boolean
                tooltip?: string
            }
        }
        getContextActionsPermissions: (dashboard?: InsightDashboard) => Record<DashboardMenuItemKey, DashboardMenuItem>
        getAddRemoveInsightsPermission: (
            dashboard?: InsightDashboard
        ) => {
            disabled: boolean
            tooltip: string | undefined
        }
    }
    insight: {
        getContextActionsPermissions: (insight: Insight) => { showYAxis: boolean }
        getCreationPermissions: () => Observable<{ available: true } | { available: false; reason: string }>
    }
}

export function useUiFeatures(): UseUiFeatures {
    const { getUiFeatures, hasInsights } = useContext(CodeInsightsBackendContext)

    const { licensed, insightsLimit } = useMemo(() => getUiFeatures(), [getUiFeatures])

    return {
        licensed,
        dashboard: {
            createPermissions: { submit: { disabled: !licensed } },
            getAddRemoveInsightsPermission: (dashboard?: InsightDashboard) => {
                const permissions = getDashboardPermissions(dashboard, true)

                return {
                    disabled: !permissions.isConfigurable,
                    tooltip: getTooltipMessage(dashboard, permissions),
                }
            },
            // Available menu items
            getContextActionsPermissions: (dashboard?: InsightDashboard) => {
                const permissions = getDashboardPermissions(dashboard, true)

                return {
                    configure: {
                        display: licensed,
                        disabled: !permissions.isConfigurable,
                        tooltip: getTooltipMessage(dashboard, permissions),
                    },
                    copy: {
                        display: licensed,
                        disabled: !dashboard,
                    },
                    delete: {
                        display: true,
                        disabled: !permissions.isConfigurable,
                        tooltip: getTooltipMessage(dashboard, permissions),
                    },
                }
            },
        },
        insight: {
            getContextActionsPermissions: (insight: Insight) => ({
                showYAxis: isSearchBasedInsight(insight) && !insight.locked,
            }),
            getCreationPermissions: () =>
                insightsLimit !== null
                    ? hasInsights(insightsLimit).pipe(
                          map(reachedLimit =>
                              reachedLimit
                                  ? { available: false, reason: 'You already have enough insights buddy' }
                                  : { available: true }
                          )
                      )
                    : of({ available: true }),
        },
    }
}

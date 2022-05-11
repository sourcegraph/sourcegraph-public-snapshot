import { useContext, useMemo } from 'react'

import { Observable, of } from 'rxjs'
import { map } from 'rxjs/operators'

import { CodeInsightsBackendContext, Insight, InsightDashboard, isSearchBasedInsight } from '../core'
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
        getCreationPermissions: () => Observable<{ available: boolean }>
        getEditPermissions: (insight: Insight) => Observable<{ available: boolean }>
    }
}

export function useUiFeatures(): UseUiFeatures {
    const { UIFeatures, getActiveInsightsCount } = useContext(CodeInsightsBackendContext)
    const { licensed, insightsLimit } = UIFeatures

    return useMemo(
        () => ({
            licensed,
            dashboard: {
                createPermissions: { submit: { disabled: !licensed } },
                getAddRemoveInsightsPermission: (dashboard?: InsightDashboard) => {
                    const permissions = getDashboardPermissions(dashboard)

                    if (!licensed) {
                        return {
                            disabled: true,
                            tooltip: 'Limited access: upgrade your license to add insights to dashboards',
                        }
                    }

                    return {
                        disabled: !permissions.isConfigurable,
                        tooltip: getTooltipMessage(permissions),
                    }
                },
                getContextActionsPermissions: (dashboard?: InsightDashboard) => {
                    const permissions = getDashboardPermissions(dashboard)

                    return {
                        configure: {
                            display: licensed,
                            disabled: !permissions.isConfigurable,
                            tooltip: getTooltipMessage(permissions),
                        },
                        copy: {
                            display: licensed,
                            disabled: !dashboard,
                        },
                        delete: {
                            display: true,
                            disabled: !permissions.isConfigurable,
                            tooltip: getTooltipMessage(permissions),
                        },
                    }
                },
            },
            insight: {
                getContextActionsPermissions: (insight: Insight) => ({
                    showYAxis: isSearchBasedInsight(insight),
                }),
                getCreationPermissions: () =>
                    insightsLimit !== null
                        ? getActiveInsightsCount(insightsLimit).pipe(
                              map(insightCount => ({ available: insightCount < insightsLimit }))
                          )
                        : of({ available: true }),
                getEditPermissions: (insight: Insight) => of({ available: licensed || !insight.isFrozen }),
            },
        }),
        [licensed, insightsLimit, getActiveInsightsCount]
    )
}

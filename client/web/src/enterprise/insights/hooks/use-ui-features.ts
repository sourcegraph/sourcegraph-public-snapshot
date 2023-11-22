import { useContext, useMemo } from 'react'

import { type Observable, of } from 'rxjs'
import { map } from 'rxjs/operators'

import { CodeInsightsBackendContext, type CustomInsightDashboard, type Insight, isSearchBasedInsight } from '../core'
import {
    getDashboardPermissions,
    getTooltipMessage,
} from '../pages/dashboards/dashboard-view/utils/get-dashboard-permissions'
import { useCodeInsightsLicenseState } from '../stores'

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
        getContextActionsPermissions: (
            dashboard?: CustomInsightDashboard
        ) => Record<DashboardMenuItemKey, DashboardMenuItem>
        getAddRemoveInsightsPermission: (dashboard?: CustomInsightDashboard) => {
            disabled: boolean
            tooltip: string | undefined
        }
    }
    insight: {
        getContextActionsPermissions: (insight: Insight) => { showYAxis: boolean }
        getCreationPermissions: () => Observable<{ available: boolean }>
        getEditPermissions: (insight: Insight | undefined | null) => Observable<{ available: boolean }>
    }
}

export function useUiFeatures(): UseUiFeatures {
    const { getActiveInsightsCount } = useContext(CodeInsightsBackendContext)
    const { licensed, insightsLimit } = useCodeInsightsLicenseState()

    return useMemo(
        () => ({
            licensed,
            dashboard: {
                createPermissions: { submit: { disabled: !licensed } },
                getAddRemoveInsightsPermission: (dashboard?: CustomInsightDashboard) => {
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
                getContextActionsPermissions: (dashboard?: CustomInsightDashboard) => {
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
                getContextActionsPermissions: (insight: Insight) => ({ showYAxis: isSearchBasedInsight(insight) }),
                getCreationPermissions: () =>
                    insightsLimit !== null
                        ? getActiveInsightsCount(insightsLimit).pipe(
                              map(insightCount => ({ available: insightCount < insightsLimit }))
                          )
                        : of({ available: true }),
                getEditPermissions: (insight: Insight | undefined | null) =>
                    insight ? of({ available: licensed || !insight?.isFrozen }) : of({ available: false }),
            },
        }),
        [licensed, insightsLimit, getActiveInsightsCount]
    )
}

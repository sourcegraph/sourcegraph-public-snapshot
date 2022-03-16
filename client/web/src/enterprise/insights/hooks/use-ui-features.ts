import { useContext, useMemo } from 'react'

import { Observable, of } from 'rxjs';
import { map } from 'rxjs/operators';

import { CodeInsightsBackendContext } from '../core/backend/code-insights-backend-context'
import { Insight, InsightDashboard, isSearchBasedInsight } from '../core/types'
import {
    getDashboardPermissions,
    getTooltipMessage
} from '../pages/dashboards/dashboard-page/utils/get-dashboard-permissions'

interface DashboardMenuItem {
    disabled?: boolean
    tooltip?: string
    display: boolean
}

type DashboardMenuItemKey = 'configure' | 'copy' | 'delete'

export interface UseUiFeatures {
    licensed: boolean
    dashboards: {
        getAddRemoveInsightsPermission: (dashboard?: InsightDashboard) => ({
            disabled: boolean
            tooltip: string | undefined
        })
        getActionPermissions: (dashboard?: InsightDashboard) => Record<DashboardMenuItemKey, DashboardMenuItem>
    }
    insights: {
        menu: {
            showYAxis: (insight: Insight) => boolean
        }
    },
    insight: {
        isCreationAvailable: () => Observable<{ available: true } | { available: false, reason: string }>
    }
}

export function useUiFeatures(): UseUiFeatures {
    const { getUiFeatures, hasInsights } = useContext(CodeInsightsBackendContext)

    const {
        licensed,
        insightsLimit
    } = useMemo(() => getUiFeatures(), [getUiFeatures])

    return {
        licensed,
        dashboards: {
            getAddRemoveInsightsPermission: (dashboard?: InsightDashboard) => {
                const permissions = getDashboardPermissions(dashboard, true)

                return {
                    disabled: !permissions.isConfigurable,
                    tooltip: getTooltipMessage(dashboard, permissions),
                }
            },
            create: {
                addDashboardButton: {
                    disabled: !licensed,
                },
            },
            // Available menu items
            getActionPermissions: (dashboard?: InsightDashboard) => {
                const permissions = getDashboardPermissions(dashboard, true)

                return {
                    configure: {
                        display: licensed,
                        disabled: !permissions.isConfigurable,
                        tooltip: getTooltipMessage(dashboard, permissions),
                    },
                    copy: {
                        display: licensed,
                    },
                    delete: {
                        display: true,
                        disabled: !permissions.isConfigurable,
                        tooltip: getTooltipMessage(dashboard, permissions),
                    },
                }
            }
        },
        insights: {
            menu: {
                showYAxis: insight => isSearchBasedInsight(insight) && !insight.isFrozen,
            },
        },
        insight: {
            isCreationAvailable: () => insightsLimit !== null
                ? hasInsights(insightsLimit).pipe(
                    map(reachedLimit => reachedLimit
                            ? { available: false, reason: 'You already have enough insights buddy' }
                            : { available: true })
                )
                : of({ available: true })
        }
    }
}

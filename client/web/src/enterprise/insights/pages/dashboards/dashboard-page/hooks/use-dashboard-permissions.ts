import { InsightDashboard, InsightsDashboardScope, isRealDashboard, isVirtualDashboard } from '../../../../core/types'
import { isCustomInsightDashboard } from '../../../../core/types/dashboard/real-dashboard'
import { isGlobalSubject, SupportedInsightSubject } from '../../../../core/types/subjects'

enum DashboardReasonDenied {
    AllVirtualDashboard,
    BuiltInCantBeEdited,
    PermissionDenied,
    UnknownDashboard,
}

type DashboardPermissions =
    | {
          isConfigurable: false
          reason: DashboardReasonDenied
      }
    | {
          isConfigurable: true
      }

const DEFAULT_DASHBOARD_PERMISSIONS: DashboardPermissions = {
    isConfigurable: false,
    reason: DashboardReasonDenied.UnknownDashboard,
}

export function useDashboardPermissions(
    dashboard: InsightDashboard | undefined,
    supportedSubjects?: SupportedInsightSubject[]
): DashboardPermissions {
    if (dashboard && 'grants' in dashboard) {
        // This means we're using the graphql api.
        // Since the api only returns info the user can see
        // We can safely assume the user has permission to edit the dashboard
        return {
            isConfigurable: true,
        }
    }

    if (!dashboard) {
        return DEFAULT_DASHBOARD_PERMISSIONS
    }

    if (!dashboard) {
        return DEFAULT_DASHBOARD_PERMISSIONS
    }

    if (isVirtualDashboard(dashboard)) {
        return {
            isConfigurable: false,
            reason: DashboardReasonDenied.AllVirtualDashboard,
        }
    }

    if (!supportedSubjects) {
        return DEFAULT_DASHBOARD_PERMISSIONS
    }

    const dashboardOwner = supportedSubjects.find(subject => subject.id === dashboard?.owner?.id)

    // No dashboard can't be modified
    if (!dashboard || !dashboardOwner) {
        return DEFAULT_DASHBOARD_PERMISSIONS
    }

    if (isRealDashboard(dashboard)) {
        // Settings based insights dashboards (custom dashboards created by users)
        if (isCustomInsightDashboard(dashboard)) {
            // Global scope permission handling
            if (isGlobalSubject(dashboardOwner)) {
                const canBeEdited = dashboardOwner.viewerCanAdminister

                if (!canBeEdited) {
                    return {
                        isConfigurable: false,
                        reason: DashboardReasonDenied.PermissionDenied,
                    }
                }
            }

            return {
                isConfigurable: true,
            }
        }

        // Not settings based dashboard (built-in-dashboard case)
        return {
            isConfigurable: false,
            reason: DashboardReasonDenied.BuiltInCantBeEdited,
        }
    }

    return DEFAULT_DASHBOARD_PERMISSIONS
}

export function getTooltipMessage(
    dashboard: InsightDashboard | undefined,
    permissions: DashboardPermissions
): string | undefined {
    if (!dashboard) {
        return 'Dashboard not found'
    }

    if (permissions.isConfigurable) {
        return
    }

    switch (permissions.reason) {
        case DashboardReasonDenied.UnknownDashboard:
            return 'Dashboard not found'
        case DashboardReasonDenied.PermissionDenied:
            return "You don't have permission to edit this dashboard"
        case DashboardReasonDenied.BuiltInCantBeEdited:
            switch (dashboard.scope) {
                case InsightsDashboardScope.Personal:
                    return "This is an automatically created dashboard that lists all your private insights. You can't edit this dashboard."
                case InsightsDashboardScope.Organization:
                case InsightsDashboardScope.Global:
                    if (!dashboard.owner) {
                        throw new Error('TODO: support GraphQL API')
                    }
                    return `This is an automatically created dashboard that lists all ${dashboard.owner.name} insights. You can't edit this dashboard.`
            }
        case DashboardReasonDenied.AllVirtualDashboard:
            return "This is an automatically created dashboard that lists all the insights you have access to. You can't edit this dashboard."
    }
}

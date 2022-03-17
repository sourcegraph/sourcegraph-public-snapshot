import { InsightDashboard, isVirtualDashboard } from '../../../../core/types'

enum DashboardReasonDenied {
    AllVirtualDashboard,
    PermissionDenied,
    UnknownDashboard,
}

interface DashboardDeniedPermissions {
    isConfigurable: false
    reason: DashboardReasonDenied
}

interface DashboardGrantedPermissions {
    isConfigurable: true
}

export type DashboardPermissions = DashboardDeniedPermissions | DashboardGrantedPermissions

const DEFAULT_DASHBOARD_PERMISSIONS: DashboardPermissions = {
    isConfigurable: false,
    reason: DashboardReasonDenied.UnknownDashboard,
}

export function getDashboardPermissions(dashboard: InsightDashboard | undefined): DashboardPermissions {
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

    if (isVirtualDashboard(dashboard)) {
        return {
            isConfigurable: false,
            reason: DashboardReasonDenied.AllVirtualDashboard,
        }
    }

    return DEFAULT_DASHBOARD_PERMISSIONS
}

export function getTooltipMessage(permissions: DashboardPermissions): string | undefined {
    if (permissions.isConfigurable) {
        return
    }

    switch (permissions.reason) {
        case DashboardReasonDenied.UnknownDashboard:
            return 'Dashboard not found'
        case DashboardReasonDenied.PermissionDenied:
            return "You don't have permission to edit this dashboard"
        case DashboardReasonDenied.AllVirtualDashboard:
            return "This is an automatically created dashboard that lists all the insights you have access to. You can't edit this dashboard."
    }
}

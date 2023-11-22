import type { CustomInsightDashboard } from '../../../../core/types'

enum DashboardReasonDenied {
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

export function getDashboardPermissions(dashboard: CustomInsightDashboard | undefined): DashboardPermissions {
    if (dashboard) {
        return { isConfigurable: true }
    }

    return {
        isConfigurable: false,
        reason: DashboardReasonDenied.UnknownDashboard,
    }
}

export function getTooltipMessage(permissions: DashboardPermissions): string | undefined {
    if (permissions.isConfigurable) {
        return
    }

    switch (permissions.reason) {
        case DashboardReasonDenied.UnknownDashboard: {
            return 'Dashboard not found'
        }
        case DashboardReasonDenied.PermissionDenied: {
            return "You don't have permission to edit this dashboard"
        }
    }
}

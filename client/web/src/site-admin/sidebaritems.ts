import AccountMultipleIcon from 'mdi-react/AccountMultipleIcon'
import ChartLineVariantIcon from 'mdi-react/ChartLineVariantIcon'
import CogsIcon from 'mdi-react/CogsIcon'
import ConsoleIcon from 'mdi-react/ConsoleIcon'
import MonitorStarIcon from 'mdi-react/MonitorStarIcon'
import SourceRepositoryIcon from 'mdi-react/SourceRepositoryIcon'

import { checkRequestAccessAllowed } from '../util/checkRequestAccessAllowed'

import { isPackagesEnabled } from './flags'
import { SiteAdminSideBarGroup, SiteAdminSideBarGroups } from './SiteAdminSidebar'

export const analyticsGroup: SiteAdminSideBarGroup = {
    condition: ({ isSourcegraphApp }) => !isSourcegraphApp,
    header: {
        label: 'Analytics',
        icon: ChartLineVariantIcon,
    },
    items: [
        {
            label: 'Overview',
            to: '/site-admin/',
            exact: true,
        },
        {
            label: 'Search',
            to: '/site-admin/analytics/search',
        },
        {
            label: 'Code navigation',
            to: '/site-admin/analytics/code-intel',
        },
        {
            label: 'Users',
            to: '/site-admin/analytics/users',
        },
        {
            label: 'Insights',
            to: '/site-admin/analytics/code-insights',
        },
        {
            label: 'Batch changes',
            to: '/site-admin/analytics/batch-changes',
        },
        {
            label: 'Notebooks',
            to: '/site-admin/analytics/notebooks',
        },
        {
            label: 'Extensions',
            to: '/site-admin/analytics/extensions',
        },
        {
            label: 'Feedback survey',
            to: '/site-admin/surveys',
            condition: ({ isSourcegraphApp }) => !isSourcegraphApp,
        },
    ],
}

export const configurationGroup: SiteAdminSideBarGroup = {
    header: {
        label: 'Configuration',
        icon: CogsIcon,
    },
    items: [
        {
            label: 'Site configuration',
            to: '/site-admin/configuration',
        },
        {
            label: 'Global settings',
            to: '/site-admin/global-settings',
            condition: ({ isSourcegraphApp }) => !isSourcegraphApp,
        },
        {
            label: 'Feature flags',
            to: '/site-admin/feature-flags',
        },
    ],
}

export const repositoriesGroup: SiteAdminSideBarGroup = {
    header: {
        label: 'Repositories',
        icon: SourceRepositoryIcon,
    },
    items: [
        {
            label: 'Code host connections',
            to: '/site-admin/external-services',
        },
        {
            label: 'Repositories',
            to: '/site-admin/repositories',
        },
        {
            label: 'Packages',
            to: '/site-admin/packages',
            condition: isPackagesEnabled,
        },
    ],
}

export const usersGroup: SiteAdminSideBarGroup = {
    header: {
        label: 'Users & auth',
        icon: AccountMultipleIcon,
    },

    condition: ({ isSourcegraphApp }) => !isSourcegraphApp,
    items: [
        {
            label: 'Users',
            to: '/site-admin/users',
        },
        {
            label: 'Account requests',
            to: '/site-admin/account-requests',
            condition: () => checkRequestAccessAllowed(window.context),
        },
        {
            label: 'Organizations',
            to: '/site-admin/organizations',
        },
        {
            label: 'Access tokens',
            to: '/site-admin/tokens',
        },
    ],
}

export const maintenanceGroupHeaderLabel = 'Maintenance'

export const maintenanceGroupMonitoringItemLabel = 'Monitoring'

export const maintenanceGroupInstrumentationItemLabel = 'Instrumentation'

export const maintenanceGroupUpdatesItemLabel = 'Updates'

export const maintenanceGroupMigrationsItemLabel = 'Migrations'

export const maintenanceGroupTracingItemLabel = 'Tracing'

export const maintenanceGroup: SiteAdminSideBarGroup = {
    header: {
        label: maintenanceGroupHeaderLabel,
        icon: MonitorStarIcon,
    },
    items: [
        {
            label: maintenanceGroupUpdatesItemLabel,
            to: '/site-admin/updates',
        },
        {
            label: 'Documentation',
            to: '/help',
            condition: ({ isSourcegraphApp }) => !isSourcegraphApp,
        },
        {
            label: 'Pings',
            to: '/site-admin/pings',
            condition: ({ isSourcegraphApp }) => !isSourcegraphApp,
        },
        {
            label: 'Report a bug',
            to: '/site-admin/report-bug',
            condition: ({ isSourcegraphApp }) => !isSourcegraphApp,
        },
        {
            label: maintenanceGroupMigrationsItemLabel,
            to: '/site-admin/migrations',
            condition: ({ isSourcegraphApp }) => !isSourcegraphApp,
        },
        {
            label: maintenanceGroupInstrumentationItemLabel,
            to: '/-/debug/',
            source: 'server',
            condition: ({ isSourcegraphApp }) => !isSourcegraphApp,
        },
        {
            label: maintenanceGroupMonitoringItemLabel,
            to: '/-/debug/grafana',
            source: 'server',
            condition: ({ isSourcegraphApp }) => !isSourcegraphApp,
        },
        {
            label: maintenanceGroupTracingItemLabel,
            to: '/-/debug/jaeger',
            source: 'server',
            condition: ({ isSourcegraphApp }) => !isSourcegraphApp,
        },
        {
            label: 'Outbound requests',
            to: '/site-admin/outbound-requests',
            condition: ({ isSourcegraphApp }) => !isSourcegraphApp,
        },
        {
            label: 'Slow requests',
            to: '/site-admin/slow-requests',
            condition: ({ isSourcegraphApp }) => !isSourcegraphApp,
        },
        {
            label: 'Background jobs',
            to: '/site-admin/background-jobs',
            condition: ({ isSourcegraphApp }) => !isSourcegraphApp,
        },
    ],
}

export const apiConsoleGroup: SiteAdminSideBarGroup = {
    header: {
        label: 'API Console',
        icon: ConsoleIcon,
    },
    items: [
        {
            label: 'API Console',
            to: '/api/console',
        },
    ],
}

export const siteAdminSidebarGroups: SiteAdminSideBarGroups = [
    analyticsGroup,
    configurationGroup,
    repositoriesGroup,
    usersGroup,
    maintenanceGroup,
    apiConsoleGroup,
]

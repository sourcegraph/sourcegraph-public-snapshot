import AccountMultipleIcon from 'mdi-react/AccountMultipleIcon'
import ChartLineVariantIcon from 'mdi-react/ChartLineVariantIcon'
import CogsIcon from 'mdi-react/CogsIcon'
import ConsoleIcon from 'mdi-react/ConsoleIcon'
import MonitorStarIcon from 'mdi-react/MonitorStarIcon'
import SourceRepositoryIcon from 'mdi-react/SourceRepositoryIcon'

import { checkRequestAccessAllowed } from '../util/checkRequestAccessAllowed'

import { isPackagesEnabled } from './flags'
import type { SiteAdminSideBarGroup, SiteAdminSideBarGroups } from './SiteAdminSidebar'

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
            condition: ({ codeInsightsEnabled }) => codeInsightsEnabled,
        },
        {
            label: 'Batch changes',
            to: '/site-admin/analytics/batch-changes',
            condition: ({ batchChangesEnabled }) => batchChangesEnabled,
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
            label: 'Code ownership',
            to: '/site-admin/analytics/own',
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
            label: 'End user onboarding',
            to: '/site-admin/end-user-onboarding',
            condition: ({ endUserOnboardingEnabled }) => endUserOnboardingEnabled,
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
            condition: ({ isSourcegraphApp }) => !isSourcegraphApp,
        },
        {
            label: 'Packages',
            to: '/site-admin/packages',
            condition: isPackagesEnabled,
        },
        {
            label: 'Gitservers',
            to: '/site-admin/gitservers',
        },
    ],
    condition: ({ isSourcegraphApp }) => !isSourcegraphApp,
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
    condition: ({ isSourcegraphApp }) => !isSourcegraphApp,
    items: [
        {
            label: maintenanceGroupUpdatesItemLabel,
            to: '/site-admin/updates',
        },
        {
            label: 'Documentation',
            to: '/help',
        },
        {
            label: 'Pings',
            to: '/site-admin/pings',
        },
        {
            label: 'Report a bug',
            to: '/site-admin/report-bug',
        },
        {
            label: maintenanceGroupMigrationsItemLabel,
            to: '/site-admin/migrations',
        },
        {
            label: maintenanceGroupInstrumentationItemLabel,
            to: '/-/debug/',
            source: 'server',
        },
        {
            label: maintenanceGroupMonitoringItemLabel,
            to: '/-/debug/grafana',
            source: 'server',
        },
        {
            label: maintenanceGroupTracingItemLabel,
            to: '/-/debug/jaeger',
            source: 'server',
        },
        {
            label: 'Outbound requests',
            to: '/site-admin/outbound-requests',
        },
        {
            label: 'Slow requests',
            to: '/site-admin/slow-requests',
        },
        {
            label: 'Background jobs',
            to: '/site-admin/background-jobs',
        },
    ],
}

export const apiConsoleGroup: SiteAdminSideBarGroup = {
    header: {
        label: 'API Console',
        icon: ConsoleIcon,
    },
    condition: ({ isSourcegraphApp }) => !isSourcegraphApp,
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

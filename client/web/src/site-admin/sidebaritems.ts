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
            label: 'Manage code hosts',
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
            label: 'Access requests',
            to: '/site-admin/access-requests',
            condition: context =>
                checkRequestAccessAllowed(
                    context.isSourcegraphDotCom,
                    window.context.allowSignup,
                    window.context.experimentalFeatures
                ),
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

export const maintenanceGroup: SiteAdminSideBarGroup = {
    header: {
        label: 'Maintenance',
        icon: MonitorStarIcon,
    },
    items: [
        {
            label: 'Updates',
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
            label: 'Migrations',
            to: '/site-admin/migrations',
            condition: ({ isSourcegraphApp }) => !isSourcegraphApp,
        },
        {
            label: 'Instrumentation',
            to: '/-/debug/',
            source: 'server',
            condition: ({ isSourcegraphApp }) => !isSourcegraphApp,
        },
        {
            label: 'Monitoring',
            to: '/-/debug/grafana',
            source: 'server',
            condition: ({ isSourcegraphApp }) => !isSourcegraphApp,
        },
        {
            label: 'Tracing',
            to: '/-/debug/jaeger',
            source: 'server',
            condition: ({ isSourcegraphApp }) => !isSourcegraphApp,
        },
        {
            label: 'Outbound requests',
            to: '/site-admin/outbound-requests',
            source: 'server',
            condition: ({ isSourcegraphApp }) => !isSourcegraphApp,
        },
        {
            label: 'Slow requests',
            to: '/site-admin/slow-requests',
            source: 'server',
            condition: ({ isSourcegraphApp }) => !isSourcegraphApp,
        },
        {
            label: 'Background jobs',
            to: '/site-admin/background-jobs',
            source: 'server',
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

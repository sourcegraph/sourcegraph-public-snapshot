import AccountMultipleIcon from 'mdi-react/AccountMultipleIcon'
import CogsIcon from 'mdi-react/CogsIcon'
import ConsoleIcon from 'mdi-react/ConsoleIcon'
import EarthIcon from 'mdi-react/EarthIcon'
import MonitorStarIcon from 'mdi-react/MonitorStarIcon'
import SourceRepositoryIcon from 'mdi-react/SourceRepositoryIcon'

import { SiteAdminSideBarGroup, SiteAdminSideBarGroups } from './SiteAdminSidebar'

export const overviewGroup: SiteAdminSideBarGroup = {
    header: {
        label: 'Statistics',
        icon: EarthIcon,
    },
    items: [
        {
            label: 'Overview',
            to: '/site-admin',
            exact: true,
        },
        {
            label: 'Usage stats',
            to: '/site-admin/usage-statistics',
        },
        {
            label: 'Feedback survey',
            to: '/site-admin/surveys',
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
    ],
}

export const usersGroup: SiteAdminSideBarGroup = {
    header: {
        label: 'Users & auth',
        icon: AccountMultipleIcon,
    },
    items: [
        {
            label: 'Users',
            to: '/site-admin/users',
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
            label: 'Migrations',
            to: '/site-admin/migrations',
        },
        {
            label: 'Instrumentation',
            to: '/-/debug/',
            source: 'server',
            condition: () =>
                window.context.deployType === 'kubernetes' ||
                window.context.deployType === 'dev' ||
                window.context.deployType === 'helm',
        },
        {
            label: 'Monitoring',
            to: '/-/debug/grafana',
            source: 'server',
        },
        {
            label: 'Tracing',
            to: '/-/debug/jaeger',
            source: 'server',
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
    overviewGroup,
    configurationGroup,
    repositoriesGroup,
    usersGroup,
    maintenanceGroup,
    apiConsoleGroup,
]

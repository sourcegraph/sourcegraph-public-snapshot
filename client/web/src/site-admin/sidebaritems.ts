import { SiteAdminSideBarGroup, SiteAdminSideBarGroups } from './SiteAdminSidebar'

export const siteAdminGroup: SiteAdminSideBarGroup = {
    header: {
        label: 'Site admin',
    },
    items: [
        {
            label: 'Overview',
            to: '/site-admin',
            exact: true,
        },
        {
            label: 'Site configuration',
            to: '/site-admin/configuration',
        },
        {
            label: 'Global settings',
            to: '/site-admin/global-settings',
        },
    ],
}

export const repositoriesGroup: SiteAdminSideBarGroup = {
    header: {
        label: 'Repositories',
    },
    items: [
        {
            label: 'Manage code hosts',
            to: '/site-admin/external-services',
        },
        {
            label: 'Repository status',
            to: '/site-admin/repositories',
        },
    ],
}

export const usersGroup: SiteAdminSideBarGroup = {
    header: {
        label: 'Users & auth',
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

export const maintenanceGroup: SiteAdminSideBarGroup = {
    header: {
        label: 'Maintenance',
    },
    items: [
        {
            label: 'Updates',
            to: '/site-admin/updates',
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
            condition: () => window.context.deployType === 'kubernetes',
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

export const siteAdminSidebarGroups: SiteAdminSideBarGroups = [
    siteAdminGroup,
    repositoriesGroup,
    usersGroup,
    maintenanceGroup,
]

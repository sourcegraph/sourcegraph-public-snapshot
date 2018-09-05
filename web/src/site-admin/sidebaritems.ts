import { SiteAdminSideBarItems } from './SiteAdminSidebar'

export const siteAdminSidebarNavItems: SiteAdminSideBarItems = {
    primary: [
        {
            label: 'Overview',
            to: '/site-admin',
        },
        {
            label: 'Configuration',
            to: '/site-admin/configuration',
        },
        {
            label: 'Repositories',
            to: '/site-admin/repositories',
        },
    ],
    secondary: [
        {
            label: 'Users',
            to: '/site-admin/users',
        },
        {
            label: 'Organizations',
            to: '/site-admin/organizations',
        },
        {
            label: 'Global settings',
            to: '/site-admin/global-settings',
        },
        {
            label: 'Code intelligence',
            to: '/site-admin/code-intelligence',
        },
    ],
    registry: [
        {
            label: 'Extensions',
            to: '/site-admin/registry/extensions',
        },
    ],
    auth: [
        {
            label: 'Access Tokens',
            to: '/site-admin/tokens',
        },
    ],
    other: [
        {
            label: 'Updates',
            to: '/site-admin/updates',
        },
        {
            label: 'Analytics',
            to: '/site-admin/analytics',
        },
        {
            label: 'User surveys',
            to: '/site-admin/surveys',
        },
        {
            label: 'Pings',
            to: '/site-admin/pings',
        },
    ],
}

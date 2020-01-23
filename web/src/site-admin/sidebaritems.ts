import LockIcon from 'mdi-react/LockIcon'
import ServerIcon from 'mdi-react/ServerIcon'
import { SiteAdminSideBarGroup, SiteAdminSideBarGroups } from './SiteAdminSidebar'
import SourceRepositoryIcon from 'mdi-react/SourceRepositoryIcon'

export const primaryGroup: SiteAdminSideBarGroup = {
    header: {
        label: 'Site admin',
        icon: ServerIcon,
    },
    items: [
        {
            label: 'Overview',
            to: '/site-admin',
            exact: true,
        },
        {
            label: 'Configuration',
            to: '/site-admin/configuration',
        },
    ],
}

export const secondaryGroup: SiteAdminSideBarGroup = {
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
            label: 'Global settings',
            to: '/site-admin/global-settings',
        },
    ],
}

const repositoriesGroup: SiteAdminSideBarGroup = {
    header: {
        label: 'Repositories',
        icon: SourceRepositoryIcon,
    },
    items: [
        {
            label: 'Manage repositories',
            to: '/site-admin/external-services',
        },
        {
            label: 'Repository status',
            to: '/site-admin/repositories',
        },
    ],
}

export const authGroup: SiteAdminSideBarGroup = {
    header: {
        label: 'Auth',
        icon: LockIcon,
    },
    items: [
        {
            label: 'Access tokens',
            to: '/site-admin/tokens',
        },
    ],
}

export const otherGroup: SiteAdminSideBarGroup = {
    items: [
        {
            label: 'Updates',
            to: '/site-admin/updates',
        },
        {
            label: 'Usage stats',
            to: '/site-admin/usage-statistics',
        },
        {
            label: 'User surveys',
            to: '/site-admin/surveys',
        },
        {
            label: 'Pings',
            to: '/site-admin/pings',
        },
        {
            label: 'Report a bug',
            to: '/site-admin/report-bug',
        },
    ],
}

export const siteAdminSidebarGroups: SiteAdminSideBarGroups = [
    primaryGroup,
    secondaryGroup,
    repositoriesGroup,
    authGroup,
    otherGroup,
]

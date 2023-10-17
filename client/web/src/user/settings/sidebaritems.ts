import type { UserSettingsSidebarItems } from './UserSettingsSidebar'

export const userSettingsSideBarItems: UserSettingsSidebarItems = [
    {
        label: 'Settings',
        to: '',
        exact: true,
    },
    {
        label: 'Profile',
        to: '/profile',
        exact: true,
        condition: ({ isCodyApp }) => !isCodyApp,
    },
    {
        label: 'Emails',
        to: '/emails',
        exact: true,
        condition: ({ isCodyApp }) => !isCodyApp,
    },
    {
        label: 'Access tokens',
        to: '/tokens',
        condition: () => window.context.accessTokensAllow !== 'none',
    },
    {
        label: 'Account security',
        to: '/security',
        exact: true,
        condition: ({ isCodyApp }) => !isCodyApp,
    },
    {
        label: 'Quotas',
        to: '/quota',
        exact: true,
        condition: ({ authenticatedUser }) => authenticatedUser.siteAdmin,
    },
    {
        label: 'Product research',
        to: '/product-research',
        condition: () => window.context.productResearchPageEnabled,
    },
]

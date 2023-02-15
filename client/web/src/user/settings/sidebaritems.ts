import { UserSettingsSidebarItems } from './UserSettingsSidebar'

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
    },
    {
        label: 'Emails',
        to: '/emails',
        exact: true,
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
    },
    {
        label: 'Product research',
        to: '/product-research',
        condition: () => window.context.productResearchPageEnabled,
    },
]

import { showAccountSecurityPage, showPasswordsPage, userExternalServicesEnabled } from './cloud-ga'
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
        label: 'Password',
        to: '/password',
        exact: true,
        // Only the builtin auth provider has a password.
        condition: showPasswordsPage,
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
    //  future GA Cloud nav items
    {
        label: 'Account security',
        to: '/security',
        exact: true,
        condition: showAccountSecurityPage,
    },
    {
        label: 'Code host connections',
        to: '/code-hosts',
        condition: userExternalServicesEnabled,
    },
    {
        label: 'Your repositories',
        to: '/repositories',
        condition: userExternalServicesEnabled,
    },
    {
        label: 'Product research',
        to: '/product-research',
        condition: () => window.context.productResearchPageEnabled,
    },
]

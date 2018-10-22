import { UserAccountSidebarItems } from './UserAccountSidebar'

export const userAccountSideBarItems: UserAccountSidebarItems = {
    account: [
        {
            label: 'Profile',
            to: `/profile`,
            exact: true,
        },
        {
            label: 'Password',
            to: `/password`,
            exact: true,
            // Only the builtin auth provider has a password.
            condition: ({ authProviders }) => authProviders.some(({ isBuiltin }) => isBuiltin),
        },
        {
            label: 'Emails',
            to: `/emails`,
            exact: true,
        },
        {
            label: 'Access tokens',
            to: `/tokens`,
            condition: () => window.context.accessTokensAllow !== 'none',
        },
    ],
}

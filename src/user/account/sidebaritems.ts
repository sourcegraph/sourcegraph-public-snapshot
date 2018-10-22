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
            condition: ({ siteAdminViewingOtherUser, externalAuthEnabled }) =>
                siteAdminViewingOtherUser && !externalAuthEnabled,
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

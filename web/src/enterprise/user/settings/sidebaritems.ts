import { userAccountSideBarItems } from '../../../user/settings/sidebaritems'
import { authExp } from '../../site-admin/SiteAdminAuthenticationProvidersPage'

export const enterpriseUserAccountSideBarItems = {
    ...userAccountSideBarItems,
    account: [
        ...userAccountSideBarItems.account.slice(0, 1),
        {
            label: 'External accounts',
            to: `/external-accounts`,
            exact: true,
            condition: () => authExp,
        },
        ...userAccountSideBarItems.account.slice(2),
    ],
}

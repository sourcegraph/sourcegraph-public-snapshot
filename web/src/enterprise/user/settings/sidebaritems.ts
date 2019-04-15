import { userSettingsSideBarItems } from '../../../user/settings/sidebaritems'
import { authExp } from '../../site-admin/SiteAdminAuthenticationProvidersPage'

export const enterpriseUserSettingsSideBarItems = {
    ...userSettingsSideBarItems,
    account: [
        ...userSettingsSideBarItems.account.slice(0, 1),
        {
            label: 'External accounts',
            to: `/external-accounts`,
            exact: true,
            condition: () => authExp,
        },
        ...userSettingsSideBarItems.account.slice(2),
    ],
}

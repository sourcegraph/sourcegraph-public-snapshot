import { siteAdminSidebarNavItems } from '../../site-admin/sidebaritems'
import { SiteAdminSideBarItems } from '../../site-admin/SiteAdminSidebar'

export const enterpriseSiteAdminSidebarNavItems: SiteAdminSideBarItems = {
    ...siteAdminSidebarNavItems,
    auth: [
        {
            label: 'Providers',
            to: '/site-admin/auth/providers',
        },
        {
            label: 'External Accounts',
            to: '/site-admin/auth/external-accounts',
        },
        ...siteAdminSidebarNavItems.auth,
    ],
}

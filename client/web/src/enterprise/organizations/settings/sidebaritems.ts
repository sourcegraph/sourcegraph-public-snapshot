import { OrgSettingsSidebarItems } from '../../../org/settings/OrgSettingsSidebar'
import { orgSettingsSideBarItems } from '../../../org/settings/sidebaritems'

export const enterpriseOrgSettingsSideBarItems: OrgSettingsSidebarItems = [
    ...orgSettingsSideBarItems,
    {
        to: '/executors/secrets',
        label: 'Executor secrets',
        condition: ({ org: { viewerCanAdminister } }) => viewerCanAdminister,
    },
]

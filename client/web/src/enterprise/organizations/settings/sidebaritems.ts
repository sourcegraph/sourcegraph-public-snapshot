import { canWriteBatchChanges } from '../../../batches/utils'
import { OrgSettingsSidebarItems } from '../../../org/settings/OrgSettingsSidebar'
import { orgSettingsSideBarItems } from '../../../org/settings/sidebaritems'

export const enterpriseOrgSettingsSideBarItems: OrgSettingsSidebarItems = [
    ...orgSettingsSideBarItems,
    {
        to: '/executors/secrets',
        label: 'Executor secrets',
        condition: ({ batchChangesEnabled, org: { viewerCanAdminister }, authenticatedUser }) =>
            batchChangesEnabled && viewerCanAdminister && canWriteBatchChanges(authenticatedUser),
    },
]

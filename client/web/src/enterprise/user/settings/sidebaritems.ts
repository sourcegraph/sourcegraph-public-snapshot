import { userSettingsSideBarItems } from '../../../user/settings/sidebaritems'
import { UserSettingsSidebarItems } from '../../../user/settings/UserSettingsSidebar'
import { canWriteBatchChanges } from '../../batches/utils'
import { SHOW_BUSINESS_FEATURES } from '../../dotcom/productSubscriptions/features'

export const enterpriseUserSettingsSideBarItems: UserSettingsSidebarItems = [
    ...userSettingsSideBarItems.slice(0, 2),
    {
        label: 'Subscriptions',
        to: '/subscriptions',
        condition: ({ user }) => SHOW_BUSINESS_FEATURES && user.viewerCanAdminister,
    },
    {
        to: '/batch-changes',
        label: 'Batch Changes',
        condition: ({ batchChangesEnabled, user: { viewerCanAdminister }, authenticatedUser }) =>
            batchChangesEnabled && viewerCanAdminister && canWriteBatchChanges(authenticatedUser),
    },
    {
        to: '/executors/secrets',
        label: 'Executor secrets',
        condition: ({ batchChangesEnabled, user: { viewerCanAdminister }, authenticatedUser }) =>
            batchChangesEnabled && viewerCanAdminister && canWriteBatchChanges(authenticatedUser),
    },
    ...userSettingsSideBarItems.slice(2),
    {
        label: 'Permissions',
        to: '/permissions',
        exact: true,
        condition: ({ isSourcegraphApp }) => !isSourcegraphApp,
    },
    {
        to: '/event-log',
        label: 'Event log',
        condition: ({ user: { viewerCanAdminister } }) => viewerCanAdminister,
    },
]

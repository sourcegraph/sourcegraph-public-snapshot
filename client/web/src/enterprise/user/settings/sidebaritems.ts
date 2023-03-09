import { userSettingsSideBarItems } from '../../../user/settings/sidebaritems'
import { UserSettingsSidebarItems } from '../../../user/settings/UserSettingsSidebar'
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
        condition: ({ batchChangesEnabled, user: { viewerCanAdminister } }) =>
            batchChangesEnabled && viewerCanAdminister,
    },
    {
        to: '/executors/secrets',
        label: 'Executor secrets',
        condition: ({ user: { viewerCanAdminister } }) => viewerCanAdminister,
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

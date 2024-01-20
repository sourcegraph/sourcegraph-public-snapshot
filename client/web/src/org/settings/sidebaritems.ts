import { canWriteBatchChanges } from '../../batches/utils'
import { isCodyOnlyLicense } from '../../util/license'

import type { OrgSettingsSidebarItems } from './OrgSettingsSidebar'

const disableCodeSearchFeatures = isCodyOnlyLicense()

export const orgSettingsSideBarItems: OrgSettingsSidebarItems = [
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
        label: 'Members',
        to: '/members',
        exact: true,
    },
    {
        to: '/executors/secrets',
        label: 'Executor secrets',
        condition: ({ batchChangesEnabled, org: { viewerCanAdminister }, authenticatedUser }) =>
            !disableCodeSearchFeatures &&
            batchChangesEnabled &&
            viewerCanAdminister &&
            canWriteBatchChanges(authenticatedUser),
    },
]
